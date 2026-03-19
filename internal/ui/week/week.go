package week

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gdiab/toggl-tui/internal/api"
	"github.com/gdiab/toggl-tui/internal/ui/common"
)

// Model is the weekly summary screen model.
type Model struct {
	client  *api.Client
	summary api.WeekSummary
	cursor  int
	loaded  bool
}

// New creates a new week view model.
func New(client *api.Client) Model {
	return Model{client: client}
}

// Init fetches the week data.
func (m Model) Init() tea.Cmd {
	return m.fetchWeek()
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.loaded && m.cursor < len(m.summary.Days)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "enter":
			if m.loaded && m.cursor >= 0 && m.cursor < len(m.summary.Days) {
				day := m.summary.Days[m.cursor]
				return m, func() tea.Msg {
					return common.WeekDaySelectedMsg{Day: day}
				}
			}
		case "esc", "b":
			return m, func() tea.Msg {
				return common.SwitchScreenMsg{Screen: common.ScreenDashboard}
			}
		}

	case common.WeekFetchedMsg:
		m.summary = msg.Summary
		m.loaded = true
		// Place cursor on today's row (last day in the slice).
		if len(m.summary.Days) > 0 {
			m.cursor = len(m.summary.Days) - 1
		}
	}

	return m, nil
}

// View renders the weekly summary table.
func (m Model) View() string {
	var b strings.Builder

	b.WriteString(common.TitleStyle.Render("Week Summary"))
	b.WriteString("\n")

	if !m.loaded {
		b.WriteString(common.MutedStyle.Render("  Loading..."))
		return b.String()
	}

	if len(m.summary.Days) == 0 {
		b.WriteString(common.MutedStyle.Render("  No data for this week"))
		return b.String()
	}

	// Header
	header := fmt.Sprintf("  %-12s %-12s %10s", "Day", "Date", "Hours")
	b.WriteString(common.TableHeaderStyle.Render(header))
	b.WriteString("\n")

	today := time.Now()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	for i, day := range m.summary.Days {
		dayName := day.Date.Weekday().String()
		dateStr := day.Date.Format("Jan 02")
		hours := formatHours(day.Total)

		line := fmt.Sprintf("  %-12s %-12s %10s", dayName, dateStr, hours)

		isToday := day.Date.Equal(todayDate)

		switch {
		case i == m.cursor && isToday:
			line = common.SelectedRowStyle.Copy().Bold(true).Render(line)
		case i == m.cursor:
			line = common.SelectedRowStyle.Render(line)
		case isToday:
			line = common.SubtitleStyle.Bold(true).Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	// Week total row
	b.WriteString("\n")
	totalLine := fmt.Sprintf("  %-12s %-12s %10s", "", "Total", formatHours(m.summary.Total))
	b.WriteString(common.SubtitleStyle.Render(totalLine))

	// Help
	b.WriteString("\n\n")
	b.WriteString(common.HelpStyle.Render("j/k navigate • enter day detail • esc/b back"))

	return b.String()
}

func (m Model) fetchWeek() tea.Cmd {
	return func() tea.Msg {
		summary, err := m.client.GetWeekEntries()
		if err != nil {
			return common.ErrMsg{Err: fmt.Errorf("fetch week: %w", err)}
		}
		return common.WeekFetchedMsg{Summary: summary}
	}
}

// formatHours formats a duration as "Xh YYm".
func formatHours(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h == 0 && m == 0 {
		return "0h 00m"
	}
	return fmt.Sprintf("%dh %02dm", h, m)
}
