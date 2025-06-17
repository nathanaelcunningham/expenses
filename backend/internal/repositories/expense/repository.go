package expense

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

// Repository implements the Store interface for expense operations
type Repository struct {
	db     *sql.DB
	logger zerolog.Logger
}

// NewRepository creates a new expense repository that operates on a family database
func NewRepository(db *sql.DB, logger zerolog.Logger) Store {
	return &Repository{
		db:     db,
		logger: logger.With().Str("component", "expense-repository").Logger(),
	}
}

// CreateExpense creates a new expense in the family database
func (r *Repository) CreateExpense(ctx context.Context, req *CreateExpenseRequest) (*Expense, error) {
	expenseID, err := r.generateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate expense ID: %w", err)
	}

	now := time.Now()
	expense := &Expense{
		ID:                expenseID,
		MemberID:          req.MemberID,
		CategoryID:        req.CategoryID,
		Amount:            req.Amount,
		Currency:          req.Currency,
		Description:       strings.TrimSpace(req.Description),
		Date:              req.Date,
		ReceiptURL:        req.ReceiptURL,
		Tags:              req.Tags,
		IsRecurring:       req.IsRecurring,
		RecurringInterval: req.RecurringInterval,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if expense.Currency == "" {
		expense.Currency = "USD"
	}

	// Begin transaction for atomic expense creation
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert expense
	query := `
		INSERT INTO expenses (id, member_id, category_id, amount, currency, description, date, 
		                     receipt_url, tags, is_recurring, recurring_interval, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.ExecContext(ctx, query,
		expense.ID, expense.MemberID, expense.CategoryID, expense.Amount, expense.Currency,
		expense.Description, expense.Date, expense.ReceiptURL, expense.Tags,
		expense.IsRecurring, expense.RecurringInterval, expense.CreatedAt, expense.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create expense: %w", err)
	}

	// Create splits if provided
	if len(req.Splits) > 0 {
		for _, split := range req.Splits {
			splitID, err := r.generateID()
			if err != nil {
				return nil, fmt.Errorf("failed to generate split ID: %w", err)
			}

			splitQuery := `
				INSERT INTO expense_splits (id, expense_id, member_id, amount, percentage, created_at)
				VALUES (?, ?, ?, ?, ?, ?)`

			_, err = tx.ExecContext(ctx, splitQuery,
				splitID, expense.ID, split.MemberID, split.Amount, split.Percentage, now)
			if err != nil {
				return nil, fmt.Errorf("failed to create expense split: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit expense creation: %w", err)
	}

	r.logger.Info().
		Str("expense_id", expense.ID).
		Str("member_id", expense.MemberID).
		Float64("amount", expense.Amount).
		Str("currency", expense.Currency).
		Msg("Expense created successfully")

	return expense, nil
}

// GetExpenseByID retrieves an expense by its ID
func (r *Repository) GetExpenseByID(ctx context.Context, expenseID string) (*Expense, error) {
	expense := &Expense{}
	query := `
		SELECT id, member_id, category_id, amount, currency, description, date,
		       receipt_url, tags, is_recurring, recurring_interval, created_at, updated_at
		FROM expenses WHERE id = ?`

	err := r.db.QueryRowContext(ctx, query, expenseID).Scan(
		&expense.ID, &expense.MemberID, &expense.CategoryID, &expense.Amount,
		&expense.Currency, &expense.Description, &expense.Date, &expense.ReceiptURL,
		&expense.Tags, &expense.IsRecurring, &expense.RecurringInterval,
		&expense.CreatedAt, &expense.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("expense not found: %s", expenseID)
		}
		return nil, fmt.Errorf("failed to get expense: %w", err)
	}

	return expense, nil
}

// GetExpenseWithSplits retrieves an expense with its splits
func (r *Repository) GetExpenseWithSplits(ctx context.Context, expenseID string) (*ExpenseWithSplits, error) {
	expense, err := r.GetExpenseByID(ctx, expenseID)
	if err != nil {
		return nil, err
	}

	splits, err := r.GetExpenseSplits(ctx, expenseID)
	if err != nil {
		return nil, err
	}

	return &ExpenseWithSplits{
		Expense: expense,
		Splits:  splits,
	}, nil
}

// UpdateExpense updates an existing expense
func (r *Repository) UpdateExpense(ctx context.Context, expenseID string, req *UpdateExpenseRequest) (*Expense, error) {
	// First, get the current expense
	expense, err := r.GetExpenseByID(ctx, expenseID)
	if err != nil {
		return nil, err
	}

	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Build update query dynamically based on provided fields
	var setParts []string
	var args []interface{}

	if req.CategoryID != nil {
		setParts = append(setParts, "category_id = ?")
		args = append(args, *req.CategoryID)
		expense.CategoryID = req.CategoryID
	}
	if req.Amount != nil {
		setParts = append(setParts, "amount = ?")
		args = append(args, *req.Amount)
		expense.Amount = *req.Amount
	}
	if req.Currency != nil {
		setParts = append(setParts, "currency = ?")
		args = append(args, *req.Currency)
		expense.Currency = *req.Currency
	}
	if req.Description != nil {
		setParts = append(setParts, "description = ?")
		args = append(args, *req.Description)
		expense.Description = *req.Description
	}
	if req.Date != nil {
		setParts = append(setParts, "date = ?")
		args = append(args, *req.Date)
		expense.Date = *req.Date
	}
	if req.ReceiptURL != nil {
		setParts = append(setParts, "receipt_url = ?")
		args = append(args, *req.ReceiptURL)
		expense.ReceiptURL = req.ReceiptURL
	}
	if req.Tags != nil {
		setParts = append(setParts, "tags = ?")
		args = append(args, *req.Tags)
		expense.Tags = req.Tags
	}
	if req.IsRecurring != nil {
		setParts = append(setParts, "is_recurring = ?")
		args = append(args, *req.IsRecurring)
		expense.IsRecurring = *req.IsRecurring
	}
	if req.RecurringInterval != nil {
		setParts = append(setParts, "recurring_interval = ?")
		args = append(args, *req.RecurringInterval)
		expense.RecurringInterval = req.RecurringInterval
	}

	if len(setParts) > 0 {
		now := time.Now()
		setParts = append(setParts, "updated_at = ?")
		args = append(args, now)
		expense.UpdatedAt = now

		query := fmt.Sprintf("UPDATE expenses SET %s WHERE id = ?", strings.Join(setParts, ", "))
		args = append(args, expenseID)

		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to update expense: %w", err)
		}
	}

	// Update splits if provided
	if req.Splits != nil {
		// Delete existing splits
		if err = r.deleteAllExpenseSplitsInTx(ctx, tx, expenseID); err != nil {
			return nil, fmt.Errorf("failed to delete existing splits: %w", err)
		}

		// Create new splits
		for _, split := range req.Splits {
			splitID, err := r.generateID()
			if err != nil {
				return nil, fmt.Errorf("failed to generate split ID: %w", err)
			}

			splitQuery := `
				INSERT INTO expense_splits (id, expense_id, member_id, amount, percentage, created_at)
				VALUES (?, ?, ?, ?, ?, ?)`

			_, err = tx.ExecContext(ctx, splitQuery,
				splitID, expenseID, split.MemberID, split.Amount, split.Percentage, time.Now())
			if err != nil {
				return nil, fmt.Errorf("failed to create expense split: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit expense update: %w", err)
	}

	r.logger.Info().
		Str("expense_id", expenseID).
		Msg("Expense updated successfully")

	return expense, nil
}

// DeleteExpense removes an expense and its splits
func (r *Repository) DeleteExpense(ctx context.Context, expenseID string) error {
	// The foreign key constraints will cascade delete expense_splits
	query := `DELETE FROM expenses WHERE id = ?`
	
	result, err := r.db.ExecContext(ctx, query, expenseID)
	if err != nil {
		return fmt.Errorf("failed to delete expense: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("expense not found: %s", expenseID)
	}

	r.logger.Info().
		Str("expense_id", expenseID).
		Msg("Expense deleted successfully")

	return nil
}

// ListExpenses retrieves expenses based on filter criteria
func (r *Repository) ListExpenses(ctx context.Context, filter *ExpenseFilter) ([]*Expense, error) {
	query := `
		SELECT id, member_id, category_id, amount, currency, description, date,
		       receipt_url, tags, is_recurring, recurring_interval, created_at, updated_at
		FROM expenses`

	whereClause, args := r.buildWhereClause(filter)
	if whereClause != "" {
		query += " WHERE " + whereClause
	}

	query += " ORDER BY date DESC, created_at DESC"

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
		return nil, fmt.Errorf("failed to list expenses: %w", err)
	}
	defer rows.Close()

	var expenses []*Expense
	for rows.Next() {
		expense := &Expense{}
		err := rows.Scan(
			&expense.ID, &expense.MemberID, &expense.CategoryID, &expense.Amount,
			&expense.Currency, &expense.Description, &expense.Date, &expense.ReceiptURL,
			&expense.Tags, &expense.IsRecurring, &expense.RecurringInterval,
			&expense.CreatedAt, &expense.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expense: %w", err)
		}
		expenses = append(expenses, expense)
	}

	return expenses, rows.Err()
}

// ListExpensesWithSplits retrieves expenses with their splits based on filter criteria
func (r *Repository) ListExpensesWithSplits(ctx context.Context, filter *ExpenseFilter) ([]*ExpenseWithSplits, error) {
	expenses, err := r.ListExpenses(ctx, filter)
	if err != nil {
		return nil, err
	}

	var expensesWithSplits []*ExpenseWithSplits
	for _, expense := range expenses {
		splits, err := r.GetExpenseSplits(ctx, expense.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get splits for expense %s: %w", expense.ID, err)
		}

		expensesWithSplits = append(expensesWithSplits, &ExpenseWithSplits{
			Expense: expense,
			Splits:  splits,
		})
	}

	return expensesWithSplits, nil
}

// CountExpenses counts expenses based on filter criteria
func (r *Repository) CountExpenses(ctx context.Context, filter *ExpenseFilter) (int, error) {
	query := "SELECT COUNT(*) FROM expenses"
	
	whereClause, args := r.buildWhereClause(filter)
	if whereClause != "" {
		query += " WHERE " + whereClause
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count expenses: %w", err)
	}

	return count, nil
}

// CreateExpenseSplit creates a new expense split
func (r *Repository) CreateExpenseSplit(ctx context.Context, expenseID string, req *CreateSplitRequest) (*ExpenseSplit, error) {
	splitID, err := r.generateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate split ID: %w", err)
	}

	now := time.Now()
	split := &ExpenseSplit{
		ID:         splitID,
		ExpenseID:  expenseID,
		MemberID:   req.MemberID,
		Amount:     req.Amount,
		Percentage: req.Percentage,
		CreatedAt:  now,
	}

	query := `
		INSERT INTO expense_splits (id, expense_id, member_id, amount, percentage, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, query,
		split.ID, split.ExpenseID, split.MemberID, split.Amount, split.Percentage, split.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create expense split: %w", err)
	}

	return split, nil
}

