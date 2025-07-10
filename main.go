package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

// sanitizeForLog removes newlines and carriage returns from user input to prevent log injection
func sanitizeForLog(input string) string {
	// Remove newlines and carriage returns to prevent log injection
	sanitized := strings.ReplaceAll(input, "\n", "")
	sanitized = strings.ReplaceAll(sanitized, "\r", "")
	// Also escape HTML to prevent HTML injection in log viewers
	sanitized = html.EscapeString(sanitized)
	return sanitized
}

type Project struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status"`
	LinkCount   int    `json:"linkCount"`
	LastUpdated string `json:"lastUpdated"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

type ProjectCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
}

type ProjectUpdateRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
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
	LatestURL   string `json:"latestURL"`
	LatestTitle string `json:"latestTitle"`
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
	
	// Only write to log file if it's initialized (not nil)
	if logFile != nil {
		if _, err := logFile.WriteString(string(jsonData) + "\n"); err != nil {
			log.Printf("Failed to write to log file: %v", err)
		}
	}
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
	defer func() {
		if err := logFile.Close(); err != nil {
			log.Printf("Failed to close log file: %v", err)
		}
	}()
	
	logStructured("INFO", "startup", "BookMinder API starting up", nil)
	
	// Initialize CORS configuration
	corsConfig = initCORSConfig()
	log.Printf("CORS configuration initialized")
	
	// Initialize security headers configuration  
	securityConfig = initSecurityConfig()
	log.Printf("Security headers configuration initialized")
	
	// Initialize database
	if err := initDatabase(); err != nil {
		logStructured("ERROR", "database", "Failed to initialize database", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}()
	
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
	http.HandleFunc("/api/bookmark/by-url", withCORS(handleBookmarkByURL))
	
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
	log.Printf("  POST /api/projects - Create a new project")
	log.Printf("  GET /api/projects/{id} - Get project by ID")
	log.Printf("  PUT /api/projects/{id} - Update project settings")
	log.Printf("  DELETE /api/projects/{id} - Delete a project")
	log.Printf("  GET /api/projects/{topic} - Get detailed view of a specific project")
	log.Printf("  GET /api/projects/id/{id} - Get detailed view of a project by ID")
	log.Printf("  PATCH /api/bookmarks/{id} - Update a bookmark (partial)")
	log.Printf("  PUT /api/bookmarks/{id} - Update a bookmark (full)")
	log.Printf("  DELETE /api/bookmarks/{id} - Soft delete a bookmark")
	log.Printf("  GET /api/bookmark/by-url?url={url} - Get bookmark by URL")
	
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
// CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxAge         string
	AllowWildcard  bool // Emergency development override
}

// SecurityHeaders configuration for HTTP security headers
type SecurityConfig struct {
	ContentSecurityPolicy string
	XFrameOptions         string
	XContentTypeOptions   string
	ReferrerPolicy        string
	PermissionsPolicy     string
	HSTSMaxAge            string
	EnableHSTS            bool
}

var corsConfig CORSConfig
var securityConfig SecurityConfig

func initCORSConfig() CORSConfig {
	// Load from environment with sensible defaults
	allowedOriginsEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	var origins []string
	
	if allowedOriginsEnv != "" {
		origins = strings.Split(allowedOriginsEnv, ",")
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
		}
		log.Printf("CORS origins loaded from environment: %v", origins)
	} else {
		// Development defaults
		origins = []string{
			"http://localhost:3000",
			"http://localhost:8080", 
			"http://127.0.0.1:3000",
			"http://127.0.0.1:8080",
		}
		log.Printf("CORS using development defaults: %v", origins)
	}
	
	// Emergency wildcard override (development only)
	allowWildcard := os.Getenv("CORS_ALLOW_WILDCARD") == "true"
	if allowWildcard {
		log.Printf("WARNING: CORS wildcard enabled - NOT FOR PRODUCTION!")
	}
	
	return CORSConfig{
		AllowedOrigins: origins,
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-Requested-With", "X-API-Key"},
		MaxAge:         "86400", // 24 hours
		AllowWildcard:  allowWildcard,
	}
}

func (c *CORSConfig) isOriginAllowed(origin string) bool {
	if origin == "" {
		return true // Same-origin requests
	}
	
	// Emergency wildcard override (development only)
	if c.AllowWildcard {
		return true
	}
	
	// Check exact matches
	for _, allowed := range c.AllowedOrigins {
		if origin == allowed {
			return true
		}
	}
	
	return false
}

func initSecurityConfig() SecurityConfig {
	// Load security headers from environment with secure defaults
	csp := os.Getenv("CSP_POLICY")
	if csp == "" {
		// Secure default CSP - restrictive but functional
		csp = "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none';"
	}
	
	hstsMaxAge := os.Getenv("HSTS_MAX_AGE")
	if hstsMaxAge == "" {
		hstsMaxAge = "31536000" // 1 year
	}
	
	enableHSTS := os.Getenv("ENABLE_HSTS") != "false" // Default to enabled
	
	return SecurityConfig{
		ContentSecurityPolicy: csp,
		XFrameOptions:         "DENY",
		XContentTypeOptions:   "nosniff",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		PermissionsPolicy:     "geolocation=(), microphone=(), camera=()",
		HSTSMaxAge:            hstsMaxAge,
		EnableHSTS:            enableHSTS,
	}
}

func securityHeadersMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set security headers
		w.Header().Set("Content-Security-Policy", securityConfig.ContentSecurityPolicy)
		w.Header().Set("X-Frame-Options", securityConfig.XFrameOptions)
		w.Header().Set("X-Content-Type-Options", securityConfig.XContentTypeOptions)
		w.Header().Set("Referrer-Policy", securityConfig.ReferrerPolicy)
		w.Header().Set("Permissions-Policy", securityConfig.PermissionsPolicy)
		
		// Only set HSTS for HTTPS requests
		if securityConfig.EnableHSTS && r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", fmt.Sprintf("max-age=%s; includeSubDomains", securityConfig.HSTSMaxAge))
		}
		
		// Call the next handler
		next.ServeHTTP(w, r)
	}
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		// Set CORS headers only for allowed origins
		if corsConfig.isOriginAllowed(origin) {
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(corsConfig.AllowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(corsConfig.AllowedHeaders, ", "))
			w.Header().Set("Access-Control-Max-Age", corsConfig.MaxAge)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Handle preflight OPTIONS requests
		if r.Method == "OPTIONS" {
			if corsConfig.isOriginAllowed(origin) {
				w.WriteHeader(http.StatusOK)
			} else {
				log.Printf("CORS: Blocked OPTIONS request from unauthorized origin: %s", origin)
				w.WriteHeader(http.StatusForbidden)
			}
			return
		}

		// For non-OPTIONS requests, check origin if present
		if origin != "" && !corsConfig.isOriginAllowed(origin) {
			log.Printf("CORS: Blocked request from unauthorized origin: %s", origin)
			logStructured("WARN", "security", "CORS blocked unauthorized origin", map[string]interface{}{
				"origin":     origin,
				"method":     r.Method,
				"path":       r.URL.Path,
				"user_agent": r.UserAgent(),
			})
			http.Error(w, "Origin not allowed", http.StatusForbidden)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	}
}

// Helper function to wrap handlers with security headers and CORS
func withCORS(handler http.HandlerFunc) http.HandlerFunc {
	return securityHeadersMiddleware(corsMiddleware(handler))
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to / from %s", sanitizeForLog(r.Method), sanitizeForLog(r.RemoteAddr))
	
	logStructured("INFO", "api", "Dashboard request received", map[string]interface{}{
		"method": r.Method,
		"remote_addr": r.RemoteAddr,
		"user_agent": r.UserAgent(),
	})
	
	if r.Method != http.MethodGet {
		log.Printf("Method not allowed: %s (expected GET)", sanitizeForLog(r.Method))
		logStructured("WARN", "api", "Method not allowed", map[string]interface{}{
			"method": r.Method,
			"expected": "GET",
		})
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate and read the dashboard HTML file
	filename := "dashboard.html"
	if err := validateHTMLFile(filename); err != nil {
		log.Printf("Invalid HTML file: %v", sanitizeForLog(err.Error()))
		http.Error(w, "File not accessible", http.StatusForbidden)
		return
	}
	
	dashboardHTML, err := os.ReadFile(filename)
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if _, err := w.Write(dashboardHTML); err != nil {
		log.Printf("Failed to write dashboard HTML: %v", err)
		http.Error(w, "Failed to serve dashboard", http.StatusInternalServerError)
		return
	}
	
	logStructured("INFO", "api", "Dashboard served successfully", nil)
}

func handleProjectsPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to /projects from %s", sanitizeForLog(r.Method), sanitizeForLog(r.RemoteAddr))
	
	logStructured("INFO", "api", "Projects page request received", map[string]interface{}{
		"method": r.Method,
		"remote_addr": r.RemoteAddr,
		"user_agent": r.UserAgent(),
	})
	
	if r.Method != http.MethodGet {
		log.Printf("Method not allowed: %s (expected GET)", sanitizeForLog(r.Method))
		logStructured("WARN", "api", "Method not allowed", map[string]interface{}{
			"method": r.Method,
			"expected": "GET",
		})
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate and read the projects HTML file
	filename := "projects.html"
	if err := validateHTMLFile(filename); err != nil {
		log.Printf("Invalid HTML file: %v", sanitizeForLog(err.Error()))
		http.Error(w, "File not accessible", http.StatusForbidden)
		return
	}
	
	projectsHTML, err := os.ReadFile(filename)
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
	if _, err := w.Write(projectsHTML); err != nil {
		log.Printf("Failed to write projects HTML: %v", err)
		http.Error(w, "Failed to serve projects page", http.StatusInternalServerError)
		return
	}
	
	logStructured("INFO", "api", "Projects page served successfully", nil)
}

func handleProjectDetailPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to /project-detail from %s", sanitizeForLog(r.Method), sanitizeForLog(r.RemoteAddr))
	
	logStructured("INFO", "api", "Project detail page request received", map[string]interface{}{
		"method": r.Method,
		"remote_addr": r.RemoteAddr,
		"user_agent": r.UserAgent(),
	})
	
	if r.Method != http.MethodGet {
		log.Printf("Method not allowed: %s (expected GET)", sanitizeForLog(r.Method))
		logStructured("WARN", "api", "Method not allowed", map[string]interface{}{
			"method": r.Method,
			"expected": "GET",
		})
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate and read the project detail HTML file
	filename := "project-detail.html"
	if err := validateHTMLFile(filename); err != nil {
		log.Printf("Invalid HTML file: %v", sanitizeForLog(err.Error()))
		http.Error(w, "File not accessible", http.StatusForbidden)
		return
	}
	
	projectDetailHTML, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("Failed to read project-detail.html: %v", err)
		logStructured("ERROR", "api", "Failed to read project detail file", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Project detail page not available", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write(projectDetailHTML); err != nil {
		log.Printf("Failed to write project detail HTML: %v", err)
		http.Error(w, "Failed to serve project detail page", http.StatusInternalServerError)
		return
	}
	
	logStructured("INFO", "api", "Project detail page served successfully", nil)
}

func handleBookmark(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to /bookmark from %s", sanitizeForLog(r.Method), sanitizeForLog(r.RemoteAddr))
	
	logStructured("INFO", "api", "Bookmark request received", map[string]interface{}{
		"method": r.Method,
		"remote_addr": r.RemoteAddr,
		"user_agent": r.UserAgent(),
	})
	
	if r.Method != http.MethodPost {
		log.Printf("Method not allowed: %s (expected POST)", sanitizeForLog(r.Method))
		logStructured("WARN", "api", "Method not allowed", map[string]interface{}{
			"method": r.Method,
			"expected": "POST",
		})
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BookmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode JSON request: %v", sanitizeForLog(err.Error()))
		logStructured("ERROR", "api", "JSON decode failed", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Parsed bookmark request: URL=%s, Title=%s, Action=%s, Topic=%s", 
		sanitizeForLog(req.URL), sanitizeForLog(req.Title), sanitizeForLog(req.Action), sanitizeForLog(req.Topic))

	logStructured("INFO", "api", "Bookmark request parsed", map[string]interface{}{
		"url": req.URL,
		"title": req.Title,
		"action": req.Action,
		"topic": req.Topic,
		"has_content": len(req.Content) > 0,
	})

	// Validate input using enhanced validation
	if err := validateBookmarkInput(req); err != nil {
		logStructured("WARN", "api", "Validation failed", map[string]interface{}{
			"error": err.Error(),
			"url":   req.URL,
			"title": req.Title,
		})
		log.Printf("Validation failed: %v", sanitizeForLog(err.Error()))
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	if err := saveBookmarkToDB(req); err != nil {
		log.Printf("Failed to save bookmark to database: %v", sanitizeForLog(err.Error()))
		logStructured("ERROR", "database", "Failed to save bookmark", map[string]interface{}{
			"error": err.Error(),
			"url": req.URL,
		})
		http.Error(w, "Failed to save bookmark", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully saved bookmark: %s", sanitizeForLog(req.URL))
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
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "success"}); err != nil {
			log.Printf("Failed to encode success response: %v", err)
		}
		return
	}
	
	// Get the complete bookmark data
	createdBookmark, err := getBookmarkByID(bookmarkID)
	if err != nil {
		log.Printf("Failed to fetch created bookmark: %v", err)
		// Still return success since the bookmark was saved
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "success"}); err != nil {
			log.Printf("Failed to encode success response: %v", err)
		}
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(createdBookmark); err != nil {
		log.Printf("Failed to encode bookmark response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func handleTopics(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to /topics from %s", sanitizeForLog(r.Method), sanitizeForLog(r.RemoteAddr))
	
	logStructured("INFO", "api", "Topics request received", map[string]interface{}{
		"method": r.Method,
		"remote_addr": r.RemoteAddr,
	})
	
	if r.Method != http.MethodGet {
		log.Printf("Method not allowed: %s (expected GET)", sanitizeForLog(r.Method))
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
	if err := json.NewEncoder(w).Encode(map[string][]string{"topics": topics}); err != nil {
		log.Printf("Failed to encode topics response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func saveBookmarkToDB(req BookmarkRequest) error {
	// Validate database connection first
	if err := validateDB(); err != nil {
		return fmt.Errorf("failed to validate database connection: %v", err)
	}

	log.Printf("Saving bookmark to database: %s", sanitizeForLog(req.URL))
	
	logStructured("INFO", "database", "Saving bookmark", map[string]interface{}{
		"url": req.URL,
		"title": req.Title,
		"action": req.Action,
		"content_length": len(req.Content),
	})
	
	// Convert tags and custom properties to JSON
	tagsJSON := tagsToJSON(req.Tags)
	customPropsJSON := customPropsToJSON(req.CustomProperties)

	// Check if bookmark already exists
	var existingID int
	checkSQL := `SELECT id FROM bookmarks WHERE url = ? AND (deleted = FALSE OR deleted IS NULL) LIMIT 1`
	err := db.QueryRow(checkSQL, req.URL).Scan(&existingID)
	
	if err == nil {
		// Bookmark exists, update it
		log.Printf("Updating existing bookmark with ID: %d", existingID)
		logStructured("INFO", "database", "Updating existing bookmark", map[string]interface{}{
			"id": existingID,
			"url": req.URL,
		})
		
		updateSQL := `
		UPDATE bookmarks 
		SET title = ?, description = ?, content = ?, action = ?, shareTo = ?, topic = ?, tags = ?, custom_properties = ?, timestamp = CURRENT_TIMESTAMP
		WHERE id = ?`
		
		_, err = db.Exec(updateSQL, req.Title, req.Description, req.Content, req.Action, req.ShareTo, req.Topic, tagsJSON, customPropsJSON, existingID)
		if err != nil {
			log.Printf("Failed to update bookmark: %v", err)
			logStructured("ERROR", "database", "Update failed", map[string]interface{}{
				"error": err.Error(),
				"id": existingID,
				"url": req.URL,
			})
			return err
		}
		
		log.Printf("Successfully updated bookmark with ID: %d", existingID)
		logStructured("INFO", "database", "Bookmark updated", map[string]interface{}{
			"id": existingID,
			"url": req.URL,
			"title": req.Title,
		})
		
		return nil
	} else if err != sql.ErrNoRows {
		// Database error
		log.Printf("Error checking for existing bookmark: %v", err)
		logStructured("ERROR", "database", "Error checking existing bookmark", map[string]interface{}{
			"error": err.Error(),
			"url": req.URL,
		})
		return err
	}
	
	// No existing bookmark found, create new one
	log.Printf("Creating new bookmark for URL: %s", sanitizeForLog(req.URL))
	logStructured("INFO", "database", "Creating new bookmark", map[string]interface{}{
		"url": req.URL,
	})
	
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
	
	log.Printf("Successfully created bookmark with ID: %d", id)
	logStructured("INFO", "database", "Bookmark created", map[string]interface{}{
		"id": id,
		"url": req.URL,
		"title": req.Title,
	})
	
	return nil
}

func getTopicsFromDB() ([]string, error) {
	log.Printf("Reading topics from database")
	
	logStructured("INFO", "database", "Querying topics", nil)
	
	querySQL := `SELECT DISTINCT topic FROM bookmarks WHERE topic IS NOT NULL AND topic != '' AND (deleted = FALSE OR deleted IS NULL) ORDER BY topic`
	
	rows, err := db.Query(querySQL)
	if err != nil {
		log.Printf("Failed to query topics: %v", err)
		logStructured("ERROR", "database", "Topics query failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Failed to close rows: %v", err)
		}
	}()
	
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
	log.Printf("Received %s request to /api/stats/summary from %s", sanitizeForLog(r.Method), sanitizeForLog(r.RemoteAddr))
	
	logStructured("INFO", "api", "Stats summary request received", map[string]interface{}{
		"method": r.Method,
		"remote_addr": r.RemoteAddr,
	})
	
	if r.Method != http.MethodGet {
		log.Printf("Method not allowed: %s (expected GET)", sanitizeForLog(r.Method))
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
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("Failed to encode stats response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func getStatsSummary() (*SummaryStats, error) {
	// Validate database connection first
	if err := validateDB(); err != nil {
		return nil, fmt.Errorf("failed to validate database connection: %v", err)
	}

	logStructured("INFO", "database", "Computing stats summary", nil)
	
	stats := &SummaryStats{}
	
	// Get total bookmarks count
	err := db.QueryRow("SELECT COUNT(*) FROM bookmarks WHERE deleted = FALSE OR deleted IS NULL").Scan(&stats.TotalBookmarks)
	if err != nil {
		return nil, fmt.Errorf("failed to count total bookmarks: %v", err)
	}
	
	// Count by action categories
	// needsTriage: bookmarks with no action or action = "read-later"
	err = db.QueryRow(`
		SELECT COUNT(*) FROM bookmarks 
		WHERE (action IS NULL OR action = '' OR action = 'read-later') AND (deleted = FALSE OR deleted IS NULL)
	`).Scan(&stats.NeedsTriage)
	if err != nil {
		return nil, fmt.Errorf("failed to count needs triage: %v", err)
	}
	
	// activeProjects: unique topics in "working" action
	err = db.QueryRow(`
		SELECT COUNT(DISTINCT topic) FROM bookmarks 
		WHERE action = 'working' AND topic IS NOT NULL AND topic != '' AND (deleted = FALSE OR deleted IS NULL)
	`).Scan(&stats.ActiveProjects)
	if err != nil {
		return nil, fmt.Errorf("failed to count active projects: %v", err)
	}
	
	// readyToShare: bookmarks with action = "share"
	err = db.QueryRow(`
		SELECT COUNT(*) FROM bookmarks 
		WHERE action = 'share' AND (deleted = FALSE OR deleted IS NULL)
	`).Scan(&stats.ReadyToShare)
	if err != nil {
		return nil, fmt.Errorf("failed to count ready to share: %v", err)
	}
	
	// archived: bookmarks with action = "archived"
	err = db.QueryRow(`
		SELECT COUNT(*) FROM bookmarks 
		WHERE action = 'archived' AND (deleted = FALSE OR deleted IS NULL)
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
			stats.topic,
			stats.count,
			stats.lastUpdated,
			latest.url as latestURL,
			latest.title as latestTitle
		FROM (
			SELECT 
				topic,
				COUNT(*) as count,
				MAX(timestamp) as lastUpdated
			FROM bookmarks 
			WHERE action = 'working' AND topic IS NOT NULL AND topic != '' AND (deleted = FALSE OR deleted IS NULL)
			GROUP BY topic
		) stats
		LEFT JOIN bookmarks latest ON stats.topic = latest.topic 
			AND latest.timestamp = stats.lastUpdated
			AND latest.action = 'working'
			AND (latest.deleted = FALSE OR latest.deleted IS NULL)
			AND latest.id = (
				SELECT MAX(id) FROM bookmarks b 
				WHERE b.topic = stats.topic 
				AND b.timestamp = stats.lastUpdated 
				AND b.action = 'working'
				AND (b.deleted = FALSE OR b.deleted IS NULL)
			)
		ORDER BY stats.lastUpdated DESC
		LIMIT 10
	`
	
	rows, err := db.Query(querySQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query project stats: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Failed to close rows: %v", err)
		}
	}()
	
	var projects []ProjectStat
	for rows.Next() {
		var project ProjectStat
		var lastUpdated string
		
		err := rows.Scan(&project.Topic, &project.Count, &lastUpdated, &project.LatestURL, &project.LatestTitle)
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
	log.Printf("Received %s request to /api/bookmarks/triage from %s", sanitizeForLog(r.Method), sanitizeForLog(r.RemoteAddr))
	
	logStructured("INFO", "api", "Triage queue request received", map[string]interface{}{
		"method":      r.Method,
		"remote_addr": r.RemoteAddr,
	})
	
	if r.Method != http.MethodGet {
		log.Printf("Method not allowed: %s (expected GET)", sanitizeForLog(r.Method))
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
	if err := json.NewEncoder(w).Encode(triageData); err != nil {
		log.Printf("Failed to encode triage response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func handleBookmarks(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to /api/bookmarks from %s", sanitizeForLog(r.Method), sanitizeForLog(r.RemoteAddr))
	
	logStructured("INFO", "api", "Bookmarks request received", map[string]interface{}{
		"method":      r.Method,
		"remote_addr": r.RemoteAddr,
	})
	
	if r.Method != http.MethodGet {
		log.Printf("Method not allowed: %s (expected GET)", sanitizeForLog(r.Method))
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
		log.Printf("Failed to get bookmarks for action %s: %v", sanitizeForLog(action), err)
		logStructured("ERROR", "database", "Failed to get bookmarks", map[string]interface{}{
			"error":  err.Error(),
			"action": action,
		})
		http.Error(w, "Failed to get bookmarks", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully retrieved %d bookmarks for action %s", len(bookmarksData.Bookmarks), sanitizeForLog(action))
	logStructured("INFO", "database", "Bookmarks retrieved", map[string]interface{}{
		"count":  len(bookmarksData.Bookmarks),
		"total":  bookmarksData.Total,
		"action": action,
		"limit":  bookmarksData.Limit,
		"offset": bookmarksData.Offset,
	})

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(bookmarksData); err != nil {
		log.Printf("Failed to encode bookmarks response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
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
		WHERE (action IS NULL OR action = '' OR action = 'read-later') AND (deleted = FALSE OR deleted IS NULL)
	`
	
	err := db.QueryRow(countSQL).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count triage bookmarks: %v", err)
	}

	// Get the bookmarks
	querySQL := `
		SELECT id, url, title, description, timestamp, topic 
		FROM bookmarks 
		WHERE (action IS NULL OR action = '' OR action = 'read-later') AND (deleted = FALSE OR deleted IS NULL)
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := db.Query(querySQL, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query triage bookmarks: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Failed to close rows: %v", err)
		}
	}()

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
	countSQL := `SELECT COUNT(*) FROM bookmarks WHERE action = ? AND (deleted = FALSE OR deleted IS NULL)`
	
	err := db.QueryRow(countSQL, action).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count bookmarks for action %s: %v", action, err)
	}

	// Get the bookmarks with all fields including tags and custom properties
	querySQL := `
		SELECT id, url, title, description, timestamp, topic, shareTo, tags, custom_properties
		FROM bookmarks 
		WHERE action = ? AND (deleted = FALSE OR deleted IS NULL)
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := db.Query(querySQL, action, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query bookmarks for action %s: %v", action, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Failed to close rows: %v", err)
		}
	}()

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

