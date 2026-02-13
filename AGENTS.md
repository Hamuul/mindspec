# MindSpec Agent Instructions — Human Reference

> **For agents**: This file is a human-readable reference document. Agent operating guidance is emitted dynamically by `mindspec instruct` (run automatically by the SessionStart hook). Agents should rely on `mindspec instruct` output, not this file.

## Mode System

All work follows a three-phase approach: **Spec → Plan → Implement**. Each mode has permitted/forbidden actions, human gates, and exit criteria. See [MODES.md](docs/core/MODES.md) for full definitions.

> **Rule**: Never create or modify code without an approved spec AND an approved plan.

## Beads Integration

Beads is the **execution tracking substrate** (not a planning system). See [ADR-0002](docs/adr/ADR-0002.md).

## Git Workflow

- **Clean tree rule**: commit or revert before mode transitions. Never auto-stash.
- **Milestone commits**: `spec(<bead-id>)`, `plan(<bead-id>)`, `impl(<bead-id>)` — agent convention, not CLI-enforced.
- **Co-commit**: `.beads/` and `.mindspec/state.json` alongside relevant work.

## ADR Governance

ADR divergence always triggers a human gate. See [ARCHITECTURE.md](docs/core/ARCHITECTURE.md).

## Documentation Sync

Every code change must update corresponding documentation. "Done" includes doc-sync.

## Key Documentation

| Document | Purpose |
|:---------|:--------|
| [mindspec.md](mindspec.md) | Product specification (source of truth) |
| [MODES.md](docs/core/MODES.md) | Mode definitions and transitions |
| [ARCHITECTURE.md](docs/core/ARCHITECTURE.md) | System design and invariants |
| [CONVENTIONS.md](docs/core/CONVENTIONS.md) | File organization and naming |
| [GLOSSARY.md](GLOSSARY.md) | Term definitions for context injection |
