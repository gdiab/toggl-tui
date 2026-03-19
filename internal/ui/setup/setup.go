package setup

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gd/toggl-tui/internal/api"
	"github.com/gd/toggl-tui/internal/config"
	"github.com/gd/toggl-tui/internal/ui/common"
)

type step int

const (
	stepToken step = iota
	stepValidating
	stepWorkspace
	stepSaving
)

// Model is the setup screen model.
type Model struct {
	step       step
	tokenInput textinput.Model
	workspaces []api.Workspace
	cursor     int
	err        string
	client     *api.Client
}

// New creates a new setup screen model.
func New() Model {
	ti := textinput.New()
	ti.Placeholder = "Enter your Toggl API token"
	ti.EchoMode = textinput.EchoPassword
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 40

	return Model{
		step:       stepToken,
		tokenInput: ti,
	}
}

// Init starts the setup screen.
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.err = ""
		switch m.step {
		case stepToken:
			switch msg.String() {
			case "enter":
				token := strings.TrimSpace(m.tokenInput.Value())
				if token == "" {
					m.err = "Token cannot be empty"
					return m, nil
				}
				m.step = stepValidating
				m.client = api.NewClient(token)
				return m, m.validateToken()
			}
		case stepWorkspace:
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.workspaces)-1 {
					m.cursor++
				}
			case "enter":
				if len(m.workspaces) > 0 {
					m.step = stepSaving
					ws := m.workspaces[m.cursor]
					return m, m.saveConfig(m.tokenInput.Value(), ws)
				}
			}
			return m, nil
		}

	case common.UserFetchedMsg:
		return m, m.fetchWorkspaces()

	case common.WorkspacesFetchedMsg:
		m.workspaces = msg.Workspaces
		m.step = stepWorkspace
		if len(m.workspaces) == 1 {
			m.step = stepSaving
			ws := m.workspaces[0]
			return m, m.saveConfig(m.tokenInput.Value(), ws)
		}
		return m, nil

	case common.ConfigSavedMsg:
		return m, func() tea.Msg {
			return common.SwitchScreenMsg{Screen: common.ScreenDashboard}
		}

	case common.ErrMsg:
		m.step = stepToken
		m.err = msg.Err.Error()
		return m, nil
	}

	if m.step == stepToken {
		var cmd tea.Cmd
		m.tokenInput, cmd = m.tokenInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) validateToken() tea.Cmd {
	return func() tea.Msg {
		user, err := m.client.GetMe()
		if err != nil {
			return common.ErrMsg{Err: fmt.Errorf("invalid token: %w", err)}
		}
		return common.UserFetchedMsg{User: user}
	}
}

func (m Model) fetchWorkspaces() tea.Cmd {
	return func() tea.Msg {
		ws, err := m.client.GetWorkspaces()
		if err != nil {
			return common.ErrMsg{Err: fmt.Errorf("fetch workspaces: %w", err)}
		}
		return common.WorkspacesFetchedMsg{Workspaces: ws}
	}
}

func (m Model) saveConfig(token string, ws api.Workspace) tea.Cmd {
	return func() tea.Msg {
		cfg := &config.Config{
			APIToken:      strings.TrimSpace(token),
			WorkspaceID:   ws.ID,
			WorkspaceName: ws.Name,
		}
		if err := config.Save(cfg); err != nil {
			return common.ErrMsg{Err: fmt.Errorf("save config: %w", err)}
		}
		return common.ConfigSavedMsg{}
	}
}

// View renders the setup screen.
func (m Model) View() string {
	var b strings.Builder

	b.WriteString(common.TitleStyle.Render("Toggl TUI Setup"))
	b.WriteString("\n\n")

	switch m.step {
	case stepToken:
		b.WriteString("Enter your Toggl API token:\n")
		b.WriteString(common.MutedStyle.Render("(Find it at https://track.toggl.com/profile)"))
		b.WriteString("\n\n")
		b.WriteString(m.tokenInput.View())
		if m.err != "" {
			b.WriteString("\n\n")
			b.WriteString(common.ErrorStyle.Render(m.err))
		}
		b.WriteString("\n\n")
		b.WriteString(common.HelpStyle.Render("enter: validate • ctrl+c: quit"))

	case stepValidating:
		b.WriteString("Validating token...")

	case stepWorkspace:
		b.WriteString("Select a workspace:\n\n")
		for i, ws := range m.workspaces {
			cursor := "  "
			style := lipgloss.NewStyle()
			if i == m.cursor {
				cursor = "> "
				style = common.SelectedRowStyle
			}
			b.WriteString(style.Render(fmt.Sprintf("%s%s", cursor, ws.Name)))
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(common.HelpStyle.Render("j/k: navigate • enter: select"))

	case stepSaving:
		b.WriteString("Saving configuration...")
	}

	return b.String()
}

// Token returns the current API token value.
func (m Model) Token() string {
	return strings.TrimSpace(m.tokenInput.Value())
}

// WorkspaceID returns the selected workspace ID (0 if none selected).
func (m Model) WorkspaceID() int {
	if m.cursor < len(m.workspaces) {
		return m.workspaces[m.cursor].ID
	}
	return 0
}
