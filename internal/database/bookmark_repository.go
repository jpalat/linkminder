package database

import (
	"fmt"
	"log"

	"bookminderapi/internal/config"
	"bookminderapi/internal/models"
)

// SaveBookmark saves a bookmark to the database
func (db *DB) SaveBookmark(req models.BookmarkRequest) error {
	// Validate database connection
	if err := db.ValidateDB(); err != nil {
		return fmt.Errorf("failed to validate database connection: %v", err)
	}

	log.Printf("Saving bookmark to database: %s", req.URL)
	
	config.LogStructured("INFO", "database", "Saving bookmark", map[string]interface{}{
		"url":            req.URL,
		"title":          req.Title,
		"action":         req.Action,
		"content_length": len(req.Content),
	})
	
	// Convert tags and custom properties to JSON
	tagsJSON := TagsToJSON(req.Tags)
	customPropsJSON := CustomPropsToJSON(req.CustomProperties)

	insertSQL := `
	INSERT INTO bookmarks (url, title, description, content, action, shareTo, topic, tags, custom_properties)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	result, err := db.conn.Exec(insertSQL, req.URL, req.Title, req.Description, req.Content, req.Action, req.ShareTo, req.Topic, tagsJSON, customPropsJSON)
	if err != nil {
		log.Printf("Failed to insert bookmark: %v", err)
		config.LogStructured("ERROR", "database", "Insert failed", map[string]interface{}{
			"error": err.Error(),
			"url":   req.URL,
		})
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Failed to get last insert ID: %v", err)
		config.LogStructured("WARN", "database", "Failed to get insert ID", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}
	
	log.Printf("Successfully saved bookmark with ID: %d", id)
	config.LogStructured("INFO", "database", "Bookmark saved", map[string]interface{}{
		"id":    id,
		"url":   req.URL,
		"title": req.Title,
	})
	
	return nil
}

// GetTopics gets distinct topics from bookmarks
func (db *DB) GetTopics() ([]string, error) {
	log.Printf("Reading topics from database")
	
	config.LogStructured("INFO", "database", "Querying topics", nil)
	
	querySQL := `SELECT DISTINCT topic FROM bookmarks WHERE topic IS NOT NULL AND topic != '' ORDER BY topic`
	
	rows, err := db.conn.Query(querySQL)
	if err != nil {
		log.Printf("Failed to query topics: %v", err)
		config.LogStructured("ERROR", "database", "Topics query failed", map[string]interface{}{
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
			config.LogStructured("ERROR", "database", "Topic scan failed", map[string]interface{}{
				"error": err.Error(),
			})
			return nil, err
		}
		topics = append(topics, topic)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating topic rows: %v", err)
		config.LogStructured("ERROR", "database", "Topic iteration failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	log.Printf("Successfully retrieved %d topics", len(topics))
	config.LogStructured("INFO", "database", "Topics retrieved successfully", map[string]interface{}{
		"count":  len(topics),
		"topics": topics,
	})

	return topics, nil
}

// GetTriageQueue gets bookmarks that need triage
func (db *DB) GetTriageQueue(limit, offset int) (*models.TriageResponse, error) {
	log.Printf("Getting triage queue: limit=%d, offset=%d", limit, offset)
	
	config.LogStructured("INFO", "database", "Querying triage queue", map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	})

	// First get the total count
	countSQL := `SELECT COUNT(*) FROM bookmarks WHERE action IS NULL OR action = '' OR action = 'read-later'`
	var total int
	err := db.conn.QueryRow(countSQL).Scan(&total)
	if err != nil {
		log.Printf("Failed to get triage count: %v", err)
		config.LogStructured("ERROR", "database", "Triage count query failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	// Then get the bookmarks
	querySQL := `
		SELECT id, url, title, description, timestamp, topic, action, shareTo, tags, custom_properties
		FROM bookmarks 
		WHERE action IS NULL OR action = '' OR action = 'read-later'
		ORDER BY timestamp DESC 
		LIMIT ? OFFSET ?`
	
	rows, err := db.conn.Query(querySQL, limit, offset)
	if err != nil {
		log.Printf("Failed to query triage queue: %v", err)
		config.LogStructured("ERROR", "database", "Triage queue query failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}
	defer rows.Close()

	var bookmarks []models.TriageBookmark
	for rows.Next() {
		var bookmark models.TriageBookmark
		var tagsJSON, customPropsJSON string
		
		err := rows.Scan(
			&bookmark.ID, &bookmark.URL, &bookmark.Title, &bookmark.Description,
			&bookmark.Timestamp, &bookmark.Topic, &bookmark.Action, &bookmark.ShareTo,
			&tagsJSON, &customPropsJSON,
		)
		if err != nil {
			log.Printf("Failed to scan triage bookmark: %v", err)
			config.LogStructured("ERROR", "database", "Triage bookmark scan failed", map[string]interface{}{
				"error": err.Error(),
			})
			return nil, err
		}

		// Process the bookmark data
		bookmark.Domain = ExtractDomain(bookmark.URL)
		bookmark.Age = CalculateAge(bookmark.Timestamp)
		bookmark.Suggested = GetSuggestedAction(bookmark.Domain, bookmark.Title, bookmark.Description)
		bookmark.Tags = TagsFromJSON(tagsJSON)
		bookmark.CustomProperties = CustomPropsFromJSON(customPropsJSON)

		bookmarks = append(bookmarks, bookmark)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating triage rows: %v", err)
		config.LogStructured("ERROR", "database", "Triage iteration failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	response := &models.TriageResponse{
		Bookmarks: bookmarks,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	}

	log.Printf("Successfully retrieved %d triage bookmarks (total: %d)", len(bookmarks), total)
	config.LogStructured("INFO", "database", "Triage queue retrieved successfully", map[string]interface{}{
		"count": len(bookmarks),
		"total": total,
	})

	return response, nil
}

// GetBookmarksByAction gets bookmarks filtered by action type
func (db *DB) GetBookmarksByAction(action string, limit, offset int) (*models.TriageResponse, error) {
	log.Printf("Getting bookmarks by action: %s, limit=%d, offset=%d", action, limit, offset)
	
	config.LogStructured("INFO", "database", "Querying bookmarks by action", map[string]interface{}{
		"action": action,
		"limit":  limit,
		"offset": offset,
	})

	// First get the total count
	countSQL := `SELECT COUNT(*) FROM bookmarks WHERE action = ?`
	var total int
	err := db.conn.QueryRow(countSQL, action).Scan(&total)
	if err != nil {
		log.Printf("Failed to get bookmark count for action %s: %v", action, err)
		config.LogStructured("ERROR", "database", "Bookmark count query failed", map[string]interface{}{
			"error":  err.Error(),
			"action": action,
		})
		return nil, err
	}

	// Then get the bookmarks
	querySQL := `
		SELECT id, url, title, description, timestamp, topic, action, shareTo, tags, custom_properties
		FROM bookmarks 
		WHERE action = ?
		ORDER BY timestamp DESC 
		LIMIT ? OFFSET ?`
	
	rows, err := db.conn.Query(querySQL, action, limit, offset)
	if err != nil {
		log.Printf("Failed to query bookmarks by action %s: %v", action, err)
		config.LogStructured("ERROR", "database", "Bookmarks by action query failed", map[string]interface{}{
			"error":  err.Error(),
			"action": action,
		})
		return nil, err
	}
	defer rows.Close()

	var bookmarks []models.TriageBookmark
	for rows.Next() {
		var bookmark models.TriageBookmark
		var tagsJSON, customPropsJSON string
		
		err := rows.Scan(
			&bookmark.ID, &bookmark.URL, &bookmark.Title, &bookmark.Description,
			&bookmark.Timestamp, &bookmark.Topic, &bookmark.Action, &bookmark.ShareTo,
			&tagsJSON, &customPropsJSON,
		)
		if err != nil {
			log.Printf("Failed to scan bookmark: %v", err)
			config.LogStructured("ERROR", "database", "Bookmark scan failed", map[string]interface{}{
				"error": err.Error(),
			})
			return nil, err
		}

		// Process the bookmark data
		bookmark.Domain = ExtractDomain(bookmark.URL)
		bookmark.Age = CalculateAge(bookmark.Timestamp)
		bookmark.Suggested = GetSuggestedAction(bookmark.Domain, bookmark.Title, bookmark.Description)
		bookmark.Tags = TagsFromJSON(tagsJSON)
		bookmark.CustomProperties = CustomPropsFromJSON(customPropsJSON)

		bookmarks = append(bookmarks, bookmark)
	}

	response := &models.TriageResponse{
		Bookmarks: bookmarks,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	}

	log.Printf("Successfully retrieved %d bookmarks for action %s (total: %d)", len(bookmarks), action, total)
	config.LogStructured("INFO", "database", "Bookmarks by action retrieved successfully", map[string]interface{}{
		"count":  len(bookmarks),
		"action": action,
		"total":  total,
	})

	return response, nil
}

// GetBookmarkByID retrieves a single bookmark by its ID
func (db *DB) GetBookmarkByID(id int) (*models.ProjectBookmark, error) {
	log.Printf("Getting bookmark by ID: %d", id)
	
	config.LogStructured("INFO", "database", "Querying bookmark by ID", map[string]interface{}{
		"id": id,
	})

	querySQL := `
		SELECT id, url, title, description, content, timestamp, action, topic, shareTo, tags, custom_properties
		FROM bookmarks 
		WHERE id = ?`
	
	var bookmark models.ProjectBookmark
	var tagsJSON, customPropsJSON string
	
	err := db.conn.QueryRow(querySQL, id).Scan(
		&bookmark.ID, &bookmark.URL, &bookmark.Title, &bookmark.Description,
		&bookmark.Content, &bookmark.Timestamp, &bookmark.Action, &bookmark.Topic,
		&bookmark.ShareTo, &tagsJSON, &customPropsJSON,
	)
	
	if err != nil {
		log.Printf("Failed to get bookmark with ID %d: %v", id, err)
		config.LogStructured("ERROR", "database", "Bookmark by ID query failed", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return nil, err
	}

	// Process the bookmark data
	bookmark.Domain = ExtractDomain(bookmark.URL)
	bookmark.Age = CalculateAge(bookmark.Timestamp)
	bookmark.Tags = TagsFromJSON(tagsJSON)
	bookmark.CustomProperties = CustomPropsFromJSON(customPropsJSON)

	log.Printf("Successfully retrieved bookmark with ID: %d", id)
	config.LogStructured("INFO", "database", "Bookmark by ID retrieved successfully", map[string]interface{}{
		"id":    id,
		"title": bookmark.Title,
	})

	return &bookmark, nil
}