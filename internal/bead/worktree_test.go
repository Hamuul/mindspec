package bead

import (
	"os/exec"
	"testing"
)

// --- parseWorktreePorcelain tests ---

func TestParseWorktreePorcelain_MultiEntry(t *testing.T) {
	output := `worktree /home/user/project
HEAD abc123
branch refs/heads/main

worktree /home/user/worktree-bead-xyz
HEAD def456
branch refs/heads/bead/bead-xyz

`
	entries := parseWorktreePorcelain(output)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	if entries[0].Path != "/home/user/project" {
		t.Errorf("entry 0 path: got %q", entries[0].Path)
	}
	if entries[0].Branch != "main" {
		t.Errorf("entry 0 branch: got %q", entries[0].Branch)
	}
	if entries[0].HEAD != "abc123" {
		t.Errorf("entry 0 HEAD: got %q", entries[0].HEAD)
	}

	if entries[1].Path != "/home/user/worktree-bead-xyz" {
		t.Errorf("entry 1 path: got %q", entries[1].Path)
	}
	if entries[1].Branch != "bead/bead-xyz" {
		t.Errorf("entry 1 branch: got %q", entries[1].Branch)
	}
}

func TestParseWorktreePorcelain_BareRepo(t *testing.T) {
	// Bare repos have a "bare" line instead of HEAD/branch
	output := `worktree /home/user/repo.git
bare

worktree /home/user/worktree-bead-abc
HEAD aaa111
branch refs/heads/bead/bead-abc

`
	entries := parseWorktreePorcelain(output)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	// Bare entry should still parse (just empty HEAD/branch)
	if entries[0].Path != "/home/user/repo.git" {
		t.Errorf("bare entry path: got %q", entries[0].Path)
	}
}

func TestParseWorktreePorcelain_Empty(t *testing.T) {
	entries := parseWorktreePorcelain("")
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestParseWorktreePorcelain_NoTrailingNewline(t *testing.T) {
	output := `worktree /home/user/project
HEAD abc123
branch refs/heads/main`

	entries := parseWorktreePorcelain(output)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Branch != "main" {
		t.Errorf("branch: got %q", entries[0].Branch)
	}
}

// --- FindWorktree matching tests ---

func TestFindWorktree_MatchByPath(t *testing.T) {
	origExec := execCommand
	defer func() { execCommand = origExec }()

	execCommand = func(name string, args ...string) *exec.Cmd {
		porcelain := "worktree /home/user/project\nHEAD abc\nbranch refs/heads/main\n\nworktree /home/user/worktree-bead-xyz\nHEAD def\nbranch refs/heads/feature\n\n"
		return exec.Command("echo", "-n", porcelain)
	}

	path, err := FindWorktree("bead-xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "/home/user/worktree-bead-xyz" {
		t.Errorf("expected worktree path, got %q", path)
	}
}

func TestFindWorktree_MatchByBranch(t *testing.T) {
	origExec := execCommand
	defer func() { execCommand = origExec }()

	execCommand = func(name string, args ...string) *exec.Cmd {
		porcelain := "worktree /home/user/project\nHEAD abc\nbranch refs/heads/main\n\nworktree /home/user/other-dir\nHEAD def\nbranch refs/heads/bead/bead-xyz\n\n"
		return exec.Command("echo", "-n", porcelain)
	}

	path, err := FindWorktree("bead-xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "/home/user/other-dir" {
		t.Errorf("expected worktree path, got %q", path)
	}
}

func TestFindWorktree_NotFound(t *testing.T) {
	origExec := execCommand
	defer func() { execCommand = origExec }()

	execCommand = func(name string, args ...string) *exec.Cmd {
		porcelain := "worktree /home/user/project\nHEAD abc\nbranch refs/heads/main\n\n"
		return exec.Command("echo", "-n", porcelain)
	}

	path, err := FindWorktree("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "" {
		t.Errorf("expected empty path for not found, got %q", path)
	}
}

// --- CreateWorktree tests ---

func TestCreateWorktree_RefusesDirtyTree(t *testing.T) {
	origExec := execCommand
	defer func() { execCommand = origExec }()

	callNum := 0
	execCommand = func(name string, args ...string) *exec.Cmd {
		callNum++
		if name == "bd" && len(args) > 0 && args[0] == "show" {
			return exec.Command("echo", `{"id":"bead-xyz","title":"test","description":"","status":"in_progress","priority":2,"issue_type":"task","owner":"","created_at":"","updated_at":""}`)
		}
		if name == "git" && len(args) > 0 && args[0] == "status" {
			// Return dirty tree
			return exec.Command("echo", "M  dirty-file.go")
		}
		return exec.Command("echo", "")
	}

	_, err := CreateWorktree("/tmp/test", "bead-xyz")
	if err == nil {
		t.Fatal("expected error for dirty tree")
	}
	if !contains(err.Error(), "dirty") {
		t.Errorf("error should mention dirty tree: %v", err)
	}
}

func TestCreateWorktree_RefusesNonInProgress(t *testing.T) {
	origExec := execCommand
	defer func() { execCommand = origExec }()

	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "bd" && len(args) > 0 && args[0] == "show" {
			return exec.Command("echo", `{"id":"bead-xyz","title":"test","description":"","status":"open","priority":2,"issue_type":"task","owner":"","created_at":"","updated_at":""}`)
		}
		return exec.Command("echo", "")
	}

	_, err := CreateWorktree("/tmp/test", "bead-xyz")
	if err == nil {
		t.Fatal("expected error for non-in_progress bead")
	}
	if !contains(err.Error(), "not in_progress") {
		t.Errorf("error should mention not in_progress: %v", err)
	}
}
