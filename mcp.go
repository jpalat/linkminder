package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func setupMCPServer() *server.MCPServer {
	s := server.NewMCPServer(
		"BookMinder",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	s.AddTool(toolGetStats(), handleMCPGetStats)
	s.AddTool(toolListBookmarks(), handleMCPListBookmarks)
	s.AddTool(toolGetBookmark(), handleMCPGetBookmark)
	s.AddTool(toolUpdateBookmark(), handleMCPUpdateBookmark)
	s.AddTool(toolBulkUpdateBookmarks(), handleMCPBulkUpdateBookmarks)
	s.AddTool(toolListProjects(), handleMCPListProjects)
	s.AddTool(toolCreateProject(), handleMCPCreateProject)

	return s
}

func startMCPServer(mcpServer *server.MCPServer) {
	port := os.Getenv("MCP_PORT")
	if port == "" {
		port = "9091"
	}
	baseURL := fmt.Sprintf("http://localhost:%s", port)
	sseServer := server.NewSSEServer(mcpServer, server.WithBaseURL(baseURL))
	log.Printf("MCP server starting on %s", baseURL)
	if err := sseServer.Start(":" + port); err != nil {
		log.Printf("MCP server error: %v", err)
	}
}

// ── Tool definitions ──────────────────────────────────────────────────────────

func toolGetStats() mcp.Tool {
	return mcp.NewTool("get_stats",
		mcp.WithDescription("Get bookmark dashboard statistics: how many need triage, active project count, ready-to-share count, and total bookmarks."),
	)
}

func toolListBookmarks() mcp.Tool {
	return mcp.NewTool("list_bookmarks",
		mcp.WithDescription("List bookmarks with optional filtering. Use action='read-later' or action='' to see the triage queue. Returns id, title, url, description, ai_summary, tags, action, domain, and age."),
		mcp.WithString("action",
			mcp.Description("Filter by action: read-later, working, share, archived, irrelevant. Omit for triage queue (no action set)."),
		),
		mcp.WithString("search",
			mcp.Description("Text search across title, description, and URL."),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum results to return (default 20, max 100)."),
		),
		mcp.WithNumber("offset",
			mcp.Description("Offset for pagination (default 0)."),
		),
	)
}

func toolGetBookmark() mcp.Tool {
	return mcp.NewTool("get_bookmark",
		mcp.WithDescription("Get full details of a single bookmark by ID, including content, tags, custom properties, and current action."),
		mcp.WithNumber("id",
			mcp.Description("Bookmark ID."),
			mcp.Required(),
		),
	)
}

func toolUpdateBookmark() mcp.Tool {
	return mcp.NewTool("update_bookmark",
		mcp.WithDescription("Update a bookmark's action, tags, project assignment, and/or AI summary. Provide only the fields you want to change."),
		mcp.WithNumber("id",
			mcp.Description("Bookmark ID to update."),
			mcp.Required(),
		),
		mcp.WithString("action",
			mcp.Description("New action: read-later, working, share, archived, irrelevant."),
		),
		mcp.WithString("project",
			mcp.Description("Project name to assign this bookmark to (must match an existing project name exactly)."),
		),
		mcp.WithString("ai_summary",
			mcp.Description("Short 1-2 sentence summary of what this bookmark is about."),
		),
		mcp.WithString("tags",
			mcp.Description("Comma-separated list of tags to set on this bookmark."),
		),
	)
}

func toolBulkUpdateBookmarks() mcp.Tool {
	return mcp.NewTool("bulk_update_bookmarks",
		mcp.WithDescription("Update the action (and optionally project) for multiple bookmarks at once. Efficient for batch triage decisions."),
		mcp.WithString("ids",
			mcp.Description("Comma-separated list of bookmark IDs to update."),
			mcp.Required(),
		),
		mcp.WithString("action",
			mcp.Description("Action to apply: read-later, working, share, archived, irrelevant."),
			mcp.Required(),
		),
		mcp.WithString("project",
			mcp.Description("Optional project name to assign all updated bookmarks to."),
		),
	)
}

func toolListProjects() mcp.Tool {
	return mcp.NewTool("list_projects",
		mcp.WithDescription("List all active projects with their bookmark counts and last-updated timestamps. Use this to see available project names before assigning bookmarks."),
	)
}

func toolCreateProject() mcp.Tool {
	return mcp.NewTool("create_project",
		mcp.WithDescription("Create a new project. Use this when a cluster of bookmarks belongs to a topic that doesn't have a project yet."),
		mcp.WithString("name",
			mcp.Description("Project name (must be unique)."),
			mcp.Required(),
		),
		mcp.WithString("description",
			mcp.Description("Optional description of what this project is about."),
		),
	)
}

// ── Handlers ──────────────────────────────────────────────────────────────────

func handleMCPGetStats(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	stats, err := getStatsSummary()
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("error: %v", err)), nil
	}
	out, _ := json.MarshalIndent(stats, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

func handleMCPListBookmarks(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	action, _ := args["action"].(string)
	search, _ := args["search"].(string)
	limit := 20
	offset := 0
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
		if limit > 100 {
			limit = 100
		}
	}
	if o, ok := args["offset"].(float64); ok && o >= 0 {
		offset = int(o)
	}

	type row struct {
		ID          int      `json:"id"`
		Title       string   `json:"title"`
		URL         string   `json:"url"`
		Description string   `json:"description,omitempty"`
		AISummary   string   `json:"ai_summary,omitempty"`
		Tags        []string `json:"tags,omitempty"`
		Action      string   `json:"action,omitempty"`
		Domain      string   `json:"domain,omitempty"`
		Age         string   `json:"age,omitempty"`
		Topic       string   `json:"topic,omitempty"`
	}

	where := []string{"(b.deleted = FALSE OR b.deleted IS NULL)"}
	params := []interface{}{}

	if action == "" {
		where = append(where, "(b.action IS NULL OR b.action = '' OR b.action = 'read-later')")
	} else {
		where = append(where, "b.action = ?")
		params = append(params, action)
	}
	if search != "" {
		where = append(where, "(b.title LIKE ? OR b.description LIKE ? OR b.url LIKE ?)")
		s := "%" + search + "%"
		params = append(params, s, s, s)
	}

	query := fmt.Sprintf(`
		SELECT b.id, b.title, b.url,
		       COALESCE(b.description,''), COALESCE(b.ai_summary,''),
		       COALESCE(b.tags,'[]'), COALESCE(b.action,''),
		       b.timestamp, COALESCE(b.topic,'')
		FROM bookmarks b
		WHERE %s
		ORDER BY b.timestamp DESC
		LIMIT ? OFFSET ?
	`, strings.Join(where, " AND "))
	params = append(params, limit, offset)

	rows, err := db.Query(query, params...)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("error querying bookmarks: %v", err)), nil
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Failed to close rows: %v", err)
		}
	}()

	var results []row
	for rows.Next() {
		var r row
		var tagsJSON, ts string
		if err := rows.Scan(&r.ID, &r.Title, &r.URL, &r.Description, &r.AISummary,
			&tagsJSON, &r.Action, &ts, &r.Topic); err != nil {
			continue
		}
		json.Unmarshal([]byte(tagsJSON), &r.Tags) //nolint
		r.Domain = extractDomain(r.URL)
		r.Age = formatAge(ts)
		results = append(results, r)
	}
	if results == nil {
		results = []row{}
	}

	var total int
	if err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM bookmarks b WHERE %s",
		strings.Join(where, " AND ")), params[:len(params)-2]...).Scan(&total); err != nil {
		log.Printf("Failed to count bookmarks: %v", err)
	}

	out, _ := json.MarshalIndent(map[string]interface{}{
		"bookmarks": results,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	}, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

func handleMCPGetBookmark(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	idF, ok := args["id"].(float64)
	if !ok {
		return mcp.NewToolResultText("error: id is required"), nil
	}
	id := int(idF)

	bookmark, err := mcpGetBookmarkFull(id)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("error: %v", err)), nil
	}
	if bookmark == nil {
		return mcp.NewToolResultText(fmt.Sprintf("error: bookmark %d not found", id)), nil
	}

	out, _ := json.MarshalIndent(bookmark, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

func handleMCPUpdateBookmark(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	idF, ok := args["id"].(float64)
	if !ok {
		return mcp.NewToolResultText("error: id is required"), nil
	}
	id := int(idF)

	sets := []string{}
	params := []interface{}{}

	if action, ok := args["action"].(string); ok && action != "" {
		valid := map[string]bool{"read-later": true, "working": true, "share": true, "archived": true, "irrelevant": true}
		if !valid[action] {
			return mcp.NewToolResultText(fmt.Sprintf("error: invalid action %q", action)), nil
		}
		sets = append(sets, "action = ?")
		params = append(params, action)
	}

	if projectName, ok := args["project"].(string); ok && projectName != "" {
		var projectID int
		err := db.QueryRow("SELECT id FROM projects WHERE name = ?", projectName).Scan(&projectID)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("error: project %q not found — use list_projects to see available names", projectName)), nil
		}
		sets = append(sets, "project_id = ?", "topic = ?")
		params = append(params, projectID, projectName)
	}

	if summary, ok := args["ai_summary"].(string); ok && summary != "" {
		sets = append(sets, "ai_summary = ?")
		params = append(params, summary)
	}

	if tagsStr, ok := args["tags"].(string); ok && tagsStr != "" {
		tags := splitTrimmed(tagsStr, ",")
		tagsJSON, _ := json.Marshal(tags)
		sets = append(sets, "tags = ?")
		params = append(params, string(tagsJSON))
	}

	if len(sets) == 0 {
		return mcp.NewToolResultText("error: no fields to update provided"), nil
	}

	params = append(params, id)
	_, err := db.Exec(
		fmt.Sprintf("UPDATE bookmarks SET %s WHERE id = ?", strings.Join(sets, ", ")),
		params...,
	)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("error: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("bookmark %d updated", id)), nil
}

