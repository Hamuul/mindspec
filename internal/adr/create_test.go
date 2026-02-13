package adr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupCreateEnv(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	adrDir := filepath.Join(root, "docs", "adr")
	tmplDir := filepath.Join(root, "docs", "templates")
	os.MkdirAll(adrDir, 0o755)
	os.MkdirAll(tmplDir, 0o755)

	// Write existing ADRs
	os.WriteFile(filepath.Join(adrDir, "ADR-0001.md"), []byte(testADR1), 0o644)
	os.WriteFile(filepath.Join(adrDir, "ADR-0002.md"), []byte(testADR2), 0o644)

	// Write template
	tmpl := `# ADR-NNNN: <Title>

- **Date**: <YYYY-MM-DD>
- **Status**: Proposed
- **Domain(s)**: <comma-separated list>
- **Deciders**: <who decides>
- **Supersedes**: n/a
- **Superseded-by**: n/a

## Context

<What is the issue?>

## Decision

<What is the change?>
`
	os.WriteFile(filepath.Join(tmplDir, "adr.md"), []byte(tmpl), 0o644)

	return root
}

func TestCreate_HappyPath(t *testing.T) {
	root := setupCreateEnv(t)

	path, err := Create(root, "Use Redis for caching", CreateOpts{})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if !strings.HasSuffix(path, "ADR-0003.md") {
		t.Errorf("path = %q, want suffix ADR-0003.md", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "# ADR-0003: Use Redis for caching") {
		t.Error("expected title in heading")
	}
	if !strings.Contains(content, "**Status**: Proposed") {
		t.Error("expected Proposed status")
	}
}

func TestCreate_EmptyTitle(t *testing.T) {
	root := setupCreateEnv(t)

	_, err := Create(root, "", CreateOpts{})
	if err == nil {
		t.Error("expected error for empty title")
	}
}

func TestCreate_WithDomains(t *testing.T) {
	root := setupCreateEnv(t)

	path, err := Create(root, "Test", CreateOpts{Domains: []string{"core", "workflow"}})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "core, workflow") {
		t.Errorf("expected domains in content, got:\n%s", string(data))
	}
}

func TestCreate_WithSupersedes(t *testing.T) {
	root := setupCreateEnv(t)

	path, err := Create(root, "New Approach", CreateOpts{Supersedes: "ADR-0001"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Check new ADR has Supersedes field
	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, "**Supersedes**: ADR-0001") {
		t.Error("new ADR should reference superseded ADR")
	}

	// Domains should be copied from old ADR
	if !strings.Contains(content, "core, context-system") {
		t.Errorf("expected inherited domains, got:\n%s", content)
	}

	// Check old ADR was updated
	oldData, _ := os.ReadFile(filepath.Join(root, "docs", "adr", "ADR-0001.md"))
	if !strings.Contains(string(oldData), "**Superseded-by**: ADR-0003") {
		t.Errorf("old ADR should reference new ADR, got:\n%s", string(oldData))
	}
}

func TestCreate_SupersedesNotFound(t *testing.T) {
	root := setupCreateEnv(t)

	_, err := Create(root, "Test", CreateOpts{Supersedes: "ADR-9999"})
	if err == nil {
		t.Error("expected error for nonexistent superseded ADR")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %q, want 'not found'", err.Error())
	}
}

func TestCreate_SupersedesWithExplicitDomains(t *testing.T) {
	root := setupCreateEnv(t)

	path, err := Create(root, "Override Domains", CreateOpts{
		Supersedes: "ADR-0001",
		Domains:    []string{"new-domain"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	// Should use explicit domains, not inherited ones
	if !strings.Contains(content, "new-domain") {
		t.Error("expected explicit domain override")
	}
	if strings.Contains(content, "context-system") {
		t.Error("should not inherit domains when explicitly provided")
	}
}
