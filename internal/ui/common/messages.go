package common

import "github.com/gdiab/toggl-tui/internal/api"

// Screen identifies which screen to display.
type Screen int

const (
	ScreenSetup Screen = iota
	ScreenDashboard
	ScreenStartTimer
	ScreenManualEntry
	ScreenWeekView
)

// SwitchScreenMsg requests a screen transition.
type SwitchScreenMsg struct{ Screen Screen }

// ErrMsg carries an error to display in the error bar.
type ErrMsg struct{ Err error }

// UserFetchedMsg carries the authenticated user.
type UserFetchedMsg struct{ User api.User }

// WorkspacesFetchedMsg carries the list of workspaces.
type WorkspacesFetchedMsg struct{ Workspaces []api.Workspace }

// ProjectsFetchedMsg carries the list of projects.
type ProjectsFetchedMsg struct{ Projects []api.Project }

// EntriesFetchedMsg carries today's time entries.
type EntriesFetchedMsg struct{ Entries []api.TimeEntry }

// CurrentTimerMsg carries the currently running timer (nil if none).
type CurrentTimerMsg struct{ Entry *api.TimeEntry }

// TimerStartedMsg indicates a timer was started.
type TimerStartedMsg struct{ Entry api.TimeEntry }

// TimerStoppedMsg indicates a timer was stopped.
type TimerStoppedMsg struct{ Entry api.TimeEntry }

// EntryCreatedMsg indicates a manual entry was created.
type EntryCreatedMsg struct{ Entry api.TimeEntry }

// EntryUpdatedMsg indicates an entry was updated.
type EntryUpdatedMsg struct{ Entry api.TimeEntry }

// ConfigSavedMsg indicates config was saved successfully.
type ConfigSavedMsg struct{}

// ClearErrMsg clears the error bar.
type ClearErrMsg struct{}

// TickMsg is sent every second for the running timer display.
type TickMsg struct{}

// RefreshMsg triggers a data refresh.
type RefreshMsg struct{}

// UpdateAvailableMsg carries an update notice (empty if up to date).
type UpdateAvailableMsg struct{ Notice string }

// WeekFetchedMsg carries the weekly summary data.
type WeekFetchedMsg struct{ Summary api.WeekSummary }
