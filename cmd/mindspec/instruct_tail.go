package main

import (
	"fmt"
	"os"

	"github.com/mindspec/mindspec/internal/guard"
	"github.com/mindspec/mindspec/internal/instruct"
	"github.com/mindspec/mindspec/internal/state"
)

// emitInstruct reads current state and prints mode-appropriate guidance.
// This is the "instruct-tail" convention: every state-changing command
// (approve, next, complete) calls this after transitioning to emit
// guidance for the new mode.
func emitInstruct(root string) error {
	s, err := state.Read(root)
	if err != nil {
		if err == state.ErrNoState {
			s = &state.State{Mode: state.ModeIdle}
		} else {
			return fmt.Errorf("reading state: %w", err)
		}
	}

	// CWD redirect: if on main with active worktree, emit redirect only.
	if wtPath := guard.ActiveWorktreePath(root); wtPath != "" && guard.IsMainCWD(root) {
		fmt.Fprintf(os.Stdout, "\n## Worktree Redirect\n\nYou are in the main worktree. Switch to:\n\n  cd %s\n\nThen run `mindspec instruct` for mode-appropriate guidance.\n", wtPath)
		return nil
	}

	ctx := instruct.BuildContext(root, s)

	// Add worktree check when an active worktree is set.
	if s.ActiveWorktree != "" {
		if warning := instruct.CheckWorktree(s.ActiveWorktree); warning != "" {
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
