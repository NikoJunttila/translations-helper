package database

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

// DB is the database connection wrapper
type DB struct {
	conn *sql.DB
}

// NewDB creates a new database connection
func NewDB() (*DB, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable not set")
	}

	// Determine driver based on URL
	var driver string
	var dsn string

	if strings.HasPrefix(dbURL, "file:") {
		// Local SQLite file
		driver = "sqlite"
		dsn = strings.TrimPrefix(dbURL, "file:")
	} else if strings.HasPrefix(dbURL, "libsql://") {
		// TursoDB / libSQL
		driver = "libsql"
		authToken := os.Getenv("DATABASE_AUTH_TOKEN")
		if authToken != "" {
			dsn = fmt.Sprintf("%s?authToken=%s", dbURL, authToken)
		} else {
			dsn = dbURL
		}
	} else {
		return nil, fmt.Errorf("unsupported DATABASE_URL format: %s", dbURL)
	}

	conn, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// GetConn returns the underlying database connection
func (db *DB) GetConn() *sql.DB {
	return db.conn
}
