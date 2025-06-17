package category

import (
	"context"
	"time"
)

// Category represents a category for organizing expenses
type Category struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Color       *string   `json:"color,omitempty"`   // Hex color for UI
	Icon        *string   `json:"icon,omitempty"`    // Icon identifier
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateCategoryRequest represents the data needed to create a category
type CreateCategoryRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty"`
	Icon        *string `json:"icon,omitempty"`
}

// UpdateCategoryRequest represents the data for updating a category
type UpdateCategoryRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty"`
	Icon        *string `json:"icon,omitempty"`
}

// CategoryWithStats represents a category with usage statistics
type CategoryWithStats struct {
	Category     *Category `json:"category"`
	ExpenseCount int       `json:"expense_count"`
	TotalAmount  float64   `json:"total_amount"`
	LastUsed     *time.Time `json:"last_used,omitempty"`
}

// CategoryFilter represents filters for category queries
type CategoryFilter struct {
	Name   *string `json:"name,omitempty"`
	Color  *string `json:"color,omitempty"`
	Icon   *string `json:"icon,omitempty"`
	Limit  int     `json:"limit,omitempty"`
	Offset int     `json:"offset,omitempty"`
}

// Store defines the interface for category repository operations
// This interface operates on family-specific databases
type Store interface {
	// Category CRUD operations
	CreateCategory(ctx context.Context, req *CreateCategoryRequest) (*Category, error)
	GetCategoryByID(ctx context.Context, categoryID string) (*Category, error)
	GetCategoryByName(ctx context.Context, name string) (*Category, error)
	UpdateCategory(ctx context.Context, categoryID string, req *UpdateCategoryRequest) (*Category, error)
	DeleteCategory(ctx context.Context, categoryID string) error

	// Category listing and filtering
	ListCategories(ctx context.Context, filter *CategoryFilter) ([]*Category, error)
	ListCategoriesWithStats(ctx context.Context, filter *CategoryFilter) ([]*CategoryWithStats, error)
	CountCategories(ctx context.Context, filter *CategoryFilter) (int, error)

	// Category analytics
	GetCategoryStats(ctx context.Context, categoryID string) (*CategoryWithStats, error)
	GetMostUsedCategories(ctx context.Context, limit int) ([]*CategoryWithStats, error)
	GetCategoriesUsedInDateRange(ctx context.Context, startDate, endDate time.Time) ([]*Category, error)

	// Utility operations
	CategoryExists(ctx context.Context, categoryID string) (bool, error)
	CategoryNameExists(ctx context.Context, name string, excludeID *string) (bool, error)
	GetDefaultCategories() []*CreateCategoryRequest
}