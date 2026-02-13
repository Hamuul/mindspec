package instruct

import (
	"fmt"
	"os"

	"github.com/mindspec/mindspec/internal/bead"
)

// worktreeInfoFn is a package-level variable for testability.
var worktreeInfoFn = bead.WorktreeInfo

// CheckWorktree verifies that the current working directory is the expected
// worktree for the active bead. Returns a warning message if mismatched,
// or empty string if OK or check is not applicable.
func CheckWorktree(activeBead string) string {
	if activeBead == "" {
		return ""
	}

	expectedName := "worktree-" + activeBead

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf("Could not determine working directory: %v", err)
	}

	info, err := worktreeInfoFn()
	if err != nil {
		// bd worktree info not available — fall back to name check
		return fmt.Sprintf("Worktree mismatch: you are in %s but bead %s expects a worktree named %s (not found). Create it with: mindspec next", cwd, activeBead, expectedName)
	}

	if info.IsWorktree && info.Name == expectedName {
		return ""
	}

	return fmt.Sprintf("Worktree mismatch: you are in %s but bead %s expects a worktree named %s (not found). Create it with: mindspec next", cwd, activeBead, expectedName)
}
