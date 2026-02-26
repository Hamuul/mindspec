package explore

import (
	"errors"
	"fmt"

	"github.com/mindspec/mindspec/internal/specinit"
	"github.com/mindspec/mindspec/internal/state"
)

// Enter validates the current state is idle (or absent) and transitions to explore mode.
func Enter(root, description string) error {
	s, err := state.Read(root)
	if err != nil && !errors.Is(err, state.ErrNoState) {
		return fmt.Errorf("reading state: %w", err)
	}

	if s != nil && s.Mode != state.ModeIdle {
		return fmt.Errorf("cannot enter explore mode: currently in %q mode (must be idle)", s.Mode)
	}

	return state.SetMode(root, state.ModeExplore, "", "")
}

// Dismiss validates the current state is explore and transitions to idle.
func Dismiss(root string) error {
	s, err := state.Read(root)
	if err != nil {
		return fmt.Errorf("reading state: %w", err)
	}

	if s.Mode != state.ModeExplore {
		return fmt.Errorf("cannot dismiss: not in explore mode (currently %q)", s.Mode)
	}

	return state.SetMode(root, state.ModeIdle, "", "")
}

// Promote validates the current state is explore and delegates to spec-init.
// specinit.Run handles the state transition to spec mode and molecule creation.
func Promote(root, specID, title string) error {
	s, err := state.Read(root)
	if err != nil {
		return fmt.Errorf("reading state: %w", err)
	}

	if s.Mode != state.ModeExplore {
		return fmt.Errorf("cannot promote: not in explore mode (currently %q)", s.Mode)
	}

	_, err = specinit.Run(root, specID, title)
	return err
}
