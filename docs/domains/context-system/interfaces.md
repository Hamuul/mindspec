# Context-System Domain — Interfaces

## Provided Interfaces

### Glossary Parsing

```go
// internal/glossary/glossary.go
glossary.Parse(root string) ([]glossary.Entry, error)
// Returns all glossary entries with Term, Label, Target, FilePath, Anchor
```

### Glossary Matching

```go
// internal/glossary/match.go
glossary.Match(entries []glossary.Entry, text string) []glossary.Entry
// Returns matched terms, longest-match-first, case-insensitive
```

### Section Extraction

```go
// internal/glossary/section.go
glossary.ExtractSection(root, filePath, anchor string) (string, error)
// Extracts a specific section from a markdown file by anchor
```

### Context Pack Generation (Planned — Spec 003)

```go
// Planned: internal/context/
contextpack.Build(specID string, mode string) (*ContextPack, error)
// Assembles a context pack for the given spec and mode
```

## Consumed Interfaces

- **core**: `workspace.FindRoot()`, `workspace.GlossaryPath()`, `workspace.DocsDir()`
- **workflow**: Spec bead metadata (impacted domains, ADR citations) for context pack routing

## Events

None defined yet. Future: context pack generation events for observability (tokens injected, cache hits).
