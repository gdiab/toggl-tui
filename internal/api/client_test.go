package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testServer(handler http.HandlerFunc) (*Client, *httptest.Server) {
	srv := httptest.NewServer(handler)
	c := NewClient("test-token")
	c.baseURL = srv.URL
	return c, srv
}

func TestGetMe(t *testing.T) {
	want := User{ID: 123, Email: "test@example.com", DefaultWorkspaceID: 456}
	c, srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me" {
			t.Errorf("path = %q, want /me", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("method = %q, want GET", r.Method)
		}
		u, p, ok := r.BasicAuth()
		if !ok || u != "test-token" || p != "api_token" {
			t.Errorf("auth = %s:%s ok=%v, want test-token:api_token", u, p, ok)
		}
		json.NewEncoder(w).Encode(want)
	})
	defer srv.Close()

	got, err := c.GetMe()
	if err != nil {
		t.Fatalf("GetMe: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("ID = %d, want %d", got.ID, want.ID)
	}
	if got.Email != want.Email {
		t.Errorf("Email = %q, want %q", got.Email, want.Email)
	}
	if got.DefaultWorkspaceID != want.DefaultWorkspaceID {
		t.Errorf("DefaultWorkspaceID = %d, want %d", got.DefaultWorkspaceID, want.DefaultWorkspaceID)
	}
}

func TestGetWorkspaces(t *testing.T) {
	want := []Workspace{{ID: 1, Name: "Work"}, {ID: 2, Name: "Personal"}}
	c, srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces" {
			t.Errorf("path = %q, want /workspaces", r.URL.Path)
		}
		json.NewEncoder(w).Encode(want)
	})
	defer srv.Close()

	got, err := c.GetWorkspaces()
	if err != nil {
		t.Fatalf("GetWorkspaces: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].ID != 1 || got[0].Name != "Work" {
		t.Errorf("got[0] = %+v, want {ID:1 Name:Work}", got[0])
	}
	if got[1].ID != 2 || got[1].Name != "Personal" {
		t.Errorf("got[1] = %+v, want {ID:2 Name:Personal}", got[1])
	}
}

func TestGetProjects(t *testing.T) {
	want := []Project{
		{ID: 10, Name: "Backend", Color: "#FF0000", Active: true},
		{ID: 20, Name: "Frontend", Color: "#00FF00", Active: false},
	}
	c, srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces/100/projects" {
			t.Errorf("path = %q, want /workspaces/100/projects", r.URL.Path)
		}
		json.NewEncoder(w).Encode(want)
	})
	defer srv.Close()

	got, err := c.GetProjects(100)
	if err != nil {
		t.Fatalf("GetProjects: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].ID != 10 || got[0].Name != "Backend" || !got[0].Active {
		t.Errorf("got[0] = %+v, want {ID:10 Name:Backend Active:true}", got[0])
	}
	if got[1].Active {
		t.Errorf("got[1].Active = true, want false")
	}
}

func TestGetTimeEntries(t *testing.T) {
	pid := 10
	want := []TimeEntry{
		{ID: 1, Description: "standup", Duration: 900, WorkspaceID: 100},
		{ID: 2, Description: "coding", Duration: 3600, WorkspaceID: 100, ProjectID: &pid},
	}
	c, srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/me/time_entries") {
			t.Errorf("path = %q, want /me/time_entries?...", r.URL.Path)
		}
		start := r.URL.Query().Get("start_date")
		end := r.URL.Query().Get("end_date")
		if start == "" || end == "" {
			t.Errorf("missing date params: start=%q end=%q", start, end)
		}
		json.NewEncoder(w).Encode(want)
	})
	defer srv.Close()

	got, err := c.GetTimeEntries("2024-01-01T00:00:00Z", "2024-01-01T23:59:59Z")
	if err != nil {
		t.Fatalf("GetTimeEntries: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Description != "standup" || got[0].Duration != 900 {
		t.Errorf("got[0] = %+v", got[0])
	}
	if got[1].ProjectID == nil || *got[1].ProjectID != 10 {
		t.Errorf("got[1].ProjectID = %v, want &10", got[1].ProjectID)
	}
}

func TestGetCurrentTimerNull(t *testing.T) {
	c, srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/time_entries/current" {
			t.Errorf("path = %q, want /me/time_entries/current", r.URL.Path)
		}
		w.Write([]byte("null"))
	})
	defer srv.Close()

	got, err := c.GetCurrentTimer()
	if err != nil {
		t.Fatalf("GetCurrentTimer: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestGetCurrentTimerRunning(t *testing.T) {
	want := TimeEntry{ID: 42, Description: "coding", Duration: -1, WorkspaceID: 100}
	c, srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/time_entries/current" {
			t.Errorf("path = %q, want /me/time_entries/current", r.URL.Path)
		}
		json.NewEncoder(w).Encode(want)
	})
	defer srv.Close()

	got, err := c.GetCurrentTimer()
	if err != nil {
		t.Fatalf("GetCurrentTimer: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil entry")
	}
	if got.ID != 42 {
		t.Errorf("ID = %d, want 42", got.ID)
	}
	if got.Description != "coding" {
		t.Errorf("Description = %q, want 'coding'", got.Description)
	}
	if got.Duration != -1 {
		t.Errorf("Duration = %d, want -1", got.Duration)
	}
}

