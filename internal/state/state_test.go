package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// --- Session tests ---

func TestSessionFile_RoundTrip(t *testing.T) {
	tmp := t.TempDir()

	s := &Session{
		SessionSource:    "startup",
		SessionStartedAt: "2026-02-27T00:00:00Z",
		BeadClaimedAt:    "2026-02-27T00:01:00Z",
	}
	if err := WriteSessionFile(tmp, s); err != nil {
		t.Fatalf("WriteSessionFile failed: %v", err)
	}

	got, err := ReadSession(tmp)
	if err != nil {
		t.Fatalf("ReadSession failed: %v", err)
	}
	if got.SessionSource != "startup" {
		t.Errorf("sessionSource: got %q, want %q", got.SessionSource, "startup")
	}
	if got.SessionStartedAt != "2026-02-27T00:00:00Z" {
		t.Errorf("sessionStartedAt: got %q, want %q", got.SessionStartedAt, "2026-02-27T00:00:00Z")
	}
	if got.BeadClaimedAt != "2026-02-27T00:01:00Z" {
		t.Errorf("beadClaimedAt: got %q, want %q", got.BeadClaimedAt, "2026-02-27T00:01:00Z")
	}
}

func TestSessionFile_MissingReturnsZero(t *testing.T) {
	tmp := t.TempDir()

	got, err := ReadSession(tmp)
	if err != nil {
		t.Fatalf("ReadSession failed: %v", err)
	}
	if got.SessionSource != "" {
		t.Errorf("expected empty sessionSource, got %q", got.SessionSource)
	}
}

func TestSessionFile_OmitsEmptyFields(t *testing.T) {
	tmp := t.TempDir()

	s := &Session{SessionSource: "clear"}
	if err := WriteSessionFile(tmp, s); err != nil {
		t.Fatalf("WriteSessionFile failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmp, ".mindspec", "session.json"))
	if err != nil {
		t.Fatalf("reading session.json: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if _, ok := parsed["beadClaimedAt"]; ok {
		t.Error("expected beadClaimedAt to be omitted when empty")
	}
}

// --- Convention function tests ---

func TestSpecBranch(t *testing.T) {
	tests := []struct {
		specID string
		want   string
	}{
		{"053-drop-state-json", "spec/053-drop-state-json"},
		{"001-skeleton", "spec/001-skeleton"},
	}
	for _, tt := range tests {
		got := SpecBranch(tt.specID)
		if got != tt.want {
			t.Errorf("SpecBranch(%q) = %q, want %q", tt.specID, got, tt.want)
		}
	}
}

func TestSpecWorktreePath(t *testing.T) {
	got := SpecWorktreePath("/project", "053-foo")
	want := filepath.Join("/project", ".worktrees", "worktree-spec-053-foo")
	if got != want {
		t.Errorf("SpecWorktreePath = %q, want %q", got, want)
	}
}

func TestBeadWorktreePath(t *testing.T) {
	specWT := filepath.Join("/project", ".worktrees", "worktree-spec-053-foo")
	got := BeadWorktreePath(specWT, "mindspec-mol-07lst")
	want := filepath.Join(specWT, ".worktrees", "worktree-mindspec-mol-07lst")
	if got != want {
		t.Errorf("BeadWorktreePath = %q, want %q", got, want)
	}
}

func TestIsValidMode(t *testing.T) {
	for _, m := range ValidModes {
		if !IsValidMode(m) {
			t.Errorf("IsValidMode(%q) = false, want true", m)
		}
	}
	if IsValidMode("invalid") {
		t.Error("IsValidMode(\"invalid\") = true, want false")
	}
}
