package family

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"expenses-backend/internal/database"
	"expenses-backend/internal/database/sql/familydb"
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
	userID := int64(req.ManagerID)
	existingFamily, err := s.dbManager.GetMasterQueries().CheckUserExistsInFamily(ctx, &userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if user exists in family: %w", err)
	}
	if existingFamily > 0 {
		return nil, ErrUserAlreadyInFamily
	}

	// Generate unique family ID and invite code
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
		logger.Int64("manager_id", userID),
		logger.Str("invite_code", inviteCode),
	)

	// Provision family database
	familydb, familydbInfo, err := s.dbManager.ProvisionFamilyDatabase(ctx, req.Name)
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
			Name:          req.Name,
			InviteCode:    inviteCode,
			ManagerID:     userID,
			SchemaVersion: &schemaVersion,
			CreatedAt:     now,
			UpdatedAt:     now,
			DatabaseUrl:   familydbInfo.URL,
		}
		sqlcFamily, err := q.CreateFamily(ctx, createFamilyParams)
		if err != nil {
			return fmt.Errorf("failed to create family: %w", err)
		}

		// Create membership for manager using the generated family ID
		createMembershipParams := masterdb.CreateFamilyMembershipParams{
			FamilyID: &sqlcFamily.ID,
			UserID:   &userID,
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

	// Add database to internal maps now that we have the family ID
	s.dbManager.AddFamilyDB(int(family.ID), familydb)

	// Add manager to family database
	if err := s.addMemberToFamilyDatabase(ctx, int(family.ID), int(req.ManagerID), req.ManagerName, req.ManagerEmail, "manager"); err != nil {
		s.logger.Warn("Failed to add manager to family database - this may cause issues",
			err,
			logger.Int64("family_id", family.ID),
			logger.Int64("manager_id", req.ManagerID),
		)
	}

	s.logger.Info("Family created successfully",
		logger.Int64("family_id", family.ID),
		logger.Str("family_name", req.Name),
		logger.Str("invite_code", inviteCode),
	)

	return family, nil
}

// JoinFamily allows a user to join a family using an invite code
func (s *Service) JoinFamily(ctx context.Context, req JoinFamilyRequest) (*masterdb.Family, error) {
	// Validate input
	if req.InviteCode == "" || req.UserID == 0 {
		return nil, &FamilyError{"INVALID_REQUEST", "Invite code and user ID are required"}
	}

	// Check if user is already in a family
	existingFamily, err := s.GetUserFamily(ctx, int(req.UserID))
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
	uID := int64(req.UserID)
	createMembershipParams := masterdb.CreateFamilyMembershipParams{
		FamilyID: &family.ID,
		UserID:   &uID,
		Role:     "member",
		JoinedAt: now,
	}
	_, err = s.dbManager.GetMasterQueries().CreateFamilyMembership(ctx, createMembershipParams)
	if err != nil {
		return nil, fmt.Errorf("failed to add user to family: %w", err)
	}

	// Add member to family database
	if err := s.addMemberToFamilyDatabase(ctx, int(family.ID), int(req.UserID), req.UserName, req.UserEmail, "member"); err != nil {
		s.logger.Warn("Failed to add member to family database", err, logger.Int64("family_id", family.ID), logger.Int64("user_id", req.UserID))
	}

	s.logger.Info("User joined family successfully", logger.Int64("family_id", family.ID), logger.Int64("user_id", req.UserID), logger.Str("invite_code", req.InviteCode))

	return family, nil
}

// GetUserFamily retrieves the family that a user belongs to
func (s *Service) GetUserFamily(ctx context.Context, userID int) (*FamilyResponse, error) {
	uID := int64(userID)
	// Get user family info first
	userFamilyInfo, err := s.dbManager.GetMasterQueries().GetUserFamilyInfo(ctx, &uID)
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
	members, err := s.getFamilyMembers(ctx, int(sqlcFamily.ID))
	if err != nil {
		s.logger.Warn("Failed to get family members", err, logger.Int64("family_id", sqlcFamily.ID))
		members = []MemberResponse{} // Empty slice on error
	}

	familyResponse := &FamilyResponse{
		Family:  sqlcFamily,
		Members: members,
	}

	return familyResponse, nil
}

// GetFamilyByID retrieves a family by its ID
func (s *Service) GetFamilyByID(ctx context.Context, familyID int) (*FamilyResponse, error) {
	sqlcFamily, err := s.dbManager.GetMasterQueries().GetFamilyByID(ctx, int64(familyID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrFamilyNotFound
		}
		return nil, fmt.Errorf("failed to get family: %w", err)
	}

	// Get family members
	members, err := s.getFamilyMembers(ctx, int(sqlcFamily.ID))
	if err != nil {
		s.logger.Warn("Failed to get family members", err, logger.Int64("family_id", sqlcFamily.ID))
		members = []MemberResponse{} // Empty slice on error
	}

	familyResponse := &FamilyResponse{
		Family:  sqlcFamily,
		Members: members,
	}

	return familyResponse, nil
}

