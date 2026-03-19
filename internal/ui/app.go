package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gdiab/toggl-tui/internal/api"
	"github.com/gdiab/toggl-tui/internal/config"
	"github.com/gdiab/toggl-tui/internal/ui/common"
	"github.com/gdiab/toggl-tui/internal/ui/dashboard"
	"github.com/gdiab/toggl-tui/internal/ui/setup"
	"github.com/gdiab/toggl-tui/internal/ui/timer"
)

// App is the root model that routes between screens.
type App struct {
	screen    common.Screen
	setup     setup.Model
	dashboard dashboard.Model
	startForm timer.StartModel
	manualForm timer.ManualModel
	cfg       *config.Config
	client    *api.Client
	errMsg    string
	width     int
	height    int
}

// NewApp creates a new root app model.
func NewApp(cfg *config.Config) App {
	app := App{}
	if cfg == nil {
		app.screen = common.ScreenSetup
		app.setup = setup.New()
	} else {
		app.cfg = cfg
		app.client = api.NewClient(cfg.APIToken)
		app.screen = common.ScreenDashboard
		app.dashboard = dashboard.New(app.client, cfg.WorkspaceID)
	}
	return app
}

// Init starts the app.
func (a App) Init() tea.Cmd {
	switch a.screen {
	case common.ScreenSetup:
		return a.setup.Init()
	case common.ScreenDashboard:
		return a.dashboard.Init()
	}
	return nil
}

// Update handles messages.
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case "q":
			if a.screen == common.ScreenDashboard {
				return a, tea.Quit
			}
		}

	case common.SwitchScreenMsg:
		return a.switchScreen(msg.Screen)

	case common.ErrMsg:
		a.errMsg = msg.Err.Error()
		return a, a.clearErrAfter(5 * time.Second)

	case common.ClearErrMsg:
		a.errMsg = ""
		return a, nil

	case common.TimerStartedMsg:
		a.screen = common.ScreenDashboard
		var m dashboard.Model
		m, cmd := a.dashboard.Update(msg)
		a.dashboard = m
		return a, cmd

	case common.EntryCreatedMsg:
		a.screen = common.ScreenDashboard
		var m dashboard.Model
		m, cmd := a.dashboard.Update(msg)
		a.dashboard = m
		return a, cmd
	}

	// Route to active screen
	var cmd tea.Cmd
	switch a.screen {
	case common.ScreenSetup:
		a.setup, cmd = a.setup.Update(msg)
		// After setup saves config, we need to init dashboard
		if sw, ok := msg.(common.ConfigSavedMsg); ok {
			_ = sw
			cfg, _ := config.Load()
			if cfg != nil {
				a.cfg = cfg
				a.client = api.NewClient(cfg.APIToken)
			}
		}
	case common.ScreenDashboard:
		a.dashboard, cmd = a.dashboard.Update(msg)
	case common.ScreenStartTimer:
		a.startForm, cmd = a.startForm.Update(msg)
	case common.ScreenManualEntry:
		a.manualForm, cmd = a.manualForm.Update(msg)
	}

	return a, cmd
}

func (a App) switchScreen(screen common.Screen) (App, tea.Cmd) {
	a.screen = screen
	switch screen {
	case common.ScreenDashboard:
		if a.client == nil {
			cfg, err := config.Load()
			if err != nil || cfg == nil {
				a.screen = common.ScreenSetup
				a.setup = setup.New()
				return a, a.setup.Init()
			}
			a.cfg = cfg
			a.client = api.NewClient(cfg.APIToken)
		}
		a.dashboard = dashboard.New(a.client, a.cfg.WorkspaceID)
		return a, a.dashboard.Init()
	case common.ScreenStartTimer:
		projects := projectSlice(a.dashboard.Projects())
		a.startForm = timer.NewStart(a.client, a.cfg.WorkspaceID, projects, a.dashboard.HasRunningTimer())
		return a, a.startForm.Init()
	case common.ScreenManualEntry:
		projects := projectSlice(a.dashboard.Projects())
		a.manualForm = timer.NewManual(a.client, a.cfg.WorkspaceID, projects, a.height)
		return a, a.manualForm.Init()
	case common.ScreenSetup:
		a.setup = setup.New()
		return a, a.setup.Init()
	}
	return a, nil
}

func (a App) clearErrAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return common.ClearErrMsg{}
	})
}

func projectSlice(m map[int]api.Project) []api.Project {
	ps := make([]api.Project, 0, len(m))
	for _, p := range m {
		if p.Active {
			ps = append(ps, p)
		}
	}
	return ps
}

// View renders the current screen.
func (a App) View() string {
	var view string
	switch a.screen {
	case common.ScreenSetup:
		view = a.setup.View()
	case common.ScreenDashboard:
		view = a.dashboard.View()
	case common.ScreenStartTimer:
		view = a.startForm.View()
	case common.ScreenManualEntry:
		view = a.manualForm.View()
	}

	if a.errMsg != "" {
		view += "\n" + common.ErrorStyle.Render(a.errMsg)
	}

	return view
}
