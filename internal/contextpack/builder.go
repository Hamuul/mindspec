package contextpack

import (
	"os"
	"path/filepath"
	"strings"
)

// ExtractSection extracts the content under a markdown ## heading, collecting
// lines until the next ## heading or EOF. Returns empty string if not found.
func ExtractSection(content, heading string) string {
	lines := strings.Split(content, "\n")
	var collecting bool
	var result []string

	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			if collecting {
				break
			}
			h := strings.TrimSpace(strings.TrimPrefix(line, "## "))
			if strings.EqualFold(h, heading) {
				collecting = true
				continue
			}
		}
		if collecting {
			result = append(result, line)
		}
	}

	return strings.TrimSpace(strings.Join(result, "\n"))
}

// ExtractFilePathsFromText scans text for references to source file paths.
func ExtractFilePathsFromText(text string) []string {
	prefixes := []string{"internal/", "cmd/", "pkg/"}
	seen := map[string]bool{}
	var paths []string

	for _, line := range strings.Split(text, "\n") {
		for _, prefix := range prefixes {
			idx := strings.Index(line, prefix)
			for idx >= 0 {
				// Extract path: take chars until whitespace, backtick, paren, comma
				end := idx
				for end < len(line) {
					c := line[end]
					if c == ' ' || c == '\t' || c == '`' || c == ')' || c == '(' || c == ',' || c == ';' {
						break
					}
					end++
				}
				path := line[idx:end]
				// Clean trailing punctuation
				path = strings.TrimRight(path, ".:;,)")
				if path != "" && !seen[path] {
					seen[path] = true
					paths = append(paths, path)
				}
				// Continue scanning the rest of the line
				remaining := line[end:]
				nextIdx := -1
				for _, p := range prefixes {
					if i := strings.Index(remaining, p); i >= 0 && (nextIdx < 0 || i < nextIdx) {
						nextIdx = i
					}
				}
				if nextIdx < 0 {
					break
				}
				idx = end + nextIdx
			}
		}
	}

	return paths
}

// readFileContent reads a file and returns its content, or empty string on error.
func readFileContent(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func relPath(root, p string) string {
	rel, err := filepath.Rel(root, p)
	if err != nil {
		return p
	}
	return rel
}
