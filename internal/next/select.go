package next

import "fmt"

// SelectWork picks a work item from the list.
// If exactly one item, it is returned directly (auto-claim).
// If multiple, it prints a numbered list and returns the selected index.
// The pick parameter selects a specific item (1-based); 0 means default to first.
func SelectWork(items []BeadInfo, pick int) (BeadInfo, error) {
	if len(items) == 0 {
		return BeadInfo{}, fmt.Errorf("no items to select from")
	}

	if len(items) == 1 {
		return items[0], nil
	}

	// Multiple items — validate pick or default to first
	if pick > 0 {
		if pick > len(items) {
			return BeadInfo{}, fmt.Errorf("pick %d out of range (1-%d)", pick, len(items))
		}
		return items[pick-1], nil
	}

	// Default: return first item
	return items[0], nil
}

// FormatWorkList returns a formatted numbered list of work items for display.
func FormatWorkList(items []BeadInfo) string {
	var result string
	for i, item := range items {
		result += fmt.Sprintf("  %d. [%s] %s (P%d, %s)\n", i+1, item.ID, item.Title, item.Priority, item.IssueType)
	}
	return result
}
