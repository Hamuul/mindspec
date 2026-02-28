package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mindspec/mindspec/internal/workspace"
)

// Valid mode values.
const (
	ModeIdle      = "idle"
	ModeExplore   = "explore"
	ModeSpec      = "spec"
	ModePlan      = "plan"
	ModeImplement = "implement"
	ModeReview    = "review"
)

// ValidModes lists all valid mode values.
var ValidModes = []string{ModeIdle, ModeExplore, ModeSpec, ModePlan, ModeImplement, ModeReview}

// Session holds transient per-session metadata persisted at .mindspec/session.json.
type Session struct {
	SessionSource    string `json:"sessionSource,omitempty"`
	SessionStartedAt string `json:"sessionStartedAt,omitempty"`
	BeadClaimedAt    string `json:"beadClaimedAt,omitempty"`
}

// ModeCache is a write-through cache for hook latency, persisted at .mindspec/mode-cache.
// Lifecycle commands write this after mutating molecule state. Hooks read it
// instead of calling bd mol show on every PreToolUse invocation.
type ModeCache struct {
	Mode           string `json:"mode"`
	ActiveSpec     string `json:"activeSpec,omitempty"`
	ActiveBead     string `json:"activeBead,omitempty"`
	ActiveWorktree string `json:"activeWorktree,omitempty"`
	SpecBranch     string `json:"specBranch,omitempty"`
	Timestamp      string `json:"timestamp"`
}

// ReadSession loads the session from .mindspec/session.json under root.
// Returns a zero Session (no error) if the file does not exist.
func ReadSession(root string) (*Session, error) {
	path := workspace.SessionPath(root)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Session{}, nil
		}
		return nil, fmt.Errorf("reading session file: %w", err)
	}

	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing session file: %w", err)
	}
	return &s, nil
}

// WriteSessionFile persists the session to .mindspec/session.json under root.
func WriteSessionFile(root string, s *Session) error {
	dir := workspace.MindspecDir(root)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating .mindspec directory: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling session: %w", err)
	}
	data = append(data, '\n')

	return os.WriteFile(workspace.SessionPath(root), data, 0644)
}

// ReadModeCache loads the mode cache from .mindspec/mode-cache under root.
// Returns nil (no error) if the file does not exist.
func ReadModeCache(root string) (*ModeCache, error) {
	path := workspace.ModeCachePath(root)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading mode-cache: %w", err)
	}

	var mc ModeCache
	if err := json.Unmarshal(data, &mc); err != nil {
		return nil, fmt.Errorf("parsing mode-cache: %w", err)
	}
	return &mc, nil
}

// WriteModeCache persists the mode cache to .mindspec/mode-cache under root.
func WriteModeCache(root string, mc *ModeCache) error {
	dir := workspace.MindspecDir(root)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating .mindspec directory: %w", err)
	}

	mc.Timestamp = time.Now().UTC().Format(time.RFC3339)

	data, err := json.MarshalIndent(mc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling mode-cache: %w", err)
	}
	data = append(data, '\n')

	return os.WriteFile(workspace.ModeCachePath(root), data, 0644)
}

// SpecBranch returns the canonical branch name for a spec.
func SpecBranch(specID string) string { return "spec/" + specID }

// SpecWorktreePath returns the canonical worktree path for a spec.
func SpecWorktreePath(root, specID string) string {
	return filepath.Join(root, ".worktrees", "worktree-spec-"+specID)
}

// BeadWorktreePath returns the canonical worktree path for a bead
// nested under its spec's worktree.
func BeadWorktreePath(specWorktree, beadID string) string {
	return filepath.Join(specWorktree, ".worktrees", "worktree-"+beadID)
}

// IsValidMode reports whether mode is a recognized lifecycle mode.
func IsValidMode(mode string) bool {
	for _, m := range ValidModes {
		if m == mode {
			return true
		}
	}
	return false
}
