package interfaces

import (
	"context"
	"database/sql"
	"expenses-backend/internal/models"
)

// UserRepository defines the interface for user-related database operations
type UserRepository interface {
	CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, id string, req *models.UpdateUserRequest) (*models.User, error)
	DeleteUser(ctx context.Context, id string) error
}

// SessionRepository defines the interface for session-related database operations
type SessionRepository interface {
	CreateSession(ctx context.Context, session *models.UserSession) (*models.UserSession, error)
	GetSession(ctx context.Context, id string) (*models.UserSession, error)
	GetUserActiveSessions(ctx context.Context, userID string, limit int) ([]*models.UserSession, error)
	UpdateSessionActivity(ctx context.Context, id string) error
	DeleteSession(ctx context.Context, id string) error
	DeleteExpiredSessions(ctx context.Context) (int64, error)
}

// FamilyRepository defines the interface for family-related database operations
type FamilyRepository interface {
	CreateFamily(ctx context.Context, req *models.CreateFamilyRequest) (*models.Family, error)
	GetFamilyByID(ctx context.Context, id string) (*models.Family, error)
	GetFamilyByInviteCode(ctx context.Context, inviteCode string) (*models.Family, error)
	UpdateFamily(ctx context.Context, id string, req *models.UpdateFamilyRequest) (*models.Family, error)
	DeleteFamily(ctx context.Context, id string) error
}

// FamilyMembershipRepository defines the interface for family membership operations
type FamilyMembershipRepository interface {
	CreateMembership(ctx context.Context, familyID, userID, role string) error
	GetMembership(ctx context.Context, familyID, userID string) (*models.Member, error)
	ListFamilyMemberships(ctx context.Context, familyID string) ([]*models.Member, error)
	ListUserMemberships(ctx context.Context, userID string) ([]*models.Family, error)
	UpdateMembershipRole(ctx context.Context, familyID, userID, role string) error
	DeleteMembership(ctx context.Context, familyID, userID string) error
}

// ExpenseRepository defines the interface for expense-related database operations
type ExpenseRepository interface {
	CreateExpense(ctx context.Context, req *models.CreateExpenseRequest) (*models.Expense, error)
	GetExpenseByID(ctx context.Context, id string) (*models.Expense, error)
	ListExpenses(ctx context.Context, filter *models.ExpenseFilter) ([]*models.Expense, error)
	ListExpensesByCategory(ctx context.Context, categoryID string) ([]*models.Expense, error)
	GetExpensesByDateRange(ctx context.Context, startDay, endDay int) ([]*models.Expense, error)
	UpdateExpense(ctx context.Context, id string, req *models.UpdateExpenseRequest) (*models.Expense, error)
	DeleteExpense(ctx context.Context, id string) error
	CountExpenses(ctx context.Context) (int64, error)
	UserCanAccessExpense(ctx context.Context, expenseID, userID string) (bool, error)
}

// CategoryRepository defines the interface for category-related database operations
type CategoryRepository interface {
	CreateCategory(ctx context.Context, req *models.CreateCategoryRequest) (*models.Category, error)
	GetCategoryByID(ctx context.Context, id string) (*models.Category, error)
	ListCategories(ctx context.Context) ([]*models.Category, error)
	UpdateCategory(ctx context.Context, id string, req *models.UpdateCategoryRequest) (*models.Category, error)
	DeleteCategory(ctx context.Context, id string) error
}

// FamilyMemberRepository defines the interface for family member operations in family database
type FamilyMemberRepository interface {
	CreateFamilyMember(ctx context.Context, member *models.Member) (*models.Member, error)
	GetFamilyMemberByID(ctx context.Context, id string) (*models.Member, error)
	GetFamilyMemberByEmail(ctx context.Context, email string) (*models.Member, error)
	ListFamilyMembers(ctx context.Context) ([]*models.Member, error)
	ListAllFamilyMembers(ctx context.Context) ([]*models.Member, error)
	UpdateFamilyMember(ctx context.Context, id string, member *models.Member) (*models.Member, error)
	DeactivateFamilyMember(ctx context.Context, id string) error
	DeleteFamilyMember(ctx context.Context, id string) error
}

// Transactional interface for repositories that support transactions
type Transactional interface {
	WithTx(tx *sql.Tx) interface{}
}

// MasterRepositories groups all master database repositories
type MasterRepositories struct {
	Users             UserRepository
	Sessions          SessionRepository
	Families          FamilyRepository
	FamilyMemberships FamilyMembershipRepository
}

// FamilyRepositories groups all family database repositories
type FamilyRepositories struct {
	Expenses      ExpenseRepository
	Categories    CategoryRepository
	FamilyMembers FamilyMemberRepository
}