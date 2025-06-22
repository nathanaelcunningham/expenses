package database

// import (
// 	"context"
// 	"database/sql"
// 	"expenses-backend/internal/database/sql/familydb"
// 	"expenses-backend/internal/database/sql/masterdb"
// 	"fmt"
// )
//
// // Factory provides access to SQLC queries with proper database connections
// type Factory struct {
// 	manager *Manager
// }
//
// // NewFactory creates a new database factory
// func NewFactory(manager *Manager) *Factory {
// 	return &Factory{
// 		manager: manager,
// 	}
// }
//
// // GetMasterQueries returns SQLC queries for the master database
// func (f *Factory) GetMasterQueries() *masterdb.Queries {
// 	masterDB := f.manager.GetMasterDatabase()
// 	return masterdb.New(masterDB)
// }
//
// // GetFamilyQueries returns SQLC queries for a specific family database
// func (f *Factory) GetFamilyQueries(ctx context.Context, familyID string) (*familydb.Queries, error) {
// 	familyDB, err := f.manager.GetFamilyDatabase(ctx, familyID)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get family database: %w", err)
// 	}
//
// 	return familydb.New(familyDB), nil
// }
//
// // GetMasterQueriesWithTx returns master queries using a transaction
// func (f *Factory) GetMasterQueriesWithTx(tx *sql.Tx) *masterdb.Queries {
// 	return masterdb.New(tx)
// }
//
// // GetFamilyQueriesWithTx returns family queries using a transaction
// func (f *Factory) GetFamilyQueriesWithTx(tx *sql.Tx) *familydb.Queries {
// 	return familydb.New(tx)
// }
//
// // TransactionManager provides transaction management across databases
// type TransactionManager struct {
// 	factory *Factory
// }
//
// // NewTransactionManager creates a new transaction manager
// func NewTransactionManager(factory *Factory) *TransactionManager {
// 	return &TransactionManager{
// 		factory: factory,
// 	}
// }
//
// // WithMasterTx executes a function within a master database transaction
// func (tm *TransactionManager) WithMasterTx(ctx context.Context, fn func(*masterdb.Queries) error) error {
// 	masterDB := tm.factory.manager.GetMasterDatabase()
//
// 	tx, err := masterDB.BeginTx(ctx, nil)
// 	if err != nil {
// 		return fmt.Errorf("failed to begin transaction: %w", err)
// 	}
// 	defer tx.Rollback()
//
// 	queries := tm.factory.GetMasterQueriesWithTx(tx)
//
// 	if err := fn(queries); err != nil {
// 		return err
// 	}
//
// 	return tx.Commit()
// }
//
// // WithFamilyTx executes a function within a family database transaction
// func (tm *TransactionManager) WithFamilyTx(ctx context.Context, familyID string, fn func(*familydb.Queries) error) error {
// 	familyDB, err := tm.factory.manager.GetFamilyDatabase(ctx, familyID)
// 	if err != nil {
// 		return fmt.Errorf("failed to get family database: %w", err)
// 	}
//
// 	tx, err := familyDB.BeginTx(ctx, nil)
// 	if err != nil {
// 		return fmt.Errorf("failed to begin transaction: %w", err)
// 	}
// 	defer tx.Rollback()
//
// 	queries := tm.factory.GetFamilyQueriesWithTx(tx)
//
// 	if err := fn(queries); err != nil {
// 		return err
// 	}
//
// 	return tx.Commit()
// }

