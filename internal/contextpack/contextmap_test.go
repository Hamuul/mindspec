package contextpack

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestParseContextMap(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "context-map.md")

	content := `# Context Map

## Bounded Contexts

### Core
**Owns**: CLI

## Relationships

### Core → Context-System (upstream)

Core provides workspace resolution.

**Contract**: [interfaces](domains/core/interfaces.md)

### Core → Workflow (upstream)

Core provides CLI shell.

**Contract**: [interfaces](domains/core/interfaces.md)

### Workflow → Context-System (upstream)

Workflow provides spec metadata.

**Contract**: [interfaces](domains/workflow/interfaces.md), [context](domains/context-system/interfaces.md)

### Context-System → Workflow (downstream)

Context-system delivers context packs.

## Source of Truth
`
	os.WriteFile(path, []byte(content), 0o644)

	rels, err := ParseContextMap(path)
	if err != nil {
		t.Fatalf("ParseContextMap: %v", err)
	}

	if len(rels) != 4 {
		t.Fatalf("got %d relationships, want 4", len(rels))
	}

	// Check first relationship
	if rels[0].From != "core" || rels[0].To != "context-system" {
		t.Errorf("rels[0] = %s→%s, want core→context-system", rels[0].From, rels[0].To)
	}
	if rels[0].Direction != "upstream" {
		t.Errorf("rels[0].Direction = %q, want upstream", rels[0].Direction)
	}
}

func TestParseContextMap_ArrowVariants(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "context-map.md")

	content := `## Relationships

### A -> B (downstream)

Some description.
`
	os.WriteFile(path, []byte(content), 0o644)

	rels, err := ParseContextMap(path)
	if err != nil {
		t.Fatalf("ParseContextMap: %v", err)
	}
	if len(rels) != 1 {
		t.Fatalf("got %d relationships, want 1", len(rels))
	}
	if rels[0].From != "a" || rels[0].To != "b" {
		t.Errorf("got %s→%s, want a→b", rels[0].From, rels[0].To)
	}
}

func TestResolveNeighbors(t *testing.T) {
	rels := []Relationship{
		{From: "core", To: "context-system", Direction: "upstream"},
		{From: "core", To: "workflow", Direction: "upstream"},
		{From: "workflow", To: "context-system", Direction: "upstream"},
		{From: "context-system", To: "workflow", Direction: "downstream"},
	}

	neighbors := ResolveNeighbors(rels, []string{"context-system"})
	sort.Strings(neighbors)

	want := []string{"core", "workflow"}
	if !reflect.DeepEqual(neighbors, want) {
		t.Errorf("ResolveNeighbors = %v, want %v", neighbors, want)
	}
}

func TestResolveNeighbors_ExcludesImpacted(t *testing.T) {
	rels := []Relationship{
		{From: "core", To: "context-system"},
		{From: "core", To: "workflow"},
	}

	neighbors := ResolveNeighbors(rels, []string{"core", "context-system"})
	want := []string{"workflow"}
	if !reflect.DeepEqual(neighbors, want) {
		t.Errorf("ResolveNeighbors = %v, want %v", neighbors, want)
	}
}
