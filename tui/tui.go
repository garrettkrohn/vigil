package tui

import (
	"fmt"
	"vigil/config"
	"vigil/monitor"
	"vigil/process"
	"vigil/tmux"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Catppuccin Mocha color palette
var (
	catppuccinGreen    = lipgloss.Color("#a6e3a1")
	catppuccinRed      = lipgloss.Color("#f38ba8")
	catppuccinMauve    = lipgloss.Color("#cba6f7")
	catppuccinPink     = lipgloss.Color("#f5c2e7")
	catppuccinText     = lipgloss.Color("#cdd6f4")
	catppuccinSubtext0 = lipgloss.Color("#a6adc8")
	catppuccinSurface0 = lipgloss.Color("#313244")
	catppuccinBlue     = lipgloss.Color("#89b4fa")
	catppuccinYellow   = lipgloss.Color("#f9e2af")
)

type model struct {
	statuses       []monitor.ServiceStatus
	cursor         int
	cfg            *config.Config
	confirmingKill bool
	killTarget     int
	killForce      bool // true for SIGKILL, false for SIGTERM
	err            string
}

func newModel(statuses []monitor.ServiceStatus, cfg *config.Config) model {
	return model{
		statuses:       statuses,
		cursor:         0,
		cfg:            cfg,
		confirmingKill: false,
		killTarget:     -1,
		killForce:      false,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle confirmation mode
		if m.confirmingKill {
			switch msg.String() {
			case "y", "Y":
				// Kill the process
				if m.killTarget >= 0 && m.killTarget < len(m.statuses) {
					status := m.statuses[m.killTarget]
					if status.Running && status.PID > 0 {
						var err error
						if m.killForce {
							err = process.KillProcessForce(status.PID)
						} else {
							err = process.KillProcessGraceful(status.PID)
						}

						if err != nil {
							m.err = fmt.Sprintf("Failed to kill PID %d: %v", status.PID, err)
						} else {
							// Refresh statuses after killing
							if m.cfg != nil {
								m.statuses = monitor.RefreshStatuses(m.cfg)
							}
							signalType := "SIGTERM"
							if m.killForce {
								signalType = "SIGKILL"
							}
							m.err = fmt.Sprintf("Sent %s to PID %d", signalType, status.PID)
						}
					}
				}
				m.confirmingKill = false
				m.killTarget = -1
				m.killForce = false
				return m, nil
			case "n", "N", "esc":
				// Cancel
				m.confirmingKill = false
				m.killTarget = -1
				m.killForce = false
				return m, nil
			}
			return m, nil
		}

		// Normal mode key handling
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "j", "down":
			if m.cursor < len(m.statuses)-1 {
				m.cursor++
			}

		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}

		case "o":
			// Attach to session
			if m.cursor < len(m.statuses) {
				status := m.statuses[m.cursor]
				if status.UseDocker {
					m.err = "Cannot attach to docker containers (use 'docker exec' instead)"
				} else if status.Session != "" {
					if err := tmux.AttachToSession(status.Session); err != nil {
						m.err = fmt.Sprintf("Failed to attach: %v", err)
					} else {
						return m, tea.Quit
					}
				} else {
					m.err = "No tmux session for this service"
				}
			}

		case "t":
			// Initiate graceful kill confirmation (SIGTERM)
			if m.cursor < len(m.statuses) {
				status := m.statuses[m.cursor]
				if status.Running && status.PID > 0 {
					m.confirmingKill = true
					m.killTarget = m.cursor
					m.killForce = false
				} else {
					m.err = "No process to kill"
				}
			}

		case "K":
			// Initiate force kill confirmation (SIGKILL)
			if m.cursor < len(m.statuses) {
				status := m.statuses[m.cursor]
				if status.Running && status.PID > 0 {
					m.confirmingKill = true
					m.killTarget = m.cursor
					m.killForce = true
				} else {
					m.err = "No process to kill"
				}
			}

		case "r":
			// Refresh statuses
			if m.cfg != nil {
				m.statuses = monitor.RefreshStatuses(m.cfg)
				m.err = "Refreshed"
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(catppuccinMauve).
		Background(catppuccinSurface0).
		Padding(0, 2).
		MarginBottom(1)

	helpStyle := lipgloss.NewStyle().
		Foreground(catppuccinSubtext0).
		Italic(true)

	keybindStyle := lipgloss.NewStyle().
		Foreground(catppuccinBlue).
		Bold(true)

	cursorStyle := lipgloss.NewStyle().
		Foreground(catppuccinPink).
		Bold(true)

	runningStyle := lipgloss.NewStyle().
		Foreground(catppuccinGreen)

	stoppedStyle := lipgloss.NewStyle().
		Foreground(catppuccinRed)

	highlightStyle := lipgloss.NewStyle().
		Background(catppuccinSurface0)

	// Build view
	var s string

	// Title
	s += titleStyle.Render("✨ Vigil - Service Monitor") + "\n\n"

	// Confirmation prompt
	if m.confirmingKill {
		if m.killTarget >= 0 && m.killTarget < len(m.statuses) {
			status := m.statuses[m.killTarget]
			signalType := "SIGTERM"
			if m.killForce {
				signalType = "SIGKILL"
			}
			s += helpStyle.Render(fmt.Sprintf("Send %s to %s (PID %d)? (y/n)", signalType, status.Name, status.PID)) + "\n\n"
		}
	} else {
		// Help text
		help := keybindStyle.Render("j/k") + " navigate  " +
			keybindStyle.Render("o") + " attach  " +
			keybindStyle.Render("t") + " terminate (SIGTERM)  " +
			keybindStyle.Render("K") + " kill (SIGKILL)  " +
			keybindStyle.Render("r") + " refresh  " +
			keybindStyle.Render("q") + " quit"
		s += helpStyle.Render(help) + "\n\n"
	}

	// Error message
	if m.err != "" {
		errorStyle := lipgloss.NewStyle().Foreground(catppuccinYellow)
		s += errorStyle.Render("⚠ "+m.err) + "\n\n"
	}

	// Service list
	if !m.confirmingKill {
		for i, status := range m.statuses {
			cursor := "  "
			if m.cursor == i {
				cursor = cursorStyle.Render("› ")
			}

			// Status indicator
			statusText := "●"
			var statusStyle lipgloss.Style
			if status.Running {
				statusStyle = runningStyle
			} else {
				statusStyle = stoppedStyle
			}

			// Build line
			line := fmt.Sprintf("%s %s %-20s :%-5d", cursor, statusStyle.Render(statusText), status.Name, status.Port)

			if status.Running {
				session := status.Session
				if session == "" {
					session = "N/A"
				}
				label := "Session:"
				if status.UseDocker {
					label = "Docker:"
				}
				line += fmt.Sprintf(" PID:%-6d %s %s", status.PID, label, session)
			} else {
				line += " [stopped]"
			}

			// Highlight current line
			if m.cursor == i {
				line = highlightStyle.Render(line)
			}

			s += line + "\n"
		}
	}

	return s
}

// Run starts the TUI
func Run(cfg *config.Config) error {
	statuses := monitor.GetAllServiceStatuses(cfg)
	m := newModel(statuses, cfg)

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}
