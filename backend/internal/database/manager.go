package database

import (
	"context"
	"database/sql"
	"expenses-backend/internal/database/migrations"
	"expenses-backend/internal/database/turso"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// Manager handles database operations across master and family databases
type Manager struct {
	tursoClient *turso.Client
	masterDB    *sql.DB
	logger      zerolog.Logger
	mu          sync.RWMutex
}

// Config holds database manager configuration
type Config struct {
	MasterDatabaseURL string       `json:"master_database_url"`
	TursoConfig       turso.Config `json:"turso_config"`
}

// FamilyDatabase represents a family-specific database
type FamilyDatabase struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	URL           string    `json:"url"`
	Created       time.Time `json:"created"`
	SchemaVersion int       `json:"schema_version"`
}

// NewManager creates a new database manager
func NewManager(ctx context.Context, config Config, logger zerolog.Logger) (*Manager, error) {
	// Initialize Turso client
	tursoClient := turso.NewClient(config.TursoConfig, logger)

	// Connect to master database
	masterDB, err := tursoClient.Connect(ctx, config.MasterDatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to master database: %w", err)
	}

	manager := &Manager{
		tursoClient: tursoClient,
		masterDB:    masterDB,
		logger:      logger.With().Str("component", "db-manager").Logger(),
	}

	return manager, nil
}

// ProvisionFamilyDatabase creates a new database for a family
func (m *Manager) ProvisionFamilyDatabase(ctx context.Context, familyID, familyName string) (*FamilyDatabase, error) {
	m.logger.Info().
		Str("family_id", familyID).
		Str("family_name", familyName).
		Msg("Provisioning family database")

	// Generate unique database name
	dbName := fmt.Sprintf("family-%s", familyID)

	// Create database via Turso API
	dbInfo, err := m.tursoClient.CreateDatabase(ctx, dbName, "ord")
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	familyDB := &FamilyDatabase{
		ID:            familyID,
		Name:          familyName,
		URL:           dbInfo.URL,
		Created:       time.Now(),
		SchemaVersion: 0, // Will be updated after migrations run
	}

	// Update master database with family database info
	if err := m.recordFamilyDatabase(ctx, familyDB); err != nil {
		return nil, fmt.Errorf("failed to record family database: %w", err)
	}

	m.logger.Info().
		Str("family_id", familyID).
		Str("database_url", dbInfo.URL).
		Msg("Family database provisioned successfully")

	return familyDB, nil
}

// DeleteFamilyDatabase removes a family database
func (m *Manager) DeleteFamilyDatabase(ctx context.Context, familyID string) error {
	m.logger.Info().
		Str("family_id", familyID).
		Msg("Deleting family database")

	// Get the database name for deletion
	dbName := fmt.Sprintf("family-%s", familyID)

	// Delete from Turso
	if err := m.tursoClient.DeleteDatabase(ctx, dbName); err != nil {
		return fmt.Errorf("failed to delete Turso database: %w", err)
	}

	// Remove from master database (this will cascade to family_memberships due to FK)
	query := `DELETE FROM families WHERE id = ?`
	if _, err := m.masterDB.ExecContext(ctx, query, familyID); err != nil {
		m.logger.Error().
			Str("family_id", familyID).
			Err(err).
			Msg("Failed to remove family from master database after Turso deletion")
		return fmt.Errorf("failed to remove family from master database: %w", err)
	}

	m.logger.Info().
		Str("family_id", familyID).
		Msg("Family database deleted successfully")

	return nil
}

// GetFamilyDatabase retrieves a connection to a family's database
func (m *Manager) GetFamilyDatabase(ctx context.Context, familyID string) (*sql.DB, error) {
	// Get family database URL from master database
	var dbURL string
	query := `SELECT database_url FROM families WHERE id = ?`

	err := m.masterDB.QueryRowContext(ctx, query, familyID).Scan(&dbURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("family database not found: %s", familyID)
		}
		return nil, fmt.Errorf("failed to get family database URL: %w", err)
	}

	// Get connection from Turso client
	return m.tursoClient.GetConnection(ctx, dbURL)
}

// GetMasterDatabase returns the master database connection
func (m *Manager) GetMasterDatabase() *sql.DB {
	return m.masterDB
}

