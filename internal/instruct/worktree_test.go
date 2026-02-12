package instruct

import (
	"strings"
	"testing"
)

func TestCheckWorktree_EmptyBead(t *testing.T) {
	warning := CheckWorktree("")
	if warning != "" {
		t.Errorf("expected no warning for empty bead, got %q", warning)
	}
}

func TestCheckWorktree_Mismatch(t *testing.T) {
	// We're not in a worktree named worktree-beads-001, so this should warn
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
