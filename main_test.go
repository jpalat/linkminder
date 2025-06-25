package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Test database setup and teardown
type TestDB struct {
	db     *sql.DB
	dbPath string
}

// setupTestDB creates a temporary SQLite database for testing
func setupTestDB(t *testing.T) *TestDB {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_bookmarks.db")
	
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	
	// Create the projects table
	createProjectsTableSQL := `
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		description TEXT,
		status TEXT DEFAULT 'active',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	
	if _, err = db.Exec(createProjectsTableSQL); err != nil {
		t.Fatalf("Failed to create test projects table: %v", err)
	}
	
	// Create the bookmarks table
	createBookmarksTableSQL := `
	CREATE TABLE IF NOT EXISTS bookmarks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		url TEXT NOT NULL,
		title TEXT NOT NULL,
		description TEXT,
		content TEXT,
		action TEXT,
		shareTo TEXT,
		topic TEXT,
		project_id INTEGER REFERENCES projects(id),
		tags TEXT DEFAULT '[]',
		custom_properties TEXT DEFAULT '{}'
	);`
	
	if _, err = db.Exec(createBookmarksTableSQL); err != nil {
		t.Fatalf("Failed to create test bookmarks table: %v", err)
	}
	
	return &TestDB{db: db, dbPath: dbPath}
}

// Project Settings API Tests

func TestProjectSettings_CreateProject(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.db.Close()
	
	// Set the global db variable for testing
	db = testDB.db
	
	tests := []struct {
		name           string
		projectData    map[string]interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name: "valid new project",
			projectData: map[string]interface{}{
				"name":        "Test Project",
				"description": "A test project for unit testing",
				"status":      "active",
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "project with minimal data",
			projectData: map[string]interface{}{
				"name": "Minimal Project",
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "duplicate project name",
			projectData: map[string]interface{}{
				"name":        "Test Project", // Same as first test
				"description": "This should fail",
			},
			expectedStatus: http.StatusConflict,
			expectError:    true,
		},
		{
			name: "missing required name",
			projectData: map[string]interface{}{
				"description": "Project without name",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "empty name",
			projectData: map[string]interface{}{
				"name":        "",
				"description": "Project with empty name",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.projectData)
			req, err := http.NewRequest("POST", "/api/projects", bytes.NewBuffer(body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handleProjects)
			handler.ServeHTTP(rr, req)
			
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", 
					tt.expectedStatus, rr.Code, rr.Body.String())
			}
			
			if !tt.expectError && rr.Code == http.StatusCreated {
				var response map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				
				// Verify response contains expected fields
				if _, ok := response["id"]; !ok {
					t.Error("Response should contain 'id' field")
				}
				if response["name"] != tt.projectData["name"] {
					t.Errorf("Expected name '%v', got '%v'", tt.projectData["name"], response["name"])
				}
			}
		})
	}
}

func TestProjectSettings_UpdateProject(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.db.Close()
	
	// Set the global db variable for testing
	db = testDB.db
	
	// Create a test project first
	createData := map[string]interface{}{
		"name":        "Original Project",
		"description": "Original description",
		"status":      "active",
	}
	
	body, _ := json.Marshal(createData)
	req, _ := http.NewRequest("POST", "/api/projects", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleProjects)
	handler.ServeHTTP(rr, req)
	
	if rr.Code != http.StatusCreated {
		t.Fatalf("Failed to create test project: %d", rr.Code)
	}
	
	var createdProject map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &createdProject)
	projectID := int(createdProject["id"].(float64))
	
	tests := []struct {
		name           string
		projectID      int
		updateData     map[string]interface{}
		expectedStatus int
		expectError    bool
	}{
		{
			name:      "valid update all fields",
			projectID: projectID,
			updateData: map[string]interface{}{
				"name":        "Updated Project",
				"description": "Updated description",
				"status":      "archived",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:      "update only description",
			projectID: projectID,
			updateData: map[string]interface{}{
				"description": "Only description updated",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:      "update only status",
			projectID: projectID,
			updateData: map[string]interface{}{
				"status": "inactive",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:      "nonexistent project",
			projectID: 99999,
			updateData: map[string]interface{}{
				"name": "This should fail",
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:      "empty name",
			projectID: projectID,
			updateData: map[string]interface{}{
				"name": "",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.updateData)
			url := fmt.Sprintf("/api/projects/%d", tt.projectID)
			req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handleProjects)
			handler.ServeHTTP(rr, req)
			
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", 
					tt.expectedStatus, rr.Code, rr.Body.String())
			}
			
			if !tt.expectError && rr.Code == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				
				// Verify updated fields
				for key, expectedValue := range tt.updateData {
					if response[key] != expectedValue {
						t.Errorf("Expected %s '%v', got '%v'", key, expectedValue, response[key])
					}
				}
			}
		})
	}
}

func TestProjectSettings_DeleteProject(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.db.Close()
	
	// Set the global db variable for testing
	db = testDB.db
	
	// Create test projects
	projects := []map[string]interface{}{
		{"name": "Project to Delete", "description": "Will be deleted"},
		{"name": "Project with Bookmarks", "description": "Has associated bookmarks"},
	}
	
	var projectIDs []int
	for _, project := range projects {
		body, _ := json.Marshal(project)
		req, _ := http.NewRequest("POST", "/api/projects", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(handleProjects)
		handler.ServeHTTP(rr, req)
		
		var createdProject map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &createdProject)
		projectIDs = append(projectIDs, int(createdProject["id"].(float64)))
	}
	
	// Add a bookmark to the second project
	_, err := testDB.db.Exec(`
		INSERT INTO bookmarks (url, title, action, topic, project_id, timestamp)
		VALUES (?, ?, ?, ?, ?, ?)
	`, "https://example.com", "Test Bookmark", "working", "Project with Bookmarks", projectIDs[1], time.Now())
	if err != nil {
		t.Fatalf("Failed to create test bookmark: %v", err)
	}
	
	tests := []struct {
		name           string
		projectID      int
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "delete empty project",
			projectID:      projectIDs[0],
			expectedStatus: http.StatusNoContent,
			expectError:    false,
		},
		{
			name:           "delete project with bookmarks (should cascade)",
			projectID:      projectIDs[1],
			expectedStatus: http.StatusNoContent,
			expectError:    false,
		},
		{
			name:           "delete nonexistent project",
			projectID:      99999,
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/projects/%d", tt.projectID)
			req, err := http.NewRequest("DELETE", url, nil)
			if err != nil {
				t.Fatal(err)
			}
			
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handleProjects)
			handler.ServeHTTP(rr, req)
			
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", 
					tt.expectedStatus, rr.Code, rr.Body.String())
			}
			
			// Verify project was actually deleted
			if !tt.expectError && rr.Code == http.StatusNoContent {
				var count int
				err := testDB.db.QueryRow("SELECT COUNT(*) FROM projects WHERE id = ?", tt.projectID).Scan(&count)
				if err != nil {
					t.Errorf("Failed to check if project was deleted: %v", err)
				}
				if count != 0 {
					t.Error("Project should have been deleted")
				}
			}
		})
	}
}

func TestProjectSettings_GetProject(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.db.Close()
	
	// Set the global db variable for testing
	db = testDB.db
	
	// Create a test project
	createData := map[string]interface{}{
		"name":        "Get Test Project",
		"description": "Project for GET testing",
		"status":      "active",
	}
	
	body, _ := json.Marshal(createData)
	req, _ := http.NewRequest("POST", "/api/projects", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleProjects)
	handler.ServeHTTP(rr, req)
	
	var createdProject map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &createdProject)
	projectID := int(createdProject["id"].(float64))
	
	tests := []struct {
		name           string
		projectID      int
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "get existing project",
			projectID:      projectID,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "get nonexistent project",
			projectID:      99999,
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/projects/%d", tt.projectID)
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				t.Fatal(err)
			}
			
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handleProjects)
			handler.ServeHTTP(rr, req)
			
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", 
					tt.expectedStatus, rr.Code, rr.Body.String())
			}
			
			if !tt.expectError && rr.Code == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				
				// Verify response contains expected fields
				expectedFields := []string{"id", "name", "description", "status", "createdAt", "updatedAt"}
				for _, field := range expectedFields {
					if _, ok := response[field]; !ok {
						t.Errorf("Response should contain '%s' field", field)
					}
				}
				
				if response["name"] != createData["name"] {
					t.Errorf("Expected name '%v', got '%v'", createData["name"], response["name"])
				}
			}
		})
	}
}

func TestProjectSettings_InvalidMethods(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.db.Close()
	
	// Set the global db variable for testing
	db = testDB.db
	
	invalidMethods := []string{"PATCH", "OPTIONS", "HEAD"}
	
	for _, method := range invalidMethods {
		t.Run(fmt.Sprintf("invalid method %s", method), func(t *testing.T) {
			req, err := http.NewRequest(method, "/api/projects", nil)
			if err != nil {
				t.Fatal(err)
			}
			
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handleProjects)
			handler.ServeHTTP(rr, req)
			
			if rr.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status %d for method %s, got %d", 
					http.StatusMethodNotAllowed, method, rr.Code)
			}
		})
	}
}

func TestProjectSettings_MalformedJSON(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.db.Close()
	
	// Set the global db variable for testing
	db = testDB.db
	
	tests := []struct {
		name        string
		method      string
		body        string
		expectedStatus int
	}{
		{
			name:   "invalid JSON in POST",
			method: "POST",
			body:   `{"name": "test", "invalid": }`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid JSON in PUT",
			method: "PUT",
			body:   `{"name": "test", "description":}`,
			expectedStatus: http.StatusBadRequest,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/projects"
			if tt.method == "PUT" {
				url = "/api/projects/1"
			}
			
			req, err := http.NewRequest(tt.method, url, strings.NewReader(tt.body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handleProjects)
			handler.ServeHTTP(rr, req)
			
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

// cleanup closes the test database and removes the file
func (tdb *TestDB) cleanup(t *testing.T) {
	if err := tdb.db.Close(); err != nil {
		t.Errorf("Failed to close test database: %v", err)
	}
}

// insertTestBookmarks adds sample data to the test database
func (tdb *TestDB) insertTestBookmarks(t *testing.T) {
	// First, create projects for topics that will be used
	createProjectSQL := `
	INSERT OR IGNORE INTO projects (name, description, status, created_at, updated_at)
	VALUES (?, ?, 'active', '2023-12-01 10:00:00', '2023-12-01 10:00:00')`
	
	projects := []struct {
		name, description string
	}{
		{"Programming", "Programming related bookmarks"},
		{"Development", "Development related bookmarks"},
	}
	
	for _, project := range projects {
		_, err := tdb.db.Exec(createProjectSQL, project.name, project.description)
		if err != nil {
			t.Fatalf("Failed to insert test project: %v", err)
		}
	}
	
	insertSQL := `
	INSERT INTO bookmarks (url, title, description, content, action, shareTo, topic, timestamp)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	
	testData := []BookmarkRequest{
		{URL: "https://example.com/1", Title: "Example 1", Description: "Test bookmark 1", Action: "read-later"},
		{URL: "https://example.com/2", Title: "Example 2", Description: "Test bookmark 2", Action: "working", Topic: "Programming"},
		{URL: "https://example.com/3", Title: "Example 3", Description: "Test bookmark 3", Action: "share", ShareTo: "team"},
		{URL: "https://example.com/4", Title: "Example 4", Description: "Test bookmark 4", Action: "working", Topic: "Development"},
		{URL: "https://example.com/5", Title: "Example 5", Description: "Test bookmark 5", Action: "working", Topic: "Programming"},
	}
	
	for _, bookmark := range testData {
		_, err := tdb.db.Exec(insertSQL, bookmark.URL, bookmark.Title, bookmark.Description,
			bookmark.Content, bookmark.Action, bookmark.ShareTo, bookmark.Topic, "2023-12-01 10:00:00")
		if err != nil {
			t.Fatalf("Failed to insert test bookmark: %v", err)
		}
	}
}

// createTestProject creates a project in the test database
func (tdb *TestDB) createTestProject(t *testing.T, name, description, status string) {
	createProjectSQL := `
	INSERT OR IGNORE INTO projects (name, description, status, created_at, updated_at)
	VALUES (?, ?, ?, '2023-12-01 10:00:00', '2023-12-01 10:00:00')`
	
	_, err := tdb.db.Exec(createProjectSQL, name, description, status)
	if err != nil {
		t.Fatalf("Failed to create test project %s: %v", name, err)
	}
}

// withTestDB is a test helper that sets up a test database, runs the test function, and cleans up
func withTestDB(t *testing.T, testFunc func(*testing.T, *TestDB)) {
	tdb := setupTestDB(t)
	defer tdb.cleanup(t)
	
	// Set global db for handlers to use
	originalDB := db
	db = tdb.db
	defer func() { db = originalDB }()
	
	testFunc(t, tdb)
}

