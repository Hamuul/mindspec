---
description: Request approval to transition from Plan Mode to Implementation Mode
---

# Plan Approval

1. Identify the active spec/plan via `mindspec state show`
2. Summarize the plan and ask: "Do you approve this plan for implementation?"
3. On **yes**: run `mindspec approve plan <id>` (validates, creates beads, resolves gate, sets state, emits guidance)
4. On **no**: ask what changes the user wants
5. After approval: advise `mindspec next` to claim the first bead
