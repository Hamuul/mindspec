package bench

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

// SessionDef describes one benchmark session configuration.
type SessionDef struct {
	Label        string // e.g., "a", "b", "c"
	Description  string // e.g., "no-docs", "baseline", "mindspec"
	Port         int
	Neutralize   func(wtPath string) error // nil for session C (full MindSpec)
	EnableTrace  bool
	Prompt       string // per-session prompt override (if non-empty, overrides cfg.Prompt)
	ExternalOTLP string // if set, skip local collector and send OTLP here
}

// SessionResult holds the outcome of a single benchmark session.
type SessionResult struct {
	Label      string
	JSONLPath  string
	OutputPath string
	EventCount int
	ExitCode   int
	TimedOut   bool
}

// RunSessionWithRetries executes a benchmark session with retry-based auto-approve.
// After each attempt, it checks whether implementation code was produced. If not,
// it auto-approves any pending workflow gates (spec/plan for session C) and retries
// with an escalating prompt directing the agent to implement.
func RunSessionWithRetries(ctx context.Context, cfg *RunConfig, def *SessionDef, wtPath string) (*SessionResult, error) {
	jsonlPath := filepath.Join(cfg.WorkDir, fmt.Sprintf("session-%s.jsonl", def.Label))
	outputPath := filepath.Join(cfg.WorkDir, fmt.Sprintf("output-%s.txt", def.Label))

	result := &SessionResult{
		Label:      def.Label,
		JSONLPath:  jsonlPath,
		OutputPath: outputPath,
	}

	baseCommit := getCurrentCommit(wtPath)

	// Start in-process collector for the entire retry loop (skip if using external OTLP)
	var collectorCancel context.CancelFunc
	var collectorDone chan error
	if def.ExternalOTLP == "" {
		collector := NewCollector(def.Port, jsonlPath)
		var collectorCtx context.Context
		collectorCtx, collectorCancel = context.WithCancel(ctx)

		collectorDone = make(chan error, 1)
		go func() {
			collectorDone <- collector.Run(collectorCtx)
		}()

		if err := waitForPort(def.Port, 5*time.Second); err != nil {
			collectorCancel()
			return nil, fmt.Errorf("collector failed to start on port %d: %w", def.Port, err)
		}
	}

	// Create output file (append across retries)
	outFile, err := os.Create(outputPath)
	if err != nil {
		if collectorCancel != nil {
			collectorCancel()
		}
		return nil, fmt.Errorf("creating output file: %w", err)
	}
	defer outFile.Close()

	// Build environment — use external endpoint or per-session port
	var env []string
	if def.ExternalOTLP != "" {
		env = buildSessionEnvEndpoint(def.ExternalOTLP, cfg.WorkDir, def.Label, def.EnableTrace)
	} else {
		env = buildSessionEnv(def.Port, cfg.WorkDir, def.Label, def.EnableTrace)
	}

	prompt := def.Prompt
	if prompt == "" {
		prompt = cfg.Prompt
	}

	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			fmt.Fprintf(cfg.Stdout, "\n── Retry %d/%d (auto-approve) ──\n\n", attempt, maxRetries)
			fmt.Fprintf(outFile, "\n\n--- RETRY %d/%d ---\n\n", attempt, maxRetries)
		}

		exitCode, timedOut, _ := runClaude(ctx, prompt, wtPath, env,
			cfg.MaxTurns, cfg.Model, cfg.Timeout, cfg.Stdout, outFile)
		result.ExitCode = exitCode
		result.TimedOut = timedOut

		// Commit any uncommitted changes
		commitWorktreeChanges(wtPath, fmt.Sprintf("%s-attempt-%d", def.Label, attempt))

		// Check if implementation code was produced
		if hasCodeChanges(wtPath, baseCommit) {
			fmt.Fprintf(cfg.Stdout, "  Implementation detected.\n")
			break
		}

		if attempt < maxRetries {
			fmt.Fprintf(cfg.Stdout, "  No implementation code detected. Auto-approving...\n")
			autoApprove(def.Label, wtPath, cfg.SpecID)
			prompt = buildRetryPrompt(def.Label, wtPath, cfg.SpecID, attempt+1)
		}
	}

	// Shut down local collector if we started one
	if collectorCancel != nil {
		time.Sleep(2 * time.Second)
		collectorCancel()
		<-collectorDone
	}

	// Count events (only meaningful when using local collector)
	if data, err := os.ReadFile(jsonlPath); err == nil {
		for _, b := range data {
			if b == '\n' {
				result.EventCount++
			}
		}
	}

	return result, nil
}