func getBookmarkByURL(urlStr string) (*TriageBookmark, error) {
	logStructured("INFO", "database", "Getting bookmark by URL", map[string]interface{}{
		"url": urlStr,
	})

	querySQL := `
		SELECT id, url, title, description, timestamp, action, topic, shareTo, tags, custom_properties
		FROM bookmarks 
		WHERE url = ? AND (deleted = FALSE OR deleted IS NULL)
		ORDER BY timestamp DESC
		LIMIT 1
	`
	
	row := db.QueryRow(querySQL, urlStr)
	
	var bookmark TriageBookmark
	var timestamp string
	var description, action, topic, shareTo, tagsJSON, customPropsJSON sql.NullString
	
	err := row.Scan(&bookmark.ID, &bookmark.URL, &bookmark.Title, &description, &timestamp, &action, &topic, &shareTo, &tagsJSON, &customPropsJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No bookmark found for this URL
		}
		return nil, fmt.Errorf("failed to scan bookmark: %v", err)
	}
	
	// Set optional fields
	if description.Valid {
		bookmark.Description = description.String
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
	
	// Parse tags from JSON
	if tagsJSON.Valid && tagsJSON.String != "" {
		var tags []string
		if err := json.Unmarshal([]byte(tagsJSON.String), &tags); err == nil {
			bookmark.Tags = tags
		}
	}
	
	// Parse custom properties from JSON
	if customPropsJSON.Valid && customPropsJSON.String != "" {
		var customProps map[string]string
		if err := json.Unmarshal([]byte(customPropsJSON.String), &customProps); err == nil {
			bookmark.CustomProperties = customProps
		}
	}
	
	// Set timestamp and calculate age
	bookmark.Timestamp = timestamp
	bookmark.Age = calculateAge(timestamp)
	
	// Extract domain from URL
	if parsedURL, err := url.Parse(bookmark.URL); err == nil {
		bookmark.Domain = parsedURL.Host
	}
	
	return &bookmark, nil
}

