package next

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mindspec/mindspec/internal/state"
)

// ResolvedWork holds the result of mode resolution for a claimed bead.
type ResolvedWork struct {
	Mode   string
	SpecID string
	Bead   BeadInfo
}

// ResolveMode maps a bead's type and artifact state to a MindSpec mode and spec ID.
//
// Mapping:
//   - task, bug → implement
//   - feature → spec (if no approved spec) or plan (if spec approved)
//
// Spec ID is parsed from the bead title prefix before the first colon.
func ResolveMode(root string, bead BeadInfo) ResolvedWork {
	specID := parseSpecID(bead.Title)

	switch bead.IssueType {
	case "task", "bug":
		return ResolvedWork{
			Mode:   state.ModeImplement,
			SpecID: specID,
			Bead:   bead,
		}
	case "feature":
		mode := resolveFeatureMode(root, specID)
		return ResolvedWork{
			Mode:   mode,
			SpecID: specID,
			Bead:   bead,
		}
	default:
		// Unknown type defaults to implement
		return ResolvedWork{
			Mode:   state.ModeImplement,
			SpecID: specID,
			Bead:   bead,
		}
	}
}

// parseSpecID extracts the spec ID from a bead title.
// Convention: title starts with the spec slug followed by a colon.
// e.g., "005-next: Implement work selection" → "005-next"
func parseSpecID(title string) string {
	idx := strings.Index(title, ":")
	if idx < 0 {
		return ""
	}
	return strings.TrimSpace(title[:idx])
}

// resolveFeatureMode checks the spec's artifact state to determine if we're in
// spec mode (draft spec) or plan mode (approved spec with plan needed).
func resolveFeatureMode(root, specID string) string {
	if specID == "" {
		return state.ModeSpec
	}

	specPath := filepath.Join(root, "docs", "specs", specID, "spec.md")
	data, err := os.ReadFile(specPath)
	if err != nil {
		return state.ModeSpec
	}

	content := string(data)
	if strings.Contains(content, "Status: APPROVED") || strings.Contains(content, "**Status**: APPROVED") {
		return state.ModePlan
	}

	return state.ModeSpec
}