func handleMCPBulkUpdateBookmarks(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	idsStr, ok := args["ids"].(string)
	if !ok || idsStr == "" {
		return mcp.NewToolResultText("error: ids is required"), nil
	}
	action, ok := args["action"].(string)
	if !ok || action == "" {
		return mcp.NewToolResultText("error: action is required"), nil
	}
	valid := map[string]bool{"read-later": true, "working": true, "share": true, "archived": true, "irrelevant": true}
	if !valid[action] {
		return mcp.NewToolResultText(fmt.Sprintf("error: invalid action %q", action)), nil
	}

	ids := splitTrimmed(idsStr, ",")
	if len(ids) == 0 {
		return mcp.NewToolResultText("error: no valid IDs provided"), nil
	}

	projectName, _ := args["project"].(string)
	var projectID int
	if projectName != "" {
		err := db.QueryRow("SELECT id FROM projects WHERE name = ?", projectName).Scan(&projectID)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("error: project %q not found", projectName)), nil
		}
	}

	updated := 0
	failed := 0
	for _, idStr := range ids {
		var id int
		if _, err := fmt.Sscanf(strings.TrimSpace(idStr), "%d", &id); err != nil {
			failed++
			continue
		}
		if projectID > 0 {
			_, err := db.Exec("UPDATE bookmarks SET action = ?, project_id = ?, topic = ? WHERE id = ?",
				action, projectID, projectName, id)
			if err != nil {
				failed++
			} else {
				updated++
			}
		} else {
			_, err := db.Exec("UPDATE bookmarks SET action = ? WHERE id = ?", action, id)
			if err != nil {
				failed++
			} else {
				updated++
			}
		}
	}

	return mcp.NewToolResultText(fmt.Sprintf("updated %d bookmarks to action=%q%s (%d failed)",
		updated, action,
		func() string {
			if projectName != "" {
				return fmt.Sprintf(", project=%q", projectName)
			}
			return ""
		}(),
		failed,
	)), nil
}

