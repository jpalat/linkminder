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

CSV output format: `timestamp,url,title,description`

## Development Notes

- Port can be changed by modifying both the print statement and `ListenAndServe` call in `main.go`
- CSV file is created automatically with headers on first write
- No external dependencies means no `go.sum` file or dependency management needed
- Repository is not yet committed to git - files are staged for initial commit