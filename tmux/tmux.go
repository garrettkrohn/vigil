package tmux

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// GetActiveSessions returns a list of active tmux session names
func GetActiveSessions() []string {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	output, err := cmd.Output()

	if err != nil {
		return []string{}
	}

	return parseSessionsList(string(output))
}

// GetAllPIDsWithSessions returns a map of PID to session name for all tmux panes
func GetAllPIDsWithSessions() map[int]string {
	cmd := exec.Command("tmux", "list-panes", "-a", "-F", "#{session_name} #{pane_pid}")
	output, err := cmd.Output()

	if err != nil {
		return make(map[int]string)
	}

	return parsePanesOutput(string(output))
}

// GetSessionForPID returns the session name for a given PID, or empty string if not found
// This function walks up the parent process tree to find a tmux session
func GetSessionForPID(pid int) string {
	pidMap := GetAllPIDsWithSessions()

	// Check if the PID itself is in a tmux pane
	if session, ok := pidMap[pid]; ok {
		return session
	}

	// Walk up the parent process tree
	currentPID := pid
	for i := 0; i < 10; i++ { // Limit to 10 levels to avoid infinite loops
		parentPID := getParentPID(currentPID)
		if parentPID <= 1 {
			break
		}

		if session, ok := pidMap[parentPID]; ok {
			return session
		}

		currentPID = parentPID
	}

	return ""
}

// getParentPID returns the parent PID of a given PID
func getParentPID(pid int) int {
	cmd := exec.Command("ps", "-o", "ppid=", "-p", strconv.Itoa(pid))
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	ppidStr := strings.TrimSpace(string(output))
	ppid, err := strconv.Atoi(ppidStr)
	if err != nil {
		return 0
	}

	return ppid
}

// AttachToSession switches to a specific tmux session
func AttachToSession(sessionName string) error {
	cmd := exec.Command("tmux", "switch-client", "-t", sessionName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to attach to session %s: %w", sessionName, err)
	}
	return nil
}

// parseSessionsList parses the output of tmux list-sessions
func parseSessionsList(output string) []string {
	output = strings.TrimSpace(output)
	if output == "" {
		return []string{}
	}

	lines := strings.Split(output, "\n")
	sessions := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			sessions = append(sessions, line)
		}
	}

	return sessions
}

// parsePanesOutput parses the output of tmux list-panes -a
// Expected format: "session_name pane_pid"
// Note: session_name can contain spaces, so PID is always the last field
func parsePanesOutput(output string) map[int]string {
	pidMap := make(map[int]string)
	output = strings.TrimSpace(output)

	if output == "" {
		return pidMap
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split by whitespace
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		// PID is always the last field
		pidStr := fields[len(fields)-1]
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		// Session name is everything except the last field
		sessionName := strings.Join(fields[:len(fields)-1], " ")

		pidMap[pid] = sessionName
	}

	return pidMap
}
