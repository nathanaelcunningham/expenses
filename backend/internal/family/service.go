package family

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"expenses-backend/internal/database"
	"expenses-backend/internal/database/sql/masterdb"

	"expenses-backend/internal/logger"
)

// Service handles family management operations
type Service struct {
	dbManager *database.DatabaseManager
	logger    logger.Logger
}

type FamilyError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *FamilyError) Error() string {
	return fmt.Sprintf("family error [%s]: %s", e.Code, e.Message)
}

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

func NewService(dbManager *database.DatabaseManager, log logger.Logger) *Service {
	return &Service{
		dbManager: dbManager,
		logger:    log.With(logger.Str("component", "family-service")),
	}
}

func (s *Service) CreateFamily(ctx context.Context, req CreateFamilyRequest) (*masterdb.Family, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if user is already in a family
	existingFamily, err := s.dbManager.GetMasterQueries().CheckUserExistsInFamily(ctx, &req.ManagerID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if user exists in family: %w", err)
	}
	if existingFamily > 0 {
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

	var family *masterdb.Family
	now := time.Now()

	// Create family record and membership in master database transaction
	err = s.dbManager.WithMasterTx(ctx, func(q *masterdb.Queries) error {
		// Create family
		var schemaVersion int64 = 0
		createFamilyParams := masterdb.CreateFamilyParams{
			ID:            familyID,
			Name:          req.Name,
			InviteCode:    inviteCode,
			DatabaseUrl:   familyDB.URL,
			ManagerID:     req.ManagerID,
			SchemaVersion: &schemaVersion,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		sqlcFamily, err := q.CreateFamily(ctx, createFamilyParams)
		if err != nil {
			return fmt.Errorf("failed to create family: %w", err)
		}

		// Create membership for manager
		createMembershipParams := masterdb.CreateFamilyMembershipParams{
			FamilyID: &familyID,
			UserID:   &req.ManagerID,
			Role:     "manager",
			JoinedAt: now,
		}
		_, err = q.CreateFamilyMembership(ctx, createMembershipParams)
		if err != nil {
			return fmt.Errorf("failed to add manager to family: %w", err)
		}

		family = sqlcFamily
		return nil
	})
	if err != nil {
		return nil, err
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
func (s *Service) JoinFamily(ctx context.Context, req JoinFamilyRequest) (*masterdb.Family, error) {
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
	family, err := s.dbManager.GetMasterQueries().GetFamilyByInviteCode(ctx, req.InviteCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvalidInviteCode
		}
		return nil, fmt.Errorf("failed to find family: %w", err)
	}

	now := time.Now()

	// Add user to family
	createMembershipParams := masterdb.CreateFamilyMembershipParams{
		FamilyID: &family.ID,
		UserID:   &req.UserID,
		Role:     "member",
		JoinedAt: now,
	}
	_, err = s.dbManager.GetMasterQueries().CreateFamilyMembership(ctx, createMembershipParams)
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
func (s *Service) GetUserFamily(ctx context.Context, userID string) (*FamilyResponse, error) {
	// Get user family info first
	userFamilyInfo, err := s.dbManager.GetMasterQueries().GetUserFamilyInfo(ctx, &userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User is not in any family
		}
		return nil, fmt.Errorf("failed to get user family info: %w", err)
	}

	if userFamilyInfo.FamilyID == nil {
		return nil, nil // User is not in any family
	}

	// Get the full family details
	sqlcFamily, err := s.dbManager.GetMasterQueries().GetFamilyByID(ctx, *userFamilyInfo.FamilyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get family details: %w", err)
	}

	// Get family members
	members, err := s.getFamilyMembers(ctx, sqlcFamily.ID)
	if err != nil {
		s.logger.Warn("Failed to get family members", err, logger.Str("family_id", sqlcFamily.ID))
		members = []MemberResponse{} // Empty slice on error
	}

	familyResponse := &FamilyResponse{
		Family:  sqlcFamily,
		Members: members,
	}

	return familyResponse, nil
}

// GetFamilyByID retrieves a family by its ID
func (s *Service) GetFamilyByID(ctx context.Context, familyID string) (*FamilyResponse, error) {
	sqlcFamily, err := s.dbManager.GetMasterQueries().GetFamilyByID(ctx, familyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrFamilyNotFound
		}
		return nil, fmt.Errorf("failed to get family: %w", err)
	}

	// Get family members
	members, err := s.getFamilyMembers(ctx, sqlcFamily.ID)
	if err != nil {
		s.logger.Warn("Failed to get family members", err, logger.Str("family_id", sqlcFamily.ID))
		members = []MemberResponse{} // Empty slice on error
	}

	familyResponse := &FamilyResponse{
		Family:  sqlcFamily,
		Members: members,
	}

	return familyResponse, nil
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

	// Remove from master database
	deleteMembershipParams := masterdb.DeleteFamilyMembershipParams{
		FamilyID: &familyID,
		UserID:   &memberID,
	}
	err := s.dbManager.GetMasterQueries().DeleteFamilyMembership(ctx, deleteMembershipParams)
	if err != nil {
		return fmt.Errorf("failed to remove family member: %w", err)
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

	// Update family record - using raw SQL since SQLC UpdateFamily doesn't include invite_code
	masterDB := s.dbManager.GetMasterDB()
	query := `UPDATE families SET invite_code = ?, updated_at = ? WHERE id = ?`

	_, err = masterDB.ExecContext(ctx, query, newInviteCode, time.Now(), familyID)
	if err != nil {
		return "", fmt.Errorf("failed to update invite code: %w", err)
	}

	s.logger.Info("Family invite code regenerated", logger.Str("family_id", familyID), logger.Str("new_invite_code", newInviteCode))

	return newInviteCode, nil
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
	_, err := s.dbManager.GetMasterQueries().GetFamilyByInviteCode(ctx, inviteCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Code doesn't exist
		}
		return false, err // Other error
	}
	return true, nil // Code exists
}

func (s *Service) getFamilyMembers(ctx context.Context, familyID string) ([]MemberResponse, error) {
	// Use SQLC to get family memberships
	sqlcMemberships, err := s.dbManager.GetMasterQueries().ListFamilyMemberships(ctx, &familyID)
	if err != nil {
		return nil, err
	}

	var members []MemberResponse
	for _, membership := range sqlcMemberships {
		if membership.UserID == nil {
			continue
		}

		// Get user details for each membership
		user, err := s.dbManager.GetMasterQueries().GetUserByID(ctx, *membership.UserID)
		if err != nil {
			s.logger.Warn("Failed to get user details for member", err, logger.Str("user_id", *membership.UserID))
			continue
		}

		member := MemberResponse{
			UserID:   *membership.UserID,
			Name:     user.Name,
			Email:    user.Email,
			Role:     membership.Role,
			JoinedAt: membership.JoinedAt,
			IsActive: true, // All members in the database are active
		}
		members = append(members, member)
	}

	return members, nil
}

func (s *Service) isUserFamilyManager(ctx context.Context, familyID, userID string) bool {
	getMembershipParams := masterdb.GetFamilyMembershipParams{
		FamilyID: &familyID,
		UserID:   &userID,
	}

	membership, err := s.dbManager.GetMasterQueries().GetFamilyMembership(ctx, getMembershipParams)
	if err != nil {
		if err != sql.ErrNoRows {
			s.logger.Error("Failed to check if user is family manager", err, logger.Str("family_id", familyID), logger.Str("user_id", userID))
		}
		return false
	}

	return membership.Role == "manager"
}

func (s *Service) addMemberToFamilyDatabase(ctx context.Context, familyID, userID, userName, userEmail, role string) error {
	familyDB, err := s.dbManager.GetFamilyDB(familyID)
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
	familyDB, err := s.dbManager.GetFamilyDB(familyID)
	if err != nil {
		return fmt.Errorf("failed to get family database: %w", err)
	}

	query := `UPDATE family_members SET is_active = FALSE WHERE id = ?`
	_, err = familyDB.ExecContext(ctx, query, userID)
	return err
}
