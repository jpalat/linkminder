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

# Test the bookmark endpoint
curl -X POST http://localhost:9090/bookmark \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com","title":"Example Site","description":"Test bookmark","action":"read-later"}'

# Test the topics endpoint
curl -X GET http://localhost:9090/topics

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

## Development Notes

- Port can be changed by modifying the `port` variable in `main.go`
- SQLite database (`bookmarks.db`) is created automatically on first run
- Requires SQLite driver dependency - run `go mod tidy` to install
- Database includes auto-incrementing IDs and automatic timestamps
- Topics are extracted dynamically from existing bookmark data

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