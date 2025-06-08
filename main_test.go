package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// Test helper to create a temporary CSV file
func createTempCSV(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp("", "test_bookmarks_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	
	if content != "" {
		_, err = tmpFile.WriteString(content)
		if err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
	}
	
	tmpFile.Close()
	return tmpFile.Name()
}

// Test helper to clean up temp files
func cleanup(t *testing.T, filename string) {
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		t.Errorf("Failed to clean up temp file %s: %v", filename, err)
	}
}

// Test helper to read CSV content
func readCSV(t *testing.T, filename string) [][]string {
	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Failed to open CSV file: %v", err)
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}
	
	return records
}

// Testable version of writeToCSV that accepts filename
func writeToCSVWithFilename(req BookmarkRequest, filename string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	if stat.Size() == 0 {
		writer.Write([]string{"timestamp", "url", "title", "description", "content", "action", "shareTo", "topic"})
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	record := []string{timestamp, req.URL, req.Title, req.Description, req.Content, req.Action, req.ShareTo, req.Topic}
	
	return writer.Write(record)
}

// Testable version of getTopicsFromCSV that accepts filename
func getTopicsFromCSVWithFilename(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	topicSet := make(map[string]bool)
	
	// Skip header row if it exists
	startIndex := 0
	if len(records) > 0 && records[0][0] == "timestamp" {
		startIndex = 1
	}
	
	// Extract topics from records (topic is in column 7)
	for i := startIndex; i < len(records); i++ {
		if len(records[i]) > 7 && records[i][7] != "" {
			topicSet[records[i][7]] = true
		}
	}
	
	// Convert to sorted slice
	topics := make([]string, 0, len(topicSet))
	for topic := range topicSet {
		topics = append(topics, topic)
	}
	
	return topics, nil
}

func TestWriteToCSV_NewFile(t *testing.T) {
	tmpFile := createTempCSV(t, "")
	defer cleanup(t, tmpFile)
	
	req := BookmarkRequest{
		URL:         "https://example.com",
		Title:       "Test Title",
		Description: "Test Description",
		Content:     "Test Content",
		Action:      "read-later",
		ShareTo:     "",
		Topic:       "",
	}
	
	err := writeToCSVWithFilename(req, tmpFile)
	if err != nil {
		t.Fatalf("writeToCSVWithFilename failed: %v", err)
	}
	
	records := readCSV(t, tmpFile)
	
	if len(records) != 2 {
		t.Fatalf("Expected 2 records (header + data), got %d", len(records))
	}
	
	// Check header
	expectedHeader := []string{"timestamp", "url", "title", "description", "content", "action", "shareTo", "topic"}
	if len(records[0]) != len(expectedHeader) {
		t.Fatalf("Header length mismatch. Expected %d, got %d", len(expectedHeader), len(records[0]))
	}
	
	for i, expected := range expectedHeader {
		if records[0][i] != expected {
			t.Errorf("Header[%d]: expected %s, got %s", i, expected, records[0][i])
		}
	}
	
	// Check data row
	dataRow := records[1]
	if dataRow[1] != req.URL {
		t.Errorf("URL: expected %s, got %s", req.URL, dataRow[1])
	}
	if dataRow[2] != req.Title {
		t.Errorf("Title: expected %s, got %s", req.Title, dataRow[2])
	}
	if dataRow[3] != req.Description {
		t.Errorf("Description: expected %s, got %s", req.Description, dataRow[3])
	}
	if dataRow[4] != req.Content {
		t.Errorf("Content: expected %s, got %s", req.Content, dataRow[4])
	}
	if dataRow[5] != req.Action {
		t.Errorf("Action: expected %s, got %s", req.Action, dataRow[5])
	}
}

func TestWriteToCSV_ExistingFile(t *testing.T) {
	existingContent := "timestamp,url,title,description,content,action,shareTo,topic\n2023-01-01 12:00:00,https://old.com,Old Title,Old Desc,Old Content,share,John,\n"
	tmpFile := createTempCSV(t, existingContent)
	defer cleanup(t, tmpFile)
	
	req := BookmarkRequest{
		URL:         "https://new.com",
		Title:       "New Title",
		Description: "New Description",
		Content:     "New Content",
		Action:      "working",
		ShareTo:     "",
		Topic:       "Development",
	}
	
	err := writeToCSVWithFilename(req, tmpFile)
	if err != nil {
		t.Fatalf("writeToCSVWithFilename failed: %v", err)
	}
	
	records := readCSV(t, tmpFile)
	
	if len(records) != 3 {
		t.Fatalf("Expected 3 records (header + 2 data rows), got %d", len(records))
	}
	
	// Check new data row
	newRow := records[2]
	if newRow[1] != req.URL {
		t.Errorf("URL: expected %s, got %s", req.URL, newRow[1])
	}
	if newRow[7] != req.Topic {
		t.Errorf("Topic: expected %s, got %s", req.Topic, newRow[7])
	}
}

func TestGetTopicsFromCSV_EmptyFile(t *testing.T) {
	tmpFile := createTempCSV(t, "")
	defer cleanup(t, tmpFile)
	
	topics, err := getTopicsFromCSVWithFilename(tmpFile)
	if err != nil {
		t.Fatalf("getTopicsFromCSVWithFilename failed: %v", err)
	}
	
	if len(topics) != 0 {
		t.Errorf("Expected 0 topics, got %d", len(topics))
	}
}

func TestGetTopicsFromCSV_WithData(t *testing.T) {
	csvContent := `timestamp,url,title,description,content,action,shareTo,topic
2023-01-01 12:00:00,https://example1.com,Title 1,Desc 1,Content 1,working,,Programming
2023-01-02 12:00:00,https://example2.com,Title 2,Desc 2,Content 2,working,,Development
2023-01-03 12:00:00,https://example3.com,Title 3,Desc 3,Content 3,working,,Programming
2023-01-04 12:00:00,https://example4.com,Title 4,Desc 4,Content 4,share,John,
2023-01-05 12:00:00,https://example5.com,Title 5,Desc 5,Content 5,read-later,,
2023-01-06 12:00:00,https://example6.com,Title 6,Desc 6,Content 6,working,,Testing
`
	tmpFile := createTempCSV(t, csvContent)
	defer cleanup(t, tmpFile)
	
	topics, err := getTopicsFromCSVWithFilename(tmpFile)
	if err != nil {
		t.Fatalf("getTopicsFromCSVWithFilename failed: %v", err)
	}
	
	expectedTopics := map[string]bool{
		"Programming":  true,
		"Development":  true,
		"Testing":      true,
	}
	
	if len(topics) != len(expectedTopics) {
		t.Errorf("Expected %d topics, got %d", len(expectedTopics), len(topics))
	}
	
	for _, topic := range topics {
		if !expectedTopics[topic] {
			t.Errorf("Unexpected topic: %s", topic)
		}
	}
}

func TestGetTopicsFromCSV_NonExistentFile(t *testing.T) {
	topics, err := getTopicsFromCSVWithFilename("/nonexistent/file.csv")
	if err != nil {
		t.Fatalf("getTopicsFromCSVWithFilename should handle non-existent file gracefully: %v", err)
	}
	
	if len(topics) != 0 {
		t.Errorf("Expected 0 topics for non-existent file, got %d", len(topics))
	}
}

func TestHandleBookmark_Success(t *testing.T) {
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
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
	
	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	if response["status"] != "success" {
		t.Errorf("Expected status 'success', got %s", response["status"])
	}
}

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

func TestHandleTopics_Success(t *testing.T) {
	// Create a test CSV file with topics
	csvContent := `timestamp,url,title,description,content,action,shareTo,topic
2023-01-01 12:00:00,https://example1.com,Title 1,Desc 1,Content 1,working,,Programming
2023-01-02 12:00:00,https://example2.com,Title 2,Desc 2,Content 2,working,,Development
`
	
	// Create test_bookmarks.csv in current directory for the test (different name)
	testFileName := "test_bookmarks.csv"
	err := os.WriteFile(testFileName, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}
	defer os.Remove(testFileName)
	
	// Temporarily backup the original CSV filename and use test file
	originalFile := "bookmarks.csv"
	if _, err := os.Stat(originalFile); err == nil {
		// Backup exists, rename it temporarily
		backupFile := "bookmarks_backup_test.csv"
		err := os.Rename(originalFile, backupFile)
		if err != nil {
			t.Fatalf("Failed to backup original CSV: %v", err)
		}
		defer func() {
			os.Remove(originalFile) // Remove any test file
			os.Rename(backupFile, originalFile) // Restore backup
		}()
	}
	
	// Copy test file to expected location
	err = os.Rename(testFileName, originalFile)
	if err != nil {
		t.Fatalf("Failed to setup test CSV: %v", err)
	}
	
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
}

func TestHandleTopics_InvalidMethod(t *testing.T) {
	req := httptest.NewRequest("POST", "/topics", nil)
	rr := httptest.NewRecorder()
	
	handleTopics(rr, req)
	
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}