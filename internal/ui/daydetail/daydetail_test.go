package daydetail

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gdiab/toggl-tui/internal/api"
	"github.com/gdiab/toggl-tui/internal/ui/common"
)

func intPtr(i int) *int { return &i }

func makeTestDay() api.DaySummary {
	return api.DaySummary{
		Date:  time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC),
		Total: 3*time.Hour + 45*time.Minute,
		Entries: []api.TimeEntry{
			{ID: 1, Description: "Standup", Duration: 1800, ProjectID: intPtr(100)},
			{ID: 2, Description: "Code review", Duration: 5400, ProjectID: intPtr(200)},
			{ID: 3, Description: "Deep work", Duration: 6300, ProjectID: nil},
		},
	}
}

func makeTestProjects() map[int]api.Project {
	return map[int]api.Project{
		100: {ID: 100, Name: "Engineering"},
		200: {ID: 200, Name: "Reviews"},
	}
}

func TestNew(t *testing.T) {
	day := makeTestDay()
	projects := makeTestProjects()
	m := New(day, projects, common.ScreenWeekView)

	if m.cursor != 0 {
		t.Error("cursor should start at 0")
	}
	if m.parentScreen != common.ScreenWeekView {
		t.Error("parentScreen should be ScreenWeekView")
	}
	if len(m.day.Entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(m.day.Entries))
	}
}

func TestInit(t *testing.T) {
	m := New(makeTestDay(), makeTestProjects(), common.ScreenWeekView)
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestNavigationJK(t *testing.T) {
	m := New(makeTestDay(), makeTestProjects(), common.ScreenWeekView)

	// Move down with j.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after j, got %d", m.cursor)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 2 {
		t.Errorf("expected cursor 2 after j, got %d", m.cursor)
	}

	// Can't go past last entry.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 2 {
		t.Errorf("expected cursor 2 (clamped), got %d", m.cursor)
	}

	// Move up with k.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after k, got %d", m.cursor)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 after k, got %d", m.cursor)
	}

	// Can't go above 0.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 (clamped), got %d", m.cursor)
	}
}

func TestNavigationArrowKeys(t *testing.T) {
	m := New(makeTestDay(), makeTestProjects(), common.ScreenWeekView)

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after down, got %d", m.cursor)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 after up, got %d", m.cursor)
	}
}

func TestEscReturnsToParent(t *testing.T) {
	m := New(makeTestDay(), makeTestProjects(), common.ScreenWeekView)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc should produce a command")
	}
	msg := cmd()
	sw, ok := msg.(common.SwitchScreenMsg)
	if !ok {
		t.Fatalf("expected SwitchScreenMsg, got %T", msg)
	}
	if sw.Screen != common.ScreenWeekView {
		t.Errorf("esc should switch to ScreenWeekView, got %d", sw.Screen)
	}
}

func TestBKeyReturnsToParent(t *testing.T) {
	m := New(makeTestDay(), makeTestProjects(), common.ScreenDashboard)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	if cmd == nil {
		t.Fatal("b should produce a command")
	}
	msg := cmd()
	sw, ok := msg.(common.SwitchScreenMsg)
	if !ok {
		t.Fatalf("expected SwitchScreenMsg, got %T", msg)
	}
	if sw.Screen != common.ScreenDashboard {
		t.Errorf("b should switch to ScreenDashboard, got %d", sw.Screen)
	}
}

func TestWindowSizeMsg(t *testing.T) {
	m := New(makeTestDay(), makeTestProjects(), common.ScreenWeekView)

	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if m.width != 120 {
		t.Errorf("expected width 120, got %d", m.width)
	}
	if m.height != 40 {
		t.Errorf("expected height 40, got %d", m.height)
	}
}

func TestViewRendersEntries(t *testing.T) {
	m := New(makeTestDay(), makeTestProjects(), common.ScreenWeekView)
	view := m.View()

	if !strings.Contains(view, "Monday") {
		t.Error("should show day name (Monday)")
	}
	if !strings.Contains(view, "March 16") {
		t.Error("should show date")
	}
	if !strings.Contains(view, "Standup") {
		t.Error("should show first entry description")
	}
	if !strings.Contains(view, "Code review") {
		t.Error("should show second entry description")
	}
	if !strings.Contains(view, "Deep work") {
		t.Error("should show third entry description")
	}
	if !strings.Contains(view, "Engineering") {
		t.Error("should show project name")
	}
	if !strings.Contains(view, "Reviews") {
		t.Error("should show project name")
	}
	if !strings.Contains(view, "Total") {
		t.Error("should show total")
	}
	if !strings.Contains(view, "esc/b back") {
		t.Error("should show navigation hint")
	}
}

func TestViewEmptyState(t *testing.T) {
	day := api.DaySummary{
		Date:    time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC),
		Total:   0,
		Entries: nil,
	}
	m := New(day, nil, common.ScreenWeekView)
	view := m.View()

	if !strings.Contains(view, "No entries") {
		t.Error("empty state should show 'No entries'")
	}
}

func TestViewNoDescription(t *testing.T) {
	day := api.DaySummary{
		Date:  time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC),
		Total: 30 * time.Minute,
		Entries: []api.TimeEntry{
			{ID: 1, Description: "", Duration: 1800},
		},
	}
	m := New(day, nil, common.ScreenWeekView)
	view := m.View()

	if !strings.Contains(view, "(no description)") {
		t.Error("should show '(no description)' for entries without description")
	}
}

func TestViewLongDescriptionTruncated(t *testing.T) {
	day := api.DaySummary{
		Date:  time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC),
		Total: 1 * time.Hour,
		Entries: []api.TimeEntry{
			{ID: 1, Description: "This is a very long description that exceeds the limit", Duration: 3600},
		},
	}
	m := New(day, nil, common.ScreenWeekView)
	view := m.View()

	if !strings.Contains(view, "..") {
		t.Error("long descriptions should be truncated with '..'")
	}
	// The full description should NOT appear.
	if strings.Contains(view, "exceeds the limit") {
		t.Error("long description should be truncated")
	}
}

func TestViewLongProjectTruncated(t *testing.T) {
	day := api.DaySummary{
		Date:  time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC),
		Total: 1 * time.Hour,
		Entries: []api.TimeEntry{
			{ID: 1, Description: "Work", Duration: 3600, ProjectID: intPtr(100)},
		},
	}
	projects := map[int]api.Project{
		100: {ID: 100, Name: "Super Long Project Name Here"},
	}
	m := New(day, projects, common.ScreenWeekView)
	view := m.View()

	if !strings.Contains(view, "..") {
		t.Error("long project names should be truncated with '..'")
	}
	if strings.Contains(view, "Name Here") {
		t.Error("long project name should be truncated")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		seconds int
		want    string
	}{
		{0, "0:00:00"},
		{61, "0:01:01"},
		{3600, "1:00:00"},
		{3661, "1:01:01"},
		{7200 + 900 + 30, "2:15:30"},
		{-1, "running"},
	}

	for _, tt := range tests {
		got := formatDuration(tt.seconds)
		if got != tt.want {
			t.Errorf("formatDuration(%d) = %q, want %q", tt.seconds, got, tt.want)
		}
	}
}

func TestCursorDoesNotMoveOnEmptyEntries(t *testing.T) {
	day := api.DaySummary{
		Date:    time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC),
		Total:   0,
		Entries: nil,
	}
	m := New(day, nil, common.ScreenWeekView)

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 0 {
		t.Errorf("cursor should stay at 0 with no entries, got %d", m.cursor)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.cursor != 0 {
		t.Errorf("cursor should stay at 0 with no entries, got %d", m.cursor)
	}
}