// createDashboardFile creates a temporary dashboard.html file for testing
func createDashboardFile(t *testing.T) string {
	tmpDir := t.TempDir()
	dashboardPath := filepath.Join(tmpDir, "dashboard.html")
	
	dashboardContent := `<!DOCTYPE html>
<html><head><title>Test Dashboard</title></head>
<body><h1>Test Dashboard</h1></body></html>`
	
	err := os.WriteFile(dashboardPath, []byte(dashboardContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test dashboard file: %v", err)
	}
	
	return dashboardPath
}

// Unit Tests for Database Functions

func TestSaveBookmarkToDB(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		req := BookmarkRequest{
			URL:         "https://example.com",
			Title:       "Test Title",
			Description: "Test Description",
			Content:     "Test Content",
			Action:      "read-later",
			ShareTo:     "",
			Topic:       "",
		}
		
		err := saveBookmarkToDB(req)
		if err != nil {
			t.Fatalf("saveBookmarkToDB failed: %v", err)
		}
		
		// Verify the bookmark was saved
		var count int
		err = tdb.db.QueryRow("SELECT COUNT(*) FROM bookmarks WHERE url = ?", req.URL).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query saved bookmark: %v", err)
		}
		
		if count != 1 {
			t.Errorf("Expected 1 bookmark, got %d", count)
		}
		
		// Verify the data
		var savedBookmark BookmarkRequest
		err = tdb.db.QueryRow(
			"SELECT url, title, description, content, action, shareTo, topic FROM bookmarks WHERE url = ?",
			req.URL).Scan(&savedBookmark.URL, &savedBookmark.Title, &savedBookmark.Description,
			&savedBookmark.Content, &savedBookmark.Action, &savedBookmark.ShareTo, &savedBookmark.Topic)
		if err != nil {
			t.Fatalf("Failed to retrieve saved bookmark: %v", err)
		}
		
		if savedBookmark.URL != req.URL {
			t.Errorf("URL: expected %s, got %s", req.URL, savedBookmark.URL)
		}
		if savedBookmark.Title != req.Title {
			t.Errorf("Title: expected %s, got %s", req.Title, savedBookmark.Title)
		}
		if savedBookmark.Action != req.Action {
			t.Errorf("Action: expected %s, got %s", req.Action, savedBookmark.Action)
		}
	})
}

func TestGetTopicsFromDB(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		tdb.insertTestBookmarks(t)
		
		topics, err := getTopicsFromDB()
		if err != nil {
			t.Fatalf("getTopicsFromDB failed: %v", err)
		}
		
		expectedTopics := map[string]bool{
			"Programming":  true,
			"Development":  true,
		}
		
		if len(topics) != len(expectedTopics) {
			t.Errorf("Expected %d topics, got %d", len(expectedTopics), len(topics))
		}
		
		for _, topic := range topics {
			if !expectedTopics[topic] {
				t.Errorf("Unexpected topic: %s", topic)
			}
		}
	})
}

func TestGetStatsSummary(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		tdb.insertTestBookmarks(t)
		
		stats, err := getStatsSummary()
		if err != nil {
			t.Fatalf("getStatsSummary failed: %v", err)
		}
		
		if stats.TotalBookmarks != 5 {
			t.Errorf("Expected 5 total bookmarks, got %d", stats.TotalBookmarks)
		}
		
		if stats.NeedsTriage != 1 {
			t.Errorf("Expected 1 bookmark needing triage, got %d", stats.NeedsTriage)
		}
		
		if stats.ActiveProjects != 2 {
			t.Errorf("Expected 2 active projects, got %d", stats.ActiveProjects)
		}
		
		if stats.ReadyToShare != 1 {
			t.Errorf("Expected 1 bookmark ready to share, got %d", stats.ReadyToShare)
		}
		
		if stats.Archived != 0 {
			t.Errorf("Expected 0 archived bookmarks, got %d", stats.Archived)
		}
		
		if len(stats.ProjectStats) == 0 {
			t.Error("Expected project stats, got none")
		}
		
		// Test the new latest resource functionality
		for _, project := range stats.ProjectStats {
			if project.LatestURL == "" {
				t.Errorf("Expected latestURL for project %s, got empty string", project.Topic)
			}
			if project.LatestTitle == "" {
				t.Errorf("Expected latestTitle for project %s, got empty string", project.Topic)
			}
			
			// Validate specific projects based on test data
			switch project.Topic {
			case "Programming":
				if project.Count != 2 {
					t.Errorf("Expected Programming project to have 2 bookmarks, got %d", project.Count)
				}
				// Should contain the latest bookmark (either Example 2 or Example 5)
				if project.LatestURL != "https://example.com/2" && project.LatestURL != "https://example.com/5" {
					t.Errorf("Expected Programming project latest URL to be from test data, got %s", project.LatestURL)
				}
			case "Development":
				if project.Count != 1 {
					t.Errorf("Expected Development project to have 1 bookmark, got %d", project.Count)
				}
				if project.LatestURL != "https://example.com/4" {
					t.Errorf("Expected Development project latest URL to be https://example.com/4, got %s", project.LatestURL)
				}
				if project.LatestTitle != "Example 4" {
					t.Errorf("Expected Development project latest title to be 'Example 4', got %s", project.LatestTitle)
				}
			}
		}
	})
}

func TestGetTriageQueue(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		tdb.insertTestBookmarks(t)
		
		triageData, err := getTriageQueue(10, 0)
		if err != nil {
			t.Fatalf("getTriageQueue failed: %v", err)
		}
		
		if triageData.Total != 1 {
			t.Errorf("Expected 1 total triage item, got %d", triageData.Total)
		}
		
		if len(triageData.Bookmarks) != 1 {
			t.Errorf("Expected 1 triage bookmark, got %d", len(triageData.Bookmarks))
		}
		
		if triageData.Limit != 10 {
			t.Errorf("Expected limit 10, got %d", triageData.Limit)
		}
		
		if triageData.Offset != 0 {
			t.Errorf("Expected offset 0, got %d", triageData.Offset)
		}
		
		// Check first bookmark
		bookmark := triageData.Bookmarks[0]
		if bookmark.URL != "https://example.com/1" {
			t.Errorf("Expected URL 'https://example.com/1', got %s", bookmark.URL)
		}
		if bookmark.Domain != "example.com" {
			t.Errorf("Expected domain 'example.com', got %s", bookmark.Domain)
		}
	})
}

func TestGetProjects(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		tdb.insertTestBookmarks(t)
		
		projects, err := getProjects()
		if err != nil {
			t.Fatalf("getProjects failed: %v", err)
		}
		
		if len(projects.ActiveProjects) != 2 {
			t.Errorf("Expected 2 active projects, got %d", len(projects.ActiveProjects))
		}
		
		// Check if we have the expected topics
		found := map[string]bool{}
		for _, project := range projects.ActiveProjects {
			found[project.Topic] = true
			if project.LinkCount == 0 {
				t.Errorf("Expected project %s to have link count > 0", project.Topic)
			}
		}
		
		if !found["Programming"] || !found["Development"] {
			t.Error("Expected to find 'Programming' and 'Development' topics")
		}
	})
}

// HTTP Handler Tests

func TestHandleBookmark_Success(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		reqBody := BookmarkRequest{
			URL:         "https://example.com",
			Title:       "Test Title",
			Description: "Test Description",
			Content:     "Test Content",
			Action:      "working",
			ShareTo:     "",
			Topic:       "Development",
		}
		
		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}
		
		req := httptest.NewRequest("POST", "/bookmark", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		
		rr := httptest.NewRecorder()
		handleBookmark(rr, req)
		
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, rr.Code, rr.Body.String())
		}
		
		var response ProjectBookmark
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		
		if response.URL != reqBody.URL {
			t.Errorf("Expected URL '%s', got '%s'", reqBody.URL, response.URL)
		}
		
		if response.Title != reqBody.Title {
			t.Errorf("Expected title '%s', got '%s'", reqBody.Title, response.Title)
		}
		
		// Verify bookmark was actually saved
		var count int
		err = tdb.db.QueryRow("SELECT COUNT(*) FROM bookmarks WHERE url = ?", reqBody.URL).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to verify bookmark was saved: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected bookmark to be saved once, found %d times", count)
		}
	})
}

func TestHandleTopics_Success(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		tdb.insertTestBookmarks(t)
		
		req := httptest.NewRequest("GET", "/topics", nil)
		rr := httptest.NewRecorder()
		
		handleTopics(rr, req)
		
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Response body: %s", http.StatusOK, rr.Code, rr.Body.String())
			return
		}
		
		var response map[string][]string
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v. Response body: %s", err, rr.Body.String())
		}
		
		topics, exists := response["topics"]
		if !exists {
			t.Fatal("Response missing 'topics' field")
		}
		
		expectedTopics := map[string]bool{
			"Programming":  true,
			"Development":  true,
		}
		
		if len(topics) != len(expectedTopics) {
			t.Errorf("Expected %d topics, got %d", len(expectedTopics), len(topics))
		}
		
		for _, topic := range topics {
			if !expectedTopics[topic] {
				t.Errorf("Unexpected topic: %s", topic)
			}
		}
	})
}

func TestHandleStatsSummary_Success(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		tdb.insertTestBookmarks(t)
		
		req := httptest.NewRequest("GET", "/api/stats/summary", nil)
		rr := httptest.NewRecorder()
		
		handleStatsSummary(rr, req)
		
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, rr.Code, rr.Body.String())
		}
		
		var stats SummaryStats
		if err := json.Unmarshal(rr.Body.Bytes(), &stats); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		
		if stats.TotalBookmarks != 5 {
			t.Errorf("Expected 5 total bookmarks, got %d", stats.TotalBookmarks)
		}
		
		if stats.ActiveProjects != 2 {
			t.Errorf("Expected 2 active projects, got %d", stats.ActiveProjects)
		}
		
		// Test the new latest resource functionality in HTTP response
		if len(stats.ProjectStats) == 0 {
			t.Error("Expected project stats in HTTP response, got none")
		}
		
		for _, project := range stats.ProjectStats {
			if project.LatestURL == "" {
				t.Errorf("Expected latestURL for project %s in HTTP response, got empty string", project.Topic)
			}
			if project.LatestTitle == "" {
				t.Errorf("Expected latestTitle for project %s in HTTP response, got empty string", project.Topic)
			}
		}
	})
}

func TestHandleTriageQueue_Success(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		tdb.insertTestBookmarks(t)
		
		req := httptest.NewRequest("GET", "/api/bookmarks/triage", nil)
		rr := httptest.NewRecorder()
		
		handleTriageQueue(rr, req)
		
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, rr.Code, rr.Body.String())
		}
		
		var triageResponse TriageResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &triageResponse); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		
		if triageResponse.Total != 1 {
			t.Errorf("Expected 1 triage item, got %d", triageResponse.Total)
		}
	})
}

func TestHandleProjects_Success(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		tdb.insertTestBookmarks(t)
		
		req := httptest.NewRequest("GET", "/api/projects", nil)
		rr := httptest.NewRecorder()
		
		handleProjects(rr, req)
		
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, rr.Code, rr.Body.String())
		}
		
		var projectsResponse ProjectsResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &projectsResponse); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		
		if len(projectsResponse.ActiveProjects) != 2 {
			t.Errorf("Expected 2 active projects, got %d", len(projectsResponse.ActiveProjects))
		}
	})
}

func TestHandleDashboard_Success(t *testing.T) {
	// Create a temporary dashboard file
	dashboardPath := createDashboardFile(t)
	originalWd, _ := os.Getwd()
	tmpDir := filepath.Dir(dashboardPath)
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)
	
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	
	handleDashboard(rr, req)
	
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}
	
	if !strings.Contains(rr.Body.String(), "Test Dashboard") {
		t.Error("Expected dashboard HTML content")
	}
	
	contentType := rr.Header().Get("Content-Type")
	if contentType != "text/html" {
		t.Errorf("Expected Content-Type 'text/html', got %s", contentType)
	}
}

// Error case tests

