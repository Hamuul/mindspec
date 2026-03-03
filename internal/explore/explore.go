package explore

import (
	"github.com/mindspec/mindspec/internal/specinit"
)

// Enter is a no-op — explore no longer changes state.
// It exists for backward compatibility with callers.
func Enter(root, description string) error {
	return nil
}

// Dismiss is a no-op — explore no longer changes state.
// It exists for backward compatibility with callers.
func Dismiss(root string) error {
	return nil
}

// Promote delegates to specinit.Run to create a spec from exploration.
// No mode check — explore is no longer a distinct mode.
func Promote(root, specID, title string) error {
	_, err := specinit.Run(root, specID, title)
	return err
}
