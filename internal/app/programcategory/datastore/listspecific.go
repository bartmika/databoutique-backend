package datastore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl ProgramCategoryStorerImpl) ListByTenantID(ctx context.Context, tid primitive.ObjectID) (*ProgramCategoryPaginationListResult, error) {
	f := &ProgramCategoryPaginationListFilter{
		Cursor:    "",
		PageSize:  1_000_000_000, // Unlimited
		SortField: "sort_number",
		SortOrder: 1,
		TenantID:  tid,
		Status:    ProgramCategoryStatusActive,
	}
	return impl.ListByFilter(ctx, f)
}
