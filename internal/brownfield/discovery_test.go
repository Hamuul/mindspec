package brownfield

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDiscoverMarkdown_DeterministicAndFiltered(t *testing.T) {
	root := t.TempDir()

	mk := func(rel string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(p, []byte("# test\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}

	mk("README.md")
	mk("docs/a.md")
	mk("docs/z.MD")
	mk("notes/todo.txt")
	mk(".git/ignored.md")
	mk(".beads/internal.md")

	got, err := DiscoverMarkdown(root)
	if err != nil {
		t.Fatalf("DiscoverMarkdown: %v", err)
	}

	want := []string{
		"README.md",
		"docs/a.md",
		"docs/z.MD",
	}
	if !reflect.DeepEqual(got.MarkdownFiles, want) {
		t.Fatalf("markdown files mismatch\ngot:  %#v\nwant: %#v", got.MarkdownFiles, want)
	}
}

func TestRun_ReportArtifactsAreDeterministic(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "docs", "adr"), 0o755); err != nil {
		t.Fatalf("mkdir docs/adr: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "adr", "ADR-0001.md"), []byte("# adr\n"), 0o644); err != nil {
		t.Fatalf("write adr file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# readme\n"), 0o644); err != nil {
		t.Fatalf("write readme: %v", err)
	}

	first, err := Run(root, RunOptions{RunID: "run-a"})
	if err == nil {
		// report-only should succeed
	} else {
		t.Fatalf("run first: %v", err)
	}
	second, err := Run(root, RunOptions{RunID: "run-b"})
	if err != nil {
		t.Fatalf("run second: %v", err)
	}

	if !reflect.DeepEqual(first.Inventory, second.Inventory) {
		t.Fatalf("inventory mismatch across runs")
	}
	if !reflect.DeepEqual(first.Classification, second.Classification) {
		t.Fatalf("classification mismatch across runs")
	}
}

func TestRun_ApplyFailsWithoutLLMWhenUnresolvedExists(t *testing.T) {
	t.Setenv("MINDSPEC_LLM_PROVIDER", "off")
	t.Setenv("MINDSPEC_LLM_MODEL", "")

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "misc"), 0o755); err != nil {
		t.Fatalf("mkdir misc: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "misc", "notes.md"), []byte("# notes\n"), 0o644); err != nil {
		t.Fatalf("write notes.md: %v", err)
	}

	report, err := Run(root, RunOptions{Apply: true, ArchiveMode: "copy", RunID: "run-apply"})
	if err == nil {
		t.Fatal("expected apply failure when LLM unavailable and unresolved docs exist")
	}
	if report == nil {
		t.Fatal("expected report on failure")
	}
	if len(report.Unresolved) == 0 {
		t.Fatal("expected unresolved docs")
	}
	if !strings.Contains(err.Error(), "no provider is configured") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRun_ApplyPromotesCanonicalAndArchivesSources(t *testing.T) {
	t.Setenv("MINDSPEC_LLM_PROVIDER", "off")
	t.Setenv("MINDSPEC_LLM_MODEL", "")

	root := t.TempDir()
	mk := func(rel, content string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}

	mk("docs/core/ARCHITECTURE.md", "# arch\n")
	mk("docs/adr/ADR-0001.md", "# adr\n")
	mk("docs/specs/001-demo/spec.md", "# spec\n")
	mk("docs/domains/core/overview.md", "# domain\n")
	mk("docs/context-map.md", "# context map\n")
	mk("GLOSSARY.md", "# glossary\n")
	mk("architecture/policies.yml", "policies:\n  - id: x\n    reference: \"docs/core/ARCHITECTURE.md\"\n")

	report, err := Run(root, RunOptions{Apply: true, ArchiveMode: "copy", RunID: "run-ok"})
	if err != nil {
		t.Fatalf("apply run failed: %v", err)
	}
	if report == nil {
		t.Fatal("expected report")
	}

	// Canonical docs promoted
	for _, rel := range []string{
		".mindspec/docs/core/ARCHITECTURE.md",
		".mindspec/docs/adr/ADR-0001.md",
		".mindspec/docs/specs/001-demo/spec.md",
		".mindspec/docs/domains/core/overview.md",
		".mindspec/docs/context-map.md",
		".mindspec/docs/glossary.md",
		".mindspec/policies.yml",
		".mindspec/lineage/manifest.json",
	} {
		if _, statErr := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); statErr != nil {
			t.Fatalf("expected canonical artifact %s: %v", rel, statErr)
		}
	}

	// Legacy markdown docs archived in copy mode.
	for _, rel := range []string{
		"docs/core/ARCHITECTURE.md",
		"docs/adr/ADR-0001.md",
		"docs/specs/001-demo/spec.md",
		"docs/domains/core/overview.md",
		"docs/context-map.md",
		"GLOSSARY.md",
	} {
		archived := filepath.Join(root, "docs_archive", "run-ok", filepath.FromSlash(rel))
		if _, statErr := os.Stat(archived); statErr != nil {
			t.Fatalf("expected archived source %s: %v", rel, statErr)
		}
		// Copy mode keeps source.
		if _, statErr := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); statErr != nil {
			t.Fatalf("expected source to remain in copy mode %s: %v", rel, statErr)
		}
	}

	policyBytes, err := os.ReadFile(filepath.Join(root, ".mindspec", "policies.yml"))
	if err != nil {
		t.Fatalf("read canonical policies: %v", err)
	}
	if !strings.Contains(string(policyBytes), "reference: \".mindspec/docs/core/ARCHITECTURE.md\"") {
		t.Fatalf("expected canonicalized policy reference, got:\n%s", string(policyBytes))
	}
}
