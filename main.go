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
	http.HandleFunc("/bookmark", handleBookmark)
	http.HandleFunc("/topics", handleTopics)
	
	fmt.Println("Server starting on :9090")
	log.Fatal(http.ListenAndServe(":9090", nil))
}

func handleBookmark(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BookmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.URL == "" || req.Title == "" {
		http.Error(w, "URL and title are required", http.StatusBadRequest)
		return
	}

	if err := writeToCSV(req); err != nil {
		http.Error(w, "Failed to save bookmark", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func handleTopics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	topics, err := getTopicsFromCSV()
	if err != nil {
		http.Error(w, "Failed to get topics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"topics": topics})
}

func writeToCSV(req BookmarkRequest) error {
	filename := "bookmarks.csv"
	
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

func getTopicsFromCSV() ([]string, error) {
	filename := "bookmarks.csv"
	
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