package models

import "time"

// UserLog represents a single login/logout session record in the database.
type UserLog struct {
	ID       int        `json:"id"`        // Auto-incrementing primary key
	UserID   string     `json:"user_id"`   // Foreign key to the users table
	LoginAt  time.Time  `json:"login_at"`  // Timestamp of login
	LogoutAt *time.Time `json:"logout_at"` // Nullable timestamp of logout
}
