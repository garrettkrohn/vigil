package tmux

import (
	"os"
	"testing"
)

func TestParseSessionsList(t *testing.T) {
	tests := []struct {
		name      string
		output    string
		wantCount int
	}{
		{
			name:      "should list all tmux sessions",
			output:    "api\nauth\ngateway\n",
			wantCount: 3,
		},
		{
			name:      "should handle empty output",
			output:    "",
			wantCount: 0,
		},
		{
			name:      "should handle single session",
			output:    "main\n",
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessions := parseSessionsList(tt.output)
			if len(sessions) != tt.wantCount {
				t.Errorf("expected %d sessions, got %d", tt.wantCount, len(sessions))
			}
		})
	}
}

func TestParsePanesOutput(t *testing.T) {
	tests := []struct {
		name       string
		output     string
		wantPIDMap map[int]string
	}{
		{
			name:   "should get all PIDs from all sessions/panes",
			output: "api 12345\nauth 12346\ngateway 12347\n",
			wantPIDMap: map[int]string{
				12345: "api",
				12346: "auth",
				12347: "gateway",
			},
		},
		{
			name:   "should parse tmux list-panes -a -F output correctly",
			output: "session1 1000\nsession1 1001\nsession2 2000\n",
			wantPIDMap: map[int]string{
				1000: "session1",
				1001: "session1",
				2000: "session2",
			},
		},
		{
			name:   "should handle session names with spaces",
			output: "pub - master 71038\ncore - master 83437\ncal-feature-branch 24108\n",
			wantPIDMap: map[int]string{
				71038: "pub - master",
				83437: "core - master",
				24108: "cal-feature-branch",
			},
		},
		{
			name:       "should handle empty output",
			output:     "",
			wantPIDMap: map[int]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pidMap := parsePanesOutput(tt.output)

			if len(pidMap) != len(tt.wantPIDMap) {
				t.Errorf("expected %d entries, got %d", len(tt.wantPIDMap), len(pidMap))
			}

			for pid, session := range tt.wantPIDMap {
				if pidMap[pid] != session {
					t.Errorf("expected PID %d to map to session %s, got %s", pid, session, pidMap[pid])
				}
			}
		})
	}
}

func TestGetActiveSessions(t *testing.T) {
	t.Run("should handle tmux not running", func(t *testing.T) {
		// This might actually succeed if tmux is running, but we're testing error handling
		sessions := GetActiveSessions()
		// Should not panic
		_ = sessions
	})
}

func TestGetSessionForPID(t *testing.T) {
	t.Run("should handle PID not in any tmux session", func(t *testing.T) {
		session := GetSessionForPID(999999)
		if session != "" {
			t.Errorf("expected empty session for non-existent PID, got %s", session)
		}
	})
}

func TestAttachToSession(t *testing.T) {
	t.Run("should handle invalid session name", func(t *testing.T) {
		err := AttachToSession("nonexistent-session-12345")
		// Should return error, not panic
		if err == nil {
			t.Error("expected error when attaching to non-existent session")
		}
	})

	t.Run("should handle tmux not running", func(t *testing.T) {
		// Just verify it doesn't panic
		_ = AttachToSession("any-session")
	})
}

func TestGetAllPIDsWithSessions(t *testing.T) {
	t.Run("should build PID to session map", func(t *testing.T) {
		pidMap := GetAllPIDsWithSessions()
		// Should not panic, map could be empty or populated
		_ = pidMap
	})
}

func TestGetParentPID(t *testing.T) {
	t.Run("should get parent PID of current process", func(t *testing.T) {
		// Get our own PID
		myPID := os.Getpid()

		// Get parent PID
		parentPID := getParentPID(myPID)

		// Parent should exist and be greater than 0
		if parentPID <= 0 {
			t.Errorf("expected positive parent PID, got %d", parentPID)
		}
	})

	t.Run("should return 0 for non-existent PID", func(t *testing.T) {
		parentPID := getParentPID(999999)
		if parentPID != 0 {
			t.Errorf("expected 0 for non-existent PID, got %d", parentPID)
		}
	})
}
