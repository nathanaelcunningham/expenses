package database

import (
	"context"
	"crypto/rand"
	"database/sql"
	"expenses-backend/internal/database/migrations"
	"expenses-backend/internal/database/sql/familydb"
	"expenses-backend/internal/database/sql/masterdb"
	"expenses-backend/internal/database/turso"
	"expenses-backend/internal/logger"
	"fmt"
	"strings"
	"sync"
	"time"
)

type DatabaseManager struct {
	masterDB      *sql.DB
	masterQueries *masterdb.Queries

	mu            sync.RWMutex
	familyDBs     map[int64]*sql.DB
	familyQueries map[int64]*familydb.Queries

	tursoClient  *turso.Client
	migrationMgr *migrations.MigrationManager
	log          logger.Logger
}

func New(masterDB *sql.DB, tursoClient *turso.Client, migrationMgr *migrations.MigrationManager, log logger.Logger) *DatabaseManager {
	return &DatabaseManager{
		masterDB:      masterDB,
		masterQueries: masterdb.New(masterDB),
		familyDBs:     make(map[int64]*sql.DB),
		familyQueries: make(map[int64]*familydb.Queries),
		tursoClient:   tursoClient,
		migrationMgr:  migrationMgr,
		log:           log.With(logger.Str("component", "db-manager")),
	}
}
func (dm *DatabaseManager) GetMasterDB() *sql.DB {
	return dm.masterDB
}

func (dm *DatabaseManager) GetMasterQueries() *masterdb.Queries {
	return dm.masterQueries
}

func (dm *DatabaseManager) AddFamilyDB(familyID int, db *sql.DB) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.familyDBs[int64(familyID)] = db
	dm.familyQueries[int64(familyID)] = familydb.New(db)
}

func (dm *DatabaseManager) GetFamilyDB(familyID int) (*sql.DB, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	db, exists := dm.familyDBs[int64(familyID)]
	if !exists {
		err := fmt.Errorf("family database not found: %d", familyID)
		dm.log.Error("family database not found", err)
		return nil, err
	}
	return db, nil
}

