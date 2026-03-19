package timer

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gd/toggl-tui/internal/api"
	"github.com/gd/toggl-tui/internal/ui/common"
)

const minFormHeight = 20

// ManualModel is the manual entry form.
type ManualModel struct {
	client        *api.Client
	workspaceID   int
	descInput     textinput.Model
	startInput    textinput.Model
	durationInput textinput.Model
	projects      []api.Project
	projectIdx    int
	focusIdx      int // 0=desc, 1=start, 2=duration, 3=project, 4=submit
	err           string
	height        int
}

// NewManual creates a new manual entry form.
func NewManual(client *api.Client, workspaceID int, projects []api.Project, height int) ManualModel {
	desc := textinput.New()
	desc.Placeholder = "What did you work on?"
	desc.Focus()
	desc.Width = 40

	start := textinput.New()
	start.Placeholder = "HH:MM (e.g. 09:30)"
	start.Width = 20

	dur := textinput.New()
	dur.Placeholder = "HH:MM (e.g. 1:30)"
	dur.Width = 20

	return ManualModel{
		client:        client,
		workspaceID:   workspaceID,
		descInput:     desc,
		startInput:    start,
		durationInput: dur,
		projects:      projects,
		projectIdx:    -1,
		height:        height,
	}
}

// Init starts the form.
func (m ManualModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages.
func (m ManualModel) Update(msg tea.Msg) (ManualModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		m.err = ""
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return common.SwitchScreenMsg{Screen: common.ScreenDashboard} }
		case "tab":
			m.focusIdx = (m.focusIdx + 1) % 5
			m.updateFocus()
		case "shift+tab":
			m.focusIdx = (m.focusIdx + 4) % 5
			m.updateFocus()
		case "enter":
			if m.focusIdx == 4 {
				return m, m.submit()
			}
		case "left", "h":
			if m.focusIdx == 3 && m.projectIdx > -1 {
				m.projectIdx--
			}
		case "right", "l":
			if m.focusIdx == 3 && m.projectIdx < len(m.projects)-1 {
				m.projectIdx++
			}
		}
	}

	var cmd tea.Cmd
	switch m.focusIdx {
	case 0:
		m.descInput, cmd = m.descInput.Update(msg)
	case 1:
		m.startInput, cmd = m.startInput.Update(msg)
	case 2:
		m.durationInput, cmd = m.durationInput.Update(msg)
	}

	return m, cmd
}

func (m *ManualModel) updateFocus() {
	m.descInput.Blur()
	m.startInput.Blur()
	m.durationInput.Blur()

	switch m.focusIdx {
	case 0:
		m.descInput.Focus()
	case 1:
		m.startInput.Focus()
	case 2:
		m.durationInput.Focus()
	}
}

func parseHHMM(s string) (int, int, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected HH:MM format")
	}
	h, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid hours: %w", err)
	}
	mins, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid minutes: %w", err)
	}
	return h, mins, nil
}

func (m ManualModel) submit() tea.Cmd {
	return func() tea.Msg {
		// Parse start time
		startStr := strings.TrimSpace(m.startInput.Value())
		if startStr == "" {
			return common.ErrMsg{Err: fmt.Errorf("start time is required")}
		}
		sh, sm, err := parseHHMM(startStr)
		if err != nil {
			return common.ErrMsg{Err: fmt.Errorf("start time: %w", err)}
		}

		// Parse duration
		durStr := strings.TrimSpace(m.durationInput.Value())
		if durStr == "" {
			return common.ErrMsg{Err: fmt.Errorf("duration is required")}
		}
		dh, dm, err := parseHHMM(durStr)
		if err != nil {
			return common.ErrMsg{Err: fmt.Errorf("duration: %w", err)}
		}
		durationSecs := dh*3600 + dm*60

		if durationSecs <= 0 {
			return common.ErrMsg{Err: fmt.Errorf("duration must be positive")}
		}

		// Build start/stop times (today, local tz)
		now := time.Now()
		startTime := time.Date(now.Year(), now.Month(), now.Day(), sh, sm, 0, 0, now.Location())
		stopTime := startTime.Add(time.Duration(durationSecs) * time.Second)

		req := api.CreateTimeEntryRequest{
			Description: strings.TrimSpace(m.descInput.Value()),
			Start:       startTime.UTC().Format(time.RFC3339),
			Stop:        stopTime.UTC().Format(time.RFC3339),
			Duration:    durationSecs,
			WorkspaceID: m.workspaceID,
		}
		if m.projectIdx >= 0 && m.projectIdx < len(m.projects) {
			pid := m.projects[m.projectIdx].ID
			req.ProjectID = &pid
		}

		entry, err := m.client.CreateTimeEntry(m.workspaceID, req)
		if err != nil {
			return common.ErrMsg{Err: fmt.Errorf("create entry: %w", err)}
		}
		return common.EntryCreatedMsg{Entry: entry}
	}
}

// View renders the manual entry form.
func (m ManualModel) View() string {
	if m.height > 0 && m.height < minFormHeight {
		return common.ErrorStyle.Render(fmt.Sprintf("Terminal too short (%d rows). Need at least %d rows — please resize.", m.height, minFormHeight))
	}

	var b strings.Builder

	b.WriteString(common.TitleStyle.Render("Manual Time Entry"))
	b.WriteString("\n\n")

	fields := []struct {
		label string
		view  string
		idx   int
	}{
		{"Description", m.descInput.View(), 0},
		{"Start Time", m.startInput.View(), 1},
		{"Duration", m.durationInput.View(), 2},
	}

	for _, f := range fields {
		b.WriteString(common.FormLabelStyle.Render(f.label))
		b.WriteString("\n")
		if m.focusIdx == f.idx {
			b.WriteString(common.FormFieldFocusedStyle.Render(f.view))
		} else {
			b.WriteString(common.FormFieldStyle.Render(f.view))
		}
		b.WriteString("\n\n")
	}

	// Project picker
	b.WriteString(common.FormLabelStyle.Render("Project"))
	b.WriteString("\n")
	projectStr := "(none)"
	if m.projectIdx >= 0 && m.projectIdx < len(m.projects) {
		projectStr = m.projects[m.projectIdx].Name
	}
	if m.focusIdx == 3 {
		b.WriteString(common.FormFieldFocusedStyle.Render(fmt.Sprintf("< %s >", projectStr)))
	} else {
		b.WriteString(common.FormFieldStyle.Render(projectStr))
	}
	b.WriteString("\n\n")

	// Submit
	if m.focusIdx == 4 {
		b.WriteString(common.FormFieldFocusedStyle.Render("[ Create Entry ]"))
	} else {
		b.WriteString(common.FormFieldStyle.Render("[ Create Entry ]"))
	}

	if m.err != "" {
		b.WriteString("\n\n")
		b.WriteString(common.ErrorStyle.Render(m.err))
	}

	b.WriteString("\n\n")
	b.WriteString(common.HelpStyle.Render("tab: next field • enter: submit • esc: cancel"))

	return b.String()
}
