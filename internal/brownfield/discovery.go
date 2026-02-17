package brownfield

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// Report captures deterministic brownfield discovery output.
type Report struct {
	MarkdownFiles []string
}

// DiscoverMarkdown scans root for markdown files and returns deterministic output.
func DiscoverMarkdown(root string) (*Report, error) {
	var files []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			name := d.Name()
			// Skip VCS and bead internals from brownfield corpus discovery.
			if name == ".git" || name == ".beads" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.EqualFold(filepath.Ext(d.Name()), ".md") {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("rel path for %s: %w", path, err)
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(files)
	return &Report{MarkdownFiles: files}, nil
}

// FormatSummary renders a compact report summary.
func (r *Report) FormatSummary() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Brownfield discovery report\n")
	fmt.Fprintf(&b, "  Markdown files discovered: %d\n", len(r.MarkdownFiles))
	for i, f := range r.MarkdownFiles {
		if i >= 20 {
			remaining := len(r.MarkdownFiles) - 20
			fmt.Fprintf(&b, "  ... and %d more\n", remaining)
			break
		}
		fmt.Fprintf(&b, "  - %s\n", f)
	}
	return b.String()
}

// RunApply is a placeholder for transactional brownfield apply.
func RunApply(root, archiveMode string) (*Report, error) {
	report, err := DiscoverMarkdown(root)
	if err != nil {
		return nil, err
	}
	return report, fmt.Errorf("brownfield apply is not implemented yet (requested archive mode: %s); use --report-only", archiveMode)
}
