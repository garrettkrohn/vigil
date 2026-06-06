package monitor

import (
	"testing"
	"vigil/config"
)

func TestGetAllServiceStatuses(t *testing.T) {
	t.Run("should build service status from config and live data", func(t *testing.T) {
		cfg := &config.Config{
			Services: []config.Service{
				{Name: "test-api", Ports: []int{99998}},
			},
		}

		statuses := GetAllServiceStatuses(cfg)

		// Should have one status per port
		if len(statuses) != 1 {
			t.Errorf("expected 1 status, got %d", len(statuses))
		}
	})

	t.Run("should mark service as stopped if port is inactive", func(t *testing.T) {
		cfg := &config.Config{
			Services: []config.Service{
				{Name: "inactive-service", Ports: []int{99999}},
			},
		}

		statuses := GetAllServiceStatuses(cfg)

		if len(statuses) == 0 {
			t.Fatal("expected at least one status")
		}

		if statuses[0].Running {
			t.Error("expected service to be marked as not running")
		}
	})

	t.Run("should handle multiple ports per service", func(t *testing.T) {
		cfg := &config.Config{
			Services: []config.Service{
				{Name: "multi-port", Ports: []int{99997, 99996, 99995}},
			},
		}

		statuses := GetAllServiceStatuses(cfg)

		// Should have one status per port
		if len(statuses) != 3 {
			t.Errorf("expected 3 statuses, got %d", len(statuses))
		}

		// All should have the same service name
		for _, status := range statuses {
			if status.Name != "multi-port" {
				t.Errorf("expected name 'multi-port', got %s", status.Name)
			}
		}
	})

	t.Run("should handle service with no tmux session (external process)", func(t *testing.T) {
		cfg := &config.Config{
			Services: []config.Service{
				{Name: "external", Ports: []int{99994}},
			},
		}

		statuses := GetAllServiceStatuses(cfg)

		if len(statuses) == 0 {
			t.Fatal("expected at least one status")
		}

		// External process should have empty session
		if statuses[0].Session != "" {
			t.Errorf("expected empty session for non-tmux process, got %s", statuses[0].Session)
		}
	})
}

func TestRefreshStatuses(t *testing.T) {
	t.Run("should refresh service status on demand", func(t *testing.T) {
		cfg := &config.Config{
			Services: []config.Service{
				{Name: "refresh-test", Ports: []int{99993}},
			},
		}

		statuses1 := GetAllServiceStatuses(cfg)
		statuses2 := RefreshStatuses(cfg)

		// Both should have same length
		if len(statuses1) != len(statuses2) {
			t.Errorf("expected same number of statuses, got %d and %d", len(statuses1), len(statuses2))
		}
	})
}

func TestServiceStatusStructure(t *testing.T) {
	t.Run("ServiceStatus should have correct fields", func(t *testing.T) {
		status := ServiceStatus{
			Name:    "test",
			Port:    8080,
			Running: true,
			PID:     12345,
			Session: "test-session",
		}

		if status.Name != "test" {
			t.Error("Name not set correctly")
		}
		if status.Port != 8080 {
			t.Error("Port not set correctly")
		}
		if !status.Running {
			t.Error("Running not set correctly")
		}
		if status.PID != 12345 {
			t.Error("PID not set correctly")
		}
		if status.Session != "test-session" {
			t.Error("Session not set correctly")
		}
	})
}
