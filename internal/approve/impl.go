package approve

import (
	"fmt"

	"github.com/mindspec/mindspec/internal/state"
)

// ImplResult holds the result of implementation approval.
type ImplResult struct {
	SpecID   string
	Warnings []string
}

// ApproveImpl transitions from review mode to idle, completing the spec lifecycle.
func ApproveImpl(root, specID string) (*ImplResult, error) {
	result := &ImplResult{SpecID: specID}

	// Verify current state is review mode for this spec
	s, err := state.Read(root)
	if err != nil {
		return nil, fmt.Errorf("reading state: %w", err)
	}
	if s.Mode != state.ModeReview {
		return nil, fmt.Errorf("expected review mode, got %q", s.Mode)
	}
	if s.ActiveSpec != specID {
		return nil, fmt.Errorf("active spec is %q, not %q", s.ActiveSpec, specID)
	}

	// Transition to idle
	idleState := &state.State{Mode: state.ModeIdle}
	if err := state.Write(root, idleState); err != nil {
		return nil, fmt.Errorf("setting state to idle: %w", err)
	}

	return result, nil
}