func TestHandleBookmark_InvalidMethod(t *testing.T) {
	req := httptest.NewRequest("GET", "/bookmark", nil)
	rr := httptest.NewRecorder()
	
	handleBookmark(rr, req)
	
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestHandleBookmark_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/bookmark", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	handleBookmark(rr, req)
	
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestHandleBookmark_MissingURL(t *testing.T) {
	reqBody := BookmarkRequest{
		Title: "Test Title",
	}
	
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}
	
	req := httptest.NewRequest("POST", "/bookmark", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	handleBookmark(rr, req)
	
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestHandleBookmark_MissingTitle(t *testing.T) {
	reqBody := BookmarkRequest{
		URL: "https://example.com",
	}
	
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}
	
	req := httptest.NewRequest("POST", "/bookmark", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	handleBookmark(rr, req)
	
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestHandleTopics_InvalidMethod(t *testing.T) {
	req := httptest.NewRequest("POST", "/topics", nil)
	rr := httptest.NewRecorder()
	
	handleTopics(rr, req)
	
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestHandleDashboard_InvalidMethod(t *testing.T) {
	req := httptest.NewRequest("POST", "/", nil)
	rr := httptest.NewRecorder()
	
	handleDashboard(rr, req)
	
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

// Integration Tests

func TestGetSuggestedAction(t *testing.T) {
	tests := []struct {
		domain      string
		title       string
		description string
		expected    string
	}{
		{"github.com", "Some Project", "Code repository", "share"},
		{"stackoverflow.com", "How to code", "Programming question", "share"},
		{"example.com", "Tutorial Guide", "Learning resource", "share"},
		{"docs.example.com", "API Documentation", "Reference guide", "working"},
		{"example.com", "Random Article", "Just reading", "read-later"},
	}
	
	for i, test := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			result := getSuggestedAction(test.domain, test.title, test.description)
			if result != test.expected {
				t.Errorf("Expected %s, got %s for domain=%s, title=%s, description=%s",
					test.expected, result, test.domain, test.title, test.description)
			}
		})
	}
}

// End-to-end integration test
func TestBookmarkWorkflow_EndToEnd(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// 1. Add a bookmark
		reqBody := BookmarkRequest{
			URL:         "https://golang.org",
			Title:       "Go Programming Language",
			Description: "Official Go website",
			Action:      "working",
			Topic:       "Programming",
		}
		
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/bookmark", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		
		handleBookmark(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("Failed to add bookmark: %d", rr.Code)
		}
		
		// 2. Check that topics include our new topic
		req = httptest.NewRequest("GET", "/topics", nil)
		rr = httptest.NewRecorder()
		
		handleTopics(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("Failed to get topics: %d", rr.Code)
		}
		
		var topicsResponse map[string][]string
		json.Unmarshal(rr.Body.Bytes(), &topicsResponse)
		
		found := false
		for _, topic := range topicsResponse["topics"] {
			if topic == "Programming" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Programming topic not found in topics list")
		}
		
		// 3. Check stats show the bookmark
		req = httptest.NewRequest("GET", "/api/stats/summary", nil)
		rr = httptest.NewRecorder()
		
		handleStatsSummary(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("Failed to get stats: %d", rr.Code)
		}
		
		var stats SummaryStats
		json.Unmarshal(rr.Body.Bytes(), &stats)
		
		if stats.TotalBookmarks == 0 {
			t.Error("Expected at least 1 bookmark in stats")
		}
		if stats.ActiveProjects == 0 {
			t.Error("Expected at least 1 active project in stats")
		}
	})
}

// ============ COMPREHENSIVE PROJECTS TESTING ============

// Projects Unit Tests - Reference Collections

func TestGetReferenceCollections_EmptyDatabase(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		collections, err := getReferenceCollections()
		if err != nil {
			t.Fatalf("getReferenceCollections failed: %v", err)
		}
		
		if len(collections) != 0 {
			t.Errorf("Expected 0 reference collections in empty DB, got %d", len(collections))
		}
	})
}

func TestGetReferenceCollections_OnlyActiveProjects(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Insert only working bookmarks (should not appear in reference collections)
		insertSQL := `INSERT INTO bookmarks (url, title, action, topic, timestamp) VALUES (?, ?, ?, ?, ?)`
		
		testData := []struct {
			url, title, action, topic string
		}{
			{"https://example1.com", "Title 1", "working", "ActiveTopic1"},
			{"https://example2.com", "Title 2", "working", "ActiveTopic2"},
		}
		
		for i, data := range testData {
			_, err := tdb.db.Exec(insertSQL, data.url, data.title, data.action, data.topic, "2023-12-01 10:00:00")
			if err != nil {
				t.Fatalf("Failed to insert test data %d: %v", i, err)
			}
		}
		
		collections, err := getReferenceCollections()
		if err != nil {
			t.Fatalf("getReferenceCollections failed: %v", err)
		}
		
		if len(collections) != 0 {
			t.Errorf("Expected 0 reference collections when all topics are active, got %d", len(collections))
		}
	})
}

func TestGetReferenceCollections_MixedTopics(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		insertSQL := `INSERT INTO bookmarks (url, title, action, topic, timestamp) VALUES (?, ?, ?, ?, ?)`
		
		// Mix of working topics and reference topics
		testData := []struct {
			url, title, action, topic string
		}{
			{"https://example1.com", "Working 1", "working", "ActiveTopic"},
			{"https://example2.com", "Working 2", "working", "ActiveTopic"},
			{"https://example3.com", "Reference 1", "read-later", "ReferenceTopic1"},
			{"https://example4.com", "Reference 2", "share", "ReferenceTopic1"}, 
			{"https://example5.com", "Reference 3", "", "ReferenceTopic2"}, // Empty action
		}
		
		for i, data := range testData {
			_, err := tdb.db.Exec(insertSQL, data.url, data.title, data.action, data.topic, "2023-12-01 10:00:00")
			if err != nil {
				t.Fatalf("Failed to insert test data %d: %v", i, err)
			}
		}
		
		collections, err := getReferenceCollections()
		if err != nil {
			t.Fatalf("getReferenceCollections failed: %v", err)
		}
		
		if len(collections) != 2 {
			t.Errorf("Expected 2 reference collections, got %d", len(collections))
		}
		
		// Verify collections are sorted by count DESC
		if len(collections) >= 2 && collections[0].LinkCount < collections[1].LinkCount {
			t.Error("Reference collections should be sorted by link count DESC")
		}
	})
}

func TestGetReferenceCollections_TimestampParsing(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Test various timestamp formats
		insertSQL := `INSERT INTO bookmarks (url, title, action, topic, timestamp) VALUES (?, ?, ?, ?, ?)`
		
		timestamps := []string{
			"2023-12-01 10:00:00",     // SQLite format
			"2023-12-01T10:00:00Z",    // ISO format 
			"invalid-timestamp",        // Invalid format
		}
		
		for i, ts := range timestamps {
			url := fmt.Sprintf("https://example%d.com", i)
			topic := fmt.Sprintf("Topic%d", i)
			_, err := tdb.db.Exec(insertSQL, url, "Title", "read-later", topic, ts)
			if err != nil {
				t.Fatalf("Failed to insert test data %d: %v", i, err)
			}
		}
		
		collections, err := getReferenceCollections()
		if err != nil {
			t.Fatalf("getReferenceCollections failed: %v", err)
		}
		
		if len(collections) != 3 {
			t.Errorf("Expected 3 reference collections, got %d", len(collections))
		}
		
		// Check that invalid timestamps are handled gracefully
		for _, collection := range collections {
			if collection.LastAccessed == "" {
				t.Error("LastAccessed should not be empty")
			}
		}
	})
}

// Projects Unit Tests - Active Projects

func TestGetActiveProjects_EdgeCases(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Test edge cases - using current time for more reliable testing
		now := time.Now()
		futureDate := now.Add(24 * time.Hour).Format("2006-01-02 15:04:05")
		oldDate := now.Add(-60 * 24 * time.Hour).Format("2006-01-02 15:04:05") // 60 days ago
		staleDate := now.Add(-15 * 24 * time.Hour).Format("2006-01-02 15:04:05") // 15 days ago
		
		testCases := []struct {
			topic     string
			timestamp string
			expected  string // expected status
		}{
			{"FutureTopic", futureDate, "active"},     // Future date
			{"OldTopic", oldDate, "inactive"},         // Very old
			{"RecentTopic", staleDate, "stale"},       // Recent but not active
		}
		
		// Create projects first
		for _, tc := range testCases {
			tdb.createTestProject(t, tc.topic, "Test project for "+tc.topic, "active")
		}
		
		insertSQL := `INSERT INTO bookmarks (url, title, action, topic, timestamp) VALUES (?, ?, ?, ?, ?)`
		for i, tc := range testCases {
			url := fmt.Sprintf("https://example%d.com", i)
			_, err := tdb.db.Exec(insertSQL, url, "Title", "working", tc.topic, tc.timestamp)
			if err != nil {
				t.Fatalf("Failed to insert test data for %s: %v", tc.topic, err)
			}
		}
		
		projects, err := getActiveProjects()
		if err != nil {
			t.Fatalf("getActiveProjects failed: %v", err)
		}
		
		if len(projects) != 3 {
			t.Errorf("Expected 3 active projects, got %d", len(projects))
		}
		
		// Verify status calculation
		statusMap := make(map[string]string)
		for _, project := range projects {
			statusMap[project.Topic] = project.Status
		}
		
		for _, tc := range testCases {
			if statusMap[tc.topic] != tc.expected {
				t.Errorf("Topic %s: expected status %s, got %s", tc.topic, tc.expected, statusMap[tc.topic])
			}
		}
	})
}

func TestGetActiveProjects_LinkCounts(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Create topics with different link counts
		testCases := []struct {
			topic         string
			linkCount     int
		}{
			{"SmallProject", 1},
			{"MediumProject", 5},
			{"LargeProject", 15},
		}
		
		// Create projects first
		for _, tc := range testCases {
			tdb.createTestProject(t, tc.topic, "Test project for "+tc.topic, "active")
		}
		
		insertSQL := `INSERT INTO bookmarks (url, title, action, topic, timestamp) VALUES (?, ?, ?, ?, ?)`
		for _, tc := range testCases {
			for i := 0; i < tc.linkCount; i++ {
				url := fmt.Sprintf("https://%s-link%d.com", tc.topic, i)
				_, err := tdb.db.Exec(insertSQL, url, "Title", "working", tc.topic, "2023-12-01 10:00:00")
				if err != nil {
					t.Fatalf("Failed to insert link %d for %s: %v", i, tc.topic, err)
				}
			}
		}
		
		projects, err := getActiveProjects()
		if err != nil {
			t.Fatalf("getActiveProjects failed: %v", err)
		}
		
		linkCountMap := make(map[string]int)
		for _, project := range projects {
			linkCountMap[project.Topic] = project.LinkCount
		}
		
		for _, tc := range testCases {
			if linkCountMap[tc.topic] != tc.linkCount {
				t.Errorf("Topic %s: expected link count %d, got %d", tc.topic, tc.linkCount, linkCountMap[tc.topic])
			}
		}
	})
}

func TestGetActiveProjects_EmptyAndNullTopics(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Create project for valid topic
		tdb.createTestProject(t, "ValidTopic", "Test project for ValidTopic", "active")
		
		insertSQL := `INSERT INTO bookmarks (url, title, action, topic, timestamp) VALUES (?, ?, ?, ?, ?)`
		
		// Test handling of empty/null topics
		testData := []struct {
			url   string
			topic interface{} // Can be string or nil
		}{
			{"https://valid.com", "ValidTopic"},
			{"https://empty.com", ""},      // Empty string
			{"https://null.com", nil},      // NULL
		}
		
		for i, data := range testData {
			_, err := tdb.db.Exec(insertSQL, data.url, "Title", "working", data.topic, "2023-12-01 10:00:00")
			if err != nil {
				t.Fatalf("Failed to insert test data %d: %v", i, err)
			}
		}
		
		projects, err := getActiveProjects()
		if err != nil {
			t.Fatalf("getActiveProjects failed: %v", err)
		}
		
		// Only valid topic should be returned
		if len(projects) != 1 {
			t.Errorf("Expected 1 project with valid topic, got %d", len(projects))
		}
		
		if len(projects) > 0 && projects[0].Topic != "ValidTopic" {
			t.Errorf("Expected topic 'ValidTopic', got %s", projects[0].Topic)
		}
	})
}

func TestGetActiveProjects_SortingOrder(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Create projects with different timestamps
		testData := []struct {
			topic     string
			timestamp string
		}{
			{"OldestProject", "2023-11-01 10:00:00"},
			{"MiddleProject", "2023-11-15 10:00:00"},
			{"NewestProject", "2023-12-01 10:00:00"},
		}
		
		// Create projects first
		for _, data := range testData {
			tdb.createTestProject(t, data.topic, "Test project for "+data.topic, "active")
		}
		
		insertSQL := `INSERT INTO bookmarks (url, title, action, topic, timestamp) VALUES (?, ?, ?, ?, ?)`
		for i, data := range testData {
			url := fmt.Sprintf("https://example%d.com", i)
			_, err := tdb.db.Exec(insertSQL, url, "Title", "working", data.topic, data.timestamp)
			if err != nil {
				t.Fatalf("Failed to insert test data for %s: %v", data.topic, err)
			}
		}
		
		projects, err := getActiveProjects()
		if err != nil {
			t.Fatalf("getActiveProjects failed: %v", err)
		}
		
		if len(projects) != 3 {
			t.Fatalf("Expected 3 projects, got %d", len(projects))
		}
		
		// Should be sorted by timestamp DESC (newest first)
		expectedOrder := []string{"NewestProject", "MiddleProject", "OldestProject"}
		for i, expected := range expectedOrder {
			if projects[i].Topic != expected {
				t.Errorf("Position %d: expected %s, got %s", i, expected, projects[i].Topic)
			}
		}
	})
}

