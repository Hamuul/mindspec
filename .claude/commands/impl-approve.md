---
description: Approve implementation and close out the spec lifecycle
---

# Implementation Approval

1. Identify the active spec via `mindspec state show`
2. Run `mindspec approve impl <id>` (verifies review mode, transitions to idle, emits guidance)
3. If approval fails, show the error and help the user resolve it
