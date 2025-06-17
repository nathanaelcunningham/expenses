package models

import "time"

// Expense represents an expense in the system
type Expense struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	Amount        float64    `json:"amount"`
	CategoryID    *string    `json:"category_id,omitempty"`
	Category      *Category  `json:"category,omitempty"`
	Date          time.Time  `json:"date"`
	MemberID      string     `json:"member_id"`
	Member        *Member    `json:"member,omitempty"`
	IsRecurring   bool       `json:"is_recurring"`
	RecurrenceRule *string   `json:"recurrence_rule,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// Category represents an expense category
type Category struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Color       *string   `json:"color,omitempty"`
	Icon        *string   `json:"icon,omitempty"`
	Budget      *float64  `json:"budget,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateExpenseRequest represents the request to create a new expense
type CreateExpenseRequest struct {
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	Amount         float64 `json:"amount"`
	CategoryID     *string `json:"category_id,omitempty"`
	Date           time.Time `json:"date"`
	MemberID       string  `json:"member_id"`
	IsRecurring    bool    `json:"is_recurring"`
	RecurrenceRule *string `json:"recurrence_rule,omitempty"`
}

// UpdateExpenseRequest represents the request to update an expense
type UpdateExpenseRequest struct {
	Name           *string    `json:"name,omitempty"`
	Description    *string    `json:"description,omitempty"`
	Amount         *float64   `json:"amount,omitempty"`
	CategoryID     *string    `json:"category_id,omitempty"`
	Date           *time.Time `json:"date,omitempty"`
	IsRecurring    *bool      `json:"is_recurring,omitempty"`
	RecurrenceRule *string    `json:"recurrence_rule,omitempty"`
}

// ExpenseFilter represents filtering options for expense queries
type ExpenseFilter struct {
	MemberID   *string    `json:"member_id,omitempty"`
	CategoryID *string    `json:"category_id,omitempty"`
	DateFrom   *time.Time `json:"date_from,omitempty"`
	DateTo     *time.Time `json:"date_to,omitempty"`
	Limit      int        `json:"limit,omitempty"`
	Offset     int        `json:"offset,omitempty"`
}

// CreateCategoryRequest represents the request to create a new category
type CreateCategoryRequest struct {
	Name   string   `json:"name"`
	Color  *string  `json:"color,omitempty"`
	Icon   *string  `json:"icon,omitempty"`
	Budget *float64 `json:"budget,omitempty"`
}

// UpdateCategoryRequest represents the request to update a category
type UpdateCategoryRequest struct {
	Name     *string  `json:"name,omitempty"`
	Color    *string  `json:"color,omitempty"`
	Icon     *string  `json:"icon,omitempty"`
	Budget   *float64 `json:"budget,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
}