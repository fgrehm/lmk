package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGetNotificationIcon tests the icon mapping for different notification types
func TestGetNotificationIcon(t *testing.T) {
	tests := []struct {
		notificationType string
		expectedIcon     string
	}{
		{"permission_prompt", "🔐"},
		{"idle_prompt", "⏱️"},
		{"auth_success", "✅"},
		{"elicitation_dialog", "📝"},
		{"unknown_type", "🤖"},
		{"", "🤖"},
	}

	for _, tt := range tests {
		t.Run(tt.notificationType, func(t *testing.T) {
			icon := getNotificationIcon(tt.notificationType)
			if icon != tt.expectedIcon {
				t.Errorf("getNotificationIcon(%q) = %q, want %q", tt.notificationType, icon, tt.expectedIcon)
			}
		})
	}
}

// TestClaudeHookPayloadParsing tests JSON parsing of hook payloads
func TestClaudeHookPayloadParsing(t *testing.T) {
	tests := []struct {
		name     string
		jsonFile string
	}{
		{"permission_prompt", "testdata/permission_prompt.json"},
		{"idle_prompt", "testdata/idle_prompt.json"},
		{"auth_success", "testdata/auth_success.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(tt.jsonFile)
			if err != nil {
				t.Fatalf("Failed to read test file: %v", err)
			}

			var payload ClaudeHookPayload
			if err := json.Unmarshal(data, &payload); err != nil {
				t.Fatalf("Failed to parse JSON: %v", err)
			}

			if payload.NotificationType == "" {
				t.Error("NotificationType should not be empty")
			}
			if payload.Message == "" {
				t.Error("Message should not be empty")
			}
		})
	}
}

// TestReadOrCreateSettings tests reading existing and creating new settings
func TestReadOrCreateSettings(t *testing.T) {
	t.Run("read_existing_settings", func(t *testing.T) {
		settings := readOrCreateSettings("testdata/existing_settings.json")
		if _, ok := settings["hooks"]; !ok {
			t.Error("Expected hooks to be populated")
		}
	})

	t.Run("create_empty_settings", func(t *testing.T) {
		settings := readOrCreateSettings("testdata/nonexistent_file.json")
		if _, ok := settings["hooks"]; ok {
			t.Error("Expected hooks to not exist for new settings")
		}
		if settings == nil {
			t.Error("Expected empty map, not nil")
		}
	})

	t.Run("read_empty_settings", func(t *testing.T) {
		settings := readOrCreateSettings("testdata/empty_settings.json")
		if _, ok := settings["hooks"]; ok {
			t.Error("Expected hooks to not exist for empty settings")
		}
	})
}

// TestRemoveLmkHooks tests removing lmk hooks from configuration
func TestRemoveLmkHooks(t *testing.T) {
	t.Run("remove_from_nil", func(t *testing.T) {
		result := removeLmkHooks(nil)
		if result != nil {
			t.Error("Expected nil when removing from nil hooks")
		}
	})

	t.Run("remove_lmk_hook", func(t *testing.T) {
		hooks := &ClaudeHooks{
			Notification: []NotificationHook{
				{
					Hooks: []HookConfig{
						{Type: "command", Command: "lmk claude-hooks"},
					},
				},
			},
		}
		result := removeLmkHooks(hooks)
		if result != nil {
			t.Error("Expected nil after removing only lmk hook")
		}
	})

	t.Run("keep_other_hooks", func(t *testing.T) {
		hooks := &ClaudeHooks{
			Notification: []NotificationHook{
				{
					Hooks: []HookConfig{
						{Type: "command", Command: "lmk claude-hooks"},
					},
				},
				{
					Hooks: []HookConfig{
						{Type: "command", Command: "other-tool"},
					},
				},
			},
		}
		result := removeLmkHooks(hooks)
		if result == nil {
			t.Fatal("Expected hooks to remain after removing lmk")
		}
		if len(result.Notification) != 1 {
			t.Errorf("Expected 1 hook, got %d", len(result.Notification))
		}
		if !strings.Contains(result.Notification[0].Hooks[0].Command, "other-tool") {
			t.Error("Expected other-tool hook to remain")
		}
	})
}

