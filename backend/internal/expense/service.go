package expense

import (
	"context"
	expensev1 "expenses-backend/pkg/expense/v1"
	"fmt"
	"slices"
	"sync"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	mu       sync.RWMutex
	expenses map[string]*expensev1.Expense
	nextID   int64
}

func NewService() *Service {
	return &Service{
		expenses: make(map[string]*expensev1.Expense),
		nextID:   1,
	}
}

func (s *Service) CreateExpense(ctx context.Context, req *connect.Request[expensev1.CreateExpenseRequest]) (*connect.Response[expensev1.CreateExpenseResponse], error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.Msg.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.Msg.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}
	if req.Msg.DayOfMonthDue < 1 || req.Msg.DayOfMonthDue > 31 {
		return nil, status.Error(codes.InvalidArgument, "day_of_month_due must be between 1 and 31")
	}

	id := fmt.Sprintf("%d", s.nextID)
	s.nextID++

	now := time.Now().Unix()
	expense := &expensev1.Expense{
		Id:            id,
		Name:          req.Msg.Name,
		Amount:        req.Msg.Amount,
		DayOfMonthDue: req.Msg.DayOfMonthDue,
		IsAutopay:     req.Msg.IsAutopay,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	s.expenses[id] = expense

	return connect.NewResponse(&expensev1.CreateExpenseResponse{
		Expense: expense,
	}), nil
}
func (s *Service) GetExpense(ctx context.Context, req *connect.Request[expensev1.GetExpenseRequest]) (*connect.Response[expensev1.GetExpenseResponse], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if req.Msg.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	expense, exists := s.expenses[req.Msg.Id]
	if !exists {
		return nil, status.Error(codes.NotFound, "expense not found")
	}

	return connect.NewResponse(&expensev1.GetExpenseResponse{
		Expense: expense,
	}), nil
}

func (s *Service) UpdateExpense(ctx context.Context, req *connect.Request[expensev1.UpdateExpenseRequest]) (*connect.Response[expensev1.UpdateExpenseResponse], error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.Msg.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	expense, exists := s.expenses[req.Msg.Id]
	if !exists {
		return nil, status.Error(codes.NotFound, "expense not found")
	}

	if req.Msg.Name != "" {
		expense.Name = req.Msg.Name
	}
	if req.Msg.Amount > 0 {
		expense.Amount = req.Msg.Amount
	}
	if req.Msg.DayOfMonthDue >= 1 && req.Msg.DayOfMonthDue <= 31 {
		expense.DayOfMonthDue = req.Msg.DayOfMonthDue
	}
	expense.IsAutopay = req.Msg.IsAutopay
	expense.UpdatedAt = time.Now().Unix()

	return connect.NewResponse(&expensev1.UpdateExpenseResponse{
		Expense: expense,
	}), nil
}

func (s *Service) DeleteExpense(ctx context.Context, req *connect.Request[expensev1.DeleteExpenseRequest]) (*connect.Response[expensev1.DeleteExpenseResponse], error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.Msg.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	_, exists := s.expenses[req.Msg.Id]
	if !exists {
		return nil, status.Error(codes.NotFound, "expense not found")
	}

	delete(s.expenses, req.Msg.Id)

	return connect.NewResponse(&expensev1.DeleteExpenseResponse{
		Success: true,
	}), nil
}

func (s *Service) ListExpenses(ctx context.Context, req *connect.Request[expensev1.ListExpensesRequest]) (*connect.Response[expensev1.ListExpensesResponse], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// populate sorted bills array by bill.Dayofmonthdue
	expensesByDayMap := make(map[int32][]*expensev1.Expense)
	for _, expense := range s.expenses {
		expensesByDayMap[expense.DayOfMonthDue] = append(expensesByDayMap[expense.DayOfMonthDue], expense)
	}

	// 2. Extract the unique days (keys of the map)
	var uniqueDays []int32
	for day := range expensesByDayMap {
		uniqueDays = append(uniqueDays, day)
	}

	// 3. Sort these unique days
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

	var sortedExpenses []*expensev1.SortedExpense

	for _, day := range uniqueDays {
		sortedExpenses = append(sortedExpenses, &expensev1.SortedExpense{
			Day:      day,
			Expenses: expensesByDayMap[day],
		})
	}

	return connect.NewResponse(&expensev1.ListExpensesResponse{
		Expenses:      sortedExpenses,
		NextPageToken: "",
	}), nil
}
