---
description: Check the current MindSpec mode and active specification
---

# Spec Status Workflow

Use this workflow to check the current operational mode and active specification.

## Trigger

User invokes `/spec-status` or asks about current mode/spec.

## Steps

### 1. Read MindSpec State

Run the following commands to get the current state:

```bash
mindspec state show
mindspec instruct
```

`mindspec state show` prints the raw state (mode, active spec, active bead).
`mindspec instruct` provides full mode-appropriate guidance including any drift warnings.

### 2. Report Status

Present the output from `mindspec instruct` to the user. If there are warnings (state drift, worktree mismatch), highlight them prominently.

### 3. Quick Summary

Provide a brief summary:

> **Mode**: <mode from state show>
>
> **Active Spec**: `<activeSpec>` (or "none")
>
> **Active Bead**: `<activeBead>` (or "none")

### 4. List Recent Specs (Optional)

If user asks, list specs in `docs/specs/`:

| Spec ID | Status | Domains | Criteria |
|:--------|:-------|:--------|:---------|
| 001-skeleton | APPROVED | core | 5 defined |
| 002-glossary | APPROVED | context-system | 5 defined |

---

## Notes

- This workflow is read-only; it doesn't change state
- If `state.json` is missing, `mindspec instruct` will fall back gracefully with a warning
- Use `/spec-init` to start a new spec
- Use `/spec-approve` to transition Spec -> Plan
- Use `/plan-approve` to transition Plan -> Implementation
