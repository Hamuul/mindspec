# MindSpec Conventions

This document outlines the file organization, naming, and structural conventions for MindSpec-managed projects.

## File Organization

| Path | Purpose |
|:-----|:--------|
| `docs/core/` | Permanent architectural and convention documents |
| `docs/domains/<domain>/` | Domain-scoped documentation (overview, architecture, interfaces, runbook, ADRs) |
| `docs/specs/` | Historical and active specifications |
| `docs/context-map.md` | Bounded context relationships and integration contracts |
| `docs/adr/` | Cross-cutting architecture decision records |
| `architecture/` | Machine-readable policies |
| `GLOSSARY.md` | Concept-to-doc-section mapping for context injection |
| `docs/templates/` | Templates for specs, ADRs, domain docs |
| `AGENTS.md` | Agent behavioral instructions |
| `CLAUDE.md` | Claude Code project instructions |
| `mindspec.md` | Product specification (source of truth) |

## Domain Doc Structure

Each domain lives at `/docs/domains/<domain>/` with:

| File | Purpose |
|:-----|:--------|
| `overview.md` | What the domain owns, its boundaries |
| `architecture.md` | Key patterns, invariants |
| `interfaces.md` | APIs, events, contracts (published language) |
| `runbook.md` | Ops/dev workflows |
| `adr/ADR-xxxx.md` | Domain-scoped architecture decision records |

## Specification Naming

Specs follow the pattern `NNN-slug-name`:
- `001-skeleton`
- `002-glossary`
- `003-context-pack`

## ADR Naming

ADRs follow the pattern `ADR-NNNN.md`:
- Cross-cutting: `docs/adr/ADR-NNNN.md`
- Domain-scoped: `docs/domains/<domain>/adr/ADR-NNNN.md`

ADR metadata must include: domain(s), status (proposed/accepted/superseded), supersedes/superseded-by links, decision + rationale + consequences.

## Beads Conventions

- Spec beads contain a **concise summary** and **link to the canonical spec file**. No long-form content.
- Implementation beads contain: scope, micro-plan (3-7 steps), verification steps, dependencies.
- Keep the active workset intentionally small. Regularly clean up completed beads.
- Rely on git history + documentation for historical traceability, not Beads as archive.

## Git Workflow Conventions

### Clean Tree Rule

A **clean working tree is a hard precondition** for:

- Starting new work (picking up a bead)
- Switching modes (Spec → Plan → Implement → Done)
- Running `mindspec next`, `mindspec pickup`, or any mode transition

If the tree is dirty: **commit or revert**. Do not auto-stash (it hides state and breaks determinism).

### Milestone Commits

Mode transitions are marked with explicit commits:

| Transition | What to commit |
|:-----------|:---------------|
| **Spec → Plan** | Spec artifact + bead update recording "spec approved" |
| **Plan → Implement** | Plan artifacts, spawned child beads, bead updates |
| **Implement → Done** | Code, tests, docs, bead closure notes |

Normal commits during a mode are expected and encouraged (especially in Implementation Mode — tests first, refactor, docs, etc.). The milestone commit marks the boundary cleanly.

### Commit Message Conventions

Use conventional-commit style scoped to the bead ID:

```
spec(<bead-id>): <summary>
plan(<bead-id>): <summary>
impl(<bead-id>): <summary>
chore(<bead-id>): <summary>
```

- `spec` — spec artifacts and related documentation
- `plan` — plan artifacts, bead creation, dependency mapping
- `impl` — implementation code, tests, doc-sync
- `chore` — cleanup, formatting, dependency bumps, tooling

### Co-committing `.beads/`

Always commit `.beads/` changes alongside the relevant work in the same commit, unless operating in a mode where Beads is not tracked in git.

### Preflight (before starting any forward-progress work)

1. Confirm you are on the correct worktree/branch for the active bead
2. Confirm working tree is clean (`git status` shows no changes). If not: commit with an appropriate message, or revert/discard the changes.
3. Confirm the active bead exists and is in the expected state
4. Only then proceed

## Worktree Conventions

- Worktrees are named with the bead ID: `worktree-<bead-id>`
- One worktree per implementation bead
- Changes are isolated per bead
- Closing a bead requires clean state sync from worktree

## Glossary Conventions

- **Pathing**: Always use **relative paths** from the project root for glossary targets (e.g., `docs/core/ARCHITECTURE.md#section-id`). Do not use absolute paths.
- **Format**: Use the standard table format: `| **Term** | [label](relative/path#anchor) |`.
- **Coverage**: Every new concept introduced in a spec or domain doc should have a glossary entry.

## Documentation Anchors

Use stable Markdown header anchors for deterministic section retrieval:
`## Component X {#component-x}`

## Tooling Interface (Tentative)

The primary interface will be a CLI. Usage pattern:

- `mindspec context pack <spec-id>`: Generate context for an agent session
- `mindspec validate spec <id>`: Check acceptance criteria quality
- `mindspec validate docs`: Verify doc-sync compliance
- `mindspec doctor`: Project structure health check
