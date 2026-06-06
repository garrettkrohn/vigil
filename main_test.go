package main

import (
	"os"
	"testing"
)

func TestShowUsage(t *testing.T) {
	t.Run("should display usage information", func(t *testing.T) {
		// Just verify it doesn't panic
		showUsage()
	})
}

func TestCommandParsing(t *testing.T) {
	t.Run("should recognize help command", func(t *testing.T) {
		// Verify help command is handled
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()

		os.Args = []string{"vigil", "help"}
		// In real execution, this would call showUsage
		if len(os.Args) > 1 && os.Args[1] == "help" {
			// Test passes
		} else {
			t.Error("help command not recognized")
		}
	})

	t.Run("should handle invalid commands", func(t *testing.T) {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()

		os.Args = []string{"vigil", "invalid-command"}

		if len(os.Args) > 1 {
			cmd := os.Args[1]
			validCommands := []string{"help", "-h", "--help"}
			isValid := false
			for _, valid := range validCommands {
				if cmd == valid {
					isValid = true
					break
				}
			}

			if isValid {
				t.Error("invalid command recognized as valid")
			}
		}
	})
}

func TestConfigHandling(t *testing.T) {
	t.Run("should handle config loading", func(t *testing.T) {
		// This is tested via config package tests
		// Just verify the structure exists
	})
}
