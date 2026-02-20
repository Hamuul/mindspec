//go:build !windows

package recording

import (
	"os"
	"syscall"
)

func isProcessAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}

func signalTerminate(proc *os.Process) {
	_ = proc.Signal(syscall.SIGTERM)
}
