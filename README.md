# BookMinder API

A simple, self-hosted bookmark management system with a Go HTTP API backend, Vue.js frontend, and Chrome extension. Save, organize, and share bookmarks across your development workflow.

## ✨ Features

- **📑 Simple Bookmark Management**: Save URLs with titles, descriptions, and custom metadata
- **🗂️ Project Organization**: Group bookmarks into projects with status tracking
- **🎯 Action-Based Workflow**: Triage bookmarks with actions (read-later, working, share, archived)
- **📊 Dashboard Analytics**: Track bookmark statistics and project progress
- **🌐 REST API**: Full-featured API for integration with tools and scripts
- **🔧 Chrome Extension**: One-click bookmark saving from any webpage
- **💻 Web Interface**: Modern Vue.js frontend for bookmark management
- **🔍 Advanced Filtering**: Search, filter, and sort bookmarks by multiple criteria
- **📱 Responsive Design**: Works on desktop and mobile devices

## 🚀 Quick Start

### Prerequisites
- Go 1.23+ 
- Node.js 16+ (for frontend development)

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd bookminderapi
   ```

2. **Install Go dependencies**
   ```bash
   go mod tidy
   ```

3. **Run the server**
   ```bash
   go run main.go
   ```

   The API server will start on `http://localhost:9090`

4. **Test the installation**
   ```bash
   curl -X POST http://localhost:9090/bookmark \
     -H "Content-Type: application/json" \
     -d '{"url": "https://golang.org", "title": "Go Programming Language", "action": "read-later"}'
   ```

## 📁 Project Structure

```
bookminderapi/
├── main.go                 # Single-file Go HTTP server
├── main_test.go           # Comprehensive test suite (71.5% coverage)
├── migrations/            # Database schema migrations
├── frontend/              # Vue.js web interface
├── extension/             # Chrome browser extension
├── docs/                  # API documentation
└── scripts/               # Utility scripts
```

## 🔧 API Endpoints

### Core Bookmark Operations
- `POST /bookmark` - Save a new bookmark
- `PATCH /api/bookmarks/{id}` - Update bookmark action/topic
- `PUT /api/bookmarks/{id}` - Update entire bookmark
- `GET /api/bookmarks?action={action}` - Get bookmarks by action

### Project Management
- `GET /api/projects` - List all projects with statistics
- `POST /api/projects` - Create a new project
- `GET /api/projects/id/{id}` - Get project details by ID
- `PUT /api/projects/{id}` - Update project settings
- `DELETE /api/projects/{id}` - Delete project

### Analytics & Discovery
- `GET /api/stats/summary` - Dashboard summary statistics
- `GET /api/bookmarks/triage` - Bookmarks needing triage
- `GET /topics` - List all bookmark topics (legacy)

### Web Interface
- `GET /` - Dashboard homepage
- `GET /projects` - Projects overview page
- `GET /project-detail?topic={name}` - Interactive project detail page

## 📊 Data Model

### Bookmark Structure
```json
{
  "url": "https://example.com",
  "title": "Example Site",
  "description": "Optional description",
  "content": "Optional full content",
  "action": "working",
  "topic": "Development",
  "projectId": 1,
  "tags": ["web", "development"],
  "customProperties": {
    "priority": "high",
    "deadline": "2024-01-15"
  }
}
```

### Action Workflow
- **`read-later`** → Needs triage and decision
- **`working`** → Actively being used for a project
- **`share`** → Ready to be shared with others  
- **`archived`** → Completed/finished work
- **`irrelevant`** → Determined not useful

## 🗄️ Database

**SQLite** database with automatic migrations:
- **bookmarks** - Main bookmark storage
- **projects** - Normalized project management
- **Automatic timestamps** and **foreign key constraints**
- **WAL mode** for better concurrent access

### Database Commands
```bash
# Check migration status
migrate -path migrations -database sqlite3://bookmarks.db version

# Manual migration
migrate -path migrations -database sqlite3://bookmarks.db up
```

