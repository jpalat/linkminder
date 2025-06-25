package database

import (
	"fmt"
	"log"

	"bookminderapi/internal/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations executes database migrations
func (db *DB) RunMigrations() error {
	if db.conn == nil {
		return fmt.Errorf("database connection is nil")
	}

	log.Printf("Starting database migrations...")
	config.LogStructured("INFO", "database", "Starting database migrations", nil)

	// Create SQLite3 migration driver
	driver, err := sqlite3.WithInstance(db.conn, &sqlite3.Config{})
	if err != nil {
		config.LogStructured("ERROR", "database", "Failed to create migration driver", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create migration driver: %v", err)
	}

	// Create migration instance
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"sqlite3",
		driver,
	)
	if err != nil {
		config.LogStructured("ERROR", "database", "Failed to create migration instance", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create migration instance: %v", err)
	}

	// Run migrations
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		config.LogStructured("ERROR", "database", "Migration failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("migration failed: %v", err)
	}

	// Get current migration version and status
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		config.LogStructured("WARN", "database", "Failed to get migration version", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		if err == migrate.ErrNilVersion {
			log.Printf("No migrations applied yet")
			config.LogStructured("INFO", "database", "No migrations applied yet", nil)
		} else {
			log.Printf("Current migration version: %d (dirty: %t)", version, dirty)
			config.LogStructured("INFO", "database", "Migration status", map[string]interface{}{
				"version": version,
				"dirty":   dirty,
			})
		}
	}

	if err == migrate.ErrNoChange {
		log.Printf("No new migrations to apply")
		config.LogStructured("INFO", "database", "No new migrations to apply", nil)
	} else {
		log.Printf("Database migrations completed successfully")
		config.LogStructured("INFO", "database", "Database migrations completed successfully", nil)
	}

	return nil
}