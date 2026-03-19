package timer

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gd/toggl-tui/internal/api"
	"github.com/gd/toggl-tui/internal/ui/common"
)

// StartModel is the start-timer form.
type StartModel struct {
	client      *api.Client
	workspaceID int
	descInput   textinput.Model
	projects    []api.Project
	projectIdx  int
	focusIdx    int // 0=description, 1=project, 2=submit
	err         string
}

// NewStart creates a new start-timer form.
func NewStart(client *api.Client, workspaceID int, projects []api.Project) StartModel {
	ti := textinput.New()
	ti.Placeholder = "What are you working on?"
	ti.Focus()
	ti.Width = 40

	return StartModel{
		client:      client,
		workspaceID: workspaceID,
		descInput:   ti,
		projects:    projects,
		projectIdx:  -1, // no project selected
	}
}

// Init starts the form.
func (m StartModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages.
func (m StartModel) Update(msg tea.Msg) (StartModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.err = ""
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return common.SwitchScreenMsg{Screen: common.ScreenDashboard} }
		case "tab":
			m.focusIdx = (m.focusIdx + 1) % 3
			m.updateFocus()
		case "shift+tab":
			m.focusIdx = (m.focusIdx + 2) % 3
			m.updateFocus()
		case "enter":
			if m.focusIdx == 2 {
				return m, m.submit()
			}
		case "left", "h":
			if m.focusIdx == 1 && m.projectIdx > -1 {
				m.projectIdx--
			}
		case "right", "l":
			if m.focusIdx == 1 && m.projectIdx < len(m.projects)-1 {
				m.projectIdx++
			}
		}
	}

	if m.focusIdx == 0 {
		var cmd tea.Cmd
		m.descInput, cmd = m.descInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *StartModel) updateFocus() {
	if m.focusIdx == 0 {
		m.descInput.Focus()
	} else {
		m.descInput.Blur()
	}
}

func (m StartModel) submit() tea.Cmd {
	return func() tea.Msg {
		now := time.Now().UTC().Format(time.RFC3339)
		req := api.CreateTimeEntryRequest{
			Description: strings.TrimSpace(m.descInput.Value()),
			Start:       now,
			Duration:    -1, // running timer
			WorkspaceID: m.workspaceID,
		}
		if m.projectIdx >= 0 && m.projectIdx < len(m.projects) {
			pid := m.projects[m.projectIdx].ID
			req.ProjectID = &pid
		}
		entry, err := m.client.CreateTimeEntry(m.workspaceID, req)
		if err != nil {
			return common.ErrMsg{Err: fmt.Errorf("start timer: %w", err)}
		}
		return common.TimerStartedMsg{Entry: entry}
	}
}

// View renders the start timer form.
func (m StartModel) View() string {
	var b strings.Builder

	b.WriteString(common.TitleStyle.Render("Start Timer"))
	b.WriteString("\n\n")

	// Description
	label := common.FormLabelStyle.Render("Description")
	b.WriteString(label)
	b.WriteString("\n")
	if m.focusIdx == 0 {
		b.WriteString(common.FormFieldFocusedStyle.Render(m.descInput.View()))
	} else {
		b.WriteString(common.FormFieldStyle.Render(m.descInput.View()))
	}
	b.WriteString("\n\n")

	// Project picker
	label = common.FormLabelStyle.Render("Project")
	b.WriteString(label)
	b.WriteString("\n")
	projectStr := "(none)"
	if m.projectIdx >= 0 && m.projectIdx < len(m.projects) {
		projectStr = m.projects[m.projectIdx].Name
	}
	if m.focusIdx == 1 {
		b.WriteString(common.FormFieldFocusedStyle.Render(fmt.Sprintf("< %s >", projectStr)))
	} else {
		b.WriteString(common.FormFieldStyle.Render(projectStr))
	}
	b.WriteString("\n\n")

	// Submit
	if m.focusIdx == 2 {
		b.WriteString(common.FormFieldFocusedStyle.Render("[ Start Timer ]"))
	} else {
		b.WriteString(common.FormFieldStyle.Render("[ Start Timer ]"))
	}

	if m.err != "" {
		b.WriteString("\n\n")
		b.WriteString(common.ErrorStyle.Render(m.err))
	}

	b.WriteString("\n\n")
	b.WriteString(common.HelpStyle.Render("tab: next field • enter: submit • esc: cancel"))

	return b.String()
}