func handleMCPListProjects(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	type projectRow struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		Status      string `json:"status"`
		LinkCount   int    `json:"linkCount"`
		LastUpdated string `json:"lastUpdated"`
	}

	rows, err := db.Query(`
		SELECT p.id, p.name, COALESCE(p.description,''), p.status,
		       COUNT(b.id) as link_count,
		       COALESCE(MAX(b.timestamp), p.updated_at) as last_updated
		FROM projects p
		LEFT JOIN bookmarks b ON b.project_id = p.id AND (b.deleted = FALSE OR b.deleted IS NULL)
		WHERE p.status != 'deleted'
		GROUP BY p.id
		ORDER BY last_updated DESC
	`)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("error: %v", err)), nil
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Failed to close rows: %v", err)
		}
	}()

	var projects []projectRow
	for rows.Next() {
		var p projectRow
		var lastUpdated string
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Status, &p.LinkCount, &lastUpdated); err != nil {
			continue
		}
		p.LastUpdated = formatAge(lastUpdated)
		projects = append(projects, p)
	}
	if projects == nil {
		projects = []projectRow{}
	}

	out, _ := json.MarshalIndent(map[string]interface{}{"projects": projects}, "", "  ")
	return mcp.NewToolResultText(string(out)), nil
}

func handleMCPCreateProject(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return mcp.NewToolResultText("error: name is required"), nil
	}
	description, _ := args["description"].(string)

	now := time.Now().UTC().Format(time.RFC3339)
	res, err := db.Exec(
		"INSERT INTO projects (name, description, status, created_at, updated_at) VALUES (?, ?, 'active', ?, ?)",
		name, description, now, now,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return mcp.NewToolResultText(fmt.Sprintf("error: project %q already exists", name)), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("error: %v", err)), nil
	}

	id, _ := res.LastInsertId()
	return mcp.NewToolResultText(fmt.Sprintf("created project %q with id=%d", name, id)), nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func formatAge(ts string) string {
	if ts == "" {
		return ""
	}
	var t time.Time
	var err error
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05Z", "2006-01-02 15:04:05"} {
		t, err = time.Parse(layout, ts)
		if err == nil {
			break
		}
	}
	if err != nil {
		return ""
	}
	d := time.Since(t)
	switch {
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dw", int(d.Hours()/(24*7)))
	default:
		return fmt.Sprintf("%dmo", int(d.Hours()/(24*30)))
	}
}