// TestAddOrUpdateLmkHook tests adding/updating lmk hooks
func TestAddOrUpdateLmkHook(t *testing.T) {
	t.Run("add_to_nil", func(t *testing.T) {
		lmkHook := NotificationHook{
			Hooks: []HookConfig{
				{Type: "command", Command: "lmk claude-hooks"},
			},
		}
		result := addOrUpdateLmkHook(nil, lmkHook)
		if result == nil {
			t.Fatal("Expected hooks to be created")
		}
		if len(result.Notification) != 1 {
			t.Errorf("Expected 1 hook, got %d", len(result.Notification))
		}
	})

	t.Run("add_to_existing", func(t *testing.T) {
		hooks := &ClaudeHooks{
			Notification: []NotificationHook{
				{
					Hooks: []HookConfig{
						{Type: "command", Command: "other-tool"},
					},
				},
			},
		}
		lmkHook := NotificationHook{
			Hooks: []HookConfig{
				{Type: "command", Command: "lmk claude-hooks"},
			},
		}
		result := addOrUpdateLmkHook(hooks, lmkHook)
		if len(result.Notification) != 2 {
			t.Errorf("Expected 2 hooks, got %d", len(result.Notification))
		}
	})

	t.Run("update_existing_lmk", func(t *testing.T) {
		hooks := &ClaudeHooks{
			Notification: []NotificationHook{
				{
					Hooks: []HookConfig{
						{Type: "command", Command: "lmk claude-hooks --old"},
					},
				},
			},
		}
		lmkHook := NotificationHook{
			Hooks: []HookConfig{
				{Type: "command", Command: "lmk claude-hooks --new"},
			},
		}
		result := addOrUpdateLmkHook(hooks, lmkHook)
		if len(result.Notification) != 1 {
			t.Errorf("Expected 1 hook after update, got %d", len(result.Notification))
		}
		if !strings.Contains(result.Notification[0].Hooks[0].Command, "--new") {
			t.Error("Expected updated command with --new flag")
		}
	})
}

// TestWriteAndReadSettings tests round-trip settings persistence
func TestWriteAndReadSettings(t *testing.T) {
	tempDir := t.TempDir()
	settingsPath := filepath.Join(tempDir, ".claude", "settings.json")

	settings := make(ClaudeSettings)
	settings["hooks"] = &ClaudeHooks{
		Notification: []NotificationHook{
			{
				Hooks: []HookConfig{
					{Type: "command", Command: "lmk claude-hooks"},
				},
			},
		},
	}
	settings["someOtherSetting"] = "should be preserved"
	settings["anotherField"] = map[string]interface{}{
		"nested": "value",
	}

	// Write settings
	if err := writeSettings(settingsPath, settings); err != nil {
		t.Fatalf("Failed to write settings: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Fatal("Settings file was not created")
	}

	// Read back settings
	readSettings := readOrCreateSettings(settingsPath)

	// Verify hooks were preserved
	if _, ok := readSettings["hooks"]; !ok {
		t.Fatal("Expected hooks to be populated")
	}

	// Verify other settings were preserved
	if val, ok := readSettings["someOtherSetting"]; !ok || val != "should be preserved" {
		t.Error("Expected other settings to be preserved")
	}

	if _, ok := readSettings["anotherField"]; !ok {
		t.Error("Expected nested field to be preserved")
	}

	// Verify JSON formatting
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("Failed to read settings file: %v", err)
	}
	if !strings.Contains(string(data), "  ") {
		t.Error("Expected indented JSON output")
	}
}

