package repositories

import (
	"context"
	"database/sql"
	"expenses-backend/internal/database"
	"expenses-backend/internal/repositories/budget"
	"expenses-backend/internal/repositories/category"
	"expenses-backend/internal/repositories/expense"
	"expenses-backend/internal/repositories/family"

	"github.com/rs/zerolog"
)

// Factory provides repository instances based on database context
type Factory struct {
	dbManager *database.Manager
	logger    zerolog.Logger
}

// NewFactory creates a new repository factory
func NewFactory(dbManager *database.Manager, logger zerolog.Logger) *Factory {
	return &Factory{
		dbManager: dbManager,
		logger:    logger.With().Str("component", "repo-factory").Logger(),
	}
}

// NewFamilyStore creates a family repository that operates on the master database
func (f *Factory) NewFamilyStore() family.Store {
	return family.NewRepository(f.dbManager.GetMasterDatabase(), f.logger)
}

// NewExpenseStore creates an expense repository for a specific family database
func (f *Factory) NewExpenseStore(ctx context.Context, familyID string) (expense.Store, error) {
	familyDB, err := f.dbManager.GetFamilyDatabase(ctx, familyID)
	if err != nil {
		return nil, err
	}
	return expense.NewRepository(familyDB, f.logger), nil
}

// NewCategoryStore creates a category repository for a specific family database
func (f *Factory) NewCategoryStore(ctx context.Context, familyID string) (category.Store, error) {
	familyDB, err := f.dbManager.GetFamilyDatabase(ctx, familyID)
	if err != nil {
		return nil, err
	}
	return category.NewRepository(familyDB, f.logger), nil
}

// NewBudgetStore creates a budget repository for a specific family database
func (f *Factory) NewBudgetStore(ctx context.Context, familyID string) (budget.Store, error) {
	familyDB, err := f.dbManager.GetFamilyDatabase(ctx, familyID)
	if err != nil {
		return nil, err
	}
	return budget.NewRepository(familyDB, f.logger), nil
}

// NewStoresForFamily creates all stores for a specific family database
func (f *Factory) NewStoresForFamily(ctx context.Context, familyID string) (*FamilyStores, error) {
	familyDB, err := f.dbManager.GetFamilyDatabase(ctx, familyID)
	if err != nil {
		return nil, err
	}

	return &FamilyStores{
		ExpenseStore:  expense.NewRepository(familyDB, f.logger),
		CategoryStore: category.NewRepository(familyDB, f.logger),
		BudgetStore:   budget.NewRepository(familyDB, f.logger),
	}, nil
}

// NewStoresForFamilyDB creates all stores for a given family database connection
func (f *Factory) NewStoresForFamilyDB(familyDB *sql.DB) *FamilyStores {
	return &FamilyStores{
		ExpenseStore:  expense.NewRepository(familyDB, f.logger),
		CategoryStore: category.NewRepository(familyDB, f.logger),
		BudgetStore:   budget.NewRepository(familyDB, f.logger),
	}
}

// FamilyStores contains all repository stores for a family
type FamilyStores struct {
	ExpenseStore  expense.Store
	CategoryStore category.Store
	BudgetStore   budget.Store
}

// GetDatabaseManager returns the underlying database manager
func (f *Factory) GetDatabaseManager() *database.Manager {
	return f.dbManager
}