func (dm *DatabaseManager) GetFamilyQueries(familyID int) (*familydb.Queries, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	queries, exists := dm.familyQueries[int64(familyID)]
	if !exists {
		err := fmt.Errorf("family queries not found: %d", familyID)
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

func (dm *DatabaseManager) WithFamilyTx(ctx context.Context, familyID int, fn func(*familydb.Queries) error) error {
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
			errors = append(errors, fmt.Errorf("family db %d: %w", familyID, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("close errors: %v", errors)
	}

	return nil
}

// ProvisionFamilyDatabase creates a new database for a family
func (dm *DatabaseManager) ProvisionFamilyDatabase(ctx context.Context, familyName string) (*sql.DB, error) {
	dm.log.Info("Provisioning family database",
		logger.Str("family_name", familyName),
	)

	// Generate unique database name using timestamp and random suffix
	timestamp := time.Now().Format("20060102-150405")
	randomBytes := make([]byte, 4)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random suffix: %w", err)
	}
	randomSuffix := fmt.Sprintf("%x", randomBytes)
	dbName := fmt.Sprintf("family-%s-%s", timestamp, randomSuffix)

	// Create Turso database using family-seed as template
	dbInfo, err := dm.tursoClient.CreateDatabase(ctx, dbName, "family-seed")
	if err != nil {
		return nil, fmt.Errorf("failed to create Turso database: %w", err)
	}

	// Connect to the new database
	db, err := dm.tursoClient.Connect(ctx, dbInfo.URL)
	if err != nil {
		// Cleanup: delete the database if we can't connect
		if deleteErr := dm.tursoClient.DeleteDatabase(ctx, dbName); deleteErr != nil {
			dm.log.Warn("Failed to cleanup database after connection failure", deleteErr,
				logger.Str("db_name", dbName))
		}
		return nil, fmt.Errorf("failed to connect to new family database: %w", err)
	}

	dm.log.Info("Family database provisioned successfully",
		logger.Str("db_name", dbName),
		logger.Str("db_url", dbInfo.URL),
	)

	return db, nil
}

// DeleteFamilyDatabase removes a family database
func (dm *DatabaseManager) DeleteFamilyDatabase(ctx context.Context, familyID int) error {
	dm.log.Info("Deleting family database", logger.Int("family_id", familyID))

	// Get family info from master database to get the database name
	family, err := dm.masterQueries.GetFamilyByID(ctx, int64(familyID))
	if err != nil {
		if err == sql.ErrNoRows {
			dm.log.Warn("Family not found in master database", err, logger.Int("family_id", familyID))
			return fmt.Errorf("family not found: %d", familyID)
		}
		return fmt.Errorf("failed to get family info: %w", err)
	}

	// Extract database name from URL (e.g., "libsql://family-123.turso.io" -> "family-123")
	dbName := dm.extractDatabaseNameFromURL(family.DatabaseUrl)
	if dbName == "" {
		return fmt.Errorf("invalid database URL format: %s", family.DatabaseUrl)
	}

	// Remove from internal maps and close connection
	dm.mu.Lock()
	if db, exists := dm.familyDBs[int64(familyID)]; exists {
		db.Close()
		delete(dm.familyDBs, int64(familyID))
		delete(dm.familyQueries, int64(familyID))
	}
	dm.mu.Unlock()

	fID := int64(familyID)
	// Delete all family memberships
	memberships, err := dm.masterQueries.ListFamilyMemberships(ctx, &fID)
	if err != nil {
		dm.log.Warn("Failed to get family memberships for cleanup", err, logger.Int64("family_id", fID))
	} else {
		for _, membership := range memberships {
			if membership.UserID != nil {
				deleteParams := masterdb.DeleteFamilyMembershipParams{
					FamilyID: &fID,
					UserID:   membership.UserID,
				}
				if err := dm.masterQueries.DeleteFamilyMembership(ctx, deleteParams); err != nil {
					dm.log.Warn("Failed to delete family membership", err,
						logger.Int("family_id", familyID),
						logger.Int64("user_id", *membership.UserID))
				}
			}
		}
	}

	// Delete family record from master database
	if err := dm.masterQueries.DeleteFamily(ctx, fID); err != nil {
		dm.log.Warn("Failed to delete family record from master database", err, logger.Int("family_id", familyID))
	}

	// Delete the Turso database
	if err := dm.tursoClient.DeleteDatabase(ctx, dbName); err != nil {
		return fmt.Errorf("failed to delete Turso database: %w", err)
	}

	dm.log.Info("Family database deleted successfully",
		logger.Int("family_id", familyID),
		logger.Str("db_name", dbName),
	)

	return nil
}

// GetFamilyDatabase gets or establishes a connection to a family database
func (dm *DatabaseManager) GetFamilyDatabase(ctx context.Context, familyID int) (*sql.DB, error) {
	fID := int64(familyID)
	// Check if already connected
	dm.mu.RLock()
	if db, exists := dm.familyDBs[fID]; exists {
		dm.mu.RUnlock()
		return db, nil
	}
	dm.mu.RUnlock()

	// Not connected, need to establish connection
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Double-check after acquiring write lock
	if db, exists := dm.familyDBs[fID]; exists {
		return db, nil
	}

	// Get family info from master database
	family, err := dm.masterQueries.GetFamilyByID(ctx, fID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("family not found: %d", fID)
		}
		return nil, fmt.Errorf("failed to get family info: %w", err)
	}

	// Connect to family database
	familyDB, err := dm.tursoClient.Connect(ctx, family.DatabaseUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to family database: %w", err)
	}

	// Run any pending migrations
	if err := dm.migrationMgr.RunMigrations(ctx, familyDB, migrations.FamilyMigration); err != nil {
		familyDB.Close()
		return nil, fmt.Errorf("failed to run migrations on family database: %w", err)
	}

	// Add to internal maps
	dm.familyDBs[fID] = familyDB
	dm.familyQueries[fID] = familydb.New(familyDB)

	dm.log.Info("Connected to family database",
		logger.Int64("family_id", fID),
		logger.Str("db_url", family.DatabaseUrl),
	)

	return familyDB, nil
}

// LoadExistingFamilyDatabases discovers and connects to existing family databases
func (dm *DatabaseManager) LoadExistingFamilyDatabases(ctx context.Context) error {
	dm.log.Info("Loading existing family databases")

	// Get all families from master database
	// Note: We need a query to list all families - let's use a simple SQL query for now
	query := "SELECT id, database_url FROM families"
	rows, err := dm.masterDB.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query families: %w", err)
	}
	defer rows.Close()

	var families []struct {
		ID          int64
		DatabaseURL string
	}

	for rows.Next() {
		var family struct {
			ID          int64
			DatabaseURL string
		}
		if err := rows.Scan(&family.ID, &family.DatabaseURL); err != nil {
			dm.log.Warn("Failed to scan family row", err)
			continue
		}
		families = append(families, family)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating family rows: %w", err)
	}

	// Attempt to connect to each family database
	connectedCount := 0
	for _, family := range families {
		if err := dm.connectToFamilyDatabase(ctx, int(family.ID), family.DatabaseURL); err != nil {
			dm.log.Warn("Failed to connect to family database", err,
				logger.Int64("family_id", family.ID),
				logger.Str("database_url", family.DatabaseURL),
			)
		} else {
			connectedCount++
		}
	}

	dm.log.Info("Finished loading family databases",
		logger.Int("total_families", len(families)),
		logger.Int("connected", connectedCount),
	)

	return nil
}

// connectToFamilyDatabase helper function to connect to a specific family database
func (dm *DatabaseManager) connectToFamilyDatabase(ctx context.Context, familyID int, databaseURL string) error {
	// Connect to family database
	familyDB, err := dm.tursoClient.Connect(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Run any pending migrations
	if err := dm.migrationMgr.RunMigrations(ctx, familyDB, migrations.FamilyMigration); err != nil {
		familyDB.Close()
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Add to internal maps (using write lock)
	dm.mu.Lock()
	dm.familyDBs[int64(familyID)] = familyDB
	dm.familyQueries[int64(familyID)] = familydb.New(familyDB)
	dm.mu.Unlock()

	return nil
}

// extractDatabaseNameFromURL extracts database name from Turso URL
// Example: "libsql://family-123.turso.io" -> "family-123"
func (dm *DatabaseManager) extractDatabaseNameFromURL(url string) string {
	// Remove protocol prefix
	url = strings.TrimPrefix(url, "libsql://")
	url = strings.TrimPrefix(url, "https://")

	// Extract hostname part before .turso.io
	if idx := strings.Index(url, ".turso.io"); idx > 0 {
		return url[:idx]
	}

	// If not standard format, return empty string
	return ""
}
