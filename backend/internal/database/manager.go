package database

//
// import (
// 	"context"
// 	"database/sql"
// 	"errors"
// 	"expenses-backend/internal/database/migrations"
// 	"expenses-backend/internal/database/turso"
// 	"fmt"
// 	"sync"
// 	"time"
//
// 	"expenses-backend/internal/logger"
// 	"maps"
// )
//
// // SeedDatabase represents a seed database configuration
// type SeedDatabase struct {
// 	Type string `json:"type"` // "database"
// 	Name string `json:"name"` // e.g., "family-seed"
// }
//
// // Manager handles database operations across master and family databases
// type Manager struct {
// 	tursoClient        *turso.Client
// 	masterDB           *sql.DB
// 	logger             logger.Logger
// 	mu                 sync.RWMutex
// 	familyDatabaseSeed *SeedDatabase // Turso database seed/template for family databases
// }
//
// // Config holds database manager configuration
// type Config struct {
// 	MasterDatabaseURL        string        `json:"master_database_url"`
// 	FamilySeedDatabaseUrl string               `json:"family_seed_database_url"`
// 	TursoConfig              turso.Config  `json:"turso_config"`
// 	FamilyDatabaseSeed       *SeedDatabase `json:"family_database_seed"` // Turso database seed/template for family databases
// }
//
// // FamilyDatabase represents a family-specific database
// type FamilyDatabase struct {
// 	ID            string    `json:"id"`
// 	Name          string    `json:"name"`
// 	URL           string    `json:"url"`
// 	Created       time.Time `json:"created"`
// 	SchemaVersion int       `json:"schema_version"`
// }
//
// // NewManager creates a new database manager
// func NewManager(ctx context.Context, config Config, log logger.Logger) (*Manager, error) {
// 	// Initialize Turso client
// 	tursoClient := turso.NewClient(config.TursoConfig, log)
//
// 	// Connect to master database
// 	masterDB, err := tursoClient.Connect(ctx, config.MasterDatabaseURL)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to connect to master database: %w", err)
// 	}
//
// 	manager := &Manager{
// 		tursoClient:        tursoClient,
// 		masterDB:           masterDB,
// 		logger:             log.With(logger.Str("component", "db-manager")),
// 		familyDatabaseSeed: config.FamilyDatabaseSeed,
// 	}
//
// 	// Log seed database configuration
// 	if config.FamilyDatabaseSeed != nil {
// 		manager.logger.Info("Family database seed configured",
// 			logger.Str("seed_type", config.FamilyDatabaseSeed.Type),
// 			logger.Str("seed_name", config.FamilyDatabaseSeed.Name))
// 	} else {
// 		manager.logger.Warn("No family database seed configured - family databases will be created empty", errors.New("no family database seed configured"))
// 	}
//
// 	return manager, nil
// }
//
// // ProvisionFamilyDatabase creates a new database for a family
// func (m *Manager) ProvisionFamilyDatabase(ctx context.Context, familyID, familyName string) (*FamilyDatabase, error) {
// 	// Validate seed database configuration before provisioning
// 	if err := m.ValidateSeedDatabase(ctx); err != nil {
// 		return nil, fmt.Errorf("seed database validation failed: %w", err)
// 	}
//
// 	m.logger.Info("Provisioning family database",
// 		logger.Str("family_id", familyID),
// 		logger.Str("family_name", familyName),
// 		logger.Str("seed", m.familyDatabaseSeed.Name))
//
// 	// Generate unique database name
// 	dbName := fmt.Sprintf("family-%s", familyID)
//
// 	// Create database via Turso API with seed if configured
// 	dbInfo, err := m.tursoClient.CreateDatabase(ctx, dbName, "ord", m.familyDatabaseSeed.Name)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create database: %w", err)
// 	}
//
// 	familyDB := &FamilyDatabase{
// 		ID:            familyID,
// 		Name:          familyName,
// 		URL:           dbInfo.URL,
// 		Created:       time.Now(),
// 		SchemaVersion: 0, // Will be updated after migrations run
// 	}
//
// 	// Update master database with family database info
// 	if err := m.recordFamilyDatabase(ctx, familyDB); err != nil {
// 		return nil, fmt.Errorf("failed to record family database: %w", err)
// 	}
//
// 	m.logger.Info("Family database provisioned successfully", logger.Str("family_id", familyID), logger.Str("database_url", dbInfo.URL))
//
// 	return familyDB, nil
// }
//
// // DeleteFamilyDatabase removes a family database
// func (m *Manager) DeleteFamilyDatabase(ctx context.Context, familyID string) error {
// 	m.logger.Info("Deleting family database", logger.Str("family_id", familyID))
//
// 	// Get the database name for deletion
// 	dbName := fmt.Sprintf("family-%s", familyID)
//
// 	// Delete from Turso
// 	if err := m.tursoClient.DeleteDatabase(ctx, dbName); err != nil {
// 		return fmt.Errorf("failed to delete Turso database: %w", err)
// 	}
//
// 	// Remove from master database (this will cascade to family_memberships due to FK)
// 	query := `DELETE FROM families WHERE id = ?`
// 	if _, err := m.masterDB.ExecContext(ctx, query, familyID); err != nil {
// 		m.logger.Error("Failed to remove family from master database after Turso deletion", err, logger.Str("family_id", familyID))
// 		return fmt.Errorf("failed to remove family from master database: %w", err)
// 	}
//
// 	m.logger.Info("Family database deleted successfully", logger.Str("family_id", familyID))
// 	return nil
// }
//
// // GetFamilyDatabase retrieves a connection to a family's database
// func (m *Manager) GetFamilyDatabase(ctx context.Context, familyID string) (*sql.DB, error) {
// 	// Get family database URL from master database
// 	var dbURL string
// 	query := `SELECT database_url FROM families WHERE id = ?`
//
// 	err := m.masterDB.QueryRowContext(ctx, query, familyID).Scan(&dbURL)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, fmt.Errorf("family database not found: %s", familyID)
// 		}
// 		return nil, fmt.Errorf("failed to get family database URL: %w", err)
// 	}
//
// 	// Get connection from Turso client
// 	return m.tursoClient.GetConnection(ctx, dbURL)
// }
//
// // GetMasterDatabase returns the master database connection
// func (m *Manager) GetMasterDatabase() *sql.DB {
// 	return m.masterDB
// }
//
// // RunFamilySeedMigrations runs migrations on the family seed database
// // Returns true if migrations were applied, false if already up to date
// func (m *Manager) RunFamilySeedMigrations(ctx context.Context, migrationManager *migrations.MigrationManager) (bool, error) {
// 	if m.familyDatabaseSeed == nil {
// 		m.logger.Info("No family database seed configured, skipping seed migrations")
// 		return false, nil
// 	}
//
// 	// Get connection to the seed database
// 	seedDB, err := m.tursoClient.GetConnection(ctx, fmt.Sprintf("libsql://%s.turso.io", m.familyDatabaseSeed.Name))
// 	if err != nil {
// 		return false, fmt.Errorf("failed to connect to seed database: %w", err)
// 	}
// 	defer seedDB.Close()
//
// 	// Check if there are pending migrations
// 	pendingMigrations, err := migrationManager.GetPendingMigrations(ctx, seedDB, migrations.FamilyMigration)
// 	if err != nil {
// 		return false, fmt.Errorf("failed to check pending migrations: %w", err)
// 	}
//
// 	if len(pendingMigrations) == 0 {
// 		m.logger.Info("Family seed database is up to date")
// 		return false, nil
// 	}
//
// 	// Run migrations on seed database
// 	if err := migrationManager.RunMigrations(ctx, seedDB, migrations.FamilyMigration); err != nil {
// 		return false, fmt.Errorf("failed to run migrations on seed database: %w", err)
// 	}
//
// 	m.logger.Info("Family seed database migrations completed", logger.Int("migrations_applied", len(pendingMigrations)))
// 	return true, nil
// }
//
// // RunFamilyDatabaseMigrations runs migrations on all family databases
// func (m *Manager) RunFamilyDatabaseMigrations(ctx context.Context, migrationManager *migrations.MigrationManager) error {
// 	// Get all family databases
// 	families, err := m.getAllFamilies(ctx)
// 	if err != nil {
// 		return fmt.Errorf("failed to get families: %w", err)
// 	}
//
// 	if len(families) == 0 {
// 		m.logger.Info("No family databases found to migrate")
// 		return nil
// 	}
//
// 	m.logger.Info("Running migrations on family databases", logger.Int("family_count", len(families)))
//
// 	// Run migrations on each family database
// 	for _, family := range families {
// 		familyDB, err := m.tursoClient.GetConnection(ctx, family.URL)
// 		if err != nil {
// 			m.logger.Error("Failed to connect to family database", err, logger.Str("family_id", family.ID))
// 			continue
// 		}
//
// 		if err := migrationManager.RunMigrations(ctx, familyDB, migrations.FamilyMigration); err != nil {
// 			m.logger.Error("Failed to run migrations for family", err, logger.Str("family_id", family.ID))
// 			continue
// 		}
//
// 		// Update schema version in master database
// 		if err := m.updateFamilySchemaVersion(ctx, family.ID); err != nil {
// 			m.logger.Warn("Failed to update family schema version", err, logger.Str("family_id", family.ID))
// 		}
// 	}
//
// 	m.logger.Info("Family database migrations completed")
// 	return nil
// }
//
// // HealthCheck performs health checks on all databases
// func (m *Manager) HealthCheck(ctx context.Context) map[string]error {
// 	results := make(map[string]error)
//
// 	// Check master database
// 	if err := m.masterDB.PingContext(ctx); err != nil {
// 		results["master"] = err
// 	}
//
// 	// Check family databases
// 	familyResults := m.tursoClient.HealthCheck(ctx)
// 	maps.Copy(results, familyResults)
//
// 	return results
// }
//
// // Close closes all database connections
// func (m *Manager) Close() error {
// 	m.logger.Info("Closing database manager")
//
// 	var errors []error
//
// 	// Close master database
// 	if err := m.masterDB.Close(); err != nil {
// 		errors = append(errors, fmt.Errorf("failed to close master database: %w", err))
// 	}
//
// 	// Close all Turso connections
// 	if err := m.tursoClient.CloseAll(); err != nil {
// 		errors = append(errors, err)
// 	}
//
// 	if len(errors) > 0 {
// 		return fmt.Errorf("errors closing databases: %v", errors)
// 	}
//
// 	return nil
// }
//
// // GetTursoClient returns the underlying Turso client
// func (m *Manager) GetTursoClient() *turso.Client {
// 	return m.tursoClient
// }
//
// // updateFamilySchemaVersion updates the schema version for a family in master database
// func (m *Manager) updateFamilySchemaVersion(ctx context.Context, familyID string) error {
// 	// Get the current schema version from the family database
// 	familyDB, err := m.GetFamilyDatabase(ctx, familyID)
// 	if err != nil {
// 		return err
// 	}
//
// 	var currentVersion int
// 	err = familyDB.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&currentVersion)
// 	if err != nil {
// 		return err
// 	}
//
// 	// Update the master database record
// 	query := `UPDATE families SET schema_version = ? WHERE id = ?`
// 	_, err = m.masterDB.ExecContext(ctx, query, currentVersion, familyID)
// 	return err
// }
//
// // recordFamilyDatabase records family database info in master database
// func (m *Manager) recordFamilyDatabase(ctx context.Context, familyDB *FamilyDatabase) error {
// 	query := `UPDATE families SET database_url = ?, schema_version = ? WHERE id = ?`
// 	_, err := m.masterDB.ExecContext(ctx, query, familyDB.URL, familyDB.SchemaVersion, familyDB.ID)
// 	return err
// }
//
// // getAllFamilies retrieves all families from master database
// func (m *Manager) getAllFamilies(ctx context.Context) ([]FamilyDatabase, error) {
// 	query := `SELECT id, name, database_url, schema_version, created_at FROM families`
// 	rows, err := m.masterDB.QueryContext(ctx, query)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
//
// 	var families []FamilyDatabase
// 	for rows.Next() {
// 		var family FamilyDatabase
// 		err := rows.Scan(&family.ID, &family.Name, &family.URL, &family.SchemaVersion, &family.Created)
// 		if err != nil {
// 			return nil, err
// 		}
// 		families = append(families, family)
// 	}
//
// 	return families, rows.Err()
// }
// // ValidateSeedDatabase checks if the seed database exists and is accessible
// func (m *Manager) ValidateSeedDatabase(ctx context.Context) error {
// 	if m.familyDatabaseSeed == nil {
// 		return fmt.Errorf("family database seed configuration is not set")
// 	}
//
// 	if m.familyDatabaseSeed.Type != "database" {
// 		return fmt.Errorf("invalid seed database type: %s (expected 'database')", m.familyDatabaseSeed.Type)
// 	}
//
// 	if m.familyDatabaseSeed.Name == "" {
// 		return fmt.Errorf("seed database name cannot be empty")
// 	}
//
// 	m.logger.Debug("Validating family seed database", logger.Str("seed_name", m.familyDatabaseSeed.Name))
//
// 	// Try to connect to the seed database to verify it exists
// 	seedDB, err := m.tursoClient.GetConnection(ctx, fmt.Sprintf("libsql://%s.turso.io", m.familyDatabaseSeed.Name))
// 	if err != nil {
// 		return fmt.Errorf("failed to connect to seed database '%s': %w", m.familyDatabaseSeed.Name, err)
// 	}
// 	defer seedDB.Close()
//
// 	// Verify the seed database has the expected schema by checking for migrations table
// 	var migrationCount int
// 	err = seedDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations").Scan(&migrationCount)
// 	if err != nil {
// 		return fmt.Errorf("seed database '%s' appears to be invalid (no schema_migrations table): %w", m.familyDatabaseSeed.Name, err)
// 	}
//
// 	if migrationCount == 0 {
// 		return fmt.Errorf("seed database '%s' has no migrations applied", m.familyDatabaseSeed.Name)
// 	}
//
// 	m.logger.Debug("Family seed database validation successful",
// 		logger.Str("seed_name", m.familyDatabaseSeed.Name),
// 		logger.Int("migration_count", migrationCount))
//
// 	return nil
// }
