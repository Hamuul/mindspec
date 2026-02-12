package contextpack

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ADR represents a parsed Architecture Decision Record.
type ADR struct {
	ID      string
	Path    string
	Status  string
	Domains []string
	Content string
}

// ScanADRs reads all ADR-*.md files from the ADR directory.
func ScanADRs(root string) ([]ADR, error) {
	adrDir := filepath.Join(root, "docs", "adr")
	pattern := filepath.Join(adrDir, "ADR-*.md")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("globbing ADRs: %w", err)
	}

	var adrs []ADR
	for _, path := range matches {
		adr, err := parseADR(path)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", filepath.Base(path), err)
		}
		adrs = append(adrs, adr)
	}

	return adrs, nil
}

func parseADR(path string) (ADR, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ADR{}, err
	}

	content := string(data)
	base := filepath.Base(path)
	id := strings.TrimSuffix(base, ".md")

	adr := ADR{
		ID:      id,
		Path:    path,
		Content: content,
	}

	// Parse metadata lines
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Match "- **Status**: X" or "* **Status**: X"
		if strings.Contains(trimmed, "**Status**:") {
			adr.Status = extractValue(trimmed, "**Status**:")
		}
		if strings.Contains(trimmed, "**Domain(s)**:") {
			domStr := extractValue(trimmed, "**Domain(s)**:")
			for _, d := range strings.Split(domStr, ",") {
				d = strings.ToLower(strings.TrimSpace(d))
				if d != "" {
					adr.Domains = append(adr.Domains, d)
				}
			}
		}
	}

	return adr, nil
}

func extractValue(line, key string) string {
	idx := strings.Index(line, key)
	if idx < 0 {
		return ""
	}
	return strings.TrimSpace(line[idx+len(key):])
}

// FilterADRs returns ADRs with Status "Accepted" that have at least one
// domain in common with the provided domain list.
func FilterADRs(adrs []ADR, domains []string) []ADR {
	domainSet := make(map[string]bool)
	for _, d := range domains {
		domainSet[strings.ToLower(strings.TrimSpace(d))] = true
	}

	var result []ADR
	for _, adr := range adrs {
		if !strings.EqualFold(adr.Status, "Accepted") {
			continue
		}
		for _, d := range adr.Domains {
			if domainSet[d] {
				result = append(result, adr)
				break
			}
		}
	}
	return result
}
