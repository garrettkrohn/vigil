package process

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// ProcessInfo contains information about a process listening on a port
type ProcessInfo struct {
	PID     int
	Port    int
	Command string
	Found   bool
}

// GetProcessOnPort checks if a process is listening on the given port
func GetProcessOnPort(port int) ProcessInfo {
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port), "-sTCP:LISTEN", "-t", "-n", "-P")
	output, err := cmd.Output()

	if err != nil {
		return ProcessInfo{Port: port, Found: false}
	}

	return parseLsofOutput(string(output))
}

// GetProcessesOnPorts checks multiple ports and returns their process info
func GetProcessesOnPorts(ports []int) []ProcessInfo {
	infos := make([]ProcessInfo, len(ports))

	for i, port := range ports {
		info := GetProcessOnPort(port)
		info.Port = port
		infos[i] = info
	}

	return infos
}

// KillProcessGraceful terminates a process by PID using SIGTERM (graceful shutdown)
func KillProcessGraceful(pid int) error {
	// Check if process exists first
	if err := syscall.Kill(pid, 0); err != nil {
		return fmt.Errorf("process not found: %w", err)
	}

	// Send SIGTERM for graceful shutdown
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM: %w", err)
	}

	// Wait briefly to give process time to shut down
	time.Sleep(200 * time.Millisecond)

	return nil
}

// KillProcessForce terminates a process by PID using SIGKILL (forceful termination)
func KillProcessForce(pid int) error {
	// Check if process exists first
	if err := syscall.Kill(pid, 0); err != nil {
		return fmt.Errorf("process not found: %w", err)
	}

	// Send SIGKILL for immediate termination
	if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
		return fmt.Errorf("failed to send SIGKILL: %w", err)
	}

	// Wait briefly for SIGKILL to take effect
	time.Sleep(200 * time.Millisecond)

	// Verify process is actually dead
	if err := syscall.Kill(pid, 0); err == nil {
		return fmt.Errorf("process still running after SIGKILL")
	}

	return nil
}

// KillProcess is a convenience function that tries SIGTERM first, then escalates to SIGKILL
// Deprecated: Use KillProcessGraceful or KillProcessForce directly for explicit control
func KillProcess(pid int) error {
	// Check if process exists first
	if err := syscall.Kill(pid, 0); err != nil {
		return fmt.Errorf("process not found: %w", err)
	}

	// Try graceful termination with SIGTERM
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM: %w", err)
	}

	// Wait up to 2 seconds for process to exit
	maxWait := 2 * time.Second
	checkInterval := 100 * time.Millisecond
	elapsed := time.Duration(0)

	for elapsed < maxWait {
		time.Sleep(checkInterval)
		elapsed += checkInterval

		// Check if process still exists (signal 0 checks existence)
		if err := syscall.Kill(pid, 0); err != nil {
			// Process is gone
			return nil
		}
	}

	// Process didn't exit gracefully, use SIGKILL
	if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
		return fmt.Errorf("failed to send SIGKILL: %w", err)
	}

	// Wait briefly for SIGKILL to take effect
	time.Sleep(200 * time.Millisecond)

	// Verify process is actually dead
	if err := syscall.Kill(pid, 0); err == nil {
		return fmt.Errorf("process still running after SIGKILL")
	}

	return nil
}

// parseLsofOutput parses the output from lsof to extract PID and command
func parseLsofOutput(output string) ProcessInfo {
	output = strings.TrimSpace(output)
	if output == "" {
		return ProcessInfo{Found: false}
	}

	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		return ProcessInfo{Found: false}
	}

	// First line contains the process info
	firstLine := lines[0]
	fields := strings.Fields(firstLine)

	if len(fields) == 0 {
		return ProcessInfo{Found: false}
	}

	// Check if this is lsof output with -t flag (just PID)
	if len(fields) == 1 {
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			return ProcessInfo{Found: false}
		}

		command := getCommandForPID(pid)
		return ProcessInfo{
			PID:     pid,
			Command: command,
			Found:   true,
		}
	}

	// Otherwise, parse full lsof output: COMMAND PID USER ...
	command := fields[0]
	pid, err := strconv.Atoi(fields[1])
	if err != nil {
		return ProcessInfo{Found: false}
	}

	return ProcessInfo{
		PID:     pid,
		Command: command,
		Found:   true,
	}
}

// getCommandForPID gets the command name for a given PID
func getCommandForPID(pid int) string {
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "comm=")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	command := strings.TrimSpace(string(output))
	// Remove path if present
	if idx := strings.LastIndex(command, "/"); idx != -1 {
		command = command[idx+1:]
	}

	return command
}
