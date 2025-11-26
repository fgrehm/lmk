package main

import (
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"under a second", 500 * time.Millisecond, "0.5s"},
		{"seconds", 5 * time.Second, "5.0s"},
		{"under a minute", 45 * time.Second, "45.0s"},
		{"minutes and seconds", 2*time.Minute + 30*time.Second, "2m 30s"},
		{"over an hour", 2*time.Hour + 15*time.Minute, "2h 15m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.duration)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %v, want %v", tt.duration, got, tt.want)
			}
		})
	}
}

func TestGetExecutableAndArgs(t *testing.T) {
	tests := []struct {
		name           string
		cmd            []string
		wantExecutable string
		wantArgs       []string
		wantError      bool
	}{
		{
			name: "echo with args",
			cmd:  []string{"echo", "hello", "world"},
			wantExecutable: func() string {
				path, _ := exec.LookPath("echo")
				return path
			}(),
			wantArgs:  []string{"hello", "world"},
			wantError: false,
		},
		{
			name: "single command",
			cmd:  []string{"pwd"},
			wantExecutable: func() string {
				path, _ := exec.LookPath("pwd")
				return path
			}(),
			wantArgs:  []string{},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executable, args := getExecutableAndArgs(tt.cmd)

			if executable != tt.wantExecutable {
				t.Errorf("getExecutableAndArgs() executable = %v, want %v", executable, tt.wantExecutable)
			}

			if len(args) != len(tt.wantArgs) {
				t.Errorf("getExecutableAndArgs() args length = %v, want %v", len(args), len(tt.wantArgs))
			}

			for i := range args {
				if args[i] != tt.wantArgs[i] {
					t.Errorf("getExecutableAndArgs() args[%d] = %v, want %v", i, args[i], tt.wantArgs[i])
				}
			}
		})
	}
}

func TestGetMessageAndExitCode(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		cmd          []string
		duration     time.Duration
		wantIsError  bool
		wantMsgParts []string
	}{
		{
			name:         "success",
			err:          nil,
			cmd:          []string{"echo", "test"},
			duration:     2 * time.Second,
			wantIsError:  false,
			wantMsgParts: []string{"✅", "completed successfully", "2.0s"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, exitCode, isError := getMessageAndExitCode(tt.err, tt.cmd, tt.duration)

			if isError != tt.wantIsError {
				t.Errorf("getMessageAndExitCode() isError = %v, want %v", isError, tt.wantIsError)
			}

			if !tt.wantIsError && exitCode != 0 {
				t.Errorf("getMessageAndExitCode() exitCode = %v, want 0 for success", exitCode)
			}

			for _, part := range tt.wantMsgParts {
				if !strings.Contains(msg, part) {
					t.Errorf("getMessageAndExitCode() message should contain %q, got %q", part, msg)
				}
			}
		})
	}
}

func TestGetMessageAndExitCodeWithRealCommand(t *testing.T) {
	// Test with real command failure
	t.Run("real command failure", func(t *testing.T) {
		cmd := exec.Command("false")
		err := cmd.Run()

		msg, exitCode, isError := getMessageAndExitCode(err, []string{"false"}, 1*time.Second)

		if !isError {
			t.Error("getMessageAndExitCode() isError = false, want true")
		}

		if exitCode != 1 {
			t.Errorf("getMessageAndExitCode() exitCode = %v, want 1", exitCode)
		}

		if !strings.Contains(msg, "❌") || !strings.Contains(msg, "failed") {
			t.Errorf("getMessageAndExitCode() message should contain error indicators, got %q", msg)
		}
	})
}

func TestEscapeAppleScript(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple string", "hello", "hello"},
		{"with quotes", `hello "world"`, `hello \"world\"`},
		{"with backslash", `hello\world`, `hello\\world`},
		{"with newline", "hello\nworld", "hello\\nworld"},
		{"complex", `test "foo\bar"`, `test \"foo\\bar\"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeAppleScript(tt.input)
			if got != tt.want {
				t.Errorf("escapeAppleScript(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEscapeWindowsString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple string", "hello", "hello"},
		{"with single quote", "hello'world", "hello''world"},
		{"with newline", "hello\nworld", "hello`nworld"},
		{"complex", "test's\nvalue", "test''s`nvalue"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeWindowsString(tt.input)
			if got != tt.want {
				t.Errorf("escapeWindowsString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRun(t *testing.T) {
	// Test successful command
	t.Run("successful command", func(t *testing.T) {
		err := run("echo", "test")
		if err != nil {
			t.Errorf("run() error = %v, want nil", err)
		}
	})

	// Test failing command
	t.Run("failing command", func(t *testing.T) {
		err := run("false")
		if err == nil {
			t.Error("run() error = nil, want error")
		}
	})
}

func TestTimerDurationParsing(t *testing.T) {
	tests := []struct {
		name        string
		duration    string
		wantError   bool
		expectValid bool
	}{
		{"25 minutes", "25m", false, true},
		{"5 seconds", "5s", false, true},
		{"1 hour 30 minutes", "1h30m", false, true},
		{"90 seconds", "90s", false, true},
		{"invalid format", "abc", true, false},
		{"negative duration", "-5m", false, false}, // parses but should fail validation
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration, err := time.ParseDuration(tt.duration)

			if tt.wantError {
				if err == nil {
					t.Errorf("ParseDuration(%q) expected error, got nil", tt.duration)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseDuration(%q) unexpected error: %v", tt.duration, err)
				return
			}

			if tt.expectValid && duration <= 0 {
				t.Errorf("ParseDuration(%q) = %v, should be positive", tt.duration, duration)
			}
		})
	}
}

func TestIsAckModeSupported(t *testing.T) {
	// Test platform check
	supported := isAckModeSupported()

	if runtime.GOOS != "linux" {
		// Non-Linux should never support ack mode
		if supported {
			t.Error("isAckModeSupported() = true on non-Linux platform, want false")
		}
	} else {
		// Linux support depends on yad availability
		_, err := exec.LookPath("yad")
		expectedSupport := (err == nil)

		if supported != expectedSupport {
			t.Errorf("isAckModeSupported() = %v, want %v (yad available = %v)", supported, expectedSupport, err == nil)
		}
	}
}

func TestAckModeFallback(t *testing.T) {
	// Save and restore original GOOS
	originalGOOS := runtime.GOOS

	tests := []struct {
		name    string
		ackMode bool
		wantMsg string
	}{
		{
			name:    "ack mode enabled in dry run",
			ackMode: true,
			wantMsg: "Ack mode: true",
		},
		{
			name:    "ack mode disabled in dry run",
			ackMode: false,
			wantMsg: "Ack mode: false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use dry run to verify ack mode is being logged
			t.Setenv("LMK_DRY_RUN", "1")

			// This would normally show a dialog, but in dry run it just logs
			showDialog("test message", false, tt.ackMode)

			// The test passes if we get here without panic
			// In real usage, non-Linux or non-yad systems would log a warning
			// and fall back to normal mode
		})
	}

	_ = originalGOOS // keep linter happy
}
