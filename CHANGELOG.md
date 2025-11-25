# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Claude Code integration** - `lmk claude-hooks` subcommand for desktop notifications from Claude Code
  - Modal dialogs when Claude needs your attention (permission prompts 🔐, idle alerts ⏱️, auth success ✅, etc.)
  - Shows project name so you know which session needs attention
  - Install with `lmk claude-hooks install` (project-local) or `--global` (all projects)
  - Debug logging to `/tmp/lmk-claude-hooks.log` for troubleshooting
  - Filter by notification type: `--type permission_prompt,idle_prompt`
  - `--dry-run` to preview changes, `--uninstall` to remove hooks

## [2.0.0] - 2025-11-17

**🤖 This release was brought to you by Claude Sonnet 4.5**

Complete modernization of lmk from 2019 codebase. See [detailed implementation notes](docs/releases/v2.0.0.md).

### Added

- **Timer mode** - Built-in pomodoro timer with `-t` / `-timer` flag
  - `lmk -t 25m -m "Pomodoro done!"` for 25-minute work sessions
  - Supports any duration: 5s, 25m, 1h30m, etc.
- **Execution time display** - Shows how long commands took to run
- **Exit code display** - Failed commands show their exit code
- **Version flag** - `lmk -version` to show current version
- **Windows support** - PowerShell-based dialogs
- **ARM64 support** - Builds for Apple Silicon, Linux ARM64, Windows ARM64
- **Configurable delay** - 3 second delay before showing dialog (prevents accidental dismissal while typing)
- **Test suite** - Unit and integration tests (39% coverage)
- **Dry-run mode** - `LMK_DRY_RUN=1` environment variable for testing without dialogs
- **GitHub Actions CI/CD** - Automated testing and releases

### Changed

- **Modal dialogs instead of notifications** - BREAKING CHANGE
  - No more `notify-send` with 30-second loops
  - Dialogs stay on top and block until acknowledged
  - No need to press Enter in terminal anymore
- **Dialog tools priority** (Linux):
  1. `yad` - Recommended, has proper always-on-top support
  2. `zenity` - Fallback, uses question dialogs
  3. `kdialog` - KDE environment support
  4. `notify-send` - Last resort (v1.0.0 behavior: shows notification + waits for Enter)
- **Migrated to Go modules** - Now requires Go 1.23+
- **Better dialog UX** - Centered, properly sized (450×150), with padding
- **Improved help text** - Shows both command and timer modes

### Removed

- Travis CI configuration (replaced with GitHub Actions)
- Notification loop behavior (replaced with modal dialogs)
- Go 1.12 support (now requires Go 1.23+)

### Fixed

- Dialogs can no longer be hidden behind other windows
- Proper string escaping for AppleScript and PowerShell
- Better error messages with clear examples

### Platform Support

- Linux: amd64, arm64 (requires yad/zenity/kdialog)
- macOS: amd64, arm64 (uses built-in osascript)
- Windows: amd64, arm64 (uses built-in PowerShell)

---

## [1.0.0] - 2019-06-26

### Added
- Support for running without a command (just show notification)
- `lmk -m "message"` now works standalone

## [0.2.0] - 2019-05-02

### Changed
- Updated to Go 1.12.1

## [0.1.0] - 2016-03-25

### Added
- macOS support with osascript
- Cross-compilation support with Go 1.6
- Travis CI for automated releases
- Binary releases for Linux and macOS

## [0.0.1] - 2014-02-16

Initial release.

### Added
- Basic command execution with notifications
- Linux support with notify-send
- Custom success messages with `-m` flag
- Notification loop every 30 seconds
- Press Enter to dismiss

[2.0.0]: https://github.com/fgrehm/lmk/compare/v1.0.0...v2.0.0
[1.0.0]: https://github.com/fgrehm/lmk/compare/v0.2.0...v1.0.0
[0.2.0]: https://github.com/fgrehm/lmk/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/fgrehm/lmk/compare/v0.0.1...v0.1.0
[0.0.1]: https://github.com/fgrehm/lmk/releases/tag/v0.0.1
