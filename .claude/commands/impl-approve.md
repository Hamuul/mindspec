---
description: Approve implementation and close out the spec lifecycle
---

# Implementation Approval

1. Identify the active spec via `mindspec state show`
2. If not in review mode, run `mindspec complete` first to transition
3. Run `mindspec approve impl <id>` (verifies review mode, transitions to idle, emits guidance)
4. If approval fails, show the error and help the user resolve it
5. On success: run the session close protocol as the LAST step (after idle, after recordings stop):
   - `bd sync`
   - `git add` all changed files (state, specs, recordings, beads)
   - `git commit`
   - `bd sync`
   - `git push`
