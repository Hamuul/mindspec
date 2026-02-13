package bead

import (
	"bytes"
	"encoding/json"
	"io"
	"os/exec"
	"strings"
	"testing"
)

// mockExec returns a fake execCommand that routes bd subcommands to handler functions.
// handlers maps "search"/"update"/"close" to functions that receive args and return stdout.
type mockHandler func(args []string) string

func mockExecRouted(handlers map[string]mockHandler) func(string, ...string) *exec.Cmd {
	return func(name string, args ...string) *exec.Cmd {
		if name != "bd" || len(args) == 0 {
			return exec.Command("echo", "")
		}
		sub := args[0]
		if h, ok := handlers[sub]; ok {
			return exec.Command("echo", h(args[1:]))
		}
		return exec.Command("echo", "[]")
	}
}

func beadJSON(id, title, status string) string {
	b := BeadInfo{ID: id, Title: title, Status: status}
	data, _ := json.Marshal([]BeadInfo{b})
	return string(data)
}

func savePropagateState(t *testing.T) {
	t.Helper()
	origExec := execCommand
	origStderr := stderr
	t.Cleanup(func() {
		execCommand = origExec
		stderr = origStderr
	})
}

func TestPropagateStart_UpdatesOpenParents(t *testing.T) {
	savePropagateState(t)
	stderr = io.Discard

	var updatedIDs []string
	execCommand = mockExecRouted(map[string]mockHandler{
		"search": func(args []string) string {
			query := args[0]
			if strings.Contains(query, "[PLAN") {
				return beadJSON("plan-1", "[PLAN 009-feat] Plan", "open")
			}
			if strings.Contains(query, "[SPEC") {
				return beadJSON("spec-1", "[SPEC 009-feat] Feature", "open")
			}
			return "[]"
		},
		"update": func(args []string) string {
			// args: [id, --status=in_progress]
			updatedIDs = append(updatedIDs, args[0])
			return ""
		},
	})

	PropagateStart("009-feat")

	if len(updatedIDs) != 2 {
		t.Fatalf("expected 2 updates, got %d: %v", len(updatedIDs), updatedIDs)
	}
	if updatedIDs[0] != "plan-1" {
		t.Errorf("first update: got %q, want %q", updatedIDs[0], "plan-1")
	}
	if updatedIDs[1] != "spec-1" {
		t.Errorf("second update: got %q, want %q", updatedIDs[1], "spec-1")
	}
}

func TestPropagateStart_SkipsNonOpen(t *testing.T) {
	savePropagateState(t)
	stderr = io.Discard

	var updatedIDs []string
	execCommand = mockExecRouted(map[string]mockHandler{
		"search": func(args []string) string {
			query := args[0]
			if strings.Contains(query, "[PLAN") {
				return beadJSON("plan-1", "[PLAN 009-feat] Plan", "in_progress")
			}
			if strings.Contains(query, "[SPEC") {
				return beadJSON("spec-1", "[SPEC 009-feat] Feature", "closed")
			}
			return "[]"
		},
		"update": func(args []string) string {
			updatedIDs = append(updatedIDs, args[0])
			return ""
		},
	})

	PropagateStart("009-feat")

	if len(updatedIDs) != 0 {
		t.Errorf("expected 0 updates for non-open parents, got %d: %v", len(updatedIDs), updatedIDs)
	}
}

func TestPropagateClose_ClosesWhenAllDone(t *testing.T) {
	savePropagateState(t)
	stderr = io.Discard

	var closedIDs []string
	execCommand = mockExecRouted(map[string]mockHandler{
		"search": func(args []string) string {
			query := args[0]
			// Check if this is a --status=open search (has that flag)
			hasStatusOpen := false
			for _, a := range args {
				if a == "--status=open" {
					hasStatusOpen = true
				}
			}

			if strings.Contains(query, "[IMPL") {
				if hasStatusOpen {
					return "[]" // no open impl beads
				}
				// SearchAny — all impl beads are closed
				return beadJSON("impl-1", "[IMPL 009-feat.1] Chunk", "closed")
			}
			if strings.Contains(query, "[PLAN") {
				return beadJSON("plan-1", "[PLAN 009-feat] Plan", "in_progress")
			}
			if strings.Contains(query, "[SPEC") {
				return beadJSON("spec-1", "[SPEC 009-feat] Feature", "in_progress")
			}
			return "[]"
		},
		"close": func(args []string) string {
			closedIDs = append(closedIDs, args[0])
			return ""
		},
	})

	PropagateClose("009-feat")

	if len(closedIDs) != 2 {
		t.Fatalf("expected 2 closes, got %d: %v", len(closedIDs), closedIDs)
	}
	if closedIDs[0] != "plan-1" {
		t.Errorf("first close: got %q, want %q", closedIDs[0], "plan-1")
	}
	if closedIDs[1] != "spec-1" {
		t.Errorf("second close: got %q, want %q", closedIDs[1], "spec-1")
	}
}

func TestPropagateClose_SkipsWhenImplsRemain(t *testing.T) {
	savePropagateState(t)
	stderr = io.Discard

	var closedIDs []string
	execCommand = mockExecRouted(map[string]mockHandler{
		"search": func(args []string) string {
			query := args[0]
			hasStatusOpen := false
			for _, a := range args {
				if a == "--status=open" {
					hasStatusOpen = true
				}
			}

			if strings.Contains(query, "[IMPL") {
				if hasStatusOpen {
					// One open impl bead remains
					return beadJSON("impl-2", "[IMPL 009-feat.2] Chunk 2", "open")
				}
				return "[]"
			}
			return "[]"
		},
		"close": func(args []string) string {
			closedIDs = append(closedIDs, args[0])
			return ""
		},
	})

	PropagateClose("009-feat")

	if len(closedIDs) != 0 {
		t.Errorf("expected 0 closes when impl beads remain open, got %d: %v", len(closedIDs), closedIDs)
	}
}

func TestPropagateClose_SkipsWhenImplsInProgress(t *testing.T) {
	savePropagateState(t)
	stderr = io.Discard

	var closedIDs []string
	execCommand = mockExecRouted(map[string]mockHandler{
		"search": func(args []string) string {
			query := args[0]
			hasStatusOpen := false
			for _, a := range args {
				if a == "--status=open" {
					hasStatusOpen = true
				}
			}

			if strings.Contains(query, "[IMPL") {
				if hasStatusOpen {
					return "[]" // no open
				}
				// SearchAny — one in_progress bead remains
				return beadJSON("impl-1", "[IMPL 009-feat.1] Chunk", "in_progress")
			}
			return "[]"
		},
		"close": func(args []string) string {
			closedIDs = append(closedIDs, args[0])
			return ""
		},
	})

	PropagateClose("009-feat")

	if len(closedIDs) != 0 {
		t.Errorf("expected 0 closes when impl beads are in_progress, got %d: %v", len(closedIDs), closedIDs)
	}
}

func TestPropagateStart_WarnsOnSearchError(t *testing.T) {
	savePropagateState(t)
	var buf bytes.Buffer
	stderr = &buf

	execCommand = func(name string, args ...string) *exec.Cmd {
		// Return a command that fails
		return exec.Command("false")
	}

	PropagateStart("009-feat") // should not panic

	if !strings.Contains(buf.String(), "Warning") {
		t.Errorf("expected warning on stderr, got %q", buf.String())
	}
}
