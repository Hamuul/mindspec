package glossary

import (
	"sort"
	"strings"
)

// Match returns glossary entries whose terms appear in the input text.
// Matching is case-insensitive. Results are ordered longest-match-first
// to prefer more specific terms (e.g. "Context Pack" before "Context").
func Match(entries []Entry, text string) []Entry {
	lower := strings.ToLower(text)

	// Sort by term length descending for longest-match-first
	sorted := make([]Entry, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool {
		return len(sorted[i].Term) > len(sorted[j].Term)
	})

	var matched []Entry
	seen := make(map[string]bool)

	for _, e := range sorted {
		termLower := strings.ToLower(e.Term)
		if strings.Contains(lower, termLower) && !seen[e.Term] {
			seen[e.Term] = true
			matched = append(matched, e)
		}
	}

	return matched
}
