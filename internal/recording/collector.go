package recording

import (
	"fmt"
	"os"
	"time"

	"github.com/mindspec/mindspec/internal/agentmind"
)

// StartCollector launches AgentMind as a detached background process
// to collect OTLP telemetry and write NDJSON to the spec's recording directory.
func StartCollector(root, specID string) error {
	eventsPath := EventsPath(root, specID)

	pid, err := agentmind.AutoStart(root, agentmind.DefaultOTLPPort, agentmind.DefaultUIPort, eventsPath)
	if err != nil {
		return fmt.Errorf("starting AgentMind collector: %w", err)
	}

	// Update manifest with PID and port
	m, err := ReadManifest(root, specID)
	if err != nil {
		return fmt.Errorf("reading manifest for PID update: %w", err)
	}
	m.CollectorPID = pid
	m.CollectorPort = agentmind.DefaultOTLPPort
	m.Status = "recording"
	if err := WriteManifest(root, specID, m); err != nil {
		return fmt.Errorf("writing manifest with PID: %w", err)
	}

	return nil
}

// StopCollector sends SIGTERM to the AgentMind process and updates the manifest.
func StopCollector(root, specID string) error {
	m, err := ReadManifest(root, specID)
	if err != nil {
		return fmt.Errorf("reading manifest: %w", err)
	}

	if m.CollectorPID > 0 {
		proc, err := os.FindProcess(m.CollectorPID)
		if err == nil {
			signalTerminate(proc)
			// Give it a moment to shut down gracefully
			time.Sleep(500 * time.Millisecond)
		}
	}

	m.Status = "complete"
	m.CollectorPID = 0

	// Close the last phase
	if len(m.Phases) > 0 {
		last := &m.Phases[len(m.Phases)-1]
		if last.EndedAt == "" {
			last.EndedAt = time.Now().UTC().Format(time.RFC3339)
		}
	}

	return WriteManifest(root, specID, m)
}
