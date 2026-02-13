package next

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/mindspec/mindspec/internal/bead"
)

// BeadInfo represents a work item from Beads.
type BeadInfo struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	Priority  int    `json:"priority"`
	IssueType string `json:"issue_type"`
	Owner     string `json:"owner"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Package-level function variables for testability.
// Tests override these to avoid calling real bd commands.
var (
	searchBeads    = bead.Search
	molReady       = bead.MolReady
	updateBead     = bead.Update
	worktreeList   = bead.WorktreeList
	worktreeCreate = bead.WorktreeCreate
	execCommand    = exec.Command
)

// QueryReady discovers ready work, preferring molecule children when available.
// Searches for molecule parents ([PLAN prefix), queries their ready children,
// and falls back to `bd ready --json` for standalone beads.
func QueryReady() ([]BeadInfo, error) {
	// Try molecule-aware discovery first
	molItems := queryMolChildren()
	if len(molItems) > 0 {
		return molItems, nil
	}

	// Fall back to regular bd ready
	out, err := execCommand("bd", "ready", "--json").Output()
	if err != nil {
		if execErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("bd ready failed: %s", string(execErr.Stderr))
		}
		return nil, fmt.Errorf("running bd ready: %w", err)
	}

	return ParseBeadsJSON(out)
}

// queryMolChildren searches for molecule parents and returns their ready children.
func queryMolChildren() []BeadInfo {
	parents, err := searchBeads("[PLAN ")
	if err != nil || len(parents) == 0 {
		return nil
	}

	var items []BeadInfo
	for _, parent := range parents {
		children, err := molReady(parent.ID)
		if err != nil {
			continue
		}
		items = append(items, convertBeadInfos(children)...)
	}
	return items
}

// convertBeadInfos converts bead.BeadInfo slice to next.BeadInfo slice.
func convertBeadInfos(src []bead.BeadInfo) []BeadInfo {
	result := make([]BeadInfo, len(src))
	for i, s := range src {
		result[i] = BeadInfo{
			ID:        s.ID,
			Title:     s.Title,
			Status:    s.Status,
			Priority:  s.Priority,
			IssueType: s.IssueType,
			Owner:     s.Owner,
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
		}
	}
	return result
}

// ParseBeadsJSON parses the JSON output from bd commands into BeadInfo slices.
func ParseBeadsJSON(data []byte) ([]BeadInfo, error) {
	var items []BeadInfo
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("parsing beads JSON: %w", err)
	}
	return items, nil
}

// ClaimBead marks a bead as in_progress via bead.Update().
func ClaimBead(id string) error {
	return updateBead(id, "in_progress")
}

// EnsureWorktree checks for an existing worktree for the bead, or creates one.
// Returns the worktree path. Returns empty string if worktree creation is not
// applicable (e.g., working on main).
func EnsureWorktree(beadID string) (string, error) {
	entries, err := worktreeList()
	if err != nil {
		return "", fmt.Errorf("listing worktrees: %w", err)
	}

	// Check for existing worktree matching this bead
	wtName := "worktree-" + beadID
	branchName := "bead/" + beadID
	for _, e := range entries {
		if e.Name == wtName || e.Branch == branchName {
			return e.Path, nil
		}
	}

	// Create new worktree via bd worktree create
	if err := worktreeCreate(wtName, branchName); err != nil {
		return "", fmt.Errorf("creating worktree: %w", err)
	}

	// Read back path from worktree list
	entries, err = worktreeList()
	if err != nil {
		return "", fmt.Errorf("reading worktree path: %w", err)
	}
	for _, e := range entries {
		if e.Name == wtName || strings.HasSuffix(e.Path, wtName) {
			return e.Path, nil
		}
	}

	// Fallback: return the name (relative path)
	return wtName, nil
}
