package monitor

import (
	"vigil/config"
	"vigil/docker"
	"vigil/process"
	"vigil/tmux"
)

// ServiceStatus represents the current status of a service on a specific port
type ServiceStatus struct {
	Name      string
	Port      int
	Running   bool
	PID       int
	Session   string
	UseDocker bool
}

// GetAllServiceStatuses builds a complete picture of all services from config
func GetAllServiceStatuses(cfg *config.Config) []ServiceStatus {
	var statuses []ServiceStatus

	// For each service, check each port
	for _, service := range cfg.Services {
		for _, port := range service.Ports {
			status := getServiceStatus(service.Name, port, service.Docker)
			statuses = append(statuses, status)
		}
	}

	return statuses
}

// RefreshStatuses updates the service statuses by re-checking all ports
func RefreshStatuses(cfg *config.Config) []ServiceStatus {
	return GetAllServiceStatuses(cfg)
}

// getServiceStatus checks the status of a single service on a single port
func getServiceStatus(serviceName string, port int, useDocker bool) ServiceStatus {
	// Check if port is active
	processInfo := process.GetProcessOnPort(port)

	status := ServiceStatus{
		Name:      serviceName,
		Port:      port,
		Running:   processInfo.Found,
		PID:       processInfo.PID,
		Session:   "",
		UseDocker: useDocker,
	}

	// If process found, check if it's in a docker container or tmux session
	if processInfo.Found {
		if useDocker {
			// First try to find container by port mapping (for Docker port forwarding)
			status.Session = docker.GetContainerForPort(port)
			// If not found by port, try walking up the process tree
			if status.Session == "" {
				status.Session = docker.GetContainerForPID(processInfo.PID)
			}
		} else {
			// Walk up the parent process tree to find the tmux session
			status.Session = tmux.GetSessionForPID(processInfo.PID)
		}
	}

	return status
}
