package bead

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// --- ParsePlanMeta tests ---

func TestParsePlanMeta_WorkChunks(t *testing.T) {
	tmp := t.TempDir()
	planContent := `---
status: Approved
spec_id: 007-beads-tooling
work_chunks:
  - id: 1
    title: "bdcli wrapper"
    scope: "internal/bead/bdcli.go"
    verify:
      - "tests pass"
    depends_on: []
  - id: 2
    title: "spec bead"
    scope: "internal/bead/spec.go"
    verify:
      - "creates bead"
    depends_on: [1]
---

# Plan
`
	planPath := filepath.Join(tmp, "plan.md")
	os.WriteFile(planPath, []byte(planContent), 0644)

	meta, err := ParsePlanMeta(planPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta.Status != "Approved" {
		t.Errorf("expected status Approved, got %q", meta.Status)
	}
	if len(meta.WorkChunks) != 2 {
		t.Fatalf("expected 2 work chunks, got %d", len(meta.WorkChunks))
	}
	if meta.WorkChunks[0].Title != "bdcli wrapper" {
		t.Errorf("expected first chunk title 'bdcli wrapper', got %q", meta.WorkChunks[0].Title)
	}
	if len(meta.WorkChunks[1].DependsOn) != 1 || meta.WorkChunks[1].DependsOn[0] != 1 {
		t.Errorf("expected chunk 2 depends_on [1], got %v", meta.WorkChunks[1].DependsOn)
	}
}

func TestParsePlanMeta_CommentedLines(t *testing.T) {
	tmp := t.TempDir()
	planContent := `---
status: Approved
spec_id: test
# This is a comment
# approved_at: not-yet
work_chunks:
  - id: 1
    title: "test"
    scope: "test.go"
    verify: []
    depends_on: []
---

body
`
	planPath := filepath.Join(tmp, "plan.md")
	os.WriteFile(planPath, []byte(planContent), 0644)

	meta, err := ParsePlanMeta(planPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.Status != "Approved" {
		t.Errorf("expected status Approved, got %q", meta.Status)
	}
}

func TestParsePlanMeta_NoFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	planPath := filepath.Join(tmp, "plan.md")
	os.WriteFile(planPath, []byte("# Plan\nNo frontmatter here\n"), 0644)

	_, err := ParsePlanMeta(planPath)
	if err == nil {
		t.Fatal("expected error for missing frontmatter")
	}
}

// --- CreatePlanBeads tests ---

func TestCreatePlanBeads_UnapprovedPlan(t *testing.T) {
	tmp := t.TempDir()
	specDir := filepath.Join(tmp, "docs", "specs", "test")
	os.MkdirAll(specDir, 0755)
	planContent := `---
status: Draft
spec_id: test
work_chunks:
  - id: 1
    title: "test"
    scope: "test.go"
    verify: []
    depends_on: []
---

# Plan
`
	os.WriteFile(filepath.Join(specDir, "plan.md"), []byte(planContent), 0644)

	_, err := CreatePlanBeads(tmp, "test")
	if err == nil {
		t.Fatal("expected error for unapproved plan")
	}
	if !contains(err.Error(), "not approved") {
		t.Errorf("error should mention not approved: %v", err)
	}
}

func TestCreatePlanBeads_MissingWorkChunks(t *testing.T) {
	tmp := t.TempDir()
	specDir := filepath.Join(tmp, "docs", "specs", "test")
	os.MkdirAll(specDir, 0755)
	planContent := `---
status: Approved
spec_id: test
---

# Plan
`
	os.WriteFile(filepath.Join(specDir, "plan.md"), []byte(planContent), 0644)

	_, err := CreatePlanBeads(tmp, "test")
	if err == nil {
		t.Fatal("expected error for missing work_chunks")
	}
	if !contains(err.Error(), "no work_chunks") {
		t.Errorf("error should mention work_chunks: %v", err)
	}
}

func TestCreatePlanBeads_CreatesAndWiresDeps(t *testing.T) {
	tmp := t.TempDir()
	specDir := filepath.Join(tmp, "docs", "specs", "test")
	os.MkdirAll(specDir, 0755)
	planContent := `---
status: Approved
spec_id: test
work_chunks:
  - id: 1
    title: "first chunk"
    scope: "first.go"
    verify:
      - "test passes"
    depends_on: []
  - id: 2
    title: "second chunk"
    scope: "second.go"
    verify:
      - "test passes"
    depends_on: [1]
---

# Plan
`
	os.WriteFile(filepath.Join(specDir, "plan.md"), []byte(planContent), 0644)

	origExec := execCommand
	defer func() { execCommand = origExec }()

	var depAddCalls [][]string
	createCount := 0

	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "bd" && len(args) > 0 {
			switch args[0] {
			case "search":
				return exec.Command("echo", `[]`) // no existing beads
			case "create":
				createCount++
				id := "bead-" + strings.Replace(args[1], " ", "", -1)[:10]
				return exec.Command("echo", `{"id":"`+id+`","title":"","description":"","status":"open","priority":2,"issue_type":"task","owner":"","created_at":"","updated_at":""}`)
			case "dep":
				if len(args) >= 4 {
					depAddCalls = append(depAddCalls, args[1:])
				}
				return exec.Command("echo", "")
			}
		}
		return exec.Command("echo", "")
	}

	mapping, err := CreatePlanBeads(tmp, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mapping) != 2 {
		t.Errorf("expected 2 mappings, got %d", len(mapping))
	}

	// At least one dep add call for chunk 2 -> chunk 1
	if len(depAddCalls) == 0 {
		t.Error("expected at least one dep add call")
	}
}

