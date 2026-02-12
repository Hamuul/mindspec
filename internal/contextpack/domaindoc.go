package contextpack

import (
	"os"
	"path/filepath"
)

// DomainDoc holds the content of a domain's documentation files.
type DomainDoc struct {
	Domain       string
	Overview     string
	Architecture string
	Interfaces   string
	Runbook      string
}

// ReadDomainDocs reads the 4 standard doc files from a domain directory.
// Missing files result in empty strings, not errors.
func ReadDomainDocs(root, domain string) (*DomainDoc, error) {
	dir := filepath.Join(root, "docs", "domains", domain)
	doc := &DomainDoc{Domain: domain}

	doc.Overview = readFileOrEmpty(filepath.Join(dir, "overview.md"))
	doc.Architecture = readFileOrEmpty(filepath.Join(dir, "architecture.md"))
	doc.Interfaces = readFileOrEmpty(filepath.Join(dir, "interfaces.md"))
	doc.Runbook = readFileOrEmpty(filepath.Join(dir, "runbook.md"))

	return doc, nil
}

func readFileOrEmpty(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}
