package family

import (
	"expenses-backend/internal/database/sql/masterdb"
	"time"
)

type CreateFamilyRequest struct {
	Name         string `json:"name"`
	ManagerID    int64  `json:"manager_id"`
	ManagerName  string `json:"manager_name"`
	ManagerEmail string `json:"manager_email"`
}

func (r CreateFamilyRequest) Validate() error {
	if r.Name == "" || len(r.Name) > 100 {
		return ErrInvalidFamilyName
	}
	if r.ManagerID == 0 {
		return &FamilyError{"INVALID_MANAGER", "Manager ID is required"}
	}
	return nil
}

type JoinFamilyRequest struct {
	InviteCode string `json:"invite_code"`
	UserID     int64  `json:"user_id"`
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
	UserID   int64     `json:"user_id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
	IsActive bool      `json:"is_active"`
}
