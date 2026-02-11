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

### 001: Go CLI Skeleton + Doctor
**Why P0**: Foundation for everything. Establishes Go binary, workspace detection, and project health validation. Ports Python prototype doctor to Go. Stubs ADR-0003 command surface.

**Scope**:
- Go module scaffolding (`cmd/mindspec/`, `internal/`, `go.mod`)
- CLI entry point with subcommand routing (cobra)
- Workspace root detection (`mindspec.md` or `.git`)
- `mindspec doctor`: validates `docs/core/`, `docs/domains/`, `docs/specs/`, `architecture/`, `GLOSSARY.md`, `.beads/` (durable state + no tracked runtime artifacts)
- Stub commands: `mindspec instruct`, `mindspec next`, `mindspec validate` (per ADR-0003)
- Retire Python `src/mindspec/` prototype

**Immediate Use**: Working `mindspec` binary for project health checks and command surface.

### 002: Glossary-Based Context Injection
**Why P0**: Enables deterministic doc retrieval based on keywords.

**Scope**:
- Parse `GLOSSARY.md` into keyword-to-target mapping
- Match keywords from input text
- Extract targeted documentation sections
- CLI: `mindspec glossary list|match|show`

**Immediate Use**: Agent can pull architectural context when working on specs.

### 003: Context Pack Generation (with DDD Routing)
**Why P0**: Reproducible context bundles for agent sessions.

**Scope** (split into two deliverables):

**003a: Pack Manifest Builder**
- Determine which files/sections to include and why
- DDD-informed assembly: start from impacted domains, 1-hop neighbor expansion via Context Map (per ADR-0001)
- Provenance tracking: record why each piece was included
- Output: structured manifest (JSON or equivalent)

**003b: Renderer**
- Turn manifest into `context-pack.md` in spec directory
- Include: spec, matched domain docs, accepted ADRs, policies, commit tuple
- Respect token budgets

**Immediate Use**: Consistent, domain-aware context for every session.

---

## P1: Core Workflow Support

### 004: `mindspec instruct` — Mode-Aware Guidance Emission
**Why P1**: Core of ADR-0003. Replaces static CLAUDE.md/AGENTS.md with dynamic, mode-aware guidance.

**Scope**:
- Detect current mode (Spec/Plan/Implement) from local state (spec status, plan status, worktree)
- Detect active work item (spec ID, bead ID) from Beads + worktree conventions
- Emit authoritative operating guidance: current mode, active work, required outputs, hard gates
- Read-only, no side effects
- Reduce CLAUDE.md/AGENTS.md to minimal bootstrap pointing to `mindspec instruct`

### 005: `mindspec next` — Work Selection + Claiming
**Why P1**: Integrates with Beads for ready-work discovery and worktree association.

**Scope**:
- Query Beads for ready work (`bd ready`)
- Claim/lock a work item
- Associate with worktree (create if needed)
- Emit guidance or instruct to run `mindspec instruct`

### 006: `mindspec validate` — Workflow Checks
**Why P1**: Consolidates doc-sync, ADR divergence, and structural validation.

**Scope**:
- `mindspec validate docs`: compare changed files against doc requirements, flag missing updates
- `mindspec validate spec <id>`: check acceptance criteria quality, section completeness
- `mindspec validate plan <id>`: verify beads have verification steps, ADR citations, acyclic deps
- ADR divergence gate checks

### 007: Beads Integration Conventions + Tooling
**Why P1**: Beads is central to the execution model; conventions must be codified.

**Scope**:
- Spec bead creation from approved spec (concise summary + link)
- Implementation bead creation from plan
- Active workset hygiene commands
- Bead-to-worktree mapping
- Reference hygiene rules established in 000

### 008: Worktree Lifecycle Management
**Why P1**: Implementation Mode requires worktree isolation.

**Scope**:
- Create worktree for a bead: `mindspec worktree create <bead-id>`
- Naming convention: `worktree-<bead-id>`
- Clean state sync on bead closure
- List active worktrees: `mindspec worktree list`

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
P0: 000 ✓ → 001 (Go skeleton + doctor)
    → 002 (glossary) → 003a (pack manifest) → 003b (renderer)

P1: 004 (instruct) → 005 (next) → 006 (validate)
    → 007 (Beads tooling) → 008 (worktrees) → 009 (domains) → 010 (ADRs) → 011 (proofs)

P2: 012 (memory) → 013 (init)
```

**Rationale**:
- Go CLI skeleton is the new foundation — everything builds on it
- Glossary + context packs are immediately useful once doctor works
- ADR-0003 commands (`instruct`, `next`, `validate`) move to P1 since they require mode detection and Beads integration that benefits from having the basic CLI solid first
- `mindspec init` demoted to P2 — manual project setup is fine while dogfooding
- 001a (workspace root fix) folded into 001 — Go rewrite handles it naturally
- Former 009 (doc-sync) and 010/011 (spec/plan validation) consolidated into 006 (`validate`)
