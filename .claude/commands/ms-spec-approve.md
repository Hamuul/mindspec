---
description: Approve a spec and transition to Plan Mode
---

# Spec Approval

1. Identify the active spec via `mindspec state show`
2. Run `mindspec approve spec <id>` (validates, closes the spec-approve molecule step, generates context pack, sets state, emits guidance)
3. If approval fails, show the validation errors and help the user fix them
4. On success: immediately begin planning (the approval is the authorization)
