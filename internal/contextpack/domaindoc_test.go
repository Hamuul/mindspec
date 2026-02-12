package contextpack

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadDomainDocs(t *testing.T) {
	root := t.TempDir()
	domainDir := filepath.Join(root, "docs", "domains", "core")
	os.MkdirAll(domainDir, 0o755)

	os.WriteFile(filepath.Join(domainDir, "overview.md"), []byte("# Core Overview"), 0o644)
	os.WriteFile(filepath.Join(domainDir, "architecture.md"), []byte("# Core Arch"), 0o644)
	// interfaces.md and runbook.md deliberately missing

	doc, err := ReadDomainDocs(root, "core")
	if err != nil {
		t.Fatalf("ReadDomainDocs: %v", err)
	}

	if doc.Domain != "core" {
		t.Errorf("Domain = %q, want %q", doc.Domain, "core")
	}
	if doc.Overview != "# Core Overview" {
		t.Errorf("Overview = %q", doc.Overview)
	}
	if doc.Architecture != "# Core Arch" {
		t.Errorf("Architecture = %q", doc.Architecture)
	}
	if doc.Interfaces != "" {
		t.Errorf("Interfaces should be empty, got %q", doc.Interfaces)
	}
	if doc.Runbook != "" {
		t.Errorf("Runbook should be empty, got %q", doc.Runbook)
	}
}

func TestReadDomainDocs_MissingDomain(t *testing.T) {
	root := t.TempDir()
	doc, err := ReadDomainDocs(root, "nonexistent")
	if err != nil {
		t.Fatalf("ReadDomainDocs should not error for missing domain: %v", err)
	}
	if doc.Overview != "" || doc.Architecture != "" {
		t.Error("expected all empty strings for missing domain")
	}
}
