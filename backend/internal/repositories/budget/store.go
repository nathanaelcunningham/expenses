package budget

import (
	"context"
	"time"
)

// Budget represents a budget for expense tracking
type Budget struct {
	ID         string     `json:"id"`
	CategoryID *string    `json:"category_id,omitempty"` // NULL for family-wide budgets
	MemberID   *string    `json:"member_id,omitempty"`   // NULL for family-wide budgets
	Amount     float64    `json:"amount"`
	Period     string     `json:"period"` // 'monthly', 'weekly', 'yearly'
	StartDate  time.Time  `json:"start_date"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	IsActive   bool       `json:"is_active"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// BudgetWithSpent represents a budget with current spending information
type BudgetWithSpent struct {
	Budget      *Budget `json:"budget"`
	AmountSpent float64 `json:"amount_spent"`
	Remaining   float64 `json:"remaining"`
	Percentage  float64 `json:"percentage"` // Percentage of budget used
	IsOverBudget bool   `json:"is_over_budget"`
}

// CreateBudgetRequest represents the data needed to create a budget
type CreateBudgetRequest struct {
	CategoryID *string    `json:"category_id,omitempty"`
	MemberID   *string    `json:"member_id,omitempty"`
	Amount     float64    `json:"amount"`
	Period     string     `json:"period"`
	StartDate  time.Time  `json:"start_date"`
	EndDate    *time.Time `json:"end_date,omitempty"`
}

// UpdateBudgetRequest represents the data for updating a budget
type UpdateBudgetRequest struct {
	CategoryID *string    `json:"category_id,omitempty"`
	MemberID   *string    `json:"member_id,omitempty"`
	Amount     *float64   `json:"amount,omitempty"`
	Period     *string    `json:"period,omitempty"`
	StartDate  *time.Time `json:"start_date,omitempty"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	IsActive   *bool      `json:"is_active,omitempty"`
}

// BudgetFilter represents filters for budget queries
type BudgetFilter struct {
	CategoryID *string    `json:"category_id,omitempty"`
	MemberID   *string    `json:"member_id,omitempty"`
	Period     *string    `json:"period,omitempty"`
	IsActive   *bool      `json:"is_active,omitempty"`
	StartDate  *time.Time `json:"start_date,omitempty"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	Limit      int        `json:"limit,omitempty"`
	Offset     int        `json:"offset,omitempty"`
}

// BudgetSummary represents budget summary statistics
type BudgetSummary struct {
	TotalBudgets    int                     `json:"total_budgets"`
	ActiveBudgets   int                     `json:"active_budgets"`
	TotalAmount     float64                 `json:"total_amount"`
	TotalSpent      float64                 `json:"total_spent"`
	TotalRemaining  float64                 `json:"total_remaining"`
	OverBudgetCount int                     `json:"over_budget_count"`
	ByCategory      map[string]*BudgetWithSpent `json:"by_category"`
	ByMember        map[string]*BudgetWithSpent `json:"by_member"`
	ByPeriod        map[string]int          `json:"by_period"`
}

// BudgetAlert represents a budget alert when spending exceeds thresholds
type BudgetAlert struct {
	BudgetID     string    `json:"budget_id"`
	Budget       *Budget   `json:"budget"`
	AmountSpent  float64   `json:"amount_spent"`
	Percentage   float64   `json:"percentage"`
	AlertType    string    `json:"alert_type"` // 'warning', 'exceeded'
	AlertDate    time.Time `json:"alert_date"`
}

// Store defines the interface for budget repository operations
// This interface operates on family-specific databases
type Store interface {
	// Budget CRUD operations
	CreateBudget(ctx context.Context, req *CreateBudgetRequest) (*Budget, error)
	GetBudgetByID(ctx context.Context, budgetID string) (*Budget, error)
	GetBudgetWithSpent(ctx context.Context, budgetID string) (*BudgetWithSpent, error)
	UpdateBudget(ctx context.Context, budgetID string, req *UpdateBudgetRequest) (*Budget, error)
	DeleteBudget(ctx context.Context, budgetID string) error

	// Budget listing and filtering
	ListBudgets(ctx context.Context, filter *BudgetFilter) ([]*Budget, error)
	ListBudgetsWithSpent(ctx context.Context, filter *BudgetFilter) ([]*BudgetWithSpent, error)
	CountBudgets(ctx context.Context, filter *BudgetFilter) (int, error)

	// Budget analytics and monitoring
	GetBudgetSummary(ctx context.Context, filter *BudgetFilter) (*BudgetSummary, error)
	GetActiveBudgets(ctx context.Context) ([]*BudgetWithSpent, error)
	GetOverBudgetAlerts(ctx context.Context) ([]*BudgetAlert, error)
	GetBudgetsByCategory(ctx context.Context, categoryID string) ([]*BudgetWithSpent, error)
	GetBudgetsByMember(ctx context.Context, memberID string) ([]*BudgetWithSpent, error)

	// Budget period calculations
	GetCurrentPeriodBudgets(ctx context.Context) ([]*BudgetWithSpent, error)
	GetBudgetForPeriod(ctx context.Context, categoryID, memberID *string, period string, date time.Time) (*BudgetWithSpent, error)
	CalculateBudgetSpent(ctx context.Context, budgetID string) (float64, error)

	// Budget validation and constraints
	ValidateBudgetConstraints(ctx context.Context, req *CreateBudgetRequest) error
	CheckBudgetConflicts(ctx context.Context, categoryID, memberID *string, period string, startDate, endDate time.Time, excludeBudgetID *string) (bool, error)

	// Utility operations
	BudgetExists(ctx context.Context, budgetID string) (bool, error)
	DeactivateExpiredBudgets(ctx context.Context) (int, error)
	GetBudgetPeriodDates(period string, referenceDate time.Time) (startDate, endDate time.Time)
}