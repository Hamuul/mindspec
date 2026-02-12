package next

import (
	"os"
	"path/filepath"
	"testing"
)

// --- ParseBeadsJSON tests ---

func TestParseBeadsJSON_SingleItem(t *testing.T) {
	input := `[{
		"id": "mindspec-25p",
		"title": "Test bead for parsing",
		"status": "open",
		"priority": 4,
		"issue_type": "task",
		"owner": "max@enubiq.com",
		"created_at": "2026-02-12T08:50:30Z",
		"updated_at": "2026-02-12T08:50:30Z"
	}]`

	items, err := ParseBeadsJSON([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ID != "mindspec-25p" {
		t.Errorf("expected ID mindspec-25p, got %s", items[0].ID)
	}
	if items[0].IssueType != "task" {
		t.Errorf("expected issue_type task, got %s", items[0].IssueType)
	}
	if items[0].Priority != 4 {
		t.Errorf("expected priority 4, got %d", items[0].Priority)
	}
}

func TestParseBeadsJSON_MultipleItems(t *testing.T) {
	input := `[
		{"id": "a", "title": "First", "status": "open", "priority": 1, "issue_type": "task", "owner": "", "created_at": "", "updated_at": ""},
		{"id": "b", "title": "Second", "status": "open", "priority": 2, "issue_type": "feature", "owner": "", "created_at": "", "updated_at": ""}
	]`

	items, err := ParseBeadsJSON([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[1].IssueType != "feature" {
		t.Errorf("expected second item type feature, got %s", items[1].IssueType)
	}
}

func TestParseBeadsJSON_EmptyArray(t *testing.T) {
	items, err := ParseBeadsJSON([]byte("[]"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestParseBeadsJSON_InvalidJSON(t *testing.T) {
	_, err := ParseBeadsJSON([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

// --- SelectWork tests ---

func TestSelectWork_SingleItem(t *testing.T) {
	items := []BeadInfo{{ID: "a", Title: "Only one"}}
	result, err := SelectWork(items, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "a" {
		t.Errorf("expected ID a, got %s", result.ID)
	}
}

func TestSelectWork_MultipleDefaultsToFirst(t *testing.T) {
	items := []BeadInfo{
		{ID: "a", Title: "First"},
		{ID: "b", Title: "Second"},
	}
	result, err := SelectWork(items, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "a" {
		t.Errorf("expected ID a, got %s", result.ID)
	}
}

func TestSelectWork_PickSpecific(t *testing.T) {
	items := []BeadInfo{
		{ID: "a", Title: "First"},
		{ID: "b", Title: "Second"},
		{ID: "c", Title: "Third"},
	}
	result, err := SelectWork(items, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "b" {
		t.Errorf("expected ID b, got %s", result.ID)
	}
}

func TestSelectWork_PickOutOfRange(t *testing.T) {
	items := []BeadInfo{
		{ID: "a", Title: "First"},
		{ID: "b", Title: "Second"},
	}
	_, err := SelectWork(items, 5)
	if err == nil {
		t.Fatal("expected error for out of range pick")
	}
}

func TestSelectWork_EmptyList(t *testing.T) {
	_, err := SelectWork([]BeadInfo{}, 0)
	if err == nil {
		t.Fatal("expected error for empty list")
	}
}

func TestFormatWorkList(t *testing.T) {
	items := []BeadInfo{
		{ID: "abc", Title: "Do something", Priority: 2, IssueType: "task"},
		{ID: "def", Title: "Plan feature", Priority: 1, IssueType: "feature"},
	}
	result := FormatWorkList(items)
	if result == "" {
		t.Fatal("expected non-empty format output")
	}
	// Check that both items appear
	if !contains(result, "abc") || !contains(result, "def") {
		t.Errorf("format output missing item IDs: %s", result)
	}
	if !contains(result, "1.") || !contains(result, "2.") {
		t.Errorf("format output missing numbering: %s", result)
	}
}

// --- ResolveMode tests ---

func TestResolveMode_Task(t *testing.T) {
	bead := BeadInfo{ID: "x", Title: "005-next: Implement something", IssueType: "task"}
	result := ResolveMode("/nonexistent", bead)
	if result.Mode != "implement" {
		t.Errorf("expected implement, got %s", result.Mode)
	}
	if result.SpecID != "005-next" {
		t.Errorf("expected spec ID 005-next, got %s", result.SpecID)
	}
}

func TestResolveMode_Bug(t *testing.T) {
	bead := BeadInfo{ID: "x", Title: "003-context: Fix rendering", IssueType: "bug"}
	result := ResolveMode("/nonexistent", bead)
	if result.Mode != "implement" {
		t.Errorf("expected implement, got %s", result.Mode)
	}
	if result.SpecID != "003-context" {
		t.Errorf("expected spec ID 003-context, got %s", result.SpecID)
	}
}

func TestResolveMode_Feature_NoSpec(t *testing.T) {
	bead := BeadInfo{ID: "x", Title: "099-future: New feature", IssueType: "feature"}
	result := ResolveMode("/nonexistent", bead)
	if result.Mode != "spec" {
		t.Errorf("expected spec, got %s", result.Mode)
	}
}

func TestResolveMode_Feature_ApprovedSpec(t *testing.T) {
	// Set up a temp dir with an approved spec
	tmp := t.TempDir()
	specDir := filepath.Join(tmp, "docs", "specs", "010-test")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}
	specContent := "# Spec\n\n## Approval\n\n- **Status**: APPROVED\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0644); err != nil {
		t.Fatal(err)
	}

	bead := BeadInfo{ID: "x", Title: "010-test: Plan a feature", IssueType: "feature"}
	result := ResolveMode(tmp, bead)
	if result.Mode != "plan" {
		t.Errorf("expected plan, got %s", result.Mode)
	}
}

func TestResolveMode_Feature_DraftSpec(t *testing.T) {
	// Set up a temp dir with a draft spec (no APPROVED status)
	tmp := t.TempDir()
	specDir := filepath.Join(tmp, "docs", "specs", "010-test")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}
	specContent := "# Spec\n\n## Approval\n\n- **Status**: DRAFT\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specContent), 0644); err != nil {
		t.Fatal(err)
	}

	bead := BeadInfo{ID: "x", Title: "010-test: Draft feature", IssueType: "feature"}
	result := ResolveMode(tmp, bead)
	if result.Mode != "spec" {
		t.Errorf("expected spec, got %s", result.Mode)
	}
}

func TestResolveMode_NoColonInTitle(t *testing.T) {
	bead := BeadInfo{ID: "x", Title: "No colon here", IssueType: "task"}
	result := ResolveMode("/nonexistent", bead)
	if result.Mode != "implement" {
		t.Errorf("expected implement, got %s", result.Mode)
	}
	if result.SpecID != "" {
		t.Errorf("expected empty spec ID, got %s", result.SpecID)
	}
}

// --- parseSpecID tests ---

func TestParseSpecID(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		{"005-next: Implement work selection", "005-next"},
		{"003-context: Fix rendering bug", "003-context"},
		{"No colon here", ""},
		{"simple:", "simple"},
		{": leading colon", ""},
	}
	for _, tt := range tests {
		result := parseSpecID(tt.title)
		if result != tt.expected {
			t.Errorf("parseSpecID(%q) = %q, want %q", tt.title, result, tt.expected)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
