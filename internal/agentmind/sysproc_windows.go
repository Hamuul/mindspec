//go:build windows

package agentmind

import "os/exec"

func detachProcess(cmd *exec.Cmd) {
	// Setsid is not available on Windows; the process is already detached
	// when started without inheriting stdio handles.
}
