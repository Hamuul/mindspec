---
description: Request approval to transition from Spec Mode to Plan Mode
---

# Spec Approval

1. Identify the active spec via `mindspec state show`
2. Summarize the spec and ask: "Do you approve this spec for planning?"
3. On **yes**: run `mindspec approve spec <id>` (validates, creates beads, resolves gate, generates context pack, sets state, emits guidance)
4. On **no**: ask what changes the user wants
