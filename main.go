package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
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

var db *sql.DB

func initDatabase() error {
	var err error
	db, err = sql.Open("sqlite3", "bookmarks.db")
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Test the connection
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
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
		return fmt.Errorf("failed to create bookmarks table: %v", err)
	}

	log.Printf("Database initialized successfully")
	return nil
}

func main() {
	log.Printf("BookMinder API starting up...")
	
	// Initialize database
	if err := initDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	
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

	if err := saveBookmarkToDB(req); err != nil {
		log.Printf("Failed to save bookmark to database: %v", err)
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

	topics, err := getTopicsFromDB()
	if err != nil {
		log.Printf("Failed to get topics from database: %v", err)
		http.Error(w, "Failed to get topics", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully retrieved %d topics", len(topics))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"topics": topics})
}

func saveBookmarkToDB(req BookmarkRequest) error {
	log.Printf("Saving bookmark to database: %s", req.URL)
	
	insertSQL := `
	INSERT INTO bookmarks (url, title, description, content, action, shareTo, topic)
	VALUES (?, ?, ?, ?, ?, ?, ?)`
	
	result, err := db.Exec(insertSQL, req.URL, req.Title, req.Description, req.Content, req.Action, req.ShareTo, req.Topic)
	if err != nil {
		log.Printf("Failed to insert bookmark: %v", err)
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Failed to get last insert ID: %v", err)
		return err
	}
	
	log.Printf("Successfully saved bookmark with ID: %d", id)
	return nil
}

func getTopicsFromDB() ([]string, error) {
	log.Printf("Reading topics from database")
	
	querySQL := `SELECT DISTINCT topic FROM bookmarks WHERE topic IS NOT NULL AND topic != '' ORDER BY topic`
	
	rows, err := db.Query(querySQL)
	if err != nil {
		log.Printf("Failed to query topics: %v", err)
		return nil, err
	}
	defer rows.Close()
	
	var topics []string
	for rows.Next() {
		var topic string
		if err := rows.Scan(&topic); err != nil {
			log.Printf("Failed to scan topic: %v", err)
			return nil, err
		}
		topics = append(topics, topic)
	}
	
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating topics: %v", err)
		return nil, err
	}
	
	log.Printf("Found %d unique topics", len(topics))
	log.Printf("Returning topics: %v", topics)
	return topics, nil
}