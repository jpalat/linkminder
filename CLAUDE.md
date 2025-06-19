# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

BookMinder API is a simple Go HTTP server that accepts bookmark submissions via POST requests and stores them in SQLite database. The entire application is contained in a single `main.go` file.

## Architecture

- **Single-file application**: All logic is in `main.go`
- **HTTP Server**: Runs on port 9090 using Go's standard `net/http`
- **Endpoints**: 
  - `POST /bookmark` accepts JSON with `url`, `title`, and optional fields
  - `GET /topics` returns list of available topics
  - `GET /api/stats/summary` returns dashboard summary statistics
  - `GET /api/projects/{topic}` returns detailed view of a specific project
- **SQLite storage**: Data is stored in `bookmarks.db` with automatic timestamps
- **Dependencies**: Uses SQLite driver (`github.com/mattn/go-sqlite3`)

## Common Development Commands

```bash
# Install dependencies
go mod tidy

# Run the server
go run main.go

# Build executable
go build -o bookminderapi main.go

# Run unit tests
go test

# Run tests with verbose output
go test -v

# Run specific test
go test -run TestBookmarkWorkflow_EndToEnd -v

# Run all project-related tests
go test -run "Projects" -v

# Stop server (if running in background)
pkill -f "bookminderapi"
```

## Data Model

The API accepts JSON requests with this structure:
```json
{
  "url": "required string",
  "title": "required string", 
  "description": "optional string",
  "content": "optional string",
  "action": "optional string (read-later, share, working)",
  "shareTo": "optional string (for share action)",
  "topic": "optional string (for working action)"
}
```

SQLite database schema:
```sql
CREATE TABLE bookmarks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    url TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    content TEXT,
    action TEXT,
    shareTo TEXT,
    topic TEXT
);
```

## API Endpoints

### Dashboard Summary Statistics
```http
GET /api/stats/summary
```

Returns analytics and overview data for dashboard display.

### Project Detail View
```http
GET /api/projects/{topic}
```

Returns detailed information about a specific project including all bookmarks within that topic. The topic parameter should be URL-encoded if it contains special characters.

**Response:**
```json
{
  "topic": "Energy",
  "linkCount": 3,
  "lastUpdated": "2025-06-16T19:37:02Z",
  "status": "active",
  "progress": 30,
  "bookmarks": [
    {
      "id": 12,
      "url": "https://example.com",
      "title": "Example Bookmark",
      "description": "Description text",
      "content": "Full content text",
      "timestamp": "2025-06-16T19:37:02Z",
      "domain": "example.com",
      "age": "2d",
      "action": "working"
    }
  ]
}
```

**Summary Response:**
```json
{
  "needsTriage": 23,
  "activeProjects": 4,
  "readyToShare": 7,
  "totalBookmarks": 347,
  "projectStats": [
    {
      "topic": "React Migration",
      "count": 15,
      "lastUpdated": "2025-06-11T10:30:00Z",
      "status": "active"
    }
  ]
}
```

**Statistics Definitions:**
- `needsTriage`: Bookmarks with no action or action="read-later" 
- `activeProjects`: Count of unique topics with action="working"
- `readyToShare`: Bookmarks with action="share"
- `totalBookmarks`: Total number of saved bookmarks
- `projectStats`: Top 10 working topics with counts and activity status

**Project Status:**
- `active`: Updated within last 7 days
- `stale`: Updated 7-30 days ago  
- `inactive`: Updated over 30 days ago

## Development Notes

- Port can be changed by modifying the `port` variable in `main.go`
- SQLite database (`bookmarks.db`) is created automatically on first run
- Requires SQLite driver dependency - run `go mod tidy` to install
- Database includes auto-incrementing IDs and automatic timestamps
- Topics are extracted dynamically from existing bookmark data

## Testing

The API includes comprehensive unit tests covering:

- **Database operations**: Bookmark saving, topic extraction, stats calculation
- **HTTP handlers**: All endpoint functionality with proper error handling  
- **API responses**: JSON structure validation and field verification
- **Edge cases**: Invalid inputs, empty databases, timestamp parsing
- **Integration tests**: End-to-end workflow validation

### Test Coverage
- `main_test.go` contains 30+ test functions
- Database functions tested with temporary SQLite databases
- HTTP handlers tested with `httptest` package
- Both success and error scenarios covered
- Project functionality comprehensively tested

## Logging

The API includes comprehensive logging for debugging and monitoring:

### Console Logging
- **Startup logs**: Server initialization and endpoint registration
- **Request logs**: HTTP method, endpoint, and client IP for all requests
- **Validation logs**: Details about failed validations (missing fields, invalid JSON)
- **Database operation logs**: SQL operations, record writing, and topic extraction
- **Success logs**: Confirmation of successful operations
- **Error logs**: Detailed error information for troubleshooting

### Structured Logging
JSON-formatted logs are written to `bookminderapi.log` with the following structure:
```json
{
  "timestamp": "2023-12-10T15:30:45Z",
  "level": "INFO|WARN|ERROR",
  "message": "Human readable message",
  "component": "startup|api|database|system",
  "data": {
    "key": "value"
  }
}
```

**Log Levels**:
- `INFO`: Normal operations, successful requests
- `WARN`: Non-fatal issues like invalid HTTP methods
- `ERROR`: Failed operations, database errors

**Components**:
- `startup`: Server initialization
- `api`: HTTP request handling
- `database`: SQLite operations
- `system`: General system events

## Development Best Practices

- **Testing**:
  - Prefer test scripts to curl tests