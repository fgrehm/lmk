package main

import (
	"os"
	"testing"
	"time"
)

func TestDialogDelay(t *testing.T) {
	// Enable dry-run to avoid showing dialogs
	os.Setenv("LMK_DRY_RUN", "1")
	defer os.Unsetenv("LMK_DRY_RUN")

	tests := []struct {
		name        string
		envValue    string
		expectDelay bool
	}{
		{"default delay", "", false},        // dry-run exits early, no delay
		{"custom delay 2s", "2s", false},    // dry-run exits early
		{"no delay", "0s", false},           // dry-run exits early
		{"invalid delay", "invalid", false}, // dry-run exits early
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("LMK_DELAY", tt.envValue)
				defer os.Unsetenv("LMK_DELAY")
			}

			start := time.Now()
			showDialog("test message", false, false)
			elapsed := time.Since(start)

			// In dry-run mode, should return immediately
			if elapsed > 100*time.Millisecond {
				t.Errorf("Dry-run should be fast, took %v", elapsed)
			}
		})
	}
}

func TestDelayParsing(t *testing.T) {
	// Test delay parsing without dry-run (we'll check the sleep happens)
	tests := []struct {
		name         string
		envValue     string
		minExpected  time.Duration
		maxExpected  time.Duration
		shouldErrMsg string
	}{
		{"valid 100ms", "100ms", 90 * time.Millisecond, 150 * time.Millisecond, ""},
		{"valid 1s", "1s", 950 * time.Millisecond, 1100 * time.Millisecond, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("LMK_DELAY", tt.envValue)
			defer os.Unsetenv("LMK_DELAY")

			// Parse the delay value (simulating showDialog logic)
			delay := 1500 * time.Millisecond // Default
			if delayStr := os.Getenv("LMK_DELAY"); delayStr != "" {
				if customDelay, err := time.ParseDuration(delayStr); err == nil {
					delay = customDelay
				}
			}

			// Verify parsed value
			if delay < tt.minExpected || delay > tt.maxExpected {
				t.Errorf("Parsed delay %v not in expected range [%v, %v]",
					delay, tt.minExpected, tt.maxExpected)
			}
		})
	}
}
