package bead

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// execCommand is a package-level variable for testability.
// Tests override this to capture arguments or return canned output.
var execCommand = exec.Command

// BeadInfo represents a work item from the Beads CLI.
type BeadInfo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Priority    int    `json:"priority"`
	IssueType   string `json:"issue_type"`
	Owner       string `json:"owner"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// Preflight checks prerequisites for bead commands:
// git repo, .beads/ directory, bd on PATH.
func Preflight(root string) error {
	// Check git repo
	cmd := execCommand("git", "rev-parse", "--git-dir")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("not a git repository (run 'git init'): %s", string(out))
	}

	// Check .beads/ exists
	beadsDir := filepath.Join(root, ".beads")
	if _, err := os.Stat(beadsDir); os.IsNotExist(err) {
		return fmt.Errorf(".beads/ directory not found (run 'beads init' to initialize)")
	}

	// Check bd on PATH
	if _, err := exec.LookPath("bd"); err != nil {
		return fmt.Errorf("bd not found on PATH (install Beads: https://github.com/beads-project/beads)")
	}

	return nil
}

// Create creates a new bead via `bd create` and returns the created bead info.
func Create(title, desc, issueType string, priority int, parent string) (*BeadInfo, error) {
	args := []string{"create", title,
		"--description=" + desc,
		"--type=" + issueType,
		fmt.Sprintf("--priority=%d", priority),
		"--json",
	}
	if parent != "" {
		args = append(args, "--parent="+parent)
	}

	cmd := execCommand("bd", args...)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("bd create failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("bd create failed: %w", err)
	}

	var info BeadInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, fmt.Errorf("parsing bd create output: %w", err)
	}
	return &info, nil
}

// Search searches for beads matching query, returning only open beads.
func Search(query string) ([]BeadInfo, error) {
	cmd := execCommand("bd", "search", query, "--json", "--status=open")
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("bd search failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("bd search failed: %w", err)
	}

	var items []BeadInfo
	if err := json.Unmarshal(out, &items); err != nil {
		return nil, fmt.Errorf("parsing bd search output: %w", err)
	}
	return items, nil
}

// Show returns details for a single bead by ID.
func Show(id string) (*BeadInfo, error) {
	cmd := execCommand("bd", "show", id, "--json")
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("bd show failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("bd show failed: %w", err)
	}

	var info BeadInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, fmt.Errorf("parsing bd show output: %w", err)
	}
	return &info, nil
}

// ListOpen returns all open beads.
func ListOpen() ([]BeadInfo, error) {
	cmd := execCommand("bd", "list", "--status=open", "--json")
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("bd list failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("bd list failed: %w", err)
	}

	var items []BeadInfo
	if err := json.Unmarshal(out, &items); err != nil {
		return nil, fmt.Errorf("parsing bd list output: %w", err)
	}
	return items, nil
}

// DepAdd adds a dependency: blocked depends on blocker.
func DepAdd(blocked, blocker string) error {
	cmd := execCommand("bd", "dep", "add", blocked, blocker)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("bd dep add failed: %s", string(out))
	}
	return nil
}

// Update changes a bead's status.
func Update(id, status string) error {
	cmd := execCommand("bd", "update", id, "--status="+status)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("bd update failed: %s", string(out))
	}
	return nil
}

// parseBeadList parses JSON output containing a list of BeadInfo.
func parseBeadList(data []byte) ([]BeadInfo, error) {
	var items []BeadInfo
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("parsing beads JSON: %w", err)
	}
	return items, nil
}
