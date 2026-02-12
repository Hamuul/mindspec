package bead

import (
	"fmt"
	"path/filepath"
	"strings"
)

// WorktreeEntry represents a single git worktree.
type WorktreeEntry struct {
	Path   string
	Branch string
	HEAD   string
}

// ParseWorktreeList parses the output of `git worktree list --porcelain`.
func ParseWorktreeList() ([]WorktreeEntry, error) {
	cmd := execCommand("git", "worktree", "list", "--porcelain")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git worktree list failed: %s", string(out))
	}
	return parseWorktreePorcelain(string(out)), nil
}

// parseWorktreePorcelain parses porcelain-format worktree output.
// Blocks are separated by blank lines. Each block has:
//
//	worktree <path>
//	HEAD <sha>
//	branch refs/heads/<name>
func parseWorktreePorcelain(output string) []WorktreeEntry {
	var entries []WorktreeEntry
	var current *WorktreeEntry

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimRight(line, "\r")

		if line == "" {
			if current != nil {
				entries = append(entries, *current)
				current = nil
			}
			continue
		}

		if strings.HasPrefix(line, "worktree ") {
			current = &WorktreeEntry{
				Path: strings.TrimPrefix(line, "worktree "),
			}
		} else if strings.HasPrefix(line, "HEAD ") && current != nil {
			current.HEAD = strings.TrimPrefix(line, "HEAD ")
		} else if strings.HasPrefix(line, "branch ") && current != nil {
			branch := strings.TrimPrefix(line, "branch ")
			// Strip refs/heads/ prefix
			current.Branch = strings.TrimPrefix(branch, "refs/heads/")
		}
	}

	// Handle last entry if no trailing newline
	if current != nil {
		entries = append(entries, *current)
	}

	return entries
}

// FindWorktree looks for an existing worktree for the given bead ID.
// Matches by path ending in `worktree-<beadID>` or branch `bead/<beadID>`.
// Returns the worktree path, or empty string if not found.
func FindWorktree(beadID string) (string, error) {
	entries, err := ParseWorktreeList()
	if err != nil {
		return "", err
	}

	pathSuffix := "worktree-" + beadID
	branchName := "bead/" + beadID

	for _, e := range entries {
		if strings.HasSuffix(e.Path, pathSuffix) || e.Branch == branchName {
			return e.Path, nil
		}
	}

	return "", nil
}

// CreateWorktree creates a git worktree for a bead.
// Validates bead is in_progress and working tree is clean.
func CreateWorktree(root, beadID string) (string, error) {
	// Validate bead status
	info, err := Show(beadID)
	if err != nil {
		return "", fmt.Errorf("cannot look up bead %s: %w", beadID, err)
	}
	if info.Status != "in_progress" {
		return "", fmt.Errorf("bead %s is not in_progress (status: %s) — claim it first with 'bd update %s --status=in_progress'", beadID, info.Status, beadID)
	}

	// Validate clean tree
	if err := checkCleanTree(); err != nil {
		return "", err
	}

	// Compute worktree path (sibling to project root)
	wtPath := filepath.Join(filepath.Dir(root), "worktree-"+beadID)
	branchName := "bead/" + beadID

	cmd := execCommand("git", "worktree", "add", wtPath, "-b", branchName)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git worktree add failed: %s", string(out))
	}

	return wtPath, nil
}

// checkCleanTree verifies the git working tree is clean.
func checkCleanTree() error {
	cmd := execCommand("git", "status", "--porcelain")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git status failed: %s", string(out))
	}
	if strings.TrimSpace(string(out)) != "" {
		return fmt.Errorf("working tree is dirty — commit or stash changes before creating a worktree")
	}
	return nil
}
