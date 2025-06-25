package models

// BookmarkRequest represents a request to create a new bookmark
type BookmarkRequest struct {
	URL              string            `json:"url"`
	Title            string            `json:"title"`
	Description      string            `json:"description,omitempty"`
	Content          string            `json:"content,omitempty"`
	Action           string            `json:"action,omitempty"`
	ShareTo          string            `json:"shareTo,omitempty"`
	Topic            string            `json:"topic,omitempty"`     // Legacy support
	ProjectID        int               `json:"projectId,omitempty"` // New field
	Tags             []string          `json:"tags,omitempty"`
	CustomProperties map[string]string `json:"customProperties,omitempty"`
}

// BookmarkUpdateRequest represents a request to partially update a bookmark
type BookmarkUpdateRequest struct {
	Action           string            `json:"action,omitempty"`
	ShareTo          string            `json:"shareTo,omitempty"`
	Topic            string            `json:"topic,omitempty"`     // Legacy support
	ProjectID        int               `json:"projectId,omitempty"` // New field
	Tags             []string          `json:"tags,omitempty"`
	CustomProperties map[string]string `json:"customProperties,omitempty"`
}

// BookmarkFullUpdateRequest represents a request to fully update a bookmark
type BookmarkFullUpdateRequest struct {
	Title            string            `json:"title"`
	URL              string            `json:"url"`
	Description      string            `json:"description,omitempty"`
	Action           string            `json:"action,omitempty"`
	ShareTo          string            `json:"shareTo,omitempty"`
	Topic            string            `json:"topic,omitempty"`
	Tags             []string          `json:"tags,omitempty"`
	CustomProperties map[string]string `json:"customProperties,omitempty"`
}

// ProjectCreateRequest represents a request to create a new project
type ProjectCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
}

// ProjectUpdateRequest represents a request to update a project
type ProjectUpdateRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
}