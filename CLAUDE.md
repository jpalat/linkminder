# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

BookMinder API is a simple Go HTTP server that accepts bookmark submissions via POST requests and stores them in CSV format. The entire application is contained in a single `main.go` file.

## Architecture

- **Single-file application**: All logic is in `main.go`
- **HTTP Server**: Runs on port 9090 using Go's standard `net/http`
- **Single endpoint**: `POST /bookmark` accepts JSON with `url`, `title`, and optional `description`
- **CSV storage**: Data is appended to `bookmarks.csv` with timestamps
- **No external dependencies**: Uses only Go standard library

## Common Development Commands

```bash
# Run the server
go run main.go

# Build executable
go build -o bookminderapi main.go

# Test the API endpoint
curl -X POST http://localhost:9090/bookmark \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com","title":"Example Site","description":"Test bookmark"}'

# Stop server (if running in background)
pkill -f "go run main.go"
```

## Data Model

The API accepts JSON requests with this structure:
```json
{
  "url": "required string",
  "title": "required string", 
  "description": "optional string"
}
```

CSV output format: `timestamp,url,title,description,content,action,shareTo,topic`

## Development Notes

- Port can be changed by modifying the `port` variable in `main.go`
- CSV file is created automatically with headers on first write
- No external dependencies means no `go.sum` file or dependency management needed
- Repository is not yet committed to git - files are staged for initial commit

## Logging

The API includes comprehensive logging for debugging and monitoring:

- **Startup logs**: Server initialization and endpoint registration
- **Request logs**: HTTP method, endpoint, and client IP for all requests
- **Validation logs**: Details about failed validations (missing fields, invalid JSON)
- **CSV operation logs**: File operations, record writing, and topic extraction
- **Success logs**: Confirmation of successful operations
- **Error logs**: Detailed error information for troubleshooting

All logs include timestamps and are written to stdout/stderr.