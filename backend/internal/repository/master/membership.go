package master

import (
	"context"
	"database/sql"
	"expenses-backend/internal/database/sql/masterdb"
	"expenses-backend/internal/models"
	"expenses-backend/internal/repository/interfaces"
	"time"
)

type familyMembershipRepository struct {
	queries *masterdb.Queries
}

// NewFamilyMembershipRepository creates a new family membership repository
func NewFamilyMembershipRepository(db *sql.DB) interfaces.FamilyMembershipRepository {
	return &familyMembershipRepository{
		queries: masterdb.New(db),
	}
}

// NewFamilyMembershipRepositoryWithTx creates a new family membership repository with a transaction
func NewFamilyMembershipRepositoryWithTx(tx *sql.Tx) interfaces.FamilyMembershipRepository {
	return &familyMembershipRepository{
		queries: masterdb.New(tx),
	}
}

// WithTx returns a new family membership repository using the provided transaction
func (r *familyMembershipRepository) WithTx(tx *sql.Tx) interface{} {
	return &familyMembershipRepository{
		queries: r.queries.WithTx(tx),
	}
}

func (r *familyMembershipRepository) CreateMembership(ctx context.Context, familyID, userID, role string) error {
	params := masterdb.CreateFamilyMembershipParams{
		FamilyID: &familyID,
		UserID:   &userID,
		Role:     role,
		JoinedAt: time.Now(),
	}

	_, err := r.queries.CreateFamilyMembership(ctx, params)
	return err
}

func (r *familyMembershipRepository) GetMembership(ctx context.Context, familyID, userID string) (*models.Member, error) {
	params := masterdb.GetFamilyMembershipParams{
		FamilyID: &familyID,
		UserID:   &userID,
	}

	result, err := r.queries.GetFamilyMembership(ctx, params)
	if err != nil {
		return nil, err
	}

	// Note: This only returns the membership info, not the full member details
	// You'd need to join with the users table to get name and email
	member := &models.Member{
		UserID:   *result.UserID,
		Role:     result.Role,
		JoinedAt: result.JoinedAt,
		IsActive: true,
	}

	return member, nil
}

func (r *familyMembershipRepository) ListFamilyMemberships(ctx context.Context, familyID string) ([]*models.Member, error) {
	results, err := r.queries.ListFamilyMemberships(ctx, &familyID)
	if err != nil {
		return nil, err
	}

	members := make([]*models.Member, len(results))
	for i, result := range results {
		members[i] = &models.Member{
			UserID:   *result.UserID,
			Role:     result.Role,
			JoinedAt: result.JoinedAt,
			IsActive: true,
		}
	}

	return members, nil
}

func (r *familyMembershipRepository) ListUserMemberships(ctx context.Context, userID string) ([]*models.Family, error) {
	results, err := r.queries.ListUserMemberships(ctx, &userID)
	if err != nil {
		return nil, err
	}

	// Note: This only returns membership info, not full family details
	// You'd need to join with families table to get family details
	families := make([]*models.Family, len(results))
	for i, result := range results {
		families[i] = &models.Family{
			ID: *result.FamilyID,
			// Other fields would need to be populated by joining with families table
		}
	}

	return families, nil
}

func (r *familyMembershipRepository) UpdateMembershipRole(ctx context.Context, familyID, userID, role string) error {
	params := masterdb.UpdateFamilyMembershipRoleParams{
		FamilyID: &familyID,
		UserID:   &userID,
		Role:     role,
	}

	_, err := r.queries.UpdateFamilyMembershipRole(ctx, params)
	return err
}

func (r *familyMembershipRepository) DeleteMembership(ctx context.Context, familyID, userID string) error {
	params := masterdb.DeleteFamilyMembershipParams{
		FamilyID: &familyID,
		UserID:   &userID,
	}

	return r.queries.DeleteFamilyMembership(ctx, params)
}