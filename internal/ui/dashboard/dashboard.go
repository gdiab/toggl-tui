package dashboard

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gd/toggl-tui/internal/api"
	"github.com/gd/toggl-tui/internal/ui/common"
)

// Model is the dashboard screen model.
type Model struct {
	client       *api.Client
	workspaceID  int
	entries      []api.TimeEntry
	projects     map[int]api.Project
	currentTimer *api.TimeEntry
	cursor       int
	showHelp     bool
	width        int
	height       int
}

// New creates a new dashboard model.
func New(client *api.Client, workspaceID int) Model {
	return Model{
		client:      client,
		workspaceID: workspaceID,
		projects:    make(map[int]api.Project),
	}
}

// Init fetches initial data.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchEntries(),
		m.fetchCurrentTimer(),
		m.fetchProjects(),
		m.tickCmd(),
		m.autoRefreshCmd(),
	)
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "s":
			return m, func() tea.Msg { return common.SwitchScreenMsg{Screen: common.ScreenStartTimer} }
		case "m":
			return m, func() tea.Msg { return common.SwitchScreenMsg{Screen: common.ScreenManualEntry} }
		case "r":
			return m, tea.Batch(m.fetchEntries(), m.fetchCurrentTimer())
		case "x":
			if m.currentTimer != nil {
				return m, m.stopTimer()
			}
		case "j", "down":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "?":
			m.showHelp = !m.showHelp
		}

	case common.EntriesFetchedMsg:
		m.entries = msg.Entries
		if m.cursor >= len(m.entries) {
			m.cursor = max(0, len(m.entries)-1)
		}

	case common.CurrentTimerMsg:
		m.currentTimer = msg.Entry

	case common.ProjectsFetchedMsg:
		m.projects = make(map[int]api.Project)
		for _, p := range msg.Projects {
			m.projects[p.ID] = p
		}

	case common.TimerStoppedMsg:
		m.currentTimer = nil
		return m, tea.Batch(m.fetchEntries(), m.fetchCurrentTimer())

	case common.TimerStartedMsg, common.EntryCreatedMsg:
		return m, tea.Batch(m.fetchEntries(), m.fetchCurrentTimer())

	case common.TickMsg:
		return m, m.tickCmd()

	case common.RefreshMsg:
		return m, tea.Batch(m.fetchEntries(), m.fetchCurrentTimer(), m.autoRefreshCmd())
	}

	return m, nil
}

func (m Model) fetchEntries() tea.Cmd {
	return func() tea.Msg {
		now := time.Now()
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).UTC().Format(time.RFC3339)
		end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location()).UTC().Format(time.RFC3339)
		entries, err := m.client.GetTimeEntries(start, end)
		if err != nil {
			return common.ErrMsg{Err: fmt.Errorf("fetch entries: %w", err)}
		}
		return common.EntriesFetchedMsg{Entries: entries}
	}
}

func (m Model) fetchCurrentTimer() tea.Cmd {
	return func() tea.Msg {
		entry, err := m.client.GetCurrentTimer()
		if err != nil {
			return common.ErrMsg{Err: fmt.Errorf("fetch timer: %w", err)}
		}
		return common.CurrentTimerMsg{Entry: entry}
	}
}

func (m Model) fetchProjects() tea.Cmd {
	return func() tea.Msg {
		projects, err := m.client.GetProjects(m.workspaceID)
		if err != nil {
			return common.ErrMsg{Err: fmt.Errorf("fetch projects: %w", err)}
		}
		return common.ProjectsFetchedMsg{Projects: projects}
	}
}

func (m Model) stopTimer() tea.Cmd {
	return func() tea.Msg {
		if m.currentTimer == nil {
			return nil
		}
		entry, err := m.client.StopTimer(m.workspaceID, m.currentTimer.ID)
		if err != nil {
			return common.ErrMsg{Err: fmt.Errorf("stop timer: %w", err)}
		}
		return common.TimerStoppedMsg{Entry: entry}
	}
}

func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return common.TickMsg{}
	})
}

