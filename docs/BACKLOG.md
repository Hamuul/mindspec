# MindSpec Product Backlog

> **Principle**: Prioritize features that enable MindSpec to assist in building MindSpec itself (dogfooding).
> **Language**: Go (per ADR-0004, accepted). All CLI work targets the Go binary.

## Priority Tiers

| Tier | Description |
|:-----|:-----------|
| **P0** | Immediately useful for the next development session |
| **P1** | Needed within the first few specs |
| **P2** | Important for scaled usage |
| **P3** | Nice-to-have / future enhancements |

---

## Done

### Documentation Alignment
- [x] Three-mode system (Spec/Plan/Implement) documented in MODES.md
- [x] ARCHITECTURE.md rewritten for Beads + worktrees + domains + Claude Code
- [x] AGENTS.md updated for three-mode system + Beads + ADR governance
- [x] Agent rules (.claude/rules/mindspec-modes.md) aligned with three modes
- [x] GLOSSARY.md rebuilt with v1 primitives
- [x] CONVENTIONS.md updated with domain/worktree/Beads conventions
- [x] policies.yml expanded for Plan mode, ADR governance, domains, Beads, worktrees
- [x] ADR-0001: DDD Enablement + DDD-Informed Context Packs (proposed)
- [x] ADR-0002: Beads Integration Strategy (proposed → accepted)
- [x] ADR-0003: Centralized Agent Instruction Emission (proposed → accepted)
- [x] ADR-0004: Go as v1 CLI Implementation Language (accepted)
- [x] INIT.md archived (superseded by mindspec.md)

### 000: Repo + Beads Hygiene ✓
- [x] Beads initialized in repo (`.beads/` with durable state)
- [x] Selective `.beads/` gitignore (runtime ignored, durable tracked)
- [x] Packaging excludes (`MANIFEST.in`)
- [x] `mindspec doctor` Beads hygiene checks (Python prototype)

---

## P0: Immediate Value (Use While Building MindSpec)

### 001: Go CLI Skeleton + Doctor ✓
### 002: Glossary-Based Context Injection ✓
### 003: Context Pack Generation (with DDD Routing) ✓

---

## P1: Core Workflow Support

### 004: `mindspec instruct` — Mode-Aware Guidance Emission ✓
### 005: `mindspec next` — Work Selection + Claiming ✓
### 006: `mindspec validate` — Workflow Checks ✓
### 007: Beads Integration Conventions + Tooling ✓

### 008: Worktree Lifecycle Management ← **NEXT**
**Why P1**: Implementation Mode requires worktree isolation. Workflow commands should handle worktrees implicitly.

**Status**: Spec APPROVED (draft), Plan in progress. ADR-0006 (branch protection) and ADR-0007 (per-worktree state) drafted as Proposed.

**Scope**:
- `mindspec next` creates worktree via `bd worktree create` after claiming bead
- New `mindspec complete` command: close bead + `bd worktree remove` + advance state
- Delegate all worktree CRUD to Beads (`bd worktree create/list/remove/info`)
- Update `implement.md` template to use `mindspec complete` as single close-out step
- Deprecate `mindspec bead worktree`

### 008a: Molecule-Based Plan Decomposition
**Why P1**: `bead plan` reimplements Beads molecule DAG creation. Beads formulas and `bd mol pour` handle this natively.

**Scope**:
- Replace `internal/bead/plan.go` (`CreatePlanBeads()`) with Beads molecule creation
- Convert plan `work_chunks` YAML to a Beads formula (or call `bd mol pour` with equivalent structure)
- Leverage `bd mol show --parallel` for DAG visualization
- `bd mol ready` replaces custom ready-work filtering for plan beads

### 008b: Human Gates for Approval Workflow
**Why P1**: Spec-approve and plan-approve are human gates tracked only in markdown frontmatter. Beads has first-class `human` gate support.

