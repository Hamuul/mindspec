package next

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/mindspec/mindspec/internal/bead"
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
		// Bracket-prefix convention
		{"[IMPL 009-feature.1] Chunk title", "009-feature"},
		{"[IMPL 009-workflow-gaps.2] Approval enhancements", "009-workflow-gaps"},
		{"[SPEC 008b-gates] Human Gates Feature", "008b-gates"},
		{"[PLAN 009-feature] Plan decomposition", "009-feature"},
		{"[IMPL 001.3] Simple numeric", "001"},
		// Legacy colon convention (fallback)
		{"005-next: Implement work selection", "005-next"},
		{"003-context: Fix rendering bug", "003-context"},
		// Edge cases
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

// --- QueryReady molecule-aware tests ---

func TestQueryReady_PrefersMolChildren(t *testing.T) {
	origSearch := searchBeads
	origMolReady := molReady
	origExec := execCommand
	defer func() {
		searchBeads = origSearch
		molReady = origMolReady
		execCommand = origExec
	}()

	searchBeads = func(query string) ([]bead.BeadInfo, error) {
		// Return a molecule parent
		return []bead.BeadInfo{
			{ID: "mol-parent-1", Title: "[PLAN test] Plan", IssueType: "epic"},
		}, nil
	}

	molReady = func(parentID string) ([]bead.BeadInfo, error) {
		// Return ready children
		return []bead.BeadInfo{
			{ID: "child-1", Title: "[IMPL test.1] First chunk", IssueType: "task", Priority: 2},
			{ID: "child-2", Title: "[IMPL test.2] Second chunk", IssueType: "task", Priority: 2},
		}, nil
	}

	// execCommand should NOT be called (mol ready should be preferred)
	execCommand = func(name string, args ...string) *exec.Cmd {
		t.Error("bd ready should not be called when mol children exist")
		return exec.Command("echo", "[]")
	}

	items, err := QueryReady()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ID != "child-1" {
		t.Errorf("items[0].ID: got %q, want %q", items[0].ID, "child-1")
	}
}

func TestQueryReady_FallsBackToBdReady(t *testing.T) {
	origSearch := searchBeads
	origMolReady := molReady
	origExec := execCommand
	defer func() {
		searchBeads = origSearch
		molReady = origMolReady
		execCommand = origExec
	}()

	// No molecule parents
	searchBeads = func(query string) ([]bead.BeadInfo, error) {
		return nil, nil
	}

	molReady = func(parentID string) ([]bead.BeadInfo, error) {
		t.Error("MolReady should not be called when no parents found")
		return nil, nil
	}

	// bd ready fallback
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("echo", `[{"id":"standalone-1","title":"Standalone work","status":"open","priority":2,"issue_type":"task","owner":"","created_at":"","updated_at":""}]`)
	}

	items, err := QueryReady()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ID != "standalone-1" {
		t.Errorf("items[0].ID: got %q, want %q", items[0].ID, "standalone-1")
	}
}

func TestQueryReady_FallsBackWhenMolEmpty(t *testing.T) {
	origSearch := searchBeads
	origMolReady := molReady
	origExec := execCommand
	defer func() {
		searchBeads = origSearch
		molReady = origMolReady
		execCommand = origExec
	}()

	// Molecule parent exists but no ready children
	searchBeads = func(query string) ([]bead.BeadInfo, error) {
		return []bead.BeadInfo{
			{ID: "mol-parent-1", Title: "[PLAN test] Plan", IssueType: "epic"},
		}, nil
	}

	molReady = func(parentID string) ([]bead.BeadInfo, error) {
		return nil, nil // empty
	}

	// Should fall back to bd ready
	bdReadyCalled := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		bdReadyCalled = true
		return exec.Command("echo", `[{"id":"fallback-1","title":"Fallback work","status":"open","priority":2,"issue_type":"task","owner":"","created_at":"","updated_at":""}]`)
	}

	items, err := QueryReady()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bdReadyCalled {
		t.Error("expected bd ready fallback to be called")
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ID != "fallback-1" {
		t.Errorf("items[0].ID: got %q, want %q", items[0].ID, "fallback-1")
	}
}

// --- ClaimBead tests ---

