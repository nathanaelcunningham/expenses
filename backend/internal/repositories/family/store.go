package family

import (
	"context"
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
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
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
	Name      string `json:"name"`
	ManagerID string `json:"manager_id"`
}

// Store defines the interface for family repository operations
type Store interface {
	// Family operations
	CreateFamily(ctx context.Context, req *CreateFamilyRequest) (*Family, error)
	GetFamilyByID(ctx context.Context, familyID string) (*Family, error)
	GetFamilyByInviteCode(ctx context.Context, inviteCode string) (*Family, error)
	UpdateFamily(ctx context.Context, family *Family) error
	DeleteFamily(ctx context.Context, familyID string) error
	GetFamiliesByManagerID(ctx context.Context, managerID string) ([]*Family, error)

	// Membership operations
	AddFamilyMember(ctx context.Context, familyID, userID, role string) error
	RemoveFamilyMember(ctx context.Context, familyID, userID string) error
	GetFamilyMemberships(ctx context.Context, familyID string) ([]*FamilyMembership, error)
	GetUserFamilyMembership(ctx context.Context, userID string) (*FamilyMembership, error)
	UpdateMemberRole(ctx context.Context, familyID, userID, newRole string) error

	// Combined operations
	GetFamilyWithMembers(ctx context.Context, familyID string) (*FamilyWithMembers, error)

	// Utility operations
	FamilyExists(ctx context.Context, familyID string) (bool, error)
	UserIsFamilyMember(ctx context.Context, familyID, userID string) (bool, error)
	UserIsFamilyManager(ctx context.Context, familyID, userID string) (bool, error)
	GenerateInviteCode(ctx context.Context) (string, error)
}