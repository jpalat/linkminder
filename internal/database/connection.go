package database

import (
	"database/sql"
	"fmt"
	"log"

	"bookminderapi/internal/config"
	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the database connection and provides repository methods
type DB struct {
	conn *sql.DB
}

// New creates a new database connection
func New(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	log.Printf("Database connection established: %s", dbPath)
	config.LogStructured("INFO", "database", "Database connection established", map[string]interface{}{
		"database_path": dbPath,
	})

	return &DB{conn: db}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// GetConn returns the underlying sql.DB connection for migrations
func (db *DB) GetConn() *sql.DB {
	return db.conn
}

// Ping tests the database connection
func (db *DB) Ping() error {
	return db.conn.Ping()
}

// ValidateDB validates the database connection
func (db *DB) ValidateDB() error {
	if db.conn == nil {
		return fmt.Errorf("database connection is nil")
	}
	
	if err := db.conn.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %v", err)
	}
	
	return nil
}