---
description: Enter, promote, or dismiss an Explore Mode session
---

# Explore

1. If the user provides a description, run `mindspec explore "description"` to enter Explore Mode
2. If the user wants to promote, ask for a spec ID (check `docs/specs/` for next available number) and run `mindspec explore promote <spec-id>`
3. If the user wants to dismiss, run `mindspec explore dismiss` (optionally with `--adr` to record an ADR)
4. If unclear what the user wants, show the three options and ask
5. On success: relay the CLI output to the user
