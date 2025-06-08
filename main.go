package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type BookmarkRequest struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Content     string `json:"content,omitempty"`
	Action      string `json:"action,omitempty"`
	ShareTo     string `json:"shareTo,omitempty"`
	Topic       string `json:"topic,omitempty"`
}

func main() {
	log.Printf("BookMinder API starting up...")
	log.Printf("Registering HTTP handlers")
	
	http.HandleFunc("/bookmark", handleBookmark)
	http.HandleFunc("/topics", handleTopics)
	
	log.Printf("Available endpoints:")
	log.Printf("  POST /bookmark - Save a new bookmark")
	log.Printf("  GET /topics - Get list of available topics")
	
	port := ":9090"
	log.Printf("Starting server on port %s", port)
	fmt.Printf("BookMinder API server starting on %s\n", port)
	
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleBookmark(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to /bookmark from %s", r.Method, r.RemoteAddr)
	
	if r.Method != http.MethodPost {
		log.Printf("Method not allowed: %s (expected POST)", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BookmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode JSON request: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Parsed bookmark request: URL=%s, Title=%s, Action=%s, Topic=%s", 
		req.URL, req.Title, req.Action, req.Topic)

	if req.URL == "" || req.Title == "" {
		log.Printf("Validation failed: missing required fields (URL=%s, Title=%s)", req.URL, req.Title)
		http.Error(w, "URL and title are required", http.StatusBadRequest)
		return
	}

	if err := writeToCSV(req); err != nil {
		log.Printf("Failed to write bookmark to CSV: %v", err)
		http.Error(w, "Failed to save bookmark", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully saved bookmark: %s", req.URL)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func handleTopics(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to /topics from %s", r.Method, r.RemoteAddr)
	
	if r.Method != http.MethodGet {
		log.Printf("Method not allowed: %s (expected GET)", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	topics, err := getTopicsFromCSV()
	if err != nil {
		log.Printf("Failed to get topics from CSV: %v", err)
		http.Error(w, "Failed to get topics", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully retrieved %d topics", len(topics))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"topics": topics})
}

func writeToCSV(req BookmarkRequest) error {
	filename := "bookmarks.csv"
	
	log.Printf("Opening CSV file: %s", filename)
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open CSV file %s: %v", filename, err)
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	stat, err := file.Stat()
	if err != nil {
		log.Printf("Failed to stat CSV file %s: %v", filename, err)
		return err
	}

	if stat.Size() == 0 {
		log.Printf("Empty CSV file detected, writing header")
		if err := writer.Write([]string{"timestamp", "url", "title", "description", "content", "action", "shareTo", "topic"}); err != nil {
			log.Printf("Failed to write CSV header: %v", err)
			return err
		}
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	record := []string{timestamp, req.URL, req.Title, req.Description, req.Content, req.Action, req.ShareTo, req.Topic}
	
	log.Printf("Writing record to CSV: %v", record)
	if err := writer.Write(record); err != nil {
		log.Printf("Failed to write record to CSV: %v", err)
		return err
	}
	
	log.Printf("Successfully wrote record to CSV")
	return nil
}

func getTopicsFromCSV() ([]string, error) {
	filename := "bookmarks.csv"
	
	log.Printf("Reading topics from CSV file: %s", filename)
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("CSV file %s does not exist, returning empty topics", filename)
			return []string{}, nil
		}
		log.Printf("Failed to open CSV file %s: %v", filename, err)
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Printf("Failed to read CSV records from %s: %v", filename, err)
		return nil, err
	}

	log.Printf("Read %d records from CSV", len(records))
	topicSet := make(map[string]bool)
	
	// Skip header row if it exists
	startIndex := 0
	if len(records) > 0 && records[0][0] == "timestamp" {
		log.Printf("Header row detected, skipping first record")
		startIndex = 1
	}
	
	// Extract topics from records (topic is in column 7)
	topicsFound := 0
	for i := startIndex; i < len(records); i++ {
		if len(records[i]) > 7 && records[i][7] != "" {
			topicSet[records[i][7]] = true
			topicsFound++
		}
	}
	
	log.Printf("Found %d topic entries, %d unique topics", topicsFound, len(topicSet))
	
	// Convert to sorted slice
	topics := make([]string, 0, len(topicSet))
	for topic := range topicSet {
		topics = append(topics, topic)
	}
	
	log.Printf("Returning topics: %v", topics)
	return topics, nil
}