// --- buildImplDescription tests ---

func TestBuildImplDescription_Format(t *testing.T) {
	chunk := WorkChunk{
		ID:    1,
		Title: "test chunk",
		Scope: "internal/bead/bdcli.go",
		Verify: []string{
			"tests pass",
			"preflight works",
		},
	}

	desc := buildImplDescription(chunk, "007-beads-tooling")
	if !contains(desc, "Scope: internal/bead/bdcli.go") {
		t.Errorf("missing Scope line: %s", desc)
	}
	if !contains(desc, "Verify:") {
		t.Errorf("missing Verify section: %s", desc)
	}
	if !contains(desc, "- tests pass") {
		t.Errorf("missing verify item: %s", desc)
	}
	if !contains(desc, "Plan: docs/specs/007-beads-tooling/plan.md") {
		t.Errorf("missing Plan line: %s", desc)
	}
}

func TestBuildImplDescription_Cap(t *testing.T) {
	chunk := WorkChunk{
		Scope:  strings.Repeat("x", 900),
		Verify: []string{"test"},
	}
	desc := buildImplDescription(chunk, "test")
	if len(desc) > 800 {
		t.Errorf("description exceeds 800 char cap: %d chars", len(desc))
	}
}

// --- WriteGeneratedBeadIDs tests ---

func TestWriteGeneratedBeadIDs_PreservesFields(t *testing.T) {
	tmp := t.TempDir()
	planContent := `---
status: Approved
spec_id: test
version: "1.0"
work_chunks:
  - id: 1
    title: "chunk one"
    scope: "test.go"
    verify: []
    depends_on: []
---

# Plan body content

This should be preserved.
`
	planPath := filepath.Join(tmp, "plan.md")
	os.WriteFile(planPath, []byte(planContent), 0644)

	mapping := map[int]string{
		1: "bead-abc",
	}

	err := WriteGeneratedBeadIDs(planPath, mapping)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read back
	data, _ := os.ReadFile(planPath)
	content := string(data)

	// Body should be preserved
	if !contains(content, "# Plan body content") {
		t.Error("body content was lost")
	}
	if !contains(content, "This should be preserved.") {
		t.Error("body detail was lost")
	}

	// Frontmatter should still have original fields
	if !contains(content, "status: Approved") {
		t.Error("status field was lost")
	}
	if !contains(content, "spec_id: test") {
		t.Error("spec_id field was lost")
	}

	// Should have generated.bead_ids
	if !contains(content, "bead_ids") {
		t.Error("generated bead_ids not found")
	}
	if !contains(content, "bead-abc") {
		t.Error("bead ID value not found")
	}
}

func TestWriteGeneratedBeadIDs_NoFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	planPath := filepath.Join(tmp, "plan.md")
	os.WriteFile(planPath, []byte("# No frontmatter\n"), 0644)

	err := WriteGeneratedBeadIDs(planPath, map[int]string{1: "bead-x"})
	if err == nil {
		t.Fatal("expected error for missing frontmatter")
	}
}