func TestClaimBead_DelegatesToBeadUpdate(t *testing.T) {
	origUpdate := updateBead
	defer func() { updateBead = origUpdate }()

	var capturedID, capturedStatus string
	updateBead = func(id, status string) error {
		capturedID = id
		capturedStatus = status
		return nil
	}

	err := ClaimBead("bead-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "bead-abc" {
		t.Errorf("ID: got %q, want %q", capturedID, "bead-abc")
	}
	if capturedStatus != "in_progress" {
		t.Errorf("status: got %q, want %q", capturedStatus, "in_progress")
	}
}

func TestClaimBead_PropagatesError(t *testing.T) {
	origUpdate := updateBead
	defer func() { updateBead = origUpdate }()

	updateBead = func(id, status string) error {
		return fmt.Errorf("bd update failed")
	}

	err := ClaimBead("bead-abc")
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- EnsureWorktree tests ---

func TestEnsureWorktree_CreatesNew(t *testing.T) {
	origList := worktreeList
	origCreate := worktreeCreate
	defer func() {
		worktreeList = origList
		worktreeCreate = origCreate
	}()

	listCallCount := 0
	worktreeList = func() ([]bead.WorktreeListEntry, error) {
		listCallCount++
		if listCallCount == 1 {
			// First call: no matching worktree
			return []bead.WorktreeListEntry{
				{Name: "mindspec", Path: "/home/user/mindspec", Branch: "main", IsMain: true},
			}, nil
		}
		// Second call: worktree was created
		return []bead.WorktreeListEntry{
			{Name: "mindspec", Path: "/home/user/mindspec", Branch: "main", IsMain: true},
			{Name: "worktree-bead-abc", Path: "/home/user/worktree-bead-abc", Branch: "bead/bead-abc", IsMain: false},
		}, nil
	}

	var createdName, createdBranch string
	worktreeCreate = func(name, branch string) error {
		createdName = name
		createdBranch = branch
		return nil
	}

	path, err := EnsureWorktree("bead-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "/home/user/worktree-bead-abc" {
		t.Errorf("path: got %q, want %q", path, "/home/user/worktree-bead-abc")
	}
	if createdName != "worktree-bead-abc" {
		t.Errorf("created name: got %q, want %q", createdName, "worktree-bead-abc")
	}
	if createdBranch != "bead/bead-abc" {
		t.Errorf("created branch: got %q, want %q", createdBranch, "bead/bead-abc")
	}
}

func TestEnsureWorktree_ReusesExisting(t *testing.T) {
	origList := worktreeList
	origCreate := worktreeCreate
	defer func() {
		worktreeList = origList
		worktreeCreate = origCreate
	}()

	worktreeList = func() ([]bead.WorktreeListEntry, error) {
		return []bead.WorktreeListEntry{
			{Name: "mindspec", Path: "/home/user/mindspec", Branch: "main", IsMain: true},
			{Name: "worktree-bead-abc", Path: "/home/user/worktree-bead-abc", Branch: "bead/bead-abc", IsMain: false},
		}, nil
	}

	worktreeCreate = func(name, branch string) error {
		t.Error("worktreeCreate should not be called when worktree exists")
		return nil
	}

	path, err := EnsureWorktree("bead-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "/home/user/worktree-bead-abc" {
		t.Errorf("path: got %q, want %q", path, "/home/user/worktree-bead-abc")
	}
}

func TestEnsureWorktree_MatchesByBranch(t *testing.T) {
	origList := worktreeList
	origCreate := worktreeCreate
	defer func() {
		worktreeList = origList
		worktreeCreate = origCreate
	}()

	worktreeList = func() ([]bead.WorktreeListEntry, error) {
		return []bead.WorktreeListEntry{
			{Name: "custom-name", Path: "/home/user/custom-name", Branch: "bead/bead-xyz", IsMain: false},
		}, nil
	}

	worktreeCreate = func(name, branch string) error {
		t.Error("should not create — matched by branch")
		return nil
	}

	path, err := EnsureWorktree("bead-xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "/home/user/custom-name" {
		t.Errorf("path: got %q, want %q", path, "/home/user/custom-name")
	}
}

// --- convertBeadInfos tests ---

func TestConvertBeadInfos(t *testing.T) {
	src := []bead.BeadInfo{
		{ID: "a", Title: "First", Status: "open", Priority: 1, IssueType: "task", Owner: "user", CreatedAt: "t1", UpdatedAt: "t2"},
	}
	result := convertBeadInfos(src)
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	if result[0].ID != "a" || result[0].Title != "First" || result[0].Priority != 1 {
		t.Errorf("conversion mismatch: %+v", result[0])
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
