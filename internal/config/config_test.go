package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if len(cfg.ProtectedBranches) != 2 {
		t.Fatalf("expected 2 protected branches, got %d", len(cfg.ProtectedBranches))
	}
	if cfg.ProtectedBranches[0] != "main" || cfg.ProtectedBranches[1] != "master" {
		t.Errorf("unexpected protected branches: %v", cfg.ProtectedBranches)
	}
	if cfg.MergeStrategy != "auto" {
		t.Errorf("expected merge_strategy=auto, got %q", cfg.MergeStrategy)
	}
	if cfg.WorktreeRoot != ".worktrees" {
		t.Errorf("expected worktree_root=.worktrees, got %q", cfg.WorktreeRoot)
	}
	if !cfg.Enforcement.PreCommitHook {
		t.Error("expected pre_commit_hook=true")
	}
	if !cfg.Enforcement.CLIGuards {
		t.Error("expected cli_guards=true")
	}
	if !cfg.Enforcement.AgentHooks {
		t.Error("expected agent_hooks=true")
	}
	if cfg.Recording.Enabled {
		t.Error("expected recording.enabled=false by default")
	}
}

func TestLoadMissing(t *testing.T) {
	cfg, err := Load(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MergeStrategy != "auto" {
		t.Errorf("expected defaults when file missing, got merge_strategy=%q", cfg.MergeStrategy)
	}
}

func TestLoadFromFile(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".mindspec")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}

	content := `
protected_branches:
  - main
  - develop
merge_strategy: pr
worktree_root: .wt
enforcement:
  pre_commit_hook: true
  cli_guards: false
  agent_hooks: true
`
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.ProtectedBranches) != 2 || cfg.ProtectedBranches[1] != "develop" {
		t.Errorf("unexpected protected branches: %v", cfg.ProtectedBranches)
	}
	if cfg.MergeStrategy != "pr" {
		t.Errorf("expected merge_strategy=pr, got %q", cfg.MergeStrategy)
	}
	if cfg.WorktreeRoot != ".wt" {
		t.Errorf("expected worktree_root=.wt, got %q", cfg.WorktreeRoot)
	}
	if cfg.Enforcement.CLIGuards {
		t.Error("expected cli_guards=false")
	}
}

func TestRecordingEnabled(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".mindspec")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}

	content := `
recording:
  enabled: true
`
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Recording.Enabled {
		t.Error("expected recording.enabled=true")
	}
}

func TestIsProtectedBranch(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.IsProtectedBranch("main") {
		t.Error("main should be protected")
	}
	if !cfg.IsProtectedBranch("master") {
		t.Error("master should be protected")
	}
	if cfg.IsProtectedBranch("feature/foo") {
		t.Error("feature/foo should not be protected")
	}
}

func TestWorktreePath(t *testing.T) {
	cfg := DefaultConfig()
	got := cfg.WorktreePath("/repo", "worktree-spec-046")
	want := filepath.Join("/repo", ".worktrees", "worktree-spec-046")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