// commitWorktreeChanges commits any changes made during the session.
func commitWorktreeChanges(wtPath, label string) {
	// Check if there are changes
	cmd := exec.Command("git", "-C", wtPath, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil || len(out) == 0 {
		return
	}

	exec.Command("git", "-C", wtPath, "add", "-A").Run()                                                      //nolint:errcheck
	exec.Command("git", "-C", wtPath, "commit", "-m", "bench: Session "+label+" output", "--no-verify").Run() //nolint:errcheck
}

// waitForPort waits until the given port is accepting TCP connections.
func waitForPort(port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	addr := fmt.Sprintf("localhost:%d", port)

	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return fmt.Errorf("port %d not ready after %s", port, timeout)
}

// CheckPortFree returns nil if the port is not in use.
func CheckPortFree(port int) error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 500*time.Millisecond)
	if err != nil {
		return nil // port is free
	}
	conn.Close()
	return fmt.Errorf("port %d is already in use", port)
}

// runClaude executes a single claude -p invocation. It is the core execution
// primitive used by both RunSession and runWithRetries.
func runClaude(ctx context.Context, prompt, wtPath string, env []string,
	maxTurns int, model string, timeout time.Duration,
	stdout io.Writer, outFile *os.File) (exitCode int, timedOut bool, err error) {

	args := []string{"-p", prompt, "--dangerously-skip-permissions", "--no-session-persistence"}
	if maxTurns > 0 {
		args = append(args, "--max-turns", fmt.Sprintf("%d", maxTurns))
	}
	if model != "" {
		args = append(args, "--model", model)
	}

	sessionCtx, sessionCancel := context.WithTimeout(ctx, timeout)
	defer sessionCancel()

	cmd := exec.CommandContext(sessionCtx, "claude", args...)
	cmd.Dir = wtPath
	cmd.Stdout = io.MultiWriter(stdout, outFile)
	cmd.Stderr = io.MultiWriter(stdout, outFile)
	cmd.Cancel = func() error {
		return cmd.Process.Signal(syscall.SIGTERM)
	}
	cmd.WaitDelay = 10 * time.Second
	cmd.Env = env

	runErr := cmd.Run()
	if runErr != nil {
		if sessionCtx.Err() == context.DeadlineExceeded {
			timedOut = true
		}
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}
	return exitCode, timedOut, nil
}

// buildSessionEnv creates the environment for a benchmark Claude session.
func buildSessionEnv(port int, workDir, label string, enableTrace bool) []string {
	return buildSessionEnvEndpoint(
		fmt.Sprintf("http://localhost:%d", port),
		workDir, label, enableTrace,
	)
}

// buildSessionEnvEndpoint creates the environment with an explicit OTLP endpoint.
func buildSessionEnvEndpoint(endpoint, workDir, label string, enableTrace bool) []string {
	env := os.Environ()
	env = append(env,
		"CLAUDECODE=",
		"CLAUDE_CODE_ENABLE_TELEMETRY=1",
		"OTEL_METRICS_EXPORTER=otlp",
		"OTEL_LOGS_EXPORTER=otlp",
		"OTEL_EXPORTER_OTLP_PROTOCOL=http/json",
		"OTEL_EXPORTER_OTLP_ENDPOINT="+endpoint,
	)
	if enableTrace {
		tracePath := filepath.Join(workDir, fmt.Sprintf("trace-%s.jsonl", label))
		env = append(env, "MINDSPEC_TRACE="+tracePath)
	}
	return env
}

