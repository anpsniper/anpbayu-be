package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/anpsniper/anpbayu-be/config" // Import your config package
	_ "github.com/lib/pq"                    // PostgreSQL driver
)

// DB is the global database connection pool.
// It's exported so other packages (like models) can access it.
var DB *sql.DB

// InitDatabase initializes the PostgreSQL database connection and creates necessary tables.
func InitDatabase(cfg *config.Config) error {
	var err error
	// Implement a retry mechanism for database connection
	for i := 0; i < 5; i++ { // Try to connect 5 times
		DB, err = sql.Open("postgres", cfg.DBURL)
		if err == nil {
			err = DB.Ping() // Ping the database to verify the connection
			if err == nil {
				log.Println("Successfully connected to the database!")
				break
			}
		}
		log.Printf("Failed to connect to database (attempt %d/5): %v. Retrying in 5 seconds...", i+1, err)
		time.Sleep(5 * time.Second) // Wait before retrying
	}

	if err != nil {
		return fmt.Errorf("could not connect to the database after multiple retries: %w", err)
	}

	// Set connection pool settings
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(10)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// SQL DDL to create tables if they don't exist
	createTablesSQL := `
	-- Function to update 'updated_at' column
	CREATE OR REPLACE FUNCTION update_updated_at_column()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.updated_at = NOW();
		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	-- Create 'roles' table
	CREATE TABLE IF NOT EXISTS roles (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(50) UNIQUE NOT NULL,
		description TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	-- Trigger for 'roles' table
	DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_roles_updated_at') THEN
			CREATE TRIGGER update_roles_updated_at
			BEFORE UPDATE ON roles
			FOR EACH ROW
			EXECUTE FUNCTION update_updated_at_column();
		END IF;
	END $$;

	-- Create 'users' table
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		username VARCHAR(100) NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		role_id UUID NOT NULL, -- Foreign key to roles table
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT fk_users_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE RESTRICT
	);

	-- Trigger for 'users' table
	DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_users_updated_at') THEN
			CREATE TRIGGER update_users_updated_at
			BEFORE UPDATE ON users
			FOR EACH ROW
			EXECUTE FUNCTION update_updated_at_column();
		END IF;
	END $$;

	-- Create 'posts' table
	CREATE TABLE IF NOT EXISTS posts (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID NOT NULL,
		title VARCHAR(255) NOT NULL,
		content TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT fk_posts_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	-- Trigger for 'posts' table
	DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_posts_updated_at') THEN
			CREATE TRIGGER update_posts_updated_at
			BEFORE UPDATE ON posts
			FOR EACH ROW
			EXECUTE FUNCTION update_updated_at_column();
		END IF;
	END $$;

	-- Create 'comments' table
	CREATE TABLE IF NOT EXISTS comments (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		post_id UUID NOT NULL,
		user_id UUID NOT NULL,
		content TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT fk_comments_post FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
		CONSTRAINT fk_comments_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	-- Trigger for 'comments' table
	DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_comments_updated_at') THEN
			CREATE TRIGGER update_comments_updated_at
			BEFORE UPDATE ON comments
			FOR EACH ROW
			EXECUTE FUNCTION update_updated_at_column();
		END IF;
	END $$;

	-- Create 'sessions' table
	CREATE TABLE IF NOT EXISTS sessions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID NOT NULL,
		token VARCHAR(255) UNIQUE NOT NULL,
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT fk_sessions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	-- Create 'user_logs' table
	CREATE TABLE IF NOT EXISTS user_logs (
		id serial4 NOT NULL,
		user_id uuid NOT NULL,
		login_at timestamptz DEFAULT now() NOT NULL,
		logout_at timestamptz NULL,
		CONSTRAINT user_logs_pkey PRIMARY KEY (id),
		CONSTRAINT user_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES db_bayneta.users(id) ON DELETE CASCADE
	);
	`

	log.Println("Creating tables if they do not exist...")
	_, err = DB.Exec(createTablesSQL)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	log.Println("Tables created or already exist.")

	return nil
}

// CloseDatabase closes the global database connection.
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