**Scope**:
- Model spec approval as a Beads human gate on the spec bead
- Model plan approval as a Beads human gate on the plan molecule
- `bd gate resolve <id>` becomes the execution signal (complements frontmatter as document record)
- `bd ready` naturally shows work unblocked by resolved gates
- Update `/spec-approve` and `/plan-approve` skills to resolve Beads gates alongside frontmatter update

### 008c: Compose `bd prime` into `mindspec instruct`
**Why P1**: `bd prime` provides Beads workflow context (~1-2k tokens), `mindspec instruct` provides spec-driven process guidance. Composing them gives agents a complete picture.

**Scope**:
- `mindspec instruct` embeds or appends `bd prime` output
- Avoids agents needing two separate context sources
- Respect token budgets — `bd prime` is already compact

### 009: Domain Scaffold + Context Map
**Why P1**: DDD primitives need tooling support.

**Scope**:
- `mindspec domain add <name>`: scaffold `/docs/domains/<domain>/` with template files
- `mindspec domain list`: show registered domains
- Domain operations produce ADR drafts

**Partial**: Initial domain structure and `docs/context-map.md` created manually.

### 010: ADR Lifecycle Tooling
**Why P1**: ADR governance needs tooling support.

**Scope**:
- `mindspec adr create <title>`: generate ADR template with metadata
- `mindspec adr list`: show ADRs by status
- Superseding workflow: create new ADR linking to superseded one
- Validate ADR citations in plans

### 011: Proof Runner (MVP)
**Why P1**: Foundation for "proof-of-done" invariant.

**Scope**:
- Parse `Validation Proofs` section from spec.md
- Execute listed commands and capture output
- Report pass/fail with artifacts
- CLI: `mindspec proof run <spec-id>`

---

## P2: Project Health + Memory

### 012: Memory Service (Basic)
**Why P2**: Persist decisions, gotchas, debugging outcomes across sessions.

**Scope**:
- Local persistent store
- CLI: `mindspec memory save`, `mindspec memory recall`
- Tag by spec-id, domain, keywords
- Memory entries reference canonical beads or specs (per ADR-0002)

### 013: `mindspec init` — Project Bootstrap
**Why P2**: Scaffolds a new MindSpec project from scratch.

**Scope**:
- Create missing docs folders, templates, context-map placeholders
- Check for Beads presence and instruct how to init
- Generate starter GLOSSARY.md, AGENTS.md, CLAUDE.md

**Note**: Deferred from P0 — manual setup is fine while dogfooding on MindSpec itself.

---

## P3: Advanced Features

### 014: Architecture Divergence Detection
- Compare implementation against documented architecture
- Auto-trigger ADR divergence protocol when violations detected

### 015: Parallel Task Dispatch
- Identify ready beads (no unresolved dependencies)
- Generate per-bead context packets for parallel agent execution

### 016: Observability / Telemetry
- Glossary hit/miss rates
- Token budgets and cache rates
- OTel-friendly event shaping for future Agent Mind Visualization

### 017: Cross-Platform Release Automation
- CI/CD pipeline for Go binary builds
- Multi-arch binaries (darwin/linux, amd64/arm64)
- GitHub Releases or homebrew tap

---

## Implementation Order

```
P0: 000 ✓ → 001 ✓ (Go skeleton + doctor)
    → 002 ✓ (glossary) → 003 ✓ (context packs)

P1: 004 ✓ (instruct) → 005 ✓ (next) → 006 ✓ (validate) → 007 ✓ (Beads tooling)
    → 008 (worktree lifecycle + mindspec complete)  ← NEXT
    → 008a (molecule-based plan decomposition)
    → 008b (human gates for approval)
    → 008c (compose bd prime into instruct)
    → 009 (domains) → 010 (ADRs) → 011 (proofs)

P2: 012 (memory) → 013 (init)
```

**Rationale**:
- 001–007 are done. 008 is the next priority.
- 008a/b/c deepen Beads integration before adding new MindSpec features. Molecules replace custom DAG code, gates formalize approval tracking, prime composition unifies agent context.
- 008a/b/c can be done in any order after 008 lands (they're independent of each other).
- 009+ resume after Beads integration is solid.
