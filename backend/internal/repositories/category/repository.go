package category

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// Repository implements the Store interface for category operations
type Repository struct {
	db     *sql.DB
	logger zerolog.Logger
}

// NewRepository creates a new category repository that operates on a family database
func NewRepository(db *sql.DB, logger zerolog.Logger) Store {
	return &Repository{
		db:     db,
		logger: logger.With().Str("component", "category-repository").Logger(),
	}
}

// CreateCategory creates a new category in the family database
func (r *Repository) CreateCategory(ctx context.Context, req *CreateCategoryRequest) (*Category, error) {
	// Check if category name already exists
	exists, err := r.CategoryNameExists(ctx, req.Name, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check category name existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("category with name '%s' already exists", req.Name)
	}

	categoryID, err := r.generateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate category ID: %w", err)
	}

	now := time.Now()
	category := &Category{
		ID:          categoryID,
		Name:        strings.TrimSpace(req.Name),
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	query := `
		INSERT INTO categories (id, name, description, color, icon, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, query,
		category.ID, category.Name, category.Description, category.Color,
		category.Icon, category.CreatedAt, category.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	r.logger.Info().
		Str("category_id", category.ID).
		Str("category_name", category.Name).
		Msg("Category created successfully")

	return category, nil
}

// GetCategoryByID retrieves a category by its ID
func (r *Repository) GetCategoryByID(ctx context.Context, categoryID string) (*Category, error) {
	category := &Category{}
	query := `
		SELECT id, name, description, color, icon, created_at, updated_at
		FROM categories WHERE id = ?`

	err := r.db.QueryRowContext(ctx, query, categoryID).Scan(
		&category.ID, &category.Name, &category.Description, &category.Color,
		&category.Icon, &category.CreatedAt, &category.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category not found: %s", categoryID)
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return category, nil
}

// GetCategoryByName retrieves a category by its name
func (r *Repository) GetCategoryByName(ctx context.Context, name string) (*Category, error) {
	category := &Category{}
	query := `
		SELECT id, name, description, color, icon, created_at, updated_at
		FROM categories WHERE name = ?`

	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&category.ID, &category.Name, &category.Description, &category.Color,
		&category.Icon, &category.CreatedAt, &category.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category not found with name: %s", name)
		}
		return nil, fmt.Errorf("failed to get category by name: %w", err)
	}

	return category, nil
}

// UpdateCategory updates an existing category
func (r *Repository) UpdateCategory(ctx context.Context, categoryID string, req *UpdateCategoryRequest) (*Category, error) {
	// First, get the current category
	category, err := r.GetCategoryByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	// Check if new name conflicts with existing categories (excluding current one)
	if req.Name != nil && *req.Name != category.Name {
		exists, err := r.CategoryNameExists(ctx, *req.Name, &categoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to check category name existence: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("category with name '%s' already exists", *req.Name)
		}
	}

	// Build update query dynamically based on provided fields
	var setParts []string
	var args []interface{}

	if req.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *req.Name)
		category.Name = *req.Name
	}
	if req.Description != nil {
		setParts = append(setParts, "description = ?")
		args = append(args, *req.Description)
		category.Description = req.Description
	}
	if req.Color != nil {
		setParts = append(setParts, "color = ?")
		args = append(args, *req.Color)
		category.Color = req.Color
	}
	if req.Icon != nil {
		setParts = append(setParts, "icon = ?")
		args = append(args, *req.Icon)
		category.Icon = req.Icon
	}

	if len(setParts) > 0 {
		now := time.Now()
		setParts = append(setParts, "updated_at = ?")
		args = append(args, now)
		category.UpdatedAt = now

		query := fmt.Sprintf("UPDATE categories SET %s WHERE id = ?", strings.Join(setParts, ", "))
		args = append(args, categoryID)

		_, err = r.db.ExecContext(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to update category: %w", err)
		}

		r.logger.Info().
			Str("category_id", categoryID).
			Msg("Category updated successfully")
	}

	return category, nil
}

// DeleteCategory removes a category
func (r *Repository) DeleteCategory(ctx context.Context, categoryID string) error {
	// Check if category is being used by any expenses
	var expenseCount int
	query := `SELECT COUNT(*) FROM expenses WHERE category_id = ?`
	err := r.db.QueryRowContext(ctx, query, categoryID).Scan(&expenseCount)
	if err != nil {
		return fmt.Errorf("failed to check category usage: %w", err)
	}

	if expenseCount > 0 {
		return fmt.Errorf("cannot delete category: it is used by %d expense(s)", expenseCount)
	}

	// Delete the category
	deleteQuery := `DELETE FROM categories WHERE id = ?`
	result, err := r.db.ExecContext(ctx, deleteQuery, categoryID)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("category not found: %s", categoryID)
	}

	r.logger.Info().
		Str("category_id", categoryID).
		Msg("Category deleted successfully")

	return nil
}

// ListCategories retrieves categories based on filter criteria
func (r *Repository) ListCategories(ctx context.Context, filter *CategoryFilter) ([]*Category, error) {
	query := `
		SELECT id, name, description, color, icon, created_at, updated_at
		FROM categories`

	whereClause, args := r.buildWhereClause(filter)
	if whereClause != "" {
		query += " WHERE " + whereClause
	}

	query += " ORDER BY name ASC"

	if filter != nil {
		if filter.Limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", filter.Limit)
		}
		if filter.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", filter.Offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	var categories []*Category
	for rows.Next() {
		category := &Category{}
		err := rows.Scan(
			&category.ID, &category.Name, &category.Description, &category.Color,
			&category.Icon, &category.CreatedAt, &category.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, rows.Err()
}

// ListCategoriesWithStats retrieves categories with usage statistics
func (r *Repository) ListCategoriesWithStats(ctx context.Context, filter *CategoryFilter) ([]*CategoryWithStats, error) {
	query := `
		SELECT c.id, c.name, c.description, c.color, c.icon, c.created_at, c.updated_at,
		       COALESCE(COUNT(e.id), 0) as expense_count,
		       COALESCE(SUM(e.amount), 0) as total_amount,
		       MAX(e.date) as last_used
		FROM categories c
		LEFT JOIN expenses e ON c.id = e.category_id`

	whereClause, args := r.buildWhereClause(filter)
	if whereClause != "" {
		query += " WHERE " + whereClause
	}

	query += " GROUP BY c.id, c.name, c.description, c.color, c.icon, c.created_at, c.updated_at"
	query += " ORDER BY c.name ASC"

	if filter != nil {
		if filter.Limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", filter.Limit)
		}
		if filter.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", filter.Offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories with stats: %w", err)
	}
	defer rows.Close()

	var categoriesWithStats []*CategoryWithStats
	for rows.Next() {
		category := &Category{}
		var expenseCount int
		var totalAmount float64
		var lastUsed sql.NullTime

		err := rows.Scan(
			&category.ID, &category.Name, &category.Description, &category.Color,
			&category.Icon, &category.CreatedAt, &category.UpdatedAt,
			&expenseCount, &totalAmount, &lastUsed)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category with stats: %w", err)
		}

		categoryWithStats := &CategoryWithStats{
			Category:     category,
			ExpenseCount: expenseCount,
			TotalAmount:  totalAmount,
		}

		if lastUsed.Valid {
			categoryWithStats.LastUsed = &lastUsed.Time
		}

		categoriesWithStats = append(categoriesWithStats, categoryWithStats)
	}

	return categoriesWithStats, rows.Err()
}

// CountCategories counts categories based on filter criteria
func (r *Repository) CountCategories(ctx context.Context, filter *CategoryFilter) (int, error) {
	query := "SELECT COUNT(*) FROM categories"
	
	whereClause, args := r.buildWhereClause(filter)
	if whereClause != "" {
		query += " WHERE " + whereClause
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count categories: %w", err)
	}

	return count, nil
}

// GetCategoryStats retrieves statistics for a specific category
func (r *Repository) GetCategoryStats(ctx context.Context, categoryID string) (*CategoryWithStats, error) {
	category, err := r.GetCategoryByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT COALESCE(COUNT(id), 0) as expense_count,
		       COALESCE(SUM(amount), 0) as total_amount,
		       MAX(date) as last_used
		FROM expenses WHERE category_id = ?`

	var expenseCount int
	var totalAmount float64
	var lastUsed sql.NullTime

	err = r.db.QueryRowContext(ctx, query, categoryID).Scan(&expenseCount, &totalAmount, &lastUsed)
	if err != nil {
		return nil, fmt.Errorf("failed to get category stats: %w", err)
	}

	stats := &CategoryWithStats{
		Category:     category,
		ExpenseCount: expenseCount,
		TotalAmount:  totalAmount,
	}

	if lastUsed.Valid {
		stats.LastUsed = &lastUsed.Time
	}

	return stats, nil
}

