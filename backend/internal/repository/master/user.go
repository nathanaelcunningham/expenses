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

type userRepository struct {
	queries *masterdb.Queries
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) interfaces.UserRepository {
	return &userRepository{
		queries: masterdb.New(db),
	}
}

// NewUserRepositoryWithTx creates a new user repository with a transaction
func NewUserRepositoryWithTx(tx *sql.Tx) interfaces.UserRepository {
	return &userRepository{
		queries: masterdb.New(tx),
	}
}

// WithTx returns a new user repository using the provided transaction
func (r *userRepository) WithTx(tx *sql.Tx) interface{} {
	return &userRepository{
		queries: r.queries.WithTx(tx),
	}
}

func (r *userRepository) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	params := masterdb.CreateUserParams{
		ID:           generateID(),
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		Name:         strings.TrimSpace(req.Name),
		PasswordHash: req.Password, // Should be hashed before calling this
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	result, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return convertSQLCUserToModel(result), nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	result, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return convertSQLCUserToModel(result), nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	result, err := r.queries.GetUserByEmail(ctx, strings.ToLower(email))
	if err != nil {
		return nil, err
	}

	return convertSQLCUserToModel(result), nil
}

func (r *userRepository) UpdateUser(ctx context.Context, id string, req *models.UpdateUserRequest) (*models.User, error) {
	// Get current user first
	current, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	params := masterdb.UpdateUserParams{
		ID:           id,
		Name:         current.Name,
		PasswordHash: current.PasswordHash,
		UpdatedAt:    time.Now(),
	}

	// Apply updates
	if req.Name != nil {
		params.Name = strings.TrimSpace(*req.Name)
	}
	if req.PasswordHash != nil {
		params.PasswordHash = *req.PasswordHash
	}

	result, err := r.queries.UpdateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return convertSQLCUserToModel(result), nil
}

func (r *userRepository) DeleteUser(ctx context.Context, id string) error {
	return r.queries.DeleteUser(ctx, id)
}

// Helper functions
func convertSQLCUserToModel(user *masterdb.User) *models.User {
	return &models.User{
		ID:           user.ID,
		Email:        user.Email,
		Name:         user.Name,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}