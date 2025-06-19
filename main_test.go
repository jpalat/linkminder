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
	
	// Create the bookmarks table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS bookmarks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		url TEXT NOT NULL,
		title TEXT NOT NULL,
		description TEXT,
		content TEXT,
		action TEXT,
		shareTo TEXT,
		topic TEXT
	);`
	
	if _, err = db.Exec(createTableSQL); err != nil {
		t.Fatalf("Failed to create test table: %v", err)
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