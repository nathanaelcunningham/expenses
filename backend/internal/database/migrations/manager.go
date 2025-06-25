// internal/database/migrations/manager.go
package migrations

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"errors"
	"expenses-backend/internal/database/sql/familydb"
	"expenses-backend/internal/database/sql/masterdb"
	"expenses-backend/internal/logger"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

//go:embed sql/master/*.sql
var masterMigrations embed.FS

//go:embed sql/family/*.sql
var familyMigrations embed.FS

// MigrationManager handles database migrations
type MigrationManager struct {
	logger   logger.Logger
	masterDB masterdb.Querier
	familyDB familydb.Querier
}

// Migration represents a database migration
type Migration struct {
	Version     int    `json:"version"`
	Name        string `json:"name"`
	Filename    string `json:"filename"`
	SQL         string `json:"sql"`
	Description string `json:"description"`
	AppliedAt   string `json:"applied_at,omitempty"`
}

// MigrationType represents the type of migration
type MigrationType string

const (
	MasterMigration MigrationType = "master"
	FamilyMigration MigrationType = "family"
)

// MigrationStatus represents the status of migrations
type MigrationStatus struct {
	CurrentVersion    int         `json:"current_version"`
	AppliedMigrations []Migration `json:"applied_migrations"`
	PendingMigrations []Migration `json:"pending_migrations"`
	TotalMigrations   int         `json:"total_migrations"`
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(log logger.Logger, masterDB masterdb.Querier, familyDB familydb.Querier) *MigrationManager {
	return &MigrationManager{
		logger:   log.With(logger.Str("component", "migration-manager")),
		masterDB: masterDB,
		familyDB: familyDB,
	}
}

// LoadMigrations loads migrations from embedded SQL files
func (mm *MigrationManager) LoadMigrations(migrationType MigrationType) ([]Migration, error) {
	var fs embed.FS
	var dirPath string

	switch migrationType {
	case MasterMigration:
		fs = masterMigrations
		dirPath = "sql/master"
	case FamilyMigration:
		fs = familyMigrations
		dirPath = "sql/family"
	default:
		return nil, fmt.Errorf("unsupported migration type: %s", migrationType)
	}

	files, err := fs.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migration directory %s: %w", dirPath, err)
	}

	var migrations []Migration
	migrationRegex := regexp.MustCompile(`^(\d+)_(.+)\.sql$`)

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		matches := migrationRegex.FindStringSubmatch(file.Name())
		if len(matches) != 3 {
			mm.logger.Warn("Skipping migration file with invalid name format (expected: 001_name.sql)", errors.New("skipping migration, invalid format"), logger.Str("filename", file.Name()))
			continue
		}

		version, err := strconv.Atoi(matches[1])
		if err != nil {
			mm.logger.Warn("Failed to parse migration version", err, logger.Str("filename", file.Name()))
			continue
		}

		name := strings.ReplaceAll(matches[2], "_", " ")
		sqlPath := filepath.Join(dirPath, file.Name())

		sqlContent, err := fs.ReadFile(sqlPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		migration := Migration{
			Version:     version,
			Name:        name,
			Filename:    file.Name(),
			SQL:         string(sqlContent),
			Description: mm.extractDescription(string(sqlContent)),
		}

		migrations = append(migrations, migration)
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	mm.logger.Debug("Loaded migrations", logger.Int("count", len(migrations)), logger.Str("type", string(migrationType)))

	return migrations, nil
}

// RunMigrations runs all pending migrations for a database
func (mm *MigrationManager) RunMigrations(ctx context.Context, db *sql.DB, migrationType MigrationType) error {
	// Ensure migrations table exists
	if err := mm.ensureMigrationsTable(ctx, migrationType); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Load migrations from files
	migrations, err := mm.LoadMigrations(migrationType)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	if len(migrations) == 0 {
		mm.logger.Debug("No migrations found", logger.Str("type", string(migrationType)))
		return nil
	}

	// Get current schema version
	currentVersion := mm.getCurrentVersion(ctx, migrationType)

	mm.logger.Info("Starting migration process", logger.Int("current_version", currentVersion), logger.Int("available_migrations", len(migrations)), logger.Str("type", string(migrationType)))

	// Run pending migrations
	migrationsRun := 0
	for _, migration := range migrations {
		if migration.Version <= currentVersion {
			mm.logger.Debug("Skipping already applied migration", logger.Int("version", migration.Version), logger.Str("name", migration.Name))
			continue
		}

		mm.logger.Info("Running migration", logger.Int("version", migration.Version), logger.Str("name", migration.Name), logger.Str("filename", migration.Filename))

		if err := mm.runSingleMigration(ctx, db, migration, migrationType); err != nil {
			return fmt.Errorf("failed to run migration %d (%s): %w",
				migration.Version, migration.Name, err)
		}

		migrationsRun++
		mm.logger.Info("Migration completed successfully", logger.Int("version", migration.Version), logger.Str("name", migration.Name))
	}

	if migrationsRun == 0 {
		mm.logger.Info("No new migrations to run - database is up to date", logger.Str("type", string(migrationType)))
	} else {
		mm.logger.Info("Migration process completed successfully", logger.Int("migrations_run", migrationsRun), logger.Str("type", string(migrationType)))
	}

	return nil
}

// GetMigrationStatus returns the current migration status
func (mm *MigrationManager) GetMigrationStatus(ctx context.Context, migrationType MigrationType) (*MigrationStatus, error) {
	// Ensure migrations table exists
	if err := mm.ensureMigrationsTable(ctx, migrationType); err != nil {
		return nil, fmt.Errorf("failed to create migrations table: %w", err)
	}

	currentVersion := mm.getCurrentVersion(ctx, migrationType)

	appliedMigrations, err := mm.GetAppliedMigrations(ctx, migrationType)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	pendingMigrations, err := mm.GetPendingMigrations(ctx, migrationType)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending migrations: %w", err)
	}

	allMigrations, err := mm.LoadMigrations(migrationType)
	if err != nil {
		return nil, fmt.Errorf("failed to load all migrations: %w", err)
	}

	return &MigrationStatus{
		CurrentVersion:    currentVersion,
		AppliedMigrations: appliedMigrations,
		PendingMigrations: pendingMigrations,
		TotalMigrations:   len(allMigrations),
	}, nil
}

