package resolve

import (
	"fmt"
	"strings"

	"github.com/mindspec/mindspec/internal/state"
)

// ErrAmbiguousTarget is returned when multiple active specs exist and no --spec was provided.
type ErrAmbiguousTarget struct {
	Active []SpecStatus
}

func (e *ErrAmbiguousTarget) Error() string {
	var sb strings.Builder
	sb.WriteString("multiple active specs found; use --spec to target one:\n")
	for _, s := range e.Active {
		sb.WriteString(fmt.Sprintf("  %s  (mode: %s)\n", s.SpecID, s.Mode))
	}
	return sb.String()
}

// ResolveTarget determines which spec to operate on.
//
// Resolution order:
//  1. If specFlag is provided (from --spec), use it directly.
//  2. Query active specs; if exactly one, auto-select it.
//  3. If multiple active specs exist, return ErrAmbiguousTarget.
//  4. If no active specs, fall back to state.json cursor (if any).
//  5. If nothing found, return an error.
func ResolveTarget(root, specFlag string) (string, error) {
	// Explicit target
	if specFlag != "" {
		return specFlag, nil
	}

	// Query active specs
	active, err := ActiveSpecs(root)
	if err != nil {
		// If resolver fails (e.g. beads unavailable), fall back to state cursor
		return fallbackToCursor(root)
	}

	switch len(active) {
	case 0:
		return fallbackToCursor(root)
	case 1:
		return active[0].SpecID, nil
	default:
		return "", &ErrAmbiguousTarget{Active: active}
	}
}

// fallbackToCursor reads the activeSpec from state.json as a last resort.
func fallbackToCursor(root string) (string, error) {
	s, err := state.Read(root)
	if err != nil {
		return "", fmt.Errorf("no active specs found and no state cursor available")
	}
	if s.ActiveSpec == "" {
		return "", fmt.Errorf("no active specs found and state cursor has no active spec")
	}
	return s.ActiveSpec, nil
}
