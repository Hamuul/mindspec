package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mindspec/mindspec/internal/trace"
	"github.com/mindspec/mindspec/internal/workspace"
)

// Valid mode values.
const (
	ModeIdle      = "idle"
	ModeExplore   = "explore"
	ModeSpec      = "spec"
	ModePlan      = "plan"
	ModeImplement = "implement"
	ModeReview    = "review"
)

// ValidModes lists all valid mode values.
var ValidModes = []string{ModeIdle, ModeExplore, ModeSpec, ModePlan, ModeImplement, ModeReview}

// Session holds transient per-session metadata persisted at .mindspec/session.json.
type Session struct {
	SessionSource    string `json:"sessionSource,omitempty"`
	SessionStartedAt string `json:"sessionStartedAt,omitempty"`
	BeadClaimedAt    string `json:"beadClaimedAt,omitempty"`
}

// ModeCache is a write-through cache for hook latency, persisted at .mindspec/mode-cache.
// Lifecycle commands write this after mutating molecule state. Hooks read it
// instead of calling bd mol show on every PreToolUse invocation.
type ModeCache struct {
	Mode           string `json:"mode"`
	ActiveSpec     string `json:"activeSpec,omitempty"`
	ActiveBead     string `json:"activeBead,omitempty"`
	ActiveWorktree string `json:"activeWorktree,omitempty"`
	SpecBranch     string `json:"specBranch,omitempty"`
	Timestamp      string `json:"timestamp"`
}

// ReadSession loads the session from .mindspec/session.json under root.
// Returns a zero Session (no error) if the file does not exist.
func ReadSession(root string) (*Session, error) {
	path := workspace.SessionPath(root)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Session{}, nil
		}
		return nil, fmt.Errorf("reading session file: %w", err)
	}

	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing session file: %w", err)
	}
	return &s, nil
}

// WriteSessionFile persists the session to .mindspec/session.json under root.
func WriteSessionFile(root string, s *Session) error {
	dir := workspace.MindspecDir(root)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating .mindspec directory: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling session: %w", err)
	}
	data = append(data, '\n')

	return os.WriteFile(workspace.SessionPath(root), data, 0644)
}

// ReadModeCache loads the mode cache from .mindspec/mode-cache under root.
// Returns nil (no error) if the file does not exist.
func ReadModeCache(root string) (*ModeCache, error) {
	path := workspace.ModeCachePath(root)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading mode-cache: %w", err)
	}

	var mc ModeCache
	if err := json.Unmarshal(data, &mc); err != nil {
		return nil, fmt.Errorf("parsing mode-cache: %w", err)
	}
	return &mc, nil
}

// WriteModeCache persists the mode cache to .mindspec/mode-cache under root.
func WriteModeCache(root string, mc *ModeCache) error {
	dir := workspace.MindspecDir(root)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating .mindspec directory: %w", err)
	}

	mc.Timestamp = time.Now().UTC().Format(time.RFC3339)

	data, err := json.MarshalIndent(mc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling mode-cache: %w", err)
	}
	data = append(data, '\n')

	return os.WriteFile(workspace.ModeCachePath(root), data, 0644)
}

// SpecBranch returns the canonical branch name for a spec.
func SpecBranch(specID string) string { return "spec/" + specID }

// SpecWorktreePath returns the canonical worktree path for a spec.
func SpecWorktreePath(root, specID string) string {
	return filepath.Join(root, ".worktrees", "worktree-spec-"+specID)
}

// BeadWorktreePath returns the canonical worktree path for a bead
// nested under its spec's worktree.
func BeadWorktreePath(specWorktree, beadID string) string {
	return filepath.Join(specWorktree, ".worktrees", "worktree-"+beadID)
}

// State represents the MindSpec workflow state persisted at .mindspec/state.json.
// Deprecated: will be removed in Bead 6. Use Session, ModeCache, and resolve.* instead.
type State struct {
	Mode             string            `json:"mode"`
	ActiveSpec       string            `json:"activeSpec"`
	ActiveBead       string            `json:"activeBead"`
	ActiveWorktree   string            `json:"activeWorktree,omitempty"`
	SpecBranch       string            `json:"specBranch,omitempty"`
	ActiveMolecule   string            `json:"activeMolecule,omitempty"`
	StepMapping      map[string]string `json:"stepMapping,omitempty"`
	SessionSource    string            `json:"sessionSource,omitempty"`
	SessionStartedAt string            `json:"sessionStartedAt,omitempty"`
	BeadClaimedAt    string            `json:"beadClaimedAt,omitempty"`
	LastUpdated      string            `json:"lastUpdated"`
}

// ErrNoState is returned when .mindspec/state.json does not exist.
// Deprecated: will be removed in Bead 6.
var ErrNoState = errors.New("no .mindspec/state.json found")

// Read loads the state from .mindspec/state.json under root.
// Deprecated: will be removed in Bead 6. Use ReadSession/ReadModeCache and resolve.* instead.
func Read(root string) (*State, error) {
	path := workspace.StatePath(root)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoState
		}
		return nil, fmt.Errorf("reading state file: %w", err)
	}

	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing state file: %w", err)
	}
	return &s, nil
}

