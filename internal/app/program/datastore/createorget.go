package datastore

import (
	"context"
)

func (impl ProgramStorerImpl) CreateOrGetByID(ctx context.Context, hh *Program) (*Program, error) {
	res, err := impl.GetByID(ctx, hh.ID)
	if err != nil {
		return nil, err
	}
	if res != nil {
		return res, nil
	}
	if err := impl.Create(ctx, hh); err != nil {
		return nil, err
	}
	return impl.GetByID(ctx, hh.ID)
}
