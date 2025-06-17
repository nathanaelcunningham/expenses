package models

import (
	"time"
)

// Expense represents an expense record from the family database
type Expense struct {
	ID            string    `json:"id"`
	CategoryID    *string   `json:"category_id,omitempty"`
	Amount        float64   `json:"amount"`
	Name          string    `json:"name"`
	DayOfMonthDue int       `json:"day_of_month_due"`
	IsAutopay     bool      `json:"is_autopay"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CreateExpenseRequest represents the data needed to create an expense
type CreateExpenseRequest struct {
	CategoryID    *string `json:"category_id,omitempty"`
	Amount        float64 `json:"amount"`
	Name          string  `json:"name"`
	DayOfMonthDue int     `json:"day_of_month_due"`
	IsAutopay     bool    `json:"is_autopay"`
}

// UpdateExpenseRequest represents the data for updating an expense
type UpdateExpenseRequest struct {
	CategoryID    *string  `json:"category_id,omitempty"`
	Amount        *float64 `json:"amount,omitempty"`
	Name          string   `json:"name"`
	DayOfMonthDue int      `json:"day_of_month_due"`
	IsAutopay     bool     `json:"is_autopay"`
}

// ExpenseSummary represents expense summary statistics
type ExpenseSummary struct {
	TotalAmount   float64            `json:"total_amount"`
	ExpenseCount  int                `json:"expense_count"`
	AverageAmount float64            `json:"average_amount"`
	ByCategory    map[string]float64 `json:"by_category"`
	ByMember      map[string]float64 `json:"by_member"`
	Currency      string             `json:"currency"`
}

