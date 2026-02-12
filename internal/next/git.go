package next

import (
	"fmt"
	"os/exec"
	"strings"
)

// CheckCleanTree verifies the git working tree is clean.
// Returns nil if clean, error with dirty state details if not.
func CheckCleanTree() error {
	out, err := exec.Command("git", "status", "--porcelain").Output()
	if err != nil {
		return fmt.Errorf("running git status: %w", err)
	}

	status := strings.TrimSpace(string(out))
	if status != "" {
		return fmt.Errorf("working tree is dirty — commit or discard changes before claiming work:\n%s", status)
	}

	return nil
}