// GetExpenseSplits retrieves all splits for an expense
func (r *Repository) GetExpenseSplits(ctx context.Context, expenseID string) ([]*ExpenseSplit, error) {
	query := `
		SELECT id, expense_id, member_id, amount, percentage, created_at
		FROM expense_splits WHERE expense_id = ? ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, expenseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get expense splits: %w", err)
	}
	defer rows.Close()

	var splits []*ExpenseSplit
	for rows.Next() {
		split := &ExpenseSplit{}
		err := rows.Scan(&split.ID, &split.ExpenseID, &split.MemberID, 
			&split.Amount, &split.Percentage, &split.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expense split: %w", err)
		}
		splits = append(splits, split)
	}

	return splits, rows.Err()
}

// UpdateExpenseSplit updates an expense split
func (r *Repository) UpdateExpenseSplit(ctx context.Context, splitID string, amount float64, percentage *float64) (*ExpenseSplit, error) {
	query := `UPDATE expense_splits SET amount = ?, percentage = ? WHERE id = ?`
	
	result, err := r.db.ExecContext(ctx, query, amount, percentage, splitID)
	if err != nil {
		return nil, fmt.Errorf("failed to update expense split: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, fmt.Errorf("expense split not found: %s", splitID)
	}

	// Return updated split
	split := &ExpenseSplit{}
	getQuery := `
		SELECT id, expense_id, member_id, amount, percentage, created_at
		FROM expense_splits WHERE id = ?`

	err = r.db.QueryRowContext(ctx, getQuery, splitID).Scan(
		&split.ID, &split.ExpenseID, &split.MemberID, 
		&split.Amount, &split.Percentage, &split.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated split: %w", err)
	}

	return split, nil
}

// DeleteExpenseSplit removes an expense split
func (r *Repository) DeleteExpenseSplit(ctx context.Context, splitID string) error {
	query := `DELETE FROM expense_splits WHERE id = ?`
	
	result, err := r.db.ExecContext(ctx, query, splitID)
	if err != nil {
		return fmt.Errorf("failed to delete expense split: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("expense split not found: %s", splitID)
	}

	return nil
}

// DeleteAllExpenseSplits removes all splits for an expense
func (r *Repository) DeleteAllExpenseSplits(ctx context.Context, expenseID string) error {
	query := `DELETE FROM expense_splits WHERE expense_id = ?`
	_, err := r.db.ExecContext(ctx, query, expenseID)
	if err != nil {
		return fmt.Errorf("failed to delete expense splits: %w", err)
	}
	return nil
}

// GetExpenseSummary generates expense summary statistics
func (r *Repository) GetExpenseSummary(ctx context.Context, filter *ExpenseFilter) (*ExpenseSummary, error) {
	// Base query for summary
	baseQuery := "FROM expenses"
	whereClause, args := r.buildWhereClause(filter)
	if whereClause != "" {
		baseQuery += " WHERE " + whereClause
	}

	// Get total amount and count
	summaryQuery := "SELECT COALESCE(SUM(amount), 0), COUNT(*), COALESCE(AVG(amount), 0) " + baseQuery
	
	var totalAmount, averageAmount float64
	var count int
	err := r.db.QueryRowContext(ctx, summaryQuery, args...).Scan(&totalAmount, &count, &averageAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to get expense summary: %w", err)
	}

	summary := &ExpenseSummary{
		TotalAmount:   totalAmount,
		ExpenseCount:  count,
		AverageAmount: averageAmount,
		ByCategory:    make(map[string]float64),
		ByMember:      make(map[string]float64),
		Currency:      "USD", // Default currency
	}

	// Get breakdown by category
	categoryQuery := "SELECT COALESCE(category_id, 'uncategorized') as category, SUM(amount) " + baseQuery + " GROUP BY category_id"
	rows, err := r.db.QueryContext(ctx, categoryQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get category breakdown: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		var amount float64
		if err := rows.Scan(&category, &amount); err != nil {
			return nil, fmt.Errorf("failed to scan category breakdown: %w", err)
		}
		summary.ByCategory[category] = amount
	}

	// Get breakdown by member
	memberQuery := "SELECT member_id, SUM(amount) " + baseQuery + " GROUP BY member_id"
	rows, err = r.db.QueryContext(ctx, memberQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get member breakdown: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var memberID string
		var amount float64
		if err := rows.Scan(&memberID, &amount); err != nil {
			return nil, fmt.Errorf("failed to scan member breakdown: %w", err)
		}
		summary.ByMember[memberID] = amount
	}

	return summary, nil
}

// GetExpensesByDateRange retrieves expenses within a date range
func (r *Repository) GetExpensesByDateRange(ctx context.Context, memberID *string, startDate, endDate time.Time) ([]*Expense, error) {
	filter := &ExpenseFilter{
		DateFrom: &startDate,
		DateTo:   &endDate,
	}
	if memberID != nil {
		filter.MemberID = memberID
	}

	return r.ListExpenses(ctx, filter)
}

// GetTotalSpentByMember calculates total spent by a member in a date range
func (r *Repository) GetTotalSpentByMember(ctx context.Context, memberID string, startDate, endDate time.Time) (float64, error) {
	query := `SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE member_id = ? AND date BETWEEN ? AND ?`
	
	var total float64
	err := r.db.QueryRowContext(ctx, query, memberID, startDate, endDate).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get total spent by member: %w", err)
	}

	return total, nil
}

// GetTotalSpentByCategory calculates total spent in a category in a date range
func (r *Repository) GetTotalSpentByCategory(ctx context.Context, categoryID string, startDate, endDate time.Time) (float64, error) {
	query := `SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE category_id = ? AND date BETWEEN ? AND ?`
	
	var total float64
	err := r.db.QueryRowContext(ctx, query, categoryID, startDate, endDate).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get total spent by category: %w", err)
	}

	return total, nil
}

// GetRecurringExpenses retrieves all recurring expenses
func (r *Repository) GetRecurringExpenses(ctx context.Context) ([]*Expense, error) {
	filter := &ExpenseFilter{
		IsRecurring: &[]bool{true}[0],
	}
	return r.ListExpenses(ctx, filter)
}

// GetExpensesByMember retrieves expenses for a specific member with pagination
func (r *Repository) GetExpensesByMember(ctx context.Context, memberID string, limit, offset int) ([]*Expense, error) {
	filter := &ExpenseFilter{
		MemberID: &memberID,
		Limit:    limit,
		Offset:   offset,
	}
	return r.ListExpenses(ctx, filter)
}

// ExpenseExists checks if an expense exists
func (r *Repository) ExpenseExists(ctx context.Context, expenseID string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM expenses WHERE id = ?`
	err := r.db.QueryRowContext(ctx, query, expenseID).Scan(&count)
	return count > 0, err
}