func TestProjects_TopicCaseHandling(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Test case sensitivity and special characters
		topics := []string{
			"JavaScript",
			"javascript", 
			"Java-Script",
			"Java_Script",
			"Java Script",
			"JAVASCRIPT",
		}
		
		// Create projects first
		for _, topic := range topics {
			tdb.createTestProject(t, topic, "Test project for "+topic, "active")
		}
		
		insertSQL := `INSERT INTO bookmarks (url, title, action, topic, timestamp) VALUES (?, ?, ?, ?, ?)`
		for i, topic := range topics {
			url := fmt.Sprintf("https://example%d.com", i)
			_, err := tdb.db.Exec(insertSQL, url, "Title", "working", topic, "2023-12-01 10:00:00")
			if err != nil {
				t.Fatalf("Failed to insert test data for topic %s: %v", topic, err)
			}
		}
		
		projects, err := getActiveProjects()
		if err != nil {
			t.Fatalf("getActiveProjects failed: %v", err)
		}
		
		// Each topic should be treated as separate
		if len(projects) != len(topics) {
			t.Errorf("Expected %d distinct topics, got %d", len(topics), len(projects))
		}
		
		// Verify all topics are present
		foundTopics := make(map[string]bool)
		for _, project := range projects {
			foundTopics[project.Topic] = true
		}
		
		for _, expectedTopic := range topics {
			if !foundTopics[expectedTopic] {
				t.Errorf("Topic %s not found in results", expectedTopic)
			}
		}
	})
}

// Projects HTTP Handler Tests

func TestHandleProjects_InvalidMethods(t *testing.T) {
	methods := []string{"POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}
	
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/projects", nil)
			rr := httptest.NewRecorder()
			
			handleProjects(rr, req)
			
			if rr.Code != http.StatusMethodNotAllowed {
				t.Errorf("Method %s: expected status %d, got %d", method, http.StatusMethodNotAllowed, rr.Code)
			}
		})
	}
}

func TestHandleProjects_Headers(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		req := httptest.NewRequest("GET", "/api/projects", nil)
		rr := httptest.NewRecorder()
		
		handleProjects(rr, req)
		
		if rr.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, rr.Code)
		}
		
		contentType := rr.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got %s", contentType)
		}
		
		// Verify it's valid JSON
		var response ProjectsResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Errorf("Response is not valid JSON: %v", err)
		}
	})
}

// Projects Integration Tests

func TestProjects_ResponseStructure(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Insert comprehensive test data
		testData := []struct {
			url, title, action, topic string
		}{
			{"https://active1.com", "Active 1", "working", "ActiveTopic1"},
			{"https://active2.com", "Active 2", "working", "ActiveTopic2"}, 
			{"https://ref1.com", "Ref 1", "read-later", "RefTopic1"},
			{"https://ref2.com", "Ref 2", "share", "RefTopic2"},
		}
		
		// Create projects for working topics
		tdb.createTestProject(t, "ActiveTopic1", "Test project for ActiveTopic1", "active")
		tdb.createTestProject(t, "ActiveTopic2", "Test project for ActiveTopic2", "active")
		
		insertSQL := `INSERT INTO bookmarks (url, title, action, topic, timestamp) VALUES (?, ?, ?, ?, ?)`
		for i, data := range testData {
			_, err := tdb.db.Exec(insertSQL, data.url, data.title, data.action, data.topic, "2023-12-01 10:00:00")
			if err != nil {
				t.Fatalf("Failed to insert test data %d: %v", i, err)
			}
		}
		
		req := httptest.NewRequest("GET", "/api/projects", nil)
		rr := httptest.NewRecorder()
		
		handleProjects(rr, req)
		
		if rr.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, rr.Code)
		}
		
		var response ProjectsResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		
		// Validate response structure
		if len(response.ActiveProjects) != 2 {
			t.Errorf("Expected 2 active projects, got %d", len(response.ActiveProjects))
		}
		
		if len(response.ReferenceCollections) != 2 {
			t.Errorf("Expected 2 reference collections, got %d", len(response.ReferenceCollections))
		}
		
		// Validate active project fields
		for _, project := range response.ActiveProjects {
			if project.Topic == "" {
				t.Error("Active project topic should not be empty")
			}
			if project.LinkCount <= 0 {
				t.Error("Active project link count should be > 0")
			}
			if project.LastUpdated == "" {
				t.Error("Active project lastUpdated should not be empty")
			}
			if project.Status == "" {
				t.Error("Active project status should not be empty")
			}
		}
		
		// Validate reference collection fields
		for _, collection := range response.ReferenceCollections {
			if collection.Topic == "" {
				t.Error("Reference collection topic should not be empty")
			}
			if collection.LinkCount <= 0 {
				t.Error("Reference collection link count should be > 0")
			}
			if collection.LastAccessed == "" {
				t.Error("Reference collection lastAccessed should not be empty")
			}
		}
	})
}

func TestProjectsWorkflow_EndToEnd(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// 1. Start with empty database
		req := httptest.NewRequest("GET", "/api/projects", nil)
		rr := httptest.NewRecorder()
		handleProjects(rr, req)
		
		var emptyResponse ProjectsResponse
		json.Unmarshal(rr.Body.Bytes(), &emptyResponse)
		
		if len(emptyResponse.ActiveProjects) != 0 || len(emptyResponse.ReferenceCollections) != 0 {
			t.Error("Expected empty projects in new database")
		}
		
		// 2. Add bookmarks and verify they appear as projects
		// Create projects first
		tdb.createTestProject(t, "WorkProject", "Test project for WorkProject", "active")
		
		insertSQL := `INSERT INTO bookmarks (url, title, action, topic, timestamp) VALUES (?, ?, ?, ?, ?)`
		
		// Add working project
		_, err := tdb.db.Exec(insertSQL, "https://work.com", "Work Item", "working", "WorkProject", "2023-12-01 10:00:00")
		if err != nil {
			t.Fatalf("Failed to insert working bookmark: %v", err)
		}
		
		// Add reference bookmark (doesn't need project since it's not "working")
		_, err = tdb.db.Exec(insertSQL, "https://ref.com", "Reference Item", "read-later", "RefProject", "2023-12-01 10:00:00")
		if err != nil {
			t.Fatalf("Failed to insert reference bookmark: %v", err)
		}
		
		// 3. Verify projects appear correctly
		req = httptest.NewRequest("GET", "/api/projects", nil)
		rr = httptest.NewRecorder()
		handleProjects(rr, req)
		
		var finalResponse ProjectsResponse
		json.Unmarshal(rr.Body.Bytes(), &finalResponse)
		
		if len(finalResponse.ActiveProjects) != 1 {
			t.Errorf("Expected 1 active project, got %d", len(finalResponse.ActiveProjects))
		}
		
		if len(finalResponse.ReferenceCollections) != 1 {
			t.Errorf("Expected 1 reference collection, got %d", len(finalResponse.ReferenceCollections))
		}
		
		// 4. Verify project details
		activeProject := finalResponse.ActiveProjects[0]
		if activeProject.Topic != "WorkProject" {
			t.Errorf("Expected active project 'WorkProject', got %s", activeProject.Topic)
		}
		if activeProject.LinkCount != 1 {
			t.Errorf("Expected link count 1, got %d", activeProject.LinkCount)
		}
		
		refCollection := finalResponse.ReferenceCollections[0]
		if refCollection.Topic != "RefProject" {
			t.Errorf("Expected reference collection 'RefProject', got %s", refCollection.Topic)
		}
		if refCollection.LinkCount != 1 {
			t.Errorf("Expected reference link count 1, got %d", refCollection.LinkCount)
		}
	})
}

// Test end states functionality
func TestEndStates(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		insertSQL := `INSERT INTO bookmarks (url, title, action, topic, timestamp) VALUES (?, ?, ?, ?, ?)`
		
		// Insert bookmarks with different end states
		testData := []struct {
			url, title, action, topic string
		}{
			{"https://archived1.com", "Archived Item 1", "archived", "TestProject"},
			{"https://archived2.com", "Archived Item 2", "archived", ""},
			{"https://irrelevant.com", "Irrelevant Item", "irrelevant", ""},
			{"https://active.com", "Active Item", "working", "TestProject"},
			{"https://share.com", "Share Item", "share", ""},
		}
		
		for i, data := range testData {
			_, err := tdb.db.Exec(insertSQL, data.url, data.title, data.action, data.topic, "2023-12-01 10:00:00")
			if err != nil {
				t.Fatalf("Failed to insert test data %d: %v", i, err)
			}
		}
		
		// Test stats calculation includes archived count
		stats, err := getStatsSummary()
		if err != nil {
			t.Fatalf("getStatsSummary failed: %v", err)
		}
		
		if stats.Archived != 2 {
			t.Errorf("Expected 2 archived bookmarks, got %d", stats.Archived)
		}
		
		if stats.TotalBookmarks != 5 {
			t.Errorf("Expected 5 total bookmarks, got %d", stats.TotalBookmarks)
		}
		
		if stats.ActiveProjects != 1 {
			t.Errorf("Expected 1 active project, got %d", stats.ActiveProjects)
		}
		
		if stats.ReadyToShare != 1 {
			t.Errorf("Expected 1 ready to share, got %d", stats.ReadyToShare)
		}
		
		// Test API response includes archived field
		req := httptest.NewRequest("GET", "/api/stats/summary", nil)
		rr := httptest.NewRecorder()
		handleStatsSummary(rr, req)
		
		if rr.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, rr.Code)
		}
		
		var apiStats SummaryStats
		if err := json.Unmarshal(rr.Body.Bytes(), &apiStats); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		
		if apiStats.Archived != 2 {
			t.Errorf("API response: expected 2 archived bookmarks, got %d", apiStats.Archived)
		}
	})
}

// Test bookmark update functionality
func TestBookmarkUpdate(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Insert a test bookmark
		insertSQL := `INSERT INTO bookmarks (url, title, action, topic, timestamp) VALUES (?, ?, ?, ?, ?)`
		result, err := tdb.db.Exec(insertSQL, "https://test.com", "Test Item", "read-later", "", "2023-12-01 10:00:00")
		if err != nil {
			t.Fatalf("Failed to insert test bookmark: %v", err)
		}
		
		bookmarkID, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("Failed to get bookmark ID: %v", err)
		}
		
		// Test updating bookmark to archived
		updateReq := BookmarkUpdateRequest{
			Action: "archived",
		}
		
		jsonBody, err := json.Marshal(updateReq)
		if err != nil {
			t.Fatalf("Failed to marshal update request: %v", err)
		}
		
		req := httptest.NewRequest("PATCH", fmt.Sprintf("/api/bookmarks/%d", bookmarkID), bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		
		handleBookmarkUpdate(rr, req)
		
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, rr.Code, rr.Body.String())
		}
		
		var response map[string]string
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		
		if response["status"] != "success" {
			t.Errorf("Expected status 'success', got %s", response["status"])
		}
		
		// Verify bookmark was actually updated in database
		var action string
		err = tdb.db.QueryRow("SELECT action FROM bookmarks WHERE id = ?", bookmarkID).Scan(&action)
		if err != nil {
			t.Fatalf("Failed to query updated bookmark: %v", err)
		}
		
		if action != "archived" {
			t.Errorf("Expected action 'archived', got %s", action)
		}
		
		// Test updating with topic
		updateReq = BookmarkUpdateRequest{
			Action: "working",
			Topic:  "TestProject",
		}
		
		jsonBody, _ = json.Marshal(updateReq)
		req = httptest.NewRequest("PATCH", fmt.Sprintf("/api/bookmarks/%d", bookmarkID), bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr = httptest.NewRecorder()
		
		handleBookmarkUpdate(rr, req)
		
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
		}
		
		// Verify topic was updated
		var updatedAction, updatedTopic string
		err = tdb.db.QueryRow("SELECT action, topic FROM bookmarks WHERE id = ?", bookmarkID).Scan(&updatedAction, &updatedTopic)
		if err != nil {
			t.Fatalf("Failed to query updated bookmark: %v", err)
		}
		
		if updatedAction != "working" {
			t.Errorf("Expected action 'working', got %s", updatedAction)
		}
		
		if updatedTopic != "TestProject" {
			t.Errorf("Expected topic 'TestProject', got %s", updatedTopic)
		}
	})
}

