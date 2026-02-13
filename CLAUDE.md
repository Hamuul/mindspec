# CLAUDE.md — MindSpec

MindSpec is a spec-driven development framework (Claude Code-first). See [mindspec.md](mindspec.md) for the full product specification.

## Guidance

Run `mindspec instruct` for mode-appropriate operating guidance. This is emitted automatically by the SessionStart hook.

## Build & Test

```bash
make build    # Build binary to ./bin/mindspec
make test     # Run all tests
```

## Custom Commands

| Command | Purpose |
|:--------|:--------|
| `/spec-init` | Initialize a new specification (enters Spec Mode) |
| `/spec-approve` | Request Spec → Plan transition |
| `/plan-approve` | Request Plan → Implementation transition |
| `/spec-status` | Check current mode and active spec/bead state |
