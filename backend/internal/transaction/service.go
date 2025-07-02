package transaction

import (
	"context"
	"errors"
	"expenses-backend/internal/simplefin"
)

func (s *Service) getSimplefinClient(ctx context.Context, familyID int64) (*simplefin.Client, error) {
	if client, ok := s.fin[familyID]; ok {
		return client, nil
	}

	queries, err := s.dbManager.GetFamilyQueries(int(familyID))
	if err != nil {
		return nil, err
	}

	res, err := queries.GetFamilySettingByKey(ctx, "simplefin_token")
	if err != nil {
		return nil, err
	}
	if res.SettingValue == nil || *res.SettingValue == "" {
		return nil, errors.New("simplefin_token not set")
	}

	c, err := simplefin.NewClient(*res.SettingValue)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()

	s.fin[familyID] = c

	s.mu.Unlock()

	return c, nil
}