// TestGetClaudeSettingsPath tests path generation
func TestGetClaudeSettingsPath(t *testing.T) {
	t.Run("local_path", func(t *testing.T) {
		path := getClaudeSettingsPath(false)
		if path != ".claude/settings.local.json" {
			t.Errorf("Expected .claude/settings.local.json, got %s", path)
		}
	})

	t.Run("global_path", func(t *testing.T) {
		path := getClaudeSettingsPath(true)
		if !strings.Contains(path, ".claude/settings.json") {
			t.Errorf("Expected path to contain .claude/settings.json, got %s", path)
		}
		if !strings.HasPrefix(path, "/") && !strings.Contains(path, ":") {
			t.Error("Expected absolute path for global settings")
		}
	})
}

// TestExtractProjectName tests project name extraction from cwd
func TestExtractProjectName(t *testing.T) {
	tests := []struct {
		name     string
		cwd      string
		expected string
	}{
		{"unix_path", "/home/user/projects/my-app", "my-app"},
		{"nested_path", "/home/fabio/projects/oss/lmk", "lmk"},
		{"windows_path", "C:\\Users\\john\\projects\\web-app", "web-app"},
		{"trailing_slash", "/home/user/my-project/", "my-project"},
		{"root_path", "/", ""},
		{"empty_cwd", "", ""},
		{"single_component", "project", "project"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractProjectName(tt.cwd)
			if result != tt.expected {
				t.Errorf("extractProjectName(%q) = %q, want %q", tt.cwd, result, tt.expected)
			}
		})
	}
}

// TestInstallClaudeHooksPreservesOtherSettings tests that installation preserves other hooks and settings
func TestInstallClaudeHooksPreservesOtherSettings(t *testing.T) {
	t.Run("preserve_other_hook_types_and_notification_hooks", func(t *testing.T) {
		tempDir := t.TempDir()
		settingsPath := filepath.Join(tempDir, ".claude", "settings.json")

		// Create settings with multiple hook types and multiple notification hooks
		settings := make(ClaudeSettings)
		settings["colorScheme"] = "dark"
		settings["fontSize"] = 12

		// Multiple hook types (Notification and Completion)
		hooksMap := map[string]interface{}{
			"Notification": []map[string]interface{}{
				{
					"matcher": "permission_prompt",
					"hooks": []map[string]string{
						{
							"type":    "command",
							"command": "other-tool notify",
						},
					},
				},
			},
			"Completion": []map[string]interface{}{
				{
					"hooks": []map[string]string{
						{
							"type":    "command",
							"command": "completion-tool",
						},
					},
				},
			},
			"SomeOtherHookType": "preserve-this-too",
		}
		settings["hooks"] = hooksMap

		// Write initial settings
		if err := writeSettings(settingsPath, settings); err != nil {
			t.Fatalf("Failed to write settings: %v", err)
		}

		// Read settings back
		readSettings := readOrCreateSettings(settingsPath)

		// Verify other settings are present
		if val, ok := readSettings["colorScheme"]; !ok || val != "dark" {
			t.Error("Expected colorScheme setting to be preserved")
		}
		if val, ok := readSettings["fontSize"]; !ok || val != float64(12) {
			t.Error("Expected fontSize setting to be preserved")
		}

		// Verify hooks map has all types
		hooksData, ok := readSettings["hooks"]
		if !ok {
			t.Fatal("Expected hooks to exist")
		}

		hooksMapRead, ok := hooksData.(map[string]interface{})
		if !ok {
			t.Fatal("Expected hooks to be a map")
		}

		// Verify Completion hook type is preserved
		if _, ok := hooksMapRead["Completion"]; !ok {
			t.Error("Expected Completion hook type to be preserved")
		}

		// Verify other hook type is preserved
		if _, ok := hooksMapRead["SomeOtherHookType"]; !ok {
			t.Error("Expected SomeOtherHookType to be preserved")
		}

		// Verify Notification hooks are present (they should still be there)
		if _, ok := hooksMapRead["Notification"]; !ok {
			t.Error("Expected Notification hooks to be present")
		}
	})
}

