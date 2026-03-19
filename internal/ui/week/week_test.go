package week

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gdiab/toggl-tui/internal/api"
	"github.com/gdiab/toggl-tui/internal/ui/common"
)

func makeTestSummary() api.WeekSummary {
	monday := time.Date(2026, 3, 16, 0, 0, 0, 0, time.Local)
	return api.WeekSummary{
		Days: []api.DaySummary{
			{Date: monday, Total: 2*time.Hour + 30*time.Minute},
			{Date: monday.AddDate(0, 0, 1), Total: 1 * time.Hour},
			{Date: monday.AddDate(0, 0, 2), Total: 3*time.Hour + 15*time.Minute},
		},
		Total: 6*time.Hour + 45*time.Minute,
	}
}

func TestNew(t *testing.T) {
	m := New(nil)
	if m.loaded {
		t.Error("new model should not be loaded")
	}
	if m.cursor != 0 {
		t.Error("cursor should start at 0")
	}
}

func TestUpdateWeekFetched(t *testing.T) {
	m := New(nil)
	summary := makeTestSummary()
	msg := common.WeekFetchedMsg{Summary: summary}

	m, _ = m.Update(msg)

	if !m.loaded {
		t.Error("model should be loaded after WeekFetchedMsg")
	}
	if len(m.summary.Days) != 3 {
		t.Errorf("expected 3 days, got %d", len(m.summary.Days))
	}
	// Cursor should be on last day (today's position).
	if m.cursor != 2 {
		t.Errorf("cursor should be 2 (last day), got %d", m.cursor)
	}
}

func TestNavigationJK(t *testing.T) {
	m := New(nil)
	m, _ = m.Update(common.WeekFetchedMsg{Summary: makeTestSummary()})

	// Cursor starts at 2 (last day). Move up.
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

	// Move down.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after j, got %d", m.cursor)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 2 {
		t.Errorf("expected cursor 2 after j, got %d", m.cursor)
	}

	// Can't go below len-1.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 2 {
		t.Errorf("expected cursor 2 (clamped), got %d", m.cursor)
	}
}

func TestEscReturns(t *testing.T) {
	m := New(nil)
	m, _ = m.Update(common.WeekFetchedMsg{Summary: makeTestSummary()})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc should produce a command")
	}
	msg := cmd()
	sw, ok := msg.(common.SwitchScreenMsg)
	if !ok {
		t.Fatalf("expected SwitchScreenMsg, got %T", msg)
	}
	if sw.Screen != common.ScreenDashboard {
		t.Error("esc should switch to dashboard")
	}
}

func TestBKeyReturns(t *testing.T) {
	m := New(nil)
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
		t.Error("b should switch to dashboard")
	}
}

func TestViewLoading(t *testing.T) {
	m := New(nil)
	view := m.View()
	if !strings.Contains(view, "Loading") {
		t.Error("should show loading state before data arrives")
	}
}

func TestViewRendersRows(t *testing.T) {
	m := New(nil)
	m, _ = m.Update(common.WeekFetchedMsg{Summary: makeTestSummary()})

	view := m.View()
	if !strings.Contains(view, "Monday") {
		t.Error("should show Monday")
	}
	if !strings.Contains(view, "Tuesday") {
		t.Error("should show Tuesday")
	}
	if !strings.Contains(view, "Wednesday") {
		t.Error("should show Wednesday")
	}
	if !strings.Contains(view, "Total") {
		t.Error("should show total row")
	}
	if !strings.Contains(view, "6h 45m") {
		t.Error("should show week total of 6h 45m")
	}
}

func TestFormatHours(t *testing.T) {
	tests := []struct {
		dur  time.Duration
		want string
	}{
		{0, "0h 00m"},
		{30 * time.Minute, "0h 30m"},
		{1 * time.Hour, "1h 00m"},
		{2*time.Hour + 15*time.Minute, "2h 15m"},
		{10*time.Hour + 5*time.Minute, "10h 05m"},
	}

	for _, tt := range tests {
		got := formatHours(tt.dur)
		if got != tt.want {
			t.Errorf("formatHours(%v) = %q, want %q", tt.dur, got, tt.want)
		}
	}
}
