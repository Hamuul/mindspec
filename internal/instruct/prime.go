package instruct

import (
	"os/exec"
	"strings"
)

// execPrime is the function used to create the bd prime command.
// Override in tests to mock bd prime output.
var execPrime = func() *exec.Cmd {
	return exec.Command("bd", "prime")
}

// CapturePrime runs `bd prime` and returns its output.
// Returns empty string on any error (graceful degradation).
func CapturePrime() string {
	cmd := execPrime()
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