// RemoveFamilyMember removes a member from a family (manager only)
func (s *Service) RemoveFamilyMember(ctx context.Context, familyID, managerID, memberID int) error {
	// Verify the requester is the family manager
	if !s.isUserFamilyManager(ctx, familyID, managerID) {
		return ErrNotFamilyManager
	}

	// Cannot remove the manager
	if managerID == memberID {
		return &FamilyError{"CANNOT_REMOVE_MANAGER", "Family manager cannot be removed"}
	}

	// Remove from master database
	fID := int64(familyID)
	mID := int64(memberID)
	deleteMembershipParams := masterdb.DeleteFamilyMembershipParams{
		FamilyID: &fID,
		UserID:   &mID,
	}
	err := s.dbManager.GetMasterQueries().DeleteFamilyMembership(ctx, deleteMembershipParams)
	if err != nil {
		return fmt.Errorf("failed to remove family member: %w", err)
	}

	// Remove from family database
	if err := s.removeMemberFromFamilyDatabase(ctx, familyID, memberID); err != nil {
		s.logger.Warn("Failed to remove member from family database", err, logger.Int64("family_id", int64(familyID)), logger.Int64("member_id", int64(memberID)))
	}

	s.logger.Info("Family member removed successfully",
		logger.Int64("family_id", int64(familyID)),
		logger.Int64("member_id", int64(memberID)),
		logger.Int64("manager_id", int64(managerID)),
	)

	return nil
}

// DeleteFamily completely removes a family and its database (manager only)
func (s *Service) DeleteFamily(ctx context.Context, familyID, managerID int) error {
	// Verify the requester is the family manager
	if !s.isUserFamilyManager(ctx, familyID, managerID) {
		return ErrNotFamilyManager
	}

	// Delete the family database and master database records
	if err := s.dbManager.DeleteFamilyDatabase(ctx, familyID); err != nil {
		return fmt.Errorf("failed to delete family database: %w", err)
	}

	s.logger.Info("Family deleted successfully", logger.Int64("family_id", int64(familyID)), logger.Int64("manager_id", int64(managerID)))

	return nil
}