// Test bookmark update error cases
func TestBookmarkFullUpdate_PUT(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	// Set the global database
	db = testDB.db

	// Insert a test bookmark first
	insertSQL := `
	INSERT INTO bookmarks (url, title, description, action, topic, timestamp)
	VALUES (?, ?, ?, ?, ?, '2023-12-01 10:00:00')`
	
	result, err := testDB.db.Exec(insertSQL, 
		"https://old-example.com", "Old Title", "Old description", "read-later", "OldTopic")
	if err != nil {
		t.Fatalf("Failed to insert test bookmark: %v", err)
	}
	
	bookmarkID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get bookmark ID: %v", err)
	}

	// Test PUT request for full bookmark update
	updateData := BookmarkFullUpdateRequest{
		Title:       "Updated Title",
		URL:         "https://updated-example.com",
		Description: "Updated description",
		Action:      "working",
		Topic:       "UpdatedTopic",
		ShareTo:     "",
	}

	requestBody, _ := json.Marshal(updateData)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/bookmarks/%d", bookmarkID), bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handleBookmarkUpdate(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "success" {
		t.Errorf("Expected success status, got %s", response["status"])
	}

	// Verify the bookmark was updated in database
	var title, url, description, action, topic string
	err = testDB.db.QueryRow(
		"SELECT title, url, description, action, topic FROM bookmarks WHERE id = ?", 
		bookmarkID).Scan(&title, &url, &description, &action, &topic)
	if err != nil {
		t.Fatalf("Failed to query updated bookmark: %v", err)
	}

	if title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got '%s'", title)
	}
	if url != "https://updated-example.com" {
		t.Errorf("Expected URL 'https://updated-example.com', got '%s'", url)
	}
	if description != "Updated description" {
		t.Errorf("Expected description 'Updated description', got '%s'", description)
	}
	if action != "working" {
		t.Errorf("Expected action 'working', got '%s'", action)
	}
	if topic != "UpdatedTopic" {
		t.Errorf("Expected topic 'UpdatedTopic', got '%s'", topic)
	}

	// Verify project was created/assigned
	var projectCount int
	err = testDB.db.QueryRow("SELECT COUNT(*) FROM projects WHERE name = ?", "UpdatedTopic").Scan(&projectCount)
	if err != nil {
		t.Fatalf("Failed to query projects: %v", err)
	}
	if projectCount != 1 {
		t.Errorf("Expected 1 project with name 'UpdatedTopic', got %d", projectCount)
	}
}

func TestBookmarkFullUpdate_ValidationErrors(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	// Set the global database
	db = testDB.db

	tests := []struct {
		name     string
		data     BookmarkFullUpdateRequest
		expected int
	}{
		{
			name: "Missing title",
			data: BookmarkFullUpdateRequest{
				Title: "",
				URL:   "https://example.com",
			},
			expected: http.StatusInternalServerError, // Will fail in updateFullBookmarkInDB
		},
		{
			name: "Missing URL",
			data: BookmarkFullUpdateRequest{
				Title: "Test Title",
				URL:   "",
			},
			expected: http.StatusInternalServerError, // Will fail in updateFullBookmarkInDB
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody, _ := json.Marshal(tt.data)
			req := httptest.NewRequest(http.MethodPut, "/api/bookmarks/999", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handleBookmarkUpdate(w, req)

			if w.Code != tt.expected {
				t.Errorf("Expected status %d, got %d", tt.expected, w.Code)
			}
		})
	}
}

func TestBookmarkUpdate_ErrorCases(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Test invalid method
		req := httptest.NewRequest("GET", "/api/bookmarks/1", nil)
		rr := httptest.NewRecorder()
		handleBookmarkUpdate(rr, req)
		
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
		}
		
		// Test missing ID
		req = httptest.NewRequest("PATCH", "/api/bookmarks/", nil)
		rr = httptest.NewRecorder()
		handleBookmarkUpdate(rr, req)
		
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
		
		// Test invalid ID
		req = httptest.NewRequest("PATCH", "/api/bookmarks/invalid", nil)
		rr = httptest.NewRecorder()
		handleBookmarkUpdate(rr, req)
		
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
		
		// Test invalid JSON
		req = httptest.NewRequest("PATCH", "/api/bookmarks/1", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		rr = httptest.NewRecorder()
		handleBookmarkUpdate(rr, req)
		
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})
}

// Test Project Detail Handlers (0% coverage)
func TestHandleProjectDetail_Success(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Insert test project data
		insertSQL := `INSERT INTO bookmarks (url, title, description, content, action, topic, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?)`
		
		testData := []struct {
			url, title, description, content, action, topic string
		}{
			{"https://example1.com", "Title 1", "Desc 1", "Content 1", "working", "TestProject"},
			{"https://example2.com", "Title 2", "Desc 2", "Content 2", "working", "TestProject"},
			{"https://example3.com", "Title 3", "Desc 3", "Content 3", "working", "OtherProject"},
		}
		
		for i, data := range testData {
			_, err := tdb.db.Exec(insertSQL, data.url, data.title, data.description, data.content, data.action, data.topic, "2023-12-01 10:00:00")
			if err != nil {
				t.Fatalf("Failed to insert test data %d: %v", i, err)
			}
		}
		
		req := httptest.NewRequest("GET", "/api/projects/TestProject", nil)
		rr := httptest.NewRecorder()
		
		handleProjectDetail(rr, req)
		
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, rr.Code, rr.Body.String())
		}
		
		var response ProjectDetailResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		
		if response.Topic != "TestProject" {
			t.Errorf("Expected topic 'TestProject', got %s", response.Topic)
		}
		
		if response.LinkCount != 2 {
			t.Errorf("Expected link count 2, got %d", response.LinkCount)
		}
		
		if len(response.Bookmarks) != 2 {
			t.Errorf("Expected 2 bookmarks, got %d", len(response.Bookmarks))
		}
	})
}

func TestHandleProjectDetail_NotFound(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		req := httptest.NewRequest("GET", "/api/projects/NonexistentProject", nil)
		rr := httptest.NewRecorder()
		
		handleProjectDetail(rr, req)
		
		if rr.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, rr.Code)
		}
	})
}

func TestHandleProjectByID_Success(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Create a project first
		_, err := tdb.db.Exec("INSERT INTO projects (name, description, status) VALUES (?, ?, ?)", "Test Project", "Test Description", "active")
		if err != nil {
			t.Fatalf("Failed to create test project: %v", err)
		}
		
		// Get the project ID
		var projectID int
		err = tdb.db.QueryRow("SELECT id FROM projects WHERE name = ?", "Test Project").Scan(&projectID)
		if err != nil {
			t.Fatalf("Failed to get project ID: %v", err)
		}
		
		// Insert bookmarks for this project
		insertSQL := `INSERT INTO bookmarks (url, title, description, content, action, project_id, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?)`
		_, err = tdb.db.Exec(insertSQL, "https://test1.com", "Test 1", "Desc 1", "Content 1", "working", projectID, "2023-12-01 10:00:00")
		if err != nil {
			t.Fatalf("Failed to insert test bookmark: %v", err)
		}
		
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/projects/id/%d", projectID), nil)
		rr := httptest.NewRecorder()
		
		handleProjectByID(rr, req)
		
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, rr.Code, rr.Body.String())
		}
		
		var response ProjectDetailResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		
		if response.Topic != "Test Project" {
			t.Errorf("Expected project topic 'Test Project', got %s", response.Topic)
		}
		
		if response.LinkCount != 1 {
			t.Errorf("Expected link count 1, got %d", response.LinkCount)
		}
	})
}

func TestHandleProjectByID_InvalidID(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		req := httptest.NewRequest("GET", "/api/projects/id/invalid", nil)
		rr := httptest.NewRecorder()
		
		handleProjectByID(rr, req)
		
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})
}

func TestHandleProjectByID_NotFound(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		req := httptest.NewRequest("GET", "/api/projects/id/99999", nil)
		rr := httptest.NewRecorder()
		
		handleProjectByID(rr, req)
		
		if rr.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, rr.Code)
		}
	})
}

// Test Projects Page Handler (0% coverage)
func TestHandleProjectsPage_Success(t *testing.T) {
	// Create a temporary projects.html file
	tmpDir := t.TempDir()
	projectsPath := filepath.Join(tmpDir, "projects.html")
	
	projectsContent := `<!DOCTYPE html>
<html><head><title>Test Projects</title></head>
<body><h1>Test Projects</h1></body></html>`
	
	err := os.WriteFile(projectsPath, []byte(projectsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test projects file: %v", err)
	}
	
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)
	
	req := httptest.NewRequest("GET", "/projects", nil)
	rr := httptest.NewRecorder()
	
	handleProjectsPage(rr, req)
	
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}
	
	if !strings.Contains(rr.Body.String(), "Test Projects") {
		t.Error("Expected projects HTML content")
	}
	
	contentType := rr.Header().Get("Content-Type")
	if contentType != "text/html" {
		t.Errorf("Expected Content-Type 'text/html', got %s", contentType)
	}
}

func TestHandleProjectsPage_FileNotFound(t *testing.T) {
	// Test when projects.html doesn't exist
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)
	
	req := httptest.NewRequest("GET", "/projects", nil)
	rr := httptest.NewRecorder()
	
	handleProjectsPage(rr, req)
	
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

// Test Database Helper Functions (0% coverage)
func TestGetProjectDetail_Success(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Insert test data
		insertSQL := `INSERT INTO bookmarks (url, title, description, content, action, topic, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?)`
		
		testData := []struct {
			url, title, description, content, action, topic string
		}{
			{"https://example1.com", "Title 1", "Desc 1", "Content 1", "working", "TestProject"},
			{"https://example2.com", "Title 2", "Desc 2", "Content 2", "working", "TestProject"},
		}
		
		for i, data := range testData {
			_, err := tdb.db.Exec(insertSQL, data.url, data.title, data.description, data.content, data.action, data.topic, "2023-12-01 10:00:00")
			if err != nil {
				t.Fatalf("Failed to insert test data %d: %v", i, err)
			}
		}
		
		response, err := getProjectDetail("TestProject")
		if err != nil {
			t.Fatalf("getProjectDetail failed: %v", err)
		}
		
		if response.Topic != "TestProject" {
			t.Errorf("Expected topic 'TestProject', got %s", response.Topic)
		}
		
		if response.LinkCount != 2 {
			t.Errorf("Expected link count 2, got %d", response.LinkCount)
		}
		
		if len(response.Bookmarks) != 2 {
			t.Errorf("Expected 2 bookmarks, got %d", len(response.Bookmarks))
		}
		
		// Verify bookmark details
		for _, bookmark := range response.Bookmarks {
			if bookmark.Domain == "" {
				t.Error("Bookmark domain should not be empty")
			}
			if bookmark.Age == "" {
				t.Error("Bookmark age should not be empty")
			}
		}
	})
}

func TestGetProjectDetail_NotFound(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		_, err := getProjectDetail("NonexistentProject")
		if err == nil {
			t.Error("Expected error for nonexistent project")
		}
	})
}

func TestGetProjectBookmarks_Success(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Insert test data
		insertSQL := `INSERT INTO bookmarks (url, title, description, content, action, topic, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?)`
		_, err := tdb.db.Exec(insertSQL, "https://example.com", "Title", "Desc", "Content", "working", "TestProject", "2023-12-01 10:00:00")
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
		
		bookmarks, err := getProjectBookmarks("TestProject")
		if err != nil {
			t.Fatalf("getProjectBookmarks failed: %v", err)
		}
		
		if len(bookmarks) != 1 {
			t.Errorf("Expected 1 bookmark, got %d", len(bookmarks))
		}
		
		bookmark := bookmarks[0]
		if bookmark.URL != "https://example.com" {
			t.Errorf("Expected URL 'https://example.com', got %s", bookmark.URL)
		}
		if bookmark.Domain != "example.com" {
			t.Errorf("Expected domain 'example.com', got %s", bookmark.Domain)
		}
	})
}

func TestGetProjectDetailByID_Success(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Create a project
		result, err := tdb.db.Exec("INSERT INTO projects (name, description, status) VALUES (?, ?, ?)", "Test Project", "Test Description", "active")
		if err != nil {
			t.Fatalf("Failed to create test project: %v", err)
		}
		
		projectID, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("Failed to get project ID: %v", err)
		}
		
		// Insert bookmarks for this project
		insertSQL := `INSERT INTO bookmarks (url, title, description, content, action, project_id, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?)`
		_, err = tdb.db.Exec(insertSQL, "https://test.com", "Test", "Desc", "Content", "working", projectID, "2023-12-01 10:00:00")
		if err != nil {
			t.Fatalf("Failed to insert test bookmark: %v", err)
		}
		
		response, err := getProjectDetailByID(int(projectID))
		if err != nil {
			t.Fatalf("getProjectDetailByID failed: %v", err)
		}
		
		if response.Topic != "Test Project" {
			t.Errorf("Expected project topic 'Test Project', got %s", response.Topic)
		}
		
		if response.LinkCount != 1 {
			t.Errorf("Expected link count 1, got %d", response.LinkCount)
		}
	})
}

func TestGetProjectDetailByID_NotFound(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		_, err := getProjectDetailByID(99999)
		if err == nil {
			t.Error("Expected error for nonexistent project ID")
		}
	})
}

func TestGetProjectBookmarksByID_Success(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Create a project
		result, err := tdb.db.Exec("INSERT INTO projects (name, description, status) VALUES (?, ?, ?)", "Test Project", "Test Description", "active")
		if err != nil {
			t.Fatalf("Failed to create test project: %v", err)
		}
		
		projectID, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("Failed to get project ID: %v", err)
		}
		
		// Insert bookmarks for this project
		insertSQL := `INSERT INTO bookmarks (url, title, description, content, action, project_id, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?)`
		_, err = tdb.db.Exec(insertSQL, "https://test.com", "Test", "Desc", "Content", "working", projectID, "2023-12-01 10:00:00")
		if err != nil {
			t.Fatalf("Failed to insert test bookmark: %v", err)
		}
		
		bookmarks, err := getProjectBookmarksByID(int(projectID))
		if err != nil {
			t.Fatalf("getProjectBookmarksByID failed: %v", err)
		}
		
		if len(bookmarks) != 1 {
			t.Errorf("Expected 1 bookmark, got %d", len(bookmarks))
		}
		
		bookmark := bookmarks[0]
		if bookmark.URL != "https://test.com" {
			t.Errorf("Expected URL 'https://test.com', got %s", bookmark.URL)
		}
	})
}