// GetAppliedMigrations returns a list of applied migrations
func (mm *MigrationManager) GetAppliedMigrations(ctx context.Context, migrationType MigrationType) ([]Migration, error) {
	var querier interface {
		GetAppliedMigrations(ctx context.Context) ([]*masterdb.GetAppliedMigrationsRow, error)
	}

	switch migrationType {
	case MasterMigration:
		querier = mm.masterDB
	case FamilyMigration:
		// Cast to the common interface for GetAppliedMigrations
		var familyQuerier interface {
			GetAppliedMigrations(ctx context.Context) ([]*familydb.GetAppliedMigrationsRow, error)
		} = mm.familyDB

		familyRows, err := familyQuerier.GetAppliedMigrations(ctx)
		if err != nil {
			if strings.Contains(err.Error(), "no such table") {
				return []Migration{}, nil
			}
			return nil, fmt.Errorf("failed to query applied migrations: %w", err)
		}

		var migrations []Migration
		for _, row := range familyRows {
			appliedAt := ""
			if row.AppliedAt != nil {
				appliedAt = row.AppliedAt.Format(time.RFC3339)
			}

			migrations = append(migrations, Migration{
				Version:     int(row.Version),
				Name:        row.Name,
				Filename:    row.Filename,
				Description: row.Description,
				AppliedAt:   appliedAt,
			})
		}
		return migrations, nil
	default:
		return nil, fmt.Errorf("unsupported migration type: %s", migrationType)
	}

	masterRows, err := querier.GetAppliedMigrations(ctx)
	if err != nil {
		// If table doesn't exist, return empty slice
		if strings.Contains(err.Error(), "no such table") {
			return []Migration{}, nil
		}
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}

	var migrations []Migration
	for _, row := range masterRows {
		appliedAt := ""
		if row.AppliedAt != nil {
			appliedAt = row.AppliedAt.Format(time.RFC3339)
		}

		migrations = append(migrations, Migration{
			Version:     int(row.Version),
			Name:        row.Name,
			Filename:    row.Filename,
			Description: row.Description,
			AppliedAt:   appliedAt,
		})
	}

	return migrations, nil
}