// GetMostUsedCategories retrieves the most frequently used categories
func (r *Repository) GetMostUsedCategories(ctx context.Context, limit int) ([]*CategoryWithStats, error) {
	query := `
		SELECT c.id, c.name, c.description, c.color, c.icon, c.created_at, c.updated_at,
		       COUNT(e.id) as expense_count,
		       COALESCE(SUM(e.amount), 0) as total_amount,
		       MAX(e.date) as last_used
		FROM categories c
		INNER JOIN expenses e ON c.id = e.category_id
		GROUP BY c.id, c.name, c.description, c.color, c.icon, c.created_at, c.updated_at
		ORDER BY expense_count DESC, total_amount DESC
		LIMIT ?`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get most used categories: %w", err)
	}
	defer rows.Close()

	var categoriesWithStats []*CategoryWithStats
	for rows.Next() {
		category := &Category{}
		var expenseCount int
		var totalAmount float64
		var lastUsed sql.NullTime

		err := rows.Scan(
			&category.ID, &category.Name, &category.Description, &category.Color,
			&category.Icon, &category.CreatedAt, &category.UpdatedAt,
			&expenseCount, &totalAmount, &lastUsed)
		if err != nil {
			return nil, fmt.Errorf("failed to scan most used category: %w", err)
		}

		categoryWithStats := &CategoryWithStats{
			Category:     category,
			ExpenseCount: expenseCount,
			TotalAmount:  totalAmount,
		}

		if lastUsed.Valid {
			categoryWithStats.LastUsed = &lastUsed.Time
		}

		categoriesWithStats = append(categoriesWithStats, categoryWithStats)
	}

	return categoriesWithStats, rows.Err()
}

