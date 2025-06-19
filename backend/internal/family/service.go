package family

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"expenses-backend/internal/database"

	"expenses-backend/internal/logger"
)

// Service handles family management operations
type Service struct {
	dbManager *database.Manager
	logger    logger.Logger
}

type FamilyError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *FamilyError) Error() string {
	return fmt.Sprintf("family error [%s]: %s", e.Code, e.Message)
}

// Common family error codes
var (
	ErrFamilyNotFound       = &FamilyError{"FAMILY_NOT_FOUND", "Family not found"}
	ErrInvalidInviteCode    = &FamilyError{"INVALID_INVITE_CODE", "Invalid or expired invite code"}
	ErrUserAlreadyInFamily  = &FamilyError{"USER_ALREADY_IN_FAMILY", "User is already a member of a family"}
	ErrInvalidFamilyName    = &FamilyError{"INVALID_FAMILY_NAME", "Family name must be between 1 and 100 characters"}
	ErrDatabaseCreationFail = &FamilyError{"DATABASE_CREATION_FAILED", "Failed to create family database"}
	ErrNotFamilyManager     = &FamilyError{"NOT_FAMILY_MANAGER", "Only family managers can perform this action"}
)

// Memorable words for generating invite codes
var memorableWords = []string{
	"apple", "brave", "cloud", "dance", "eagle", "flame", "grape", "heart",
	"island", "joy", "kite", "light", "moon", "nature", "ocean", "peace",
	"quiet", "river", "star", "tree", "unity", "voice", "water", "bright",
	"calm", "dream", "free", "green", "happy", "love", "magic", "pure",
}

// NewService creates a new family service
func NewService(dbManager *database.DatabaseManager, log logger.Logger) *Service {
	return &Service{
		dbManager: dbManager,
		logger:    log.With(logger.Str("component", "family-service")),
	}
}

