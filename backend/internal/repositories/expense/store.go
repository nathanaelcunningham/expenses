package expense

import (
	"context"
	"expenses-backend/internal/models"
)

// Store defines the interface for expense repository operations
// This interface operates on family-specific databases
type Store interface {
	// Expense CRUD operations
	CreateExpense(ctx context.Context, req *models.CreateExpenseRequest) (*models.Expense, error)
	GetExpenseByID(ctx context.Context, expenseID string) (*models.Expense, error)
	UpdateExpense(ctx context.Context, expenseID string, req *models.UpdateExpenseRequest) (*models.Expense, error)
	DeleteExpense(ctx context.Context, expenseID string) error

	// Expense listing and filtering
	ListExpenses(ctx context.Context) ([]*models.Expense, error)

	// Expense analytics and summaries
	GetExpenseSummary(ctx context.Context) (*models.ExpenseSummary, error)
}

