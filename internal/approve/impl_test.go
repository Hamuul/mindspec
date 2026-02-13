package approve

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mindspec/mindspec/internal/state"
)

func TestApproveImpl_HappyPath(t *testing.T) {
	tmp := t.TempDir()

	// Create spec dir and .mindspec dir
	os.MkdirAll(filepath.Join(tmp, "docs", "specs", "010-test"), 0755)
	os.MkdirAll(filepath.Join(tmp, ".mindspec"), 0755)

	// Set state to review mode
	state.Write(tmp, &state.State{
		Mode:       state.ModeReview,
		ActiveSpec: "010-test",
	})

	result, err := ApproveImpl(tmp, "010-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SpecID != "010-test" {
		t.Errorf("SpecID: got %q, want %q", result.SpecID, "010-test")
	}

	// Verify state is now idle
	s, err := state.Read(tmp)
	if err != nil {
		t.Fatalf("reading state: %v", err)
	}
	if s.Mode != state.ModeIdle {
		t.Errorf("mode: got %q, want %q", s.Mode, state.ModeIdle)
	}
}

func TestApproveImpl_WrongMode(t *testing.T) {
	tmp := t.TempDir()

	os.MkdirAll(filepath.Join(tmp, "docs", "specs", "010-test"), 0755)
	os.MkdirAll(filepath.Join(tmp, ".mindspec"), 0755)

	state.Write(tmp, &state.State{
		Mode:       state.ModeImplement,
		ActiveSpec: "010-test",
		ActiveBead: "bead-1",
	})

	_, err := ApproveImpl(tmp, "010-test")
	if err == nil {
		t.Fatal("expected error for wrong mode")
	}
	if !strings.Contains(err.Error(), "expected review mode") {
		t.Errorf("error should mention expected review mode: %v", err)
	}
}

func TestApproveImpl_WrongSpec(t *testing.T) {
	tmp := t.TempDir()

	os.MkdirAll(filepath.Join(tmp, "docs", "specs", "010-test"), 0755)
	os.MkdirAll(filepath.Join(tmp, ".mindspec"), 0755)

	state.Write(tmp, &state.State{
		Mode:       state.ModeReview,
		ActiveSpec: "010-test",
	})

	_, err := ApproveImpl(tmp, "011-other")
	if err == nil {
		t.Fatal("expected error for wrong spec")
	}
	if !strings.Contains(err.Error(), "active spec") {
		t.Errorf("error should mention active spec mismatch: %v", err)
	}
}
