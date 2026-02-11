package glossary

import (
	"os"
	"path/filepath"
	"testing"
)

const testGlossary = `# MindSpec Glossary

| Term | Target |
|:-----|:-------|
| **ADR** | [docs/core/ARCHITECTURE.md#adr-lifecycle](docs/core/ARCHITECTURE.md#adr-lifecycle) |
| **Bead** | [docs/core/ARCHITECTURE.md#beads](docs/core/ARCHITECTURE.md#beads) |
| **Context Pack** | [docs/core/ARCHITECTURE.md#context-system](docs/core/ARCHITECTURE.md#context-system) |
| **Spec Mode** | [docs/core/MODES.md#spec-mode](docs/core/MODES.md#spec-mode) |
`

func TestParseContent(t *testing.T) {
	entries, err := ParseContent(testGlossary)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}

	// Check first entry
	e := entries[0]
	if e.Term != "ADR" {
		t.Errorf("expected term 'ADR', got %q", e.Term)
	}
	if e.FilePath != "docs/core/ARCHITECTURE.md" {
		t.Errorf("expected FilePath 'docs/core/ARCHITECTURE.md', got %q", e.FilePath)
	}
	if e.Anchor != "adr-lifecycle" {
		t.Errorf("expected Anchor 'adr-lifecycle', got %q", e.Anchor)
	}
	if e.Target != "docs/core/ARCHITECTURE.md#adr-lifecycle" {
		t.Errorf("expected full Target, got %q", e.Target)
	}
}

func TestParseContent_NoAnchor(t *testing.T) {
	content := `| Term | Target |
|:-----|:-------|
| **Simple** | [docs/readme.md](docs/readme.md) |
`
	entries, err := ParseContent(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Anchor != "" {
		t.Errorf("expected empty Anchor, got %q", entries[0].Anchor)
	}
	if entries[0].FilePath != "docs/readme.md" {
		t.Errorf("expected FilePath 'docs/readme.md', got %q", entries[0].FilePath)
	}
}

func TestParse_MissingFile(t *testing.T) {
	root := t.TempDir()
	_, err := Parse(root)
	if err == nil {
		t.Error("expected error for missing GLOSSARY.md")
	}
}

func TestParse_RealFile(t *testing.T) {
	root := t.TempDir()
	glossaryPath := filepath.Join(root, "GLOSSARY.md")
	if err := os.WriteFile(glossaryPath, []byte(testGlossary), 0644); err != nil {
		t.Fatal(err)
	}

	entries, err := Parse(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 4 {
		t.Errorf("expected 4 entries, got %d", len(entries))
	}
}

func TestSplitTarget(t *testing.T) {
	tests := []struct {
		target   string
		wantPath string
		wantAnch string
	}{
		{"docs/core/ARCHITECTURE.md#beads", "docs/core/ARCHITECTURE.md", "beads"},
		{"docs/core/MODES.md", "docs/core/MODES.md", ""},
		{"file.md#a#b", "file.md", "a#b"},
	}
	for _, tt := range tests {
		path, anchor := splitTarget(tt.target)
		if path != tt.wantPath {
			t.Errorf("splitTarget(%q) path = %q, want %q", tt.target, path, tt.wantPath)
		}
		if anchor != tt.wantAnch {
			t.Errorf("splitTarget(%q) anchor = %q, want %q", tt.target, anchor, tt.wantAnch)
		}
	}
}
