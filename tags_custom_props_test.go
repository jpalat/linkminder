package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test helper functions for JSON conversion
func TestTagsJSONHelpers(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		expected string
	}{
		{
			name:     "empty tags",
			tags:     []string{},
			expected: "[]",
		},
		{
			name:     "nil tags",
			tags:     nil,
			expected: "[]",
		},
		{
			name:     "single tag",
			tags:     []string{"react"},
			expected: `["react"]`,
		},
		{
			name:     "multiple tags",
			tags:     []string{"react", "javascript", "frontend"},
			expected: `["react","javascript","frontend"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tagsToJSON(tt.tags)
			if result != tt.expected {
				t.Errorf("tagsToJSON() = %v, want %v", result, tt.expected)
			}

			// Test round trip conversion
			parsed := tagsFromJSON(result)
			if len(tt.tags) == 0 && len(parsed) != 0 {
				t.Errorf("tagsFromJSON() = %v, want empty", parsed)
			} else if len(tt.tags) > 0 {
				if len(parsed) != len(tt.tags) {
					t.Errorf("tagsFromJSON() length = %v, want %v", len(parsed), len(tt.tags))
				}
				for i, tag := range tt.tags {
					if parsed[i] != tag {
						t.Errorf("tagsFromJSON()[%d] = %v, want %v", i, parsed[i], tag)
					}
				}
			}
		})
	}
}

func TestCustomPropsJSONHelpers(t *testing.T) {
	tests := []struct {
		name     string
		props    map[string]string
		expected string
	}{
		{
			name:     "empty props",
			props:    map[string]string{},
			expected: "{}",
		},
		{
			name:     "nil props",
			props:    nil,
			expected: "{}",
		},
		{
			name:     "single prop",
			props:    map[string]string{"priority": "high"},
			expected: `{"priority":"high"}`,
		},
		{
			name: "multiple props",
			props: map[string]string{
				"priority": "high",
				"category": "work",
				"status":   "pending",
			},
			expected: "", // We'll check this differently since map ordering isn't guaranteed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := customPropsToJSON(tt.props)
			
			if tt.name == "multiple props" {
				// For multiple props, just check that it's valid JSON and round-trips correctly
				var parsed map[string]string
				if err := json.Unmarshal([]byte(result), &parsed); err != nil {
					t.Errorf("customPropsToJSON() produced invalid JSON: %v", err)
				}
			} else if result != tt.expected {
				t.Errorf("customPropsToJSON() = %v, want %v", result, tt.expected)
			}

			// Test round trip conversion
			parsed := customPropsFromJSON(result)
			if len(tt.props) == 0 && len(parsed) != 0 {
				t.Errorf("customPropsFromJSON() = %v, want empty", parsed)
			} else if len(tt.props) > 0 {
				if len(parsed) != len(tt.props) {
					t.Errorf("customPropsFromJSON() length = %v, want %v", len(parsed), len(tt.props))
				}
				for key, value := range tt.props {
					if parsed[key] != value {
						t.Errorf("customPropsFromJSON()[%s] = %v, want %v", key, parsed[key], value)
					}
				}
			}
		})
	}
}

// Test bookmark creation with tags and custom properties
func TestCreateBookmarkWithTagsAndProps(t *testing.T) {
	// Create a test database
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Set up database schema with all migrations
	if err := createTestTablesWithMigrations(testDB); err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}

	// Store original db and replace with test db
	originalDB := db
	db = testDB
	defer func() { db = originalDB }()

	bookmark := BookmarkRequest{
		URL:         "https://react.dev/learn",
		Title:       "Learn React",
		Description: "Official React documentation",
		Action:      "working",
		Topic:       "frontend-development",
		Tags:        []string{"react", "javascript", "frontend"},
		CustomProperties: map[string]string{
			"priority": "high",
			"category": "documentation",
			"status":   "active",
		},
	}

	err = saveBookmarkToDB(bookmark)
	if err != nil {
		t.Fatalf("saveBookmarkToDB failed: %v", err)
	}

	// Verify the bookmark was saved with tags and custom properties
	var id int
	var url, title, tagsJSON, customPropsJSON string
	err = db.QueryRow("SELECT id, url, title, tags, custom_properties FROM bookmarks WHERE url = ?", bookmark.URL).
		Scan(&id, &url, &title, &tagsJSON, &customPropsJSON)
	if err != nil {
		t.Fatalf("Failed to query saved bookmark: %v", err)
	}

	// Verify basic fields
	if url != bookmark.URL {
		t.Errorf("URL = %v, want %v", url, bookmark.URL)
	}
	if title != bookmark.Title {
		t.Errorf("Title = %v, want %v", title, bookmark.Title)
	}

	// Verify tags were saved correctly
	savedTags := tagsFromJSON(tagsJSON)
	if len(savedTags) != len(bookmark.Tags) {
		t.Errorf("Tags length = %v, want %v", len(savedTags), len(bookmark.Tags))
	}
	for i, tag := range bookmark.Tags {
		if savedTags[i] != tag {
			t.Errorf("Tag[%d] = %v, want %v", i, savedTags[i], tag)
		}
	}

	// Verify custom properties were saved correctly
	savedProps := customPropsFromJSON(customPropsJSON)
	if len(savedProps) != len(bookmark.CustomProperties) {
		t.Errorf("Custom properties length = %v, want %v", len(savedProps), len(bookmark.CustomProperties))
	}
	for key, value := range bookmark.CustomProperties {
		if savedProps[key] != value {
			t.Errorf("CustomProperty[%s] = %v, want %v", key, savedProps[key], value)
		}
	}
}

// Test bookmark update with tags and custom properties
func TestUpdateBookmarkWithTagsAndProps(t *testing.T) {
	// Create a test database
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Set up database schema
	if err := createTestTablesWithMigrations(testDB); err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}

	// Store original db and replace with test db
	originalDB := db
	db = testDB
	defer func() { db = originalDB }()

	// Create initial bookmark
	initial := BookmarkRequest{
		URL:   "https://example.com",
		Title: "Example",
		Tags:  []string{"old-tag"},
		CustomProperties: map[string]string{
			"status": "draft",
		},
	}

	err = saveBookmarkToDB(initial)
	if err != nil {
		t.Fatalf("saveBookmarkToDB failed: %v", err)
	}

	// Get the ID of the created bookmark
	var bookmarkID int
	err = db.QueryRow("SELECT id FROM bookmarks WHERE url = ?", initial.URL).Scan(&bookmarkID)
	if err != nil {
		t.Fatalf("Failed to get bookmark ID: %v", err)
	}

	// Update bookmark with new tags and properties
	updateReq := BookmarkUpdateRequest{
		Action: "working",
		Topic:  "development",
		Tags:   []string{"react", "javascript", "updated"},
		CustomProperties: map[string]string{
			"priority": "high",
			"status":   "active",
			"category": "frontend",
		},
	}

	err = updateBookmarkInDB(bookmarkID, updateReq)
	if err != nil {
		t.Fatalf("updateBookmarkInDB failed: %v", err)
	}

	// Verify the update
	updatedBookmark, err := getBookmarkByID(bookmarkID)
	if err != nil {
		t.Fatalf("getBookmarkByID failed: %v", err)
	}

	// Verify tags were updated
	if len(updatedBookmark.Tags) != len(updateReq.Tags) {
		t.Errorf("Updated tags length = %v, want %v", len(updatedBookmark.Tags), len(updateReq.Tags))
	}
	for i, tag := range updateReq.Tags {
		if updatedBookmark.Tags[i] != tag {
			t.Errorf("Updated tag[%d] = %v, want %v", i, updatedBookmark.Tags[i], tag)
		}
	}

	// Verify custom properties were updated
	if len(updatedBookmark.CustomProperties) != len(updateReq.CustomProperties) {
		t.Errorf("Updated custom properties length = %v, want %v", len(updatedBookmark.CustomProperties), len(updateReq.CustomProperties))
	}
	for key, value := range updateReq.CustomProperties {
		if updatedBookmark.CustomProperties[key] != value {
			t.Errorf("Updated custom property[%s] = %v, want %v", key, updatedBookmark.CustomProperties[key], value)
		}
	}
}

// Test API endpoint with tags and custom properties
func TestBookmarkAPIWithTagsAndProps(t *testing.T) {
	// Create a test database
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Set up database schema
	if err := createTestTablesWithMigrations(testDB); err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}

	// Store original db and replace with test db
	originalDB := db
	db = testDB
	defer func() { db = originalDB }()

	// Test POST /bookmark with tags and custom properties
	bookmark := BookmarkRequest{
		URL:         "https://vuejs.org/guide",
		Title:       "Vue.js Guide",
		Description: "Official Vue.js documentation",
		Action:      "working",
		Topic:       "frontend",
		Tags:        []string{"vue", "javascript", "spa"},
		CustomProperties: map[string]string{
			"framework": "vue",
			"difficulty": "beginner",
			"completed":  "false",
		},
	}

	reqBody, _ := json.Marshal(bookmark)
	req := httptest.NewRequest("POST", "/bookmark", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.0.2.1:1234"

	w := httptest.NewRecorder()
	handleBookmark(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify tags are in response
	if tags, ok := response["tags"].([]interface{}); ok {
		if len(tags) != len(bookmark.Tags) {
			t.Errorf("Response tags length = %v, want %v", len(tags), len(bookmark.Tags))
		}
	} else {
		t.Error("Response missing tags field")
	}

	// Verify custom properties are in response
	if customProps, ok := response["customProperties"].(map[string]interface{}); ok {
		if len(customProps) != len(bookmark.CustomProperties) {
			t.Errorf("Response custom properties length = %v, want %v", len(customProps), len(bookmark.CustomProperties))
		}
	} else {
		t.Error("Response missing customProperties field")
	}
}

// Helper function to create test tables with all migrations applied
func createTestTablesWithMigrations(testDB *sql.DB) error {
	// Apply all migrations in order
	migrations := []string{
		// Migration 1: Initial schema
		`CREATE TABLE bookmarks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			url TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			content TEXT,
			action TEXT,
			shareTo TEXT,
			topic TEXT
		)`,
		// Migration 2: Projects table
		`CREATE TABLE projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			status TEXT DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Migration 3: Add project_id column
		`ALTER TABLE bookmarks ADD COLUMN project_id INTEGER REFERENCES projects(id)`,
		// Migration 5: Add tags and custom_properties columns
		`ALTER TABLE bookmarks ADD COLUMN tags TEXT DEFAULT '[]'`,
		`ALTER TABLE bookmarks ADD COLUMN custom_properties TEXT DEFAULT '{}'`,
		// Migration 6: Add deleted column for soft delete
		`ALTER TABLE bookmarks ADD COLUMN deleted BOOLEAN DEFAULT FALSE`,
	}

	for i, migration := range migrations {
		if _, err := testDB.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %v", i+1, err)
		}
	}

	return nil
}