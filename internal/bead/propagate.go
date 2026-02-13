package bead

import (
	"fmt"
	"io"
	"os"
)

// stderr is an io.Writer for warning output; overridable in tests.
var stderr io.Writer = os.Stderr

// PropagateStart moves parent PLAN and SPEC beads to in_progress when an impl
// bead is claimed. Best-effort: errors are logged to stderr, never returned.
func PropagateStart(specID string) {
	// Move PLAN bead to in_progress if still open
	planBeads, err := SearchAny("[PLAN " + specID + "]")
	if err != nil {
		fmt.Fprintf(stderr, "Warning: propagate start — PLAN search failed: %v\n", err)
	} else if len(planBeads) > 0 && planBeads[0].Status == "open" {
		if err := Update(planBeads[0].ID, "in_progress"); err != nil {
			fmt.Fprintf(stderr, "Warning: propagate start — PLAN update failed: %v\n", err)
		}
	}

	// Move SPEC bead to in_progress if still open
	specBeads, err := SearchAny("[SPEC " + specID + "]")
	if err != nil {
		fmt.Fprintf(stderr, "Warning: propagate start — SPEC search failed: %v\n", err)
	} else if len(specBeads) > 0 && specBeads[0].Status == "open" {
		if err := Update(specBeads[0].ID, "in_progress"); err != nil {
			fmt.Fprintf(stderr, "Warning: propagate start — SPEC update failed: %v\n", err)
		}
	}
}

// PropagateClose closes parent PLAN and SPEC beads when all impl beads for a
// spec are done (no open or in_progress impl beads remain).
// Best-effort: errors are logged to stderr, never returned.
func PropagateClose(specID string) {
	// Check if any impl beads are still open
	implPrefix := "[IMPL " + specID + "."
	openImpls, err := Search(implPrefix)
	if err != nil {
		fmt.Fprintf(stderr, "Warning: propagate close — IMPL search (open) failed: %v\n", err)
		return
	}
	if len(openImpls) > 0 {
		return // still have open impl beads
	}

	// Also check for in_progress impl beads via SearchAny
	allImpls, err := SearchAny(implPrefix)
	if err != nil {
		fmt.Fprintf(stderr, "Warning: propagate close — IMPL search (any) failed: %v\n", err)
		return
	}
	for _, b := range allImpls {
		if b.Status != "closed" {
			return // still have non-closed impl beads
		}
	}

	// All impl beads are closed — close PLAN and SPEC beads
	planBeads, err := SearchAny("[PLAN " + specID + "]")
	if err != nil {
		fmt.Fprintf(stderr, "Warning: propagate close — PLAN search failed: %v\n", err)
	} else if len(planBeads) > 0 && planBeads[0].Status != "closed" {
		if err := Close(planBeads[0].ID); err != nil {
			fmt.Fprintf(stderr, "Warning: propagate close — PLAN close failed: %v\n", err)
		}
	}

	specBeads, err := SearchAny("[SPEC " + specID + "]")
	if err != nil {
		fmt.Fprintf(stderr, "Warning: propagate close — SPEC search failed: %v\n", err)
	} else if len(specBeads) > 0 && specBeads[0].Status != "closed" {
		if err := Close(specBeads[0].ID); err != nil {
			fmt.Fprintf(stderr, "Warning: propagate close — SPEC close failed: %v\n", err)
		}
	}
}