// Test Database Initialization Functions (0% coverage - these are tricky to test)
func TestValidateDB_Success(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		originalDB := db
		db = tdb.db
		defer func() { db = originalDB }()
		
		err := validateDB()
		if err != nil {
			t.Errorf("validateDB failed on valid database: %v", err)
		}
	})
}

func TestValidateDB_MissingTable(t *testing.T) {
	// validateDB only checks connectivity, not schema - an empty DB should pass
	// Schema validation is handled by the migration system during startup
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "empty_test.db")
	
	testDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()
	
	originalDB := db
	db = testDB
	defer func() { db = originalDB }()
	
	err = validateDB()
	if err != nil {
		t.Errorf("validateDB should pass for empty database (only checks connectivity): %v", err)
	}
}

// Test Database Error Handling
func TestSaveBookmarkToDB_DatabaseError(t *testing.T) {
	// Test with closed database to trigger error
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "closed_test.db")
	
	testDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	testDB.Close() // Close it to cause errors
	
	originalDB := db
	db = testDB
	defer func() { db = originalDB }()
	
	req := BookmarkRequest{
		URL:   "https://example.com",
		Title: "Test Title",
	}
	
	err = saveBookmarkToDB(req)
	if err == nil {
		t.Error("Expected saveBookmarkToDB to fail with closed database")
	}
}

func TestUpdateBookmarkInDB_DatabaseError(t *testing.T) {
	// Test with closed database to trigger error
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "closed_test.db")
	
	testDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	testDB.Close() // Close it to cause errors
	
	originalDB := db
	db = testDB
	defer func() { db = originalDB }()
	
	req := BookmarkUpdateRequest{
		Action: "archived",
	}
	
	err = updateBookmarkInDB(1, req)
	if err == nil {
		t.Error("Expected updateBookmarkInDB to fail with closed database")
	}
}

// Test Logging Functions
func TestLogStructured_Success(t *testing.T) {
	// Create a temporary log file
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	
	// Create the log file
	logFile_test, err := os.Create(logPath)
	if err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}
	defer logFile_test.Close()
	
	// Save original state
	originalLogFile := logFile
	logFile = logFile_test
	defer func() { logFile = originalLogFile }()
	
	// Test logging
	logStructured("INFO", "test", "test message", map[string]interface{}{
		"key": "value",
	})
	
	// Verify log was written
	logFile_test.Close()
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	if !strings.Contains(string(content), "test message") {
		t.Error("Expected log message to be written")
	}
	
	if !strings.Contains(string(content), "INFO") {
		t.Error("Expected log level to be written")
	}
}

func TestLogStructured_WithNilFile(t *testing.T) {
	// Save original state
	originalLogFile := logFile
	logFile = nil
	defer func() { logFile = originalLogFile }()
	
	// This should not panic
	logStructured("INFO", "test", "test message", nil)
}

// Test Additional HTTP Handler Edge Cases
func TestHandleTriageQueue_WithPagination(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Insert multiple triage items
		insertSQL := `INSERT INTO bookmarks (url, title, action, timestamp) VALUES (?, ?, ?, ?)`
		
		for i := 0; i < 5; i++ {
			url := fmt.Sprintf("https://example%d.com", i)
			title := fmt.Sprintf("Title %d", i)
			_, err := tdb.db.Exec(insertSQL, url, title, "read-later", "2023-12-01 10:00:00")
			if err != nil {
				t.Fatalf("Failed to insert test data %d: %v", i, err)
			}
		}
		
		// Test with limit and offset
		req := httptest.NewRequest("GET", "/api/bookmarks/triage?limit=2&offset=1", nil)
		rr := httptest.NewRecorder()
		
		handleTriageQueue(rr, req)
		
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
		}
		
		var response TriageResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		
		if response.Limit != 2 {
			t.Errorf("Expected limit 2, got %d", response.Limit)
		}
		
		if response.Offset != 1 {
			t.Errorf("Expected offset 1, got %d", response.Offset)
		}
		
		if len(response.Bookmarks) > 2 {
			t.Errorf("Expected at most 2 bookmarks, got %d", len(response.Bookmarks))
		}
	})
}

func TestHandleTriageQueue_InvalidParameters(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Test with invalid limit
		req := httptest.NewRequest("GET", "/api/bookmarks/triage?limit=invalid", nil)
		rr := httptest.NewRecorder()
		
		handleTriageQueue(rr, req)
		
		// Should still work with default limit
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
		}
		
		// Test with invalid offset
		req = httptest.NewRequest("GET", "/api/bookmarks/triage?offset=invalid", nil)
		rr = httptest.NewRecorder()
		
		handleTriageQueue(rr, req)
		
		// Should still work with default offset
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
		}
	})
}

// Test Dashboard Error Cases
func TestHandleDashboard_FileNotFound(t *testing.T) {
	// Test when dashboard.html doesn't exist
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)
	
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	
	handleDashboard(rr, req)
	
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestHandleDashboard_FileReadError(t *testing.T) {
	// Create a directory instead of a file to cause read error
	tmpDir := t.TempDir()
	dashboardDir := filepath.Join(tmpDir, "dashboard.html")
	
	err := os.Mkdir(dashboardDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create dashboard directory: %v", err)
	}
	
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)
	
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	
	handleDashboard(rr, req)
	
	// Should return an error when trying to read a directory as a file
	if rr.Code == http.StatusOK {
		t.Error("Expected error when reading directory as file")
	}
}

// Test Stats Summary Edge Cases
func TestHandleStatsSummary_DatabaseError(t *testing.T) {
	// Test with closed database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "closed_test.db")
	
	testDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	testDB.Close() // Close it to cause errors
	
	originalDB := db
	db = testDB
	defer func() { db = originalDB }()
	
	req := httptest.NewRequest("GET", "/api/stats/summary", nil)
	rr := httptest.NewRecorder()
	
	handleStatsSummary(rr, req)
	
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestGetTopicsFromDB_DatabaseError(t *testing.T) {
	// Test with closed database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "closed_test.db")
	
	testDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	testDB.Close() // Close it to cause errors
	
	originalDB := db
	db = testDB
	defer func() { db = originalDB }()
	
	_, err = getTopicsFromDB()
	if err == nil {
		t.Error("Expected getTopicsFromDB to fail with closed database")
	}
}

func TestGetStatsSummary_DatabaseError(t *testing.T) {
	// Test with closed database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "closed_test.db")
	
	testDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	testDB.Close() // Close it to cause errors
	
	originalDB := db
	db = testDB
	defer func() { db = originalDB }()
	
	_, err = getStatsSummary()
	if err == nil {
		t.Error("Expected getStatsSummary to fail with closed database")
	}
}

// Test Project Stats Edge Cases
func TestGetProjectStats_DatabaseError(t *testing.T) {
	// Test with closed database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "closed_test.db")
	
	testDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	testDB.Close() // Close it to cause errors
	
	originalDB := db
	db = testDB
	defer func() { db = originalDB }()
	
	_, err = getProjectStats()
	if err == nil {
		t.Error("Expected getProjectStats to fail with closed database")
	}
}

func TestGetTriageQueue_DatabaseError(t *testing.T) {
	// Test with closed database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "closed_test.db")
	
	testDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	testDB.Close() // Close it to cause errors
	
	originalDB := db
	db = testDB
	defer func() { db = originalDB }()
	
	_, err = getTriageQueue(10, 0)
	if err == nil {
		t.Error("Expected getTriageQueue to fail with closed database")
	}
}

func TestGetProjects_DatabaseError(t *testing.T) {
	// Test with closed database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "closed_test.db")
	
	testDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	testDB.Close() // Close it to cause errors
	
	originalDB := db
	db = testDB
	defer func() { db = originalDB }()
	
	_, err = getProjects()
	if err == nil {
		t.Error("Expected getProjects to fail with closed database")
	}
}

// Test Additional Bookmark Validation Edge Cases
func TestSaveBookmarkToDB_EdgeCases(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Test with projectId
		req := BookmarkRequest{
			URL:       "https://example.com",
			Title:     "Test Title",
			Action:    "working",
			ProjectID: 1, // Will be ignored since project doesn't exist
		}
		
		err := saveBookmarkToDB(req)
		if err != nil {
			t.Errorf("saveBookmarkToDB failed: %v", err)
		}
		
		// Verify it was saved
		var count int
		err = tdb.db.QueryRow("SELECT COUNT(*) FROM bookmarks WHERE url = ?", req.URL).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query saved bookmark: %v", err)
		}
		
		if count != 1 {
			t.Errorf("Expected 1 bookmark, got %d", count)
		}
	})
}

func TestUpdateBookmarkInDB_EdgeCases(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Insert a test bookmark
		insertSQL := `INSERT INTO bookmarks (url, title, action, topic, timestamp) VALUES (?, ?, ?, ?, ?)`
		result, err := tdb.db.Exec(insertSQL, "https://test.com", "Test", "read-later", "", "2023-12-01 10:00:00")
		if err != nil {
			t.Fatalf("Failed to insert test bookmark: %v", err)
		}
		
		bookmarkID, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("Failed to get bookmark ID: %v", err)
		}
		
		// Create a test project first
		tdb.createTestProject(t, "TestProject", "Test project", "active")
		
		// Get the project ID
		var projectID int
		err = tdb.db.QueryRow("SELECT id FROM projects WHERE name = ?", "TestProject").Scan(&projectID)
		if err != nil {
			t.Fatalf("Failed to get project ID: %v", err)
		}
		
		// Test updating with valid projectId
		req := BookmarkUpdateRequest{
			Action:    "working",
			ProjectID: projectID,
		}
		
		err = updateBookmarkInDB(int(bookmarkID), req)
		if err != nil {
			t.Errorf("updateBookmarkInDB failed: %v", err)
		}
		
		// Verify it was updated
		var action string
		var updatedProjectId sql.NullInt64
		err = tdb.db.QueryRow("SELECT action, project_id FROM bookmarks WHERE id = ?", bookmarkID).Scan(&action, &updatedProjectId)
		if err != nil {
			t.Fatalf("Failed to query updated bookmark: %v", err)
		}
		
		if action != "working" {
			t.Errorf("Expected action 'working', got %s", action)
		}
	})
}

// Test URL Parsing Edge Cases
func TestBookmarkDetailResponseDomain(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Insert bookmarks with various URL formats
		insertSQL := `INSERT INTO bookmarks (url, title, action, topic, timestamp) VALUES (?, ?, ?, ?, ?)`
		
		testCases := []struct {
			url            string
			expectedDomain string
		}{
			{"https://example.com/path", "example.com"},
			{"http://sub.example.com", "sub.example.com"},
			{"https://example.com:8080/path", "example.com:8080"},
			{"invalid-url", "invalid-url"}, // Should handle invalid URLs gracefully
			{"", ""},                       // Empty URL
		}
		
		for i, tc := range testCases {
			title := fmt.Sprintf("Test %d", i)
			_, err := tdb.db.Exec(insertSQL, tc.url, title, "read-later", "TestTopic", "2023-12-01 10:00:00")
			if err != nil {
				t.Fatalf("Failed to insert test data %d: %v", i, err)
			}
		}
		
		// Get triage queue to test domain parsing
		triageData, err := getTriageQueue(10, 0)
		if err != nil {
			t.Fatalf("getTriageQueue failed: %v", err)
		}
		
		// Verify domain parsing
		for i, bookmark := range triageData.Bookmarks {
			if i < len(testCases) {
				expectedDomain := testCases[i].expectedDomain
				if bookmark.Domain != expectedDomain {
					t.Errorf("Bookmark %d: expected domain %s, got %s", i, expectedDomain, bookmark.Domain)
				}
			}
		}
	})
}

// ============ ENHANCED PROJECT DETAIL TESTS ============

// Test Enhanced Project Detail Page Handler
func TestHandleProjectDetailPage_Success(t *testing.T) {
	req := httptest.NewRequest("GET", "/project-detail", nil)
	rr := httptest.NewRecorder()
	
	handleProjectDetailPage(rr, req)
	
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
	
	contentType := rr.Header().Get("Content-Type")
	if contentType != "text/html" {
		t.Errorf("Expected Content-Type 'text/html', got %s", contentType)
	}
	
	// Check for essential HTML elements
	body := rr.Body.String()
	expectedElements := []string{
		"<title>Project Detail - BookMinder</title>",
		"id=\"searchFilter\"",
		"id=\"actionFilter\"",
		"id=\"domainFilter\"",
		"id=\"sortField\"",
		"loadProjectData()",
		"applyFilters()",
	}
	
	for _, element := range expectedElements {
		if !strings.Contains(body, element) {
			t.Errorf("Expected HTML to contain %s", element)
		}
	}
}

