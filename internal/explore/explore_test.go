package explore

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestProject(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	// Create .mindspec marker dir
	os.MkdirAll(filepath.Join(tmp, ".mindspec"), 0755)
	// Create docs/specs for spec-init
	os.MkdirAll(filepath.Join(tmp, "docs", "specs"), 0755)
	return tmp
}

func TestEnter_NoOp(t *testing.T) {
	root := setupTestProject(t)

	if err := Enter(root, "test idea"); err != nil {
		t.Fatalf("Enter failed: %v", err)
	}
	// Enter is now a no-op — no state change expected
}

func TestEnter_AlwaysSucceeds(t *testing.T) {
	root := setupTestProject(t)

	// Enter should succeed regardless of current state
	if err := Enter(root, "test idea"); err != nil {
		t.Fatalf("Enter failed: %v", err)
	}

	// Call again — still succeeds (no state check)
	if err := Enter(root, "another idea"); err != nil {
		t.Fatalf("second Enter failed: %v", err)
	}
}

func TestDismiss_NoOp(t *testing.T) {
	root := setupTestProject(t)

	if err := Dismiss(root); err != nil {
		t.Fatalf("Dismiss failed: %v", err)
	}
	// Dismiss is now a no-op — no state change expected
}

func TestDismiss_AlwaysSucceeds(t *testing.T) {
	root := setupTestProject(t)

	// Dismiss should succeed regardless of current state
	if err := Dismiss(root); err != nil {
		t.Fatalf("Dismiss failed: %v", err)
	}
}
