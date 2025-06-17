package master

import (
	"context"
	"database/sql"
	"expenses-backend/internal/database/sql/masterdb"
	"expenses-backend/internal/models"
	"expenses-backend/internal/repository/interfaces"
	"strings"
	"time"
)

type familyRepository struct {
	queries *masterdb.Queries
}

// NewFamilyRepository creates a new family repository
func NewFamilyRepository(db *sql.DB) interfaces.FamilyRepository {
	return &familyRepository{
		queries: masterdb.New(db),
	}
}

// NewFamilyRepositoryWithTx creates a new family repository with a transaction
func NewFamilyRepositoryWithTx(tx *sql.Tx) interfaces.FamilyRepository {
	return &familyRepository{
		queries: masterdb.New(tx),
	}
}

// WithTx returns a new family repository using the provided transaction
func (r *familyRepository) WithTx(tx *sql.Tx) interface{} {
	return &familyRepository{
		queries: r.queries.WithTx(tx),
	}
}

func (r *familyRepository) CreateFamily(ctx context.Context, req *models.CreateFamilyRequest) (*models.Family, error) {
	params := masterdb.CreateFamilyParams{
		ID:          generateID(),
		Name:        strings.TrimSpace(req.Name),
		InviteCode:  generateInviteCode(),
		DatabaseUrl: "", // Will be updated after database provisioning
		ManagerID:   req.ManagerID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result, err := r.queries.CreateFamily(ctx, params)
	if err != nil {
		return nil, err
	}

	return convertSQLCFamilyToModel(result), nil
}

func (r *familyRepository) GetFamilyByID(ctx context.Context, id string) (*models.Family, error) {
	result, err := r.queries.GetFamilyByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return convertSQLCFamilyToModel(result), nil
}

func (r *familyRepository) GetFamilyByInviteCode(ctx context.Context, inviteCode string) (*models.Family, error) {
	result, err := r.queries.GetFamilyByInviteCode(ctx, inviteCode)
	if err != nil {
		return nil, err
	}

	return convertSQLCFamilyToModel(result), nil
}

func (r *familyRepository) UpdateFamily(ctx context.Context, id string, req *models.UpdateFamilyRequest) (*models.Family, error) {
	// Get current family first
	current, err := r.queries.GetFamilyByID(ctx, id)
	if err != nil {
		return nil, err
	}

	params := masterdb.UpdateFamilyParams{
		ID:        id,
		Name:      current.Name,
		UpdatedAt: time.Now(),
	}

	// Apply updates
	if req.Name != nil {
		params.Name = strings.TrimSpace(*req.Name)
	}

	result, err := r.queries.UpdateFamily(ctx, params)
	if err != nil {
		return nil, err
	}

	return convertSQLCFamilyToModel(result), nil
}

func (r *familyRepository) DeleteFamily(ctx context.Context, id string) error {
	return r.queries.DeleteFamily(ctx, id)
}

// Helper functions
func convertSQLCFamilyToModel(family *masterdb.Family) *models.Family {
	return &models.Family{
		ID:          family.ID,
		Name:        family.Name,
		InviteCode:  family.InviteCode,
		DatabaseURL: family.DatabaseUrl,
		ManagerID:   family.ManagerID,
		CreatedAt:   family.CreatedAt,
		UpdatedAt:   family.UpdatedAt,
	}
}