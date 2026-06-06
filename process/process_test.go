package process

import (
	"testing"
)

func TestParseLsofOutput(t *testing.T) {
	tests := []struct {
		name       string
		output     string
		wantPID    int
		wantCmd    string
		wantFound  bool
	}{
		{
			name:      "should parse lsof output to extract PID for port",
			output:    "node      12345 user   20u  IPv4 0x1234      0t0  TCP *:8080 (LISTEN)\n",
			wantPID:   12345,
			wantCmd:   "node",
			wantFound: true,
		},
		{
			name:      "should return empty result if port not in use",
			output:    "",
			wantPID:   0,
			wantCmd:   "",
			wantFound: false,
		},
		{
			name: "should handle multiple processes (return first)",
			output: `node      12345 user   20u  IPv4 0x1234      0t0  TCP *:8080 (LISTEN)
node      12346 user   21u  IPv4 0x1235      0t0  TCP *:8080 (LISTEN)
`,
			wantPID:   12345,
			wantCmd:   "node",
			wantFound: true,
		},
		{
			name:      "should extract command name from lsof output",
			output:    "python3   98765 user   10u  IPv6 0xabc       0t0  TCP *:3000 (LISTEN)\n",
			wantPID:   98765,
			wantCmd:   "python3",
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := parseLsofOutput(tt.output)

			if info.Found != tt.wantFound {
				t.Errorf("expected found=%v, got %v", tt.wantFound, info.Found)
			}

			if info.Found {
				if info.PID != tt.wantPID {
					t.Errorf("expected PID %d, got %d", tt.wantPID, info.PID)
				}
				if info.Command != tt.wantCmd {
					t.Errorf("expected command %s, got %s", tt.wantCmd, info.Command)
				}
			}
		})
	}
}

func TestGetProcessOnPort(t *testing.T) {
	t.Run("should handle lsof command errors gracefully", func(t *testing.T) {
		// Test with a port that's unlikely to be in use
		info := GetProcessOnPort(99999)

		// Should not panic, should return empty result
		if info.Found {
			t.Error("expected no process found on port 99999")
		}
	})
}

func TestGetProcessesOnPorts(t *testing.T) {
	t.Run("should batch check multiple ports efficiently", func(t *testing.T) {
		ports := []int{99998, 99999}
		infos := GetProcessesOnPorts(ports)

		if len(infos) != 2 {
			t.Errorf("expected 2 results, got %d", len(infos))
		}

		for i, info := range infos {
			if info.Port != ports[i] {
				t.Errorf("expected port %d, got %d", ports[i], info.Port)
			}
		}
	})
}

func TestKillProcess(t *testing.T) {
	t.Run("should handle kill errors (process not found)", func(t *testing.T) {
		// Try to kill a PID that doesn't exist
		err := KillProcess(999999)

		// Should return error, not panic
		if err == nil {
			t.Error("expected error when killing non-existent process")
		}
	})

	t.Run("should handle permission denied gracefully", func(t *testing.T) {
		// Try to kill PID 1 (init) which should fail with permission denied
		err := KillProcess(1)

		// Should return error
		if err == nil {
			t.Error("expected error when killing init process")
		}
	})
}

func TestProcessInfo(t *testing.T) {
	t.Run("ProcessInfo should have correct structure", func(t *testing.T) {
		info := ProcessInfo{
			PID:     12345,
			Port:    8080,
			Command: "node",
			Found:   true,
		}

		if info.PID != 12345 {
			t.Error("PID not set correctly")
		}
		if info.Port != 8080 {
			t.Error("Port not set correctly")
		}
		if info.Command != "node" {
			t.Error("Command not set correctly")
		}
		if !info.Found {
			t.Error("Found not set correctly")
		}
	})
}
