package gitops

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureGitignoreEntry_New(t *testing.T) {
	root := t.TempDir()

	if err := EnsureGitignoreEntry(root, ".worktrees"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		t.Fatalf("reading .gitignore: %v", err)
	}

	if !strings.Contains(string(data), ".worktrees/") {
		t.Errorf("expected .worktrees/ in .gitignore, got: %s", data)
	}
}

func TestEnsureGitignoreEntry_Idempotent(t *testing.T) {
	root := t.TempDir()

	// Write twice
	if err := EnsureGitignoreEntry(root, ".worktrees"); err != nil {
		t.Fatal(err)
	}
	if err := EnsureGitignoreEntry(root, ".worktrees"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}

	count := strings.Count(string(data), ".worktrees/")
	if count != 1 {
		t.Errorf("expected exactly 1 .worktrees/ entry, got %d in:\n%s", count, data)
	}
}

// initGitRepo creates a git repo with an initial commit and returns the path.
func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %s", args, out)
		}
	}
	run("init", "-b", "main")
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Test\n"), 0644)
	run("add", ".")
	run("commit", "-m", "initial")
	return dir
}

func TestDiffStat(t *testing.T) {
	dir := initGitRepo(t)
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %s", args, out)
		}
	}

	// Create a feature branch with changes
	run("checkout", "-b", "feature")
	os.WriteFile(filepath.Join(dir, "new.txt"), []byte("hello\n"), 0644)
	run("add", ".")
	run("commit", "-m", "add new file")

	stat, err := DiffStat(dir, "main", "feature")
	if err != nil {
		t.Fatalf("DiffStat: %v", err)
	}
	if !strings.Contains(stat, "new.txt") {
		t.Errorf("expected new.txt in diffstat, got: %s", stat)
	}
}

func TestCommitCount(t *testing.T) {
	dir := initGitRepo(t)
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %s", args, out)
		}
	}

	run("checkout", "-b", "feature")
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a\n"), 0644)
	run("add", ".")
	run("commit", "-m", "commit 1")
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b\n"), 0644)
	run("add", ".")
	run("commit", "-m", "commit 2")

	count, err := CommitCount(dir, "main", "feature")
	if err != nil {
		t.Fatalf("CommitCount: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 commits, got %d", count)
	}
}

func TestEnsureGitignoreEntry_ExistingFile(t *testing.T) {
	root := t.TempDir()
	gitignorePath := filepath.Join(root, ".gitignore")

	// Pre-existing content without trailing newline
	if err := os.WriteFile(gitignorePath, []byte("node_modules/"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := EnsureGitignoreEntry(root, ".worktrees"); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "node_modules/") {
		t.Error("existing content should be preserved")
	}
	if !strings.Contains(content, ".worktrees/") {
		t.Error("new entry should be appended")
	}
}
