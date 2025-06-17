package budget

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

// Repository implements the Store interface for budget operations
type Repository struct {
	db     *sql.DB
	logger zerolog.Logger
}

// NewRepository creates a new budget repository that operates on a family database
func NewRepository(db *sql.DB, logger zerolog.Logger) Store {
	return &Repository{
		db:     db,
		logger: logger.With().Str("component", "budget-repository").Logger(),
	}
}

// CreateBudget creates a new budget in the family database
func (r *Repository) CreateBudget(ctx context.Context, req *CreateBudgetRequest) (*Budget, error) {
	// Validate the request
	if err := r.ValidateBudgetConstraints(ctx, req); err != nil {
		return nil, err
	}

	// Check for conflicts with existing budgets
	endDate := req.EndDate
	if endDate == nil {
		// Calculate default end date based on period
		_, calculatedEndDate := r.GetBudgetPeriodDates(req.Period, req.StartDate)
		endDate = &calculatedEndDate
	}

	hasConflicts, err := r.CheckBudgetConflicts(ctx, req.CategoryID, req.MemberID, req.Period, req.StartDate, *endDate, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to check budget conflicts: %w", err)
	}
	if hasConflicts {
		return nil, fmt.Errorf("budget conflicts with existing budget for the same category/member and period")
	}

	budgetID, err := r.generateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate budget ID: %w", err)
	}

	now := time.Now()
	budget := &Budget{
		ID:         budgetID,
		CategoryID: req.CategoryID,
		MemberID:   req.MemberID,
		Amount:     req.Amount,
		Period:     req.Period,
		StartDate:  req.StartDate,
		EndDate:    endDate,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	query := `
		INSERT INTO budgets (id, category_id, member_id, amount, period, start_date, end_date, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, query,
		budget.ID, budget.CategoryID, budget.MemberID, budget.Amount, budget.Period,
		budget.StartDate, budget.EndDate, budget.IsActive, budget.CreatedAt, budget.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create budget: %w", err)
	}

	r.logger.Info().
		Str("budget_id", budget.ID).
		Float64("amount", budget.Amount).
		Str("period", budget.Period).
		Msg("Budget created successfully")

	return budget, nil
}

// GetBudgetByID retrieves a budget by its ID
func (r *Repository) GetBudgetByID(ctx context.Context, budgetID string) (*Budget, error) {
	budget := &Budget{}
	query := `
		SELECT id, category_id, member_id, amount, period, start_date, end_date, is_active, created_at, updated_at
		FROM budgets WHERE id = ?`

	err := r.db.QueryRowContext(ctx, query, budgetID).Scan(
		&budget.ID, &budget.CategoryID, &budget.MemberID, &budget.Amount, &budget.Period,
		&budget.StartDate, &budget.EndDate, &budget.IsActive, &budget.CreatedAt, &budget.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("budget not found: %s", budgetID)
		}
		return nil, fmt.Errorf("failed to get budget: %w", err)
	}

	return budget, nil
}

// GetBudgetWithSpent retrieves a budget with current spending information
func (r *Repository) GetBudgetWithSpent(ctx context.Context, budgetID string) (*BudgetWithSpent, error) {
	budget, err := r.GetBudgetByID(ctx, budgetID)
	if err != nil {
		return nil, err
	}

	amountSpent, err := r.CalculateBudgetSpent(ctx, budgetID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate budget spent: %w", err)
	}

	remaining := budget.Amount - amountSpent
	percentage := 0.0
	if budget.Amount > 0 {
		percentage = (amountSpent / budget.Amount) * 100
	}

	return &BudgetWithSpent{
		Budget:       budget,
		AmountSpent:  amountSpent,
		Remaining:    remaining,
		Percentage:   percentage,
		IsOverBudget: amountSpent > budget.Amount,
	}, nil
}

// UpdateBudget updates an existing budget
func (r *Repository) UpdateBudget(ctx context.Context, budgetID string, req *UpdateBudgetRequest) (*Budget, error) {
	// First, get the current budget
	budget, err := r.GetBudgetByID(ctx, budgetID)
	if err != nil {
		return nil, err
	}

	// Build update query dynamically based on provided fields
	var setParts []string
	var args []interface{}

	if req.CategoryID != nil {
		setParts = append(setParts, "category_id = ?")
		args = append(args, *req.CategoryID)
		budget.CategoryID = req.CategoryID
	}
	if req.MemberID != nil {
		setParts = append(setParts, "member_id = ?")
		args = append(args, *req.MemberID)
		budget.MemberID = req.MemberID
	}
	if req.Amount != nil {
		setParts = append(setParts, "amount = ?")
		args = append(args, *req.Amount)
		budget.Amount = *req.Amount
	}
	if req.Period != nil {
		setParts = append(setParts, "period = ?")
		args = append(args, *req.Period)
		budget.Period = *req.Period
	}
	if req.StartDate != nil {
		setParts = append(setParts, "start_date = ?")
		args = append(args, *req.StartDate)
		budget.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		setParts = append(setParts, "end_date = ?")
		args = append(args, *req.EndDate)
		budget.EndDate = req.EndDate
	}
	if req.IsActive != nil {
		setParts = append(setParts, "is_active = ?")
		args = append(args, *req.IsActive)
		budget.IsActive = *req.IsActive
	}

	if len(setParts) > 0 {
		now := time.Now()
		setParts = append(setParts, "updated_at = ?")
		args = append(args, now)
		budget.UpdatedAt = now

		query := fmt.Sprintf("UPDATE budgets SET %s WHERE id = ?", strings.Join(setParts, ", "))
		args = append(args, budgetID)

		_, err = r.db.ExecContext(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to update budget: %w", err)
		}

		r.logger.Info().
			Str("budget_id", budgetID).
			Msg("Budget updated successfully")
	}

	return budget, nil
}

// DeleteBudget removes a budget
func (r *Repository) DeleteBudget(ctx context.Context, budgetID string) error {
	query := `DELETE FROM budgets WHERE id = ?`
	
	result, err := r.db.ExecContext(ctx, query, budgetID)
	if err != nil {
		return fmt.Errorf("failed to delete budget: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("budget not found: %s", budgetID)
	}

	r.logger.Info().
		Str("budget_id", budgetID).
		Msg("Budget deleted successfully")

	return nil
}

// ListBudgets retrieves budgets based on filter criteria
func (r *Repository) ListBudgets(ctx context.Context, filter *BudgetFilter) ([]*Budget, error) {
	query := `
		SELECT id, category_id, member_id, amount, period, start_date, end_date, is_active, created_at, updated_at
		FROM budgets`

	whereClause, args := r.buildWhereClause(filter)
	if whereClause != "" {
		query += " WHERE " + whereClause
	}

	query += " ORDER BY start_date DESC, created_at DESC"

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
		return nil, fmt.Errorf("failed to list budgets: %w", err)
	}
	defer rows.Close()

	var budgets []*Budget
	for rows.Next() {
		budget := &Budget{}
		err := rows.Scan(
			&budget.ID, &budget.CategoryID, &budget.MemberID, &budget.Amount, &budget.Period,
			&budget.StartDate, &budget.EndDate, &budget.IsActive, &budget.CreatedAt, &budget.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan budget: %w", err)
		}
		budgets = append(budgets, budget)
	}

	return budgets, rows.Err()
}

// ListBudgetsWithSpent retrieves budgets with spending information
func (r *Repository) ListBudgetsWithSpent(ctx context.Context, filter *BudgetFilter) ([]*BudgetWithSpent, error) {
	budgets, err := r.ListBudgets(ctx, filter)
	if err != nil {
		return nil, err
	}

	var budgetsWithSpent []*BudgetWithSpent
	for _, budget := range budgets {
		amountSpent, err := r.CalculateBudgetSpent(ctx, budget.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate spent for budget %s: %w", budget.ID, err)
		}

		remaining := budget.Amount - amountSpent
		percentage := 0.0
		if budget.Amount > 0 {
			percentage = (amountSpent / budget.Amount) * 100
		}

		budgetsWithSpent = append(budgetsWithSpent, &BudgetWithSpent{
			Budget:       budget,
			AmountSpent:  amountSpent,
			Remaining:    remaining,
			Percentage:   percentage,
			IsOverBudget: amountSpent > budget.Amount,
		})
	}

	return budgetsWithSpent, nil
}

// CountBudgets counts budgets based on filter criteria
func (r *Repository) CountBudgets(ctx context.Context, filter *BudgetFilter) (int, error) {
	query := "SELECT COUNT(*) FROM budgets"
	
	whereClause, args := r.buildWhereClause(filter)
	if whereClause != "" {
		query += " WHERE " + whereClause
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count budgets: %w", err)
	}

	return count, nil
}

// GetBudgetSummary generates budget summary statistics
func (r *Repository) GetBudgetSummary(ctx context.Context, filter *BudgetFilter) (*BudgetSummary, error) {
	// Get basic budget counts and totals
	baseQuery := "FROM budgets"
	whereClause, args := r.buildWhereClause(filter)
	if whereClause != "" {
		baseQuery += " WHERE " + whereClause
	}

	summaryQuery := "SELECT COUNT(*), COUNT(CASE WHEN is_active = 1 THEN 1 END), COALESCE(SUM(amount), 0) " + baseQuery
	
	var totalBudgets, activeBudgets int
	var totalAmount float64
	err := r.db.QueryRowContext(ctx, summaryQuery, args...).Scan(&totalBudgets, &activeBudgets, &totalAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to get budget summary: %w", err)
	}

	summary := &BudgetSummary{
		TotalBudgets:  totalBudgets,
		ActiveBudgets: activeBudgets,
		TotalAmount:   totalAmount,
		ByCategory:    make(map[string]*BudgetWithSpent),
		ByMember:      make(map[string]*BudgetWithSpent),
		ByPeriod:      make(map[string]int),
	}

	// Calculate total spent and remaining (this is expensive but necessary for accuracy)
	budgetsWithSpent, err := r.ListBudgetsWithSpent(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get budgets with spent: %w", err)
	}

	var totalSpent, totalRemaining float64
	var overBudgetCount int

	for _, bws := range budgetsWithSpent {
		totalSpent += bws.AmountSpent
		totalRemaining += bws.Remaining
		if bws.IsOverBudget {
			overBudgetCount++
		}

		// Group by period
		summary.ByPeriod[bws.Budget.Period]++
	}

	summary.TotalSpent = totalSpent
	summary.TotalRemaining = totalRemaining
	summary.OverBudgetCount = overBudgetCount

	return summary, nil
}

// GetActiveBudgets retrieves all active budgets with spending information
func (r *Repository) GetActiveBudgets(ctx context.Context) ([]*BudgetWithSpent, error) {
	filter := &BudgetFilter{
		IsActive: &[]bool{true}[0],
	}
	return r.ListBudgetsWithSpent(ctx, filter)
}

// GetOverBudgetAlerts retrieves budgets that have exceeded their limits
func (r *Repository) GetOverBudgetAlerts(ctx context.Context) ([]*BudgetAlert, error) {
	activeBudgets, err := r.GetActiveBudgets(ctx)
	if err != nil {
		return nil, err
	}

	var alerts []*BudgetAlert
	now := time.Now()

	for _, bws := range activeBudgets {
		if bws.Percentage >= 80 { // Alert at 80% and above
			alertType := "warning"
			if bws.IsOverBudget {
				alertType = "exceeded"
			}

			alerts = append(alerts, &BudgetAlert{
				BudgetID:    bws.Budget.ID,
				Budget:      bws.Budget,
				AmountSpent: bws.AmountSpent,
				Percentage:  bws.Percentage,
				AlertType:   alertType,
				AlertDate:   now,
			})
		}
	}

	return alerts, nil
}

// GetBudgetsByCategory retrieves budgets for a specific category
func (r *Repository) GetBudgetsByCategory(ctx context.Context, categoryID string) ([]*BudgetWithSpent, error) {
	filter := &BudgetFilter{
		CategoryID: &categoryID,
	}
	return r.ListBudgetsWithSpent(ctx, filter)
}

// GetBudgetsByMember retrieves budgets for a specific member
func (r *Repository) GetBudgetsByMember(ctx context.Context, memberID string) ([]*BudgetWithSpent, error) {
	filter := &BudgetFilter{
		MemberID: &memberID,
	}
	return r.ListBudgetsWithSpent(ctx, filter)
}

// GetCurrentPeriodBudgets retrieves budgets that are active in the current period
func (r *Repository) GetCurrentPeriodBudgets(ctx context.Context) ([]*BudgetWithSpent, error) {
	now := time.Now()
	filter := &BudgetFilter{
		IsActive:  &[]bool{true}[0],
		StartDate: &now, // Budgets that started before or on today
	}

	budgets, err := r.ListBudgetsWithSpent(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Filter budgets that are currently active (haven't ended yet)
	var currentBudgets []*BudgetWithSpent
	for _, budget := range budgets {
		if budget.Budget.EndDate == nil || budget.Budget.EndDate.After(now) {
			currentBudgets = append(currentBudgets, budget)
		}
	}

	return currentBudgets, nil
}

// GetBudgetForPeriod finds an active budget for a specific category/member and period
func (r *Repository) GetBudgetForPeriod(ctx context.Context, categoryID, memberID *string, period string, date time.Time) (*BudgetWithSpent, error) {
	query := `
		SELECT id, category_id, member_id, amount, period, start_date, end_date, is_active, created_at, updated_at
		FROM budgets 
		WHERE period = ? AND is_active = 1 
		AND start_date <= ? AND (end_date IS NULL OR end_date >= ?)
		AND (? IS NULL OR category_id = ?) 
		AND (? IS NULL OR member_id = ?)
		ORDER BY created_at DESC LIMIT 1`

	budget := &Budget{}
	err := r.db.QueryRowContext(ctx, query, period, date, date, categoryID, categoryID, memberID, memberID).Scan(
		&budget.ID, &budget.CategoryID, &budget.MemberID, &budget.Amount, &budget.Period,
		&budget.StartDate, &budget.EndDate, &budget.IsActive, &budget.CreatedAt, &budget.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no active budget found for period %s", period)
		}
		return nil, fmt.Errorf("failed to get budget for period: %w", err)
	}

	return r.GetBudgetWithSpent(ctx, budget.ID)
}

// CalculateBudgetSpent calculates the total amount spent against a budget
func (r *Repository) CalculateBudgetSpent(ctx context.Context, budgetID string) (float64, error) {
	budget, err := r.GetBudgetByID(ctx, budgetID)
	if err != nil {
		return 0, err
	}

	query := `
		SELECT COALESCE(SUM(amount), 0) 
		FROM expenses 
		WHERE date BETWEEN ? AND ? `

	args := []interface{}{budget.StartDate}
	
	endDate := time.Now()
	if budget.EndDate != nil && budget.EndDate.Before(endDate) {
		endDate = *budget.EndDate
	}
	args = append(args, endDate)

	// Add category filter if budget is category-specific
	if budget.CategoryID != nil {
		query += " AND category_id = ?"
		args = append(args, *budget.CategoryID)
	}

	// Add member filter if budget is member-specific
	if budget.MemberID != nil {
		query += " AND member_id = ?"
		args = append(args, *budget.MemberID)
	}

	var amountSpent float64
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&amountSpent)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate budget spent: %w", err)
	}

	return amountSpent, nil
}

// ValidateBudgetConstraints validates budget creation/update constraints
func (r *Repository) ValidateBudgetConstraints(ctx context.Context, req *CreateBudgetRequest) error {
	if req.Amount <= 0 {
		return fmt.Errorf("budget amount must be positive")
	}

	validPeriods := map[string]bool{
		"weekly":  true,
		"monthly": true,
		"yearly":  true,
	}

	if !validPeriods[req.Period] {
		return fmt.Errorf("invalid period: %s (must be weekly, monthly, or yearly)", req.Period)
	}

	if req.EndDate != nil && req.EndDate.Before(req.StartDate) {
		return fmt.Errorf("end date cannot be before start date")
	}

	return nil
}

// CheckBudgetConflicts checks if a budget conflicts with existing budgets
func (r *Repository) CheckBudgetConflicts(ctx context.Context, categoryID, memberID *string, period string, startDate, endDate time.Time, excludeBudgetID *string) (bool, error) {
	query := `
		SELECT COUNT(*) FROM budgets 
		WHERE period = ? AND is_active = 1
		AND (start_date <= ? AND (end_date IS NULL OR end_date >= ?))
		AND (? IS NULL OR category_id = ?) 
		AND (? IS NULL OR member_id = ?)`

	args := []interface{}{period, endDate, startDate, categoryID, categoryID, memberID, memberID}

	if excludeBudgetID != nil {
		query += " AND id != ?"
		args = append(args, *excludeBudgetID)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check budget conflicts: %w", err)
	}

	return count > 0, nil
}

// BudgetExists checks if a budget exists
func (r *Repository) BudgetExists(ctx context.Context, budgetID string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM budgets WHERE id = ?`
	err := r.db.QueryRowContext(ctx, query, budgetID).Scan(&count)
	return count > 0, err
}