// getCurrentCommit returns the HEAD commit SHA of a worktree.
func getCurrentCommit(wtPath string) string {
	cmd := exec.Command("git", "-C", wtPath, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return trimNewline(string(out))
}

// hasCodeChanges checks if implementation source files were created since baseCommit.
func hasCodeChanges(wtPath, baseCommit string) bool {
	cmd := exec.Command("git", "-C", wtPath, "diff", "--name-only", baseCommit+"..HEAD")
	out, err := cmd.Output()
	if err != nil {
		return false
	}

	codeExts := regexp.MustCompile(`\.(go|js|ts|html|css|jsx|tsx)$`)
	excludeDirs := []string{"docs/", ".claude/", ".mindspec/", ".beads/"}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !codeExts.MatchString(line) {
			continue
		}
		excluded := false
		for _, dir := range excludeDirs {
			if strings.HasPrefix(line, dir) {
				excluded = true
				break
			}
		}
		if !excluded {
			return true
		}
	}
	return false
}

// autoApprove advances workflow gates between retries.
func autoApprove(label, wtPath, specID string) {
	if label != "c" {
		return // A/B: no state to advance, rely on retry prompt
	}

	// Read MindSpec state
	stateData, err := os.ReadFile(filepath.Join(wtPath, ".mindspec", "state.json"))
	if err != nil {
		return
	}
	var state map[string]string
	if err := json.Unmarshal(stateData, &state); err != nil {
		return
	}

	mode := state["mode"]
	switch mode {
	case "spec":
		// Approve the spec: update frontmatter, advance to plan mode
		specPath := findSpecFile(wtPath, specID)
		if specPath != "" {
			updateFrontmatterApproval(specPath)
		}
		writeState(wtPath, "plan", specID, "")

	case "plan":
		// Approve the plan: update frontmatter, advance to implement mode
		planPath := filepath.Join(wtPath, "docs", "specs", specID, "plan.md")
		if _, err := os.Stat(planPath); err == nil {
			updateFrontmatterApproval(planPath)
		}
		writeState(wtPath, "implement", specID, "bench-impl")
	}
}

// buildRetryPrompt generates the prompt for a retry attempt.
func buildRetryPrompt(label, wtPath, specID string, attempt int) string {
	if label != "c" {
		// Sessions A/B: escalating implementation prompts
		if attempt == 1 {
			return "Your plan is approved. Proceed to implementation. Write all code and tests, then commit your changes."
		}
		return "Implementation is required. Write the code now and commit all changes."
	}

	// Session C: check MindSpec state and give workflow-appropriate prompt
	stateData, err := os.ReadFile(filepath.Join(wtPath, ".mindspec", "state.json"))
	if err != nil {
		return "Continue implementing. Write all remaining code and commit."
	}
	var state map[string]string
	if err := json.Unmarshal(stateData, &state); err != nil {
		return "Continue implementing. Write all remaining code and commit."
	}

	switch state["mode"] {
	case "plan":
		return fmt.Sprintf("The spec is approved. Create a plan at docs/specs/%s/plan.md, then use /plan-approve to approve it. After approval, implement all code and tests. Commit your changes when complete.", specID)
	case "implement":
		return "The plan is approved. Implement all code and tests described in the plan. Commit your changes when complete."
	default:
		return "Continue implementing. Write all remaining code and commit."
	}
}

// prepareSessionC sets MindSpec state to spec mode so hooks emit spec-mode guidance.
func prepareSessionC(wtPath, specID string) {
	stateDir := filepath.Join(wtPath, ".mindspec")
	os.MkdirAll(stateDir, 0755) //nolint:errcheck

	state := map[string]string{
		"mode":        "spec",
		"activeSpec":  specID,
		"activeBead":  "",
		"lastUpdated": time.Now().UTC().Format(time.RFC3339),
	}

	data, _ := json.MarshalIndent(state, "", "  ")
	data = append(data, '\n')
	os.WriteFile(filepath.Join(stateDir, "state.json"), data, 0644) //nolint:errcheck
}

