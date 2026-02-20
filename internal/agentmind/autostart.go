// Package agentmind provides utilities for managing the AgentMind process —
// the unified OTLP collector and live visualization server.
package agentmind

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"

	"time"
)

const (
	// DefaultOTLPPort is the default OTLP/HTTP receiver port for AgentMind.
	DefaultOTLPPort = 4318
	// DefaultUIPort is the default web UI port for AgentMind.
	DefaultUIPort = 8420
)

// IsRunning checks whether AgentMind (or any OTLP receiver) is listening on the given port.
func IsRunning(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 500*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// AutoStart ensures AgentMind is running. If already listening on otlpPort,
// returns 0 (reusing existing instance). Otherwise starts it as a detached
// background process and returns the PID.
func AutoStart(root string, otlpPort, uiPort int, outputPath string) (int, error) {
	if IsRunning(otlpPort) {
		return 0, nil
	}

	binPath, err := findBinary(root)
	if err != nil {
		return 0, err
	}

	args := []string{"agentmind", "serve",
		"--otlp-port", fmt.Sprintf("%d", otlpPort),
		"--ui-port", fmt.Sprintf("%d", uiPort),
	}
	if outputPath != "" {
		args = append(args, "--output", outputPath)
	}

	cmd := exec.Command(binPath, args...)
	detachProcess(cmd)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("starting AgentMind: %w", err)
	}

	pid := cmd.Process.Pid
	cmd.Process.Release() //nolint:errcheck

	if err := WaitForPort(otlpPort, 5*time.Second); err != nil {
		return pid, fmt.Errorf("AgentMind started (PID %d) but not responding: %w", pid, err)
	}

	fmt.Fprintf(os.Stderr, "AgentMind started — watch live at http://localhost:%d\n", uiPort)
	return pid, nil
}

// WaitForPort polls until a TCP connection to the given port succeeds or timeout expires.
func WaitForPort(port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 200*time.Millisecond)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("port %d not ready after %s", port, timeout)
}

// Probe sends a lightweight HTTP request to check if the OTLP receiver is responding.
func Probe(port int) bool {
	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/v1/logs", port))
	if err != nil {
		return false
	}
	resp.Body.Close()
	// Any response (even 405 Method Not Allowed) means the server is up
	return true
}

// findBinary locates the mindspec binary.
func findBinary(root string) (string, error) {
	binPath := root + "/bin/mindspec"
	if _, err := os.Stat(binPath); err == nil {
		return binPath, nil
	}

	path, err := exec.LookPath("mindspec")
	if err == nil {
		return path, nil
	}

	return "", fmt.Errorf("mindspec binary not found in %s/bin/ or PATH", root)
}
