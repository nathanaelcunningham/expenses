package family

import (
	"context"
	"database/sql"
	"expenses-backend/internal/database/sql/familydb"
	"expenses-backend/internal/models"
	"expenses-backend/internal/repository/interfaces"
	"strings"
)

type familyMemberRepository struct {
	queries *familydb.Queries
}

// NewFamilyMemberRepository creates a new family member repository
func NewFamilyMemberRepository(db *sql.DB) interfaces.FamilyMemberRepository {
	return &familyMemberRepository{
		queries: familydb.New(db),
	}
}

// NewFamilyMemberRepositoryWithTx creates a new family member repository with a transaction
func NewFamilyMemberRepositoryWithTx(tx *sql.Tx) interfaces.FamilyMemberRepository {
	return &familyMemberRepository{
		queries: familydb.New(tx),
	}
}

// WithTx returns a new family member repository using the provided transaction
func (r *familyMemberRepository) WithTx(tx *sql.Tx) interface{} {
	return &familyMemberRepository{
		queries: r.queries.WithTx(tx),
	}
}

func (r *familyMemberRepository) CreateFamilyMember(ctx context.Context, member *models.Member) (*models.Member, error) {
	params := familydb.CreateFamilyMemberParams{
		ID:       member.UserID,
		Name:     member.Name,
		Email:    member.Email,
		Role:     member.Role,
		JoinedAt: member.JoinedAt,
		IsActive: &member.IsActive,
	}

	result, err := r.queries.CreateFamilyMember(ctx, params)
	if err != nil {
		return nil, err
	}

	return convertSQLCFamilyMemberToModel(result), nil
}

func (r *familyMemberRepository) GetFamilyMemberByID(ctx context.Context, id string) (*models.Member, error) {
	result, err := r.queries.GetFamilyMemberByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return convertSQLCFamilyMemberToModel(result), nil
}

func (r *familyMemberRepository) GetFamilyMemberByEmail(ctx context.Context, email string) (*models.Member, error) {
	result, err := r.queries.GetFamilyMemberByEmail(ctx, strings.ToLower(email))
	if err != nil {
		return nil, err
	}

	return convertSQLCFamilyMemberToModel(result), nil
}

func (r *familyMemberRepository) ListFamilyMembers(ctx context.Context) ([]*models.Member, error) {
	results, err := r.queries.ListFamilyMembers(ctx)
	if err != nil {
		return nil, err
	}

	members := make([]*models.Member, len(results))
	for i, result := range results {
		members[i] = convertSQLCFamilyMemberToModel(result)
	}

	return members, nil
}

func (r *familyMemberRepository) ListAllFamilyMembers(ctx context.Context) ([]*models.Member, error) {
	results, err := r.queries.ListAllFamilyMembers(ctx)
	if err != nil {
		return nil, err
	}

	members := make([]*models.Member, len(results))
	for i, result := range results {
		members[i] = convertSQLCFamilyMemberToModel(result)
	}

	return members, nil
}

func (r *familyMemberRepository) UpdateFamilyMember(ctx context.Context, id string, member *models.Member) (*models.Member, error) {
	params := familydb.UpdateFamilyMemberParams{
		ID:    id,
		Name:  member.Name,
		Email: member.Email,
		Role:  member.Role,
	}

	result, err := r.queries.UpdateFamilyMember(ctx, params)
	if err != nil {
		return nil, err
	}

	return convertSQLCFamilyMemberToModel(result), nil
}

func (r *familyMemberRepository) DeactivateFamilyMember(ctx context.Context, id string) error {
	return r.queries.DeactivateFamilyMember(ctx, id)
}

func (r *familyMemberRepository) DeleteFamilyMember(ctx context.Context, id string) error {
	return r.queries.DeleteFamilyMember(ctx, id)
}

// Helper functions
func convertSQLCFamilyMemberToModel(member *familydb.FamilyMember) *models.Member {
	isActive := true
	if member.IsActive != nil {
		isActive = *member.IsActive
	}

	return &models.Member{
		UserID:   member.ID,
		Name:     member.Name,
		Email:    member.Email,
		Role:     member.Role,
		JoinedAt: member.JoinedAt,
		IsActive: isActive,
	}
}