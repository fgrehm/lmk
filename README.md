# Let me know

`lmk` is a simple command line tool written in Go that shows a dialog when a long-running command finishes.

## What's New in v2.0.0

- **Modern Go**: Updated to Go 1.23+ with module support
- **Better Notifications**: Switched from repeating notifications to modal dialogs using zenity (Linux), native dialogs (macOS), and PowerShell (Windows)
- **Timer Mode**: Built-in pomodoro timer support (`--timer 25m`)
- **More Information**: Shows exit code and execution duration
- **Multi-platform**: Support for Linux (amd64/arm64), macOS (amd64/arm64), and Windows (amd64/arm64)
- **GitHub Actions**: Automated builds and releases with GoReleaser

## Why?

How often do you run a command that takes a long time to complete, switch to something else, and forget about it? Even worse, what if it errored along the way and you didn't notice?

Throughout the day you might run many `npm install`s, `docker build`s, test suites, etc. that take more than a few seconds. With `lmk`, you'll get a clear notification when they complete - and it won't disappear until you acknowledge it.

## How does it work?

Let's say you want to run tests that take 5 minutes. With `lmk`:

```bash
lmk npm test
```

When the command finishes, a dialog will appear showing:
- ✅ Success or ❌ failure
- Command that was run
- Exit code (if failed)
- Total execution time

The dialog **blocks and waits** for you to click OK - no more missed notifications!

## Installation

### Binary releases

Download the latest [compiled binaries](https://github.com/fgrehm/lmk/releases) for your platform and add it to your `$PATH`.

### Homebrew

```sh
brew tap fgrehm/lmk
brew install lmk
```

### From source

```sh
go install github.com/fgrehm/lmk@latest
```

## Usage

```
Usage: lmk [options...] command
   or: lmk -t <duration> [-m <text>]

Options:
  -m            Message to display in case of success, defaults to "[command] has completed successfully"
  -t, -timer    Timer duration (e.g., 25m, 1h30m, 90s) - runs a countdown timer instead of a command
  -version      Show version information
```

### Examples

#### Running Commands
```bash
# Run tests and get notified
lmk npm test

# Custom success message
lmk -m "Build completed!" make build

# Works with any command
lmk cargo build --release
lmk docker compose up
lmk bundle install

# Just show a notification (no command)
lmk -m "Time to take a break!"
```

#### Pomodoro Timer (Poor Man's Edition)
```bash
# Classic 25-minute pomodoro
lmk -t 25m -m "Pomodoro done! Time for a break"

# 5-minute break
lmk -t 5m -m "Break over, back to work!"

# Long break (15 minutes)
lmk -t 15m -m "Long break finished"

# Custom durations
lmk -t 1h30m -m "Deep work session complete"
lmk -t 90s -m "Quick break done"

# Simple timer with default message
lmk -t 10m
```

**Pro tip**: Create shell aliases for your pomodoro workflow:
```bash
alias pomo='lmk -t 25m -m "Pomodoro complete! Take a 5min break 🍅"'
alias short-break='lmk -t 5m -m "Break over! Time to focus 💪"'
alias long-break='lmk -t 15m -m "Long break done! Ready for another session?"'
```

### Configuration

**Delay before showing dialog** (prevents accidental dismissal while typing):
```bash
# Default: 3 second delay
lmk npm test

# Custom delay
LMK_DELAY=5s lmk npm test           # Longer delay if you type fast
LMK_DELAY=1s lmk npm test           # Shorter delay
LMK_DELAY=0s lmk npm test           # No delay (instant dialog)
```

**Dry-run mode** - See what lmk would do without showing dialogs:
```bash
LMK_DRY_RUN=1 lmk npm test          # Test without actual dialogs
LMK_DRY_RUN=1 lmk -t 25m -m "Test"  # Test timer without dialog
```

Output shows what would happen:
```
[DRY RUN] Dialog message: ✅ npm test has completed successfully
[DRY RUN] Is error: false
```

### Platform Dependencies

- **Linux**: Requires one of the following (in order of preference):
  - `yad` - **Recommended**, has proper always-on-top support
    - Install: `apt install yad` or `yum install yad`
  - `zenity` - GTK dialogs
    - Install: `apt install zenity` or `yum install zenity`
  - `kdialog` - KDE environments
    - Install: `apt install kdialog`
  - `notify-send` - Last resort fallback (requires pressing Enter to dismiss)
    - Install: `apt install libnotify-bin`
- **macOS**: Built-in `osascript` (no installation needed)
- **Windows**: Built-in PowerShell (no installation needed)


## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request
