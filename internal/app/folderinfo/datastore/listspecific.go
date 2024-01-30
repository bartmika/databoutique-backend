package datastore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl FolderInfoStorerImpl) ListByTenantID(ctx context.Context, tid primitive.ObjectID) (*FolderInfoPaginationListResult, error) {
	f := &FolderInfoPaginationListFilter{
		Cursor:    "",
		PageSize:  1_000_000_000, // Unlimited
		SortField: "sort_number",
		SortOrder: 1,
		TenantID:  tid,
		Status:    FolderInfoStatusActive,
	}
	return impl.ListByFilter(ctx, f)
}
