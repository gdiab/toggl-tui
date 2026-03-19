package daydetail

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gdiab/toggl-tui/internal/api"
	"github.com/gdiab/toggl-tui/internal/ui/common"
)

// Model is the read-only day detail screen showing time entries for a single day.
type Model struct {
	day          api.DaySummary
	projects     map[int]api.Project
	parentScreen common.Screen
	cursor       int
	width        int
	height       int
}

// New creates a new day detail model. parentScreen controls where esc/b navigates back to.
func New(day api.DaySummary, projects map[int]api.Project, parentScreen common.Screen) Model {
	return Model{
		day:          day,
		projects:     projects,
		parentScreen: parentScreen,
	}
}

// Init returns nil — no async work needed.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "b":
			back := m.parentScreen
			return m, func() tea.Msg {
				return common.SwitchScreenMsg{Screen: back}
			}
		case "j", "down":
			if m.cursor < len(m.day.Entries)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		}
	}

	return m, nil
}

// View renders the day detail screen.
func (m Model) View() string {
	var b strings.Builder

	// Title: day name and date
	title := m.day.Date.Format("Monday, January 2")
	b.WriteString(common.TitleStyle.Render(title))
	b.WriteString("\n")

	// Entries table
	b.WriteString(m.renderEntries())

	// Total
	b.WriteString("\n")
	b.WriteString(common.SubtitleStyle.Render(
		fmt.Sprintf("Total: %s", formatDuration(int(m.day.Total.Seconds()))),
	))

	// Navigation hint
	b.WriteString("\n\n")
	b.WriteString(common.HelpStyle.Render("esc/b back • j/k navigate"))

	return b.String()
}

func (m Model) renderEntries() string {
	if len(m.day.Entries) == 0 {
		return common.MutedStyle.Render("  No entries")
	}

	var b strings.Builder

	// Header — same format as dashboard
	header := fmt.Sprintf("  %-30s %-15s %10s", "Description", "Project", "Duration")
	b.WriteString(common.TableHeaderStyle.Render(header))
	b.WriteString("\n")

	for i, e := range m.day.Entries {
		desc := e.Description
		if desc == "" {
			desc = "(no description)"
		}
		if len(desc) > 28 {
			desc = desc[:28] + ".."
		}

		project := ""
		if e.ProjectID != nil {
			if p, ok := m.projects[*e.ProjectID]; ok {
				project = p.Name
			}
		}
		if len(project) > 13 {
			project = project[:13] + ".."
		}

		dur := formatDuration(e.Duration)
		if e.Duration < 0 {
			startTime, err := time.Parse(time.RFC3339, e.Start)
			if err == nil {
				dur = formatDuration(int(time.Since(startTime).Seconds()))
			}
		}

		line := fmt.Sprintf("  %-30s %-15s %10s", desc, project, dur)
		if i == m.cursor {
			line = common.SelectedRowStyle.Render(line)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

func formatDuration(seconds int) string {
	if seconds < 0 {
		return "running"
	}
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	return fmt.Sprintf("%d:%02d:%02d", h, m, s)
}
