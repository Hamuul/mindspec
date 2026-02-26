package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallPreCommit_NewHook(t *testing.T) {
	root := t.TempDir()
	hooksDir := filepath.Join(root, ".git", "hooks")
	os.MkdirAll(hooksDir, 0755)

	if err := InstallPreCommit(root); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(hooksDir, "pre-commit"))
	if err != nil {
		t.Fatalf("reading hook: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "MindSpec pre-commit hook") {
		t.Error("hook should contain MindSpec marker")
	}
	if !strings.Contains(content, "MINDSPEC_ALLOW_MAIN") {
		t.Error("hook should contain escape hatch")
	}
}

func TestInstallPreCommit_Idempotent(t *testing.T) {
	root := t.TempDir()
	hooksDir := filepath.Join(root, ".git", "hooks")
	os.MkdirAll(hooksDir, 0755)

	// Install twice
	if err := InstallPreCommit(root); err != nil {
		t.Fatal(err)
	}
	if err := InstallPreCommit(root); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(hooksDir, "pre-commit"))
	count := strings.Count(string(data), "MindSpec pre-commit hook")
	if count != 1 {
		t.Errorf("expected exactly 1 marker, got %d", count)
	}
}

func TestInstallPreCommit_ChainsExisting(t *testing.T) {
	root := t.TempDir()
	hooksDir := filepath.Join(root, ".git", "hooks")
	os.MkdirAll(hooksDir, 0755)

	// Write existing hook
	existing := "#!/bin/bash\necho 'existing hook'\n"
	os.WriteFile(filepath.Join(hooksDir, "pre-commit"), []byte(existing), 0755)

	if err := InstallPreCommit(root); err != nil {
		t.Fatal(err)
	}

	// Check backup exists
	backup, err := os.ReadFile(filepath.Join(hooksDir, "pre-commit.pre-mindspec"))
	if err != nil {
		t.Fatal("backup not created")
	}
	if string(backup) != existing {
		t.Error("backup content doesn't match original")
	}

	// Check new hook chains
	data, _ := os.ReadFile(filepath.Join(hooksDir, "pre-commit"))
	content := string(data)
	if !strings.Contains(content, "MindSpec pre-commit hook") {
		t.Error("new hook should contain MindSpec marker")
	}
	if !strings.Contains(content, "pre-commit.pre-mindspec") {
		t.Error("new hook should chain to backup")
	}
}

func TestInstallPreCommit_NoGitDir(t *testing.T) {
	root := t.TempDir()
	// No .git/hooks — should skip silently
	if err := InstallPreCommit(root); err != nil {
		t.Errorf("expected nil error for non-git dir, got: %v", err)
	}
}
