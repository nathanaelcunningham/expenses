package family

import (
	"time"
	"expenses-backend/internal/database/sql/masterdb"
)

type CreateFamilyRequest struct {
	Name         string `json:"name"`
	ManagerID    string `json:"manager_id"`
	ManagerName  string `json:"manager_name"`
	ManagerEmail string `json:"manager_email"`
}

func (r CreateFamilyRequest) Validate() error {
	if r.Name == "" || len(r.Name) > 100 {
		return ErrInvalidFamilyName
	}
	if r.ManagerID == "" {
		return &FamilyError{"INVALID_MANAGER", "Manager ID is required"}
	}
	return nil
}

type JoinFamilyRequest struct {
	InviteCode string `json:"invite_code"`
	UserID     string `json:"user_id"`
	UserName   string `json:"user_name"`
	UserEmail  string `json:"user_email"`
}

// FamilyResponse combines SQLC Family with additional computed data
type FamilyResponse struct {
	*masterdb.Family
	Members []MemberResponse `json:"members,omitempty"`
}

// MemberResponse combines user info with membership details
type MemberResponse struct {
	UserID   string    `json:"user_id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
	IsActive bool      `json:"is_active"`
}

