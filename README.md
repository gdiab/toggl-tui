# toggl-tui

A terminal UI for [Toggl Track](https://toggl.com/track/) built with [Bubble Tea](https://github.com/charmbracelet/bubbletea). Start/stop timers, log manual entries, edit entries, view today's time log, and browse your weekly summary — all from your terminal.

```
Toggl TUI

● Writing docs [toggl-tui]  0:12:34

Today's Entries
  Description                    Project         Duration
  ────────────────────────────────────────────────────────
  Stand-up                       Engineering      0:15:00
  Code review                    Engineering      0:45:00
  Writing docs                   toggl-tui        0:12:34

Total: 1:12:34

? help • s start • m manual • x stop • e edit • r refresh • q quit
```

## Prerequisites

- [Go](https://go.dev/dl/) 1.21 or later (for building from source)

## Install

### Pre-built binaries (recommended)

Download the latest release for your platform from [GitHub Releases](https://github.com/gdiab/toggl-tui/releases):

| Platform       | Binary                          |
|----------------|----------------------------------|
| macOS (Apple)  | `toggl-tui-darwin-arm64`         |
| macOS (Intel)  | `toggl-tui-darwin-amd64`         |
| Linux (x86_64) | `toggl-tui-linux-amd64`          |
| Linux (ARM)    | `toggl-tui-linux-arm64`          |
| Windows        | `toggl-tui-windows-amd64.exe`    |

```bash
# Example: macOS Apple Silicon
curl -Lo toggl-tui https://github.com/gdiab/toggl-tui/releases/latest/download/toggl-tui-darwin-arm64
chmod +x toggl-tui
mv toggl-tui /usr/local/bin/
```

### go install

```bash
go install github.com/gdiab/toggl-tui@latest
```

This installs to `$GOPATH/bin` (usually `~/go/bin`). Make sure it's in your `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### From source

```bash
git clone https://github.com/gdiab/toggl-tui.git
cd toggl-tui
make install    # Builds and copies to ~/go/bin/toggl-tui
```

## First-Run Setup

On first launch, you'll be guided through setup:

1. Enter your Toggl API token (find it at [toggl.com/profile](https://track.toggl.com/profile))
2. Select a workspace (auto-selected if you only have one)

Config is saved to `~/.config/toggl-tui/config.toml`.

## Keyboard Shortcuts

| Context     | Key              | Action                       |
|-------------|-----------------|-------------------------------|
| Dashboard   | `s`             | Start timer                   |
| Dashboard   | `m`             | Manual time entry             |
| Dashboard   | `w`             | Weekly summary view           |
| Dashboard   | `x`             | Stop running timer            |
| Dashboard   | `e`             | Edit entry (description + project) |
| Dashboard   | `r`             | Refresh entries               |
| Dashboard   | `j`/`k`         | Navigate entries              |
| Dashboard   | `?`             | Toggle help                   |
| Week view   | `j`/`k`         | Navigate days                 |
| Week view   | `enter`         | View day detail               |
| Week view   | `esc`/`b`       | Back to dashboard             |
| Day detail  | `j`/`k`         | Navigate entries              |
| Day detail  | `esc`/`b`       | Back to previous screen       |
| Edit mode   | `tab`           | Switch between desc/project   |
| Edit mode   | `h`/`l`         | Change project                |
| Edit mode   | `enter`         | Save changes                  |
| Edit mode   | `esc`           | Cancel                        |
| Forms       | `tab`/`shift+tab` | Next/prev field             |
| Forms       | `enter`         | Submit                        |
| Forms       | `esc`           | Cancel                        |
| Global      | `q`             | Quit                          |
| Global      | `ctrl+c`        | Force quit                    |

## Build From Source

```bash
make build          # Build for current platform
make run            # Build and run
make test           # Run tests
make install        # Install to $GOPATH/bin/toggl-tui
make uninstall      # Remove from $GOPATH/bin
make cross          # Cross-compile all platforms
make clean          # Remove build artifacts
```

## Releasing

Releases are automated with [GoReleaser](https://goreleaser.com/). To create a release:

```bash
git tag v0.1.0
git push origin v0.1.0
# GoReleaser builds binaries and creates the GitHub Release
```

To test locally: `goreleaser release --snapshot --clean`

## How It Works

- Talks to the [Toggl API v9](https://engineering.toggl.com/docs/) with HTTP basic auth
- Rate-limited to 1 request/second per Toggl's guidelines
- Auto-refreshes every 60 seconds; press `r` for manual refresh
- Starting a timer while one is running will stop the current timer (Toggl behavior)
