package family

import (
	"context"
	"database/sql"
	"errors"
	"time"

	appcontext "expenses-backend/internal/context"
	"expenses-backend/internal/database/sql/familydb"
	v1 "expenses-backend/pkg/family/v1"

	"connectrpc.com/connect"
)

func (s *Service) CreateFamilySetting(ctx context.Context, req *connect.Request[v1.CreateFamilySettingRequest]) (*connect.Response[v1.CreateFamilySettingResponse], error) {
	authCtx, err := appcontext.RequireFamily(ctx)
	if err != nil {
		return nil, err
	}

	queries, err := s.dbManager.GetFamilyQueries(int(authCtx.FamilyID))
	if err != nil {
		return nil, err
	}

	setting, err := queries.CreateFamilySetting(ctx, familydb.CreateFamilySettingParams{
		SettingKey:   req.Msg.SettingKey,
		SettingValue: req.Msg.SettingValue,
		DataType:     req.Msg.DataType,
	})

	return connect.NewResponse(&v1.CreateFamilySettingResponse{
		FamilySetting: &v1.FamilySetting{
			Id:           setting.ID,
			SettingKey:   setting.SettingKey,
			SettingValue: setting.SettingValue,
			DataType:     setting.DataType,
		},
	}), nil
}

func (s *Service) ListFamilySettings(ctx context.Context, req *connect.Request[v1.ListFamilySettingsRequest]) (*connect.Response[v1.ListFamilySettingsResponse], error) {
	authCtx, err := appcontext.RequireFamily(ctx)
	if err != nil {
		return nil, err
	}

	queries, err := s.dbManager.GetFamilyQueries(int(authCtx.FamilyID))
	if err != nil {
		return nil, err
	}

	settings, err := queries.ListFamilySettings(ctx)
	if err != nil {
		return nil, err
	}

	resp := []*v1.FamilySetting{}

	for _, s := range settings {
		resp = append(resp, &v1.FamilySetting{
			Id:           s.ID,
			SettingKey:   s.SettingKey,
			SettingValue: s.SettingValue,
			DataType:     s.DataType,
		})
	}

	return connect.NewResponse(&v1.ListFamilySettingsResponse{
		FamilySettings: resp,
	}), nil
}

func (s *Service) GetFamilySettingByKey(ctx context.Context, req *connect.Request[v1.GetFamilySettingByKeyRequest]) (*connect.Response[v1.GetFamilySettingByKeyResponse], error) {
	authCtx, err := appcontext.RequireFamily(ctx)
	if err != nil {
		return nil, err
	}

	queries, err := s.dbManager.GetFamilyQueries(int(authCtx.FamilyID))
	if err != nil {
		return nil, err
	}

	setting, err := queries.GetFamilySettingByKey(ctx, req.Msg.Key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return connect.NewResponse(&v1.GetFamilySettingByKeyResponse{}), nil
		}

		return nil, err
	}

	return connect.NewResponse(&v1.GetFamilySettingByKeyResponse{
		FamilySetting: &v1.FamilySetting{
			Id:           setting.ID,
			SettingKey:   setting.SettingKey,
			SettingValue: setting.SettingValue,
			DataType:     setting.DataType,
		},
	}), nil
}

func (s *Service) UpdateFamilySetting(ctx context.Context, req *connect.Request[v1.UpdateFamilySettingRequest]) (*connect.Response[v1.UpdateFamilySettingResponse], error) {
	authCtx, err := appcontext.RequireFamily(ctx)
	if err != nil {
		return nil, err
	}

	queries, err := s.dbManager.GetFamilyQueries(int(authCtx.FamilyID))
	if err != nil {
		return nil, err
	}

	setting, err := queries.UpdateFamilySetting(ctx, familydb.UpdateFamilySettingParams{
		SettingValue: req.Msg.SettingValue,
		DataType:     req.Msg.DataType,
		ID:           req.Msg.Id,
	})
	if err != nil {
		return nil, err
	}

	return &connect.Response[v1.UpdateFamilySettingResponse]{
		Msg: &v1.UpdateFamilySettingResponse{
			FamilySetting: &v1.FamilySetting{
				Id:           setting.ID,
				SettingKey:   setting.SettingKey,
				SettingValue: setting.SettingValue,
				DataType:     setting.DataType,
			},
		},
	}, nil
}

func (s *Service) DeleteFamilySetting(ctx context.Context, req *connect.Request[v1.DeleteFamilySettingRequest]) (*connect.Response[v1.DeleteFamilySettingResponse], error) {
	return nil, nil
}

