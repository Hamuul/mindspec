package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const preCommitScript = `#!/usr/bin/env bash
# MindSpec pre-commit hook (Layer 1 enforcement — ADR-0019)
# Prevents commits on protected branches when mindspec is active.

# Escape hatch: MINDSPEC_ALLOW_MAIN=1 git commit
if [ "${MINDSPEC_ALLOW_MAIN:-}" = "1" ]; then
  exit 0
fi

# Read focus — if no cache file, allow commit
MODE_CACHE=".mindspec/focus"
if [ ! -f "$MODE_CACHE" ]; then
  exit 0
fi

MODE=$(cat "$MODE_CACHE" 2>/dev/null | grep -o '"mode"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"mode"[[:space:]]*:[[:space:]]*"//;s/"$//')
if [ -z "$MODE" ] || [ "$MODE" = "idle" ]; then
  exit 0
fi

# Check enforcement config
CONFIG_FILE=".mindspec/config.yaml"
if [ -f "$CONFIG_FILE" ]; then
  if grep -q 'pre_commit_hook.*false' "$CONFIG_FILE" 2>/dev/null; then
    exit 0
  fi
fi

# Get current branch
BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null)
if [ -z "$BRANCH" ]; then
  exit 0
fi

# Read protected branches from config (default: main, master)
PROTECTED="main master"
if [ -f "$CONFIG_FILE" ]; then
  CUSTOM=$(grep -A5 'protected_branches' "$CONFIG_FILE" 2>/dev/null | grep '^\s*-' | sed 's/^[[:space:]]*-[[:space:]]*//' | tr '\n' ' ')
  if [ -n "$CUSTOM" ]; then
    PROTECTED="$CUSTOM"
  fi
fi

# Check if current branch is protected
for p in $PROTECTED; do
  if [ "$BRANCH" = "$p" ]; then
    WORKTREE=$(cat "$MODE_CACHE" 2>/dev/null | grep -o '"activeWorktree"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"activeWorktree"[[:space:]]*:[[:space:]]*"//;s/"$//')
    echo "mindspec: commits on '$BRANCH' are blocked while mindspec is active (mode: $MODE)." >&2
    if [ -n "$WORKTREE" ]; then
      echo "  Switch to your worktree: cd $WORKTREE" >&2
    fi
    echo "  Escape hatch: MINDSPEC_ALLOW_MAIN=1 git commit ..." >&2
    exit 1
  fi
done

exit 0
`

const postCheckoutScript = `#!/usr/bin/env bash
# MindSpec post-checkout hook v1 (Layer 1 enforcement)
# Prevents the main worktree from checking out non-protected branches.
# post-checkout receives: $1=prev_ref $2=new_ref $3=is_branch_checkout

# Only act on branch checkouts (flag=1), not file checkouts
if [ "$3" != "1" ]; then exit 0; fi

# Escape hatch
if [ "${MINDSPEC_ALLOW_MAIN:-}" = "1" ]; then exit 0; fi

# Only enforce in main worktree (git-common-dir == git-dir means main worktree)
GIT_DIR=$(git rev-parse --git-dir 2>/dev/null)
COMMON_DIR=$(git rev-parse --git-common-dir 2>/dev/null)
if [ "$GIT_DIR" != "$COMMON_DIR" ]; then exit 0; fi

# Must have mindspec focus file
MODE_CACHE=".mindspec/focus"
if [ ! -f "$MODE_CACHE" ]; then exit 0; fi

# Check enforcement config
CONFIG_FILE=".mindspec/config.yaml"
if [ -f "$CONFIG_FILE" ]; then
  if grep -q 'pre_commit_hook.*false' "$CONFIG_FILE" 2>/dev/null; then
    exit 0
  fi
fi

# Get current branch
BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null)
if [ -z "$BRANCH" ]; then exit 0; fi

# Read protected branches
PROTECTED="main master"
if [ -f "$CONFIG_FILE" ]; then
  CUSTOM=$(grep -A5 'protected_branches' "$CONFIG_FILE" 2>/dev/null | grep '^\s*-' | sed 's/^[[:space:]]*-[[:space:]]*//' | tr '\n' ' ')
  if [ -n "$CUSTOM" ]; then PROTECTED="$CUSTOM"; fi
fi

# If branch is protected, allow
for p in $PROTECTED; do
  if [ "$BRANCH" = "$p" ]; then exit 0; fi
done

# Find which protected branch to revert to (first one that exists)
REVERT_TO=""
for p in $PROTECTED; do
  if git rev-parse --verify "$p" >/dev/null 2>&1; then
    REVERT_TO="$p"; break
  fi
done
if [ -z "$REVERT_TO" ]; then exit 0; fi

# Auto-revert: switch back to protected branch
echo "mindspec: checkout blocked — main worktree must stay on '$REVERT_TO'." >&2
echo "  Use: git worktree add ../<name> -b $BRANCH" >&2
echo "  Escape hatch: MINDSPEC_ALLOW_MAIN=1 git checkout $BRANCH" >&2
git checkout "$REVERT_TO" --quiet 2>/dev/null
exit 1
`

