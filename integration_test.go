// +build integration

package main

import (
	"os"
	"os/exec"
	"testing"
	"time"
	"vigil/config"
	"vigil/monitor"
	"vigil/process"
	"vigil/tmux"
)

// TestIntegrationFullWorkflow tests the complete vigil workflow with real tmux
func TestIntegrationFullWorkflow(t *testing.T) {
	// Skip if tmux is not available
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux not found, skipping integration tests")
	}

	// Check if we're in a tmux session (required for testing)
	if os.Getenv("TMUX") == "" {
		t.Skip("not in tmux session, skipping integration tests")
	}

	t.Run("should detect service running in tmux session", func(t *testing.T) {
		// Create a temporary config
		cfg := &config.Config{
			Services: []config.Service{
				{Name: "test-service", Ports: []int{18888}},
			},
		}

		// Start a test server in background
		cmd := exec.Command("sh", "-c", "while true; do echo test | nc -l 18888 2>/dev/null || sleep 1; done")
		if err := cmd.Start(); err != nil {
			t.Skipf("Failed to start test server: %v", err)
		}
		defer cmd.Process.Kill()

		// Give it time to start
		time.Sleep(500 * time.Millisecond)

		// Get service status
		statuses := monitor.GetAllServiceStatuses(cfg)

		if len(statuses) == 0 {
			t.Fatal("expected at least one status")
		}

		status := statuses[0]
		if !status.Running {
			t.Error("expected service to be running")
		}

		if status.PID == 0 {
			t.Error("expected non-zero PID")
		}
	})

	t.Run("should handle service not running", func(t *testing.T) {
		cfg := &config.Config{
			Services: []config.Service{
				{Name: "stopped-service", Ports: []int{19999}},
			},
		}

		statuses := monitor.GetAllServiceStatuses(cfg)

		if len(statuses) == 0 {
			t.Fatal("expected at least one status")
		}

		status := statuses[0]
		if status.Running {
			t.Error("expected service to not be running")
		}

		if status.PID != 0 {
			t.Error("expected zero PID for stopped service")
		}
	})

	t.Run("should get tmux sessions", func(t *testing.T) {
		sessions := tmux.GetActiveSessions()

		// Should have at least one session (the current one)
		if len(sessions) == 0 {
			t.Error("expected at least one tmux session")
		}
	})

	t.Run("should map PIDs to sessions", func(t *testing.T) {
		pidMap := tmux.GetAllPIDsWithSessions()

		// Should have at least one entry (current shell)
		if len(pidMap) == 0 {
			t.Error("expected at least one PID in tmux")
		}
	})
}

// TestKillProcess tests killing a process (requires permission)
func TestKillProcess(t *testing.T) {
	t.Run("should kill process and verify it stops", func(t *testing.T) {
		// Start a test process
		cmd := exec.Command("sleep", "3600")
		if err := cmd.Start(); err != nil {
			t.Fatalf("failed to start test process: %v", err)
		}

		pid := cmd.Process.Pid

		// Kill it
		if err := process.KillProcess(pid); err != nil {
			t.Errorf("failed to kill process: %v", err)
		}

		// Wait for process to exit
		done := make(chan error)
		go func() {
			done <- cmd.Wait()
		}()

		select {
		case <-done:
			// Process exited successfully
		case <-time.After(2 * time.Second):
			t.Error("process did not exit within timeout")
		}
	})
}

// TestConfigRoundTrip tests saving and loading config
func TestConfigRoundTrip(t *testing.T) {
	t.Run("should save and load config", func(t *testing.T) {
		tmpFile := t.TempDir() + "/test-config.yaml"

		original := &config.Config{
			Services: []config.Service{
				{Name: "test1", Ports: []int{8080}},
				{Name: "test2", Ports: []int{8081, 9091}},
			},
		}

		// Save
		if err := config.SaveConfigToPath(original, tmpFile); err != nil {
			t.Fatalf("failed to save config: %v", err)
		}

		// Load
		loaded, err := config.LoadConfigFromPath(tmpFile)
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		// Verify
		if len(loaded.Services) != len(original.Services) {
			t.Errorf("expected %d services, got %d", len(original.Services), len(loaded.Services))
		}

		for i, svc := range original.Services {
			if loaded.Services[i].Name != svc.Name {
				t.Errorf("service %d: expected name %s, got %s", i, svc.Name, loaded.Services[i].Name)
			}
			if len(loaded.Services[i].Ports) != len(svc.Ports) {
				t.Errorf("service %d: port count mismatch", i)
			}
		}
	})
}
