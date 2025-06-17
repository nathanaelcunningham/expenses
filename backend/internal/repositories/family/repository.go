package family

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"expenses-backend/internal/models"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// Repository implements the Store interface for family operations
type Repository struct {
	db     *sql.DB
	logger zerolog.Logger
}

// NewRepository creates a new family repository
func NewRepository(db *sql.DB, logger zerolog.Logger) Store {
	return &Repository{
		db:     db,
		logger: logger.With().Str("component", "family-repository").Logger(),
	}
}

// CreateFamily creates a new family in the master database
func (r *Repository) CreateFamily(ctx context.Context, req *models.CreateFamilyRequest) (*models.Family, error) {
	// Generate family ID and invite code
	familyID, err := r.generateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate family ID: %w", err)
	}

	inviteCode, err := r.GenerateInviteCode(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invite code: %w", err)
	}

	now := time.Now()
	family := &models.Family{
		ID:            familyID,
		Name:          strings.TrimSpace(req.Name),
		InviteCode:    inviteCode,
		ManagerID:     req.ManagerID,
		SchemaVersion: 0, // Will be updated after database provisioning
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Begin transaction for atomic family creation
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert family record
	query := `
		INSERT INTO families (id, name, invite_code, database_url, manager_id, schema_version, created_at, updated_at)
		VALUES (?, ?, ?, '', ?, ?, ?, ?)`

	_, err = tx.ExecContext(ctx, query,
		family.ID, family.Name, family.InviteCode, family.ManagerID,
		family.SchemaVersion, family.CreatedAt, family.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create family: %w", err)
	}

	// Add manager as family member
	memberQuery := `
		INSERT INTO family_memberships (family_id, user_id, role, joined_at)
		VALUES (?, ?, 'manager', ?)`

	_, err = tx.ExecContext(ctx, memberQuery, family.ID, family.ManagerID, now)
	if err != nil {
		return nil, fmt.Errorf("failed to add manager as family member: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit family creation: %w", err)
	}

	r.logger.Info().
		Str("family_id", family.ID).
		Str("manager_id", family.ManagerID).
		Str("family_name", family.Name).
		Msg("Family created successfully")

	return family, nil
}

// GetFamilyByID retrieves a family by its ID
func (r *Repository) GetFamilyByID(ctx context.Context, familyID string) (*models.Family, error) {
	family := &models.Family{}
	query := `
		SELECT id, name, invite_code, database_url, manager_id, schema_version, created_at, updated_at
		FROM families WHERE id = ?`

	err := r.db.QueryRowContext(ctx, query, familyID).Scan(
		&family.ID, &family.Name, &family.InviteCode, &family.DatabaseURL,
		&family.ManagerID, &family.SchemaVersion, &family.CreatedAt, &family.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("family not found: %s", familyID)
		}
		return nil, fmt.Errorf("failed to get family: %w", err)
	}

	return family, nil
}

// GetFamilyByInviteCode retrieves a family by its invite code
func (r *Repository) GetFamilyByInviteCode(ctx context.Context, inviteCode string) (*models.Family, error) {
	family := &models.Family{}
	query := `
		SELECT id, name, invite_code, database_url, manager_id, schema_version, created_at, updated_at
		FROM families WHERE invite_code = ?`

	err := r.db.QueryRowContext(ctx, query, inviteCode).Scan(
		&family.ID, &family.Name, &family.InviteCode, &family.DatabaseURL,
		&family.ManagerID, &family.SchemaVersion, &family.CreatedAt, &family.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("family not found with invite code: %s", inviteCode)
		}
		return nil, fmt.Errorf("failed to get family by invite code: %w", err)
	}

	return family, nil
}

// UpdateFamily updates a family's information
func (r *Repository) UpdateFamily(ctx context.Context, family *models.Family) error {
	family.UpdatedAt = time.Now()
	query := `
		UPDATE families 
		SET name = ?, invite_code = ?, database_url = ?, manager_id = ?, schema_version = ?, updated_at = ?
		WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query,
		family.Name, family.InviteCode, family.DatabaseURL, family.ManagerID,
		family.SchemaVersion, family.UpdatedAt, family.ID)
	if err != nil {
		return fmt.Errorf("failed to update family: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("family not found: %s", family.ID)
	}

	return nil
}

// DeleteFamily removes a family and all its memberships
func (r *Repository) DeleteFamily(ctx context.Context, familyID string) error {
	// The foreign key constraints will cascade delete family_memberships
	query := `DELETE FROM families WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, familyID)
	if err != nil {
		return fmt.Errorf("failed to delete family: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("family not found: %s", familyID)
	}

	r.logger.Info().
		Str("family_id", familyID).
		Msg("Family deleted successfully")

	return nil
}