// RegenerateInviteCode generates a new invite code for a family (manager only)
func (s *Service) RegenerateInviteCode(ctx context.Context, familyID, managerID int) (string, error) {
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

	s.logger.Info("Family invite code regenerated", logger.Int64("family_id", int64(familyID)), logger.Str("new_invite_code", newInviteCode))

	return newInviteCode, nil
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

func (s *Service) getFamilyMembers(ctx context.Context, familyID int) ([]MemberResponse, error) {
	fID := int64(familyID)
	// Use SQLC to get family memberships
	sqlcMemberships, err := s.dbManager.GetMasterQueries().ListFamilyMemberships(ctx, &fID)
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
			s.logger.Warn("Failed to get user details for member", err, logger.Int64("user_id", *membership.UserID))
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

func (s *Service) isUserFamilyManager(ctx context.Context, familyID, userID int) bool {
	fID := int64(familyID)
	uID := int64(userID)
	getMembershipParams := masterdb.GetFamilyMembershipParams{
		FamilyID: &fID,
		UserID:   &uID,
	}

	membership, err := s.dbManager.GetMasterQueries().GetFamilyMembership(ctx, getMembershipParams)
	if err != nil {
		if err != sql.ErrNoRows {
			s.logger.Error("Failed to check if user is family manager", err, logger.Int64("family_id", int64(familyID)), logger.Int64("user_id", int64(userID)))
		}
		return false
	}

	return membership.Role == "manager"
}

func (s *Service) addMemberToFamilyDatabase(ctx context.Context, familyID, userID int, userName, userEmail, role string) error {
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

func (s *Service) removeMemberFromFamilyDatabase(ctx context.Context, familyID, userID int) error {
	familyDB, err := s.dbManager.GetFamilyDB(familyID)
	if err != nil {
		return fmt.Errorf("failed to get family database: %w", err)
	}

	query := `UPDATE family_members SET is_active = FALSE WHERE id = ?`
	_, err = familyDB.ExecContext(ctx, query, userID)
	return err
}

// Income management methods

// GetMonthlyIncomeInternal retrieves the family's monthly income
func (s *Service) getMonthlyIncomeInternal(ctx context.Context, familyID int) (*MonthlyIncome, error) {
	familyDB, err := s.dbManager.GetFamilyDB(familyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get family database: %w", err)
	}

	familyQueries := familydb.New(familyDB)

	setting, err := familyQueries.GetFamilySettingByKey(ctx, "monthly_income")
	if err != nil {
		if err == sql.ErrNoRows {
			// Return empty income if not set
			return &MonthlyIncome{
				TotalAmount: 0,
				Sources:     []IncomeSource{},
				UpdatedAt:   time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to get monthly income setting: %w", err)
	}

	var income MonthlyIncome
	if setting.SettingValue != nil {
		if err := json.Unmarshal([]byte(*setting.SettingValue), &income); err != nil {
			return nil, fmt.Errorf("failed to unmarshal income data: %w", err)
		}
	}

	return &income, nil
}

// setMonthlyIncomeInternal sets the family's monthly income
func (s *Service) setMonthlyIncomeInternal(ctx context.Context, familyID int, income *MonthlyIncome) error {
	if income == nil {
		return fmt.Errorf("income cannot be nil")
	}

	familyDB, err := s.dbManager.GetFamilyDB(familyID)
	if err != nil {
		return fmt.Errorf("failed to get family database: %w", err)
	}

	familyQueries := familydb.New(familyDB)

	// Calculate total amount from sources
	var totalAmount float64
	for _, source := range income.Sources {
		if source.IsActive {
			totalAmount += source.Amount
		}
	}
	income.TotalAmount = totalAmount
	income.UpdatedAt = time.Now()

	// Serialize income data
	incomeJSON, err := json.Marshal(income)
	if err != nil {
		return fmt.Errorf("failed to marshal income data: %w", err)
	}

	// Check if setting exists
	_, err = familyQueries.GetFamilySettingByKey(ctx, "monthly_income")
	if err != nil {
		if err == sql.ErrNoRows {
			// Create new setting
			incomeValue := string(incomeJSON)
			_, err = familyQueries.CreateFamilySetting(ctx, familydb.CreateFamilySettingParams{
				SettingKey:   "monthly_income",
				SettingValue: &incomeValue,
				DataType:     "json",
			})
			if err != nil {
				return fmt.Errorf("failed to create monthly income setting: %w", err)
			}
		} else {
			return fmt.Errorf("failed to check existing monthly income setting: %w", err)
		}
	} else {
		// Update existing setting
		setting, err := familyQueries.GetFamilySettingByKey(ctx, "monthly_income")
		if err != nil {
			return fmt.Errorf("failed to get existing setting: %w", err)
		}

		incomeValue := string(incomeJSON)
		_, err = familyQueries.UpdateFamilySetting(ctx, familydb.UpdateFamilySettingParams{
			ID:           setting.ID,
			SettingValue: &incomeValue,
			DataType:     "json",
		})
		if err != nil {
			return fmt.Errorf("failed to update monthly income setting: %w", err)
		}
	}

	s.logger.Info("Monthly income updated successfully", logger.Int64("family_id", int64(familyID)), logger.Str("total_amount", fmt.Sprintf("%.2f", totalAmount)))

	return nil
}

// addIncomeSourceInternal adds a new income source to the family
func (s *Service) addIncomeSourceInternal(ctx context.Context, familyID int, source IncomeSource) error {
	income, err := s.getMonthlyIncomeInternal(ctx, familyID)
	if err != nil {
		return fmt.Errorf("failed to get current income: %w", err)
	}

	income.Sources = append(income.Sources, source)

	return s.setMonthlyIncomeInternal(ctx, familyID, income)
}

// removeIncomeSourceInternal removes an income source from the family
func (s *Service) removeIncomeSourceInternal(ctx context.Context, familyID int, sourceName string) error {
	income, err := s.getMonthlyIncomeInternal(ctx, familyID)
	if err != nil {
		return fmt.Errorf("failed to get current income: %w", err)
	}

	// Find and remove the source
	for i, source := range income.Sources {
		if source.Name == sourceName {
			income.Sources = append(income.Sources[:i], income.Sources[i+1:]...)
			break
		}
	}

	return s.setMonthlyIncomeInternal(ctx, familyID, income)
}

// updateIncomeSourceInternal updates an existing income source
func (s *Service) updateIncomeSourceInternal(ctx context.Context, familyID int, sourceName string, updatedSource IncomeSource) error {
	income, err := s.getMonthlyIncomeInternal(ctx, familyID)
	if err != nil {
		return fmt.Errorf("failed to get current income: %w", err)
	}

	// Find and update the source
	for i, source := range income.Sources {
		if source.Name == sourceName {
			income.Sources[i] = updatedSource
			break
		}
	}

	return s.setMonthlyIncomeInternal(ctx, familyID, income)
}