// CreateFamily creates a new family and provisions its database
func (s *Service) CreateFamily(ctx context.Context, req CreateFamilyRequest) (*Family, error) {
	// Validate input
	if err := s.validateCreateFamilyRequest(req); err != nil {
		return nil, err
	}

	// Check if user is already in a family
	existingFamily, err := s.GetUserFamily(ctx, req.ManagerID)
	if err == nil && existingFamily != nil {
		return nil, ErrUserAlreadyInFamily
	}

	// Generate unique family ID and invite code
	familyID, err := s.generateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate family ID: %w", err)
	}

	inviteCode, err := s.generateInviteCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate invite code: %w", err)
	}

	// Ensure invite code is unique
	for range 5 {
		exists, err := s.inviteCodeExists(ctx, inviteCode)
		if err != nil {
			return nil, fmt.Errorf("failed to check invite code uniqueness: %w", err)
		}
		if !exists {
			break
		}
		// Generate a new code if this one exists
		inviteCode, err = s.generateInviteCode()
		if err != nil {
			return nil, fmt.Errorf("failed to generate unique invite code: %w", err)
		}
	}

	s.logger.Info("Creating family and provisioning database",
		logger.Str("family_id", familyID),
		logger.Str("manager_id", req.ManagerID),
		logger.Str("invite_code", inviteCode),
	)

	// Provision family database
	familyDB, err := s.dbManager.ProvisionFamilyDatabase(ctx, familyID, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to provision family database: %w", err)
	}

	// Create family record in master database
	masterDB := s.dbManager.GetMasterDatabase()
	tx, err := masterDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert family
	now := time.Now()
	family := &Family{
		ID:          familyID,
		Name:        req.Name,
		InviteCode:  inviteCode,
		DatabaseURL: familyDB.URL,
		ManagerID:   req.ManagerID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	familyQuery := `
		INSERT INTO families (id, name, invite_code, database_url, manager_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.ExecContext(ctx, familyQuery,
		family.ID, family.Name, family.InviteCode, family.DatabaseURL,
		family.ManagerID, family.CreatedAt, family.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create family: %w", err)
	}

	// Add manager as family member
	membershipQuery := `
		INSERT INTO family_memberships (family_id, user_id, role, joined_at)
		VALUES (?, ?, 'manager', ?)`

	_, err = tx.ExecContext(ctx, membershipQuery, familyID, req.ManagerID, now)
	if err != nil {
		return nil, fmt.Errorf("failed to add manager to family: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit family creation: %w", err)
	}

	// Add manager to family database
	if err := s.addMemberToFamilyDatabase(ctx, familyID, req.ManagerID, req.ManagerName, req.ManagerEmail, "manager"); err != nil {
		s.logger.Warn("Failed to add manager to family database - this may cause issues",
			err,
			logger.Str("family_id", familyID),
			logger.Str("manager_id", req.ManagerID),
		)
	}

	s.logger.Info("Family created successfully",
		logger.Str("family_id", familyID),
		logger.Str("family_name", req.Name),
		logger.Str("invite_code", inviteCode),
	)

	return family, nil
}

// JoinFamily allows a user to join a family using an invite code
func (s *Service) JoinFamily(ctx context.Context, req JoinFamilyRequest) (*Family, error) {
	// Validate input
	if req.InviteCode == "" || req.UserID == "" {
		return nil, &FamilyError{"INVALID_REQUEST", "Invite code and user ID are required"}
	}

	// Check if user is already in a family
	existingFamily, err := s.GetUserFamily(ctx, req.UserID)
	if err == nil && existingFamily != nil {
		return nil, ErrUserAlreadyInFamily
	}

	// Find family by invite code
	family, err := s.getFamilyByInviteCode(ctx, req.InviteCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvalidInviteCode
		}
		return nil, fmt.Errorf("failed to find family: %w", err)
	}

	// Add user to family
	masterDB := s.dbManager.GetMasterDatabase()
	membershipQuery := `
		INSERT INTO family_memberships (family_id, user_id, role, joined_at)
		VALUES (?, ?, 'member', ?)`

	_, err = masterDB.ExecContext(ctx, membershipQuery, family.ID, req.UserID, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to add user to family: %w", err)
	}

	// Add member to family database
	if err := s.addMemberToFamilyDatabase(ctx, family.ID, req.UserID, req.UserName, req.UserEmail, "member"); err != nil {
		s.logger.Warn("Failed to add member to family database", err, logger.Str("family_id", family.ID), logger.Str("user_id", req.UserID))
	}

	s.logger.Info("User joined family successfully", logger.Str("family_id", family.ID), logger.Str("user_id", req.UserID), logger.Str("invite_code", req.InviteCode))

	return family, nil
}

// GetUserFamily retrieves the family that a user belongs to
func (s *Service) GetUserFamily(ctx context.Context, userID string) (*Family, error) {
	masterDB := s.dbManager.GetMasterDatabase()

	query := `
		SELECT f.id, f.name, f.invite_code, f.database_url, f.manager_id, f.created_at, f.updated_at
		FROM families f
		JOIN family_memberships fm ON f.id = fm.family_id
		WHERE fm.user_id = ?`

	family := &Family{}
	err := masterDB.QueryRowContext(ctx, query, userID).Scan(
		&family.ID, &family.Name, &family.InviteCode, &family.DatabaseURL,
		&family.ManagerID, &family.CreatedAt, &family.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User is not in any family
		}
		return nil, fmt.Errorf("failed to get user family: %w", err)
	}

	// Get family members
	members, err := s.getFamilyMembers(ctx, family.ID)
	if err != nil {
		s.logger.Warn("Failed to get family members", err, logger.Str("family_id", family.ID))
	} else {
		family.Members = members
	}

	return family, nil
}

// GetFamilyByID retrieves a family by its ID
func (s *Service) GetFamilyByID(ctx context.Context, familyID string) (*Family, error) {
	masterDB := s.dbManager.GetMasterDatabase()

	query := `
		SELECT id, name, invite_code, database_url, manager_id, created_at, updated_at
		FROM families WHERE id = ?`

	family := &Family{}
	err := masterDB.QueryRowContext(ctx, query, familyID).Scan(
		&family.ID, &family.Name, &family.InviteCode, &family.DatabaseURL,
		&family.ManagerID, &family.CreatedAt, &family.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrFamilyNotFound
		}
		return nil, fmt.Errorf("failed to get family: %w", err)
	}

	// Get family members
	members, err := s.getFamilyMembers(ctx, family.ID)
	if err != nil {
		s.logger.Warn("Failed to get family members", err, logger.Str("family_id", family.ID))
	} else {
		family.Members = members
	}

	return family, nil
}

// RemoveFamilyMember removes a member from a family (manager only)
func (s *Service) RemoveFamilyMember(ctx context.Context, familyID, managerID, memberID string) error {
	// Verify the requester is the family manager
	if !s.isUserFamilyManager(ctx, familyID, managerID) {
		return ErrNotFamilyManager
	}

	// Cannot remove the manager
	if managerID == memberID {
		return &FamilyError{"CANNOT_REMOVE_MANAGER", "Family manager cannot be removed"}
	}

	masterDB := s.dbManager.GetMasterDatabase()

	// Remove from master database
	query := `DELETE FROM family_memberships WHERE family_id = ? AND user_id = ?`
	result, err := masterDB.ExecContext(ctx, query, familyID, memberID)
	if err != nil {
		return fmt.Errorf("failed to remove family member: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return &FamilyError{"MEMBER_NOT_FOUND", "Member not found in family"}
	}

	// Remove from family database
	if err := s.removeMemberFromFamilyDatabase(ctx, familyID, memberID); err != nil {
		s.logger.Warn("Failed to remove member from family database", err, logger.Str("family_id", familyID), logger.Str("member_id", memberID))
	}

	s.logger.Info("Family member removed successfully",
		logger.Str("family_id", familyID),
		logger.Str("member_id", memberID),
		logger.Str("manager_id", managerID),
	)

	return nil
}

// DeleteFamily completely removes a family and its database (manager only)
func (s *Service) DeleteFamily(ctx context.Context, familyID, managerID string) error {
	// Verify the requester is the family manager
	if !s.isUserFamilyManager(ctx, familyID, managerID) {
		return ErrNotFamilyManager
	}

	// Delete the family database and master database records
	if err := s.dbManager.DeleteFamilyDatabase(ctx, familyID); err != nil {
		return fmt.Errorf("failed to delete family database: %w", err)
	}

	s.logger.Info("Family deleted successfully", logger.Str("family_id", familyID), logger.Str("manager_id", managerID))

	return nil
}

// RegenerateInviteCode generates a new invite code for a family (manager only)
func (s *Service) RegenerateInviteCode(ctx context.Context, familyID, managerID string) (string, error) {
	// Verify the requester is the family manager
	if !s.isUserFamilyManager(ctx, familyID, managerID) {
		return "", ErrNotFamilyManager
	}

	// Generate new invite code
	newInviteCode, err := s.generateInviteCode()
	if err != nil {
		return "", fmt.Errorf("failed to generate new invite code: %w", err)
	}

	// Ensure it's unique
	for range 5 {
		exists, err := s.inviteCodeExists(ctx, newInviteCode)
		if err != nil {
			return "", fmt.Errorf("failed to check invite code uniqueness: %w", err)
		}
		if !exists {
			break
		}
		newInviteCode, err = s.generateInviteCode()
		if err != nil {
			return "", fmt.Errorf("failed to generate unique invite code: %w", err)
		}
	}

	// Update family record
	masterDB := s.dbManager.GetMasterDatabase()
	query := `UPDATE families SET invite_code = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

	_, err = masterDB.ExecContext(ctx, query, newInviteCode, familyID)
	if err != nil {
		return "", fmt.Errorf("failed to update invite code: %w", err)
	}

	s.logger.Info("Family invite code regenerated", logger.Str("family_id", familyID), logger.Str("new_invite_code", newInviteCode))

	return newInviteCode, nil
}