// Income management gRPC endpoints

func (s *Service) GetMonthlyIncome(ctx context.Context, req *connect.Request[v1.GetMonthlyIncomeRequest]) (*connect.Response[v1.GetMonthlyIncomeResponse], error) {
	authCtx, err := appcontext.RequireFamily(ctx)
	if err != nil {
		return nil, err
	}

	income, err := s.GetMonthlyIncomeInternal(ctx, int(authCtx.FamilyID))
	if err != nil {
		return nil, err
	}

	// Convert to proto format
	protoSources := make([]*v1.IncomeSource, len(income.Sources))
	for i, source := range income.Sources {
		protoSources[i] = &v1.IncomeSource{
			Name:        source.Name,
			Amount:      source.Amount,
			Description: source.Description,
			IsActive:    source.IsActive,
		}
	}

	return connect.NewResponse(&v1.GetMonthlyIncomeResponse{
		MonthlyIncome: &v1.MonthlyIncome{
			TotalAmount: income.TotalAmount,
			Sources:     protoSources,
			UpdatedAt:   income.UpdatedAt.Unix(),
		},
	}), nil
}

func (s *Service) SetMonthlyIncome(ctx context.Context, req *connect.Request[v1.SetMonthlyIncomeRequest]) (*connect.Response[v1.SetMonthlyIncomeResponse], error) {
	authCtx, err := appcontext.RequireFamily(ctx)
	if err != nil {
		return nil, err
	}

	// Convert from proto format
	sources := make([]IncomeSource, len(req.Msg.MonthlyIncome.Sources))
	for i, protoSource := range req.Msg.MonthlyIncome.Sources {
		sources[i] = IncomeSource{
			Name:        protoSource.Name,
			Amount:      protoSource.Amount,
			Description: protoSource.Description,
			IsActive:    protoSource.IsActive,
		}
	}

	income := &MonthlyIncome{
		TotalAmount: req.Msg.MonthlyIncome.TotalAmount,
		Sources:     sources,
		UpdatedAt:   time.Unix(req.Msg.MonthlyIncome.UpdatedAt, 0),
	}

	err = s.setMonthlyIncomeInternal(ctx, int(authCtx.FamilyID), income)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&v1.SetMonthlyIncomeResponse{
		Success: true,
	}), nil
}

func (s *Service) AddIncomeSource(ctx context.Context, req *connect.Request[v1.AddIncomeSourceRequest]) (*connect.Response[v1.AddIncomeSourceResponse], error) {
	authCtx, err := appcontext.RequireFamily(ctx)
	if err != nil {
		return nil, err
	}

	source := IncomeSource{
		Name:        req.Msg.IncomeSource.Name,
		Amount:      req.Msg.IncomeSource.Amount,
		Description: req.Msg.IncomeSource.Description,
		IsActive:    req.Msg.IncomeSource.IsActive,
	}

	err = s.addIncomeSourceInternal(ctx, int(authCtx.FamilyID), source)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&v1.AddIncomeSourceResponse{
		Success: true,
	}), nil
}

func (s *Service) RemoveIncomeSource(ctx context.Context, req *connect.Request[v1.RemoveIncomeSourceRequest]) (*connect.Response[v1.RemoveIncomeSourceResponse], error) {
	authCtx, err := appcontext.RequireFamily(ctx)
	if err != nil {
		return nil, err
	}

	err = s.removeIncomeSourceInternal(ctx, int(authCtx.FamilyID), req.Msg.SourceName)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&v1.RemoveIncomeSourceResponse{
		Success: true,
	}), nil
}

func (s *Service) UpdateIncomeSource(ctx context.Context, req *connect.Request[v1.UpdateIncomeSourceRequest]) (*connect.Response[v1.UpdateIncomeSourceResponse], error) {
	authCtx, err := appcontext.RequireFamily(ctx)
	if err != nil {
		return nil, err
	}

	updatedSource := IncomeSource{
		Name:        req.Msg.UpdatedSource.Name,
		Amount:      req.Msg.UpdatedSource.Amount,
		Description: req.Msg.UpdatedSource.Description,
		IsActive:    req.Msg.UpdatedSource.IsActive,
	}

	err = s.updateIncomeSourceInternal(ctx, int(authCtx.FamilyID), req.Msg.SourceName, updatedSource)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&v1.UpdateIncomeSourceResponse{
		Success: true,
	}), nil
}