// GetPendingMigrations returns migrations that haven't been applied yet
func (mm *MigrationManager) GetPendingMigrations(ctx context.Context, migrationType MigrationType) ([]Migration, error) {
	allMigrations, err := mm.LoadMigrations(migrationType)
	if err != nil {
		return nil, err
	}

	currentVersion := mm.getCurrentVersion(ctx, migrationType)

	var pendingMigrations []Migration
	for _, migration := range allMigrations {
		if migration.Version > currentVersion {
			pendingMigrations = append(pendingMigrations, migration)
		}
	}

	return pendingMigrations, nil
}

// ValidateMigrations performs basic validation of migration files
func (mm *MigrationManager) ValidateMigrations(migrationType MigrationType) error {
	migrations, err := mm.LoadMigrations(migrationType)
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		return nil
	}

	// Basic validation - check for empty SQL content
	for _, migration := range migrations {
		if strings.TrimSpace(migration.SQL) == "" {
			return fmt.Errorf("migration %s has empty SQL content", migration.Filename)
		}
	}

	mm.logger.Info("Migration validation completed", logger.Int("count", len(migrations)), logger.Str("type", string(migrationType)))
	return nil
}

// DryRun shows what migrations would be executed without actually running them
func (mm *MigrationManager) DryRun(ctx context.Context, migrationType MigrationType) ([]Migration, error) {
	return mm.GetPendingMigrations(ctx, migrationType)
}

// ensureMigrationsTable creates the migrations table if it doesn't exist
// This is a fallback - normally migration 000 should handle this
func (mm *MigrationManager) ensureMigrationsTable(ctx context.Context, migrationType MigrationType) error {
	var querier interface {
		CheckMigrationsTableExists(ctx context.Context) (int64, error)
		CreateMigrationsTable(ctx context.Context) error
	}

	switch migrationType {
	case MasterMigration:
		querier = mm.masterDB
	case FamilyMigration:
		querier = mm.familyDB
	default:
		return fmt.Errorf("unsupported migration type: %s", migrationType)
	}

	// Try to query the migrations table - if it fails, table doesn't exist
	if _, err := querier.CheckMigrationsTableExists(ctx); err != nil {
		// Table doesn't exist, create it
		mm.logger.Warn("Migration table doesn't exist - this suggests migration 000 hasn't been applied", errors.New("migrations table required"))

		if err := querier.CreateMigrationsTable(ctx); err != nil {
			return fmt.Errorf("failed to create schema_migrations table: %w", err)
		}

		mm.logger.Info("Created schema_migrations table as fallback")
	}

	return nil
}

// getCurrentVersion gets the current schema version
func (mm *MigrationManager) getCurrentVersion(ctx context.Context, migrationType MigrationType) int {
	var querier interface {
		GetCurrentMigrationVersion(ctx context.Context) (int64, error)
	}

	switch migrationType {
	case MasterMigration:
		querier = mm.masterDB
	case FamilyMigration:
		querier = mm.familyDB
	default:
		mm.logger.Debug("Unsupported migration type, assuming version 0")
		return 0
	}

	version, err := querier.GetCurrentMigrationVersion(ctx)
	if err != nil {
		// If the table doesn't exist or query fails, we're at version 0
		mm.logger.Debug("Could not get current version, assuming 0 (this is normal for new databases)", logger.Err(err))
		return 0
	}

	return int(version)
}

