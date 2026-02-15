# MindSpec

**Spec-driven development for AI coding agents.**

MindSpec is a CLI framework that structures how AI agents move from idea to implementation. Specs, plans, architecture decisions, domain documentation, and context packs are all first-class, versioned artifacts — not afterthoughts bolted on after the code is written.

Every phase transition goes through a human gate. The agent drafts, the human approves. Architecture divergence is detected and blocked until explicitly resolved. Documentation stays in sync because the system won't let you close work without it.

```
Idle ──→ Spec Mode ──human gate──→ Plan Mode ──human gate──→ Implementation ──→ Review ──human gate──→ Idle
```

## The Problem

AI coding agents are powerful but unstructured. Without guardrails they:

- **Drift from intent** — the agent builds what it infers, not what you specified
- **Ignore architecture** — existing ADRs and design decisions get steamrolled
- **Lose context between sessions** — every conversation starts from scratch
- **Skip documentation** — code ships, docs rot
- **Resist scope discipline** — a "small feature" becomes a refactor of three subsystems

MindSpec treats these as system design problems, not prompting problems.

## Quick Start

```bash
# Build from source
make build

# Bootstrap a new project
./bin/mindspec init

# Check project health
./bin/mindspec doctor

# Start a new feature
./bin/mindspec spec-init 025-my-feature

# See what the agent should do right now
./bin/mindspec instruct
```

## How It Works

### Gated Lifecycle

MindSpec enforces a phased workflow where each transition requires explicit human approval:

**Spec Mode** — Define what "done" looks like. Problem statement, acceptance criteria, impacted domains, ADR touchpoints. No code allowed. The agent and human iterate on the spec until it's right, then the human approves.

**Plan Mode** — Decompose the spec into bounded work chunks. Review applicable ADRs. Check architectural fitness. Wire up dependencies. If existing architecture doesn't fit, the divergence protocol fires — you can't quietly ignore a design decision. The human approves the plan.

**Implementation Mode** — Execute in an isolated git worktree. One bead per worktree, scoped to exactly what the plan defined. Doc-sync is mandatory — you can't close a bead without updating documentation. Discovered work becomes new beads, not scope creep.

**Review Mode** — Validate the implementation against the original spec's acceptance criteria. Human approves to return to idle.

### Context Packs

MindSpec assembles deterministic, token-budgeted context for each phase:

```bash
mindspec context pack 009-my-feature
```

A context pack pulls from the spec, relevant domain docs, applicable ADRs, glossary terms, neighboring bounded contexts (via the Context Map), and active policies — then deduplicates and respects token budgets. The agent gets exactly what it needs, nothing more.

This replaces the "go read file X, now read file Y, also check Z" pattern with a single deterministic bundle that includes provenance for every section.

### Architecture Decision Records

ADRs are a governed primitive, not a dusty folder:

```bash
mindspec adr create --title "Use WebSockets for real-time updates" --domain viz
mindspec adr list --status accepted
```

Plans must cite the ADRs they rely on. If implementation needs to deviate from a cited ADR, the agent stops and escalates — you approve a new superseding ADR or reject the divergence. Architecture stays coherent across dozens of specs because the system enforces it.

### Dynamic Agent Guidance

Instead of maintaining sprawling static CLAUDE.md files that try to anticipate every scenario, MindSpec emits agent instructions at runtime:

```bash
mindspec instruct
```

This reads current state (mode, active spec, active bead, worktree status) and generates mode-appropriate guidance. In spec mode the agent gets spec-writing rules. In implementation mode it gets scope discipline, doc-sync requirements, and commit conventions. The guidance adapts to where you actually are.

### Validation

Pre-flight checks catch problems before they compound:

```bash
mindspec validate spec 009-my-feature    # Structure, acceptance criteria quality
mindspec validate plan 009-my-feature    # Frontmatter, ADR citations, verification steps
mindspec validate docs                   # Doc-sync across the project
```

Vague acceptance criteria get flagged. Missing ADR citations get flagged. Plans without verification steps get flagged. You fix these before implementation, not after.

## Domain-Driven Design

MindSpec borrows heavily from DDD — specifically the idea that **bounded contexts** are as valuable for structuring agent work as they are for structuring microservices.

```
docs/
├── domains/
│   ├── viz/
│   │   ├── overview.md
│   │   ├── architecture.md
│   │   ├── interfaces.md
│   │   └── adr/
│   └── workflow/
│       ├── overview.md
│       ├── architecture.md
│       └── ...
├── context-map.md          # Bounded context relationships
└── specs/
    └── 009-my-feature/
        ├── spec.md
        ├── plan.md
        └── context-pack.md
```

Specs declare impacted domains. Context packs route through the Context Map, expanding one hop to include neighboring bounded contexts. Domain-scoped ADRs live alongside domain docs. When you split or merge domains, their documentation and decisions move with them.

The insight from DDD is that **context boundaries reduce ambiguity**. When an agent knows it's working within the `viz` domain and the Context Map shows `viz` depends on `observability` for trace data, the context pack includes both — and nothing else. No prompt engineering required.

