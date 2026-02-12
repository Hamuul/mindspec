package instruct

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CheckWorktree verifies that the current working directory is the expected
// worktree for the active bead. Returns a warning message if mismatched,
// or empty string if OK or check is not applicable.
func CheckWorktree(activeBead string) string {
	if activeBead == "" {
		return ""
	}

	expectedSuffix := "worktree-" + activeBead

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf("Could not determine working directory: %v", err)
	}

	// Check if current directory name matches the expected worktree
	currentDir := filepath.Base(cwd)
	if currentDir == expectedSuffix {
		return ""
	}

	// Also check against git worktree list to find the correct path
	worktreePath := findWorktreePath(expectedSuffix)
	if worktreePath != "" {
		if cwd == worktreePath {
			return ""
		}
		return fmt.Sprintf("Worktree mismatch: you are in %s but bead %s expects worktree at %s. Run: cd %s", cwd, activeBead, worktreePath, worktreePath)
	}

	return fmt.Sprintf("Worktree mismatch: you are in %s but bead %s expects a worktree named %s (not found). Create it with: git worktree add ../%s", cwd, activeBead, expectedSuffix, expectedSuffix)
}

// findWorktreePath runs `git worktree list` and looks for a worktree matching the expected name.
func findWorktreePath(expectedName string) string {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "worktree ") {
			path := strings.TrimPrefix(line, "worktree ")
			if filepath.Base(path) == expectedName {
				return path
			}
		}
	}

	return ""
}