// DeactivateExpiredBudgets deactivates budgets that have passed their end date
func (r *Repository) DeactivateExpiredBudgets(ctx context.Context) (int, error) {
	now := time.Now()
	query := `UPDATE budgets SET is_active = 0, updated_at = ? WHERE is_active = 1 AND end_date IS NOT NULL AND end_date < ?`
	
	result, err := r.db.ExecContext(ctx, query, now, now)
	if err != nil {
		return 0, fmt.Errorf("failed to deactivate expired budgets: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	
	if rowsAffected > 0 {
		r.logger.Info().
			Int64("budgets_deactivated", rowsAffected).
			Msg("Deactivated expired budgets")
	}

	return int(rowsAffected), nil
}

// GetBudgetPeriodDates calculates start and end dates for a budget period
func (r *Repository) GetBudgetPeriodDates(period string, referenceDate time.Time) (startDate, endDate time.Time) {
	switch period {
	case "weekly":
		// Start of week (Monday)
		weekday := int(referenceDate.Weekday())
		if weekday == 0 { // Sunday
			weekday = 7
		}
		startDate = referenceDate.AddDate(0, 0, -weekday+1)
		endDate = startDate.AddDate(0, 0, 6)
	case "monthly":
		// Start of month
		startDate = time.Date(referenceDate.Year(), referenceDate.Month(), 1, 0, 0, 0, 0, referenceDate.Location())
		endDate = startDate.AddDate(0, 1, -1)
	case "yearly":
		// Start of year
		startDate = time.Date(referenceDate.Year(), 1, 1, 0, 0, 0, 0, referenceDate.Location())
		endDate = startDate.AddDate(1, 0, -1)
	default:
		// Default to monthly
		startDate = time.Date(referenceDate.Year(), referenceDate.Month(), 1, 0, 0, 0, 0, referenceDate.Location())
		endDate = startDate.AddDate(0, 1, -1)
	}

	// Set time to end of day for end date
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())
	
	return startDate, endDate
}

// Helper methods

// buildWhereClause builds the WHERE clause for filtering budgets
func (r *Repository) buildWhereClause(filter *BudgetFilter) (string, []interface{}) {
	if filter == nil {
		return "", nil
	}

	var conditions []string
	var args []interface{}

	if filter.CategoryID != nil {
		conditions = append(conditions, "category_id = ?")
		args = append(args, *filter.CategoryID)
	}

	if filter.MemberID != nil {
		conditions = append(conditions, "member_id = ?")
		args = append(args, *filter.MemberID)
	}

	if filter.Period != nil {
		conditions = append(conditions, "period = ?")
		args = append(args, *filter.Period)
	}

	if filter.IsActive != nil {
		conditions = append(conditions, "is_active = ?")
		args = append(args, *filter.IsActive)
	}

	if filter.StartDate != nil {
		conditions = append(conditions, "start_date <= ?")
		args = append(args, *filter.StartDate)
	}

	if filter.EndDate != nil {
		conditions = append(conditions, "(end_date IS NULL OR end_date >= ?)")
		args = append(args, *filter.EndDate)
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