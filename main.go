package main

import (
	"fmt"
	"os"

	"vigil/config"
	"vigil/tui"
)

func showUsage() {
	fmt.Println("Usage:")
	fmt.Println("  vigil        - Launch interactive TUI to monitor services")
	fmt.Println("  vigil help   - Show this help message")
	fmt.Println("")
	fmt.Println("Configuration:")
	fmt.Println("  Vigil looks for config in this order:")
	fmt.Println("    1. ./config.yaml (current directory)")
	fmt.Println("    2. ~/.config/vigil/config.yaml (global)")
	fmt.Println("")
	fmt.Println("TUI Keybindings:")
	fmt.Println("  j/k      - Navigate up/down")
	fmt.Println("  o        - Attach to tmux session")
	fmt.Println("  K        - Kill process (with confirmation)")
	fmt.Println("  r        - Refresh service list")
	fmt.Println("  q        - Quit")
}

func runInteractiveMode() {
	// Ensure config exists
	if err := config.EnsureConfigExists(); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating config: %v\n", err)
		fmt.Fprintf(os.Stderr, "You may need to create ~/.config/vigil/config.yaml manually\n")
		os.Exit(1)
	}

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Check if config has services
	if len(cfg.Services) == 0 {
		fmt.Fprintf(os.Stderr, "No services configured in ~/.config/vigil/config.yaml\n")
		fmt.Fprintf(os.Stderr, "Add services to the config file and try again.\n")
		os.Exit(1)
	}

	// Run TUI
	if err := tui.Run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	// Parse command line arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "help", "-h", "--help":
			showUsage()
			return
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
			showUsage()
			os.Exit(1)
		}
	}

	// No arguments - run interactive mode
	runInteractiveMode()
}
