package common

import "github.com/charmbracelet/lipgloss"

var (
	ColorPrimary   = lipgloss.Color("#FF6B6B")
	ColorSecondary = lipgloss.Color("#4ECDC4")
	ColorMuted     = lipgloss.Color("#666666")
	ColorError     = lipgloss.Color("#FF0000")
	ColorRunning   = lipgloss.Color("#00FF00")

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorError).
			Padding(0, 1)

	TimerRunningStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorRunning)

	TimerStoppedStyle = lipgloss.NewStyle().
				Foreground(ColorMuted)

	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(ColorMuted)

	SelectedRowStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#333333"))

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			MarginTop(1)

	FormLabelStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true)

	FormFieldStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted).
			Padding(0, 1)

	FormFieldFocusedStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Padding(0, 1)

	StatusBarStyle = lipgloss.NewStyle().
			Padding(0, 1).
			MarginBottom(1)
)
