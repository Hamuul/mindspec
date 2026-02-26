package instruct

import (
	"strings"
	"testing"
)

func TestCheckWorktree_EmptyPath(t *testing.T) {
	warning := CheckWorktree("")
	if warning != "" {
		t.Errorf("expected no warning for empty path, got %q", warning)
	}
}

func TestCheckWorktree_CWDMatchesWorktree(t *testing.T) {
	orig := getwdFn
	t.Cleanup(func() { getwdFn = orig })

	getwdFn = func() (string, error) { return "/repo/.worktrees/worktree-bead-abc", nil }

	warning := CheckWorktree("/repo/.worktrees/worktree-bead-abc")
	if warning != "" {
		t.Errorf("expected no warning when CWD matches worktree, got %q", warning)
	}
}

func TestCheckWorktree_CWDSubdirOfWorktree(t *testing.T) {
	orig := getwdFn
	t.Cleanup(func() { getwdFn = orig })

	getwdFn = func() (string, error) { return "/repo/.worktrees/worktree-bead-abc/internal/pkg", nil }

	warning := CheckWorktree("/repo/.worktrees/worktree-bead-abc")
	if warning != "" {
		t.Errorf("expected no warning when CWD is subdir of worktree, got %q", warning)
	}
}

func TestCheckWorktree_CWDDiffers(t *testing.T) {
	orig := getwdFn
	t.Cleanup(func() { getwdFn = orig })

	getwdFn = func() (string, error) { return "/repo", nil }

	warning := CheckWorktree("/repo/.worktrees/worktree-bead-abc")
	if warning == "" {
		t.Error("expected warning when CWD differs from worktree")
	}
	if !strings.Contains(warning, "Switch to worktree") {
		t.Errorf("expected 'Switch to worktree' message, got %q", warning)
	}
	if !strings.Contains(warning, "/repo/.worktrees/worktree-bead-abc") {
		t.Errorf("warning should contain worktree path, got %q", warning)
	}
}
