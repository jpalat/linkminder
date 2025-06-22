package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

type Project struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status"`
	LinkCount   int    `json:"linkCount"`
	LastUpdated string `json:"lastUpdated"`
	CreatedAt   string `json:"createdAt"`
}

type BookmarkRequest struct {
	URL              string            `json:"url"`
	Title            string            `json:"title"`
	Description      string            `json:"description,omitempty"`
	Content          string            `json:"content,omitempty"`
	Action           string            `json:"action,omitempty"`
	ShareTo          string            `json:"shareTo,omitempty"`
	Topic            string            `json:"topic,omitempty"`     // Legacy support
	ProjectID        int               `json:"projectId,omitempty"` // New field
	Tags             []string          `json:"tags,omitempty"`
	CustomProperties map[string]string `json:"customProperties,omitempty"`
}

type BookmarkUpdateRequest struct {
	Action           string            `json:"action,omitempty"`
	ShareTo          string            `json:"shareTo,omitempty"`
	Topic            string            `json:"topic,omitempty"`     // Legacy support
	ProjectID        int               `json:"projectId,omitempty"` // New field
	Tags             []string          `json:"tags,omitempty"`
	CustomProperties map[string]string `json:"customProperties,omitempty"`
}

type BookmarkFullUpdateRequest struct {
	Title            string            `json:"title"`
	URL              string            `json:"url"`
	Description      string            `json:"description,omitempty"`
	Action           string            `json:"action,omitempty"`
	ShareTo          string            `json:"shareTo,omitempty"`
	Topic            string            `json:"topic,omitempty"`
	Tags             []string          `json:"tags,omitempty"`
	CustomProperties map[string]string `json:"customProperties,omitempty"`
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
	Archived        int           `json:"archived"`
	TotalBookmarks  int           `json:"totalBookmarks"`
	ProjectStats    []ProjectStat `json:"projectStats"`
}

type TriageBookmark struct {
	ID               int               `json:"id"`
	URL              string            `json:"url"`
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	Timestamp        string            `json:"timestamp"`
	Domain           string            `json:"domain"`
	Age              string            `json:"age"`
	Suggested        string            `json:"suggested"`
	Topic            string            `json:"topic"`
	Action           string            `json:"action,omitempty"`
	ShareTo          string            `json:"shareTo,omitempty"`
	Tags             []string          `json:"tags,omitempty"`
	CustomProperties map[string]string `json:"customProperties,omitempty"`
}

type TriageResponse struct {
	Bookmarks []TriageBookmark `json:"bookmarks"`
	Total     int              `json:"total"`
	Limit     int              `json:"limit"`
	Offset    int              `json:"offset"`
}

type ActiveProject struct {
	ID          int    `json:"id"`
	Topic       string `json:"topic"`
	LinkCount   int    `json:"linkCount"`
	LastUpdated string `json:"lastUpdated"`
	Status      string `json:"status"`
}

type ReferenceCollection struct {
	Topic        string `json:"topic"`
	LinkCount    int    `json:"linkCount"`
	LastAccessed string `json:"lastAccessed"`
}

type ProjectsResponse struct {
	ActiveProjects       []ActiveProject       `json:"activeProjects"`
	ReferenceCollections []ReferenceCollection `json:"referenceCollections"`
}

type ProjectBookmark struct {
	ID               int               `json:"id"`
	URL              string            `json:"url"`
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	Content          string            `json:"content"`
	Timestamp        string            `json:"timestamp"`
	Domain           string            `json:"domain"`
	Age              string            `json:"age"`
	Action           string            `json:"action"`
	Topic            string            `json:"topic"`
	ShareTo          string            `json:"shareTo"`
	Tags             []string          `json:"tags,omitempty"`
	CustomProperties map[string]string `json:"customProperties,omitempty"`
}

type ProjectDetailResponse struct {
	Topic       string            `json:"topic"`
	LinkCount   int               `json:"linkCount"`
	LastUpdated string            `json:"lastUpdated"`
	Status      string            `json:"status"`
	Bookmarks   []ProjectBookmark `json:"bookmarks"`
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
	db, err = sql.Open("sqlite3", "bookmarks.db?_busy_timeout=10000&_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Configure connection pool for better concurrent handling
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// Run migrations
	if err = runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	// Validate connection after migrations
	if err = db.Ping(); err != nil {
		return fmt.Errorf("database connection lost after migrations: %v", err)
	}

	log.Printf("Database initialized successfully")
	return nil
}

func runMigrations() error {
	// Create migration driver
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %v", err)
	}

	// Create migration instance
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"sqlite3",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %v", err)
	}
	// Don't defer close here as it may close the underlying database connection

	// Run migrations
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	if err == migrate.ErrNoChange {
		log.Printf("No new migrations to apply")
	} else {
		log.Printf("Migrations applied successfully")
	}

	// Log current migration version
	version, dirty, err := m.Version()
	if err != nil {
		log.Printf("Could not get migration version: %v", err)
	} else {
		log.Printf("Current migration version: %d (dirty: %t)", version, dirty)
		logStructured("INFO", "database", "Migration status", map[string]interface{}{
			"version": version,
			"dirty":   dirty,
		})
	}

	return nil
}

func validateDB() error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}
	if err := db.Ping(); err != nil {
		return fmt.Errorf("database connection lost: %v", err)
	}
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
	
	http.HandleFunc("/", withCORS(handleDashboard))
	http.HandleFunc("/projects", withCORS(handleProjectsPage))
	http.HandleFunc("/project-detail", withCORS(handleProjectDetailPage))
	http.HandleFunc("/bookmark", withCORS(handleBookmark))
	http.HandleFunc("/topics", withCORS(handleTopics))
	http.HandleFunc("/api/stats/summary", withCORS(handleStatsSummary))
	http.HandleFunc("/api/bookmarks/triage", withCORS(handleTriageQueue))
	http.HandleFunc("/api/bookmarks", withCORS(handleBookmarks))
	http.HandleFunc("/api/projects", withCORS(handleProjects))
	http.HandleFunc("/api/projects/", withCORS(handleProjectDetail))
	http.HandleFunc("/api/projects/id/", withCORS(handleProjectByID))
	http.HandleFunc("/api/bookmarks/", withCORS(handleBookmarkUpdate))
	
	log.Printf("Available endpoints:")
	log.Printf("  GET / - Dashboard interface")
	log.Printf("  GET /projects - Projects page interface")
	log.Printf("  GET /project-detail - Enhanced project detail page with filtering")
	log.Printf("  POST /bookmark - Save a new bookmark")
	log.Printf("  GET /topics - Get list of available topics")
	log.Printf("  GET /api/stats/summary - Get dashboard summary statistics")
	log.Printf("  GET /api/bookmarks/triage - Get bookmarks needing triage")
	log.Printf("  GET /api/bookmarks?action={action} - Get bookmarks by action type")
	log.Printf("  GET /api/projects - Get active projects and reference collections")
	log.Printf("  GET /api/projects/{topic} - Get detailed view of a specific project")
	log.Printf("  GET /api/projects/id/{id} - Get detailed view of a project by ID")
	log.Printf("  PATCH /api/bookmarks/{id} - Update a bookmark (partial)")
	log.Printf("  PUT /api/bookmarks/{id} - Update a bookmark (full)")
	
	port := ":9090"
	log.Printf("Starting server on port %s", port)
	fmt.Printf("BookMinder API server starting on %s\n", port)
	
	logStructured("INFO", "startup", "Server starting", map[string]interface{}{
		"port": port,
		"endpoints": []string{"/", "/projects", "/bookmark", "/topics", "/api/stats/summary", "/api/bookmarks/triage", "/api/projects", "/api/projects/{topic}", "/api/projects/id/{id}", "/api/bookmarks/{id}"},
	})
	
	if err := http.ListenAndServe(port, nil); err != nil {
		logStructured("ERROR", "server", "Server failed to start", map[string]interface{}{
			"error": err.Error(),
			"port": port,
		})
		log.Fatalf("Server failed to start: %v", err)
	}
}

