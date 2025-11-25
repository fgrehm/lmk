package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var (
	flagMessage = flag.String("m", "", "")
	flagTimer   string
	flagVersion = flag.Bool("version", false, "")

	version        = "2.0.0"
	defaultMessage = "%s has completed successfully"
)

var usage = `Usage: lmk [options...] command
   or: lmk -t <duration> [-m <text>]
   or: lmk claude-hooks [install [options]]

Options:
  -m            Message to display in case of success, defaults to "[command] has completed successfully"
  -t, -timer    Timer duration (e.g., 25m, 1h30m, 90s) - runs a countdown timer instead of a command
  -version      Show version information

Subcommands:
  claude-hooks           Process Claude Code notification hooks (reads JSON from stdin)
  claude-hooks install   Install lmk hooks into Claude Code settings
    --global               Install to ~/.claude/settings.json (default: .claude/settings.local.json)
    --type TYPES           Only install for specific notification types (comma-separated)
    --uninstall            Remove lmk hooks from configuration
    --dry-run              Show what would be changed without modifying files

Examples:
  lmk npm test                                    Run command and notify when done
  lmk -t 25m -m "Pomodoro done!"                 25 minute timer
  lmk -t 5m -m "Break over!"                     5 minute break timer
  lmk claude-hooks install                        Install Claude Code hooks (project-local)
  lmk claude-hooks install --global               Install Claude Code hooks (globally)
  lmk claude-hooks install --type permission_prompt,idle_prompt
                                                  Install hooks for specific notification types
`

func init() {
	flag.StringVar(&flagTimer, "t", "", "")
	flag.StringVar(&flagTimer, "timer", "", "")
}

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}
	flag.Parse()

	if *flagVersion {
		fmt.Printf("lmk version %s\n", version)
		os.Exit(0)
	}

	// Check for claude-hooks subcommand
	args := flag.Args()
	if len(args) > 0 && args[0] == "claude-hooks" {
		handleClaudeHooks(args[1:])
		return
	}

	// Handle timer mode
	if flagTimer != "" {
		runTimer(flagTimer, *flagMessage)
		return
	}

	cmd := flag.Args()

	var msg string
	var duration time.Duration
	var isError bool

	if len(cmd) > 0 {
		executable, args := getExecutableAndArgs(cmd)
		log.Printf("Running %s", cmd)
		start := time.Now()
		err := run(executable, args...)
		duration = time.Since(start)
		msg, _, isError = getMessageAndExitCode(err, cmd, duration)
	} else {
		log.Print("Nothing to run, showing notification")
		if *flagMessage != "" {
			msg = *flagMessage
		} else {
			msg = "Take a look at your terminal"
		}
		isError = false
	}

	showDialog(msg, isError)
}

func runTimer(timerDuration string, message string) {
	duration, err := time.ParseDuration(timerDuration)
	if err != nil {
		log.Fatalf("Invalid timer duration '%s': %v\nExamples: 25m, 1h30m, 90s", timerDuration, err)
	}

	if duration <= 0 {
		log.Fatalf("Timer duration must be positive, got: %s", timerDuration)
	}

	log.Printf("Timer started for %s", duration)
	time.Sleep(duration)

	msg := message
	if msg == "" {
		msg = fmt.Sprintf("⏰ Timer finished!\n\nDuration: %s", formatDuration(duration))
	} else {
		msg = fmt.Sprintf("⏰ %s\n\nDuration: %s", msg, formatDuration(duration))
	}

	showDialog(msg, false)
}