func (m Model) autoRefreshCmd() tea.Cmd {
	return tea.Tick(60*time.Second, func(t time.Time) tea.Msg {
		return common.RefreshMsg{}
	})
}

// Projects returns the project map for use by other screens.
func (m Model) Projects() map[int]api.Project {
	return m.projects
}

// Client returns the API client.
func (m Model) Client() *api.Client {
	return m.client
}

// WorkspaceID returns the workspace ID.
func (m Model) WorkspaceID() int {
	return m.workspaceID
}

// View renders the dashboard.
func (m Model) View() string {
	var b strings.Builder

	b.WriteString(common.TitleStyle.Render("Toggl TUI"))
	b.WriteString("\n")

	// Timer status bar
	b.WriteString(m.renderTimerBar())
	b.WriteString("\n\n")

	// Today's entries
	b.WriteString(common.SubtitleStyle.Render("Today's Entries"))
	b.WriteString("\n")
	b.WriteString(m.renderEntries())

	// Total time
	b.WriteString("\n")
	b.WriteString(m.renderTotal())

	// Help
	if m.showHelp {
		b.WriteString("\n\n")
		b.WriteString(m.renderHelp())
	} else {
		b.WriteString("\n\n")
		b.WriteString(common.HelpStyle.Render("? help • s start • m manual • x stop • r refresh • q quit"))
	}

	return b.String()
}

func (m Model) renderTimerBar() string {
	if m.currentTimer == nil {
		return common.TimerStoppedStyle.Render("No timer running")
	}

	desc := m.currentTimer.Description
	if desc == "" {
		desc = "(no description)"
	}

	// Calculate elapsed time
	startTime, err := time.Parse(time.RFC3339, m.currentTimer.Start)
	if err != nil {
		return common.TimerRunningStyle.Render(fmt.Sprintf("● %s (running)", desc))
	}
	elapsed := time.Since(startTime)
	h := int(elapsed.Hours())
	mins := int(elapsed.Minutes()) % 60
	secs := int(elapsed.Seconds()) % 60

	project := ""
	if m.currentTimer.ProjectID != nil {
		if p, ok := m.projects[*m.currentTimer.ProjectID]; ok {
			project = fmt.Sprintf(" [%s]", p.Name)
		}
	}

	return common.TimerRunningStyle.Render(
		fmt.Sprintf("● %s%s  %d:%02d:%02d", desc, project, h, mins, secs),
	)
}

func (m Model) renderEntries() string {
	if len(m.entries) == 0 {
		return common.MutedStyle.Render("  No entries yet today")
	}

	var b strings.Builder

	// Header
	header := fmt.Sprintf("  %-30s %-15s %10s", "Description", "Project", "Duration")
	b.WriteString(common.TableHeaderStyle.Render(header))
	b.WriteString("\n")

	for i, e := range m.entries {
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
			// Running timer — show elapsed
			startTime, err := time.Parse(time.RFC3339, e.Start)
			if err == nil {
				elapsed := time.Since(startTime)
				dur = formatDurationSecs(int(elapsed.Seconds()))
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

func (m Model) renderTotal() string {
	var total int
	for _, e := range m.entries {
		if e.Duration > 0 {
			total += e.Duration
		} else if e.Duration < 0 {
			// Running timer
			startTime, err := time.Parse(time.RFC3339, e.Start)
			if err == nil {
				total += int(time.Since(startTime).Seconds())
			}
		}
	}
	return common.SubtitleStyle.Render(fmt.Sprintf("Total: %s", formatDuration(total)))
}

func (m Model) renderHelp() string {
	help := `Keyboard Shortcuts:
  s     Start timer         m     Manual entry
  x     Stop timer          r     Refresh
  j/k   Navigate            ?     Toggle help
  q     Quit`
	return common.HelpStyle.Render(help)
}

func formatDuration(seconds int) string {
	if seconds < 0 {
		return "running"
	}
	h := seconds / 3600
	m := (seconds % 3600) / 60
	return fmt.Sprintf("%d:%02d", h, m)
}

func formatDurationSecs(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	return fmt.Sprintf("%d:%02d:%02d", h, m, s)
}