// CORSMiddleware adds CORS headers to all responses
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight OPTIONS requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	}
}

// Helper function to wrap handlers with CORS
func withCORS(handler http.HandlerFunc) http.HandlerFunc {
	return corsMiddleware(handler)
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to / from %s", r.Method, r.RemoteAddr)
	
	logStructured("INFO", "api", "Dashboard request received", map[string]interface{}{
		"method": r.Method,
		"remote_addr": r.RemoteAddr,
		"user_agent": r.UserAgent(),
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

	// Read the dashboard HTML file
	dashboardHTML, err := os.ReadFile("dashboard.html")
	if err != nil {
		log.Printf("Failed to read dashboard.html: %v", err)
		logStructured("ERROR", "api", "Failed to read dashboard file", map[string]interface{}{
			"error": err.Error(),
		})
		if os.IsNotExist(err) {
			http.Error(w, "Dashboard not found", http.StatusNotFound)
		} else {
			http.Error(w, "Dashboard not available", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(dashboardHTML)
	
	logStructured("INFO", "api", "Dashboard served successfully", nil)
}

func handleProjectsPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to /projects from %s", r.Method, r.RemoteAddr)
	
	logStructured("INFO", "api", "Projects page request received", map[string]interface{}{
		"method": r.Method,
		"remote_addr": r.RemoteAddr,
		"user_agent": r.UserAgent(),
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

	// Read the projects HTML file
	projectsHTML, err := os.ReadFile("projects.html")
	if err != nil {
		log.Printf("Failed to read projects.html: %v", err)
		logStructured("ERROR", "api", "Failed to read projects file", map[string]interface{}{
			"error": err.Error(),
		})
		if os.IsNotExist(err) {
			http.Error(w, "Projects page not found", http.StatusNotFound)
		} else {
			http.Error(w, "Projects page not available", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(projectsHTML)
	
	logStructured("INFO", "api", "Projects page served successfully", nil)
}

func handleProjectDetailPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to /project-detail from %s", r.Method, r.RemoteAddr)
	
	logStructured("INFO", "api", "Project detail page request received", map[string]interface{}{
		"method": r.Method,
		"remote_addr": r.RemoteAddr,
		"user_agent": r.UserAgent(),
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

	// Read the project detail HTML file
	projectDetailHTML, err := os.ReadFile("project-detail.html")
	if err != nil {
		log.Printf("Failed to read project-detail.html: %v", err)
		logStructured("ERROR", "api", "Failed to read project detail file", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Project detail page not available", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(projectDetailHTML)
	
	logStructured("INFO", "api", "Project detail page served successfully", nil)
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
	
	// Fetch the created bookmark to return complete data
	var bookmarkID int
	err := db.QueryRow("SELECT id FROM bookmarks WHERE url = ? ORDER BY id DESC LIMIT 1", req.URL).Scan(&bookmarkID)
	if err != nil {
		log.Printf("Failed to fetch created bookmark ID: %v", err)
		// Still return success since the bookmark was saved
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
		return
	}
	
	// Get the complete bookmark data
	createdBookmark, err := getBookmarkByID(bookmarkID)
	if err != nil {
		log.Printf("Failed to fetch created bookmark: %v", err)
		// Still return success since the bookmark was saved
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdBookmark)
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
	// Validate database connection first
	if err := validateDB(); err != nil {
		return fmt.Errorf("failed to validate database connection: %v", err)
	}

	log.Printf("Saving bookmark to database: %s", req.URL)
	
	logStructured("INFO", "database", "Saving bookmark", map[string]interface{}{
		"url": req.URL,
		"title": req.Title,
		"action": req.Action,
		"content_length": len(req.Content),
	})
	
	// Convert tags and custom properties to JSON
	tagsJSON := tagsToJSON(req.Tags)
	customPropsJSON := customPropsToJSON(req.CustomProperties)

	insertSQL := `
	INSERT INTO bookmarks (url, title, description, content, action, shareTo, topic, tags, custom_properties)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	result, err := db.Exec(insertSQL, req.URL, req.Title, req.Description, req.Content, req.Action, req.ShareTo, req.Topic, tagsJSON, customPropsJSON)
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
		"archived": stats.Archived,
	})
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func getStatsSummary() (*SummaryStats, error) {
	// Validate database connection first
	if err := validateDB(); err != nil {
		return nil, fmt.Errorf("failed to validate database connection: %v", err)
	}

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
	
	// archived: bookmarks with action = "archived"
	err = db.QueryRow(`
		SELECT COUNT(*) FROM bookmarks 
		WHERE action = 'archived'
	`).Scan(&stats.Archived)
	if err != nil {
		return nil, fmt.Errorf("failed to count archived: %v", err)
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
		"archived": stats.Archived,
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

func handleTriageQueue(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to /api/bookmarks/triage from %s", r.Method, r.RemoteAddr)
	
	logStructured("INFO", "api", "Triage queue request received", map[string]interface{}{
		"method":      r.Method,
		"remote_addr": r.RemoteAddr,
	})
	
	if r.Method != http.MethodGet {
		log.Printf("Method not allowed: %s (expected GET)", r.Method)
		logStructured("WARN", "api", "Method not allowed", map[string]interface{}{
			"method":   r.Method,
			"expected": "GET",
		})
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")
	
	limit := 10 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	
	offset := 0 // default
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	triageData, err := getTriageQueue(limit, offset)
	if err != nil {
		log.Printf("Failed to get triage queue: %v", err)
		logStructured("ERROR", "database", "Failed to get triage queue", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Failed to get triage queue", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully retrieved triage queue with %d bookmarks", len(triageData.Bookmarks))
	logStructured("INFO", "database", "Triage queue retrieved", map[string]interface{}{
		"count":  len(triageData.Bookmarks),
		"total":  triageData.Total,
		"limit":  triageData.Limit,
		"offset": triageData.Offset,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(triageData)
}

func handleBookmarks(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to /api/bookmarks from %s", r.Method, r.RemoteAddr)
	
	logStructured("INFO", "api", "Bookmarks request received", map[string]interface{}{
		"method":      r.Method,
		"remote_addr": r.RemoteAddr,
	})
	
	if r.Method != http.MethodGet {
		log.Printf("Method not allowed: %s (expected GET)", r.Method)
		logStructured("WARN", "api", "Method not allowed", map[string]interface{}{
			"method":   r.Method,
			"expected": "GET",
		})
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	action := query.Get("action")
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")
	
	// Default to getting share bookmarks if no action specified
	if action == "" {
		action = "share"
	}
	
	limit := 50 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	
	offset := 0 // default
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get bookmarks by action
	bookmarksData, err := getBookmarksByAction(action, limit, offset)
	if err != nil {
		log.Printf("Failed to get bookmarks for action %s: %v", action, err)
		logStructured("ERROR", "database", "Failed to get bookmarks", map[string]interface{}{
			"error":  err.Error(),
			"action": action,
		})
		http.Error(w, "Failed to get bookmarks", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully retrieved %d bookmarks for action %s", len(bookmarksData.Bookmarks), action)
	logStructured("INFO", "database", "Bookmarks retrieved", map[string]interface{}{
		"count":  len(bookmarksData.Bookmarks),
		"total":  bookmarksData.Total,
		"action": action,
		"limit":  bookmarksData.Limit,
		"offset": bookmarksData.Offset,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookmarksData)
}

func getTriageQueue(limit, offset int) (*TriageResponse, error) {
	logStructured("INFO", "database", "Getting triage queue", map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	})

	// First get the total count
	var total int
	countSQL := `
		SELECT COUNT(*) FROM bookmarks 
		WHERE action IS NULL OR action = '' OR action = 'read-later'
	`
	
	err := db.QueryRow(countSQL).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count triage bookmarks: %v", err)
	}

	// Get the bookmarks
	querySQL := `
		SELECT id, url, title, description, timestamp, topic 
		FROM bookmarks 
		WHERE action IS NULL OR action = '' OR action = 'read-later'
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := db.Query(querySQL, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query triage bookmarks: %v", err)
	}
	defer rows.Close()

	var bookmarks []TriageBookmark
	for rows.Next() {
		var bookmark TriageBookmark
		var timestamp string
		var description, topic sql.NullString
		
		err := rows.Scan(&bookmark.ID, &bookmark.URL, &bookmark.Title, &description, &timestamp, &topic)
		if err != nil {
			return nil, fmt.Errorf("failed to scan triage bookmark: %v", err)
		}
		
		// Handle nullable description (store raw data)
		if description.Valid {
			bookmark.Description = description.String
		} else {
			bookmark.Description = ""
		}
		
		// Handle nullable topic (store raw data)
		if topic.Valid {
			bookmark.Topic = topic.String
		} else {
			bookmark.Topic = ""
		}
		
		// Store raw data (HTML escaping will be handled by frontend for display)
		
		// Parse and format timestamp
		if ts, err := time.Parse("2006-01-02 15:04:05", timestamp); err == nil {
			bookmark.Timestamp = ts.UTC().Format(time.RFC3339)
			
			// Calculate age
			age := time.Since(ts)
			if age.Hours() < 24 {
				bookmark.Age = fmt.Sprintf("%.0fh", age.Hours())
			} else {
				bookmark.Age = fmt.Sprintf("%.0fd", age.Hours()/24)
			}
		} else if ts, err := time.Parse(time.RFC3339, timestamp); err == nil {
			bookmark.Timestamp = timestamp
			
			// Calculate age for RFC3339 format
			age := time.Since(ts)
			if age.Hours() < 24 {
				bookmark.Age = fmt.Sprintf("%.0fh", age.Hours())
			} else {
				bookmark.Age = fmt.Sprintf("%.0fd", age.Hours()/24)
			}
		} else {
			bookmark.Timestamp = timestamp
			bookmark.Age = "unknown"
		}
		
		// Extract domain from URL
		if bookmark.URL == "" {
			bookmark.Domain = ""
		} else if u, err := url.Parse(bookmark.URL); err == nil && u.Host != "" {
			bookmark.Domain = u.Host // Use Host instead of Hostname to preserve port
		} else {
			bookmark.Domain = bookmark.URL // Return original URL for invalid URLs
		}
		
		// Generate suggested action
		bookmark.Suggested = getSuggestedAction(bookmark.Domain, bookmark.Title, bookmark.Description)
		
		bookmarks = append(bookmarks, bookmark)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating triage bookmarks: %v", err)
	}

	return &TriageResponse{
		Bookmarks: bookmarks,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	}, nil
}

func getBookmarksByAction(action string, limit, offset int) (*TriageResponse, error) {
	logStructured("INFO", "database", "Getting bookmarks by action", map[string]interface{}{
		"action": action,
		"limit":  limit,
		"offset": offset,
	})

	// First get the total count
	var total int
	countSQL := `SELECT COUNT(*) FROM bookmarks WHERE action = ?`
	
	err := db.QueryRow(countSQL, action).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count bookmarks for action %s: %v", action, err)
	}

	// Get the bookmarks with all fields including tags and custom properties
	querySQL := `
		SELECT id, url, title, description, timestamp, topic, shareTo, tags, custom_properties
		FROM bookmarks 
		WHERE action = ?
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := db.Query(querySQL, action, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query bookmarks for action %s: %v", action, err)
	}
	defer rows.Close()

	var bookmarks []TriageBookmark
	for rows.Next() {
		var bookmark TriageBookmark
		var timestamp string
		var description, topic, shareTo, tagsJSON, customPropsJSON sql.NullString
		
		err := rows.Scan(&bookmark.ID, &bookmark.URL, &bookmark.Title, &description, &timestamp, &topic, &shareTo, &tagsJSON, &customPropsJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bookmark: %v", err)
		}
		
		// Set optional fields
		if description.Valid {
			bookmark.Description = description.String
		}
		if topic.Valid {
			bookmark.Topic = topic.String
		}
		if shareTo.Valid {
			bookmark.ShareTo = shareTo.String
		}
		
		// Parse tags and custom properties from JSON
		if tagsJSON.Valid && tagsJSON.String != "" {
			bookmark.Tags = tagsFromJSON(tagsJSON.String)
		}
		
		if customPropsJSON.Valid && customPropsJSON.String != "" {
			bookmark.CustomProperties = customPropsFromJSON(customPropsJSON.String)
		}
		
		// Set the action for the response
		bookmark.Action = action
		
		// Parse timestamp
		bookmark.Timestamp = timestamp
		
		// Extract domain from URL
		if bookmark.URL == "" {
			bookmark.Domain = ""
		} else if u, err := url.Parse(bookmark.URL); err == nil && u.Host != "" {
			bookmark.Domain = u.Host
		} else {
			bookmark.Domain = bookmark.URL
		}
		
		// Calculate age
		bookmark.Age = calculateAge(timestamp)
		
		// Generate suggested action
		bookmark.Suggested = getSuggestedAction(bookmark.Domain, bookmark.Title, bookmark.Description)
		
		bookmarks = append(bookmarks, bookmark)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bookmark rows: %v", err)
	}

	return &TriageResponse{
		Bookmarks: bookmarks,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	}, nil
}

func getSuggestedAction(domain, title, description string) string {
	// Simple heuristics for suggested actions
	domain = strings.ToLower(domain)
	title = strings.ToLower(title)
	description = strings.ToLower(description)
	
	// Check for sharing indicators
	if strings.Contains(domain, "github") || strings.Contains(domain, "stackoverflow") ||
		strings.Contains(title, "tutorial") || strings.Contains(title, "guide") ||
		strings.Contains(description, "share") || strings.Contains(description, "useful") {
		return "share"
	}
	
	// Check for working indicators
	if strings.Contains(title, "documentation") || strings.Contains(title, "docs") ||
		strings.Contains(title, "api") || strings.Contains(title, "reference") ||
		strings.Contains(description, "work") || strings.Contains(description, "project") {
		return "working"
	}
	
	// Default to read-later
	return "read-later"
}

func handleProjects(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to /api/projects from %s", r.Method, r.RemoteAddr)
	
	logStructured("INFO", "api", "Projects request received", map[string]interface{}{
		"method":      r.Method,
		"remote_addr": r.RemoteAddr,
	})
	
	if r.Method != http.MethodGet {
		log.Printf("Method not allowed: %s (expected GET)", r.Method)
		logStructured("WARN", "api", "Method not allowed", map[string]interface{}{
			"method":   r.Method,
			"expected": "GET",
		})
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	projects, err := getProjects()
	if err != nil {
		log.Printf("Failed to get projects: %v", err)
		logStructured("ERROR", "database", "Failed to get projects", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Failed to get projects", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully retrieved projects")
	logStructured("INFO", "database", "Projects retrieved", map[string]interface{}{
		"activeProjects":       len(projects.ActiveProjects),
		"referenceCollections": len(projects.ReferenceCollections),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func getProjects() (*ProjectsResponse, error) {
	logStructured("INFO", "database", "Getting projects data", nil)
	
	response := &ProjectsResponse{
		ActiveProjects:       []ActiveProject{},
		ReferenceCollections: []ReferenceCollection{},
	}

	// Get active projects (topics with action = 'working')
	activeProjects, err := getActiveProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to get active projects: %v", err)
	}
	response.ActiveProjects = activeProjects

	// Get reference collections (topics that are frequently accessed but not actively worked on)
	referenceCollections, err := getReferenceCollections()
	if err != nil {
		return nil, fmt.Errorf("failed to get reference collections: %v", err)
	}
	response.ReferenceCollections = referenceCollections

	return response, nil
}

func getActiveProjects() ([]ActiveProject, error) {
	// Validate database connection first
	if err := validateDB(); err != nil {
		return nil, fmt.Errorf("failed to validate database connection: %v", err)
	}

	querySQL := `
		SELECT 
			p.id,
			p.name as topic,
			COUNT(b.id) as linkCount,
			COALESCE(MAX(b.timestamp), p.updated_at) as lastUpdated
		FROM projects p
		LEFT JOIN bookmarks b ON (b.project_id = p.id OR b.topic = p.name)
		WHERE p.status = 'active'
		GROUP BY p.id, p.name, p.updated_at
		HAVING COUNT(b.id) > 0
		ORDER BY MAX(COALESCE(b.timestamp, p.updated_at)) DESC
	`
	
	rows, err := db.Query(querySQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query active projects: %v", err)
	}
	defer rows.Close()

	var projects []ActiveProject
	for rows.Next() {
		var project ActiveProject
		var lastUpdated string
		
		err := rows.Scan(&project.ID, &project.Topic, &project.LinkCount, &lastUpdated)
		if err != nil {
			return nil, fmt.Errorf("failed to scan active project: %v", err)
		}
		
		// Parse timestamp and format as ISO 8601
		if timestamp, err := time.Parse("2006-01-02 15:04:05", lastUpdated); err == nil {
			project.LastUpdated = timestamp.UTC().Format(time.RFC3339)
		} else {
			project.LastUpdated = lastUpdated
		}
		
		// Determine status based on recency and calculate progress
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
		return nil, fmt.Errorf("error iterating active projects: %v", err)
	}

	return projects, nil
}

func getReferenceCollections() ([]ReferenceCollection, error) {
	// Validate database connection first
	if err := validateDB(); err != nil {
		return nil, fmt.Errorf("failed to validate database connection: %v", err)
	}

	// Get topics that have bookmarks but aren't actively being worked on
	// These could be documentation, resources, etc.
	querySQL := `
		SELECT 
			topic,
			COUNT(*) as linkCount,
			MAX(timestamp) as lastAccessed
		FROM bookmarks 
		WHERE topic IS NOT NULL AND topic != '' 
		AND topic NOT IN (
			SELECT DISTINCT topic FROM bookmarks 
			WHERE action = 'working' AND topic IS NOT NULL AND topic != ''
		)
		GROUP BY topic
		ORDER BY COUNT(*) DESC, MAX(timestamp) DESC
		LIMIT 10
	`
	
	rows, err := db.Query(querySQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query reference collections: %v", err)
	}
	defer rows.Close()

	var collections []ReferenceCollection
	for rows.Next() {
		var collection ReferenceCollection
		var lastAccessed string
		
		err := rows.Scan(&collection.Topic, &collection.LinkCount, &lastAccessed)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reference collection: %v", err)
		}
		
		// Parse timestamp and format as ISO 8601
		if timestamp, err := time.Parse("2006-01-02 15:04:05", lastAccessed); err == nil {
			collection.LastAccessed = timestamp.UTC().Format(time.RFC3339)
		} else {
			collection.LastAccessed = lastAccessed
		}
		
		collections = append(collections, collection)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reference collections: %v", err)
	}

	return collections, nil
}

func handleProjectDetail(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	
	logStructured("INFO", "api", "Project detail request received", map[string]interface{}{
		"method":      r.Method,
		"path":        r.URL.Path,
		"remote_addr": r.RemoteAddr,
	})
	
	if r.Method != http.MethodGet {
		log.Printf("Method not allowed: %s (expected GET)", r.Method)
		logStructured("WARN", "api", "Method not allowed", map[string]interface{}{
			"method":   r.Method,
			"expected": "GET",
		})
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract topic from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/projects/")
	if path == "" {
		log.Printf("Topic not provided in URL path")
		logStructured("WARN", "api", "Topic not provided", map[string]interface{}{
			"path": r.URL.Path,
		})
		http.Error(w, "Topic is required", http.StatusBadRequest)
		return
	}

	// URL decode the topic
	topic, err := url.QueryUnescape(path)
	if err != nil {
		log.Printf("Failed to decode topic from URL: %v", err)
		logStructured("ERROR", "api", "Failed to decode topic", map[string]interface{}{
			"error": err.Error(),
			"path":  path,
		})
		http.Error(w, "Invalid topic format", http.StatusBadRequest)
		return
	}

	projectDetail, err := getProjectDetail(topic)
	if err != nil {
		if strings.Contains(err.Error(), "project not found") {
			log.Printf("Project not found: %s", topic)
			logStructured("WARN", "api", "Project not found", map[string]interface{}{
				"topic": topic,
			})
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		log.Printf("Failed to get project detail for topic '%s': %v", topic, err)
		logStructured("ERROR", "database", "Failed to get project detail", map[string]interface{}{
			"error": err.Error(),
			"topic": topic,
		})
		http.Error(w, "Failed to get project detail", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully retrieved project detail for '%s' with %d bookmarks", topic, len(projectDetail.Bookmarks))
	logStructured("INFO", "database", "Project detail retrieved", map[string]interface{}{
		"topic":          topic,
		"bookmarkCount":  len(projectDetail.Bookmarks),
		"status":         projectDetail.Status,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projectDetail)
}

func handleProjectByID(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	
	logStructured("INFO", "api", "Project by ID request received", map[string]interface{}{
		"method": r.Method,
		"path": r.URL.Path,
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

	// Extract project ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/projects/id/")
	if path == "" {
		log.Printf("Project ID not provided in URL path")
		logStructured("WARN", "api", "Project ID not provided", map[string]interface{}{
			"path": r.URL.Path,
		})
		http.Error(w, "Project ID required", http.StatusBadRequest)
		return
	}

	projectID, err := strconv.Atoi(path)
	if err != nil {
		log.Printf("Invalid project ID: %s", path)
		logStructured("WARN", "api", "Invalid project ID", map[string]interface{}{
			"provided_id": path,
			"error": err.Error(),
		})
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	projectDetail, err := getProjectDetailByID(projectID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Printf("Project not found with ID: %d", projectID)
			logStructured("WARN", "api", "Project not found by ID", map[string]interface{}{
				"project_id": projectID,
			})
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		log.Printf("Failed to get project detail for ID %d: %v", projectID, err)
		logStructured("ERROR", "database", "Failed to get project detail by ID", map[string]interface{}{
			"project_id": projectID,
			"error": err.Error(),
		})
		http.Error(w, "Failed to get project detail", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully retrieved project detail for ID %d with %d bookmarks", projectID, len(projectDetail.Bookmarks))
	logStructured("INFO", "database", "Project detail retrieved by ID", map[string]interface{}{
		"project_id":     projectID,
		"project_name":   projectDetail.Topic,
		"bookmarkCount":  len(projectDetail.Bookmarks),
		"status":         projectDetail.Status,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projectDetail)
}

func getProjectDetail(topic string) (*ProjectDetailResponse, error) {
	logStructured("INFO", "database", "Getting project detail", map[string]interface{}{
		"topic": topic,
	})

	// First check if the project exists and get basic info
	var linkCount int
	var lastUpdated string
	var hasWorkingBookmarks bool

	// Check for working bookmarks in this topic
	var nullableLastUpdated sql.NullString
	err := db.QueryRow(`
		SELECT COUNT(*), MAX(timestamp) 
		FROM bookmarks 
		WHERE topic = ? AND action = 'working'
	`, topic).Scan(&linkCount, &nullableLastUpdated)
	
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get working project info: %v", err)
	}
	
	hasWorkingBookmarks = linkCount > 0
	if nullableLastUpdated.Valid {
		lastUpdated = nullableLastUpdated.String
	}

	// If no working bookmarks, check for any bookmarks with this topic
	if !hasWorkingBookmarks {
		err = db.QueryRow(`
			SELECT COUNT(*), MAX(timestamp) 
			FROM bookmarks 
			WHERE topic = ?
		`, topic).Scan(&linkCount, &nullableLastUpdated)
		
		if err != nil {
			return nil, fmt.Errorf("failed to get project info: %v", err)
		}
		
		if linkCount == 0 {
			return nil, fmt.Errorf("project not found: %s", topic)
		}
		
		if nullableLastUpdated.Valid {
			lastUpdated = nullableLastUpdated.String
		}
	}

	// Parse timestamp and format as ISO 8601
	var formattedLastUpdated string
	if timestamp, err := time.Parse("2006-01-02 15:04:05", lastUpdated); err == nil {
		formattedLastUpdated = timestamp.UTC().Format(time.RFC3339)
	} else {
		formattedLastUpdated = lastUpdated
	}

	// Determine status based on recency
	var status string
	if timestamp, err := time.Parse(time.RFC3339, formattedLastUpdated); err == nil {
		daysSince := time.Since(timestamp).Hours() / 24
		if daysSince <= 7 {
			status = "active"
		} else if daysSince <= 30 {
			status = "stale"
		} else {
			status = "inactive"
		}
	} else {
		status = "unknown"
	}


	// Get all bookmarks for this topic
	bookmarks, err := getProjectBookmarks(topic)
	if err != nil {
		return nil, fmt.Errorf("failed to get project bookmarks: %v", err)
	}

	response := &ProjectDetailResponse{
		Topic:       topic,
		LinkCount:   linkCount,
		LastUpdated: formattedLastUpdated,
		Status:      status,
		Bookmarks:   bookmarks,
	}

	return response, nil
}

func getProjectBookmarks(topic string) ([]ProjectBookmark, error) {
	querySQL := `
		SELECT id, url, title, description, content, timestamp, action
		FROM bookmarks 
		WHERE topic = ?
		ORDER BY timestamp DESC
	`
	
	rows, err := db.Query(querySQL, topic)
	if err != nil {
		return nil, fmt.Errorf("failed to query project bookmarks: %v", err)
	}
	defer rows.Close()

	var bookmarks []ProjectBookmark
	for rows.Next() {
		var bookmark ProjectBookmark
		var timestamp string
		var description, content, action sql.NullString
		
		err := rows.Scan(&bookmark.ID, &bookmark.URL, &bookmark.Title, 
			&description, &content, &timestamp, &action)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project bookmark: %v", err)
		}
		
		// Handle nullable fields (store raw data)
		if description.Valid {
			bookmark.Description = description.String
		}
		if content.Valid {
			bookmark.Content = content.String
		}
		if action.Valid {
			bookmark.Action = action.String
		}
		
		// Store raw data (HTML escaping will be handled by frontend for display)
		
		// Parse and format timestamp
		if ts, err := time.Parse("2006-01-02 15:04:05", timestamp); err == nil {
			bookmark.Timestamp = ts.UTC().Format(time.RFC3339)
			
			// Calculate age
			age := time.Since(ts)
			if age.Hours() < 24 {
				bookmark.Age = fmt.Sprintf("%.0fh", age.Hours())
			} else {
				bookmark.Age = fmt.Sprintf("%.0fd", age.Hours()/24)
			}
		} else if ts, err := time.Parse(time.RFC3339, timestamp); err == nil {
			bookmark.Timestamp = timestamp
			
			// Calculate age for RFC3339 format
			age := time.Since(ts)
			if age.Hours() < 24 {
				bookmark.Age = fmt.Sprintf("%.0fh", age.Hours())
			} else {
				bookmark.Age = fmt.Sprintf("%.0fd", age.Hours()/24)
			}
		} else {
			bookmark.Timestamp = timestamp
			bookmark.Age = "unknown"
		}
		
		// Extract domain from URL
		if bookmark.URL == "" {
			bookmark.Domain = ""
		} else if u, err := url.Parse(bookmark.URL); err == nil && u.Host != "" {
			bookmark.Domain = u.Host // Use Host instead of Hostname to preserve port
		} else {
			bookmark.Domain = bookmark.URL // Return original URL for invalid URLs
		}
		
		bookmarks = append(bookmarks, bookmark)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating project bookmarks: %v", err)
	}

	return bookmarks, nil
}

func getProjectDetailByID(projectID int) (*ProjectDetailResponse, error) {
	logStructured("INFO", "database", "Getting project detail by ID", map[string]interface{}{
		"project_id": projectID,
	})

	// Get project information from projects table
	var project Project
	err := db.QueryRow(`
		SELECT id, name, description, status, created_at, updated_at
		FROM projects 
		WHERE id = ?
	`, projectID).Scan(&project.ID, &project.Name, &project.Description, 
		&project.Status, &project.CreatedAt, &project.LastUpdated)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project with ID %d not found", projectID)
		}
		return nil, fmt.Errorf("failed to get project info: %v", err)
	}

	// Get bookmark count and last updated from bookmarks
	var linkCount int
	var lastBookmarkUpdate sql.NullString
	err = db.QueryRow(`
		SELECT COUNT(*), MAX(timestamp) 
		FROM bookmarks 
		WHERE project_id = ?
	`, projectID).Scan(&linkCount, &lastBookmarkUpdate)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get bookmark stats: %v", err)
	}

	// Use the most recent timestamp (project updated_at or bookmark timestamp)
	lastUpdated := project.LastUpdated
	if lastBookmarkUpdate.Valid {
		if bookmarkTime, err := time.Parse("2006-01-02 15:04:05", lastBookmarkUpdate.String); err == nil {
			if projectTime, err := time.Parse(time.RFC3339, project.LastUpdated); err == nil {
				if bookmarkTime.After(projectTime) {
					lastUpdated = bookmarkTime.UTC().Format(time.RFC3339)
				}
			}
		}
	}

	// Get all bookmarks for this project
	bookmarks, err := getProjectBookmarksByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project bookmarks: %v", err)
	}

	// Determine status based on activity
	var status string
	if timestamp, err := time.Parse(time.RFC3339, lastUpdated); err == nil {
		daysSince := time.Since(timestamp).Hours() / 24
		if daysSince <= 7 {
			status = "active"
		} else if daysSince <= 30 {
			status = "stale"
		} else {
			status = "inactive"
		}
	} else {
		status = "unknown"
	}

	response := &ProjectDetailResponse{
		Topic:       project.Name,
		LinkCount:   linkCount,
		LastUpdated: lastUpdated,
		Status:      status,
		Bookmarks:   bookmarks,
	}

	return response, nil
}

func getProjectBookmarksByID(projectID int) ([]ProjectBookmark, error) {
	querySQL := `
		SELECT id, url, title, description, content, timestamp, action
		FROM bookmarks 
		WHERE project_id = ?
		ORDER BY timestamp DESC
	`
	
	rows, err := db.Query(querySQL, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query project bookmarks: %v", err)
	}
	defer rows.Close()

	var bookmarks []ProjectBookmark
	for rows.Next() {
		var bookmark ProjectBookmark
		var timestamp string
		var description, content, action sql.NullString
		
		err := rows.Scan(&bookmark.ID, &bookmark.URL, &bookmark.Title, 
			&description, &content, &timestamp, &action)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project bookmark: %v", err)
		}
		
		// Handle nullable fields (store raw data)
		if description.Valid {
			bookmark.Description = description.String
		}
		if content.Valid {
			bookmark.Content = content.String
		}
		if action.Valid {
			bookmark.Action = action.String
		}
		
		// Store raw data (HTML escaping will be handled by frontend for display)
		
		// Parse timestamp and calculate age
		if ts, err := time.Parse("2006-01-02 15:04:05", timestamp); err == nil {
			bookmark.Timestamp = ts.UTC().Format(time.RFC3339)
			
			// Calculate age for RFC3339 format
			age := time.Since(ts)
			if age.Hours() < 24 {
				bookmark.Age = fmt.Sprintf("%.0fh", age.Hours())
			} else {
				bookmark.Age = fmt.Sprintf("%.0fd", age.Hours()/24)
			}
		} else {
			bookmark.Timestamp = timestamp
			bookmark.Age = "unknown"
		}
		
		// Extract domain from URL
		if bookmark.URL == "" {
			bookmark.Domain = ""
		} else if u, err := url.Parse(bookmark.URL); err == nil && u.Host != "" {
			bookmark.Domain = u.Host // Use Host instead of Hostname to preserve port
		} else {
			bookmark.Domain = bookmark.URL // Return original URL for invalid URLs
		}
		
		bookmarks = append(bookmarks, bookmark)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating project bookmarks: %v", err)
	}

	return bookmarks, nil
}

func handleBookmarkUpdate(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	
	logStructured("INFO", "api", "Bookmark update request received", map[string]interface{}{
		"method":      r.Method,
		"path":        r.URL.Path,
		"remote_addr": r.RemoteAddr,
	})
	
	if r.Method != http.MethodPatch && r.Method != http.MethodPut {
		log.Printf("Method not allowed: %s (expected PATCH or PUT)", r.Method)
		logStructured("WARN", "api", "Method not allowed", map[string]interface{}{
			"method":   r.Method,
			"expected": "PATCH or PUT",
		})
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract bookmark ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/bookmarks/")
	if path == "" {
		log.Printf("Bookmark ID not provided in URL path")
		logStructured("WARN", "api", "Bookmark ID not provided", map[string]interface{}{
			"path": r.URL.Path,
		})
		http.Error(w, "Bookmark ID is required", http.StatusBadRequest)
		return
	}

	bookmarkID, err := strconv.Atoi(path)
	if err != nil {
		log.Printf("Invalid bookmark ID: %s", path)
		logStructured("ERROR", "api", "Invalid bookmark ID", map[string]interface{}{
			"error": err.Error(),
			"id":    path,
		})
		http.Error(w, "Invalid bookmark ID", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodPut {
		// Handle full bookmark update (PUT)
		var req BookmarkFullUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Failed to decode JSON request: %v", err)
			logStructured("ERROR", "api", "JSON decode failed", map[string]interface{}{
				"error": err.Error(),
			})
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		log.Printf("Parsed full bookmark update request: ID=%d, Title=%s, URL=%s, Action=%s", 
			bookmarkID, req.Title, req.URL, req.Action)

		logStructured("INFO", "api", "Full bookmark update request parsed", map[string]interface{}{
			"id":     bookmarkID,
			"title":  req.Title,
			"url":    req.URL,
			"action": req.Action,
		})

		if err := updateFullBookmarkInDB(bookmarkID, req); err != nil {
			log.Printf("Failed to update bookmark in database: %v", err)
			logStructured("ERROR", "database", "Failed to update bookmark", map[string]interface{}{
				"error": err.Error(),
				"id":    bookmarkID,
			})
			http.Error(w, "Failed to update bookmark", http.StatusInternalServerError)
			return
		}
	} else {
		// Handle partial bookmark update (PATCH)
		var req BookmarkUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Failed to decode JSON request: %v", err)
			logStructured("ERROR", "api", "JSON decode failed", map[string]interface{}{
				"error": err.Error(),
			})
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		log.Printf("Parsed bookmark update request: ID=%d, Action=%s, Topic=%s", 
			bookmarkID, req.Action, req.Topic)

		logStructured("INFO", "api", "Bookmark update request parsed", map[string]interface{}{
			"id":     bookmarkID,
			"action": req.Action,
			"topic":  req.Topic,
		})

		if err := updateBookmarkInDB(bookmarkID, req); err != nil {
			log.Printf("Failed to update bookmark in database: %v", err)
			logStructured("ERROR", "database", "Failed to update bookmark", map[string]interface{}{
				"error": err.Error(),
				"id":    bookmarkID,
			})
			http.Error(w, "Failed to update bookmark", http.StatusInternalServerError)
			return
		}
	}

	log.Printf("Successfully updated bookmark: %d", bookmarkID)
	logStructured("INFO", "database", "Bookmark updated successfully", map[string]interface{}{
		"id": bookmarkID,
	})
	
	// Fetch and return the updated bookmark
	updatedBookmark, err := getBookmarkByID(bookmarkID)
	if err != nil {
		log.Printf("Failed to fetch updated bookmark: %v", err)
		logStructured("ERROR", "database", "Failed to fetch updated bookmark", map[string]interface{}{
			"error": err.Error(),
			"id":    bookmarkID,
		})
		http.Error(w, "Failed to fetch updated bookmark", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedBookmark)
}

func getBookmarkByID(id int) (*ProjectBookmark, error) {
	// Validate database connection
	if err := validateDB(); err != nil {
		return nil, fmt.Errorf("failed to validate database connection: %v", err)
	}

	var bookmark ProjectBookmark
	var description, content, action, topic, shareTo, tagsJSON, customPropsJSON sql.NullString
	
	err := db.QueryRow(`
		SELECT id, url, title, description, content, timestamp, action, topic, shareTo, tags, custom_properties
		FROM bookmarks 
		WHERE id = ?`, id).Scan(
		&bookmark.ID,
		&bookmark.URL,
		&bookmark.Title,
		&description,
		&content,
		&bookmark.Timestamp,
		&action,
		&topic,
		&shareTo,
		&tagsJSON,
		&customPropsJSON,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("bookmark not found")
		}
		return nil, fmt.Errorf("failed to query bookmark: %v", err)
	}

	// Handle nullable fields
	if description.Valid {
		bookmark.Description = description.String
	}
	if content.Valid {
		bookmark.Content = content.String
	}
	if action.Valid {
		bookmark.Action = action.String
	}
	if topic.Valid {
		bookmark.Topic = topic.String
	}
	if shareTo.Valid {
		bookmark.ShareTo = shareTo.String
	}

	// Parse tags and custom properties from JSON
	if tagsJSON.Valid && tagsJSON.String != "" {
		bookmark.Tags = tagsFromJSON(tagsJSON.String)
	}
	
	if customPropsJSON.Valid && customPropsJSON.String != "" {
		bookmark.CustomProperties = customPropsFromJSON(customPropsJSON.String)
	}

	// Extract domain from URL
	bookmark.Domain = extractDomain(bookmark.URL)
	
	// Calculate age
	bookmark.Age = calculateAge(bookmark.Timestamp)
	
	return &bookmark, nil
}

func extractDomain(urlStr string) string {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return "unknown"
	}
	return parsed.Hostname()
}

func calculateAge(timestamp string) string {
	// Parse the timestamp
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		// Try alternative formats
		t, err = time.Parse("2006-01-02 15:04:05", timestamp)
		if err != nil {
			return "unknown"
		}
	}
	
	now := time.Now()
	diff := now.Sub(t)
	
	minutes := int(diff.Minutes())
	hours := int(diff.Hours())
	days := int(diff.Hours() / 24)
	weeks := days / 7
	months := days / 30
	
	if minutes < 1 {
		return "just now"
	} else if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	} else if hours < 24 {
		return fmt.Sprintf("%dh", hours)
	} else if days < 7 {
		return fmt.Sprintf("%dd", days)
	} else if weeks < 4 {
		return fmt.Sprintf("%dw", weeks)
	} else {
		return fmt.Sprintf("%dmo", months)
	}
}

// Helper functions for handling JSON fields in database
func tagsToJSON(tags []string) string {
	if len(tags) == 0 {
		return "[]"
	}
	jsonBytes, err := json.Marshal(tags)
	if err != nil {
		log.Printf("Error marshaling tags: %v", err)
		return "[]"
	}
	return string(jsonBytes)
}

func tagsFromJSON(jsonStr string) []string {
	if jsonStr == "" || jsonStr == "[]" {
		return nil
	}
	var tags []string
	if err := json.Unmarshal([]byte(jsonStr), &tags); err != nil {
		log.Printf("Error unmarshaling tags: %v", err)
		return nil
	}
	return tags
}

func customPropsToJSON(props map[string]string) string {
	if len(props) == 0 {
		return "{}"
	}
	jsonBytes, err := json.Marshal(props)
	if err != nil {
		log.Printf("Error marshaling custom properties: %v", err)
		return "{}"
	}
	return string(jsonBytes)
}

func customPropsFromJSON(jsonStr string) map[string]string {
	if jsonStr == "" || jsonStr == "{}" {
		return nil
	}
	var props map[string]string
	if err := json.Unmarshal([]byte(jsonStr), &props); err != nil {
		log.Printf("Error unmarshaling custom properties: %v", err)
		return nil
	}
	return props
}

func updateBookmarkInDB(id int, req BookmarkUpdateRequest) error {
	log.Printf("Updating bookmark in database: %d", id)
	
	logStructured("INFO", "database", "Updating bookmark", map[string]interface{}{
		"id":        id,
		"action":    req.Action,
		"topic":     req.Topic,
		"projectId": req.ProjectID,
	})
	
	// Handle project assignment - support both topic and project_id
	var projectID *int
	var topic string
	
	if req.ProjectID > 0 {
		// Use provided project ID
		projectID = &req.ProjectID
		// Get project name for backward compatibility
		err := db.QueryRow("SELECT name FROM projects WHERE id = ?", req.ProjectID).Scan(&topic)
		if err != nil {
			log.Printf("Failed to find project with ID %d: %v", req.ProjectID, err)
			return fmt.Errorf("project with ID %d not found", req.ProjectID)
		}
	} else if req.Topic != "" {
		// Use topic name - find or create project
		var existingProjectID int
		err := db.QueryRow("SELECT id FROM projects WHERE name = ?", req.Topic).Scan(&existingProjectID)
		if err != nil {
			// Project doesn't exist, create it
			result, err := db.Exec(`
				INSERT INTO projects (name, description, status, created_at, updated_at)
				VALUES (?, ?, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			`, req.Topic, fmt.Sprintf("Auto-created for topic: %s", req.Topic))
			if err != nil {
				log.Printf("Failed to create project for topic %s: %v", req.Topic, err)
				return fmt.Errorf("failed to create project for topic %s", req.Topic)
			}
			
			newID, err := result.LastInsertId()
			if err != nil {
				return fmt.Errorf("failed to get new project ID")
			}
			existingProjectID = int(newID)
		}
		projectID = &existingProjectID
		topic = req.Topic
	} else {
		// Clear project assignment
		projectID = nil
		topic = ""
	}
	
	// Convert tags and custom properties to JSON
	tagsJSON := tagsToJSON(req.Tags)
	customPropsJSON := customPropsToJSON(req.CustomProperties)

	updateSQL := `UPDATE bookmarks SET action = ?, shareTo = ?, topic = ?, project_id = ?, tags = ?, custom_properties = ? WHERE id = ?`
	
	result, err := db.Exec(updateSQL, req.Action, req.ShareTo, topic, projectID, tagsJSON, customPropsJSON, id)
	if err != nil {
		log.Printf("Failed to update bookmark: %v", err)
		logStructured("ERROR", "database", "Update failed", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Failed to get rows affected: %v", err)
		logStructured("WARN", "database", "Failed to get rows affected", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}
	
	if rowsAffected == 0 {
		log.Printf("No bookmark found with ID: %d", id)
		logStructured("WARN", "database", "No bookmark found", map[string]interface{}{
			"id": id,
		})
		return fmt.Errorf("bookmark not found")
	}
	
	log.Printf("Successfully updated bookmark with ID: %d", id)
	logStructured("INFO", "database", "Bookmark updated", map[string]interface{}{
		"id":           id,
		"rowsAffected": rowsAffected,
	})
	
	return nil
}

func updateFullBookmarkInDB(id int, req BookmarkFullUpdateRequest) error {
	// Validate database connection first
	if err := validateDB(); err != nil {
		return fmt.Errorf("failed to validate database connection: %v", err)
	}

	log.Printf("Updating full bookmark in database: %d", id)
	
	// Validate required fields
	if req.Title == "" || req.URL == "" {
		return fmt.Errorf("title and URL are required fields")
	}
	
	// Handle project assignment logic similar to partial update
	var projectID sql.NullInt64
	var actualTopic string
	
	if req.Topic != "" {
		// Look for existing project with this topic/name
		var existingProjectID int
		err := db.QueryRow("SELECT id FROM projects WHERE name = ?", req.Topic).Scan(&existingProjectID)
		if err == sql.ErrNoRows {
			// Create new project
			result, err := db.Exec(`
				INSERT INTO projects (name, description, status, created_at, updated_at)
				VALUES (?, ?, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
				req.Topic, fmt.Sprintf("Project for %s bookmarks", req.Topic))
			if err != nil {
				logStructured("ERROR", "database", "Failed to create new project", map[string]interface{}{
					"error": err.Error(),
					"topic": req.Topic,
				})
				return fmt.Errorf("failed to create new project: %v", err)
			}
			
			newProjectID, err := result.LastInsertId()
			if err != nil {
				return fmt.Errorf("failed to get new project ID: %v", err)
			}
			
			projectID = sql.NullInt64{Int64: newProjectID, Valid: true}
			actualTopic = req.Topic
			
			logStructured("INFO", "database", "Created new project", map[string]interface{}{
				"projectId": newProjectID,
				"topic":     req.Topic,
			})
		} else if err != nil {
			logStructured("ERROR", "database", "Failed to query existing project", map[string]interface{}{
				"error": err.Error(),
				"topic": req.Topic,
			})
			return fmt.Errorf("failed to query existing project: %v", err)
		} else {
			// Use existing project
			projectID = sql.NullInt64{Int64: int64(existingProjectID), Valid: true}
			actualTopic = req.Topic
			
			logStructured("INFO", "database", "Using existing project", map[string]interface{}{
				"projectId": existingProjectID,
				"topic":     req.Topic,
			})
		}
	}
	
	// Convert tags and custom properties to JSON
	tagsJSON := tagsToJSON(req.Tags)
	customPropsJSON := customPropsToJSON(req.CustomProperties)

	// Update bookmark with all fields
	updateSQL := `
		UPDATE bookmarks 
		SET url = ?, title = ?, description = ?, action = ?, shareTo = ?, topic = ?, project_id = ?, tags = ?, custom_properties = ?
		WHERE id = ?`
	
	result, err := db.Exec(updateSQL, 
		req.URL, req.Title, req.Description, req.Action, req.ShareTo, actualTopic, projectID, tagsJSON, customPropsJSON, id)
	if err != nil {
		logStructured("ERROR", "database", "Failed to execute full bookmark update", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return fmt.Errorf("failed to update bookmark: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logStructured("ERROR", "database", "Failed to get rows affected", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return fmt.Errorf("failed to check update result: %v", err)
	}
	
	if rowsAffected == 0 {
		logStructured("WARN", "database", "No bookmark found with given ID", map[string]interface{}{
			"id": id,
		})
		return fmt.Errorf("no bookmark found with ID %d", id)
	}
	
	log.Printf("Successfully updated full bookmark with ID: %d", id)
	logStructured("INFO", "database", "Full bookmark update completed", map[string]interface{}{
		"id":           id,
		"title":        req.Title,
		"url":          req.URL,
		"action":       req.Action,
		"topic":        actualTopic,
		"rowsAffected": rowsAffected,
	})
	
	return nil
}