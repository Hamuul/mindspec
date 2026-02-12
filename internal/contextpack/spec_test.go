package contextpack

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSpec(t *testing.T) {
	dir := t.TempDir()
	specDir := filepath.Join(dir, "docs", "specs", "001-test")
	os.MkdirAll(specDir, 0o755)

	content := `# Spec 001: Test Spec

## Goal

Build a test feature for validation purposes.

## Impacted Domains

- core: CLI and workspace
- context-system: context pack generation

## Requirements

1. Something
`
	os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(content), 0o644)

	meta, err := ParseSpec(specDir)
	if err != nil {
		t.Fatalf("ParseSpec: %v", err)
	}

	if meta.SpecID != "001-test" {
		t.Errorf("SpecID = %q, want %q", meta.SpecID, "001-test")
	}

	if meta.Goal != "Build a test feature for validation purposes." {
		t.Errorf("Goal = %q, want %q", meta.Goal, "Build a test feature for validation purposes.")
	}

	if len(meta.Domains) != 2 {
		t.Fatalf("Domains count = %d, want 2", len(meta.Domains))
	}
	if meta.Domains[0] != "core" {
		t.Errorf("Domains[0] = %q, want %q", meta.Domains[0], "core")
	}
	if meta.Domains[1] != "context-system" {
		t.Errorf("Domains[1] = %q, want %q", meta.Domains[1], "context-system")
	}
}

func TestParseSpec_MissingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := ParseSpec(dir)
	if err == nil {
		t.Fatal("expected error for missing spec.md")
	}
}
