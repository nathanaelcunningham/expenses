package expense

import (
	"context"
	"expenses-backend/internal/database/sql/familydb"
	"expenses-backend/internal/middleware"
	expensev1 "expenses-backend/pkg/expense/v1"
	"slices"
	"time"

	"connectrpc.com/connect"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Service handles expense operations using repository pattern
type Service struct {
	logger zerolog.Logger
}

// NewService creates a new expense service
func NewService(logger zerolog.Logger) *Service {
	return &Service{
		logger: logger.With().Str("component", "expense-service").Logger(),
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

	createReq := &familydb.CreateExpenseParams{
		Name:          req.Msg.Name,
		CategoryID:    new(string),
		Amount:        req.Msg.Amount,
		DayOfMonthDue: int64(req.Msg.DayOfMonthDue),
		IsAutopay:     req.Msg.IsAutopay,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Get expense store for the family
	expenseStore, err := s.repoFactory.NewExpenseStore(ctx, authCtx.FamilyID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get expense store")
		return nil, status.Error(codes.Internal, "failed to access family database")
	}

	// Create expense using repository
	expenseResult, err := expenseStore.CreateExpense(ctx, createReq)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create expense")
		return nil, status.Error(codes.Internal, "failed to create expense")
	}

	// Convert back to protobuf format
	pbExpense := s.convertToProtoExpense(expenseResult)

	s.logger.Info().
		Str("expense_id", expenseResult.ID).
		Str("user_id", authCtx.UserID).
		Str("family_id", authCtx.FamilyID).
		Msg("Expense created successfully")

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

	// Get expense store for the family
	expenseStore, err := s.repoFactory.NewExpenseStore(ctx, authCtx.FamilyID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get expense store")
		return nil, status.Error(codes.Internal, "failed to access family database")
	}

	// Get expense from repository
	expenseResult, err := expenseStore.GetExpenseByID(ctx, req.Msg.Id)
	if err != nil {
		s.logger.Error().Err(err).Str("expense_id", req.Msg.Id).Msg("Failed to get expense")
		return nil, status.Error(codes.NotFound, "expense not found")
	}

	// Verify user has access to this expense
	canAccess, err := expenseStore.UserCanAccessExpense(ctx, req.Msg.Id, authCtx.UserID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to check expense access")
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

	// Get expense store for the family
	expenseStore, err := s.repoFactory.NewExpenseStore(ctx, authCtx.FamilyID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get expense store")
		return nil, status.Error(codes.Internal, "failed to access family database")
	}

	// Verify user has access to this expense
	canAccess, err := expenseStore.UserCanAccessExpense(ctx, req.Msg.Id, authCtx.UserID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to check expense access")
		return nil, status.Error(codes.Internal, "failed to verify access")
	}
	if !canAccess {
		return nil, status.Error(codes.PermissionDenied, "access denied to expense")
	}

	// Build update request
	updateReq := &models.UpdateExpenseRequest{}

	if req.Msg.Name != "" {
		updateReq.Description = &req.Msg.Name
	}
	if req.Msg.Amount > 0 {
		updateReq.Amount = &req.Msg.Amount
	}
	if req.Msg.DayOfMonthDue >= 1 && req.Msg.DayOfMonthDue <= 31 {
		newDate := s.calculateDueDate(req.Msg.DayOfMonthDue)
		updateReq.Date = &newDate
	}
	// Note: IsAutopay doesn't have a direct mapping in our expense model

	// Update expense using repository
	expenseResult, err := expenseStore.UpdateExpense(ctx, req.Msg.Id, updateReq)
	if err != nil {
		s.logger.Error().Err(err).Str("expense_id", req.Msg.Id).Msg("Failed to update expense")
		return nil, status.Error(codes.Internal, "failed to update expense")
	}

	// Convert to protobuf format
	pbExpense := s.convertToProtoExpense(expenseResult)

	s.logger.Info().
		Str("expense_id", req.Msg.Id).
		Str("user_id", authCtx.UserID).
		Msg("Expense updated successfully")

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

	// Get expense store for the family
	expenseStore, err := s.repoFactory.NewExpenseStore(ctx, authCtx.FamilyID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get expense store")
		return nil, status.Error(codes.Internal, "failed to access family database")
	}

	// Verify user has access to this expense
	canAccess, err := expenseStore.UserCanAccessExpense(ctx, req.Msg.Id, authCtx.UserID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to check expense access")
		return nil, status.Error(codes.Internal, "failed to verify access")
	}
	if !canAccess {
		return nil, status.Error(codes.PermissionDenied, "access denied to expense")
	}

	// Delete expense using repository
	err = expenseStore.DeleteExpense(ctx, req.Msg.Id)
	if err != nil {
		s.logger.Error().Err(err).Str("expense_id", req.Msg.Id).Msg("Failed to delete expense")
		return nil, status.Error(codes.Internal, "failed to delete expense")
	}

	s.logger.Info().
		Str("expense_id", req.Msg.Id).
		Str("user_id", authCtx.UserID).
		Msg("Expense deleted successfully")

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

	// Create filter for user's expenses
	filter := &models.ExpenseFilter{
		MemberID: &authCtx.UserID,
	}

	// Set pagination if provided
	if req.Msg.PageSize > 0 {
		filter.Limit = int(req.Msg.PageSize)
	}

	// Get expense store for the family
	expenseStore, err := s.repoFactory.NewExpenseStore(ctx, authCtx.FamilyID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get expense store")
		return nil, status.Error(codes.Internal, "failed to access family database")
	}

	// Get expenses from repository
	expenses, err := expenseStore.ListExpenses(ctx, filter)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to list expenses")
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

	s.logger.Info().
		Int("expense_count", len(expenses)).
		Str("user_id", authCtx.UserID).
		Msg("Listed expenses successfully")

	return connect.NewResponse(&expensev1.ListExpensesResponse{
		Expenses:      sortedExpenses,
		NextPageToken: "", // TODO: Implement proper pagination
	}), nil
}

// Helper methods

// convertToProtoExpense converts repository expense model to protobuf expense
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
