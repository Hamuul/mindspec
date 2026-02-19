package specmeta

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractFrontmatter_WithFrontmatter(t *testing.T) {
	content := "---\nmolecule_id: mol-123\n---\n# Spec 001: Test\n\n## Goal\n"
	fm, body := extractFrontmatter(content)

	if fm != "molecule_id: mol-123" {
		t.Errorf("expected frontmatter 'molecule_id: mol-123', got %q", fm)
	}
	if !strings.Contains(body, "# Spec 001: Test") {
		t.Errorf("expected body to contain heading, got %q", body)
	}
}

func TestExtractFrontmatter_NoFrontmatter(t *testing.T) {
	content := "# Spec 001: Test\n\n## Goal\n"
	fm, body := extractFrontmatter(content)

	if fm != "" {
		t.Errorf("expected empty frontmatter, got %q", fm)
	}
	if body != content {
		t.Errorf("expected body to be full content")
	}
}

func TestExtractFrontmatter_EmptyFrontmatter(t *testing.T) {
	content := "---\n---\n# Spec\n"
	fm, body := extractFrontmatter(content)

	if fm != "" {
		t.Errorf("expected empty frontmatter string, got %q", fm)
	}
	if !strings.Contains(body, "# Spec") {
		t.Errorf("expected body to contain heading, got %q", body)
	}
}

func TestReadWrite_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "spec.md")

	original := "# Spec 001: Test\n\n## Goal\n\nDo something.\n"
	os.WriteFile(specPath, []byte(original), 0644)

	// Write molecule binding
	m := &Meta{
		MoleculeID: "mol-abc",
		StepMapping: map[string]string{
			"spec":         "step-1",
			"spec-approve": "step-2",
		},
	}
	if err := Write(dir, m); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	// Read back
	got, err := Read(dir)
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}

	if got.MoleculeID != "mol-abc" {
		t.Errorf("MoleculeID = %q, want %q", got.MoleculeID, "mol-abc")
	}
	if got.StepMapping["spec"] != "step-1" {
		t.Errorf("StepMapping[spec] = %q, want %q", got.StepMapping["spec"], "step-1")
	}
	if got.StepMapping["spec-approve"] != "step-2" {
		t.Errorf("StepMapping[spec-approve] = %q, want %q", got.StepMapping["spec-approve"], "step-2")
	}

	// Verify the file still contains the original heading
	data, _ := os.ReadFile(specPath)
	content := string(data)
	if !strings.Contains(content, "# Spec 001: Test") {
		t.Errorf("original heading lost after write, content:\n%s", content)
	}
	if !strings.Contains(content, "## Goal") {
		t.Errorf("Goal section lost after write, content:\n%s", content)
	}
}

func TestRead_NoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "spec.md")
	os.WriteFile(specPath, []byte("# Spec 001\n\n## Goal\n"), 0644)

	m, err := Read(dir)
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if m.MoleculeID != "" {
		t.Errorf("expected empty MoleculeID, got %q", m.MoleculeID)
	}
}

func TestRead_EmptyMoleculeID(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "spec.md")
	content := "---\nmolecule_id: \"\"\n---\n# Spec 001\n"
	os.WriteFile(specPath, []byte(content), 0644)

	m, err := Read(dir)
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if m.MoleculeID != "" {
		t.Errorf("expected empty MoleculeID, got %q", m.MoleculeID)
	}
}

func TestWrite_PreservesExistingFrontmatter(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "spec.md")

	// Start with existing frontmatter
	original := "---\ncustom_field: value\n---\n# Spec 001\n"
	os.WriteFile(specPath, []byte(original), 0644)

	m := &Meta{MoleculeID: "mol-xyz"}
	if err := Write(dir, m); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	data, _ := os.ReadFile(specPath)
	content := string(data)

	if !strings.Contains(content, "custom_field: value") {
		t.Errorf("existing frontmatter field lost, content:\n%s", content)
	}
	if !strings.Contains(content, "molecule_id: mol-xyz") {
		t.Errorf("molecule_id not written, content:\n%s", content)
	}
}

func TestWrite_UpdatesExistingMoleculeID(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "spec.md")

	original := "---\nmolecule_id: old-id\n---\n# Spec 001\n"
	os.WriteFile(specPath, []byte(original), 0644)

	m := &Meta{MoleculeID: "new-id"}
	if err := Write(dir, m); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	got, err := Read(dir)
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if got.MoleculeID != "new-id" {
		t.Errorf("MoleculeID = %q, want %q", got.MoleculeID, "new-id")
	}
}

func TestRead_MissingFile(t *testing.T) {
	_, err := Read("/nonexistent")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
