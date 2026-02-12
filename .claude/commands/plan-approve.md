---
description: Request approval to transition from Plan Mode to Implementation Mode
---

# Plan Approval Workflow

Use this workflow to gate the transition from Plan Mode to Implementation Mode.

## Trigger

User invokes `/plan-approve` or expresses readiness to implement.

## Steps

### 1. Identify Active Spec and Plan

Find the active spec and its associated implementation beads. If unclear, ask the user.

### 2. Run Automated Validation

Run `mindspec validate plan <id>` to check structural quality. If there are errors:

> **Plan not ready for approval**
>
> <paste validate output>
>
> Please update the plan and try again.

Remain in Plan Mode.

### 3. Review Plan and Check Coverage

Read the plan and verify:
- All spec requirements are covered by at least one bead
- Each bead has bounded scope (the automated checks cover steps/verification/deps)

### 4. Present Plan Summary

If validation passes, present a summary:

> **Plan Summary for Spec <id>**
>
> **Implementation Beads** (<N> total):
>
> | Bead | Scope | Deps | Verification Steps |
> |:-----|:------|:-----|:-------------------|
> | <bead-1> | <scope> | <deps> | <N steps> |
> | <bead-2> | <scope> | <deps> | <N steps> |
>
> **ADRs Cited**: <list>
>
> **Ready to approve and begin implementation?**

### 5. Request Explicit Approval

Ask the user:

> Do you approve this plan for implementation? (yes/no)

### 6. On Approval

If user approves:

1. Update MindSpec state to Implementation Mode (**before** the milestone commit):
   ```bash
   mindspec state set --mode=implement --spec=<id> --bead=<first-bead-id>
   ```
   Use the first bead with no unresolved dependencies as the initial active bead.

2. Inform user:
   > **Plan approved!**
   >
   > You are now in **Implementation Mode**.
   >
   > **Next steps:**
   > 1. Run `mindspec next` to claim the first ready bead and get guidance
   > 2. Implement within the bead's scope
   > 3. Verify, update docs, and close the bead
   > 4. Run `mindspec next` again for the next bead

### 7. On Rejection

If user declines:

> Plan remains unapproved.
>
> What changes would you like to make?

Remain in Plan Mode.

---

## Notes

- This is a human gate — the user must explicitly confirm
- Each bead should be implementable independently (respecting dependencies)
- Implementation work must happen in isolated worktrees
