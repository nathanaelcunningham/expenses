package family

import "time"

// Temporary stub models for family service - these should be replaced with SQLC types

type CreateFamilyRequest struct {
	Name         string `json:"name"`
	ManagerID    string `json:"manager_id"`
	ManagerName  string `json:"manager_name"`
	ManagerEmail string `json:"manager_email"`
}

type JoinFamilyRequest struct {
	InviteCode string `json:"invite_code"`
	UserID     string `json:"user_id"`
	UserName   string `json:"user_name"`
	UserEmail  string `json:"user_email"`
}

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

type Member struct {
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
	IsActive bool      `json:"is_active"`
}