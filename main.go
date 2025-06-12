package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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

type ProjectStat struct {
	Topic       string `json:"topic"`
	Count       int    `json:"count"`
	LastUpdated string `json:"lastUpdated"`
	Status      string `json:"status"`
}

type SummaryStats struct {
	NeedsTriage     int           `json:"needsTriage"`
	ActiveProjects  int           `json:"activeProjects"`
	ReadyToShare    int           `json:"readyToShare"`
	TotalBookmarks  int           `json:"totalBookmarks"`
	ProjectStats    []ProjectStat `json:"projectStats"`
}

var db *sql.DB
var logFile *os.File

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Component string `json:"component"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

func initLogging() error {
	var err error
	logFile, err = os.OpenFile("bookminderapi.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	
	log.Printf("Structured logging initialized: bookminderapi.log")
	logStructured("INFO", "system", "Logging system initialized", nil)
	return nil
}

func logStructured(level, component, message string, data map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		Component: component,
		Data:      data,
	}
	
	jsonData, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Failed to marshal log entry: %v", err)
		return
	}
	
	logFile.WriteString(string(jsonData) + "\n")
}

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
	
	// Initialize logging
	if err := initLogging(); err != nil {
		log.Fatalf("Failed to initialize logging: %v", err)
	}
	defer logFile.Close()
	
	logStructured("INFO", "startup", "BookMinder API starting up", nil)
	
	// Initialize database
	if err := initDatabase(); err != nil {
		logStructured("ERROR", "database", "Failed to initialize database", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	
	log.Printf("Registering HTTP handlers")
	logStructured("INFO", "startup", "Registering HTTP handlers", nil)
	
	http.HandleFunc("/bookmark", handleBookmark)
	http.HandleFunc("/topics", handleTopics)
	http.HandleFunc("/api/stats/summary", handleStatsSummary)
	
	log.Printf("Available endpoints:")
	log.Printf("  POST /bookmark - Save a new bookmark")
	log.Printf("  GET /topics - Get list of available topics")
	log.Printf("  GET /api/stats/summary - Get dashboard summary statistics")
	
	port := ":9090"
	log.Printf("Starting server on port %s", port)
	fmt.Printf("BookMinder API server starting on %s\n", port)
	
	logStructured("INFO", "startup", "Server starting", map[string]interface{}{
		"port": port,
		"endpoints": []string{"/bookmark", "/topics", "/api/stats/summary"},
	})
	
	if err := http.ListenAndServe(port, nil); err != nil {
		logStructured("ERROR", "server", "Server failed to start", map[string]interface{}{
			"error": err.Error(),
			"port": port,
		})
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleBookmark(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to /bookmark from %s", r.Method, r.RemoteAddr)
	
	logStructured("INFO", "api", "Bookmark request received", map[string]interface{}{
		"method": r.Method,
		"remote_addr": r.RemoteAddr,
		"user_agent": r.UserAgent(),
	})
	
	if r.Method != http.MethodPost {
		log.Printf("Method not allowed: %s (expected POST)", r.Method)
		logStructured("WARN", "api", "Method not allowed", map[string]interface{}{
			"method": r.Method,
			"expected": "POST",
		})
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BookmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode JSON request: %v", err)
		logStructured("ERROR", "api", "JSON decode failed", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Parsed bookmark request: URL=%s, Title=%s, Action=%s, Topic=%s", 
		req.URL, req.Title, req.Action, req.Topic)

	logStructured("INFO", "api", "Bookmark request parsed", map[string]interface{}{
		"url": req.URL,
		"title": req.Title,
		"action": req.Action,
		"topic": req.Topic,
		"has_content": len(req.Content) > 0,
	})

	if req.URL == "" || req.Title == "" {
		log.Printf("Validation failed: missing required fields (URL=%s, Title=%s)", req.URL, req.Title)
		logStructured("WARN", "api", "Validation failed", map[string]interface{}{
			"missing_url": req.URL == "",
			"missing_title": req.Title == "",
		})
		http.Error(w, "URL and title are required", http.StatusBadRequest)
		return
	}

	if err := saveBookmarkToDB(req); err != nil {
		log.Printf("Failed to save bookmark to database: %v", err)
		logStructured("ERROR", "database", "Failed to save bookmark", map[string]interface{}{
			"error": err.Error(),
			"url": req.URL,
		})
		http.Error(w, "Failed to save bookmark", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully saved bookmark: %s", req.URL)
	logStructured("INFO", "database", "Bookmark saved successfully", map[string]interface{}{
		"url": req.URL,
		"title": req.Title,
		"action": req.Action,
	})
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func handleTopics(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to /topics from %s", r.Method, r.RemoteAddr)
	
	logStructured("INFO", "api", "Topics request received", map[string]interface{}{
		"method": r.Method,
		"remote_addr": r.RemoteAddr,
	})
	
	if r.Method != http.MethodGet {
		log.Printf("Method not allowed: %s (expected GET)", r.Method)
		logStructured("WARN", "api", "Method not allowed", map[string]interface{}{
			"method": r.Method,
			"expected": "GET",
		})
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	topics, err := getTopicsFromDB()
	if err != nil {
		log.Printf("Failed to get topics from database: %v", err)
		logStructured("ERROR", "database", "Failed to get topics", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Failed to get topics", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully retrieved %d topics", len(topics))
	logStructured("INFO", "database", "Topics retrieved successfully", map[string]interface{}{
		"count": len(topics),
		"topics": topics,
	})
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"topics": topics})
}

func saveBookmarkToDB(req BookmarkRequest) error {
	log.Printf("Saving bookmark to database: %s", req.URL)
	
	logStructured("INFO", "database", "Saving bookmark", map[string]interface{}{
		"url": req.URL,
		"title": req.Title,
		"action": req.Action,
		"content_length": len(req.Content),
	})
	
	insertSQL := `
	INSERT INTO bookmarks (url, title, description, content, action, shareTo, topic)
	VALUES (?, ?, ?, ?, ?, ?, ?)`
	
	result, err := db.Exec(insertSQL, req.URL, req.Title, req.Description, req.Content, req.Action, req.ShareTo, req.Topic)
	if err != nil {
		log.Printf("Failed to insert bookmark: %v", err)
		logStructured("ERROR", "database", "Insert failed", map[string]interface{}{
			"error": err.Error(),
			"url": req.URL,
		})
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Failed to get last insert ID: %v", err)
		logStructured("WARN", "database", "Failed to get insert ID", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}
	
	log.Printf("Successfully saved bookmark with ID: %d", id)
	logStructured("INFO", "database", "Bookmark saved", map[string]interface{}{
		"id": id,
		"url": req.URL,
		"title": req.Title,
	})
	
	return nil
}

func getTopicsFromDB() ([]string, error) {
	log.Printf("Reading topics from database")
	
	logStructured("INFO", "database", "Querying topics", nil)
	
	querySQL := `SELECT DISTINCT topic FROM bookmarks WHERE topic IS NOT NULL AND topic != '' ORDER BY topic`
	
	rows, err := db.Query(querySQL)
	if err != nil {
		log.Printf("Failed to query topics: %v", err)
		logStructured("ERROR", "database", "Topics query failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}
	defer rows.Close()
	
	var topics []string
	for rows.Next() {
		var topic string
		if err := rows.Scan(&topic); err != nil {
			log.Printf("Failed to scan topic: %v", err)
			logStructured("ERROR", "database", "Topic scan failed", map[string]interface{}{
				"error": err.Error(),
			})
			return nil, err
		}
		topics = append(topics, topic)
	}
	
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating topics: %v", err)
		logStructured("ERROR", "database", "Topics iteration failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}
	
	log.Printf("Found %d unique topics", len(topics))
	log.Printf("Returning topics: %v", topics)
	logStructured("INFO", "database", "Topics query completed", map[string]interface{}{
		"count": len(topics),
		"topics": topics,
	})
	
	return topics, nil
}

func handleStatsSummary(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to /api/stats/summary from %s", r.Method, r.RemoteAddr)
	
	logStructured("INFO", "api", "Stats summary request received", map[string]interface{}{
		"method": r.Method,
		"remote_addr": r.RemoteAddr,
	})
	
	if r.Method != http.MethodGet {
		log.Printf("Method not allowed: %s (expected GET)", r.Method)
		logStructured("WARN", "api", "Method not allowed", map[string]interface{}{
			"method": r.Method,
			"expected": "GET",
		})
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := getStatsSummary()
	if err != nil {
		log.Printf("Failed to get stats summary: %v", err)
		logStructured("ERROR", "database", "Failed to get stats summary", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Failed to get stats summary", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully retrieved stats summary")
	logStructured("INFO", "database", "Stats summary retrieved", map[string]interface{}{
		"totalBookmarks": stats.TotalBookmarks,
		"needsTriage": stats.NeedsTriage,
		"activeProjects": stats.ActiveProjects,
		"readyToShare": stats.ReadyToShare,
	})
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func getStatsSummary() (*SummaryStats, error) {
	logStructured("INFO", "database", "Computing stats summary", nil)
	
	stats := &SummaryStats{}
	
	// Get total bookmarks count
	err := db.QueryRow("SELECT COUNT(*) FROM bookmarks").Scan(&stats.TotalBookmarks)
	if err != nil {
		return nil, fmt.Errorf("failed to count total bookmarks: %v", err)
	}
	
	// Count by action categories
	// needsTriage: bookmarks with no action or action = "read-later"
	err = db.QueryRow(`
		SELECT COUNT(*) FROM bookmarks 
		WHERE action IS NULL OR action = '' OR action = 'read-later'
	`).Scan(&stats.NeedsTriage)
	if err != nil {
		return nil, fmt.Errorf("failed to count needs triage: %v", err)
	}
	
	// activeProjects: unique topics in "working" action
	err = db.QueryRow(`
		SELECT COUNT(DISTINCT topic) FROM bookmarks 
		WHERE action = 'working' AND topic IS NOT NULL AND topic != ''
	`).Scan(&stats.ActiveProjects)
	if err != nil {
		return nil, fmt.Errorf("failed to count active projects: %v", err)
	}
	
	// readyToShare: bookmarks with action = "share"
	err = db.QueryRow(`
		SELECT COUNT(*) FROM bookmarks 
		WHERE action = 'share'
	`).Scan(&stats.ReadyToShare)
	if err != nil {
		return nil, fmt.Errorf("failed to count ready to share: %v", err)
	}
	
	// Get project stats for working topics
	projectStats, err := getProjectStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get project stats: %v", err)
	}
	stats.ProjectStats = projectStats
	
	logStructured("INFO", "database", "Stats summary computed", map[string]interface{}{
		"totalBookmarks": stats.TotalBookmarks,
		"needsTriage": stats.NeedsTriage,
		"activeProjects": stats.ActiveProjects,
		"readyToShare": stats.ReadyToShare,
		"projectCount": len(stats.ProjectStats),
	})
	
	return stats, nil
}

func getProjectStats() ([]ProjectStat, error) {
	querySQL := `
		SELECT 
			topic,
			COUNT(*) as count,
			MAX(timestamp) as lastUpdated
		FROM bookmarks 
		WHERE action = 'working' AND topic IS NOT NULL AND topic != ''
		GROUP BY topic
		ORDER BY MAX(timestamp) DESC
		LIMIT 10
	`
	
	rows, err := db.Query(querySQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query project stats: %v", err)
	}
	defer rows.Close()
	
	var projects []ProjectStat
	for rows.Next() {
		var project ProjectStat
		var lastUpdated string
		
		err := rows.Scan(&project.Topic, &project.Count, &lastUpdated)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project stat: %v", err)
		}
		
		// Parse timestamp and format as ISO 8601
		if timestamp, err := time.Parse("2006-01-02 15:04:05", lastUpdated); err == nil {
			project.LastUpdated = timestamp.UTC().Format(time.RFC3339)
		} else {
			project.LastUpdated = lastUpdated
		}
		
		// Determine status based on recency
		if timestamp, err := time.Parse(time.RFC3339, project.LastUpdated); err == nil {
			daysSince := time.Since(timestamp).Hours() / 24
			if daysSince <= 7 {
				project.Status = "active"
			} else if daysSince <= 30 {
				project.Status = "stale"
			} else {
				project.Status = "inactive"
			}
		} else {
			project.Status = "unknown"
		}
		
		projects = append(projects, project)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating project stats: %v", err)
	}
	
	return projects, nil
}