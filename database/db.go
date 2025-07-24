package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" // PostgreSQL driver, imported for its side effects (init function)

	"github.com/anpsniper/anpbayu-be/config" // Import your config package with the correct module name
)

// DB is the global database connection pool.
var DB *sql.DB

// InitDatabase initializes the database connection pool.
// It takes the application configuration as input.
func InitDatabase(cfg *config.Config) error {
	// Use the single DBURL from the configuration.
	// This string should already be in a format suitable for sql.Open (e.g., "host=... user=...").
	connStr := cfg.DBURL

	var err error
	// Open a database connection. This does not establish a connection to the database server,
	// nor does it validate your connection string. Instead, it prepares the database
	// abstraction for later use.
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	// Ping the database to verify the connection is alive.
	// This actually attempts to connect to the database server.
	err = DB.Ping()
	if err != nil {
		// Close the database connection if the ping fails
		DB.Close()
		return fmt.Errorf("error connecting to the database: %w", err)
	}

	log.Println("Successfully connected to the database!")
	return nil
}

// CloseDatabase closes the database connection pool.
func CloseDatabase() {
	if DB != nil {
		err := DB.Close()
		if err != nil {
			log.Printf("Error closing database connection: %v", err)
		} else {
			log.Println("Database connection closed.")
		}
	}
}
