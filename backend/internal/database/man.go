package database

import (
	"context"
	"database/sql"
	"expenses-backend/internal/database/sql/familydb"
	"expenses-backend/internal/database/sql/masterdb"
	"expenses-backend/internal/database/turso"
	"expenses-backend/internal/logger"
	"fmt"
	"sync"
)

type DatabaseManager struct {
	masterDB      *sql.DB
	masterQueries *masterdb.Queries

	mu            sync.RWMutex
	familyDBs     map[string]*sql.DB
	familyQueries map[string]*familydb.Queries

	tursoClient *turso.Client
	log         logger.Logger
}

func New(masterDB *sql.DB, tursoClient *turso.Client, log logger.Logger) *DatabaseManager {
	return &DatabaseManager{
		masterDB:      masterDB,
		masterQueries: masterdb.New(masterDB),
		familyDBs:     make(map[string]*sql.DB),
		familyQueries: make(map[string]*familydb.Queries),
		tursoClient:   tursoClient,
		log:           log.With(logger.Str("component", "db-manager")),
	}
}
func (dm *DatabaseManager) GetMasterDB() *sql.DB {
	return dm.masterDB
}

func (dm *DatabaseManager) GetMasterQueries() *masterdb.Queries {
	return dm.masterQueries
}

func (dm *DatabaseManager) AddFamilyDB(familyID string, db *sql.DB) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.familyDBs[familyID] = db
	dm.familyQueries[familyID] = familydb.New(db)
}

func (dm *DatabaseManager) GetFamilyDB(familyID string) (*sql.DB, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	db, exists := dm.familyDBs[familyID]
	if !exists {
		err := fmt.Errorf("family database not found: %s", familyID)
		dm.log.Error("family database not found", err)
		return nil, err
	}
	return db, nil
}

func (dm *DatabaseManager) GetFamilyQueries(familyID string) (*familydb.Queries, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	queries, exists := dm.familyQueries[familyID]
	if !exists {
		err := fmt.Errorf("family queries not found: %s", familyID)
		dm.log.Error("family queries not found", err)
		return nil, err
	}
	return queries, nil
}

func (dm *DatabaseManager) WithMasterTx(ctx context.Context, fn func(*masterdb.Queries) error) error {
	tx, err := dm.masterDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := dm.masterQueries.WithTx(tx)
	if err := fn(qtx); err != nil {
		return err
	}

	return tx.Commit()
}

func (dm *DatabaseManager) WithFamilyTx(ctx context.Context, familyID string, fn func(*familydb.Queries) error) error {
	db, err := dm.GetFamilyDB(familyID)
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queries := familydb.New(db)
	qtx := queries.WithTx(tx)

	if err := fn(qtx); err != nil {
		return err
	}

	return tx.Commit()
}

func (dm *DatabaseManager) Close() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	var errors []error

	// Close master database
	if err := dm.masterDB.Close(); err != nil {
		errors = append(errors, fmt.Errorf("master db: %w", err))
	}

	// Close family databases
	for familyID, db := range dm.familyDBs {
		if err := db.Close(); err != nil {
			errors = append(errors, fmt.Errorf("family db %s: %w", familyID, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("close errors: %v", errors)
	}

	return nil
}

// ProvisionFamilyDatabase creates a new database for a family
func (dm *DatabaseManager) ProvisionFamilyDatabase(ctx context.Context, familyID, familyName string) (*turso.DatabaseInfo, error) {
	return nil, nil
}

// DeleteFamilyDatabase removes a family database
func (dm *DatabaseManager) DeleteFamilyDatabase(ctx context.Context, familyID string) error {
	return nil
}
