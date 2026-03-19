package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		u, p, ok := r.BasicAuth()
		if !ok || u != "test-token" || p != "api_token" {
			t.Errorf("bad auth: %s:%s ok=%v", u, p, ok)
		}
		json.NewEncoder(w).Encode(want)
	})
	defer srv.Close()

	got, err := c.GetMe()
	if err != nil {
		t.Fatalf("GetMe: %v", err)
	}
	if got.ID != want.ID || got.Email != want.Email {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestGetWorkspaces(t *testing.T) {
	want := []Workspace{{ID: 1, Name: "Work"}, {ID: 2, Name: "Personal"}}
	c, srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(want)
	})
	defer srv.Close()

	got, err := c.GetWorkspaces()
	if err != nil {
		t.Fatalf("GetWorkspaces: %v", err)
	}
	if len(got) != 2 || got[0].Name != "Work" {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestGetCurrentTimerNull(t *testing.T) {
	c, srv := testServer(func(w http.ResponseWriter, r *http.Request) {
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
	want := TimeEntry{ID: 42, Description: "coding", Duration: -1}
	c, srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(want)
	})
	defer srv.Close()

	got, err := c.GetCurrentTimer()
	if err != nil {
		t.Fatalf("GetCurrentTimer: %v", err)
	}
	if got == nil || got.ID != 42 {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestCreateTimeEntry(t *testing.T) {
	want := TimeEntry{ID: 789, Description: "test"}
	c, srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var req CreateTimeEntryRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.CreatedWith != "toggl-tui" {
			t.Errorf("expected created_with toggl-tui, got %s", req.CreatedWith)
		}
		json.NewEncoder(w).Encode(want)
	})
	defer srv.Close()

	got, err := c.CreateTimeEntry(100, CreateTimeEntryRequest{Description: "test", Duration: -1})
	if err != nil {
		t.Fatalf("CreateTimeEntry: %v", err)
	}
	if got.ID != 789 {
		t.Errorf("got ID %d, want 789", got.ID)
	}
}

func TestStopTimer(t *testing.T) {
	stop := "2024-01-01T12:00:00Z"
	want := TimeEntry{ID: 42, Duration: 3600, Stop: &stop}
	c, srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(want)
	})
	defer srv.Close()

	got, err := c.StopTimer(100, 42)
	if err != nil {
		t.Fatalf("StopTimer: %v", err)
	}
	if got.Duration != 3600 {
		t.Errorf("got duration %d, want 3600", got.Duration)
	}
}

func TestRateLimitError(t *testing.T) {
	c, srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		w.Write([]byte("too many requests"))
	})
	defer srv.Close()

	_, err := c.GetMe()
	if err == nil {
		t.Fatal("expected error for 429")
	}
}

func TestAPIError(t *testing.T) {
	c, srv := testServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		w.Write([]byte("forbidden"))
	})
	defer srv.Close()

	_, err := c.GetMe()
	if err == nil {
		t.Fatal("expected error for 403")
	}
}