func splitTrimmed(s, sep string) []string {
	parts := strings.Split(s, sep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// getBookmarkByIDFull returns a bookmark for MCP with all fields including content.
// Reuses the existing getBookmarkByID which already returns ProjectBookmark.
func mcpGetBookmarkFull(id int) (interface{}, error) {
	type fullBookmark struct {
		ID               int               `json:"id"`
		URL              string            `json:"url"`
		Title            string            `json:"title"`
		Description      string            `json:"description,omitempty"`
		Content          string            `json:"content,omitempty"`
		AISummary        string            `json:"ai_summary,omitempty"`
		Action           string            `json:"action,omitempty"`
		Topic            string            `json:"topic,omitempty"`
		ProjectID        int               `json:"project_id,omitempty"`
		Tags             []string          `json:"tags,omitempty"`
		CustomProperties map[string]string `json:"custom_properties,omitempty"`
		Domain           string            `json:"domain"`
		Age              string            `json:"age"`
		Timestamp        string            `json:"timestamp"`
	}

	row := db.QueryRow(`
		SELECT id, url, title,
		       COALESCE(description,''), COALESCE(content,''), COALESCE(ai_summary,''),
		       COALESCE(action,''), COALESCE(topic,''), COALESCE(project_id,0),
		       COALESCE(tags,'[]'), COALESCE(custom_properties,'{}'),
		       timestamp
		FROM bookmarks
		WHERE id = ? AND (deleted = FALSE OR deleted IS NULL)
	`, id)

	var b fullBookmark
	var tagsJSON, customJSON, ts string
	var projectID sql.NullInt64
	err := row.Scan(
		&b.ID, &b.URL, &b.Title,
		&b.Description, &b.Content, &b.AISummary,
		&b.Action, &b.Topic, &projectID,
		&tagsJSON, &customJSON, &ts,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if projectID.Valid {
		b.ProjectID = int(projectID.Int64)
	}
	json.Unmarshal([]byte(tagsJSON), &b.Tags)               //nolint
	json.Unmarshal([]byte(customJSON), &b.CustomProperties) //nolint
	b.Domain = extractDomain(b.URL)
	b.Age = formatAge(ts)
	b.Timestamp = ts
	return &b, nil
}
