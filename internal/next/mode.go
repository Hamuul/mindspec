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
//
// Tries bracket-prefix convention first:
//
//	"[IMPL 009-feature.1] Chunk title" → "009-feature"
//	"[SPEC 008b-gates] Feature"        → "008b-gates"
//	"[PLAN 009-feature] Plan decomp"   → "009-feature"
//
// Falls back to legacy colon convention:
//
//	"005-next: Implement work selection" → "005-next"
func parseSpecID(title string) string {
	// Try bracket-prefix: [TAG specID] or [TAG specID.chunk]
	if strings.HasPrefix(title, "[") {
		closeBracket := strings.Index(title, "]")
		if closeBracket > 0 {
			inner := title[1:closeBracket] // e.g. "IMPL 009-feature.1"
			spaceIdx := strings.Index(inner, " ")
			if spaceIdx > 0 {
				slug := inner[spaceIdx+1:] // e.g. "009-feature.1"
				// Strip chunk suffix (everything from the last dot if it's followed by digits)
				if dotIdx := strings.LastIndex(slug, "."); dotIdx > 0 {
					suffix := slug[dotIdx+1:]
					allDigits := true
					for _, c := range suffix {
						if c < '0' || c > '9' {
							allDigits = false
							break
						}
					}
					if allDigits {
						slug = slug[:dotIdx]
					}
				}
				return slug
			}
		}
	}

	// Fallback: colon convention
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