// Helper methods

func (s *Service) validateCreateFamilyRequest(req CreateFamilyRequest) error {
	if req.Name == "" || len(req.Name) > 100 {
		return ErrInvalidFamilyName
	}
	if req.ManagerID == "" {
		return &FamilyError{"INVALID_MANAGER", "Manager ID is required"}
	}
	return nil
}

func (s *Service) generateID() (string, error) {
	// Generate a random 16-byte ID
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Convert to hex string
	id := fmt.Sprintf("%x", bytes)
	return id, nil
}

func (s *Service) generateInviteCode() (string, error) {
	// Generate a memorable 3-word invite code
	words := make([]string, 3)

	for i := range 3 {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(memorableWords))))
		if err != nil {
			return "", err
		}
		words[i] = memorableWords[index.Int64()]
	}

	// Add a random number for uniqueness
	num, err := rand.Int(rand.Reader, big.NewInt(999))
	if err != nil {
		return "", err
	}

	code := fmt.Sprintf("%s-%s-%s-%03d", words[0], words[1], words[2], num.Int64())
	return code, nil
}

func (s *Service) inviteCodeExists(ctx context.Context, inviteCode string) (bool, error) {
	masterDB := s.dbManager.GetMasterDatabase()
	var count int
	query := `SELECT COUNT(*) FROM families WHERE invite_code = ?`
	err := masterDB.QueryRowContext(ctx, query, inviteCode).Scan(&count)
	return count > 0, err
}

