package instruct

import (
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

func TestCheckWorktree_Mismatch(t *testing.T) {
	orig := worktreeInfoFn
	t.Cleanup(func() { worktreeInfoFn = orig })

	worktreeInfoFn = func() (*bead.WorktreeInfoResult, error) {
		return &bead.WorktreeInfoResult{
			IsWorktree: true,
			Name:       "worktree-other-bead",
		}, nil
	}

	warning := CheckWorktree("beads-001")
	if warning == "" {
		t.Error("expected worktree mismatch warning")
	}
	if !strings.Contains(warning, "beads-001") {
		t.Errorf("warning should mention bead ID, got %q", warning)
	}
	if !strings.Contains(warning, "worktree-beads-001") {
		t.Errorf("warning should mention expected worktree name, got %q", warning)
	}
}

func TestCheckWorktree_Match(t *testing.T) {
	orig := worktreeInfoFn
	t.Cleanup(func() { worktreeInfoFn = orig })

	worktreeInfoFn = func() (*bead.WorktreeInfoResult, error) {
		return &bead.WorktreeInfoResult{
			IsWorktree: true,
			Name:       "worktree-beads-001",
		}, nil
	}

	warning := CheckWorktree("beads-001")
	if warning != "" {
		t.Errorf("expected no warning for matching worktree, got %q", warning)
	}
}

func TestCheckWorktree_NotInWorktree(t *testing.T) {
	orig := worktreeInfoFn
	t.Cleanup(func() { worktreeInfoFn = orig })

	worktreeInfoFn = func() (*bead.WorktreeInfoResult, error) {
		return &bead.WorktreeInfoResult{
			IsWorktree: false,
		}, nil
	}

	warning := CheckWorktree("beads-001")
	if warning == "" {
		t.Error("expected warning when not in a worktree")
	}
	if !strings.Contains(warning, "mindspec next") {
		t.Errorf("warning should suggest mindspec next, got %q", warning)
	}
}
