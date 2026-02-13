package instruct

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mindspec/mindspec/internal/bead"
)

func TestCheckWorktree_EmptyBead(t *testing.T) {
	warning := CheckWorktree("")
	if warning != "" {
		t.Errorf("expected no warning for empty bead, got %q", warning)
	}
}

func TestCheckWorktree_WorktreeExistsCWDMatches(t *testing.T) {
	orig := worktreeListFn
	t.Cleanup(func() { worktreeListFn = orig })

	worktreeListFn = func() ([]bead.WorktreeListEntry, error) {
		// Return a worktree whose path matches CWD — but we can't easily
		// guarantee CWD, so we test the "exists but different CWD" case instead.
		return []bead.WorktreeListEntry{
			{Name: "worktree-beads-001", Path: "/some/other/path", Branch: "bead/beads-001"},
		}, nil
	}

	warning := CheckWorktree("beads-001")
	// Worktree exists but CWD differs — should get informational message, not a scary warning
	if warning == "" {
		t.Error("expected informational message when worktree exists but CWD differs")
	}
	if !strings.Contains(warning, "Switch to worktree") {
		t.Errorf("expected 'Switch to worktree' message, got %q", warning)
	}
	if strings.Contains(warning, "mismatch") {
		t.Errorf("should not contain 'mismatch' — this is informational, got %q", warning)
	}
}

func TestCheckWorktree_WorktreeDoesNotExist(t *testing.T) {
	orig := worktreeListFn
	t.Cleanup(func() { worktreeListFn = orig })

	worktreeListFn = func() ([]bead.WorktreeListEntry, error) {
		return []bead.WorktreeListEntry{
			{Name: "worktree-other", Path: "/some/path", Branch: "bead/other"},
		}, nil
	}

	warning := CheckWorktree("beads-001")
	if warning == "" {
		t.Error("expected warning when worktree doesn't exist")
	}
	if !strings.Contains(warning, "No worktree found") {
		t.Errorf("expected 'No worktree found' message, got %q", warning)
	}
	if !strings.Contains(warning, "mindspec next") {
		t.Errorf("warning should suggest mindspec next, got %q", warning)
	}
}

func TestCheckWorktree_ListUnavailable(t *testing.T) {
	orig := worktreeListFn
	t.Cleanup(func() { worktreeListFn = orig })

	worktreeListFn = func() ([]bead.WorktreeListEntry, error) {
		return nil, fmt.Errorf("bd not available")
	}

	warning := CheckWorktree("beads-001")
	if warning != "" {
		t.Errorf("expected no warning when bd unavailable, got %q", warning)
	}
}