func TestCreateTimeEntry(t *testing.T) {
	want := TimeEntry{ID: 789, Description: "test", Duration: -1, WorkspaceID: 100}
	c, srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/workspaces/100/time_entries" {
			t.Errorf("path = %q, want /workspaces/100/time_entries", r.URL.Path)
		}
		var req CreateTimeEntryRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.CreatedWith != "toggl-tui" {
			t.Errorf("CreatedWith = %q, want 'toggl-tui'", req.CreatedWith)
		}
		if req.Description != "test" {
			t.Errorf("Description = %q, want 'test'", req.Description)
		}
		if req.Duration != -1 {
			t.Errorf("Duration = %d, want -1", req.Duration)
		}
		if req.WorkspaceID != 100 {
			t.Errorf("WorkspaceID = %d, want 100", req.WorkspaceID)
		}
		json.NewEncoder(w).Encode(want)
	})
	defer srv.Close()

	got, err := c.CreateTimeEntry(100, CreateTimeEntryRequest{Description: "test", Duration: -1})
	if err != nil {
		t.Fatalf("CreateTimeEntry: %v", err)
	}
	if got.ID != 789 {
		t.Errorf("ID = %d, want 789", got.ID)
	}
	if got.Description != "test" {
		t.Errorf("Description = %q, want 'test'", got.Description)
	}
}

func TestStopTimer(t *testing.T) {
	stop := "2024-01-01T12:00:00Z"
	want := TimeEntry{ID: 42, Duration: 3600, Stop: &stop, WorkspaceID: 100}
	c, srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("method = %q, want PATCH", r.Method)
		}
		if r.URL.Path != "/workspaces/100/time_entries/42/stop" {
			t.Errorf("path = %q, want /workspaces/100/time_entries/42/stop", r.URL.Path)
		}
		json.NewEncoder(w).Encode(want)
	})
	defer srv.Close()

	got, err := c.StopTimer(100, 42)
	if err != nil {
		t.Fatalf("StopTimer: %v", err)
	}
	if got.Duration != 3600 {
		t.Errorf("Duration = %d, want 3600", got.Duration)
	}
	if got.Stop == nil || *got.Stop != stop {
		t.Errorf("Stop = %v, want %q", got.Stop, stop)
	}
}

func TestUpdateTimeEntry(t *testing.T) {
	pid := 10
	want := TimeEntry{ID: 42, Description: "updated", ProjectID: &pid}
	c, srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("method = %q, want PUT", r.Method)
		}
		if r.URL.Path != "/workspaces/100/time_entries/42" {
			t.Errorf("path = %q, want /workspaces/100/time_entries/42", r.URL.Path)
		}
		var req UpdateTimeEntryRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Description != "updated" {
			t.Errorf("Description = %q, want 'updated'", req.Description)
		}
		if req.WorkspaceID != 100 {
			t.Errorf("WorkspaceID = %d, want 100", req.WorkspaceID)
		}
		json.NewEncoder(w).Encode(want)
	})
	defer srv.Close()

	got, err := c.UpdateTimeEntry(100, 42, UpdateTimeEntryRequest{Description: "updated", ProjectID: &pid})
	if err != nil {
		t.Fatalf("UpdateTimeEntry: %v", err)
	}
	if got.Description != "updated" {
		t.Errorf("Description = %q, want 'updated'", got.Description)
	}
	if got.ProjectID == nil || *got.ProjectID != 10 {
		t.Errorf("ProjectID = %v, want &10", got.ProjectID)
	}
}

func TestHTTPErrors(t *testing.T) {
	tests := []struct {
		code    int
		wantSub string
	}{
		{400, "API error 400"},
		{401, "API error 401"},
		{403, "API error 403"},
		{404, "API error 404"},
		{429, "rate limited"},
		{500, "API error 500"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("status_%d", tt.code), func(t *testing.T) {
			c, srv := testServer(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.code)
				w.Write([]byte("error body"))
			})
			defer srv.Close()

			_, err := c.GetMe()
			if err == nil {
				t.Fatalf("expected error for status %d", tt.code)
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.wantSub)
			}
		})
	}
}

func TestServerUnreachable(t *testing.T) {
	c := NewClient("test-token")
	c.baseURL = "http://127.0.0.1:1" // nothing listening

	_, err := c.GetMe()
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
	if !strings.Contains(err.Error(), "request failed") {
		t.Errorf("error = %q, want 'request failed' substring", err.Error())
	}
}

func TestMalformedJSON(t *testing.T) {
	c, srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{not json!!!"))
	})
	defer srv.Close()

	_, err := c.GetMe()
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}