func TestHandleProjectDetailPage_InvalidMethod(t *testing.T) {
	methods := []string{"POST", "PUT", "DELETE", "PATCH"}
	
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/project-detail", nil)
			rr := httptest.NewRecorder()
			
			handleProjectDetailPage(rr, req)
			
			if rr.Code != http.StatusMethodNotAllowed {
				t.Errorf("Method %s: expected status %d, got %d", method, http.StatusMethodNotAllowed, rr.Code)
			}
		})
	}
}

// Test Enhanced ActiveProject Structure
func TestActiveProject_IncludesID(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Create a project in the projects table first
		_, err := tdb.db.Exec("INSERT INTO projects (name, description, status) VALUES (?, ?, ?)", 
			"Test Project", "Test Description", "active")
		if err != nil {
			t.Fatalf("Failed to create test project: %v", err)
		}
		
		// Add a bookmark for this project
		_, err = tdb.db.Exec(`INSERT INTO bookmarks (url, title, action, topic, timestamp) VALUES (?, ?, ?, ?, ?)`,
			"https://test.com", "Test Bookmark", "working", "Test Project", "2023-12-01 10:00:00")
		if err != nil {
			t.Fatalf("Failed to insert test bookmark: %v", err)
		}
		
		projects, err := getActiveProjects()
		if err != nil {
			t.Fatalf("getActiveProjects failed: %v", err)
		}
		
		if len(projects) == 0 {
			t.Fatal("Expected at least one active project")
		}
		
		project := projects[0]
		if project.ID == 0 {
			t.Error("Expected project ID to be non-zero")
		}
		
		if project.Topic == "" {
			t.Error("Expected project topic to be non-empty")
		}
		
		if project.LinkCount == 0 {
			t.Error("Expected project link count to be non-zero")
		}
		
		if project.Status == "" {
			t.Error("Expected project status to be non-empty")
		}
	})
}

func TestGetActiveProjects_ProjectsTable(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Create multiple projects
		projects := []struct {
			name, description, status string
		}{
			{"Project A", "Description A", "active"},
			{"Project B", "Description B", "active"},
			{"Project C", "Description C", "inactive"}, // Should not appear
		}
		
		var projectIDs []int64
		for _, proj := range projects {
			result, err := tdb.db.Exec("INSERT INTO projects (name, description, status) VALUES (?, ?, ?)", 
				proj.name, proj.description, proj.status)
			if err != nil {
				t.Fatalf("Failed to create project %s: %v", proj.name, err)
			}
			id, _ := result.LastInsertId()
			projectIDs = append(projectIDs, id)
		}
		
		// Add bookmarks for active projects only
		for i, proj := range projects[:2] { // Only first 2 (active ones)
			_, err := tdb.db.Exec(`INSERT INTO bookmarks (url, title, action, topic, project_id, timestamp) VALUES (?, ?, ?, ?, ?, ?)`,
				fmt.Sprintf("https://test%d.com", i), fmt.Sprintf("Test %d", i), "working", proj.name, projectIDs[i], "2023-12-01 10:00:00")
			if err != nil {
				t.Fatalf("Failed to insert bookmark for project %s: %v", proj.name, err)
			}
		}
		
		activeProjects, err := getActiveProjects()
		if err != nil {
			t.Fatalf("getActiveProjects failed: %v", err)
		}
		
		// Should only return active projects with bookmarks
		if len(activeProjects) != 2 {
			t.Errorf("Expected 2 active projects, got %d", len(activeProjects))
		}
		
		// Verify project IDs are included and correct
		foundProjects := make(map[string]int)
		for _, project := range activeProjects {
			foundProjects[project.Topic] = project.ID
			
			if project.ID == 0 {
				t.Errorf("Project %s has zero ID", project.Topic)
			}
			
			if project.LinkCount == 0 {
				t.Errorf("Project %s has zero link count", project.Topic)
			}
		}
		
		if _, found := foundProjects["Project A"]; !found {
			t.Error("Expected to find Project A in active projects")
		}
		
		if _, found := foundProjects["Project B"]; !found {
			t.Error("Expected to find Project B in active projects")
		}
		
		if _, found := foundProjects["Project C"]; found {
			t.Error("Did not expect to find inactive Project C in active projects")
		}
	})
}

// Test Project Detail by ID Functionality  
func TestProjectDetailByID_Integration(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Create a project
		result, err := tdb.db.Exec("INSERT INTO projects (name, description, status) VALUES (?, ?, ?)", 
			"Integration Test Project", "Test Description", "active")
		if err != nil {
			t.Fatalf("Failed to create test project: %v", err)
		}
		
		projectID, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("Failed to get project ID: %v", err)
		}
		
		// Add multiple bookmarks with different actions and domains
		bookmarks := []struct {
			url, title, description, action string
		}{
			{"https://example.com/1", "Example 1", "First example", "working"},
			{"https://github.com/test", "GitHub Test", "GitHub repository", "working"},
			{"https://example.com/2", "Example 2", "Second example", "share"},
			{"https://docs.example.com", "Documentation", "API docs", "read-later"},
		}
		
		for i, bookmark := range bookmarks {
			_, err := tdb.db.Exec(`INSERT INTO bookmarks (url, title, description, action, project_id, timestamp) VALUES (?, ?, ?, ?, ?, ?)`,
				bookmark.url, bookmark.title, bookmark.description, bookmark.action, projectID, fmt.Sprintf("2023-12-0%d 10:00:00", i+1))
			if err != nil {
				t.Fatalf("Failed to insert bookmark %d: %v", i, err)
			}
		}
		
		// Test the project detail by ID endpoint
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/projects/id/%d", projectID), nil)
		rr := httptest.NewRecorder()
		
		handleProjectByID(rr, req)
		
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, rr.Code, rr.Body.String())
		}
		
		var response ProjectDetailResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		
		// Verify project details
		if response.Topic != "Integration Test Project" {
			t.Errorf("Expected topic 'Integration Test Project', got %s", response.Topic)
		}
		
		if response.LinkCount != 4 {
			t.Errorf("Expected link count 4, got %d", response.LinkCount)
		}
		
		if len(response.Bookmarks) != 4 {
			t.Errorf("Expected 4 bookmarks, got %d", len(response.Bookmarks))
		}
		
		// Verify bookmark details for client-side filtering
		domainCounts := make(map[string]int)
		actionCounts := make(map[string]int)
		
		for _, bookmark := range response.Bookmarks {
			// Verify required fields for filtering
			if bookmark.URL == "" {
				t.Error("Bookmark URL should not be empty")
			}
			if bookmark.Title == "" {
				t.Error("Bookmark title should not be empty")
			}
			if bookmark.Domain == "" {
				t.Error("Bookmark domain should not be empty for client-side filtering")
			}
			if bookmark.Timestamp == "" {
				t.Error("Bookmark timestamp should not be empty for date filtering")
			}
			if bookmark.Age == "" {
				t.Error("Bookmark age should not be empty")
			}
			
			domainCounts[bookmark.Domain]++
			actionCounts[bookmark.Action]++
		}
		
		// Verify we have the expected domains for filtering
		if domainCounts["example.com"] != 2 {
			t.Errorf("Expected 2 bookmarks from example.com, got %d", domainCounts["example.com"])
		}
		
		if domainCounts["github.com"] != 1 {
			t.Errorf("Expected 1 bookmark from github.com, got %d", domainCounts["github.com"])
		}
		
		if domainCounts["docs.example.com"] != 1 {
			t.Errorf("Expected 1 bookmark from docs.example.com, got %d", domainCounts["docs.example.com"])
		}
		
		// Verify we have the expected actions for filtering
		if actionCounts["working"] != 2 {
			t.Errorf("Expected 2 working bookmarks, got %d", actionCounts["working"])
		}
		
		if actionCounts["share"] != 1 {
			t.Errorf("Expected 1 share bookmark, got %d", actionCounts["share"])
		}
		
		if actionCounts["read-later"] != 1 {
			t.Errorf("Expected 1 read-later bookmark, got %d", actionCounts["read-later"])
		}
	})
}

// Test Projects API Response Structure
func TestProjectsAPI_IncludesProjectIDs(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Create test projects
		projects := []struct {
			name, status string
		}{
			{"API Test Project 1", "active"},
			{"API Test Project 2", "active"},
		}
		
		for _, proj := range projects {
			result, err := tdb.db.Exec("INSERT INTO projects (name, status) VALUES (?, ?)", proj.name, proj.status)
			if err != nil {
				t.Fatalf("Failed to create project %s: %v", proj.name, err)
			}
			
			// Add a bookmark to make it appear in active projects
			projectID, _ := result.LastInsertId()
			_, err = tdb.db.Exec(`INSERT INTO bookmarks (url, title, action, project_id, timestamp) VALUES (?, ?, ?, ?, ?)`,
				"https://test.com", "Test", "working", projectID, "2023-12-01 10:00:00")
			if err != nil {
				t.Fatalf("Failed to insert bookmark for project %s: %v", proj.name, err)
			}
		}
		
		// Test the projects API endpoint
		req := httptest.NewRequest("GET", "/api/projects", nil)
		rr := httptest.NewRecorder()
		
		handleProjects(rr, req)
		
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
		}
		
		var response ProjectsResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		
		if len(response.ActiveProjects) < 2 {
			t.Errorf("Expected at least 2 active projects, got %d", len(response.ActiveProjects))
		}
		
		// Verify all active projects have IDs
		for i, project := range response.ActiveProjects {
			if project.ID == 0 {
				t.Errorf("Active project %d has zero ID", i)
			}
			
			if project.Topic == "" {
				t.Errorf("Active project %d has empty topic", i)
			}
			
			if project.LinkCount == 0 {
				t.Errorf("Active project %d has zero link count", i)
			}
		}
	})
}

// Test Client-Side Filtering Data Integrity
func TestProjectDetail_FilteringDataIntegrity(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Create test project first
		tdb.createTestProject(t, "TestProject", "Test project for filtering", "active")
		
		// Insert test data with various scenarios for filtering
		insertSQL := `INSERT INTO bookmarks (url, title, description, content, action, topic, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?)`
		
		testCases := []struct {
			url, title, description, content, action, topic, timestamp string
		}{
			// Different domains
			{"https://github.com/test", "GitHub Repo", "Repository", "Code", "working", "TestProject", "2023-12-01 10:00:00"},
			{"https://stackoverflow.com/q/123", "Stack Question", "Programming help", "Answer", "share", "TestProject", "2023-12-02 11:00:00"},
			{"https://docs.github.com", "GitHub Docs", "Documentation", "Guide", "read-later", "TestProject", "2023-12-03 12:00:00"},
			
			// Different actions
			{"https://example.com/archive", "Archived Item", "Old stuff", "Legacy", "archived", "TestProject", "2023-11-01 10:00:00"},
			{"https://example.com/irrelevant", "Irrelevant Item", "Not useful", "Ignore", "irrelevant", "TestProject", "2023-11-02 10:00:00"},
			
			// Edge cases
			{"https://test.com", "Empty Description", "", "", "", "TestProject", "2023-12-04 13:00:00"},
			{"https://special-chars.com", "Special & Characters", "Test <script>", "Content & stuff", "working", "TestProject", "2023-12-05 14:00:00"},
		}
		
		for i, tc := range testCases {
			_, err := tdb.db.Exec(insertSQL, tc.url, tc.title, tc.description, tc.content, tc.action, tc.topic, tc.timestamp)
			if err != nil {
				t.Fatalf("Failed to insert test case %d: %v", i, err)
			}
		}
		
		// Get project detail
		projectDetail, err := getProjectDetail("TestProject")
		if err != nil {
			t.Fatalf("getProjectDetail failed: %v", err)
		}
		
		if projectDetail == nil {
			t.Fatal("Expected project detail, got nil")
		}
		
		if len(projectDetail.Bookmarks) != len(testCases) {
			t.Errorf("Expected %d bookmarks, got %d", len(testCases), len(projectDetail.Bookmarks))
		}
		
		// Verify data integrity for client-side filtering
		domains := make(map[string]bool)
		actions := make(map[string]bool)
		timestamps := make([]string, 0)
		
		for _, bookmark := range projectDetail.Bookmarks {
			// Check domain extraction
			if bookmark.Domain != "" {
				domains[bookmark.Domain] = true
			}
			
			// Check action handling
			if bookmark.Action != "" {
				actions[bookmark.Action] = true
			}
			
			// Check timestamp format
			if bookmark.Timestamp != "" {
				timestamps = append(timestamps, bookmark.Timestamp)
			}
			
			// Note: HTML escaping is now handled by frontend for display
			// Backend APIs return raw data for proper data integrity
		}
		
		// Verify expected domains are present for filtering
		expectedDomains := []string{"github.com", "stackoverflow.com", "docs.github.com", "example.com", "test.com", "special-chars.com"}
		for _, domain := range expectedDomains {
			if !domains[domain] {
				t.Errorf("Expected domain %s not found in results", domain)
			}
		}
		
		// Verify expected actions are present for filtering
		expectedActions := []string{"working", "share", "read-later", "archived", "irrelevant"}
		for _, action := range expectedActions {
			if action != "" && !actions[action] {
				t.Errorf("Expected action %s not found in results", action)
			}
		}
		
		// Verify timestamp format for date filtering
		for i, timestamp := range timestamps {
			if _, err := time.Parse(time.RFC3339, timestamp); err != nil {
				t.Errorf("Timestamp %d (%s) is not in RFC3339 format: %v", i, timestamp, err)
			}
		}
	})
}

