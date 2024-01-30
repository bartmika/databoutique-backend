package datastore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl ExecutableStorerImpl) ListByTenantID(ctx context.Context, tid primitive.ObjectID) (*ExecutablePaginationListResult, error) {
	f := &ExecutablePaginationListFilter{
		Cursor:    "",
		PageSize:  1_000_000_000, // Unlimited
		SortField: "sort_number",
		SortOrder: 1,
		TenantID:  tid,
		Status:    ExecutableStatusActive,
	}
	return impl.ListByFilter(ctx, f)
}
