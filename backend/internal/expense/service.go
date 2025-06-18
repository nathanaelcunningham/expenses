package expense

import (
	"context"
	"database/sql"
	"expenses-backend/internal/database"
	"expenses-backend/internal/database/sql/familydb"
	"expenses-backend/internal/logger"
	"expenses-backend/internal/middleware"
	expensev1 "expenses-backend/pkg/expense/v1"
	"slices"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Service handles expense operations using database queries
type Service struct {
	dbFactory *database.Factory
	logger    logger.Logger
}

// NewService creates a new expense service
func NewService(dbFactory *database.Factory, log logger.Logger) *Service {
	return &Service{
		dbFactory: dbFactory,
		logger:    log.With(logger.Str("component", "expense-service")),
	}
}

func (s *Service) CreateExpense(ctx context.Context, req *connect.Request[expensev1.CreateExpenseRequest]) (*connect.Response[expensev1.CreateExpenseResponse], error) {
	// Get authentication context
	authCtx, err := middleware.RequireFamily(ctx)
	if err != nil {
		return nil, err
	}

	// Validate input
	if req.Msg.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.Msg.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}
	if req.Msg.DayOfMonthDue < 1 || req.Msg.DayOfMonthDue > 31 {
		return nil, status.Error(codes.InvalidArgument, "day_of_month_due must be between 1 and 31")
	}

	now := time.Now()

	// Validate day of month
	day := max(min(req.Msg.DayOfMonthDue, 31), 1)

	createParams := familydb.CreateExpenseParams{
		CategoryID:    nil, // Set to nil for now, could be mapped from request
		Amount:        req.Msg.Amount,
		Name:          req.Msg.Name,
		DayOfMonthDue: int64(day),
		IsAutopay:     req.Msg.IsAutopay,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Get family database queries
	familyQueries, err := s.dbFactory.GetFamilyQueries(ctx, authCtx.FamilyID)
	if err != nil {
		s.logger.Error("Failed to get family database", err)
		return nil, status.Error(codes.Internal, "failed to access family database")
	}

	// Create expense using SQLC
	expenseResult, err := familyQueries.CreateExpense(ctx, createParams)
	if err != nil {
		s.logger.Error("Failed to create expense", err)
		return nil, status.Error(codes.Internal, "failed to create expense")
	}

	// Convert back to protobuf format
	pbExpense := s.convertToProtoExpense(expenseResult)

	s.logger.Info("Expense created successfully",
		logger.Str("expense_id", expenseResult.ID),
		logger.Str("user_id", authCtx.UserID),
		logger.Str("family_id", authCtx.FamilyID))

	return connect.NewResponse(&expensev1.CreateExpenseResponse{
		Expense: pbExpense,
	}), nil
}
func (s *Service) GetExpense(ctx context.Context, req *connect.Request[expensev1.GetExpenseRequest]) (*connect.Response[expensev1.GetExpenseResponse], error) {
	// Get authentication context
	authCtx, err := middleware.RequireFamily(ctx)
	if err != nil {
		return nil, err
	}

	if req.Msg.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Get family database queries
	familyQueries, err := s.dbFactory.GetFamilyQueries(ctx, authCtx.FamilyID)
	if err != nil {
		s.logger.Error("Failed to get family database", err)
		return nil, status.Error(codes.Internal, "failed to access family database")
	}

	// Get expense from database
	expenseResult, err := familyQueries.GetExpenseByID(ctx, req.Msg.Id)
	if err != nil {
		s.logger.Error("Failed to get expense", err,
			logger.Str("expense_id", req.Msg.Id))
		return nil, status.Error(codes.NotFound, "expense not found")
	}

	// Verify user has access to this expense (simplified for family-based access)
	canAccess, err := s.userCanAccessExpense(ctx, familyQueries, req.Msg.Id, authCtx.UserID)
	if err != nil {
		s.logger.Error("Failed to check expense access", err)
		return nil, status.Error(codes.Internal, "failed to verify access")
	}
	if !canAccess {
		return nil, status.Error(codes.PermissionDenied, "access denied to expense")
	}

	// Convert to protobuf format
	pbExpense := s.convertToProtoExpense(expenseResult)

	return connect.NewResponse(&expensev1.GetExpenseResponse{
		Expense: pbExpense,
	}), nil
}

func (s *Service) UpdateExpense(ctx context.Context, req *connect.Request[expensev1.UpdateExpenseRequest]) (*connect.Response[expensev1.UpdateExpenseResponse], error) {
	// Get authentication context
	authCtx, err := middleware.RequireFamily(ctx)
	if err != nil {
		return nil, err
	}

	if req.Msg.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Get family database queries
	familyQueries, err := s.dbFactory.GetFamilyQueries(ctx, authCtx.FamilyID)
	if err != nil {
		s.logger.Error("Failed to get family database", err)
		return nil, status.Error(codes.Internal, "failed to access family database")
	}

	// Verify user has access to this expense
	canAccess, err := s.userCanAccessExpense(ctx, familyQueries, req.Msg.Id, authCtx.UserID)
	if err != nil {
		s.logger.Error("Failed to check expense access", err)
		return nil, status.Error(codes.Internal, "failed to verify access")
	}
	if !canAccess {
		return nil, status.Error(codes.PermissionDenied, "access denied to expense")
	}

	// Get current expense to build update parameters
	current, err := familyQueries.GetExpenseByID(ctx, req.Msg.Id)
	if err != nil {
		s.logger.Error("Failed to get current expense", err,
			logger.Str("expense_id", req.Msg.Id))
		return nil, status.Error(codes.NotFound, "expense not found")
	}

	// Build update parameters
	updateParams := familydb.UpdateExpenseParams{
		ID:            req.Msg.Id,
		CategoryID:    current.CategoryID,
		Amount:        current.Amount,
		Name:          current.Name,
		DayOfMonthDue: current.DayOfMonthDue,
		IsAutopay:     current.IsAutopay,
		UpdatedAt:     time.Now(),
	}

	// Apply updates
	if req.Msg.Name != "" {
		updateParams.Name = req.Msg.Name
	}
	if req.Msg.Amount > 0 {
		updateParams.Amount = req.Msg.Amount
	}
	if req.Msg.DayOfMonthDue >= 1 && req.Msg.DayOfMonthDue <= 31 {
		updateParams.DayOfMonthDue = int64(req.Msg.DayOfMonthDue)
	}
	if req.Msg.IsAutopay != current.IsAutopay {
		updateParams.IsAutopay = req.Msg.IsAutopay
	}

	// Update expense using SQLC
	expenseResult, err := familyQueries.UpdateExpense(ctx, updateParams)
	if err != nil {
		s.logger.Error("Failed to update expense", err,
			logger.Str("expense_id", req.Msg.Id))
		return nil, status.Error(codes.Internal, "failed to update expense")
	}

	// Convert to protobuf format
	pbExpense := s.convertToProtoExpense(expenseResult)

	s.logger.Info("Expense updated successfully",
		logger.Str("expense_id", req.Msg.Id),
		logger.Str("user_id", authCtx.UserID))

	return connect.NewResponse(&expensev1.UpdateExpenseResponse{
		Expense: pbExpense,
	}), nil
}

func (s *Service) DeleteExpense(ctx context.Context, req *connect.Request[expensev1.DeleteExpenseRequest]) (*connect.Response[expensev1.DeleteExpenseResponse], error) {
	// Get authentication context
	authCtx, err := middleware.RequireFamily(ctx)
	if err != nil {
		return nil, err
	}

	if req.Msg.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// Get family database queries
	familyQueries, err := s.dbFactory.GetFamilyQueries(ctx, authCtx.FamilyID)
	if err != nil {
		s.logger.Error("Failed to get family database", err)
		return nil, status.Error(codes.Internal, "failed to access family database")
	}

	// Verify user has access to this expense
	canAccess, err := s.userCanAccessExpense(ctx, familyQueries, req.Msg.Id, authCtx.UserID)
	if err != nil {
		s.logger.Error("Failed to check expense access", err)
		return nil, status.Error(codes.Internal, "failed to verify access")
	}
	if !canAccess {
		return nil, status.Error(codes.PermissionDenied, "access denied to expense")
	}

	// Delete expense using SQLC
	err = familyQueries.DeleteExpense(ctx, req.Msg.Id)
	if err != nil {
		s.logger.Error("Failed to delete expense", err,
			logger.Str("expense_id", req.Msg.Id))
		return nil, status.Error(codes.Internal, "failed to delete expense")
	}

	s.logger.Info("Expense deleted successfully",
		logger.Str("expense_id", req.Msg.Id),
		logger.Str("user_id", authCtx.UserID))

	return connect.NewResponse(&expensev1.DeleteExpenseResponse{
		Success: true,
	}), nil
}

func (s *Service) ListExpenses(ctx context.Context, req *connect.Request[expensev1.ListExpensesRequest]) (*connect.Response[expensev1.ListExpensesResponse], error) {
	// Get authentication context
	authCtx, err := middleware.RequireFamily(ctx)
	if err != nil {
		return nil, err
	}

	// Set pagination parameters
	limit := int64(50) // default limit
	if req.Msg.PageSize > 0 {
		limit = int64(req.Msg.PageSize)
	}

	listParams := familydb.ListExpensesParams{
		Limit:  limit,
		Offset: 0, // TODO: Implement proper pagination with page tokens
	}

	// Get family database queries
	familyQueries, err := s.dbFactory.GetFamilyQueries(ctx, authCtx.FamilyID)
	if err != nil {
		s.logger.Error("Failed to get family database", err)
		return nil, status.Error(codes.Internal, "failed to access family database")
	}

	// Get expenses from database
	expenses, err := familyQueries.ListExpenses(ctx, listParams)
	if err != nil {
		s.logger.Error("Failed to list expenses", err)
		return nil, status.Error(codes.Internal, "failed to list expenses")
	}

	// Convert to protobuf format and group by day of month
	expensesByDayMap := make(map[int32][]*expensev1.Expense)
	for _, exp := range expenses {
		pbExpense := s.convertToProtoExpense(exp)
		expensesByDayMap[pbExpense.DayOfMonthDue] = append(expensesByDayMap[pbExpense.DayOfMonthDue], pbExpense)
	}

	// Extract and sort unique days
	var uniqueDays []int32
	for day := range expensesByDayMap {
		uniqueDays = append(uniqueDays, day)
	}

	slices.SortFunc(uniqueDays, func(a, b int32) int {
		if a == b {
			return 0
		}
		if a > b {
			return 1
		}
		if a < b {
			return -1
		}
		return 0
	})

	// Build sorted expenses response
	var sortedExpenses []*expensev1.SortedExpense
	for _, day := range uniqueDays {
		sortedExpenses = append(sortedExpenses, &expensev1.SortedExpense{
			Day:      day,
			Expenses: expensesByDayMap[day],
		})
	}

	s.logger.Info("Listed expenses successfully",
		logger.Int("expense_count", len(expenses)),
		logger.Str("user_id", authCtx.UserID))

	return connect.NewResponse(&expensev1.ListExpensesResponse{
		Expenses:      sortedExpenses,
		NextPageToken: "", // TODO: Implement proper pagination
	}), nil
}

// Helper methods

// convertToProtoExpense converts SQLC expense to protobuf expense
func (s *Service) convertToProtoExpense(exp *familydb.Expense) *expensev1.Expense {
	return &expensev1.Expense{
		Id:            exp.ID,
		Name:          exp.Name,
		Amount:        exp.Amount,
		DayOfMonthDue: int32(exp.DayOfMonthDue),
		IsAutopay:     exp.IsAutopay,
		CreatedAt:     exp.CreatedAt.Unix(),
		UpdatedAt:     exp.UpdatedAt.Unix(),
	}
}

// userCanAccessExpense checks if a user can access an expense (simplified for family-based access)
func (s *Service) userCanAccessExpense(ctx context.Context, queries *familydb.Queries, expenseID, userID string) (bool, error) {
	// For family databases, all family members can access all expenses
	// In a more complex system, you might check expense ownership or permissions
	_, err := queries.GetExpenseByID(ctx, expenseID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
