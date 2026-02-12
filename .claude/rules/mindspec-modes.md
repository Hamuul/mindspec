# MindSpec Mode Rules (Bootstrap)

> **Authoritative guidance is emitted dynamically by `mindspec instruct`.**
> A SessionStart hook runs this automatically. If you don't see mode guidance in your context, run `mindspec instruct` manually.

## Core Invariant

Before writing any code, you MUST verify:

1. **Spec exists and is approved**: A spec in `docs/specs/<id>/spec.md` with `Status: APPROVED`
2. **Plan exists and is approved**: Implementation beads are defined with explicit verification steps
3. **You are working on a specific bead**: A bead ID is active

If these are not met, you are in **Spec Mode** or **Plan Mode**. Only proceed to Implementation Mode when all three hold.

## Quick Reference

- **Check state**: `mindspec state show` or `mindspec instruct`
- **Set state**: `mindspec state set --mode=<mode> --spec=<id> [--bead=<id>]`
- **Pick up work**: `mindspec next` — queries ready beads, claims one, sets state, emits guidance
- **Pre-check**: `mindspec validate spec|plan|docs` — structural quality checks before approval
- **Spec Mode**: Markdown only. No code. Use `/spec-approve` to transition.
- **Plan Mode**: Plan + beads only. No code. Use `/plan-approve` to transition.
- **Implementation Mode**: Code within bead scope. Doc-sync mandatory. ADR compliance required.

## Human Gates

- Spec approval → Plan Mode
- Plan approval → Implementation Mode
- ADR divergence → stop and inform user
- Domain operations → human approval required
- Scope expansion → human approval required

## Offline Fallback

If `mindspec instruct` is unavailable, refer to:
- [AGENTS.md](../../AGENTS.md) — Full agent behavioral rules
- [MODES.md](../../docs/core/MODES.md) — Full mode definitions
- [CONVENTIONS.md](../../docs/core/CONVENTIONS.md) — File organization