func run(executable string, args ...string) error {
	cmd := exec.Command(executable, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	return cmd.Run()
}

func getExecutableAndArgs(cmd []string) (string, []string) {
	if len(cmd) == 0 {
		log.Fatalf("No command was provided to lmk")
	}

	executable, lookErr := exec.LookPath(cmd[0])
	if lookErr != nil {
		log.Fatal(lookErr)
	}
	return executable, cmd[1:]
}

func getMessageAndExitCode(err error, cmd []string, duration time.Duration) (msg string, exitCode int, isError bool) {
	durationStr := formatDuration(duration)

	if err != nil {
		exitCode = 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		isError = true
		msg = fmt.Sprintf("❌ %s failed!\n\nExit code: %d\nDuration: %s", strings.Join(cmd, " "), exitCode, durationStr)
	} else {
		exitCode = 0
		isError = false
		if *flagMessage != "" {
			msg = *flagMessage
		} else {
			msg = fmt.Sprintf("✅ %s\n\nDuration: %s", fmt.Sprintf(defaultMessage, strings.Join(cmd, " ")), durationStr)
		}
	}

	return
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

func showDialog(msg string, isError bool) {
	// Dry-run mode for testing and debugging
	if os.Getenv("LMK_DRY_RUN") != "" {
		fmt.Fprintf(os.Stderr, "[DRY RUN] Dialog message: %s\n", msg)
		fmt.Fprintf(os.Stderr, "[DRY RUN] Is error: %t\n", isError)
		return
	}

	// Delay before showing dialog to prevent accidental dismissal
	// Gives user time to finish typing in other apps before dialog steals focus
	// Can be customized via LMK_DELAY environment variable
	delay := 3000 * time.Millisecond // Default: 3 seconds
	if delayStr := os.Getenv("LMK_DELAY"); delayStr != "" {
		if customDelay, err := time.ParseDuration(delayStr); err == nil {
			delay = customDelay
		}
	}

	if delay > 0 {
		log.Printf("Waiting %v before showing dialog...", delay)
		time.Sleep(delay)
	}

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		// Try yad first (has best always-on-top support)
		if _, err := exec.LookPath("yad"); err == nil {
			button := "gtk-ok:0"
			image := "dialog-information"
			if isError {
				image = "dialog-error"
			}
			// yad has proper --on-top and --center support with better sizing
			cmd = exec.Command("yad",
				"--text="+msg,
				"--title=lmk",
				"--width=450",
				"--height=150",
				"--center",
				"--button="+button,
				"--image="+image,
				"--on-top",
				"--no-escape",
				"--borders=10")
		} else if _, err := exec.LookPath("zenity"); err == nil {
			// Fallback to zenity
			// Use question dialog which stays on top better than info/error dialogs
			cmd = exec.Command("zenity", "--question", "--title=lmk", "--text="+msg, "--width=400",
				"--ok-label=OK", "--no-cancel", "--ellipsize")
		} else if _, err := exec.LookPath("kdialog"); err == nil {
			// Fallback to kdialog for KDE
			dialogType := "--msgbox"
			if isError {
				dialogType = "--error"
			}
			cmd = exec.Command("kdialog", dialogType, msg, "--title", "lmk")
		} else if _, err := exec.LookPath("notify-send"); err == nil {
			// Last resort: notify-send with Enter prompt (v1.0.0 behavior)
			log.Printf("Warning: No dialog tools found (yad/zenity/kdialog)")
			log.Printf("Falling back to notify-send + Enter prompt")
			showNotificationAndWait(msg, isError)
			return
		} else {
			log.Printf("Error: No notification tools found!")
			log.Printf("Please install one of: yad, zenity, kdialog, or notify-send")
			log.Printf("Message: %s", msg)
			return
		}

	case "darwin":
		// macOS: use osascript with dialog - make it giving application
		// This brings the dialog to the front
		icon := "note"
		if isError {
			icon = "stop"
		}
		script := fmt.Sprintf(`tell application "System Events"
	activate
	display dialog "%s" with title "lmk" with icon %s buttons {"OK"} default button "OK" giving up after 3600
end tell`,
			escapeAppleScript(msg), icon)
		cmd = exec.Command("osascript", "-e", script)

	case "windows":
		// Windows: use PowerShell dialogs with TopMost property
		icon := "Information"
		if isError {
			icon = "Error"
		}
		// Create a form-based messagebox that stays on top
		psScript := fmt.Sprintf(`Add-Type -AssemblyName System.Windows.Forms; $form = New-Object System.Windows.Forms.Form; $form.TopMost = $true; $form.WindowState = 'Minimized'; $form.Show(); [System.Windows.Forms.MessageBox]::Show($form, '%s', 'lmk', 'OK', '%s'); $form.Close()`,
			escapeWindowsString(msg), icon)
		cmd = exec.Command("powershell", "-Command", psScript)

	default:
		log.Printf("Unsupported platform: %s", runtime.GOOS)
		log.Printf("Message: %s", msg)
		return
	}

	log.Print("Showing dialog")

	// Show what would be executed in dry-run (after cmd is built)
	if os.Getenv("LMK_DRY_RUN") != "" {
		fmt.Fprintf(os.Stderr, "[DRY RUN] Would execute: %s %v\n", cmd.Path, cmd.Args[1:])
		return
	}

	if err := cmd.Run(); err != nil {
		log.Printf("Error showing dialog: %v", err)
		log.Printf("Message: %s", msg)
	}
}

func escapeAppleScript(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

func escapeWindowsString(s string) string {
	s = strings.ReplaceAll(s, "'", "''")
	s = strings.ReplaceAll(s, "\n", "`n")
	return s
}

func showNotificationAndWait(msg string, isError bool) {
	// Dry-run mode for testing and debugging
	if os.Getenv("LMK_DRY_RUN") != "" {
		fmt.Fprintf(os.Stderr, "[DRY RUN] Would send notification: %s\n", msg)
		fmt.Fprintf(os.Stderr, "[DRY RUN] Would wait for Enter\n")
		return
	}

	// Fallback to v1.0.0 behavior: notify-send + wait for Enter
	icon := "emblem-default"
	if isError {
		icon = "dialog-error"
	}

	log.Print("Sending notification (press Enter to dismiss)")
	exec.Command("notify-send", "-i", icon, "lmk", msg).Run()

	fmt.Println("\nPress Enter to dismiss...")
	fmt.Scanln()
}

// ClaudeHookPayload represents the JSON payload from Claude Code hooks
type ClaudeHookPayload struct {
	SessionID        string `json:"session_id"`
	TranscriptPath   string `json:"transcript_path"`
	Cwd              string `json:"cwd"`
	PermissionMode   string `json:"permission_mode"`
	HookEventName    string `json:"hook_event_name"`
	Message          string `json:"message"`
	NotificationType string `json:"notification_type"`
}

// ClaudeSettings represents Claude Code settings file structure
type ClaudeSettings struct {
	Hooks *ClaudeHooks `json:"hooks,omitempty"`
}

// ClaudeHooks represents the hooks section of Claude settings
type ClaudeHooks struct {
	Notification []NotificationHook `json:"Notification,omitempty"`
}

// NotificationHook represents a single notification hook configuration
type NotificationHook struct {
	Matcher string       `json:"matcher,omitempty"`
	Hooks   []HookConfig `json:"hooks"`
}

// HookConfig represents a hook command configuration
type HookConfig struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// handleClaudeHooks processes Claude Code hook events from stdin or handles install subcommand
func handleClaudeHooks(args []string) {
	// Check for install subcommand
	if len(args) > 0 && args[0] == "install" {
		installClaudeHooks(args[1:])
		return
	}

	// Otherwise, process hook payload from stdin
	processHookPayload()
}

// processHookPayload reads and processes a hook payload from stdin
func processHookPayload() {
	// Read JSON from stdin
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("Error reading stdin: %v", err)
	}

	// Parse JSON
	var payload ClaudeHookPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	// Validate required fields
	if payload.NotificationType == "" {
		log.Fatalf("Missing required field: notification_type")
	}
	if payload.Message == "" {
		log.Fatalf("Missing required field: message")
	}

	// Get icon based on notification type
	icon := getNotificationIcon(payload.NotificationType)

	// Format message
	msg := fmt.Sprintf("%s Claude Code\n\n%s", icon, payload.Message)

	// Claude Code hooks need immediate feedback - disable delay
	os.Setenv("LMK_DELAY", "0s")

	// Show dialog (notifications are informational, not errors)
	showDialog(msg, false)
}

// getNotificationIcon returns an emoji icon for the notification type
func getNotificationIcon(notificationType string) string {
	switch notificationType {
	case "permission_prompt":
		return "🔐"
	case "idle_prompt":
		return "⏱️"
	case "auth_success":
		return "✅"
	case "elicitation_dialog":
		return "📝"
	default:
		return "🤖"
	}
}

// installClaudeHooks handles the install subcommand
func installClaudeHooks(args []string) {
	// Parse install flags
	installFlags := flag.NewFlagSet("install", flag.ExitOnError)
	global := installFlags.Bool("global", false, "Install to ~/.claude/settings.json instead of project-local")
	typesStr := installFlags.String("type", "", "Comma-separated list of notification types to install")
	uninstall := installFlags.Bool("uninstall", false, "Remove lmk hooks from configuration")
	dryRun := installFlags.Bool("dry-run", false, "Show what would be changed without modifying files")

	installFlags.Parse(args)

	// Parse types if provided
	var types []string
	if *typesStr != "" {
		types = strings.Split(*typesStr, ",")
		// Validate types
		validTypes := map[string]bool{
			"permission_prompt":  true,
			"idle_prompt":        true,
			"auth_success":       true,
			"elicitation_dialog": true,
		}
		for _, t := range types {
			t = strings.TrimSpace(t)
			if !validTypes[t] {
				log.Fatalf("Invalid notification type: %s\nValid types: permission_prompt, idle_prompt, auth_success, elicitation_dialog", t)
			}
		}
	}

	// Get settings file path
	settingsPath := getClaudeSettingsPath(*global)

	// Read or create settings
	settings := readOrCreateSettings(settingsPath)

	if *uninstall {
		// Remove lmk hooks
		settings.Hooks = removeLmkHooks(settings.Hooks)
	} else {
		// Build lmk hook configuration
		command := "lmk claude-hooks"
		if *typesStr != "" {
			command += " --type " + *typesStr
		}

		lmkHook := NotificationHook{
			Hooks: []HookConfig{
				{Type: "command", Command: command},
			},
		}

		// Add or update
		settings.Hooks = addOrUpdateLmkHook(settings.Hooks, lmkHook)
	}

	// Write back to file
	if !*dryRun {
		if err := writeSettings(settingsPath, settings); err != nil {
			log.Fatalf("Failed to write settings: %v", err)
		}
	}

	// Report success
	printInstallSummary(settingsPath, types, *uninstall, *dryRun)
}

// getClaudeSettingsPath returns the path to the Claude settings file
func getClaudeSettingsPath(global bool) string {
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
		return fmt.Sprintf("%s/.claude/settings.json", home)
	}
	return ".claude/settings.local.json"
}