func handleBookmarkByURL(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to /api/bookmark/by-url from %s", sanitizeForLog(r.Method), sanitizeForLog(r.RemoteAddr))
	
	logStructured("INFO", "api", "Bookmark by URL request received", map[string]interface{}{
		"method":      r.Method,
		"remote_addr": r.RemoteAddr,
	})
	
	if r.Method != "GET" {
		log.Printf("Method not allowed: %s", sanitizeForLog(r.Method))
		logStructured("WARN", "api", "Method not allowed", map[string]interface{}{
			"method": r.Method,
			"expected": "GET",
		})
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Get URL parameter
	urlParam := r.URL.Query().Get("url")
	if urlParam == "" {
		log.Printf("Missing URL parameter")
		logStructured("WARN", "api", "Missing URL parameter", nil)
		http.Error(w, "URL parameter is required", http.StatusBadRequest)
		return
	}
	
	// Validate URL format
	if _, err := url.Parse(urlParam); err != nil {
		log.Printf("Invalid URL format: %v", err)
		logStructured("WARN", "api", "Invalid URL format", map[string]interface{}{
			"url": urlParam,
			"error": err.Error(),
		})
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	
	// Get bookmark from database
	bookmark, err := getBookmarkByURL(urlParam)
	if err != nil {
		log.Printf("Failed to get bookmark by URL: %v", err)
		logStructured("ERROR", "api", "Failed to get bookmark by URL", map[string]interface{}{
			"url": urlParam,
			"error": err.Error(),
		})
		http.Error(w, "Failed to retrieve bookmark", http.StatusInternalServerError)
		return
	}
	
	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	
	// Return empty response if no bookmark found
	if bookmark == nil {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte(`{"found": false}`)); err != nil {
			log.Printf("Failed to write not found response: %v", err)
		}
		return
	}
	
	// Return the bookmark
	response := map[string]interface{}{
		"found": true,
		"bookmark": bookmark,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode bookmark response: %v", err)
		logStructured("ERROR", "api", "Failed to encode response", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	
	logStructured("INFO", "api", "Bookmark by URL served successfully", map[string]interface{}{
		"url": urlParam,
		"found": true,
	})
}

func handleProjects(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to /api/projects from %s", sanitizeForLog(r.Method), sanitizeForLog(r.RemoteAddr))
	
	logStructured("INFO", "api", "Projects request received", map[string]interface{}{
		"method":      r.Method,
		"remote_addr": r.RemoteAddr,
	})
	
	// Route to handleProjectSettings for individual project operations (path includes ID)
	pathWithoutPrefix := strings.TrimPrefix(r.URL.Path, "/api/projects")
	if pathWithoutPrefix != "" && pathWithoutPrefix != "/" {
		handleProjectSettings(w, r)
		return
	}
	
	switch r.Method {
	case http.MethodGet:
		handleGetProjects(w, r)
	case http.MethodPost:
		handleCreateProject(w, r)
	default:
		log.Printf("Method not allowed: %s", sanitizeForLog(r.Method))
		logStructured("WARN", "api", "Method not allowed", map[string]interface{}{
			"method": r.Method,
			"allowed": []string{"GET", "POST"},
		})
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetProjects(w http.ResponseWriter, r *http.Request) {

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
	if err := json.NewEncoder(w).Encode(projects); err != nil {
		log.Printf("Failed to encode projects response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var req ProjectCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode project creation request: %v", sanitizeForLog(err.Error()))
		logStructured("ERROR", "api", "Invalid JSON in project creation", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Validate required fields
	if strings.TrimSpace(req.Name) == "" {
		log.Printf("Project name is required")
		logStructured("WARN", "api", "Project name missing", nil)
		http.Error(w, "Project name is required", http.StatusBadRequest)
		return
	}
	
	// Set default status if not provided
	if req.Status == "" {
		req.Status = "active"
	}
	
	// Create the project
	project, err := createProject(req)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			log.Printf("Project name already exists: %s", sanitizeForLog(req.Name))
			logStructured("WARN", "database", "Duplicate project name", map[string]interface{}{
				"name": req.Name,
			})
			http.Error(w, "Project name already exists", http.StatusConflict)
			return
		}
		
		log.Printf("Failed to create project: %v", err)
		logStructured("ERROR", "database", "Failed to create project", map[string]interface{}{
			"error": err.Error(),
			"name":  req.Name,
		})
		http.Error(w, "Failed to create project", http.StatusInternalServerError)
		return
	}
	
	log.Printf("Successfully created project: %s (ID: %d)", sanitizeForLog(project.Name), project.ID)
	logStructured("INFO", "database", "Project created", map[string]interface{}{
		"id":   project.ID,
		"name": project.Name,
	})
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(project); err != nil {
		log.Printf("Failed to encode created project response: %v", err)
		// Can't call http.Error after WriteHeader, so just log the error
		return
	}
}

func handleProjectSettings(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to project settings from %s", sanitizeForLog(r.Method), sanitizeForLog(r.RemoteAddr))
	
	// Extract project ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/projects/")
	if path == "" || path == "/" {
		http.Error(w, "Project ID required", http.StatusBadRequest)
		return
	}
	
	// Handle the existing topic-based routing
	if !isNumeric(path) {
		// This is probably a topic-based request, route to existing handler
		if r.Method == http.MethodGet {
			handleProjectDetail(w, r)
			return
		}
		http.Error(w, "Only GET method supported for topic-based projects", http.StatusMethodNotAllowed)
		return
	}
	
	projectID, err := strconv.Atoi(path)
	if err != nil {
		log.Printf("Invalid project ID: %s", sanitizeForLog(path))
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}
	
	switch r.Method {
	case http.MethodGet:
		handleGetProject(w, r, projectID)
	case http.MethodPut:
		handleUpdateProject(w, r, projectID)
	case http.MethodDelete:
		handleDeleteProject(w, r, projectID)
	default:
		log.Printf("Method not allowed: %s", sanitizeForLog(r.Method))
		logStructured("WARN", "api", "Method not allowed for project settings", map[string]interface{}{
			"method": r.Method,
			"allowed": []string{"GET", "PUT", "DELETE"},
		})
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetProject(w http.ResponseWriter, r *http.Request, projectID int) {
	project, err := getProjectByID(projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Project not found: %d", projectID)
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		
		log.Printf("Failed to get project %d: %v", projectID, err)
		logStructured("ERROR", "database", "Failed to get project", map[string]interface{}{
			"error":     err.Error(),
			"projectId": projectID,
		})
		http.Error(w, "Failed to get project", http.StatusInternalServerError)
		return
	}
	
	log.Printf("Successfully retrieved project: %d", projectID)
	logStructured("INFO", "database", "Project retrieved", map[string]interface{}{
		"projectId": projectID,
		"name":      project.Name,
	})
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(project); err != nil {
		log.Printf("Failed to encode project response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func handleUpdateProject(w http.ResponseWriter, r *http.Request, projectID int) {
	// Read the request body once and parse it for both struct and raw data
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	
	var req ProjectUpdateRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		log.Printf("Failed to decode project update request: %v", sanitizeForLog(err.Error()))
		logStructured("ERROR", "api", "Invalid JSON in project update", map[string]interface{}{
			"error":     err.Error(),
			"projectId": projectID,
		})
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Parse raw JSON to check if name field was explicitly provided
	var rawData map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &rawData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// If name field is explicitly provided, validate it's not empty
	if nameValue, nameExists := rawData["name"]; nameExists {
		if nameStr, ok := nameValue.(string); ok && strings.TrimSpace(nameStr) == "" {
			log.Printf("Project name cannot be empty")
			logStructured("WARN", "api", "Empty project name in update", map[string]interface{}{
				"projectId": projectID,
				"name":      nameStr,
			})
			http.Error(w, "Project name cannot be empty", http.StatusBadRequest)
			return
		}
	}
	
	// Update the project
	project, err := updateProject(projectID, req)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Project not found for update: %d", projectID)
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			log.Printf("Project name already exists: %s", sanitizeForLog(req.Name))
			logStructured("WARN", "database", "Duplicate project name in update", map[string]interface{}{
				"name":      req.Name,
				"projectId": projectID,
			})
			http.Error(w, "Project name already exists", http.StatusConflict)
			return
		}
		
		log.Printf("Failed to update project %d: %v", projectID, err)
		logStructured("ERROR", "database", "Failed to update project", map[string]interface{}{
			"error":     err.Error(),
			"projectId": projectID,
		})
		http.Error(w, "Failed to update project", http.StatusInternalServerError)
		return
	}
	
	log.Printf("Successfully updated project: %d", projectID)
	logStructured("INFO", "database", "Project updated", map[string]interface{}{
		"projectId": projectID,
		"name":      project.Name,
	})
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(project); err != nil {
		log.Printf("Failed to encode updated project response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func handleDeleteProject(w http.ResponseWriter, r *http.Request, projectID int) {
	// Check if project exists first
	_, err := getProjectByID(projectID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Project not found for deletion: %d", projectID)
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		
		log.Printf("Failed to check project existence %d: %v", projectID, err)
		logStructured("ERROR", "database", "Failed to check project for deletion", map[string]interface{}{
			"error":     err.Error(),
			"projectId": projectID,
		})
		http.Error(w, "Failed to check project", http.StatusInternalServerError)
		return
	}
	
	// Delete the project (this should cascade to bookmarks)
	err = deleteProject(projectID)
	if err != nil {
		log.Printf("Failed to delete project %d: %v", projectID, err)
		logStructured("ERROR", "database", "Failed to delete project", map[string]interface{}{
			"error":     err.Error(),
			"projectId": projectID,
		})
		http.Error(w, "Failed to delete project", http.StatusInternalServerError)
		return
	}
	
	log.Printf("Successfully deleted project: %d", projectID)
	logStructured("INFO", "database", "Project deleted", map[string]interface{}{
		"projectId": projectID,
	})
	
	w.WriteHeader(http.StatusNoContent)
}

// Database functions for project settings

func createProject(req ProjectCreateRequest) (*Project, error) {
	logStructured("INFO", "database", "Creating project", map[string]interface{}{
		"name": req.Name,
	})
	
	now := time.Now()
	
	result, err := db.Exec(`
		INSERT INTO projects (name, description, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, req.Name, req.Description, req.Status, now, now)
	
	if err != nil {
		return nil, err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	
	project := &Project{
		ID:          int(id),
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		LinkCount:   0,
		CreatedAt:   now.Format(time.RFC3339),
		UpdatedAt:   now.Format(time.RFC3339),
	}
	
	return project, nil
}

func getProjectByID(projectID int) (*Project, error) {
	logStructured("INFO", "database", "Getting project by ID", map[string]interface{}{
		"projectId": projectID,
	})
	
	var project Project
	var createdAt, updatedAt time.Time
	
	err := db.QueryRow(`
		SELECT p.id, p.name, p.description, p.status, p.created_at, p.updated_at,
		       COUNT(b.id) as link_count
		FROM projects p
		LEFT JOIN bookmarks b ON (p.name = b.topic OR p.id = b.project_id) AND b.action = 'working' AND (b.deleted = FALSE OR b.deleted IS NULL)
		WHERE p.id = ?
		GROUP BY p.id, p.name, p.description, p.status, p.created_at, p.updated_at
	`, projectID).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.Status,
		&createdAt,
		&updatedAt,
		&project.LinkCount,
	)
	
	if err != nil {
		return nil, err
	}
	
	project.CreatedAt = createdAt.Format(time.RFC3339)
	project.UpdatedAt = updatedAt.Format(time.RFC3339)
	project.LastUpdated = updatedAt.Format(time.RFC3339)
	
	return &project, nil
}

func updateProject(projectID int, req ProjectUpdateRequest) (*Project, error) {
	logStructured("INFO", "database", "Updating project", map[string]interface{}{
		"projectId": projectID,
	})
	
	// Build dynamic query based on provided fields
	var setParts []string
	var args []interface{}
	
	if req.Name != "" {
		setParts = append(setParts, "name = ?")
		args = append(args, req.Name)
	}
	
	if req.Description != "" {
		setParts = append(setParts, "description = ?")
		args = append(args, req.Description)
	}
	
	if req.Status != "" {
		setParts = append(setParts, "status = ?")
		args = append(args, req.Status)
	}
	
	if len(setParts) == 0 {
		// No fields to update, just return current project
		return getProjectByID(projectID)
	}
	
	// Always update the updated_at timestamp
	setParts = append(setParts, "updated_at = ?")
	args = append(args, time.Now())
	
	// Add projectID to args for WHERE clause
	args = append(args, projectID)
	
	// Use whitelist approach to prevent SQL injection
	allowedColumns := map[string]bool{
		"name = ?":        true,
		"description = ?": true,
		"status = ?":      true,
		"updated_at = ?":  true,
	}
	
	// Validate all setParts against whitelist
	for _, part := range setParts {
		if !allowedColumns[part] {
			return nil, fmt.Errorf("invalid column in update: %s", part)
		}
	}
	
	query := fmt.Sprintf("UPDATE projects SET %s WHERE id = ?", strings.Join(setParts, ", "))
	
	result, err := db.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	
	if rowsAffected == 0 {
		return nil, sql.ErrNoRows
	}
	
	// Return updated project
	return getProjectByID(projectID)
}

func deleteProject(projectID int) error {
	logStructured("INFO", "database", "Deleting project", map[string]interface{}{
		"projectId": projectID,
	})
	
	// First, update any bookmarks that reference this project to remove the reference
	// We'll set project_id to NULL and keep the topic for backward compatibility
	_, err := db.Exec(`
		UPDATE bookmarks 
		SET project_id = NULL 
		WHERE project_id = ?
	`, projectID)
	
	if err != nil {
		return fmt.Errorf("failed to update bookmarks: %v", err)
	}
	
	// Now delete the project
	result, err := db.Exec("DELETE FROM projects WHERE id = ?", projectID)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	
	return nil
}

// Helper function to check if a string is numeric
func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
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
		LEFT JOIN bookmarks b ON (b.project_id = p.id OR b.topic = p.name) AND (b.deleted = FALSE OR b.deleted IS NULL)
		WHERE p.status = 'active'
		GROUP BY p.id, p.name, p.updated_at
		HAVING COUNT(b.id) > 0
		ORDER BY MAX(COALESCE(b.timestamp, p.updated_at)) DESC
	`
	
	rows, err := db.Query(querySQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query active projects: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Failed to close rows: %v", err)
		}
	}()

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
		WHERE topic IS NOT NULL AND topic != '' AND (deleted = FALSE OR deleted IS NULL)
		AND topic NOT IN (
			SELECT DISTINCT topic FROM bookmarks 
			WHERE action = 'working' AND topic IS NOT NULL AND topic != '' AND (deleted = FALSE OR deleted IS NULL)
		)
		GROUP BY topic
		ORDER BY COUNT(*) DESC, MAX(timestamp) DESC
		LIMIT 10
	`
	
	rows, err := db.Query(querySQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query reference collections: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Failed to close rows: %v", err)
		}
	}()

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
	log.Printf("Received %s request to %s from %s", sanitizeForLog(r.Method), sanitizeForLog(r.URL.Path), sanitizeForLog(r.RemoteAddr))
	
	logStructured("INFO", "api", "Project detail request received", map[string]interface{}{
		"method":      r.Method,
		"path":        r.URL.Path,
		"remote_addr": r.RemoteAddr,
	})
	
	if r.Method != http.MethodGet {
		log.Printf("Method not allowed: %s (expected GET)", sanitizeForLog(r.Method))
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
		log.Printf("Failed to decode topic from URL: %v", sanitizeForLog(err.Error()))
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
			log.Printf("Project not found: %s", sanitizeForLog(topic))
			logStructured("WARN", "api", "Project not found", map[string]interface{}{
				"topic": topic,
			})
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		log.Printf("Failed to get project detail for topic '%s': %v", sanitizeForLog(topic), err)
		logStructured("ERROR", "database", "Failed to get project detail", map[string]interface{}{
			"error": err.Error(),
			"topic": topic,
		})
		http.Error(w, "Failed to get project detail", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully retrieved project detail for '%s' with %d bookmarks", sanitizeForLog(topic), len(projectDetail.Bookmarks))
	logStructured("INFO", "database", "Project detail retrieved", map[string]interface{}{
		"topic":          topic,
		"bookmarkCount":  len(projectDetail.Bookmarks),
		"status":         projectDetail.Status,
	})

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(projectDetail); err != nil {
		log.Printf("Failed to encode project detail response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func handleProjectByID(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to %s from %s", sanitizeForLog(r.Method), sanitizeForLog(r.URL.Path), sanitizeForLog(r.RemoteAddr))
	
	logStructured("INFO", "api", "Project by ID request received", map[string]interface{}{
		"method": r.Method,
		"path": r.URL.Path,
		"remote_addr": r.RemoteAddr,
	})
	
	if r.Method != http.MethodGet {
		log.Printf("Method not allowed: %s (expected GET)", sanitizeForLog(r.Method))
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
		log.Printf("Invalid project ID: %s", sanitizeForLog(path))
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
	if err := json.NewEncoder(w).Encode(projectDetail); err != nil {
		log.Printf("Failed to encode project detail response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
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
		WHERE topic = ? AND action = 'working' AND (deleted = FALSE OR deleted IS NULL)
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
			WHERE topic = ? AND (deleted = FALSE OR deleted IS NULL)
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
		WHERE topic = ? AND (deleted = FALSE OR deleted IS NULL)
		ORDER BY timestamp DESC
	`
	
	rows, err := db.Query(querySQL, topic)
	if err != nil {
		return nil, fmt.Errorf("failed to query project bookmarks: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Failed to close rows: %v", err)
		}
	}()

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
		WHERE project_id = ? AND (deleted = FALSE OR deleted IS NULL)
		ORDER BY timestamp DESC
	`
	
	rows, err := db.Query(querySQL, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query project bookmarks: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Failed to close rows: %v", err)
		}
	}()

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
	log.Printf("Received %s request to %s from %s", sanitizeForLog(r.Method), sanitizeForLog(r.URL.Path), sanitizeForLog(r.RemoteAddr))
	
	logStructured("INFO", "api", "Bookmark update request received", map[string]interface{}{
		"method":      r.Method,
		"path":        r.URL.Path,
		"remote_addr": r.RemoteAddr,
	})
	
	if r.Method != http.MethodPatch && r.Method != http.MethodPut && r.Method != http.MethodDelete {
		log.Printf("Method not allowed: %s (expected PATCH, PUT, or DELETE)", sanitizeForLog(r.Method))
		logStructured("WARN", "api", "Method not allowed", map[string]interface{}{
			"method":   r.Method,
			"expected": "PATCH, PUT, or DELETE",
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
		log.Printf("Invalid bookmark ID: %s", sanitizeForLog(path))
		logStructured("ERROR", "api", "Invalid bookmark ID", map[string]interface{}{
			"error": err.Error(),
			"id":    path,
		})
		http.Error(w, "Invalid bookmark ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodDelete:
		// Handle bookmark soft delete (DELETE)
		log.Printf("Soft deleting bookmark: %d", bookmarkID)
		logStructured("INFO", "api", "Bookmark soft delete request", map[string]interface{}{
			"id": bookmarkID,
		})

		if err := softDeleteBookmarkInDB(bookmarkID); err != nil {
			if err == sql.ErrNoRows {
				log.Printf("Bookmark not found: %d", bookmarkID)
				logStructured("WARN", "api", "Bookmark not found", map[string]interface{}{
					"id": bookmarkID,
				})
				http.Error(w, "Bookmark not found", http.StatusNotFound)
				return
			}
			log.Printf("Failed to soft delete bookmark: %v", err)
			logStructured("ERROR", "database", "Failed to soft delete bookmark", map[string]interface{}{
				"error": err.Error(),
				"id":    bookmarkID,
			})
			http.Error(w, "Failed to delete bookmark", http.StatusInternalServerError)
			return
		}

		log.Printf("Successfully soft deleted bookmark: %d", bookmarkID)
		logStructured("INFO", "database", "Bookmark soft deleted successfully", map[string]interface{}{
			"id": bookmarkID,
		})

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Bookmark deleted successfully",
			"id":      bookmarkID,
		}); err != nil {
			log.Printf("Failed to encode JSON response: %v", err)
		}
		return
	case http.MethodPut:
		// Handle full bookmark update (PUT)
		var req BookmarkFullUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Failed to decode JSON request: %v", sanitizeForLog(err.Error()))
			logStructured("ERROR", "api", "JSON decode failed", map[string]interface{}{
				"error": err.Error(),
			})
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		log.Printf("Parsed full bookmark update request: ID=%d, Title=%s, URL=%s, Action=%s", 
			bookmarkID, sanitizeForLog(req.Title), sanitizeForLog(req.URL), sanitizeForLog(req.Action))

		logStructured("INFO", "api", "Full bookmark update request parsed", map[string]interface{}{
			"id":     bookmarkID,
			"title":  req.Title,
			"url":    req.URL,
			"action": req.Action,
		})

		if err := updateFullBookmarkInDB(bookmarkID, req); err != nil {
			log.Printf("Failed to update bookmark in database: %v", sanitizeForLog(err.Error()))
			logStructured("ERROR", "database", "Failed to update bookmark", map[string]interface{}{
				"error": err.Error(),
				"id":    bookmarkID,
			})
			http.Error(w, "Failed to update bookmark", http.StatusInternalServerError)
			return
		}
	case http.MethodPatch:
		// Handle partial bookmark update (PATCH)
		var req BookmarkUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Failed to decode JSON request: %v", sanitizeForLog(err.Error()))
			logStructured("ERROR", "api", "JSON decode failed", map[string]interface{}{
				"error": err.Error(),
			})
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		log.Printf("Parsed bookmark update request: ID=%d, Action=%s, Topic=%s", 
			bookmarkID, sanitizeForLog(req.Action), sanitizeForLog(req.Topic))

		logStructured("INFO", "api", "Bookmark update request parsed", map[string]interface{}{
			"id":     bookmarkID,
			"action": req.Action,
			"topic":  req.Topic,
		})

		if err := updateBookmarkInDB(bookmarkID, req); err != nil {
			log.Printf("Failed to update bookmark in database: %v", sanitizeForLog(err.Error()))
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
	if err := json.NewEncoder(w).Encode(updatedBookmark); err != nil {
		log.Printf("Failed to encode updated bookmark response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
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
		WHERE id = ? AND (deleted = FALSE OR deleted IS NULL)`, id).Scan(
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
				log.Printf("Failed to create project for topic %s: %v", sanitizeForLog(req.Topic), err)
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

func softDeleteBookmarkInDB(id int) error {
	log.Printf("Soft deleting bookmark in database: %d", id)
	
	logStructured("INFO", "database", "Soft deleting bookmark", map[string]interface{}{
		"id": id,
	})
	
	// Validate database connection first
	if err := validateDB(); err != nil {
		return fmt.Errorf("failed to validate database connection: %v", err)
	}
	
	// Update the bookmark to mark it as deleted
	result, err := db.Exec("UPDATE bookmarks SET deleted = TRUE WHERE id = ? AND (deleted = FALSE OR deleted IS NULL)", id)
	if err != nil {
		logStructured("ERROR", "database", "Failed to soft delete bookmark", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return fmt.Errorf("failed to soft delete bookmark: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	
	logStructured("INFO", "database", "Bookmark soft deleted", map[string]interface{}{
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

// validateHTMLFile validates that the file path is safe to serve
func validateHTMLFile(filename string) error {
	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(filename)
	
	// Ensure the file has .html extension
	if !strings.HasSuffix(cleanPath, ".html") {
		return fmt.Errorf("invalid file extension")
	}
	
	// Get absolute path of current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}
	
	// Get absolute path of the requested file
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}
	
	// Ensure the file is within the current working directory
	if !strings.HasPrefix(absPath, cwd) {
		return fmt.Errorf("file path outside allowed directory")
	}
	
	// Additional check: prevent any path containing ".."
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid file path contains directory traversal")
	}
	
	return nil
}

// validateBookmarkInput validates bookmark request data
func validateBookmarkInput(req BookmarkRequest) error {
	// Validate required fields
	if strings.TrimSpace(req.URL) == "" {
		return fmt.Errorf("URL is required")
	}
	if strings.TrimSpace(req.Title) == "" {
		return fmt.Errorf("title is required")
	}
	
	// Validate URL format
	parsedURL, err := url.Parse(req.URL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return fmt.Errorf("invalid URL format")
	}
	
	// Validate input lengths
	if len(req.URL) > 2048 {
		return fmt.Errorf("URL too long (max 2048 characters)")
	}
	if len(req.Title) > 500 {
		return fmt.Errorf("title too long (max 500 characters)")
	}
	if len(req.Description) > 2000 {
		return fmt.Errorf("description too long (max 2000 characters)")
	}
	
	return nil
}