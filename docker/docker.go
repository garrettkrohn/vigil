package docker

import (
	"os/exec"
	"strconv"
	"strings"
)

// GetActiveContainers returns a list of active docker container names
func GetActiveContainers() []string {
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}")
	output, err := cmd.Output()

	if err != nil {
		return []string{}
	}

	return parseContainersList(string(output))
}

// GetAllPIDsWithContainers returns a map of PID to container name for all docker containers
func GetAllPIDsWithContainers() map[int]string {
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}} {{.ID}}")
	output, err := cmd.Output()

	if err != nil {
		return make(map[int]string)
	}

	return parseContainersOutput(string(output))
}

// GetContainerForPort returns the container name for a given port mapping
// This checks if any container has the port mapped (e.g., 0.0.0.0:8080->8080/tcp)
func GetContainerForPort(port int) string {
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}} {{.Ports}}")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	portStr := strconv.Itoa(port)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		// Format: "container_name 0.0.0.0:8080->8080/tcp, ..."
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}

		containerName := parts[0]
		ports := parts[1]

		// Check if this container has the port mapped
		// Look for patterns like ":8080->" or "0.0.0.0:8080->"
		if strings.Contains(ports, ":"+portStr+"->") {
			return containerName
		}
	}

	return ""
}

// GetContainerForPID returns the container name for a given PID, or empty string if not found
// This function walks up the parent process tree to find a docker container
func GetContainerForPID(pid int) string {
	pidMap := GetAllPIDsWithContainers()

	// Check if the PID itself is in a container
	if container, ok := pidMap[pid]; ok {
		return container
	}

	// Walk up the parent process tree
	currentPID := pid
	for i := 0; i < 10; i++ { // Limit to 10 levels to avoid infinite loops
		parentPID := getParentPID(currentPID)
		if parentPID <= 1 {
			break
		}

		if container, ok := pidMap[parentPID]; ok {
			return container
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

// parseContainersList parses the output of docker ps
func parseContainersList(output string) []string {
	output = strings.TrimSpace(output)
	if output == "" {
		return []string{}
	}

	lines := strings.Split(output, "\n")
	containers := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			containers = append(containers, line)
		}
	}

	return containers
}

// parseContainersOutput parses the output of docker ps to get container names and PIDs
// Expected format: "container_name container_id"
func parseContainersOutput(output string) map[int]string {
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

		containerName := fields[0]
		containerID := fields[1]

		// Get the main PID of the container
		pid := getContainerPID(containerID)
		if pid > 0 {
			pidMap[pid] = containerName
		}
	}

	return pidMap
}

// getContainerPID gets the main PID of a docker container
func getContainerPID(containerID string) int {
	cmd := exec.Command("docker", "inspect", "--format", "{{.State.Pid}}", containerID)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	pidStr := strings.TrimSpace(string(output))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0
	}

	return pid
}
