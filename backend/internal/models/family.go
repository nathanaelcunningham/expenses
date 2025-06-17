package models

import (
	"time"
)

// Family represents a family group
type Family struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	InviteCode    string    `json:"invite_code"`
	DatabaseURL   string    `json:"database_url"`
	ManagerID     string    `json:"manager_id"`
	SchemaVersion int       `json:"schema_version"`
	Members       []Member  `json:"members,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Member represents a family member
type Member struct {
	UserID   string    `json:"user_id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
	IsActive bool      `json:"is_active"`
}

// FamilyMembership represents a user's membership in a family
type FamilyMembership struct {
	FamilyID string    `json:"family_id"`
	UserID   string    `json:"user_id"`
	Role     string    `json:"role"` // 'manager' or 'member'
	JoinedAt time.Time `json:"joined_at"`
}

// FamilyWithMembers represents a family with its member details
type FamilyWithMembers struct {
	Family      *Family             `json:"family"`
	Memberships []*FamilyMembership `json:"memberships"`
}

// CreateFamilyRequest represents the data needed to create a family
type CreateFamilyRequest struct {
	Name         string `json:"name"`
	ManagerID    string `json:"manager_id"`
	ManagerName  string `json:"manager_name"`
	ManagerEmail string `json:"manager_email"`
}

// JoinFamilyRequest represents a request to join a family
type JoinFamilyRequest struct {
	InviteCode string `json:"invite_code"`
	UserID     string `json:"user_id"`
	UserName   string `json:"user_name"`
	UserEmail  string `json:"user_email"`
}