// GetFamiliesByManagerID retrieves all families managed by a user
func (r *Repository) GetFamiliesByManagerID(ctx context.Context, managerID string) ([]*models.Family, error) {
	query := `
		SELECT id, name, invite_code, database_url, manager_id, schema_version, created_at, updated_at
		FROM families WHERE manager_id = ? ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, managerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get families by manager: %w", err)
	}
	defer rows.Close()

	var families []*models.Family
	for rows.Next() {
		family := &models.Family{}
		err := rows.Scan(
			&family.ID, &family.Name, &family.InviteCode, &family.DatabaseURL,
			&family.ManagerID, &family.SchemaVersion, &family.CreatedAt, &family.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan family: %w", err)
		}
		families = append(families, family)
	}

	return families, rows.Err()
}

// AddFamilyMember adds a user to a family
func (r *Repository) AddFamilyMember(ctx context.Context, familyID, userID, role string) error {
	// Validate role
	if role != "manager" && role != "member" {
		return fmt.Errorf("invalid role: %s", role)
	}

	query := `
		INSERT INTO family_memberships (family_id, user_id, role, joined_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)`

	_, err := r.db.ExecContext(ctx, query, familyID, userID, role)
	if err != nil {
		return fmt.Errorf("failed to add family member: %w", err)
	}

	r.logger.Info().
		Str("family_id", familyID).
		Str("user_id", userID).
		Str("role", role).
		Msg("User added to family")

	return nil
}

// RemoveFamilyMember removes a user from a family
func (r *Repository) RemoveFamilyMember(ctx context.Context, familyID, userID string) error {
	query := `DELETE FROM family_memberships WHERE family_id = ? AND user_id = ?`

	result, err := r.db.ExecContext(ctx, query, familyID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove family member: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("family membership not found")
	}

	r.logger.Info().
		Str("family_id", familyID).
		Str("user_id", userID).
		Msg("User removed from family")

	return nil
}

// GetFamilyMemberships retrieves all memberships for a family
func (r *Repository) GetFamilyMemberships(ctx context.Context, familyID string) ([]*models.FamilyMembership, error) {
	query := `
		SELECT family_id, user_id, role, joined_at
		FROM family_memberships WHERE family_id = ? ORDER BY joined_at ASC`

	rows, err := r.db.QueryContext(ctx, query, familyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get family memberships: %w", err)
	}
	defer rows.Close()

	var memberships []*models.FamilyMembership
	for rows.Next() {
		membership := &models.FamilyMembership{}
		err := rows.Scan(&membership.FamilyID, &membership.UserID, &membership.Role, &membership.JoinedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan membership: %w", err)
		}
		memberships = append(memberships, membership)
	}

	return memberships, rows.Err()
}

// GetUserFamilyMembership retrieves a user's family membership
func (r *Repository) GetUserFamilyMembership(ctx context.Context, userID string) (*models.FamilyMembership, error) {
	membership := &models.FamilyMembership{}
	query := `
		SELECT family_id, user_id, role, joined_at
		FROM family_memberships WHERE user_id = ?`

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&membership.FamilyID, &membership.UserID, &membership.Role, &membership.JoinedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user is not a member of any family: %s", userID)
		}
		return nil, fmt.Errorf("failed to get user family membership: %w", err)
	}

	return membership, nil
}

// UpdateMemberRole updates a family member's role
func (r *Repository) UpdateMemberRole(ctx context.Context, familyID, userID, newRole string) error {
	// Validate role
	if newRole != "manager" && newRole != "member" {
		return fmt.Errorf("invalid role: %s", newRole)
	}

	query := `UPDATE family_memberships SET role = ? WHERE family_id = ? AND user_id = ?`

	result, err := r.db.ExecContext(ctx, query, newRole, familyID, userID)
	if err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("family membership not found")
	}

	r.logger.Info().
		Str("family_id", familyID).
		Str("user_id", userID).
		Str("new_role", newRole).
		Msg("Member role updated")

	return nil
}

// GetFamilyWithMembers retrieves a family along with its memberships
func (r *Repository) GetFamilyWithMembers(ctx context.Context, familyID string) (*models.FamilyWithMembers, error) {
	family, err := r.GetFamilyByID(ctx, familyID)
	if err != nil {
		return nil, err
	}

	memberships, err := r.GetFamilyMemberships(ctx, familyID)
	if err != nil {
		return nil, err
	}

	return &models.FamilyWithMembers{
		Family:      family,
		Memberships: memberships,
	}, nil
}

// FamilyExists checks if a family exists
func (r *Repository) FamilyExists(ctx context.Context, familyID string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM families WHERE id = ?`
	err := r.db.QueryRowContext(ctx, query, familyID).Scan(&count)
	return count > 0, err
}

// UserIsFamilyMember checks if a user is a member of a specific family
func (r *Repository) UserIsFamilyMember(ctx context.Context, familyID, userID string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM family_memberships WHERE family_id = ? AND user_id = ?`
	err := r.db.QueryRowContext(ctx, query, familyID, userID).Scan(&count)
	return count > 0, err
}

// UserIsFamilyManager checks if a user is a manager of a specific family
func (r *Repository) UserIsFamilyManager(ctx context.Context, familyID, userID string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM family_memberships WHERE family_id = ? AND user_id = ? AND role = 'manager'`
	err := r.db.QueryRowContext(ctx, query, familyID, userID).Scan(&count)
	return count > 0, err
}

// GenerateInviteCode generates a unique invite code for a family
func (r *Repository) GenerateInviteCode(ctx context.Context) (string, error) {
	const maxAttempts = 10

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Generate 6-character invite code
		bytes := make([]byte, 3)
		if _, err := rand.Read(bytes); err != nil {
			return "", fmt.Errorf("failed to generate random bytes: %w", err)
		}

		inviteCode := strings.ToUpper(hex.EncodeToString(bytes))

		// Check if code already exists
		var count int
		query := `SELECT COUNT(*) FROM families WHERE invite_code = ?`
		err := r.db.QueryRowContext(ctx, query, inviteCode).Scan(&count)
		if err != nil {
			return "", fmt.Errorf("failed to check invite code uniqueness: %w", err)
		}

		if count == 0 {
			return inviteCode, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique invite code after %d attempts", maxAttempts)
}

// Helper method to generate IDs
func (r *Repository) generateID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

