package main

import (
	"flag"
	"fmt"
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

Options:
  -m            Message to display in case of success, defaults to "[command] has completed successfully"
  -t, -timer    Timer duration (e.g., 25m, 1h30m, 90s) - runs a countdown timer instead of a command
  -version      Show version information

Examples:
  lmk npm test                          Run command and notify when done
  lmk -t 25m -m "Pomodoro done!"       25 minute timer
  lmk -t 5m -m "Break over!"           5 minute break timer
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