// GetCategoriesUsedInDateRange retrieves categories used within a date range
func (r *Repository) GetCategoriesUsedInDateRange(ctx context.Context, startDate, endDate time.Time) ([]*Category, error) {
	query := `
		SELECT DISTINCT c.id, c.name, c.description, c.color, c.icon, c.created_at, c.updated_at
		FROM categories c
		INNER JOIN expenses e ON c.id = e.category_id
		WHERE e.date BETWEEN ? AND ?
		ORDER BY c.name ASC`

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories used in date range: %w", err)
	}
	defer rows.Close()

	var categories []*Category
	for rows.Next() {
		category := &Category{}
		err := rows.Scan(
			&category.ID, &category.Name, &category.Description, &category.Color,
			&category.Icon, &category.CreatedAt, &category.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, rows.Err()
}

// CategoryExists checks if a category exists
func (r *Repository) CategoryExists(ctx context.Context, categoryID string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM categories WHERE id = ?`
	err := r.db.QueryRowContext(ctx, query, categoryID).Scan(&count)
	return count > 0, err
}

// CategoryNameExists checks if a category name already exists (optionally excluding a specific ID)
func (r *Repository) CategoryNameExists(ctx context.Context, name string, excludeID *string) (bool, error) {
	query := `SELECT COUNT(*) FROM categories WHERE name = ?`
	args := []interface{}{name}

	if excludeID != nil {
		query += " AND id != ?"
		args = append(args, *excludeID)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count > 0, err
}

// GetDefaultCategories returns a list of default categories to be created for new families
func (r *Repository) GetDefaultCategories() []*CreateCategoryRequest {
	return []*CreateCategoryRequest{
		{
			Name:        "Food & Dining",
			Description: stringPtr("Groceries, restaurants, takeout"),
			Color:       stringPtr("#FF6B6B"),
			Icon:        stringPtr("utensils"),
		},
		{
			Name:        "Transportation",
			Description: stringPtr("Gas, public transit, rideshare"),
			Color:       stringPtr("#4ECDC4"),
			Icon:        stringPtr("car"),
		},
		{
			Name:        "Utilities",
			Description: stringPtr("Electricity, water, internet, phone"),
			Color:       stringPtr("#45B7D1"),
			Icon:        stringPtr("bolt"),
		},
		{
			Name:        "Entertainment",
			Description: stringPtr("Movies, games, subscriptions"),
			Color:       stringPtr("#96CEB4"),
			Icon:        stringPtr("play"),
		},
		{
			Name:        "Shopping",
			Description: stringPtr("Clothing, household items"),
			Color:       stringPtr("#FFEAA7"),
			Icon:        stringPtr("shopping-bag"),
		},
		{
			Name:        "Healthcare",
			Description: stringPtr("Medical, dental, prescriptions"),
			Color:       stringPtr("#DDA0DD"),
			Icon:        stringPtr("heart"),
		},
		{
			Name:        "Education",
			Description: stringPtr("Books, courses, school supplies"),
			Color:       stringPtr("#98D8C8"),
			Icon:        stringPtr("book"),
		},
		{
			Name:        "Other",
			Description: stringPtr("Miscellaneous expenses"),
			Color:       stringPtr("#A8A8A8"),
			Icon:        stringPtr("more-horizontal"),
		},
	}
}

// Helper methods

// buildWhereClause builds the WHERE clause for filtering categories
func (r *Repository) buildWhereClause(filter *CategoryFilter) (string, []interface{}) {
	if filter == nil {
		return "", nil
	}

	var conditions []string
	var args []interface{}

	if filter.Name != nil {
		conditions = append(conditions, "name LIKE ?")
		args = append(args, "%"+*filter.Name+"%")
	}

	if filter.Color != nil {
		conditions = append(conditions, "color = ?")
		args = append(args, *filter.Color)
	}

	if filter.Icon != nil {
		conditions = append(conditions, "icon = ?")
		args = append(args, *filter.Icon)
	}

	return strings.Join(conditions, " AND "), args
}

// generateID generates a random ID
func (r *Repository) generateID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}