// RunMigrations runs pending migrations on all family databases
func (m *Manager) RunMigrations(ctx context.Context, migrationManager *migrations.MigrationManager) error {
	m.logger.Info().Msg("Running database migrations")

	// Run master database migrations first
	if err := migrationManager.RunMigrations(ctx, m.masterDB, migrations.MasterMigration); err != nil {
		return fmt.Errorf("failed to run master migrations: %w", err)
	}

	// Get all family databases
	families, err := m.getAllFamilies(ctx)
	if err != nil {
		return fmt.Errorf("failed to get families: %w", err)
	}

	// Run migrations on each family database
	for _, family := range families {
		familyDB, err := m.tursoClient.GetConnection(ctx, family.URL)
		if err != nil {
			m.logger.Error().
				Str("family_id", family.ID).
				Err(err).
				Msg("Failed to connect to family database")
			continue
		}

		if err := migrationManager.RunMigrations(ctx, familyDB, migrations.FamilyMigration); err != nil {
			m.logger.Error().
				Str("family_id", family.ID).
				Err(err).
				Msg("Failed to run migrations for family")
			continue
		}

		// Update schema version in master database
		if err := m.updateFamilySchemaVersion(ctx, family.ID); err != nil {
			m.logger.Warn().
				Str("family_id", family.ID).
				Err(err).
				Msg("Failed to update family schema version")
		}
	}

	m.logger.Info().Msg("Database migrations completed")
	return nil
}

// HealthCheck performs health checks on all databases
func (m *Manager) HealthCheck(ctx context.Context) map[string]error {
	results := make(map[string]error)

	// Check master database
	if err := m.masterDB.PingContext(ctx); err != nil {
		results["master"] = err
	}

	// Check family databases
	familyResults := m.tursoClient.HealthCheck(ctx)
	for url, err := range familyResults {
		results[url] = err
	}

	return results
}

// Close closes all database connections
func (m *Manager) Close() error {
	m.logger.Info().Msg("Closing database manager")

	var errors []error

	// Close master database
	if err := m.masterDB.Close(); err != nil {
		errors = append(errors, fmt.Errorf("failed to close master database: %w", err))
	}

	// Close all Turso connections
	if err := m.tursoClient.CloseAll(); err != nil {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing databases: %v", errors)
	}

	return nil
}

// GetTursoClient returns the underlying Turso client
func (m *Manager) GetTursoClient() *turso.Client {
	return m.tursoClient
}

// updateFamilySchemaVersion updates the schema version for a family in master database
func (m *Manager) updateFamilySchemaVersion(ctx context.Context, familyID string) error {
	// Get the current schema version from the family database
	familyDB, err := m.GetFamilyDatabase(ctx, familyID)
	if err != nil {
		return err
	}

	var currentVersion int
	err = familyDB.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&currentVersion)
	if err != nil {
		return err
	}

	// Update the master database record
	query := `UPDATE families SET schema_version = ? WHERE id = ?`
	_, err = m.masterDB.ExecContext(ctx, query, currentVersion, familyID)
	return err
}

// recordFamilyDatabase records family database info in master database
func (m *Manager) recordFamilyDatabase(ctx context.Context, familyDB *FamilyDatabase) error {
	query := `UPDATE families SET database_url = ?, schema_version = ? WHERE id = ?`
	_, err := m.masterDB.ExecContext(ctx, query, familyDB.URL, familyDB.SchemaVersion, familyDB.ID)
	return err
}

// getAllFamilies retrieves all families from master database
func (m *Manager) getAllFamilies(ctx context.Context) ([]FamilyDatabase, error) {
	query := `SELECT id, name, database_url, schema_version, created_at FROM families`
	rows, err := m.masterDB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var families []FamilyDatabase
	for rows.Next() {
		var family FamilyDatabase
		err := rows.Scan(&family.ID, &family.Name, &family.URL, &family.SchemaVersion, &family.Created)
		if err != nil {
			return nil, err
		}
		families = append(families, family)
	}

	return families, rows.Err()
}

// runFamilyMigrations runs migrations for a specific family database
func (m *Manager) runFamilyMigrations(ctx context.Context, family FamilyDatabase) error {
	db, err := m.tursoClient.GetConnection(ctx, family.URL)
	if err != nil {
		return err
	}

	// Check current schema version
	var currentVersion int
	err = db.QueryRowContext(ctx, "SELECT MAX(version) FROM schema_migrations").Scan(&currentVersion)
	if err != nil {
		return err
	}

	// Run pending migrations (this would be expanded with actual migration logic)
	if currentVersion < 1 {
		// Migration logic would go here
		m.logger.Info().
			Str("family_id", family.ID).
			Int("version", 1).
			Msg("Running migration")
	}

	return nil
}
