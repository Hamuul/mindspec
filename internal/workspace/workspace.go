package workspace

import (
	"errors"
	"os"
	"path/filepath"
)

// ErrNoRoot is returned when no project root marker is found.
var ErrNoRoot = errors.New("no mindspec project root found (looked for mindspec.md or .git)")

// FindRoot walks up from startDir looking for mindspec.md or .git at each level.
// It checks mindspec.md first, then .git, matching the Python prototype behavior.
func FindRoot(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	for {
		if exists(filepath.Join(dir, "mindspec.md")) {
			return dir, nil
		}
		if exists(filepath.Join(dir, ".git")) {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", ErrNoRoot
}

// DocsDir returns the path to the docs directory under root.
func DocsDir(root string) string {
	return filepath.Join(root, "docs")
}

// GlossaryPath returns the path to GLOSSARY.md under root.
func GlossaryPath(root string) string {
	return filepath.Join(root, "GLOSSARY.md")
}

// SpecDir returns the path to a specific spec directory under root.
func SpecDir(root, specID string) string {
	return filepath.Join(root, "docs", "specs", specID)
}

// ContextMapPath returns the path to docs/context-map.md under root.
func ContextMapPath(root string) string {
	return filepath.Join(root, "docs", "context-map.md")
}

// ADRDir returns the path to docs/adr/ under root.
func ADRDir(root string) string {
	return filepath.Join(root, "docs", "adr")
}

// PoliciesPath returns the path to architecture/policies.yml under root.
func PoliciesPath(root string) string {
	return filepath.Join(root, "architecture", "policies.yml")
}

// DomainDir returns the path to a specific domain's doc directory under root.
func DomainDir(root, domain string) string {
	return filepath.Join(root, "docs", "domains", domain)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
