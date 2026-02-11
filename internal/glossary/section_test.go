package glossary

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testMarkdown = `# Top Level

Intro text.

## Beads

Beads are the work graph primitive.

They track issues and dependencies.

## Context Map

The context map shows relationships.

### Context Map Details

More details here.

## Worktrees

Worktrees isolate execution.
`

func writeTestFile(t *testing.T, root, name, content string) {
	t.Helper()
	dir := filepath.Dir(filepath.Join(root, name))
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestExtractSection_ByAnchor(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "doc.md", testMarkdown)

	section, err := ExtractSection(root, "doc.md", "beads")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(section, "## Beads") {
		t.Errorf("expected section to start with '## Beads', got: %s", section[:min(50, len(section))])
	}
	if !strings.Contains(section, "work graph primitive") {
		t.Error("expected section to contain bead content")
	}
	// Should NOT contain the next same-level section
	if strings.Contains(section, "## Context Map") {
		t.Error("section should stop before next same-level heading")
	}
}

func TestExtractSection_StopsAtSameLevel(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "doc.md", testMarkdown)

	section, err := ExtractSection(root, "doc.md", "context-map")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(section, "## Context Map") {
		t.Errorf("expected section to start with '## Context Map'")
	}
	// Should include subsection (###)
	if !strings.Contains(section, "### Context Map Details") {
		t.Error("expected subsection to be included")
	}
	// Should NOT include next same-level heading
	if strings.Contains(section, "## Worktrees") {
		t.Error("section should stop before next same-level heading")
	}
}

func TestExtractSection_AnchorNotFound(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "doc.md", testMarkdown)

	_, err := ExtractSection(root, "doc.md", "nonexistent")
	if err == nil {
		t.Error("expected error for missing anchor")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error should mention the anchor, got: %v", err)
	}
}

func TestExtractSection_NoAnchor(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "doc.md", testMarkdown)

	content, err := ExtractSection(root, "doc.md", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(content, "# Top Level") {
		t.Error("expected full file content when no anchor")
	}
	if !strings.Contains(content, "## Worktrees") {
		t.Error("expected full file content to include all sections")
	}
}

func TestExtractSection_FileNotFound(t *testing.T) {
	root := t.TempDir()
	_, err := ExtractSection(root, "missing.md", "anchor")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestExtractSection_ExplicitID(t *testing.T) {
	root := t.TempDir()
	content := `# Doc

## Human-in-the-Loop Gates {#human-gates}

Always stop and request confirmation.

## Next Section

Other content.
`
	writeTestFile(t, root, "doc.md", content)

	section, err := ExtractSection(root, "doc.md", "human-gates")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(section, "Human-in-the-Loop Gates") {
		t.Error("expected section with explicit {#id} to be found")
	}
	if strings.Contains(section, "## Next Section") {
		t.Error("section should stop before next heading")
	}
}

func TestHeadingID(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"## Human-in-the-Loop Gates {#human-gates}", "human-gates"},
		{"## No ID Here", ""},
		{"### Beads {#beads}", "beads"},
		{"## Just text", ""},
	}
	for _, tt := range tests {
		got := headingID(tt.line)
		if got != tt.want {
			t.Errorf("headingID(%q) = %q, want %q", tt.line, got, tt.want)
		}
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Beads", "beads"},
		{"ADR Lifecycle", "adr-lifecycle"},
		{"Context Map", "context-map"},
		{"Spec Mode", "spec-mode"},
		{"Responsibility Boundaries", "responsibility-boundaries"},
	}
	for _, tt := range tests {
		got := slugify(tt.input)
		if got != tt.want {
			t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestHeadingLevel(t *testing.T) {
	tests := []struct {
		line string
		want int
	}{
		{"# Top", 1},
		{"## Section", 2},
		{"### Sub", 3},
		{"Not a heading", 0},
		{"#NoSpace", 0},
		{"", 0},
	}
	for _, tt := range tests {
		got := headingLevel(tt.line)
		if got != tt.want {
			t.Errorf("headingLevel(%q) = %d, want %d", tt.line, got, tt.want)
		}
	}
}
