package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const defaultBaseURL = "https://api.track.toggl.com/api/v9"

// Client is a Toggl API client with rate limiting.
type Client struct {
	token      string
	baseURL    string
	httpClient *http.Client
	mu         sync.Mutex
	lastReq    time.Time
}

// NewClient creates a new API client with the given API token.
func NewClient(token string) *Client {
	return &Client{
		token:      token,
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) rateLimit() {
	c.mu.Lock()
	defer c.mu.Unlock()
	since := time.Since(c.lastReq)
	if since < time.Second {
		time.Sleep(time.Second - since)
	}
	c.lastReq = time.Now()
}

func (c *Client) do(method, path string, body string) ([]byte, error) {
	c.rateLimit()

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.SetBasicAuth(c.token, "api_token")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == 402 {
		return nil, fmt.Errorf("payment required — check your Toggl subscription")
	}
	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("rate limited — please wait a moment")
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(data))
	}

	return data, nil
}

// GetMe returns the authenticated user.
func (c *Client) GetMe() (User, error) {
	data, err := c.do("GET", "/me", "")
	if err != nil {
		return User{}, err
	}
	var u User
	return u, json.Unmarshal(data, &u)
}

// GetWorkspaces returns the user's workspaces.
func (c *Client) GetWorkspaces() ([]Workspace, error) {
	data, err := c.do("GET", "/workspaces", "")
	if err != nil {
		return nil, err
	}
	var ws []Workspace
	return ws, json.Unmarshal(data, &ws)
}

// GetProjects returns projects for a workspace.
func (c *Client) GetProjects(workspaceID int) ([]Project, error) {
	path := fmt.Sprintf("/workspaces/%d/projects", workspaceID)
	data, err := c.do("GET", path, "")
	if err != nil {
		return nil, err
	}
	var ps []Project
	return ps, json.Unmarshal(data, &ps)
}

// GetCurrentTimer returns the running time entry, or nil if none.
func (c *Client) GetCurrentTimer() (*TimeEntry, error) {
	data, err := c.do("GET", "/me/time_entries/current", "")
	if err != nil {
		return nil, err
	}
	if string(data) == "null" {
		return nil, nil
	}
	var te TimeEntry
	return &te, json.Unmarshal(data, &te)
}

// GetTimeEntries returns time entries in the given date range (RFC 3339).
func (c *Client) GetTimeEntries(startDate, endDate string) ([]TimeEntry, error) {
	path := fmt.Sprintf("/me/time_entries?start_date=%s&end_date=%s", startDate, endDate)
	data, err := c.do("GET", path, "")
	if err != nil {
		return nil, err
	}
	var entries []TimeEntry
	return entries, json.Unmarshal(data, &entries)
}

// CreateTimeEntry creates a new time entry.
func (c *Client) CreateTimeEntry(workspaceID int, req CreateTimeEntryRequest) (TimeEntry, error) {
	req.WorkspaceID = workspaceID
	req.CreatedWith = "toggl-tui"

	body, err := json.Marshal(req)
	if err != nil {
		return TimeEntry{}, fmt.Errorf("marshal request: %w", err)
	}

	path := fmt.Sprintf("/workspaces/%d/time_entries", workspaceID)
	data, err := c.do("POST", path, string(body))
	if err != nil {
		return TimeEntry{}, err
	}
	var te TimeEntry
	return te, json.Unmarshal(data, &te)
}

// UpdateTimeEntry updates a time entry (e.g. description).
func (c *Client) UpdateTimeEntry(workspaceID, entryID int, req UpdateTimeEntryRequest) (TimeEntry, error) {
	req.WorkspaceID = workspaceID

	body, err := json.Marshal(req)
	if err != nil {
		return TimeEntry{}, fmt.Errorf("marshal request: %w", err)
	}

	path := fmt.Sprintf("/workspaces/%d/time_entries/%d", workspaceID, entryID)
	data, err := c.do("PUT", path, string(body))
	if err != nil {
		return TimeEntry{}, err
	}
	var te TimeEntry
	return te, json.Unmarshal(data, &te)
}

// StopTimer stops the running time entry.
func (c *Client) StopTimer(workspaceID, entryID int) (TimeEntry, error) {
	path := fmt.Sprintf("/workspaces/%d/time_entries/%d/stop", workspaceID, entryID)
	data, err := c.do("PATCH", path, "")
	if err != nil {
		return TimeEntry{}, err
	}
	var te TimeEntry
	return te, json.Unmarshal(data, &te)
}
