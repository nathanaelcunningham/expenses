package transaction

import (
	"context"
	"expenses-backend/internal/database"
	"expenses-backend/internal/logger"
	"expenses-backend/internal/simplefin"
	"fmt"
	"strings"

	v1 "expenses-backend/pkg/transaction/v1"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	dbManager *database.DatabaseManager
	logger    logger.Logger
	fin       *simplefin.Client
}

func NewService(dbManager *database.DatabaseManager, log logger.Logger, fin *simplefin.Client) *Service {
	return &Service{
		dbManager: dbManager,
		logger:    log,
		fin:       fin,
	}
}

func (s *Service) GetAccounts(ctx context.Context, req *connect.Request[v1.GetAccountsRequest]) (*connect.Response[v1.GetAccountsResponse], error) {
	accounts, err := s.fin.Accounts(ctx)
	if err != nil {
		s.logger.Error("failed to load accounts from simplefin", err)
		return nil, err
	}

	if len(accounts.Errors) > 0 {
		s.logger.Error("failed to load accounts from simplefin", err)
		return nil, fmt.Errorf("failed to load accounts from simplefin %s", strings.Join(accounts.Errors, "\n"))
	}

	resp := v1.GetAccountsResponse{}

	for _, a := range accounts.Accounts {
		resp.Accounts = append(resp.Accounts, &v1.Account{
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

	return connect.NewResponse(&resp), nil
}
