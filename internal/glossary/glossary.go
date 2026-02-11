package glossary

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/mindspec/mindspec/internal/workspace"
)

// Entry represents a single glossary term and its documentation target.
type Entry struct {
	Term     string // display name, e.g. "Context Pack"
	Label    string // link label text
	Target   string // full target with anchor, e.g. "docs/core/ARCHITECTURE.md#context-system"
	FilePath string // path part, e.g. "docs/core/ARCHITECTURE.md"
	Anchor   string // fragment part, e.g. "context-system" (empty if no anchor)
}

// termPattern matches glossary table rows: | **Term** | ... |
var termPattern = regexp.MustCompile(`\|\s*\*\*([^*]+)\*\*\s*\|`)

// linkPattern extracts label and target from a markdown link: [label](target)
var linkPattern = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

// Parse reads GLOSSARY.md from the given project root and returns all entries.
func Parse(root string) ([]Entry, error) {
	glossaryPath := workspace.GlossaryPath(root)
	data, err := os.ReadFile(glossaryPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read glossary: %w", err)
	}

	return ParseContent(string(data))
}

// ParseContent parses glossary entries from the given markdown content.
func ParseContent(content string) ([]Entry, error) {
	var entries []Entry

	for _, line := range strings.Split(content, "\n") {
		termMatch := termPattern.FindStringSubmatch(line)
		if termMatch == nil {
			continue
		}
		term := strings.TrimSpace(termMatch[1])

		linkMatch := linkPattern.FindStringSubmatch(line)
		if linkMatch == nil {
			continue
		}
		label := linkMatch[1]
		target := linkMatch[2]

		filePath, anchor := splitTarget(target)

		entries = append(entries, Entry{
			Term:     term,
			Label:    label,
			Target:   target,
			FilePath: filePath,
			Anchor:   anchor,
		})
	}

	return entries, nil
}

// splitTarget splits a target like "docs/foo.md#anchor" into path and anchor parts.
func splitTarget(target string) (string, string) {
	parts := strings.SplitN(target, "#", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}
