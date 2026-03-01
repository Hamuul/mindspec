package main

import (
	"fmt"
	"os"

	"github.com/mindspec/mindspec/internal/guard"
	"github.com/mindspec/mindspec/internal/instruct"
	"github.com/mindspec/mindspec/internal/state"
)

// emitInstruct reads focus and prints mode-appropriate guidance.
// This is the "instruct-tail" convention: every state-changing command
// (approve, next, complete) calls this after transitioning to emit
// guidance for the new mode.
//
// root is the main repo root (for spec dirs and guard). Focus is read
// from the local root (per-worktree focus).
func emitInstruct(root string) error {
	// Read focus from local root (per-worktree).
	localRoot, localErr := findLocalRoot()
	if localErr != nil {
		localRoot = root
	}
	mc, err := state.ReadFocus(localRoot)
	if err != nil || mc == nil {
		// Fallback to main-root focus when local worktree focus is absent.
		// This happens on freshly created worktrees before focus is copied.
		if localRoot != root {
			if rootFocus, rootErr := state.ReadFocus(root); rootErr == nil && rootFocus != nil {
				mc = rootFocus
			}
		}
	}
	if mc == nil {
		mc = &state.Focus{Mode: state.ModeIdle}
	}

	// CWD redirect: if on main with active worktree, emit redirect only.
	if wtPath := guard.ActiveWorktreePath(root); wtPath != "" && guard.IsMainCWD(root) {
		fmt.Fprintf(os.Stdout, "\n## Worktree Redirect\n\nYou are in the main worktree. Switch to:\n\n  cd %s\n\nThen run `mindspec instruct` for mode-appropriate guidance.\n", wtPath)
		return nil
	}

	ctx := instruct.BuildContext(root, mc)

	// Add worktree check when an active worktree is set.
	if mc.ActiveWorktree != "" {
		if warning := instruct.CheckWorktree(mc.ActiveWorktree); warning != "" {
			ctx.Warnings = append(ctx.Warnings, "[worktree] "+warning)
		}
	}

	output, err := instruct.Render(ctx)
	if err != nil {
		return fmt.Errorf("rendering guidance: %w", err)
	}

	fmt.Fprint(os.Stdout, output)
	return nil
}
