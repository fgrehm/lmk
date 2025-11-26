package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"
)

func TestShowDialogDryRun(t *testing.T) {
	// Enable dry-run mode
	os.Setenv("LMK_DRY_RUN", "1")
	defer os.Unsetenv("LMK_DRY_RUN")

	tests := []struct {
		name    string
		msg     string
		isError bool
	}{
		{"success message", "Command completed successfully", false},
		{"error message", "Command failed with exit code 1", true},
		{"timer message", "⏰ Pomodoro done!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			showDialog(tt.msg, tt.isError, false)

			w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			// Verify dry-run output
			if !strings.Contains(output, "[DRY RUN]") {
				t.Errorf("Expected [DRY RUN] in output, got: %s", output)
			}

			if !strings.Contains(output, tt.msg) {
				t.Errorf("Expected message %q in output, got: %s", tt.msg, output)
			}
		})
	}
}

func TestShowNotificationAndWaitDryRun(t *testing.T) {
	// Enable dry-run mode
	os.Setenv("LMK_DRY_RUN", "1")
	defer os.Unsetenv("LMK_DRY_RUN")

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	showNotificationAndWait("Test notification", false)

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify dry-run output
	if !strings.Contains(output, "[DRY RUN]") {
		t.Errorf("Expected [DRY RUN] in output, got: %s", output)
	}

	if !strings.Contains(output, "Would send notification") {
		t.Errorf("Expected 'Would send notification' in output, got: %s", output)
	}

	if !strings.Contains(output, "Would wait for Enter") {
		t.Errorf("Expected 'Would wait for Enter' in output, got: %s", output)
	}
}

func TestTimerDryRun(t *testing.T) {
	// Enable dry-run mode
	os.Setenv("LMK_DRY_RUN", "1")
	defer os.Unsetenv("LMK_DRY_RUN")

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Run a short timer
	start := time.Now()
	runTimer("1s", "Test timer", false)
	elapsed := time.Since(start)

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify timer actually ran (not skipped in dry-run)
	if elapsed < 900*time.Millisecond {
		t.Errorf("Timer should still sleep in dry-run mode, but elapsed only %v", elapsed)
	}

	// Verify dry-run dialog output
	if !strings.Contains(output, "[DRY RUN]") {
		t.Errorf("Expected [DRY RUN] in output, got: %s", output)
	}

	if !strings.Contains(output, "Test timer") {
		t.Errorf("Expected timer message in output, got: %s", output)
	}
}

func TestCommandExecutionDryRunDoesNotAffectCommand(t *testing.T) {
	// Enable dry-run mode
	os.Setenv("LMK_DRY_RUN", "1")
	defer os.Unsetenv("LMK_DRY_RUN")

	// Commands should still execute normally, only dialogs are skipped
	err := run("echo", "test")
	if err != nil {
		t.Errorf("Commands should still execute in dry-run mode, got error: %v", err)
	}
}
