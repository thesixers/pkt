package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/genesix/pkt/internal/config"
	_ "github.com/lib/pq"
)

//go:embed schema.sql
var schemaSQL string

var DB *sql.DB
var dbConfig *config.Config

// SetConfig sets the database configuration
func SetConfig(cfg *config.Config) {
	dbConfig = cfg
}

// Connect establishes a connection to the PostgreSQL database
func Connect() error {
	// Get connection parameters from config or environment
	host := getEnv("PKT_DB_HOST", "127.0.0.1")
	port := getEnv("PKT_DB_PORT", "5432")
	user := getEnv("PKT_DB_USER", "postgres")
	password := getEnv("PKT_DB_PASSWORD", "")
	dbname := getEnv("PKT_DB_NAME", "pkt_db")

	if dbConfig != nil {
		if dbConfig.DBHost != "" {
			host = dbConfig.DBHost
		}
		if dbConfig.DBPort != "" {
			port = dbConfig.DBPort
		}
		if dbConfig.DBUser != "" {
			user = dbConfig.DBUser
		}
		if dbConfig.DBPassword != "" {
			password = dbConfig.DBPassword
		}
		if dbConfig.DBName != "" {
			dbname = dbConfig.DBName
		}
	}

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable",
		host, port, user, dbname)
	
	if password != "" {
		connStr += fmt.Sprintf(" password=%s", password)
	}

	// Open database connection
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	if err := DB.Ping(); err != nil {
		return FormatDBError(err)
	}

	return nil
}

// InitDB creates the database if it doesn't exist and runs migrations
func InitDB() error {
	// First, connect to postgres database to create pkt_db if needed
	host := getEnv("PKT_DB_HOST", "127.0.0.1")
	port := getEnv("PKT_DB_PORT", "5432")
	user := getEnv("PKT_DB_USER", "postgres")
	password := getEnv("PKT_DB_PASSWORD", "")
	dbname := getEnv("PKT_DB_NAME", "pkt_db")

	if dbConfig != nil {
		if dbConfig.DBHost != "" {
			host = dbConfig.DBHost
		}
		if dbConfig.DBPort != "" {
			port = dbConfig.DBPort
		}
		if dbConfig.DBUser != "" {
			user = dbConfig.DBUser
		}
		if dbConfig.DBPassword != "" {
			password = dbConfig.DBPassword
		}
		if dbConfig.DBName != "" {
			dbname = dbConfig.DBName
		}
	}

	// Connect to default postgres database
	connStr := fmt.Sprintf("host=%s port=%s user=%s dbname=postgres sslmode=disable",
		host, port, user)
	
	if password != "" {
		connStr += fmt.Sprintf(" password=%s", password)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Create database if it doesn't exist
	// Ignore error if database already exists - we'll check this by trying to connect to it
	_, _ = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))

	// Now connect to the pkt_db database
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

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// FormatDBError provides user-friendly error messages for common PostgreSQL errors
func FormatDBError(err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// Check for Ident authentication failure
	if strings.Contains(errMsg, "Ident authentication failed") || strings.Contains(errMsg, "no pg_hba.conf entry") {
		return fmt.Errorf(`PostgreSQL authentication failed.

⚠️  Please configure PostgreSQL authentication:

Option 1: Create a dedicated user and database:
  $ sudo -u postgres psql
  postgres=# CREATE USER pkt_user WITH PASSWORD 'yourpassword';
  postgres=# CREATE DATABASE pkt_db OWNER pkt_user;
  postgres=# \q

  Then set environment variables:
  export PKT_DB_USER=pkt_user
  export PKT_DB_PASSWORD=yourpassword
  export PKT_DB_NAME=pkt_db

Option 2: Configure pg_hba.conf to allow password authentication:
  Edit /etc/postgresql/*/main/pg_hba.conf and change:
    local   all   all   peer
  to:
    local   all   all   md5
  Then restart PostgreSQL: sudo systemctl restart postgresql`)
	}

	// Check for connection refused
	if strings.Contains(errMsg, "connection refused") {
		return fmt.Errorf(`PostgreSQL connection refused.

⚠️  PostgreSQL may not be running. Try:
  $ sudo systemctl start postgresql
  $ sudo systemctl enable postgresql

Or check if PostgreSQL is installed:
  $ psql --version`)
	}

	// Check for password authentication failed
	if strings.Contains(errMsg, "password authentication failed") {
		return fmt.Errorf(`PostgreSQL password authentication failed.

⚠️  Please check your credentials and set environment variables:
  export PKT_DB_USER=your_username
  export PKT_DB_PASSWORD=your_password
  export PKT_DB_NAME=pkt_db`)
	}

	// Check for database does not exist
	if strings.Contains(errMsg, "database") && strings.Contains(errMsg, "does not exist") {
		return fmt.Errorf("database does not exist - this is normal on first run, the database will be created automatically")
	}

	return fmt.Errorf("database error: %w", err)
}

// TestConnection tests the database connection without storing it
func TestConnection() error {
	host := getEnv("PKT_DB_HOST", "127.0.0.1")
	port := getEnv("PKT_DB_PORT", "5432")
	user := getEnv("PKT_DB_USER", "postgres")
	password := getEnv("PKT_DB_PASSWORD", "")
	dbname := getEnv("PKT_DB_NAME", "pkt_db")

	if dbConfig != nil {
		if dbConfig.DBHost != "" {
			host = dbConfig.DBHost
		}
		if dbConfig.DBPort != "" {
			port = dbConfig.DBPort
		}
		if dbConfig.DBUser != "" {
			user = dbConfig.DBUser
		}
		if dbConfig.DBPassword != "" {
			password = dbConfig.DBPassword
		}
		if dbConfig.DBName != "" {
			dbname = dbConfig.DBName
		}
	}

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable",
		host, port, user, dbname)
	
	if password != "" {
		connStr += fmt.Sprintf(" password=%s", password)
	}

	// Open database connection
	testDB, err := sql.Open("postgres", connStr)
	if err != nil {
		return FormatDBError(err)
	}
	defer func() { _ = testDB.Close() }()

	// Test connection
	if err := testDB.Ping(); err != nil {
		return FormatDBError(err)
	}

	return nil
}
