package instruct

import (
	"fmt"
	"os"

	"github.com/mindspec/mindspec/internal/bead"
)

// worktreeListFn is a package-level variable for testability.
var worktreeListFn = bead.WorktreeList

// CheckWorktree verifies that the current working directory is the expected
// worktree for the active bead. Returns an informational message if the
// worktree exists but CWD doesn't match, a warning if the worktree doesn't
// exist, or empty string if OK or check is not applicable.
func CheckWorktree(activeBead string) string {
	if activeBead == "" {
		return ""
	}

	expectedName := "worktree-" + activeBead

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf("Could not determine working directory: %v", err)
	}

	entries, err := worktreeListFn()
	if err != nil {
		// bd worktree list not available — can't check
		return ""
	}

	for _, e := range entries {
		if e.Name == expectedName {
			// Worktree exists — check if we're in it
			if e.Path == cwd {
				return ""
			}
			// Exists but CWD differs — informational, not a warning
			return fmt.Sprintf("Switch to worktree to begin work: cd %s", e.Path)
		}
	}

	// Worktree doesn't exist
	return fmt.Sprintf("No worktree found for bead %s. Create it with: mindspec next", activeBead)
}
