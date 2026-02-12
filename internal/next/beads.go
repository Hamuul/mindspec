package next

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

// BeadInfo represents a work item from Beads.
type BeadInfo struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	Priority  int    `json:"priority"`
	IssueType string `json:"issue_type"`
	Owner     string `json:"owner"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// QueryReady shells out to `bd ready --json` and parses the result.
func QueryReady() ([]BeadInfo, error) {
	out, err := exec.Command("bd", "ready", "--json").Output()
	if err != nil {
		if execErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("bd ready failed: %s", string(execErr.Stderr))
		}
		return nil, fmt.Errorf("running bd ready: %w", err)
	}

	return ParseBeadsJSON(out)
}

// ParseBeadsJSON parses the JSON output from bd commands into BeadInfo slices.
func ParseBeadsJSON(data []byte) ([]BeadInfo, error) {
	var items []BeadInfo
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("parsing beads JSON: %w", err)
	}
	return items, nil
}

// ClaimBead marks a bead as in_progress via `bd update <id> --status=in_progress`.
func ClaimBead(id string) error {
	out, err := exec.Command("bd", "update", id, "--status=in_progress").CombinedOutput()
	if err != nil {
		return fmt.Errorf("claiming bead %s: %s", id, string(out))
	}
	return nil
}