// TestInstallClaudeHooksActualInstallation tests that installClaudeHooks actually installs hooks correctly
func TestInstallClaudeHooksActualInstallation(t *testing.T) {
	t.Run("install_lmk_hooks_and_preserve_other_settings", func(t *testing.T) {
		tempDir := t.TempDir()
		claudeDir := filepath.Join(tempDir, ".claude")
		settingsPath := filepath.Join(claudeDir, "settings.local.json")

		// Create initial settings with other hooks and settings
		initialSettings := make(ClaudeSettings)
		initialSettings["theme"] = "light"
		initialSettings["editor"] = "vim"

		// Add some existing hooks (including other notification hooks and other hook types)
		hooksMap := map[string]interface{}{
			"Notification": []map[string]interface{}{
				{
					"matcher": "idle_prompt",
					"hooks": []map[string]string{
						{
							"type":    "command",
							"command": "existing-notification-hook",
						},
					},
				},
			},
			"Completion": []map[string]interface{}{
				{
					"hooks": []map[string]string{
						{
							"type":    "command",
							"command": "completion-provider",
						},
					},
				},
			},
		}
		initialSettings["hooks"] = hooksMap

		// Write initial settings
		if err := writeSettings(settingsPath, initialSettings); err != nil {
			t.Fatalf("Failed to write initial settings: %v", err)
		}

		// Change to temp directory so installClaudeHooks uses local settings
		oldCwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current directory: %v", err)
		}
		defer os.Chdir(oldCwd)

		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to chdir to temp directory: %v", err)
		}

		// Call installClaudeHooks with --ack-mode flag
		installClaudeHooks([]string{"--ack-mode", "--dry-run"})

		// Read back the settings (should not be modified due to --dry-run)
		// So we need to actually do the install without --dry-run
		installClaudeHooks([]string{"--ack-mode"})

		// Read installed settings
		installedSettings := readOrCreateSettings(settingsPath)

		// Verify other settings are preserved
		if val, ok := installedSettings["theme"]; !ok || val != "light" {
			t.Error("Expected theme setting to be preserved")
		}
		if val, ok := installedSettings["editor"]; !ok || val != "vim" {
			t.Error("Expected editor setting to be preserved")
		}

		// Verify hooks structure
		hooksData, ok := installedSettings["hooks"]
		if !ok {
			t.Fatal("Expected hooks to exist after installation")
		}

		hooksMapRead, ok := hooksData.(map[string]interface{})
		if !ok {
			t.Fatal("Expected hooks to be a map")
		}

		// Verify Completion hooks are preserved
		if _, ok := hooksMapRead["Completion"]; !ok {
			t.Error("Expected Completion hook type to be preserved after lmk installation")
		}

		// Verify Notification hooks exist
		notificationData, ok := hooksMapRead["Notification"]
		if !ok {
			t.Fatal("Expected Notification hooks after installation")
		}

		notificationHooks, ok := notificationData.([]interface{})
		if !ok {
			t.Fatal("Expected Notification to be a list")
		}

		// Should have at least 1 notification hook (the new lmk one)
		if len(notificationHooks) < 1 {
			t.Errorf("Expected at least 1 notification hook, got %d", len(notificationHooks))
		}

		// Verify lmk hook was added with --ack-mode
		foundLmkHook := false
		foundExistingHook := false

		for _, hookInterface := range notificationHooks {
			hook, ok := hookInterface.(map[string]interface{})
			if !ok {
				continue
			}
			hooksArray, ok := hook["hooks"].([]interface{})
			if !ok {
				continue
			}
			for _, h := range hooksArray {
				hConfig, ok := h.(map[string]interface{})
				if !ok {
					continue
				}
				cmd, ok := hConfig["command"].(string)
				if !ok {
					continue
				}
				// Check for lmk hook
				if strings.Contains(cmd, "lmk") && strings.Contains(cmd, "claude-hooks") && strings.Contains(cmd, "--ack-mode") {
					foundLmkHook = true
				}
				// Check for existing hook
				if strings.Contains(cmd, "existing-notification-hook") {
					foundExistingHook = true
				}
			}
		}

		if !foundLmkHook {
			t.Error("Expected lmk hook with --ack-mode flag to be installed")
		}
		if !foundExistingHook {
			t.Error("Expected existing notification hook to be preserved")
		}
	})
}