// InstallPreCommit installs the MindSpec pre-commit hook.
// It uses the git hooks path and chains with any existing pre-commit hook.
func InstallPreCommit(root string) error {
	hooksDir := filepath.Join(root, ".git", "hooks")
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		// Not a git repo or bare repo — skip
		return nil
	}

	hookPath := filepath.Join(hooksDir, "pre-commit")
	marker := "# MindSpec pre-commit hook"

	// Check if already installed
	if data, err := os.ReadFile(hookPath); err == nil {
		if strings.Contains(string(data), marker) {
			return nil // already installed
		}
		// Existing hook — chain by renaming and calling it
		backupPath := hookPath + ".pre-mindspec"
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			if err := os.Rename(hookPath, backupPath); err != nil {
				return fmt.Errorf("backing up existing pre-commit hook: %w", err)
			}
		}
		// Write chained hook
		chained := preCommitScript + "\n# Chain to previous hook\nif [ -x .git/hooks/pre-commit.pre-mindspec ]; then\n  .git/hooks/pre-commit.pre-mindspec\nfi\n"
		return os.WriteFile(hookPath, []byte(chained), 0755)
	}

	// No existing hook — write directly
	return os.WriteFile(hookPath, []byte(preCommitScript), 0755)
}

// InstallPostCheckout installs the MindSpec post-checkout hook.
// It uses the git hooks path and chains with any existing post-checkout hook.
func InstallPostCheckout(root string) error {
	hooksDir := filepath.Join(root, ".git", "hooks")
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		return nil
	}

	hookPath := filepath.Join(hooksDir, "post-checkout")
	marker := "# MindSpec post-checkout hook"

	if data, err := os.ReadFile(hookPath); err == nil {
		if strings.Contains(string(data), marker) {
			return nil // already installed
		}
		backupPath := hookPath + ".pre-mindspec"
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			if err := os.Rename(hookPath, backupPath); err != nil {
				return fmt.Errorf("backing up existing post-checkout hook: %w", err)
			}
		}
		chained := postCheckoutScript + "\n# Chain to previous hook\nif [ -x .git/hooks/post-checkout.pre-mindspec ]; then\n  .git/hooks/post-checkout.pre-mindspec \"$@\"\nfi\n"
		return os.WriteFile(hookPath, []byte(chained), 0755)
	}

	return os.WriteFile(hookPath, []byte(postCheckoutScript), 0755)
}

// InstallAll installs all MindSpec git hooks (pre-commit and post-checkout).
func InstallAll(root string) error {
	if err := InstallPreCommit(root); err != nil {
		return err
	}
	return InstallPostCheckout(root)
}