// UserCanAccessExpense checks if a user can access an expense (is the member who created it or part of splits)
func (r *Repository) UserCanAccessExpense(ctx context.Context, expenseID, userID string) (bool, error) {
	query := `
		SELECT COUNT(*) FROM expenses e
		LEFT JOIN expense_splits es ON e.id = es.expense_id
		WHERE e.id = ? AND (e.member_id = ? OR es.member_id = ?)`
	
	var count int
	err := r.db.QueryRowContext(ctx, query, expenseID, userID, userID).Scan(&count)
	return count > 0, err
}

// Helper methods

// buildWhereClause builds the WHERE clause for filtering expenses
func (r *Repository) buildWhereClause(filter *ExpenseFilter) (string, []interface{}) {
	if filter == nil {
		return "", nil
	}

	var conditions []string
	var args []interface{}

	if filter.MemberID != nil {
		conditions = append(conditions, "member_id = ?")
		args = append(args, *filter.MemberID)
	}

	if filter.CategoryID != nil {
		conditions = append(conditions, "category_id = ?")
		args = append(args, *filter.CategoryID)
	}

	if filter.DateFrom != nil {
		conditions = append(conditions, "date >= ?")
		args = append(args, *filter.DateFrom)
	}

	if filter.DateTo != nil {
		conditions = append(conditions, "date <= ?")
		args = append(args, *filter.DateTo)
	}

	if filter.MinAmount != nil {
		conditions = append(conditions, "amount >= ?")
		args = append(args, *filter.MinAmount)
	}

	if filter.MaxAmount != nil {
		conditions = append(conditions, "amount <= ?")
		args = append(args, *filter.MaxAmount)
	}

	if filter.IsRecurring != nil {
		conditions = append(conditions, "is_recurring = ?")
		args = append(args, *filter.IsRecurring)
	}

	if len(filter.Tags) > 0 {
		// This would need more complex JSON querying depending on the database
		// For now, we'll do a simple LIKE search
		for _, tag := range filter.Tags {
			conditions = append(conditions, "tags LIKE ?")
			args = append(args, "%"+tag+"%")
		}
	}

	return strings.Join(conditions, " AND "), args
}

// deleteAllExpenseSplitsInTx removes all splits for an expense within a transaction
func (r *Repository) deleteAllExpenseSplitsInTx(ctx context.Context, tx *sql.Tx, expenseID string) error {
	query := `DELETE FROM expense_splits WHERE expense_id = ?`
	_, err := tx.ExecContext(ctx, query, expenseID)
	return err
}

// generateID generates a random ID
func (r *Repository) generateID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}