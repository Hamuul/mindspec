package recording

import "fmt"

// HealthStatus represents the state of a recording's collector.
type HealthStatus int

const (
	HealthNoRecording HealthStatus = iota
	HealthAlive
	HealthDead
)

// HealthCheck checks if the collector process is alive.
func HealthCheck(root, specID string) (HealthStatus, error) {
	if !IsEnabled(root) {
		return HealthNoRecording, nil
	}
	if !HasRecording(root, specID) {
		return HealthNoRecording, nil
	}

	m, err := ReadManifest(root, specID)
	if err != nil {
		return HealthNoRecording, err
	}

	if m.Status != "recording" || m.CollectorPID <= 0 {
		return HealthNoRecording, nil
	}

	if isProcessAlive(m.CollectorPID) {
		return HealthAlive, nil
	}
	return HealthDead, nil
}

// RestartIfDead restarts the collector if the manifest says recording but the PID is dead.
func RestartIfDead(root, specID string) error {
	status, err := HealthCheck(root, specID)
	if err != nil || status != HealthDead {
		return err
	}

	fmt.Println("Recording collector is dead — restarting...")
	return StartCollector(root, specID)
}

// isProcessAlive is defined in proc_unix.go / proc_windows.go.
