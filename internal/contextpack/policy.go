package contextpack

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Policy represents a single policy entry from policies.yml.
type Policy struct {
	ID          string `yaml:"id"`
	Description string `yaml:"description"`
	Severity    string `yaml:"severity"`
	Scope       string `yaml:"scope,omitempty"`
	Mode        string `yaml:"mode,omitempty"`
	Reference   string `yaml:"reference,omitempty"`
}

type policiesFile struct {
	Policies []Policy `yaml:"policies"`
}

// ParsePolicies reads and parses the policies.yml file.
func ParsePolicies(path string) ([]Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading policies: %w", err)
	}

	// The file has a markdown header line before YAML content.
	// Strip everything before the first "policies:" key.
	content := string(data)
	idx := strings.Index(content, "policies:")
	if idx < 0 {
		return nil, fmt.Errorf("no 'policies:' key found in %s", path)
	}
	content = content[idx:]

	var pf policiesFile
	if err := yaml.Unmarshal([]byte(content), &pf); err != nil {
		return nil, fmt.Errorf("parsing policies YAML: %w", err)
	}

	return pf.Policies, nil
}

// FilterPolicies returns policies that apply to the given mode.
// Policies with an empty Mode field apply to all modes.
func FilterPolicies(policies []Policy, mode string) []Policy {
	var result []Policy
	for _, p := range policies {
		if p.Mode == "" || strings.EqualFold(p.Mode, mode) {
			result = append(result, p)
		}
	}
	return result
}
