package approve

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/mindspec/mindspec/internal/bead"
	"github.com/mindspec/mindspec/internal/config"
	"github.com/mindspec/mindspec/internal/gitops"
	"github.com/mindspec/mindspec/internal/recording"
	"github.com/mindspec/mindspec/internal/specmeta"
	"github.com/mindspec/mindspec/internal/state"
)

var (
	implRunBDCombinedFn = bead.RunBDCombined
	implRunBDFn         = bead.RunBD
	loadConfigFn        = config.Load
	mergeBranchFn       = gitops.MergeBranch
	deleteBranchFn      = gitops.DeleteBranch
	worktreeRemoveFn    = bead.WorktreeRemove
	hasRemoteFn         = gitops.HasRemote
	pushBranchFn        = gitops.PushBranch
	createPRFn          = gitops.CreatePR
	diffStatFn          = gitops.DiffStat
	commitCountFn       = gitops.CommitCount
	prStatusFn          = gitops.PRStatus
	prChecksWatchFn     = gitops.PRChecksWatch
	mergePRFn           = gitops.MergePR
)

// ImplOpts holds options for implementation approval.
type ImplOpts struct {
	Wait bool // If true and strategy is PR, wait for CI checks then merge.
}

// ImplResult holds the result of implementation approval.
type ImplResult struct {
	SpecID        string
	Warnings      []string
	MergeStrategy string // "direct", "pr", or "" if no merge
	SpecBranch    string
	CommitCount   int
	DiffStat      string
	PRURL         string // set when strategy is "pr"
	PRMerged      bool   // true if PR was merged via --wait
}

// ApproveImpl transitions from review mode to idle, completing the spec lifecycle.
func ApproveImpl(root, specID string, opts ...ImplOpts) (*ImplResult, error) {
	var opt ImplOpts
	if len(opts) > 0 {
		opt = opts[0]
	}
	_ = opt // used below in merge flow
	result := &ImplResult{SpecID: specID}

	// Verify current state is review mode for this spec
	s, err := state.Read(root)
	if err != nil {
		return nil, fmt.Errorf("reading state: %w", err)
	}
	if s.Mode != state.ModeReview {
		return nil, fmt.Errorf("expected review mode, got %q", s.Mode)
	}
	if s.ActiveSpec != specID {
		return nil, fmt.Errorf("active spec is %q, not %q", s.ActiveSpec, specID)
	}

	// Resolve and enforce molecule binding before mutation.
	meta, err := specmeta.EnsureFullyBound(root, specID)
	if err != nil {
		return nil, fmt.Errorf("resolving molecule binding for %s: %w", specID, err)
	}

	// Close parent epic + all unique mapped steps (best-effort).
	for _, targetID := range closeoutTargets(meta) {
		status, err := readBeadStatus(targetID)
		if err == nil && status == "closed" {
			continue
		}

		if _, err := implRunBDCombinedFn("close", targetID); err != nil {
			if isAlreadyClosedErr(err) {
				continue
			}
			result.Warnings = append(result.Warnings, fmt.Sprintf("could not close molecule member %s: %v", targetID, err))
		}
	}

	// Merge spec branch → main (ADR-0006: one PR per spec lifecycle).
	if s.SpecBranch != "" {
		cfg, cfgErr := loadConfigFn(root)
		if cfgErr != nil {
			cfg = config.DefaultConfig()
		}

		mergeErr := mergeSpecToMain(root, s, cfg, result, opt)
		if mergeErr != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("spec→main merge: %v", mergeErr))
		} else if result.MergeStrategy == "direct" || result.PRMerged {
			// Clean up spec worktree and branch after successful merge.
			specWtName := "worktree-spec-" + specID
			if err := worktreeRemoveFn(specWtName); err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("could not remove spec worktree: %v", err))
			}
			if err := deleteBranchFn(s.SpecBranch); err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("could not delete spec branch: %v", err))
			}
		}
	}

	// Stop recording (best-effort — before transitioning to idle)
	if err := recording.StopRecording(root, specID); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("could not stop recording: %v", err))
	}

	// Transition to idle
	idleState := &state.State{Mode: state.ModeIdle}
	if err := state.Write(root, idleState); err != nil {
		return nil, fmt.Errorf("setting state to idle: %w", err)
	}

	return result, nil
}

func closeoutTargets(meta *specmeta.Meta) []string {
	seen := map[string]struct{}{}
	var targets []string

	add := func(id string) {
		id = strings.TrimSpace(id)
		if id == "" {
			return
		}
		if _, exists := seen[id]; exists {
			return
		}
		seen[id] = struct{}{}
		targets = append(targets, id)
	}

	add(meta.MoleculeID)

	var remaining []string
	for _, id := range meta.StepMapping {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		remaining = append(remaining, id)
	}
	sort.Strings(remaining)
	targets = append(targets, remaining...)

	return targets
}

func readBeadStatus(id string) (string, error) {
	out, err := implRunBDFn("show", id, "--json")
	if err != nil {
		return "", err
	}

	var payload []struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(out, &payload); err != nil {
		return "", fmt.Errorf("parsing bd show output for %s: %w", id, err)
	}
	if len(payload) == 0 {
		return "", fmt.Errorf("no bead returned for %s", id)
	}
	return strings.ToLower(strings.TrimSpace(payload[0].Status)), nil
}

// mergeSpecToMain merges the spec branch to main using the configured strategy.
// It populates result with merge metadata (strategy, stats, PR URL).
func mergeSpecToMain(root string, s *state.State, cfg *config.Config, result *ImplResult, opt ImplOpts) error {
	strategy := cfg.MergeStrategy

	// "auto" resolves to "pr" if a remote exists, "direct" otherwise.
	if strategy == "auto" {
		if hasRemoteFn() {
			strategy = "pr"
		} else {
			strategy = "direct"
		}
	}

	result.MergeStrategy = strategy
	result.SpecBranch = s.SpecBranch

	// Gather pre-merge stats (best-effort).
	if count, err := commitCountFn(root, "main", s.SpecBranch); err == nil {
		result.CommitCount = count
	}
	if stat, err := diffStatFn(root, "main", s.SpecBranch); err == nil {
		result.DiffStat = stat
	}

	switch strategy {
	case "pr":
		if err := pushBranchFn(s.SpecBranch); err != nil {
			return fmt.Errorf("pushing spec branch: %w", err)
		}
		title := fmt.Sprintf("[SPEC %s] Merge spec branch to main", s.ActiveSpec)
		body := fmt.Sprintf("Automated PR for spec %s lifecycle completion.", s.ActiveSpec)
		prURL, err := createPRFn(s.SpecBranch, "main", title, body)
		if err != nil {
			return fmt.Errorf("creating PR: %w", err)
		}
		result.PRURL = prURL

		if opt.Wait {
			fmt.Printf("Waiting for CI checks on %s...\n", prURL)
			if err := prChecksWatchFn(prURL); err != nil {
				return fmt.Errorf("CI checks failed: %w", err)
			}
			if err := mergePRFn(prURL); err != nil {
				return fmt.Errorf("merging PR: %w", err)
			}
			result.PRMerged = true
		}
		return nil

	case "direct":
		return mergeBranchFn(root, s.SpecBranch, "main")

	default:
		return fmt.Errorf("unknown merge strategy: %s", strategy)
	}
}

func isAlreadyClosedErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "already closed")
}