func (s *Service) getFamilyByInviteCode(ctx context.Context, inviteCode string) (*Family, error) {
	masterDB := s.dbManager.GetMasterDatabase()

	query := `
		SELECT id, name, invite_code, database_url, manager_id, created_at, updated_at
		FROM families WHERE invite_code = ?`

	family := &Family{}
	err := masterDB.QueryRowContext(ctx, query, inviteCode).Scan(
		&family.ID, &family.Name, &family.InviteCode, &family.DatabaseURL,
		&family.ManagerID, &family.CreatedAt, &family.UpdatedAt)

	return family, err
}

func (s *Service) getFamilyMembers(ctx context.Context, familyID string) ([]Member, error) {
	masterDB := s.dbManager.GetMasterDatabase()

	query := `
		SELECT u.id, u.name, u.email, fm.role, fm.joined_at
		FROM family_memberships fm
		JOIN users u ON fm.user_id = u.id
		WHERE fm.family_id = ?
		ORDER BY fm.joined_at ASC`

	rows, err := masterDB.QueryContext(ctx, query, familyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []Member
	for rows.Next() {
		var member Member
		err := rows.Scan(&member.UserID, &member.Name, &member.Email, &member.Role, &member.JoinedAt)
		if err != nil {
			return nil, err
		}
		member.IsActive = true // All members in the database are active
		members = append(members, member)
	}

	return members, rows.Err()
}

func (s *Service) isUserFamilyManager(ctx context.Context, familyID, userID string) bool {
	masterDB := s.dbManager.GetMasterDatabase()

	query := `
		SELECT COUNT(*) FROM family_memberships 
		WHERE family_id = ? AND user_id = ? AND role = 'manager'`

	var count int
	err := masterDB.QueryRowContext(ctx, query, familyID, userID).Scan(&count)
	if err != nil {
		s.logger.Error("Failed to check if user is family manager", err, logger.Str("family_id", familyID), logger.Str("user_id", userID))
		return false
	}

	return count > 0
}

func (s *Service) addMemberToFamilyDatabase(ctx context.Context, familyID, userID, userName, userEmail, role string) error {
	familyDB, err := s.dbManager.GetFamilyDatabase(ctx, familyID)
	if err != nil {
		return fmt.Errorf("failed to get family database: %w", err)
	}

	query := `
		INSERT OR REPLACE INTO family_members (id, name, email, role, joined_at, is_active)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, TRUE)`

	_, err = familyDB.ExecContext(ctx, query, userID, userName, userEmail, role)
	return err
}

func (s *Service) removeMemberFromFamilyDatabase(ctx context.Context, familyID, userID string) error {
	familyDB, err := s.dbManager.GetFamilyDatabase(ctx, familyID)
	if err != nil {
		return fmt.Errorf("failed to get family database: %w", err)
	}

	query := `UPDATE family_members SET is_active = FALSE WHERE id = ?`
	_, err = familyDB.ExecContext(ctx, query, userID)
	return err
}
