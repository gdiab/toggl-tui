package api

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestGetWeekEntries(t *testing.T) {
	pid := 5
	entries := []TimeEntry{
		{ID: 1, Description: "standup", Start: "2026-03-16T09:00:00Z", Duration: 900, WorkspaceID: 100},
		{ID: 2, Description: "coding", Start: "2026-03-16T10:00:00Z", Duration: 3600, WorkspaceID: 100, ProjectID: &pid},
		{ID: 3, Description: "review", Start: "2026-03-17T14:00:00Z", Duration: 1800, WorkspaceID: 100},
	}

	c, srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/time_entries" {
			t.Errorf("path = %q, want /me/time_entries", r.URL.Path)
		}
		start := r.URL.Query().Get("start_date")
		end := r.URL.Query().Get("end_date")
		if start == "" || end == "" {
			t.Errorf("missing date params: start=%q end=%q", start, end)
		}
		json.NewEncoder(w).Encode(entries)
	})
	defer srv.Close()

	// Wednesday 2026-03-18, so Monday is 2026-03-16
	now := time.Date(2026, 3, 18, 15, 0, 0, 0, time.UTC)
	got, err := c.getWeekEntries(now)
	if err != nil {
		t.Fatalf("getWeekEntries: %v", err)
	}

	// Should have 3 days: Mon, Tue, Wed
	if len(got.Days) != 3 {
		t.Fatalf("len(Days) = %d, want 3", len(got.Days))
	}

	// Monday: 2 entries, 900+3600 = 4500s
	if len(got.Days[0].Entries) != 2 {
		t.Errorf("Monday entries = %d, want 2", len(got.Days[0].Entries))
	}
	if got.Days[0].Total != 4500*time.Second {
		t.Errorf("Monday total = %v, want %v", got.Days[0].Total, 4500*time.Second)
	}

	// Tuesday: 1 entry, 1800s
	if len(got.Days[1].Entries) != 1 {
		t.Errorf("Tuesday entries = %d, want 1", len(got.Days[1].Entries))
	}
	if got.Days[1].Total != 1800*time.Second {
		t.Errorf("Tuesday total = %v, want %v", got.Days[1].Total, 1800*time.Second)
	}

	// Wednesday: 0 entries
	if len(got.Days[2].Entries) != 0 {
		t.Errorf("Wednesday entries = %d, want 0", len(got.Days[2].Entries))
	}
	if got.Days[2].Total != 0 {
		t.Errorf("Wednesday total = %v, want 0", got.Days[2].Total)
	}

	// Week total: 4500 + 1800 = 6300s
	if got.Total != 6300*time.Second {
		t.Errorf("week total = %v, want %v", got.Total, 6300*time.Second)
	}
}

func TestGetWeekEntriesMonday(t *testing.T) {
	// When "today" is Monday, should return exactly 1 day.
	c, srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]TimeEntry{})
	})
	defer srv.Close()

	now := time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC) // Monday
	got, err := c.getWeekEntries(now)
	if err != nil {
		t.Fatalf("getWeekEntries: %v", err)
	}
	if len(got.Days) != 1 {
		t.Fatalf("len(Days) = %d, want 1", len(got.Days))
	}
	if got.Days[0].Date.Weekday() != time.Monday {
		t.Errorf("day = %v, want Monday", got.Days[0].Date.Weekday())
	}
}

func TestGetWeekEntriesSunday(t *testing.T) {
	// Sunday belongs to the previous Monday's week — should return 7 days.
	c, srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]TimeEntry{})
	})
	defer srv.Close()

	now := time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC) // Sunday
	got, err := c.getWeekEntries(now)
	if err != nil {
		t.Fatalf("getWeekEntries: %v", err)
	}
	if len(got.Days) != 7 {
		t.Fatalf("len(Days) = %d, want 7", len(got.Days))
	}
	if got.Days[0].Date.Weekday() != time.Monday {
		t.Errorf("first day = %v, want Monday", got.Days[0].Date.Weekday())
	}
	if got.Days[6].Date.Weekday() != time.Sunday {
		t.Errorf("last day = %v, want Sunday", got.Days[6].Date.Weekday())
	}
}

func TestGetWeekEntriesRunningTimer(t *testing.T) {
	// A running entry (Duration == -1) should use elapsed time from start to now.
	entries := []TimeEntry{
		{ID: 1, Description: "in progress", Start: "2026-03-18T14:00:00Z", Duration: -1, WorkspaceID: 100},
	}

	c, srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(entries)
	})
	defer srv.Close()

	// "now" is 1 hour after the entry started.
	now := time.Date(2026, 3, 18, 15, 0, 0, 0, time.UTC)
	got, err := c.getWeekEntries(now)
	if err != nil {
		t.Fatalf("getWeekEntries: %v", err)
	}

	// Wednesday (index 2): running for 1 hour = 3600s
	if len(got.Days) != 3 {
		t.Fatalf("len(Days) = %d, want 3", len(got.Days))
	}
	if got.Days[2].Total != 3600*time.Second {
		t.Errorf("running timer total = %v, want %v", got.Days[2].Total, 3600*time.Second)
	}
}

func TestGetWeekEntriesAPIError(t *testing.T) {
	c, srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("internal error"))
	})
	defer srv.Close()

	now := time.Date(2026, 3, 18, 15, 0, 0, 0, time.UTC)
	_, err := c.getWeekEntries(now)
	if err == nil {
		t.Fatal("expected error for API failure")
	}
}

func TestWeekMonday(t *testing.T) {
	tests := []struct {
		name string
		in   time.Time
		want time.Time
	}{
		{"monday", time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC), time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)},
		{"wednesday", time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC), time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)},
		{"sunday", time.Date(2026, 3, 22, 12, 0, 0, 0, time.UTC), time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)},
		{"saturday", time.Date(2026, 3, 21, 12, 0, 0, 0, time.UTC), time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := weekMonday(tt.in)
			if !got.Equal(tt.want) {
				t.Errorf("weekMonday(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
