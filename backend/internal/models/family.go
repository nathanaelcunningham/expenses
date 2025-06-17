package models

import "time"

// Family represents a family unit with its metadata
type Family struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	InviteCode  string    `json:"invite_code"`
	DatabaseURL string    `json:"database_url"`
	ManagerID   string    `json:"manager_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Members     []Member  `json:"members,omitempty"`
}

// Member represents a family member
type Member struct {
	UserID   string    `json:"user_id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Role     string    `json:"role"` // "manager" or "member"
	JoinedAt time.Time `json:"joined_at"`
	IsActive bool      `json:"is_active"`
}

// CreateFamilyRequest represents the request to create a new family
type CreateFamilyRequest struct {
	Name         string `json:"name"`
	ManagerID    string `json:"manager_id"`
	ManagerName  string `json:"manager_name"`
	ManagerEmail string `json:"manager_email"`
}

// JoinFamilyRequest represents the request to join a family
type JoinFamilyRequest struct {
	InviteCode string `json:"invite_code"`
	UserID     string `json:"user_id"`
	UserName   string `json:"user_name"`
	UserEmail  string `json:"user_email"`
}

// UpdateFamilyRequest represents the request to update family details
type UpdateFamilyRequest struct {
	Name *string `json:"name,omitempty"`
}