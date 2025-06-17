package models

import "time"

// User represents a user in the system
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"` // Never expose password hash in JSON
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserSession represents an active user session
type UserSession struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	FamilyID   *string   `json:"family_id,omitempty"`
	UserRole   *string   `json:"user_role,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	LastActive time.Time `json:"last_active"`
	ExpiresAt  time.Time `json:"expires_at"`
	UserAgent  *string   `json:"user_agent,omitempty"`
	IPAddress  *string   `json:"ip_address,omitempty"`
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// UpdateUserRequest represents the request to update user details
type UpdateUserRequest struct {
	Name         *string `json:"name,omitempty"`
	Email        *string `json:"email,omitempty"`
	PasswordHash *string `json:"password_hash,omitempty"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SessionValidationResult contains session validation results
type SessionValidationResult struct {
	Valid    bool         `json:"valid"`
	Session  *UserSession `json:"session,omitempty"`
	User     *User        `json:"user,omitempty"`
	FamilyID *string      `json:"family_id,omitempty"`
}