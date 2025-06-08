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
}

func main() {
	http.HandleFunc("/bookmark", handleBookmark)
	
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
		writer.Write([]string{"timestamp", "url", "title", "description"})
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	record := []string{timestamp, req.URL, req.Title, req.Description}
	
	return writer.Write(record)
}