// Test Error Handling for Enhanced Project Detail
func TestProjectDetailPage_ErrorHandling(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Test project not found by ID
		req := httptest.NewRequest("GET", "/api/projects/id/99999", nil)
		rr := httptest.NewRecorder()
		
		handleProjectByID(rr, req)
		
		if rr.Code != http.StatusNotFound {
			t.Errorf("Expected status %d for non-existent project ID, got %d", http.StatusNotFound, rr.Code)
		}
		
		// Test invalid project ID format
		req = httptest.NewRequest("GET", "/api/projects/id/invalid", nil)
		rr = httptest.NewRecorder()
		
		handleProjectByID(rr, req)
		
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for invalid project ID, got %d", http.StatusBadRequest, rr.Code)
		}
		
		// Test missing project ID
		req = httptest.NewRequest("GET", "/api/projects/id/", nil)
		rr = httptest.NewRecorder()
		
		handleProjectByID(rr, req)
		
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for missing project ID, got %d", http.StatusBadRequest, rr.Code)
		}
	})
}

// Test Bookmark Update Endpoints - PUT vs PATCH
func TestBookmarkUpdate_PutVsPatch(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Insert a test bookmark
		insertSQL := `
		INSERT INTO bookmarks (url, title, description, action, topic, shareTo, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, '2023-12-01 10:00:00')`
		
		result, err := tdb.db.Exec(insertSQL, 
			"https://original.example.com", 
			"Original Title", 
			"Original description", 
			"read-later", 
			"OriginalTopic",
			"")
		if err != nil {
			t.Fatalf("Failed to insert test bookmark: %v", err)
		}
		
		bookmarkID, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("Failed to get bookmark ID: %v", err)
		}

		t.Run("PATCH should update metadata only", func(t *testing.T) {
			// Test PATCH request (partial update - metadata only)
			patchData := BookmarkUpdateRequest{
				Action:  "working",
				Topic:   "UpdatedTopic",
				ShareTo: "Newsletter",
			}
			
			jsonData, _ := json.Marshal(patchData)
			req := httptest.NewRequest("PATCH", fmt.Sprintf("/api/bookmarks/%d", bookmarkID), bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			
			handleBookmarkUpdate(rr, req)
			
			if rr.Code != http.StatusOK {
				t.Errorf("PATCH request failed with status %d, body: %s", rr.Code, rr.Body.String())
			}
			
			// Verify response contains updated bookmark
			var response ProjectBookmark
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal PATCH response: %v", err)
			}
			
			// Check that metadata was updated
			if response.Action != "working" {
				t.Errorf("Expected action 'working', got %s", response.Action)
			}
			if response.Topic != "UpdatedTopic" {
				t.Errorf("Expected topic 'UpdatedTopic', got %s", response.Topic)
			}
			if response.ShareTo != "Newsletter" {
				t.Errorf("Expected shareTo 'Newsletter', got %s", response.ShareTo)
			}
			
			// Check that content fields were preserved
			if response.Title != "Original Title" {
				t.Errorf("Expected title preserved as 'Original Title', got %s", response.Title)
			}
			if response.URL != "https://original.example.com" {
				t.Errorf("Expected URL preserved, got %s", response.URL)
			}
			if response.Description != "Original description" {
				t.Errorf("Expected description preserved, got %s", response.Description)
			}
			
			// Check computed fields
			if response.Domain != "original.example.com" {
				t.Errorf("Expected domain 'original.example.com', got %s", response.Domain)
			}
			if response.Age == "" {
				t.Error("Expected age to be calculated")
			}
		})

		t.Run("PUT should update all fields", func(t *testing.T) {
			// Test PUT request (full update - can update title, URL, description)
			putData := BookmarkFullUpdateRequest{
				Title:       "UPDATED: New Title",
				URL:         "https://updated.example.com/new-path",
				Description: "Completely new description",
				Action:      "share",
				Topic:       "NewTopic",
				ShareTo:     "Team Slack",
			}
			
			jsonData, _ := json.Marshal(putData)
			req := httptest.NewRequest("PUT", fmt.Sprintf("/api/bookmarks/%d", bookmarkID), bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			
			handleBookmarkUpdate(rr, req)
			
			if rr.Code != http.StatusOK {
				t.Errorf("PUT request failed with status %d, body: %s", rr.Code, rr.Body.String())
			}
			
			// Verify response contains updated bookmark
			var response ProjectBookmark
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal PUT response: %v", err)
			}
			
			// Check that ALL fields were updated
			if response.Title != "UPDATED: New Title" {
				t.Errorf("Expected title 'UPDATED: New Title', got %s", response.Title)
			}
			if response.URL != "https://updated.example.com/new-path" {
				t.Errorf("Expected URL 'https://updated.example.com/new-path', got %s", response.URL)
			}
			if response.Description != "Completely new description" {
				t.Errorf("Expected description 'Completely new description', got %s", response.Description)
			}
			if response.Action != "share" {
				t.Errorf("Expected action 'share', got %s", response.Action)
			}
			if response.Topic != "NewTopic" {
				t.Errorf("Expected topic 'NewTopic', got %s", response.Topic)
			}
			if response.ShareTo != "Team Slack" {
				t.Errorf("Expected shareTo 'Team Slack', got %s", response.ShareTo)
			}
			
			// Check computed fields were recalculated
			if response.Domain != "updated.example.com" {
				t.Errorf("Expected domain 'updated.example.com', got %s", response.Domain)
			}
			if response.Age == "" {
				t.Error("Expected age to be calculated")
			}
			
			// Verify the changes persisted in database
			var dbTitle, dbURL, dbDescription, dbAction, dbTopic, dbShareTo string
			err = tdb.db.QueryRow(`
				SELECT title, url, description, action, topic, shareTo 
				FROM bookmarks WHERE id = ?`, bookmarkID).Scan(
				&dbTitle, &dbURL, &dbDescription, &dbAction, &dbTopic, &dbShareTo)
			if err != nil {
				t.Fatalf("Failed to query updated bookmark from database: %v", err)
			}
			
			if dbTitle != "UPDATED: New Title" {
				t.Errorf("Title not persisted in database. Expected 'UPDATED: New Title', got %s", dbTitle)
			}
			if dbURL != "https://updated.example.com/new-path" {
				t.Errorf("URL not persisted in database. Got %s", dbURL)
			}
		})
	})
}

// Test that PUT endpoint validates required fields
func TestBookmarkUpdate_PUT_Validation(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Insert a test bookmark
		insertSQL := `
		INSERT INTO bookmarks (url, title, description, timestamp)
		VALUES (?, ?, ?, '2023-12-01 10:00:00')`
		
		result, err := tdb.db.Exec(insertSQL, 
			"https://test.example.com", "Test Title", "Test description")
		if err != nil {
			t.Fatalf("Failed to insert test bookmark: %v", err)
		}
		
		bookmarkID, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("Failed to get bookmark ID: %v", err)
		}

		t.Run("PUT should reject missing title", func(t *testing.T) {
			putData := BookmarkFullUpdateRequest{
				// Title missing
				URL:         "https://test.example.com",
				Description: "Test description",
			}
			
			jsonData, _ := json.Marshal(putData)
			req := httptest.NewRequest("PUT", fmt.Sprintf("/api/bookmarks/%d", bookmarkID), bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			
			handleBookmarkUpdate(rr, req)
			
			if rr.Code != http.StatusInternalServerError {
				t.Errorf("Expected status %d for missing title, got %d", http.StatusInternalServerError, rr.Code)
			}
		})

		t.Run("PUT should reject missing URL", func(t *testing.T) {
			putData := BookmarkFullUpdateRequest{
				Title: "Test Title",
				// URL missing
				Description: "Test description",
			}
			
			jsonData, _ := json.Marshal(putData)
			req := httptest.NewRequest("PUT", fmt.Sprintf("/api/bookmarks/%d", bookmarkID), bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			
			handleBookmarkUpdate(rr, req)
			
			if rr.Code != http.StatusInternalServerError {
				t.Errorf("Expected status %d for missing URL, got %d", http.StatusInternalServerError, rr.Code)
			}
		})
	})
}

// Test error handling for non-existent bookmarks
func TestBookmarkUpdate_ErrorHandling(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		t.Run("PATCH should handle non-existent bookmark", func(t *testing.T) {
			patchData := BookmarkUpdateRequest{Action: "working"}
			jsonData, _ := json.Marshal(patchData)
			
			req := httptest.NewRequest("PATCH", "/api/bookmarks/99999", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			
			handleBookmarkUpdate(rr, req)
			
			if rr.Code != http.StatusInternalServerError {
				t.Errorf("Expected status %d for non-existent bookmark, got %d", http.StatusInternalServerError, rr.Code)
			}
		})

		t.Run("PUT should handle non-existent bookmark", func(t *testing.T) {
			putData := BookmarkFullUpdateRequest{
				Title: "Test",
				URL:   "https://test.com",
			}
			jsonData, _ := json.Marshal(putData)
			
			req := httptest.NewRequest("PUT", "/api/bookmarks/99999", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			
			handleBookmarkUpdate(rr, req)
			
			if rr.Code != http.StatusInternalServerError {
				t.Errorf("Expected status %d for non-existent bookmark, got %d", http.StatusInternalServerError, rr.Code)
			}
		})

		t.Run("Should reject invalid bookmark ID", func(t *testing.T) {
			patchData := BookmarkUpdateRequest{Action: "working"}
			jsonData, _ := json.Marshal(patchData)
			
			req := httptest.NewRequest("PATCH", "/api/bookmarks/invalid-id", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			
			handleBookmarkUpdate(rr, req)
			
			if rr.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d for invalid bookmark ID, got %d", http.StatusBadRequest, rr.Code)
			}
		})

		t.Run("Should reject unsupported HTTP methods", func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/api/bookmarks/1", nil)
			rr := httptest.NewRecorder()
			
			handleBookmarkUpdate(rr, req)
			
			if rr.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status %d for unsupported method, got %d", http.StatusMethodNotAllowed, rr.Code)
			}
		})
	})
}

// Test that response format matches frontend expectations
func TestBookmarkUpdate_ResponseFormat(t *testing.T) {
	withTestDB(t, func(t *testing.T, tdb *TestDB) {
		// Insert a test bookmark
		insertSQL := `
		INSERT INTO bookmarks (url, title, description, action, topic, timestamp)
		VALUES (?, ?, ?, ?, ?, '2023-12-01 10:00:00')`
		
		result, err := tdb.db.Exec(insertSQL, 
			"https://format-test.example.com", "Format Test", "Test description", "read-later", "TestTopic")
		if err != nil {
			t.Fatalf("Failed to insert test bookmark: %v", err)
		}
		
		bookmarkID, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("Failed to get bookmark ID: %v", err)
		}

		t.Run("Response should include all expected fields", func(t *testing.T) {
			patchData := BookmarkUpdateRequest{Action: "working"}
			jsonData, _ := json.Marshal(patchData)
			
			req := httptest.NewRequest("PATCH", fmt.Sprintf("/api/bookmarks/%d", bookmarkID), bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			
			handleBookmarkUpdate(rr, req)
			
			if rr.Code != http.StatusOK {
				t.Fatalf("Request failed with status %d", rr.Code)
			}
			
			var response ProjectBookmark
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}
			
			// Check all expected fields are present and have correct types
			if response.ID == 0 {
				t.Error("Expected ID to be set")
			}
			if response.URL == "" {
				t.Error("Expected URL to be set")
			}
			if response.Title == "" {
				t.Error("Expected Title to be set")
			}
			if response.Timestamp == "" {
				t.Error("Expected Timestamp to be set")
			}
			if response.Domain == "" {
				t.Error("Expected Domain to be calculated")
			}
			if response.Age == "" {
				t.Error("Expected Age to be calculated")
			}
			
			// Verify domain calculation
			if response.Domain != "format-test.example.com" {
				t.Errorf("Expected domain 'format-test.example.com', got %s", response.Domain)
			}
			
			// Verify age calculation format
			validAgeFormats := []string{"just now", "1m", "1h", "1d", "1w", "1mo"}
			ageValid := false
			for _, format := range validAgeFormats {
				if strings.HasSuffix(response.Age, format[len(format)-1:]) || response.Age == "just now" {
					ageValid = true
					break
				}
			}
			if !ageValid {
				t.Errorf("Age format seems invalid: %s", response.Age)
			}
		})
	})
}