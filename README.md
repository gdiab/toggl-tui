# toggl-tui

A terminal UI for [Toggl Track](https://toggl.com/track/) built with [Bubble Tea](https://github.com/charmbracelet/bubbletea). Start/stop timers, log manual entries, and view today's time log — all from your terminal.

```
Toggl TUI

● Writing docs [toggl-tui]  0:12:34

Today's Entries
  Description                    Project         Duration
  ────────────────────────────────────────────────────────
  Stand-up                       Engineering       0:15:00
  Code review                    Engineering       0:45:00
  Writing docs                   toggl-tui         0:12:34

Total: 1:12:34

? help • s start • m manual • x stop • e edit • r refresh • q quit
```

## Prerequisites

- [Go](https://go.dev/dl/) 1.21 or later

## Install

```bash
# From source
git clone https://github.com/gdiab/toggl-tui.git
cd toggl-tui
go install .

# Or just build
make build
./toggl-tui
```

## First-Run Setup

On first launch, you'll be guided through setup:

1. Enter your Toggl API token (find it at [toggl.com/profile](https://track.toggl.com/profile))
2. Select a workspace (auto-selected if you only have one)

Config is saved to `~/.config/toggl-tui/config.toml`.

## Keyboard Shortcuts

| Context   | Key            | Action              |
|-----------|---------------|----------------------|
| Dashboard | `s`           | Start timer          |
| Dashboard | `m`           | Manual time entry    |
| Dashboard | `x`           | Stop running timer   |
| Dashboard | `e`           | Edit entry description |
| Dashboard | `r`           | Refresh entries      |
| Dashboard | `j`/`k`       | Navigate entries     |
| Dashboard | `?`           | Toggle help          |
| Forms     | `tab`/`shift+tab` | Next/prev field  |
| Forms     | `enter`       | Submit               |
| Forms     | `esc`         | Cancel               |
| Global    | `q`           | Quit                 |
| Global    | `ctrl+c`      | Force quit           |

## Build From Source

```bash
make build          # Build for current platform
make run            # Build and run
make test           # Run tests
make cross          # Cross-compile (darwin/linux/windows)
make install        # Install to $GOPATH/bin
```

## How It Works

- Talks to the [Toggl API v9](https://engineering.toggl.com/docs/) with HTTP basic auth
- Rate-limited to 1 request/second per Toggl's guidelines
- Auto-refreshes every 60 seconds; press `r` for manual refresh
- Starting a timer while one is running will stop the current timer (Toggl behavior)
