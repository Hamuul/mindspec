package glossary

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExtractSection reads a markdown file and extracts the section matching the
// given anchor. The anchor is matched against heading text by converting both
// to a slug form (lowercase, spaces/special chars to hyphens). Content is
// returned from the matching heading through to the next heading at the same
// or higher level.
//
// If anchor is empty, the full file content is returned.
// If the anchor is not found, an actionable error is returned.
func ExtractSection(root, filePath, anchor string) (string, error) {
	fullPath := filepath.Join(root, filePath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("cannot read %s: %w", filePath, err)
	}

	content := string(data)

	if anchor == "" {
		return content, nil
	}

	lines := strings.Split(content, "\n")
	startIdx := -1
	startLevel := 0

	for i, line := range lines {
		level := headingLevel(line)
		if level == 0 {
			continue
		}
		// Check explicit {#id} attribute first, then fall back to slugified text
		slug := headingID(line)
		if slug == "" {
			slug = slugify(headingText(line))
		}
		if slug == anchor {
			startIdx = i
			startLevel = level
			break
		}
	}

	if startIdx < 0 {
		return "", fmt.Errorf("anchor %q not found in %s — check that a heading with this slug exists", anchor, filePath)
	}

	// Collect lines from startIdx until the next heading at same or higher level
	endIdx := len(lines)
	for i := startIdx + 1; i < len(lines); i++ {
		level := headingLevel(lines[i])
		if level > 0 && level <= startLevel {
			endIdx = i
			break
		}
	}

	section := strings.Join(lines[startIdx:endIdx], "\n")
	return strings.TrimRight(section, "\n"), nil
}

// headingLevel returns the heading level (1-6) for a markdown heading line,
// or 0 if the line is not a heading.
func headingLevel(line string) int {
	trimmed := strings.TrimLeft(line, " ")
	level := 0
	for _, ch := range trimmed {
		if ch == '#' {
			level++
		} else {
			break
		}
	}
	if level > 0 && level <= 6 && len(trimmed) > level && trimmed[level] == ' ' {
		return level
	}
	return 0
}

// headingID extracts an explicit {#id} attribute from a heading line.
// Returns empty string if no explicit ID is present.
func headingID(line string) string {
	idx := strings.LastIndex(line, "{#")
	if idx < 0 || !strings.HasSuffix(strings.TrimSpace(line), "}") {
		return ""
	}
	trimmed := strings.TrimSpace(line)
	end := strings.LastIndex(trimmed, "}")
	start := strings.LastIndex(trimmed, "{#")
	if start >= 0 && end > start+2 {
		return trimmed[start+2 : end]
	}
	return ""
}

// headingText extracts the text content from a markdown heading line,
// stripping the leading `#` characters, any trailing `{#id}` attribute, and whitespace.
func headingText(line string) string {
	trimmed := strings.TrimLeft(line, " ")
	// Strip leading '#' characters
	for len(trimmed) > 0 && trimmed[0] == '#' {
		trimmed = trimmed[1:]
	}
	trimmed = strings.TrimSpace(trimmed)
	// Strip trailing {#id} attribute if present
	if idx := strings.LastIndex(trimmed, "{#"); idx >= 0 && strings.HasSuffix(trimmed, "}") {
		trimmed = strings.TrimSpace(trimmed[:idx])
	}
	return trimmed
}

// slugify converts heading text to a markdown anchor slug.
// Lowercase, replace spaces and non-alphanumeric with hyphens, collapse hyphens.
func slugify(text string) string {
	text = strings.ToLower(text)
	var b strings.Builder
	prevHyphen := false
	for _, ch := range text {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			b.WriteRune(ch)
			prevHyphen = false
		} else if ch == ' ' || ch == '-' || ch == '_' {
			if !prevHyphen && b.Len() > 0 {
				b.WriteByte('-')
				prevHyphen = true
			}
		}
		// other characters are dropped
	}
	result := b.String()
	return strings.TrimRight(result, "-")
}
