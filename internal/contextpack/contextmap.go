package contextpack

import (
	"bufio"
	"os"
	"regexp"
	"sort"
	"strings"
)

// Relationship represents a directed relationship in the Context Map.
type Relationship struct {
	From      string
	To        string
	Direction string
	Contract  string
}

// headingRe matches relationship headings like "### Core → Context-System (upstream)"
// Supports both → and ->
var headingRe = regexp.MustCompile(`^###\s+(.+?)\s*(?:→|->)\s*(.+?)\s*\(([^)]+)\)\s*$`)

// contractRe matches "**Contract**: [text](path)" or "**Contract**: path"
var contractRe = regexp.MustCompile(`\*\*Contract\*\*:\s*(?:\[.*?\]\(([^)]+)\)|(\S+))`)

// ParseContextMap reads a context-map.md file and extracts relationships.
func ParseContextMap(path string) ([]Relationship, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var rels []Relationship
	scanner := bufio.NewScanner(f)

	inRelationships := false
	var current *Relationship

	for scanner.Scan() {
		line := scanner.Text()

		// Detect ## Relationships section
		if strings.HasPrefix(line, "## ") {
			heading := strings.TrimSpace(strings.TrimPrefix(line, "## "))
			if strings.EqualFold(heading, "Relationships") {
				inRelationships = true
				continue
			}
			if inRelationships {
				// Hit next ## section, stop
				break
			}
			continue
		}

		if !inRelationships {
			continue
		}

		// Check for relationship heading
		if m := headingRe.FindStringSubmatch(line); m != nil {
			// Save previous if exists
			if current != nil {
				rels = append(rels, *current)
			}
			current = &Relationship{
				From:      normalizeDomain(m[1]),
				To:        normalizeDomain(m[2]),
				Direction: strings.TrimSpace(m[3]),
			}
			continue
		}

		// Check for contract line within current relationship
		if current != nil {
			if m := contractRe.FindStringSubmatch(line); m != nil {
				contract := m[1]
				if contract == "" {
					contract = m[2]
				}
				if current.Contract != "" {
					current.Contract += ", "
				}
				current.Contract += contract
			}
		}
	}

	// Don't forget the last one
	if current != nil {
		rels = append(rels, *current)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return rels, nil
}

// ResolveNeighbors returns domain names that are 1-hop neighbors of the impacted domains,
// excluding the impacted domains themselves.
func ResolveNeighbors(rels []Relationship, impactedDomains []string) []string {
	impacted := make(map[string]bool)
	for _, d := range impactedDomains {
		impacted[normalizeDomain(d)] = true
	}

	neighbors := make(map[string]bool)
	for _, r := range rels {
		fromNorm := normalizeDomain(r.From)
		toNorm := normalizeDomain(r.To)

		if impacted[fromNorm] && !impacted[toNorm] {
			neighbors[toNorm] = true
		}
		if impacted[toNorm] && !impacted[fromNorm] {
			neighbors[fromNorm] = true
		}
	}

	result := make([]string, 0, len(neighbors))
	for d := range neighbors {
		result = append(result, d)
	}
	sort.Strings(result)
	return result
}

// normalizeDomain lowercases and trims a domain name.
func normalizeDomain(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
