package transaction

import (
	"context"
	"expenses-backend/internal/database"
	"expenses-backend/internal/database/sql/familydb"
	"expenses-backend/internal/logger"
	"expenses-backend/internal/simplefin"
	"fmt"
	"strings"
	"sync"

	appcontext "expenses-backend/internal/context"
	v1 "expenses-backend/pkg/transaction/v1"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	dbManager *database.DatabaseManager
	logger    logger.Logger
	mu        sync.RWMutex
	fin       map[int64]*simplefin.Client
}

func NewService(dbManager *database.DatabaseManager, log logger.Logger) *Service {
	return &Service{
		dbManager: dbManager,
		logger:    log,
		fin:       make(map[int64]*simplefin.Client),
	}
}
func (s *Service) GetSimplefinAccounts(ctx context.Context, req *connect.Request[v1.GetSimplefinAccountsRequest]) (*connect.Response[v1.GetSimplefinAccountsResponse], error) {
	authCtx, err := appcontext.RequireFamily(ctx)
	if err != nil {
		return nil, err
	}

	queries, err := s.dbManager.GetFamilyQueries(int(authCtx.FamilyID))
	if err != nil {
		return nil, err
	}

	sfc, err := s.getSimplefinClient(ctx, authCtx.FamilyID)
	if err != nil {
		return nil, err
	}

	savedAccounts, err := queries.GetAccounts(ctx)
	if err != nil {
		return nil, err
	}

	accounts, err := sfc.Accounts(ctx)
	if err != nil {
		s.logger.Error("failed to load accounts from simplefin", err)
		return nil, err
	}

	if len(accounts.Errors) > 0 {
		s.logger.Error("failed to load accounts from simplefin", err)
		return nil, fmt.Errorf("failed to load accounts from simplefin %s", strings.Join(accounts.Errors, "\n"))
	}

	sa := make(map[string]bool)

	for _, a := range savedAccounts {
		sa[a.AccountID] = true
	}

	resp := v1.GetSimplefinAccountsResponse{}

	for _, a := range accounts.Accounts {
		if _, ok := sa[a.ID]; !ok {
			resp.Accounts = append(resp.Accounts, &v1.SimplefinAccount{
				Id: a.ID,
				Org: &v1.Organization{
					Domain: a.Org.Domain,
					Name:   a.Org.Name,
				},
				Name:             a.Name,
				Currency:         a.Currency,
				Balance:          a.Balance,
				AvailableBalance: &a.AvailableBalance,
				BalanceDate:      &timestamppb.Timestamp{},
				Transactions:     []*v1.Transaction{},
			})
		}
	}

	return connect.NewResponse(&resp), nil
}

func (s *Service) GetAccounts(ctx context.Context, req *connect.Request[v1.GetAccountsRequest]) (*connect.Response[v1.GetAccountsResponse], error) {
	authCtx, err := appcontext.RequireFamily(ctx)
	if err != nil {
		return nil, err
	}

	queries, err := s.dbManager.GetFamilyQueries(int(authCtx.FamilyID))
	if err != nil {
		return nil, err
	}

	savedAccounts, err := queries.GetAccounts(ctx)
	if err != nil {
		return nil, err
	}

	resp := v1.GetAccountsResponse{}

	sa := make(map[string]bool)

	for _, a := range savedAccounts {
		resp.Accounts = append(resp.Accounts, &v1.Account{
			Id:        a.ID,
			Name:      a.Name,
			AccountId: a.AccountID,
		})
		sa[a.AccountID] = true
	}

	return connect.NewResponse(&resp), nil
}

func (s *Service) AddAccount(ctx context.Context, req *connect.Request[v1.AddAccountRequest]) (*connect.Response[v1.AddAccountResponse], error) {
	authCtx, err := appcontext.RequireFamily(ctx)
	if err != nil {
		return nil, err
	}

	queries, err := s.dbManager.GetFamilyQueries(int(authCtx.FamilyID))
	if err != nil {
		return nil, err
	}

	account, err := queries.CreateAccount(ctx, familydb.CreateAccountParams{
		AccountID: req.Msg.AccountId,
		Name:      req.Msg.Name,
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&v1.AddAccountResponse{
		Account: &v1.Account{
			Id:        account.ID,
			Name:      account.Name,
			AccountId: account.AccountID,
		},
	}), nil

}