// Write persists the state to .mindspec/state.json under root.
// Deprecated: will be removed in Bead 6. Use WriteSessionFile/WriteModeCache instead.
func Write(root string, s *State) error {
	dir := workspace.MindspecDir(root)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating .mindspec directory: %w", err)
	}

	s.LastUpdated = time.Now().UTC().Format(time.RFC3339)

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}
	data = append(data, '\n')

	path := workspace.StatePath(root)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}

	// Dual-write: if we're in a worktree, also write to the main worktree.
	if mainRoot, ok := mainWorktreeRoot(root); ok && mainRoot != root {
		mainDir := workspace.MindspecDir(mainRoot)
		if err := os.MkdirAll(mainDir, 0755); err == nil {
			mainPath := workspace.StatePath(mainRoot)
			// Best-effort — don't fail the primary write if this fails.
			_ = os.WriteFile(mainPath, data, 0644)
		}
	}

	return nil
}

// WriteSession records session freshness metadata (source and timestamp).
// Deprecated: will be removed in Bead 6. Use WriteSessionFile instead.
func WriteSession(root, source string) error {
	s, err := Read(root)
	if err != nil {
		if err == ErrNoState {
			s = &State{Mode: ModeIdle}
		} else {
			return err
		}
	}
	s.SessionSource = source
	s.SessionStartedAt = time.Now().UTC().Format(time.RFC3339)
	return Write(root, s)
}

// mainWorktreeRoot returns the main worktree's root path if the given root
// is inside a git worktree. Returns ("", false) if root is the main worktree
// or detection fails.
func mainWorktreeRoot(root string) (string, bool) {
	// In a git worktree, .git is a file containing "gitdir: <path>".
	// In the main worktree, .git is a directory.
	gitPath := filepath.Join(root, ".git")
	info, err := os.Lstat(gitPath)
	if err != nil {
		return "", false
	}
	if info.IsDir() {
		// This is the main worktree — no propagation needed.
		return "", false
	}

	// .git is a file — read it to find the main repo.
	data, err := os.ReadFile(gitPath)
	if err != nil {
		return "", false
	}
	// Format: "gitdir: /path/to/main/.git/worktrees/<name>\n"
	line := strings.TrimSpace(string(data))
	if !strings.HasPrefix(line, "gitdir: ") {
		return "", false
	}
	gitdir := strings.TrimPrefix(line, "gitdir: ")

	// Walk up from gitdir to find the main .git directory.
	// gitdir is typically: <main>/.git/worktrees/<name>
	// We need: <main>
	idx := strings.Index(gitdir, filepath.Join(".git", "worktrees"))
	if idx <= 0 {
		return "", false
	}
	mainRoot := gitdir[:idx-1] // strip trailing separator

	// Verify the main root has .mindspec/
	if _, err := os.Stat(workspace.MindspecDir(mainRoot)); err != nil {
		return "", false
	}

	return mainRoot, true
}

// SetMode validates inputs and writes a new state. Emits a trace event on transition.
// Deprecated: will be removed in Bead 6. Use WriteModeCache instead.
func SetMode(root, mode, spec, bead string) error {
	return SetModeWithMetadata(root, mode, spec, bead, "", nil)
}

// SetModeWithMetadata validates inputs and writes a new state.
// Deprecated: will be removed in Bead 6. Use WriteModeCache instead.
func SetModeWithMetadata(root, mode, spec, bead, moleculeID string, stepMapping map[string]string) error {
	// Read previous state for trace event
	prevMode := "none"
	var prev *State
	if p, err := Read(root); err == nil {
		prev = p
		prevMode = p.Mode
	}
	trace.Emit(trace.NewEvent("state.transition").
		WithSpec(spec).
		WithData(map[string]any{
			"from":    prevMode,
			"to":      mode,
			"spec_id": spec,
		}))
	if !isValidMode(mode) {
		return fmt.Errorf("invalid mode %q: must be one of %v", mode, ValidModes)
	}

	if mode == ModeSpec || mode == ModePlan || mode == ModeImplement || mode == ModeReview {
		if spec == "" {
			return fmt.Errorf("mode %q requires --spec", mode)
		}
		specDir := workspace.SpecDir(root, spec)
		if _, err := os.Stat(specDir); os.IsNotExist(err) {
			return fmt.Errorf("spec directory does not exist: %s", specDir)
		}
	}

	if mode == ModeImplement && bead == "" {
		return fmt.Errorf("mode %q requires --bead", mode)
	}

	s := &State{
		Mode:       mode,
		ActiveSpec: spec,
		ActiveBead: bead,
	}
	if spec != "" && mode != ModeIdle {
		if moleculeID != "" {
			s.ActiveMolecule = moleculeID
			s.StepMapping = copyStepMapping(stepMapping)
		} else if prev != nil && prev.ActiveSpec == spec {
			s.ActiveMolecule = prev.ActiveMolecule
			s.StepMapping = copyStepMapping(prev.StepMapping)
		}
		// Preserve worktree/branch bindings across transitions for the same spec.
		if prev != nil && prev.ActiveSpec == spec {
			s.ActiveWorktree = prev.ActiveWorktree
			s.SpecBranch = prev.SpecBranch
		}
	}

	return Write(root, s)
}

func copyStepMapping(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func isValidMode(mode string) bool {
	for _, m := range ValidModes {
		if m == mode {
			return true
		}
	}
	return false
}
