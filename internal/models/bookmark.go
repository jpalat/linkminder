package models

// ProjectBookmark represents a bookmark within a project context
type ProjectBookmark struct {
	ID               int               `json:"id"`
	URL              string            `json:"url"`
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	Content          string            `json:"content"`
	Timestamp        string            `json:"timestamp"`
	Domain           string            `json:"domain"`
	Age              string            `json:"age"`
	Action           string            `json:"action"`
	Topic            string            `json:"topic"`
	ShareTo          string            `json:"shareTo"`
	Tags             []string          `json:"tags,omitempty"`
	CustomProperties map[string]string `json:"customProperties,omitempty"`
}

// TriageBookmark represents a bookmark in the triage queue
type TriageBookmark struct {
	ID               int               `json:"id"`
	URL              string            `json:"url"`
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	Timestamp        string            `json:"timestamp"`
	Domain           string            `json:"domain"`
	Age              string            `json:"age"`
	Suggested        string            `json:"suggested"`
	Topic            string            `json:"topic"`
	Action           string            `json:"action,omitempty"`
	ShareTo          string            `json:"shareTo,omitempty"`
	Tags             []string          `json:"tags,omitempty"`
	CustomProperties map[string]string `json:"customProperties,omitempty"`
}