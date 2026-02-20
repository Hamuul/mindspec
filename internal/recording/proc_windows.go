//go:build windows

package recording

import "os"

func isProcessAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Windows, FindProcess always succeeds; Kill with nil signal isn't
	// supported, so we just assume the process exists.
	_ = proc
	return true
}

func signalTerminate(proc *os.Process) {
	_ = proc.Kill()
}
