package datastore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl ProgramStorerImpl) ListByTenantID(ctx context.Context, tid primitive.ObjectID) (*ProgramPaginationListResult, error) {
	f := &ProgramPaginationListFilter{
		Cursor:    "",
		PageSize:  1_000_000_000, // Unlimited
		SortField: "sort_number",
		SortOrder: 1,
		TenantID:  tid,
		Status:    ProgramStatusActive,
	}
	return impl.ListByFilter(ctx, f)
}
