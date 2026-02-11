package doctor

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// requiredDirs are directories that must exist under docs/ or project root.
var requiredDirs = []struct {
	path string // relative to project root
	name string // display name
}{
	{"docs/core", "docs/core/"},
	{"docs/domains", "docs/domains/"},
	{"docs/specs", "docs/specs/"},
	{"architecture", "architecture/"},
}

// expectedDomains and their expected files.
var expectedDomains = []string{"core", "context-system", "workflow"}
var domainFiles = []string{"overview.md", "architecture.md", "interfaces.md", "runbook.md"}

// termPattern matches glossary table rows: | **Term** | [Link](Target) |
var termPattern = regexp.MustCompile(`\|\s*\*\*([^*]+)\*\*\s*\|`)

// linkPattern extracts the target from a glossary link: [text](target)
var linkPattern = regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)

func checkDocs(r *Report, root string) {
	// Check required directories
	for _, d := range requiredDirs {
		p := filepath.Join(root, d.path)
		if dirExists(p) {
			r.Checks = append(r.Checks, Check{Name: d.name, Status: OK})
		} else {
			r.Checks = append(r.Checks, Check{
				Name:    d.name,
				Status:  Missing,
				Message: fmt.Sprintf("create %s directory", d.path),
			})
		}
	}

	// GLOSSARY.md
	checkGlossary(r, root)

	// context-map.md
	cmPath := filepath.Join(root, "docs", "context-map.md")
	if fileExists(cmPath) {
		r.Checks = append(r.Checks, Check{Name: "docs/context-map.md", Status: OK})
	} else {
		r.Checks = append(r.Checks, Check{
			Name:    "docs/context-map.md",
			Status:  Missing,
			Message: "create docs/context-map.md",
		})
	}

	// Domain subdirectory checks
	checkDomains(r, root)
}

func checkGlossary(r *Report, root string) {
	glossaryPath := filepath.Join(root, "GLOSSARY.md")
	data, err := os.ReadFile(glossaryPath)
	if err != nil {
		r.Checks = append(r.Checks, Check{
			Name:    "GLOSSARY.md",
			Status:  Missing,
			Message: "create GLOSSARY.md in project root",
		})
		return
	}

	content := string(data)

	// Count terms (lines matching | **Term** |)
	terms := termPattern.FindAllString(content, -1)
	r.Checks = append(r.Checks, Check{
		Name:    "GLOSSARY.md",
		Status:  OK,
		Message: fmt.Sprintf("(%d terms)", len(terms)),
	})

	// Check for broken links
	lines := strings.Split(content, "\n")
	var broken []string
	for _, line := range lines {
		if !termPattern.MatchString(line) {
			continue
		}
		linkMatch := linkPattern.FindStringSubmatch(line)
		if linkMatch == nil {
			continue
		}
		target := linkMatch[1]
		// Strip anchor for file existence check
		pathPart := strings.SplitN(target, "#", 2)[0]
		if pathPart == "" {
			continue
		}
		fullPath := filepath.Join(root, pathPart)
		if !fileExists(fullPath) {
			// Extract term name
			termMatch := termPattern.FindStringSubmatch(line)
			term := ""
			if termMatch != nil {
				term = strings.TrimSpace(termMatch[1])
			}
			broken = append(broken, fmt.Sprintf("%s -> %s", term, target))
		}
	}

	if len(broken) > 0 {
		r.Checks = append(r.Checks, Check{
			Name:    "Glossary links",
			Status:  Error,
			Message: fmt.Sprintf("%d broken: %s", len(broken), strings.Join(broken, "; ")),
		})
	} else {
		r.Checks = append(r.Checks, Check{
			Name:    "Glossary links",
			Status:  OK,
			Message: "all links verified",
		})
	}
}

func checkDomains(r *Report, root string) {
	domainsDir := filepath.Join(root, "docs", "domains")
	if !dirExists(domainsDir) {
		return // already reported as missing in requiredDirs
	}

	for _, domain := range expectedDomains {
		domainDir := filepath.Join(domainsDir, domain)
		if !dirExists(domainDir) {
			r.Checks = append(r.Checks, Check{
				Name:    fmt.Sprintf("docs/domains/%s/", domain),
				Status:  Warn,
				Message: "domain directory not found",
			})
			continue
		}

		for _, f := range domainFiles {
			fp := filepath.Join(domainDir, f)
			name := fmt.Sprintf("docs/domains/%s/%s", domain, f)
			if fileExists(fp) {
				r.Checks = append(r.Checks, Check{Name: name, Status: OK})
			} else {
				r.Checks = append(r.Checks, Check{
					Name:    name,
					Status:  Warn,
					Message: "file not found",
				})
			}
		}
	}
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