// runSingleMigration runs a single migration in a transaction with enhanced tracking
func (mm *MigrationManager) runSingleMigration(ctx context.Context, db *sql.DB, migration Migration, migrationType MigrationType) error {
	startTime := time.Now()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure transaction is properly handled
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			mm.logger.Error("Failed to rollback transaction", err, logger.Int("version", migration.Version))
		}
	}()

	// Execute migration SQL
	if _, err = tx.ExecContext(ctx, migration.SQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL for version %d: %w", migration.Version, err)
	}

	executionTime := time.Since(startTime).Milliseconds()
	checksum := mm.calculateChecksum(migration.SQL)

	// Record migration in migrations table with enhanced metadata
	switch migrationType {
	case MasterMigration:
		// Use masterdb queries with transaction
		masterQuerier := masterdb.New(tx)
		description := &migration.Description
		if migration.Description == "" {
			description = nil
		}
		migrTypeStr := string(migrationType)
		appliedBy := "system"

		if err = masterQuerier.RecordMigration(ctx, masterdb.RecordMigrationParams{
			Version:         int64(migration.Version),
			Name:            migration.Name,
			Filename:        migration.Filename,
			Description:     description,
			Checksum:        &checksum,
			ExecutionTimeMs: &executionTime,
			MigrationType:   &migrTypeStr,
			AppliedBy:       &appliedBy,
		}); err != nil {
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}
	case FamilyMigration:
		// Use familydb queries with transaction
		familyQuerier := familydb.New(tx)
		description := &migration.Description
		if migration.Description == "" {
			description = nil
		}
		migrTypeStr := string(migrationType)
		appliedBy := "system"

		if err = familyQuerier.RecordMigration(ctx, familydb.RecordMigrationParams{
			Version:         int64(migration.Version),
			Name:            migration.Name,
			Filename:        migration.Filename,
			Description:     description,
			Checksum:        &checksum,
			ExecutionTimeMs: &executionTime,
			MigrationType:   &migrTypeStr,
			AppliedBy:       &appliedBy,
		}); err != nil {
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}
	default:
		return fmt.Errorf("unsupported migration type: %s", migrationType)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration transaction for version %d: %w", migration.Version, err)
	}

	mm.logger.Debug("Migration tracking recorded", logger.Int("version", migration.Version), logger.Int64("execution_time_ms", executionTime), logger.Str("checksum", checksum[:8]))

	return nil
}

// extractDescription extracts description from SQL comments
func (mm *MigrationManager) extractDescription(sql string) string {
	lines := strings.Split(sql, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "-- Description:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "-- Description:"))
		}
	}
	return ""
}

// calculateChecksum generates a SHA256 checksum for migration SQL
func (mm *MigrationManager) calculateChecksum(sql string) string {
	h := sha256.Sum256([]byte(sql))
	return fmt.Sprintf("%x", h)
}

// RunStartupMigrations runs master migrations automatically on application startup
// This ensures the master database is always up to date when the application starts
func (mm *MigrationManager) RunStartupMigrations(ctx context.Context, masterDB *sql.DB, familyMasterDB *sql.DB) error {
	mm.logger.Info("Running startup migrations for master database")

	if err := mm.RunMigrations(ctx, masterDB, MasterMigration); err != nil {
		return fmt.Errorf("failed to run master startup migrations: %w", err)
	}

	if err := mm.RunMigrations(ctx, familyMasterDB, FamilyMigration); err != nil {
		return fmt.Errorf("failed to run family master startup migrations: %w", err)
	}

	mm.logger.Info("Startup migrations completed successfully")
	return nil
}

// Helper method to check if a specific migration version has been applied
func (mm *MigrationManager) IsMigrationApplied(ctx context.Context, migrationType MigrationType, version int) (bool, error) {
	var querier interface {
		CheckMigrationApplied(ctx context.Context, version int64) (int64, error)
	}

	switch migrationType {
	case MasterMigration:
		querier = mm.masterDB
	case FamilyMigration:
		querier = mm.familyDB
	default:
		return false, fmt.Errorf("unsupported migration type: %s", migrationType)
	}

	count, err := querier.CheckMigrationApplied(ctx, int64(version))
	if err != nil {
		// If table doesn't exist, migration is not applied
		if strings.Contains(err.Error(), "no such table") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check migration status: %w", err)
	}

	return count > 0, nil
}
