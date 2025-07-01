package family

import (
	"context"
	"database/sql"
	"errors"

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
