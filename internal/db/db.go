package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

var DB *sql.DB

// dbPath returns the path to the SQLite database file
func dbPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".pkt", "pkt2.db"), nil
}

// Connect establishes a connection to the SQLite database
func Connect() error {
	path, err := dbPath()
	if err != nil {
		return err
	}

	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	DB, err = sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Enable foreign key support
	if _, err := DB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Test connection
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations to ensure tables exist
	if err := RunMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// InitDB creates the database file and runs migrations
func InitDB() error {
	// Connect first (this will create the file if it doesn't exist)
	if err := Connect(); err != nil {
		return err
	}

	// Run migrations
	return RunMigrations()
}

// RunMigrations executes the schema SQL to create tables and indexes
func RunMigrations() error {
	if DB == nil {
		return fmt.Errorf("database connection not established")
	}

	_, err := DB.Exec(schemaSQL)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// SetConfig is kept for API compatibility but is no longer needed for SQLite
func SetConfig(cfg interface{}) {
	// No-op for SQLite - database path is fixed at ~/.pkt/pkt.db
}

// TestConnection tests the database connection without storing it
func TestConnection() error {
	path, err := dbPath()
	if err != nil {
		return err
	}

	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	testDB, err := sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = testDB.Close() }()

	// Test connection
	if err := testDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}
