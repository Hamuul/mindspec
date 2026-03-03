package resolve

import (
	"fmt"
	"strings"

	"github.com/mindspec/mindspec/internal/phase"
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
//  2. Derive from beads context (worktree path + beads query).
//  3. Query active specs; if exactly one, auto-select it.
//  4. If multiple active specs exist, return ErrAmbiguousTarget.
//  5. If nothing found, return an error.
func ResolveTarget(root, specFlag string) (string, error) {
	// Explicit target
	if specFlag != "" {
		return specFlag, nil
	}

	// Beads-derived context: resolve from root directory + beads query.
	ctx, err := phase.ResolveContextFromDir(root, root)
	if err == nil && ctx != nil && ctx.SpecID != "" {
		return ctx.SpecID, nil
	}

	// Query active specs (no context available)
	active, err := ActiveSpecs(root)
	if err == nil {
		switch len(active) {
		case 1:
			return active[0].SpecID, nil
		default:
			if len(active) > 1 {
				return "", &ErrAmbiguousTarget{Active: active}
			}
		}
	}

	return "", fmt.Errorf("no active specs found; use --spec flag")
}
