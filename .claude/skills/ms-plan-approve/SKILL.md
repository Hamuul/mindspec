---
name: ms-plan-approve
description: Approve a plan and transition toward Implementation Mode
---

# Plan Approval

1. Identify the active spec/plan via `mindspec state show`
2. Run `mindspec approve plan <id>` in the terminal (validates, closes the plan-approve gate, sets state, emits guidance)
3. If approval fails, show the validation errors and help the user fix them
4. On success: **stop**. Tell the user to run `/clear` then `mindspec next` to claim their first bead with a fresh context
