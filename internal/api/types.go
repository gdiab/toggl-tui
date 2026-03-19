package api

// User represents the authenticated Toggl user.
type User struct {
	ID                 int    `json:"id"`
	Email              string `json:"email"`
	DefaultWorkspaceID int    `json:"default_workspace_id"`
}

// Workspace represents a Toggl workspace.
type Workspace struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Project represents a Toggl project.
type Project struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Color  string `json:"color"`
	Active bool   `json:"active"`
}

// TimeEntry represents a Toggl time entry.
type TimeEntry struct {
	ID          int     `json:"id"`
	WorkspaceID int     `json:"workspace_id"`
	Description string  `json:"description"`
	Start       string  `json:"start"`
	Stop        *string `json:"stop"`
	Duration    int     `json:"duration"`
	ProjectID   *int    `json:"project_id"`
}

// CreateTimeEntryRequest is the payload for creating a time entry.
type CreateTimeEntryRequest struct {
	Description string `json:"description"`
	Start       string `json:"start"`
	Duration    int    `json:"duration"`
	CreatedWith string `json:"created_with"`
	WorkspaceID int    `json:"workspace_id"`
	ProjectID   *int   `json:"project_id,omitempty"`
	Stop        string `json:"stop,omitempty"`
}

// UpdateTimeEntryRequest is the payload for updating a time entry.
type UpdateTimeEntryRequest struct {
	Description string `json:"description"`
	WorkspaceID int    `json:"workspace_id"`
	ProjectID   *int   `json:"project_id"`
}
