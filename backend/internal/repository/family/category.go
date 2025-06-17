package family

import (
	"context"
	"database/sql"
	"expenses-backend/internal/database/sql/familydb"
	"expenses-backend/internal/models"
	"expenses-backend/internal/repository/interfaces"
	"strings"
	"time"
)

type categoryRepository struct {
	queries *familydb.Queries
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *sql.DB) interfaces.CategoryRepository {
	return &categoryRepository{
		queries: familydb.New(db),
	}
}

// NewCategoryRepositoryWithTx creates a new category repository with a transaction
func NewCategoryRepositoryWithTx(tx *sql.Tx) interfaces.CategoryRepository {
	return &categoryRepository{
		queries: familydb.New(tx),
	}
}

// WithTx returns a new category repository using the provided transaction
func (r *categoryRepository) WithTx(tx *sql.Tx) interface{} {
	return &categoryRepository{
		queries: r.queries.WithTx(tx),
	}
}

func (r *categoryRepository) CreateCategory(ctx context.Context, req *models.CreateCategoryRequest) (*models.Category, error) {
	params := familydb.CreateCategoryParams{
		Name:      strings.TrimSpace(req.Name),
		Color:     req.Color,
		Icon:      req.Icon,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := r.queries.CreateCategory(ctx, params)
	if err != nil {
		return nil, err
	}

	return convertSQLCCategoryToModel(result), nil
}

func (r *categoryRepository) GetCategoryByID(ctx context.Context, id string) (*models.Category, error) {
	result, err := r.queries.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return convertSQLCCategoryToModel(result), nil
}

func (r *categoryRepository) ListCategories(ctx context.Context) ([]*models.Category, error) {
	results, err := r.queries.ListCategories(ctx)
	if err != nil {
		return nil, err
	}

	categories := make([]*models.Category, len(results))
	for i, result := range results {
		categories[i] = convertSQLCCategoryToModel(result)
	}

	return categories, nil
}

func (r *categoryRepository) UpdateCategory(ctx context.Context, id string, req *models.UpdateCategoryRequest) (*models.Category, error) {
	// Get current category first
	current, err := r.queries.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}

	params := familydb.UpdateCategoryParams{
		ID:        id,
		Name:      current.Name,
		Color:     current.Color,
		Icon:      current.Icon,
		UpdatedAt: time.Now(),
	}

	// Apply updates
	if req.Name != nil {
		params.Name = strings.TrimSpace(*req.Name)
	}
	if req.Color != nil {
		params.Color = req.Color
	}
	if req.Icon != nil {
		params.Icon = req.Icon
	}

	result, err := r.queries.UpdateCategory(ctx, params)
	if err != nil {
		return nil, err
	}

	return convertSQLCCategoryToModel(result), nil
}

func (r *categoryRepository) DeleteCategory(ctx context.Context, id string) error {
	return r.queries.DeleteCategory(ctx, id)
}

// Helper functions
func convertSQLCCategoryToModel(category *familydb.Category) *models.Category {
	return &models.Category{
		ID:        category.ID,
		Name:      category.Name,
		Color:     category.Color,
		Icon:      category.Icon,
		Budget:    nil, // Not tracked in current schema
		IsActive:  true, // Assume active if exists
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}
}