// updateFrontmatterApproval updates a markdown file's YAML frontmatter to set
// status: Approved and record the approval date.
func updateFrontmatterApproval(filePath string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	content := string(data)

	// Replace status field
	statusRe := regexp.MustCompile(`(?mi)^(\s*(?:-\s+)?\*?\*?Status\*?\*?\s*:\s*).*$`)
	if statusRe.MatchString(content) {
		content = statusRe.ReplaceAllString(content, "${1}APPROVED")
	}

	// Also handle YAML frontmatter status field
	fmStatusRe := regexp.MustCompile(`(?m)^status:\s*.*$`)
	if fmStatusRe.MatchString(content) {
		content = fmStatusRe.ReplaceAllString(content, "status: Approved")
	}

	// Set approval date in frontmatter
	approvedAtRe := regexp.MustCompile(`(?m)^approved_at:\s*.*$`)
	now := time.Now().UTC().Format(time.RFC3339)
	if approvedAtRe.MatchString(content) {
		content = approvedAtRe.ReplaceAllString(content, fmt.Sprintf("approved_at: %q", now))
	}

	approvedByRe := regexp.MustCompile(`(?m)^approved_by:\s*.*$`)
	if approvedByRe.MatchString(content) {
		content = approvedByRe.ReplaceAllString(content, "approved_by: bench")
	}

	// Handle markdown-style approval section
	mdApprovedBy := regexp.MustCompile(`(?mi)^(\s*-\s+\*\*Approved By\*\*:\s*).*$`)
	if mdApprovedBy.MatchString(content) {
		content = mdApprovedBy.ReplaceAllString(content, "${1}bench")
	}
	mdApprovedDate := regexp.MustCompile(`(?mi)^(\s*-\s+\*\*Approval Date\*\*:\s*).*$`)
	if mdApprovedDate.MatchString(content) {
		content = mdApprovedDate.ReplaceAllString(content, fmt.Sprintf("${1}%s", time.Now().Format("2006-01-02")))
	}

	os.WriteFile(filePath, []byte(content), 0644) //nolint:errcheck
}

// writeState writes a MindSpec state.json file.
func writeState(wtPath, mode, specID, beadID string) {
	stateDir := filepath.Join(wtPath, ".mindspec")
	os.MkdirAll(stateDir, 0755) //nolint:errcheck

	state := map[string]string{
		"mode":        mode,
		"activeSpec":  specID,
		"activeBead":  beadID,
		"lastUpdated": time.Now().UTC().Format(time.RFC3339),
	}

	data, _ := json.MarshalIndent(state, "", "  ")
	data = append(data, '\n')
	os.WriteFile(filepath.Join(stateDir, "state.json"), data, 0644) //nolint:errcheck
}

// findSpecFile locates the spec.md for a given spec ID in the worktree.
func findSpecFile(wtPath, specID string) string {
	// Try the standard location first
	p := filepath.Join(wtPath, "docs", "specs", specID, "spec.md")
	if _, err := os.Stat(p); err == nil {
		return p
	}
	// Try without the full slug (e.g., "022" instead of "022-agentmind-viz-mvp")
	entries, err := os.ReadDir(filepath.Join(wtPath, "docs", "specs"))
	if err != nil {
		return ""
	}
	prefix := strings.SplitN(specID, "-", 2)[0]
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), prefix) {
			p := filepath.Join(wtPath, "docs", "specs", e.Name(), "spec.md")
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}
	return ""
}

// findSpecRelPath returns the relative path to spec.md from the worktree root.
func findSpecRelPath(wtPath, specID string) string {
	abs := findSpecFile(wtPath, specID)
	if abs == "" {
		return fmt.Sprintf("docs/specs/%s/spec.md", specID)
	}
	rel, err := filepath.Rel(wtPath, abs)
	if err != nil {
		return fmt.Sprintf("docs/specs/%s/spec.md", specID)
	}
	return rel
}