## 🧪 Testing

Comprehensive test suite covering database operations, HTTP handlers, and edge cases:

```bash
# Run all tests
go test

# Run with coverage
go test -cover

# Run specific test
go test -run TestBookmarkWorkflow_EndToEnd -v

# Run project-related tests only
go test -run "Projects" -v
```

**Current Coverage**: 71.5% with 30+ test functions

## 🌐 Frontend Development

The Vue.js frontend provides a modern interface for bookmark management:

```bash
cd frontend
npm install
npm run dev          # Development server
npm run build        # Production build
npm test            # Run frontend tests
```

**Features**: Real-time filtering, project management, responsive design, toast notifications

## 🔌 Browser Extension

Chrome extension for one-click bookmark saving:

```bash
cd extension
npm install
./build.sh          # Build extension
```

**Installation**: Load unpacked extension from `extension/` directory in Chrome Developer mode

## 🚀 Production Deployment

### Option 1: Binary Deployment
```bash
go build -o bookminderapi main.go
./bookminderapi
```

### Option 2: Docker (Create Dockerfile)
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o bookminderapi main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=builder /app/bookminderapi .
COPY --from=builder /app/migrations ./migrations
EXPOSE 9090
CMD ["./bookminderapi"]
```

### Option 3: Systemd Service
```ini
[Unit]
Description=BookMinder API Server
After=network.target

[Service]
Type=simple
User=bookminderapi
WorkingDirectory=/opt/bookminderapi
ExecStart=/opt/bookminderapi/bookminderapi
Restart=always

[Install]
WantedBy=multi-user.target
```

## ⚙️ Configuration

### Environment Variables
- `PORT` - Server port (default: 9090)
- `DB_PATH` - Database file path (default: bookmarks.db)
- `LOG_LEVEL` - Logging level (INFO, WARN, ERROR)

### Security Features
- **CORS configuration** for cross-origin requests
- **Security headers** (HSTS, XSS protection, content type options)
- **Input validation** and **SQL injection protection**
- **Request logging** and **error tracking**

## 🔍 Monitoring & Logs

**Console Logging**: Request details, validation errors, database operations
**File Logging**: Structured JSON logs in `bookminderapi.log`

```json
{
  "timestamp": "2023-12-10T15:30:45Z",
  "level": "INFO",
  "message": "Successfully saved bookmark",
  "component": "api",
  "data": {"bookmarkId": 123, "url": "https://example.com"}
}
```

## 🤝 Contributing

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Run tests**: `go test` and `cd frontend && npm test`
4. **Commit changes**: `git commit -m 'Add amazing feature'`
5. **Push to branch**: `git push origin feature/amazing-feature`
6. **Open a Pull Request**

### Development Guidelines
- Maintain single-file simplicity for Go backend
- Follow existing code patterns and conventions
- Add tests for new functionality
- Update documentation for API changes

## 📈 Performance

- **SQLite WAL mode** for concurrent access
- **Connection pooling** for database efficiency
- **Lightweight single-binary** deployment
- **Optimized queries** for dashboard statistics
- **Client-side filtering** for responsive UI

## 🐛 Troubleshooting

### Common Issues

**Database locked error**:
```bash
# Check for zombie processes
pkill -f bookminderapi
rm -f bookmarks.db-shm bookmarks.db-wal
```

**Migration errors**:
```bash
# Check current version
migrate -path migrations -database sqlite3://bookmarks.db version

# Force specific version
migrate -path migrations -database sqlite3://bookmarks.db force 4
```

**Port already in use**:
```bash
# Find process using port 9090
lsof -i :9090
kill -9 <PID>
```

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🔗 Related Projects

- **Frontend**: Vue.js 3 with Composition API and TypeScript
- **Extension**: Manifest V3 Chrome extension with modern APIs
- **Database**: SQLite with golang-migrate for schema management

---

**Made with ❤️ for developers who love organized bookmarks**