## Beads Integration

MindSpec uses [Beads](https://github.com/steveyegge/beads) as its execution tracking substrate. Where MindSpec handles orchestration, specs, and context, Beads handles the work graph.

### What Beads Provides

- **Durable issue tracking** — git-native, survives across sessions without external services
- **Dependency graph** — `bd ready` shows what's unblocked, `bd blocked` shows what's waiting
- **Molecule decomposition** — parent beads (specs) contain child beads (implementation chunks) with wired dependencies
- **Hash-based IDs** — stable, collision-free identifiers that work offline
- **JSONL export** — `bd sync --flush-only` serializes state for portability
- **SQLite + JSONL dual backend** — fast queries locally, durable flat-file fallback

### How MindSpec Leverages Beads

**Gate beads** model human approval as dependency-blocking issues. A plan's implementation beads are blocked by a `[GATE plan-approve]` bead. When the human approves, the gate resolves and implementation beads become ready. This makes the human-in-the-loop workflow a natural part of the dependency graph rather than a special case.

**Worktree isolation** ties each bead to a git worktree (`worktree-<bead-id>`). The agent works in the worktree, and closing the bead requires clean state, documentation updates, and verification evidence.

**Session continuity** comes from Beads persistence. When a new session starts, `bd prime` recovers the full work graph. The agent picks up where it left off — no manual context recovery.

**Scope discipline** is enforced structurally. A bead has a defined scope from the plan. If the agent discovers additional work during implementation, it creates a new bead rather than expanding the current one. The dependency graph absorbs the new work naturally.

```bash
# Typical implementation flow
bd ready                              # What's available?
mindspec next                         # Claim bead, create worktree, set state
# ... implement ...
mindspec complete                     # Close bead, remove worktree, advance state
```

## Project Structure

```
your-project/
├── .mindspec/
│   └── state.json              # Current mode, active spec/bead (committed)
├── .beads/                     # Beads work graph (committed)
├── docs/
│   ├── core/                   # Architecture, modes, conventions, usage
│   ├── domains/<name>/         # Domain-scoped documentation
│   ├── adr/                    # Cross-cutting architecture decisions
│   ├── specs/<id>/             # Specifications with plans and context packs
│   ├── context-map.md          # Bounded context relationships
│   └── templates/              # Templates for specs, plans, ADRs
├── architecture/
│   └── policies.yml            # Machine-checkable architectural policies
├── GLOSSARY.md                 # Term → doc section mapping
└── CLAUDE.md                   # Minimal bootstrap (points to CLI)
```

## CLI Reference

### Workflow

| Command | Description |
|:--------|:------------|
| `mindspec instruct` | Emit mode-appropriate agent guidance |
| `mindspec state show` | Show current mode and active work |
| `mindspec next` | Claim next ready bead, create worktree |
| `mindspec complete` | Close bead, remove worktree, advance state |
| `mindspec approve spec <id>` | Approve spec, transition to Plan Mode |
| `mindspec approve plan <id>` | Approve plan, transition to Implementation |
| `mindspec approve impl <id>` | Approve implementation, return to Idle |

### Context & Documentation

| Command | Description |
|:--------|:------------|
| `mindspec context pack <id>` | Generate token-budgeted context pack |
| `mindspec glossary list\|match\|show` | Term lookup and section extraction |
| `mindspec adr create\|list\|show` | ADR lifecycle management |
| `mindspec validate spec\|plan\|docs` | Pre-flight validation checks |

### Project Management

| Command | Description |
|:--------|:------------|
| `mindspec init` | Bootstrap project structure |
| `mindspec spec-init <id>` | Create new specification |
| `mindspec doctor` | Project health checks |

### Observability

| Command | Description |
|:--------|:------------|
| `mindspec trace summary <file>` | Summarize NDJSON trace events |
| `mindspec bench setup\|collect\|report` | Benchmark agent sessions |
| `mindspec viz live\|replay` | Real-time 3D agent activity visualization |

## Design Principles

1. **Docs-first** — every code change updates documentation, enforced by the system
2. **Spec-anchored** — all implementation traces back to a versioned specification
3. **Human gates for divergence** — architecture deviations require approval and a new ADR
4. **Proof of done** — beads close only with verification evidence
5. **Scope discipline** — discovered work becomes new beads, never scope creep
6. **Dynamic over static** — runtime guidance beats static files that drift
7. **CLI-first** — logic lives in testable, versionable Go; IDE integrations are thin shims
8. **Deterministic context** — token-budgeted context packs, not "go read this file" prompting

## Requirements

- Go 1.22+
- [Beads](https://github.com/steveyegge/beads) CLI (`bd`)
- Git (for worktree support)
- Claude Code (for agent integration; MindSpec is Claude Code-first but the CLI is standalone)

## Building

```bash
make build      # Build to ./bin/mindspec
make test       # Run all tests
make install    # Install to $GOPATH/bin
```

## License

MIT
