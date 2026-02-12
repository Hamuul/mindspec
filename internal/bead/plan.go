package bead

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// WorkChunk represents a single work chunk from plan frontmatter.
type WorkChunk struct {
	ID        int      `yaml:"id"`
	Title     string   `yaml:"title"`
	Scope     string   `yaml:"scope"`
	Verify    []string `yaml:"verify"`
	DependsOn []int    `yaml:"depends_on"`
}

// Generated holds machine-written metadata in plan frontmatter.
type Generated struct {
	BeadIDs map[string]string `yaml:"bead_ids,omitempty"`
}

// PlanMeta represents the YAML frontmatter of a plan.md file.
type PlanMeta struct {
	Status     string      `yaml:"status"`
	SpecID     string      `yaml:"spec_id"`
	WorkChunks []WorkChunk `yaml:"work_chunks"`
	Generated  *Generated  `yaml:"generated,omitempty"`
}

// ParsePlanMeta extracts and parses YAML frontmatter from a plan file.
func ParsePlanMeta(planPath string) (*PlanMeta, error) {
	data, err := os.ReadFile(planPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read plan: %v", err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil, fmt.Errorf("no frontmatter found (expected leading ---)")
	}

	var fmLines []string
	found := false
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "---" {
			found = true
			break
		}
		fmLines = append(fmLines, line)
	}

	if !found {
		return nil, fmt.Errorf("unclosed frontmatter (missing closing ---)")
	}

	// Filter out comment lines
	var activeFmLines []string
	for _, line := range fmLines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "#") {
			activeFmLines = append(activeFmLines, line)
		}
	}

	fmContent := strings.Join(activeFmLines, "\n")

	var meta PlanMeta
	if err := yaml.Unmarshal([]byte(fmContent), &meta); err != nil {
		return nil, fmt.Errorf("parsing plan YAML: %w", err)
	}

	return &meta, nil
}

// CreatePlanBeads creates implementation beads from an approved plan.
// Returns a mapping of chunk ID to bead ID. Idempotent.
func CreatePlanBeads(root, specID string) (map[int]string, error) {
	planPath := fmt.Sprintf("%s/docs/specs/%s/plan.md", root, specID)
	meta, err := ParsePlanMeta(planPath)
	if err != nil {
		return nil, err
	}

	// Validate approved
	if !strings.EqualFold(meta.Status, "approved") {
		return nil, fmt.Errorf("plan is not approved (status: %q)", meta.Status)
	}

	// Validate work_chunks present
	if len(meta.WorkChunks) == 0 {
		return nil, fmt.Errorf("plan has no work_chunks defined")
	}

	// Find spec bead as parent
	specPrefix := fmt.Sprintf("[SPEC %s]", specID)
	var parentID string
	specBeads, err := Search(specPrefix)
	if err == nil && len(specBeads) > 0 {
		parentID = specBeads[0].ID
	}

	// Create beads per chunk (idempotent)
	mapping := make(map[int]string)
	for _, chunk := range meta.WorkChunks {
		implPrefix := fmt.Sprintf("[IMPL %s.%d]", specID, chunk.ID)

		// Check for existing
		existing, err := Search(implPrefix)
		if err == nil && len(existing) > 0 {
			mapping[chunk.ID] = existing[0].ID
			continue
		}

		// Build description (capped at 800 chars)
		desc := buildImplDescription(chunk, specID)

		title := fmt.Sprintf("%s %s", implPrefix, chunk.Title)
		bead, err := Create(title, desc, "task", 2, parentID)
		if err != nil {
			return nil, fmt.Errorf("creating bead for chunk %d: %w", chunk.ID, err)
		}
		mapping[chunk.ID] = bead.ID
	}

	// Wire dependencies
	for _, chunk := range meta.WorkChunks {
		for _, depID := range chunk.DependsOn {
			blockedBead, ok := mapping[chunk.ID]
			if !ok {
				continue
			}
			blockerBead, ok := mapping[depID]
			if !ok {
				continue
			}
			if err := DepAdd(blockedBead, blockerBead); err != nil {
				return nil, fmt.Errorf("wiring dep %d->%d: %w", chunk.ID, depID, err)
			}
		}
	}

	return mapping, nil
}

// buildImplDescription creates a structured description for an impl bead.
func buildImplDescription(chunk WorkChunk, specID string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Scope: %s", chunk.Scope)
	if len(chunk.Verify) > 0 {
		sb.WriteString("\nVerify:")
		for _, v := range chunk.Verify {
			fmt.Fprintf(&sb, "\n- %s", v)
		}
	}
	fmt.Fprintf(&sb, "\nPlan: docs/specs/%s/plan.md", specID)

	desc := sb.String()
	if len(desc) > 800 {
		desc = desc[:797] + "..."
	}
	return desc
}

// WriteGeneratedBeadIDs writes bead IDs into the plan frontmatter under generated.bead_ids.
// Preserves existing frontmatter fields via map[string]interface{} round-trip.
func WriteGeneratedBeadIDs(planPath string, mapping map[int]string) error {
	data, err := os.ReadFile(planPath)
	if err != nil {
		return fmt.Errorf("cannot read plan: %v", err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	// Find frontmatter boundaries
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return fmt.Errorf("no frontmatter found")
	}

	fmEndIdx := -1
	for i, line := range lines[1:] {
		if strings.TrimSpace(line) == "---" {
			fmEndIdx = i + 1
			break
		}
	}
	if fmEndIdx == -1 {
		return fmt.Errorf("unclosed frontmatter")
	}

	// Extract frontmatter lines (including comments, which we'll filter for parsing)
	fmLines := lines[1:fmEndIdx]
	var activeFmLines []string
	for _, line := range fmLines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "#") {
			activeFmLines = append(activeFmLines, line)
		}
	}

	// Parse into generic map to preserve all fields
	fmContent := strings.Join(activeFmLines, "\n")
	var fmMap map[string]interface{}
	if err := yaml.Unmarshal([]byte(fmContent), &fmMap); err != nil {
		return fmt.Errorf("parsing frontmatter: %w", err)
	}
	if fmMap == nil {
		fmMap = make(map[string]interface{})
	}

	// Build bead_ids as map with string keys for YAML
	beadIDs := make(map[string]string)
	for chunkID, beadID := range mapping {
		beadIDs[fmt.Sprintf("%d", chunkID)] = beadID
	}

	// Set generated.bead_ids
	gen, ok := fmMap["generated"].(map[string]interface{})
	if !ok {
		gen = make(map[string]interface{})
	}
	gen["bead_ids"] = beadIDs
	fmMap["generated"] = gen

	// Re-marshal frontmatter
	newFm, err := yaml.Marshal(fmMap)
	if err != nil {
		return fmt.Errorf("marshaling frontmatter: %w", err)
	}

	// Splice back: new frontmatter + body after closing ---
	body := strings.Join(lines[fmEndIdx+1:], "\n")
	result := "---\n" + string(newFm) + "---\n" + body

	if err := os.WriteFile(planPath, []byte(result), 0644); err != nil {
		return fmt.Errorf("writing plan: %w", err)
	}

	return nil
}
