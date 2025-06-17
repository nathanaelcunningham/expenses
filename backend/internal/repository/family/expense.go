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

type expenseRepository struct {
	queries *familydb.Queries
}

// NewExpenseRepository creates a new expense repository
func NewExpenseRepository(db *sql.DB) interfaces.ExpenseRepository {
	return &expenseRepository{
		queries: familydb.New(db),
	}
}

// NewExpenseRepositoryWithTx creates a new expense repository with a transaction
func NewExpenseRepositoryWithTx(tx *sql.Tx) interfaces.ExpenseRepository {
	return &expenseRepository{
		queries: familydb.New(tx),
	}
}

// WithTx returns a new expense repository using the provided transaction
func (r *expenseRepository) WithTx(tx *sql.Tx) interface{} {
	return &expenseRepository{
		queries: r.queries.WithTx(tx),
	}
}

func (r *expenseRepository) CreateExpense(ctx context.Context, req *models.CreateExpenseRequest) (*models.Expense, error) {
	params := familydb.CreateExpenseParams{
		CategoryID:    req.CategoryID,
		Amount:        req.Amount,
		Name:          req.Name,
		DayOfMonthDue: int64(req.Date.Day()),
		IsAutopay:     req.IsRecurring,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	result, err := r.queries.CreateExpense(ctx, params)
	if err != nil {
		return nil, err
	}

	return convertSQLCExpenseToModel(result), nil
}

func (r *expenseRepository) GetExpenseByID(ctx context.Context, id string) (*models.Expense, error) {
	result, err := r.queries.GetExpenseByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return convertSQLCExpenseToModel(result), nil
}

func (r *expenseRepository) ListExpenses(ctx context.Context, filter *models.ExpenseFilter) ([]*models.Expense, error) {
	limit := int64(50) // default limit
	offset := int64(0)

	if filter != nil {
		if filter.Limit > 0 {
			limit = int64(filter.Limit)
		}
		if filter.Offset > 0 {
			offset = int64(filter.Offset)
		}
	}

	params := familydb.ListExpensesParams{
		Limit:  limit,
		Offset: offset,
	}

	results, err := r.queries.ListExpenses(ctx, params)
	if err != nil {
		return nil, err
	}

	expenses := make([]*models.Expense, len(results))
	for i, result := range results {
		expenses[i] = convertSQLCExpenseToModel(result)
	}

	return expenses, nil
}

func (r *expenseRepository) ListExpensesByCategory(ctx context.Context, categoryID string) ([]*models.Expense, error) {
	results, err := r.queries.ListExpensesByCategory(ctx, &categoryID)
	if err != nil {
		return nil, err
	}

	expenses := make([]*models.Expense, len(results))
	for i, result := range results {
		expenses[i] = convertSQLCExpenseToModel(result)
	}

	return expenses, nil
}

func (r *expenseRepository) GetExpensesByDateRange(ctx context.Context, startDay, endDay int) ([]*models.Expense, error) {
	params := familydb.GetExpensesByDateRangeParams{
		FromDayOfMonthDue: int64(startDay),
		ToDayOfMonthDue:   int64(endDay),
	}

	results, err := r.queries.GetExpensesByDateRange(ctx, params)
	if err != nil {
		return nil, err
	}

	expenses := make([]*models.Expense, len(results))
	for i, result := range results {
		expenses[i] = convertSQLCExpenseToModel(result)
	}

	return expenses, nil
}

func (r *expenseRepository) UpdateExpense(ctx context.Context, id string, req *models.UpdateExpenseRequest) (*models.Expense, error) {
	// Get current expense first
	current, err := r.queries.GetExpenseByID(ctx, id)
	if err != nil {
		return nil, err
	}

	params := familydb.UpdateExpenseParams{
		ID:            id,
		CategoryID:    current.CategoryID,
		Amount:        current.Amount,
		Name:          current.Name,
		DayOfMonthDue: current.DayOfMonthDue,
		IsAutopay:     current.IsAutopay,
		UpdatedAt:     time.Now(),
	}

	// Apply updates
	if req.Name != nil {
		params.Name = strings.TrimSpace(*req.Name)
	}
	if req.Amount != nil {
		params.Amount = *req.Amount
	}
	if req.CategoryID != nil {
		params.CategoryID = req.CategoryID
	}
	if req.Date != nil {
		params.DayOfMonthDue = int64(req.Date.Day())
	}
	if req.IsRecurring != nil {
		params.IsAutopay = *req.IsRecurring
	}

	result, err := r.queries.UpdateExpense(ctx, params)
	if err != nil {
		return nil, err
	}

	return convertSQLCExpenseToModel(result), nil
}

func (r *expenseRepository) DeleteExpense(ctx context.Context, id string) error {
	return r.queries.DeleteExpense(ctx, id)
}

func (r *expenseRepository) CountExpenses(ctx context.Context) (int64, error) {
	return r.queries.CountExpenses(ctx)
}

func (r *expenseRepository) UserCanAccessExpense(ctx context.Context, expenseID, userID string) (bool, error) {
	// For family databases, all family members can access all expenses
	// In a more complex system, you might check expense ownership or permissions
	_, err := r.queries.GetExpenseByID(ctx, expenseID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Helper functions
func convertSQLCExpenseToModel(expense *familydb.Expense) *models.Expense {
	// Calculate date from day of month (use current month for now)
	now := time.Now()
	year, month, _ := now.Date()
	day := int(expense.DayOfMonthDue)
	if day > 31 {
		day = 31
	}
	if day < 1 {
		day = 1
	}

	date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

	return &models.Expense{
		ID:          expense.ID,
		Name:        expense.Name,
		Description: expense.Name, // Using name as description for now
		Amount:      expense.Amount,
		CategoryID:  expense.CategoryID,
		Date:        date,
		MemberID:    "", // Would need to be tracked separately or derived
		IsRecurring: expense.IsAutopay,
		CreatedAt:   expense.CreatedAt,
		UpdatedAt:   expense.UpdatedAt,
	}
}