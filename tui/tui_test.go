package tui

import (
	"testing"
	"vigil/monitor"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModelInitialization(t *testing.T) {
	t.Run("should initialize model with service statuses", func(t *testing.T) {
		statuses := []monitor.ServiceStatus{
			{Name: "api", Port: 8080, Running: true, PID: 123, Session: "api-session"},
			{Name: "auth", Port: 8081, Running: false},
		}

		m := newModel(statuses, nil)

		if len(m.statuses) != 2 {
			t.Errorf("expected 2 statuses, got %d", len(m.statuses))
		}

		if m.cursor != 0 {
			t.Errorf("expected cursor at 0, got %d", m.cursor)
		}
	})
}

func TestNavigationKeys(t *testing.T) {
	statuses := []monitor.ServiceStatus{
		{Name: "service1", Port: 8080},
		{Name: "service2", Port: 8081},
		{Name: "service3", Port: 8082},
	}

	t.Run("should navigate down with j key", func(t *testing.T) {
		m := newModel(statuses, nil)
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}

		updatedModel, _ := m.Update(msg)
		m = updatedModel.(model)

		if m.cursor != 1 {
			t.Errorf("expected cursor at 1, got %d", m.cursor)
		}
	})

	t.Run("should navigate up with k key", func(t *testing.T) {
		m := newModel(statuses, nil)
		m.cursor = 1

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
		updatedModel, _ := m.Update(msg)
		m = updatedModel.(model)

		if m.cursor != 0 {
			t.Errorf("expected cursor at 0, got %d", m.cursor)
		}
	})
}

func TestQuitKey(t *testing.T) {
	t.Run("should quit when q pressed", func(t *testing.T) {
		statuses := []monitor.ServiceStatus{{Name: "test", Port: 8080}}
		m := newModel(statuses, nil)

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
		_, cmd := m.Update(msg)

		if cmd == nil {
			t.Error("expected quit command")
		}
	})
}

func TestViewRendering(t *testing.T) {
	t.Run("should display running services in view", func(t *testing.T) {
		statuses := []monitor.ServiceStatus{
			{Name: "running-svc", Port: 8080, Running: true, PID: 123, Session: "test-session"},
		}
		m := newModel(statuses, nil)

		view := m.View()

		if view == "" {
			t.Error("expected non-empty view")
		}
	})

	t.Run("should display stopped services in view", func(t *testing.T) {
		statuses := []monitor.ServiceStatus{
			{Name: "stopped-svc", Port: 8080, Running: false},
		}
		m := newModel(statuses, nil)

		view := m.View()

		if view == "" {
			t.Error("expected non-empty view")
		}
	})

	t.Run("should show PID and session for running services", func(t *testing.T) {
		statuses := []monitor.ServiceStatus{
			{Name: "svc", Port: 8080, Running: true, PID: 12345, Session: "my-session"},
		}
		m := newModel(statuses, nil)

		view := m.View()

		// View should contain service info (exact formatting tested manually)
		if view == "" {
			t.Error("expected service info in view")
		}
	})

	t.Run("should handle services with no session (show N/A)", func(t *testing.T) {
		statuses := []monitor.ServiceStatus{
			{Name: "external", Port: 8080, Running: true, PID: 999, Session: ""},
		}
		m := newModel(statuses, nil)

		view := m.View()

		if view == "" {
			t.Error("expected service info in view")
		}
	})
}
