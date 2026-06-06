# Vigil

**Vigil** is a terminal-based service monitor for local development with microservices. It helps you track which services are running on specific ports, identifies their tmux sessions, and allows you to manage processes directly from a beautiful TUI.

## Features

- 🎯 **Port Monitoring**: Track services by port number using `lsof`
- 🖥️ **Tmux Integration**: See which tmux session each service is running in
- 🐳 **Docker Support**: Track services running in Docker containers
- 🎨 **Beautiful TUI**: Clean interface with Catppuccin color scheme
- ⚡ **Process Control**: Kill processes with confirmation
- 🔄 **Live Refresh**: Update service status on demand
- 🚀 **Session Switching**: Jump directly to any service's tmux session

## Installation

### From Source

```bash
# Clone the repository
git clone <repository-url>
cd vigil

# Build the binary
go build

# Optionally, move to your PATH
mv vigil /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/yourusername/vigil@latest
```

## Configuration

Vigil looks for a configuration file in this order:
1. `./config.yaml` (current directory)
2. `~/.config/vigil/config.yaml` (global config)

### Example Configuration

```yaml
services:
  - name: "api-service"
    ports: [8080]
  
  - name: "auth-service"
    ports: [8081, 9091]  # Main port and debug port
  
  - name: "gateway"
    ports: [3000, 3001, 3002]
  
  - name: "redis-cache"
    ports: [6379]
    docker: true  # Look for Docker container instead of tmux session
```

### Configuration Format

- **`name`**: Descriptive name for the service
- **`ports`**: Array of port numbers to monitor
- **`docker`** (optional): Set to `true` to track Docker container names instead of tmux sessions

Each port is displayed as a separate line in the TUI, making it easy to track services with multiple ports (e.g., main service port + debug port).

## Usage

### Launch the TUI

```bash
vigil
```

### Show Help

```bash
vigil help
```

## TUI Keybindings

| Key | Action |
|-----|--------|
| `j` / `↓` | Navigate down |
| `k` / `↑` | Navigate up |
| `o` | Attach to the service's tmux session |
| `K` | Kill the process (with confirmation) |
| `r` | Refresh service status |
| `q` | Quit |

## How It Works

1. **Port Detection**: Vigil uses `lsof -i :PORT -sTCP:LISTEN` to detect processes listening on configured ports
2. **PID Lookup**: Extracts the process ID (PID) from lsof output
3. **Session/Container Mapping**: 
   - For tmux services: Uses `tmux list-panes -a -F` to map PIDs to tmux sessions
   - For docker services: Uses `docker ps` and `docker inspect` to map PIDs to container names
4. **Parent Process Walking**: If a PID isn't directly in a tmux pane or container, Vigil walks up the parent process tree to find the session/container (e.g., when a service is launched from a shell)
5. **Status Display**: Shows running/stopped status, PID, and associated tmux session or Docker container

## TUI Display

The TUI shows:

- **Status Indicator**: Green ● for running, Red ● for stopped
- **Service Name**: From your config file
- **Port Number**: Which port is being monitored
- **PID**: Process ID (when running)
- **Session**: Tmux session name or Docker container name (or "N/A" if not found)

Example:
```
✨ Vigil - Service Monitor

j/k navigate  o attach  K kill  r refresh  q quit

  ● api-service        :8080  PID:12345  Session: api
› ● auth-service       :8081  PID:12346  Session: auth
  ● auth-service       :9091  PID:12346  Session: auth
  ● gateway            :3000  [stopped]
```

## Workflow Example

1. **Create a local config** (optional):
   ```bash
   # In your project directory
   cp example.config.yaml config.yaml
   # Edit config.yaml to match your services
   ```

2. **Start your services in tmux**:
   ```bash
   tmux new -s api "cd ~/api && npm start"
   tmux new -s auth "cd ~/auth && go run main.go"
   ```

3. **Launch Vigil**:
   ```bash
   vigil
   ```

3. **Monitor and manage**:
   - See which services are running
   - Press `o` to jump to a service's tmux session
   - Press `K` to kill a stuck process
   - Press `r` to refresh the status

## Troubleshooting

### "No services configured"

Create `config.yaml` in your current directory or `~/.config/vigil/config.yaml` with your services. See `example.config.yaml` for reference.

**Tip**: Use a local `config.yaml` for project-specific services, or a global `~/.config/vigil/config.yaml` for all your services.

### "Failed to attach to session"

Make sure you're running Vigil from within a tmux session. Vigil uses `tmux switch-client` to switch between sessions.

### Process shown but no tmux session

The process is running outside of tmux. Vigil will still show the PID and let you kill it, but you can't attach to it.

### Ports not detected

Ensure processes are actually listening on the configured ports. Use `lsof -i :PORT` manually to verify.

## Requirements

- **Go 1.20+** (for building)
- **tmux** (for session management, optional)
- **docker** (for container tracking, optional)
- **lsof** (for port detection - pre-installed on macOS and most Linux)

## Development

### Run Tests

```bash
go test ./...
```

### Run Tests with Coverage

```bash
go test ./... -cover
```

### Build

```bash
go build
```

## License

MIT

## Contributing

Contributions welcome! Please open an issue or PR.

## Acknowledgments

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework
- Styled with [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- Colors from [Catppuccin](https://github.com/catppuccin/catppuccin) theme
# vigil
