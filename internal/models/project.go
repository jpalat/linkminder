package models

// Project represents a project entity in the database
type Project struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status"`
	LinkCount   int    `json:"linkCount"`
	LastUpdated string `json:"lastUpdated"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

// ActiveProject represents a project with active status and link count
type ActiveProject struct {
	ID          int    `json:"id"`
	Topic       string `json:"topic"`
	LinkCount   int    `json:"linkCount"`
	LastUpdated string `json:"lastUpdated"`
	Status      string `json:"status"`
}

// ReferenceCollection represents a collection of reference links
type ReferenceCollection struct {
	Topic        string `json:"topic"`
	LinkCount    int    `json:"linkCount"`
	LastAccessed string `json:"lastAccessed"`
}

// ProjectStat represents project statistics including latest resource info
type ProjectStat struct {
	Topic       string `json:"topic"`
	Count       int    `json:"count"`
	LastUpdated string `json:"lastUpdated"`
	Status      string `json:"status"`
	LatestURL   string `json:"latestURL"`
	LatestTitle string `json:"latestTitle"`
}