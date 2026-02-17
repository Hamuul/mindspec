package brownfield

import (
	"os"
	"path/filepath"
	"reflect"
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

func TestRunApply_NotImplementedYet(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# test\n"), 0o644); err != nil {
		t.Fatalf("write README.md: %v", err)
	}

	report, err := RunApply(root, "copy")
	if err == nil {
		t.Fatal("expected not implemented error")
	}
	if report == nil {
		t.Fatal("expected report output on apply failure")
	}
}
