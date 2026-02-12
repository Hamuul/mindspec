package contextpack

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SpecMeta holds metadata extracted from a spec file.
type SpecMeta struct {
	SpecID  string
	Goal    string
	Domains []string
}

// ParseSpec reads a spec.md file and extracts the goal and impacted domains.
func ParseSpec(specDir string) (*SpecMeta, error) {
	specPath := filepath.Join(specDir, "spec.md")
	f, err := os.Open(specPath)
	if err != nil {
		return nil, fmt.Errorf("opening spec: %w", err)
	}
	defer f.Close()

	meta := &SpecMeta{
		SpecID: filepath.Base(specDir),
	}

	scanner := bufio.NewScanner(f)
	var section string
	var goalLines []string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "## ") {
			heading := strings.TrimPrefix(line, "## ")
			heading = strings.TrimSpace(heading)

			// If we were collecting goal lines, stop
			if section == "goal" {
				meta.Goal = strings.TrimSpace(strings.Join(goalLines, "\n"))
			}

			switch strings.ToLower(heading) {
			case "goal":
				section = "goal"
				goalLines = nil
			case "impacted domains":
				section = "domains"
			default:
				section = ""
			}
			continue
		}

		switch section {
		case "goal":
			goalLines = append(goalLines, line)
		case "domains":
			if strings.HasPrefix(strings.TrimSpace(line), "- ") {
				bullet := strings.TrimPrefix(strings.TrimSpace(line), "- ")
				// Domain name is text before first colon
				domain := bullet
				if idx := strings.Index(bullet, ":"); idx >= 0 {
					domain = strings.TrimSpace(bullet[:idx])
				}
				domain = strings.ToLower(strings.TrimSpace(domain))
				if domain != "" {
					meta.Domains = append(meta.Domains, domain)
				}
			}
		}
	}

	// Handle goal at end of file
	if section == "goal" && meta.Goal == "" {
		meta.Goal = strings.TrimSpace(strings.Join(goalLines, "\n"))
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading spec: %w", err)
	}

	return meta, nil
}
