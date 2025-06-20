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
		project_id INTEGER REFERENCES projects(id)
	);`
	
	if _, err = db.Exec(createBookmarksTableSQL); err != nil {
		t.Fatalf("Failed to create test bookmarks table: %v", err)
	}
	
	return &TestDB{db: db, dbPath: dbPath}
}

// cleanup closes the test database and removes the file
func (tdb *TestDB) cleanup(t *testing.T) {
	if err := tdb.db.Close(); err != nil {
		t.Errorf("Failed to close test database: %v", err)
	}
}

// insertTestBookmarks adds sample data to the test database
func (tdb *TestDB) insertTestBookmarks(t *testing.T) {
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
		
		var response map[string]string
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		
		if response["status"] != "success" {
			t.Errorf("Expected status 'success', got %s", response["status"])
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
		insertSQL := `INSERT INTO bookmarks (url, title, action, topic, timestamp) VALUES (?, ?, ?, ?, ?)`
		
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
		insertSQL := `INSERT INTO bookmarks (url, title, action, topic, timestamp) VALUES (?, ?, ?, ?, ?)`
		
		// Create topics with different link counts
		testCases := []struct {
			topic         string
			linkCount     int
		}{
			{"SmallProject", 1},
			{"MediumProject", 5},
			{"LargeProject", 15},
		}
		
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
		insertSQL := `INSERT INTO bookmarks (url, title, action, topic, timestamp) VALUES (?, ?, ?, ?, ?)`
		
		// Create projects with different timestamps
		testData := []struct {
			topic     string
			timestamp string
		}{
			{"OldestProject", "2023-11-01 10:00:00"},
			{"MiddleProject", "2023-11-15 10:00:00"},
			{"NewestProject", "2023-12-01 10:00:00"},
		}
		
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
		insertSQL := `INSERT INTO bookmarks (url, title, action, topic, timestamp) VALUES (?, ?, ?, ?, ?)`
		
		// Test case sensitivity and special characters
		topics := []string{
			"JavaScript",
			"javascript", 
			"Java-Script",
			"Java_Script",
			"Java Script",
			"JAVASCRIPT",
		}
		
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
		insertSQL := `INSERT INTO bookmarks (url, title, action, topic, timestamp) VALUES (?, ?, ?, ?, ?)`
		
		testData := []struct {
			url, title, action, topic string
		}{
			{"https://active1.com", "Active 1", "working", "ActiveTopic1"},
			{"https://active2.com", "Active 2", "working", "ActiveTopic2"}, 
			{"https://ref1.com", "Ref 1", "read-later", "RefTopic1"},
			{"https://ref2.com", "Ref 2", "share", "RefTopic2"},
		}
		
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
		insertSQL := `INSERT INTO bookmarks (url, title, action, topic, timestamp) VALUES (?, ?, ?, ?, ?)`
		
		// Add working project
		_, err := tdb.db.Exec(insertSQL, "https://work.com", "Work Item", "working", "WorkProject", "2023-12-01 10:00:00")
		if err != nil {
			t.Fatalf("Failed to insert working bookmark: %v", err)
		}
		
		// Add reference bookmark
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