package repository

import (
	"context"
	"database/sql"
	"expenses-backend/internal/database"
	"expenses-backend/internal/models"
	"expenses-backend/internal/repository/family"
	"expenses-backend/internal/repository/interfaces"
	"expenses-backend/internal/repository/master"
	"fmt"
)

// RepositoryFactory provides access to all repositories with proper database connections
type RepositoryFactory struct {
	dbManager *database.Manager
}

// NewRepositoryFactory creates a new repository factory
func NewRepositoryFactory(dbManager *database.Manager) *RepositoryFactory {
	return &RepositoryFactory{
		dbManager: dbManager,
	}
}

// GetMasterRepositories returns repositories for the master database
func (f *RepositoryFactory) GetMasterRepositories() *interfaces.MasterRepositories {
	masterDB := f.dbManager.GetMasterDatabase()
	
	return &interfaces.MasterRepositories{
		Users:             master.NewUserRepository(masterDB),
		Sessions:          master.NewSessionRepository(masterDB),
		Families:          master.NewFamilyRepository(masterDB),
		FamilyMemberships: master.NewFamilyMembershipRepository(masterDB),
	}
}

// GetFamilyRepositories returns repositories for a specific family database
func (f *RepositoryFactory) GetFamilyRepositories(ctx context.Context, familyID string) (*interfaces.FamilyRepositories, error) {
	familyDB, err := f.dbManager.GetFamilyDatabase(ctx, familyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get family database: %w", err)
	}

	return &interfaces.FamilyRepositories{
		Expenses:      family.NewExpenseRepository(familyDB),
		Categories:    family.NewCategoryRepository(familyDB),
		FamilyMembers: family.NewFamilyMemberRepository(familyDB),
	}, nil
}

// GetMasterRepositoriesWithTx returns master repositories using a transaction
func (f *RepositoryFactory) GetMasterRepositoriesWithTx(tx *sql.Tx) *interfaces.MasterRepositories {
	return &interfaces.MasterRepositories{
		Users:             master.NewUserRepositoryWithTx(tx),
		Sessions:          master.NewSessionRepositoryWithTx(tx),
		Families:          master.NewFamilyRepositoryWithTx(tx),
		FamilyMemberships: master.NewFamilyMembershipRepositoryWithTx(tx),
	}
}

// GetFamilyRepositoriesWithTx returns family repositories using a transaction
func (f *RepositoryFactory) GetFamilyRepositoriesWithTx(tx *sql.Tx) *interfaces.FamilyRepositories {
	return &interfaces.FamilyRepositories{
		Expenses:      family.NewExpenseRepositoryWithTx(tx),
		Categories:    family.NewCategoryRepositoryWithTx(tx),
		FamilyMembers: family.NewFamilyMemberRepositoryWithTx(tx),
	}
}

// TransactionManager provides transaction management across repositories
type TransactionManager struct {
	factory *RepositoryFactory
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(factory *RepositoryFactory) *TransactionManager {
	return &TransactionManager{
		factory: factory,
	}
}

// WithMasterTx executes a function within a master database transaction
func (tm *TransactionManager) WithMasterTx(ctx context.Context, fn func(*interfaces.MasterRepositories) error) error {
	masterDB := tm.factory.dbManager.GetMasterDatabase()
	
	tx, err := masterDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	repos := tm.factory.GetMasterRepositoriesWithTx(tx)
	
	if err := fn(repos); err != nil {
		return err
	}

	return tx.Commit()
}

// WithFamilyTx executes a function within a family database transaction
func (tm *TransactionManager) WithFamilyTx(ctx context.Context, familyID string, fn func(*interfaces.FamilyRepositories) error) error {
	familyDB, err := tm.factory.dbManager.GetFamilyDatabase(ctx, familyID)
	if err != nil {
		return fmt.Errorf("failed to get family database: %w", err)
	}

	tx, err := familyDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	repos := tm.factory.GetFamilyRepositoriesWithTx(tx)
	
	if err := fn(repos); err != nil {
		return err
	}

	return tx.Commit()
}

// ExpenseStore interface for backward compatibility with existing expense service
type ExpenseStore interface {
	CreateExpense(ctx context.Context, req interface{}) (interface{}, error)
	GetExpenseByID(ctx context.Context, id string) (interface{}, error)
	UpdateExpense(ctx context.Context, id string, req interface{}) (interface{}, error)
	DeleteExpense(ctx context.Context, id string) error
	ListExpenses(ctx context.Context, filter interface{}) (interface{}, error)
	UserCanAccessExpense(ctx context.Context, expenseID, userID string) (bool, error)
}

// expenseStoreAdapter adapts the ExpenseRepository to the ExpenseStore interface
type expenseStoreAdapter struct {
	repo interfaces.ExpenseRepository
}

// NewExpenseStore creates a new expense store for backward compatibility
func (f *RepositoryFactory) NewExpenseStore(ctx context.Context, familyID string) (ExpenseStore, error) {
	familyRepos, err := f.GetFamilyRepositories(ctx, familyID)
	if err != nil {
		return nil, err
	}

	return &expenseStoreAdapter{
		repo: familyRepos.Expenses,
	}, nil
}

// Implement ExpenseStore interface
func (a *expenseStoreAdapter) CreateExpense(ctx context.Context, req interface{}) (interface{}, error) {
	// This is a bit of a hack for backward compatibility
	// In practice, you'd want to properly type these interfaces
	return a.repo.CreateExpense(ctx, req.(*models.CreateExpenseRequest))
}

func (a *expenseStoreAdapter) GetExpenseByID(ctx context.Context, id string) (interface{}, error) {
	return a.repo.GetExpenseByID(ctx, id)
}

func (a *expenseStoreAdapter) UpdateExpense(ctx context.Context, id string, req interface{}) (interface{}, error) {
	return a.repo.UpdateExpense(ctx, id, req.(*models.UpdateExpenseRequest))
}

func (a *expenseStoreAdapter) DeleteExpense(ctx context.Context, id string) error {
	return a.repo.DeleteExpense(ctx, id)
}

func (a *expenseStoreAdapter) ListExpenses(ctx context.Context, filter interface{}) (interface{}, error) {
	return a.repo.ListExpenses(ctx, filter.(*models.ExpenseFilter))
}

func (a *expenseStoreAdapter) UserCanAccessExpense(ctx context.Context, expenseID, userID string) (bool, error) {
	return a.repo.UserCanAccessExpense(ctx, expenseID, userID)
}