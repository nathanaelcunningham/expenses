package expense

import (
	"context"
	"time"
)

// Expense represents an expense record from the family database
type Expense struct {
	ID               string     `json:"id"`
	MemberID         string     `json:"member_id"`
	CategoryID       *string    `json:"category_id,omitempty"`
	Amount           float64    `json:"amount"`
	Currency         string     `json:"currency"`
	Description      string     `json:"description"`
	Date             time.Time  `json:"date"`
	ReceiptURL       *string    `json:"receipt_url,omitempty"`
	Tags             *string    `json:"tags,omitempty"` // JSON array of tags
	IsRecurring      bool       `json:"is_recurring"`
	RecurringInterval *string   `json:"recurring_interval,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// ExpenseSplit represents how an expense is split among family members
type ExpenseSplit struct {
	ID         string   `json:"id"`
	ExpenseID  string   `json:"expense_id"`
	MemberID   string   `json:"member_id"`
	Amount     float64  `json:"amount"`
	Percentage *float64 `json:"percentage,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// ExpenseWithSplits represents an expense with its splits
type ExpenseWithSplits struct {
	Expense *Expense        `json:"expense"`
	Splits  []*ExpenseSplit `json:"splits"`
}

// CreateExpenseRequest represents the data needed to create an expense
type CreateExpenseRequest struct {
	MemberID          string             `json:"member_id"`
	CategoryID        *string            `json:"category_id,omitempty"`
	Amount            float64            `json:"amount"`
	Currency          string             `json:"currency"`
	Description       string             `json:"description"`
	Date              time.Time          `json:"date"`
	ReceiptURL        *string            `json:"receipt_url,omitempty"`
	Tags              *string            `json:"tags,omitempty"`
	IsRecurring       bool               `json:"is_recurring"`
	RecurringInterval *string            `json:"recurring_interval,omitempty"`
	Splits            []*CreateSplitRequest `json:"splits,omitempty"`
}

// CreateSplitRequest represents the data needed to create an expense split
type CreateSplitRequest struct {
	MemberID   string   `json:"member_id"`
	Amount     float64  `json:"amount"`
	Percentage *float64 `json:"percentage,omitempty"`
}

// UpdateExpenseRequest represents the data for updating an expense
type UpdateExpenseRequest struct {
	CategoryID        *string            `json:"category_id,omitempty"`
	Amount            *float64           `json:"amount,omitempty"`
	Currency          *string            `json:"currency,omitempty"`
	Description       *string            `json:"description,omitempty"`
	Date              *time.Time         `json:"date,omitempty"`
	ReceiptURL        *string            `json:"receipt_url,omitempty"`
	Tags              *string            `json:"tags,omitempty"`
	IsRecurring       *bool              `json:"is_recurring,omitempty"`
	RecurringInterval *string            `json:"recurring_interval,omitempty"`
	Splits            []*CreateSplitRequest `json:"splits,omitempty"`
}

// ExpenseFilter represents filters for expense queries
type ExpenseFilter struct {
	MemberID    *string    `json:"member_id,omitempty"`
	CategoryID  *string    `json:"category_id,omitempty"`
	DateFrom    *time.Time `json:"date_from,omitempty"`
	DateTo      *time.Time `json:"date_to,omitempty"`
	MinAmount   *float64   `json:"min_amount,omitempty"`
	MaxAmount   *float64   `json:"max_amount,omitempty"`
	IsRecurring *bool      `json:"is_recurring,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	Limit       int        `json:"limit,omitempty"`
	Offset      int        `json:"offset,omitempty"`
}

// ExpenseSummary represents expense summary statistics
type ExpenseSummary struct {
	TotalAmount    float64            `json:"total_amount"`
	ExpenseCount   int                `json:"expense_count"`
	AverageAmount  float64            `json:"average_amount"`
	ByCategory     map[string]float64 `json:"by_category"`
	ByMember       map[string]float64 `json:"by_member"`
	Currency       string             `json:"currency"`
}

// Store defines the interface for expense repository operations
// This interface operates on family-specific databases
type Store interface {
	// Expense CRUD operations
	CreateExpense(ctx context.Context, req *CreateExpenseRequest) (*Expense, error)
	GetExpenseByID(ctx context.Context, expenseID string) (*Expense, error)
	GetExpenseWithSplits(ctx context.Context, expenseID string) (*ExpenseWithSplits, error)
	UpdateExpense(ctx context.Context, expenseID string, req *UpdateExpenseRequest) (*Expense, error)
	DeleteExpense(ctx context.Context, expenseID string) error

	// Expense listing and filtering
	ListExpenses(ctx context.Context, filter *ExpenseFilter) ([]*Expense, error)
	ListExpensesWithSplits(ctx context.Context, filter *ExpenseFilter) ([]*ExpenseWithSplits, error)
	CountExpenses(ctx context.Context, filter *ExpenseFilter) (int, error)

	// Expense splits operations
	CreateExpenseSplit(ctx context.Context, expenseID string, req *CreateSplitRequest) (*ExpenseSplit, error)
	GetExpenseSplits(ctx context.Context, expenseID string) ([]*ExpenseSplit, error)
	UpdateExpenseSplit(ctx context.Context, splitID string, amount float64, percentage *float64) (*ExpenseSplit, error)
	DeleteExpenseSplit(ctx context.Context, splitID string) error
	DeleteAllExpenseSplits(ctx context.Context, expenseID string) error

	// Expense analytics and summaries
	GetExpenseSummary(ctx context.Context, filter *ExpenseFilter) (*ExpenseSummary, error)
	GetExpensesByDateRange(ctx context.Context, memberID *string, startDate, endDate time.Time) ([]*Expense, error)
	GetTotalSpentByMember(ctx context.Context, memberID string, startDate, endDate time.Time) (float64, error)
	GetTotalSpentByCategory(ctx context.Context, categoryID string, startDate, endDate time.Time) (float64, error)

	// Recurring expenses
	GetRecurringExpenses(ctx context.Context) ([]*Expense, error)
	GetExpensesByMember(ctx context.Context, memberID string, limit, offset int) ([]*Expense, error)

	// Utility operations
	ExpenseExists(ctx context.Context, expenseID string) (bool, error)
	UserCanAccessExpense(ctx context.Context, expenseID, userID string) (bool, error)
}