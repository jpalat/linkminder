package models

// TriageResponse represents the response for triage queue requests
type TriageResponse struct {
	Bookmarks []TriageBookmark `json:"bookmarks"`
	Total     int              `json:"total"`
	Limit     int              `json:"limit"`
	Offset    int              `json:"offset"`
}

// ProjectsResponse represents the response for projects listing
type ProjectsResponse struct {
	ActiveProjects       []ActiveProject       `json:"activeProjects"`
	ReferenceCollections []ReferenceCollection `json:"referenceCollections"`
}

// ProjectDetailResponse represents the detailed view of a project
type ProjectDetailResponse struct {
	Topic       string            `json:"topic"`
	LinkCount   int               `json:"linkCount"`
	LastUpdated string            `json:"lastUpdated"`
	Status      string            `json:"status"`
	Bookmarks   []ProjectBookmark `json:"bookmarks"`
}

// SummaryStats represents dashboard summary statistics
type SummaryStats struct {
	NeedsTriage     int           `json:"needsTriage"`
	ActiveProjects  int           `json:"activeProjects"`
	ReadyToShare    int           `json:"readyToShare"`
	Archived        int           `json:"archived"`
	TotalBookmarks  int           `json:"totalBookmarks"`
	ProjectStats    []ProjectStat `json:"projectStats"`
}