// readOrCreateSettings reads existing settings or creates empty structure
func readOrCreateSettings(path string) ClaudeSettings {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create empty settings
			return ClaudeSettings{}
		}
		log.Fatalf("Failed to read settings file: %v", err)
	}

	var settings ClaudeSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		log.Fatalf("Failed to parse settings file: %v", err)
	}

	return settings
}

// writeSettings writes settings to file with proper formatting
func writeSettings(path string, settings ClaudeSettings) error {
	// Ensure directory exists
	dir := fmt.Sprintf("%s", path[:strings.LastIndex(path, "/")])
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal with indentation
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// removeLmkHooks removes lmk hooks from the configuration
func removeLmkHooks(hooks *ClaudeHooks) *ClaudeHooks {
	if hooks == nil {
		return nil
	}

	var filtered []NotificationHook
	for _, hook := range hooks.Notification {
		// Keep hooks that don't have lmk command
		hasLmk := false
		for _, h := range hook.Hooks {
			if strings.Contains(h.Command, "lmk claude-hooks") {
				hasLmk = true
				break
			}
		}
		if !hasLmk {
			filtered = append(filtered, hook)
		}
	}

	if len(filtered) == 0 {
		return nil
	}

	return &ClaudeHooks{Notification: filtered}
}

// addOrUpdateLmkHook adds or updates the lmk hook in the configuration
func addOrUpdateLmkHook(hooks *ClaudeHooks, lmkHook NotificationHook) *ClaudeHooks {
	if hooks == nil {
		hooks = &ClaudeHooks{}
	}

	// Remove existing lmk hooks first
	hooks = removeLmkHooks(hooks)
	if hooks == nil {
		hooks = &ClaudeHooks{}
	}

	// Add new lmk hook
	hooks.Notification = append(hooks.Notification, lmkHook)

	return hooks
}

// printInstallSummary prints a summary of the installation
func printInstallSummary(path string, types []string, uninstall bool, dryRun bool) {
	if dryRun {
		fmt.Println("🔍 Dry run mode - no files were modified")
		fmt.Println()
	}

	if uninstall {
		if dryRun {
			fmt.Printf("Would remove lmk hooks from: %s\n", path)
		} else {
			fmt.Println("✅ Claude Code hooks uninstalled successfully!")
			fmt.Println()
			fmt.Printf("Configuration updated: %s\n", path)
		}
	} else {
		if dryRun {
			fmt.Printf("Would install lmk hooks to: %s\n", path)
		} else {
			fmt.Println("✅ Claude Code hooks installed successfully!")
			fmt.Println()
			fmt.Printf("Configuration updated: %s\n", path)
		}
		fmt.Println()
		fmt.Println("Installed hooks:")
		if len(types) > 0 {
			fmt.Printf("  - Notification (types: %s)\n", strings.Join(types, ", "))
		} else {
			fmt.Println("  - Notification (all types)")
		}
		fmt.Println()
		fmt.Println("Command: lmk claude-hooks")
		fmt.Println()
		fmt.Println("To test: Start a new Claude Code session and trigger a notification")
	}
}
