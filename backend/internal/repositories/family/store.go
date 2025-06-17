package family

import (
	"context"
	"expenses-backend/internal/models"
)

type Store interface {
	// Family operations
	CreateFamily(ctx context.Context, req *models.CreateFamilyRequest) (*models.Family, error)
	GetFamilyByID(ctx context.Context, familyID string) (*models.Family, error)
	GetFamilyByInviteCode(ctx context.Context, inviteCode string) (*models.Family, error)
	UpdateFamily(ctx context.Context, family *models.Family) error
	DeleteFamily(ctx context.Context, familyID string) error
	GetFamiliesByManagerID(ctx context.Context, managerID string) ([]*models.Family, error)

	// Membership operations
	AddFamilyMember(ctx context.Context, familyID, userID, role string) error
	RemoveFamilyMember(ctx context.Context, familyID, userID string) error
	GetFamilyMemberships(ctx context.Context, familyID string) ([]*models.FamilyMembership, error)
	GetUserFamilyMembership(ctx context.Context, userID string) (*models.FamilyMembership, error)
	UpdateMemberRole(ctx context.Context, familyID, userID, newRole string) error

	// Combined operations
	GetFamilyWithMembers(ctx context.Context, familyID string) (*models.FamilyWithMembers, error)

	// Utility operations
	FamilyExists(ctx context.Context, familyID string) (bool, error)
	UserIsFamilyMember(ctx context.Context, familyID, userID string) (bool, error)
	UserIsFamilyManager(ctx context.Context, familyID, userID string) (bool, error)
	GenerateInviteCode(ctx context.Context) (string, error)
}


