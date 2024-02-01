package datastore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl UploadDirectoryStorerImpl) ListByTenantID(ctx context.Context, tid primitive.ObjectID) (*UploadDirectoryPaginationListResult, error) {
	f := &UploadDirectoryPaginationListFilter{
		Cursor:    "",
		PageSize:  1_000_000_000, // Unlimited
		SortField: "sort_number",
		SortOrder: 1,
		TenantID:  tid,
		Status:    UploadDirectoryStatusActive,
	}
	return impl.ListByFilter(ctx, f)
}

func (impl UploadDirectoryStorerImpl) ListByIDs(ctx context.Context, ids []*primitive.ObjectID) (*UploadDirectoryPaginationListResult, error) {
	f := &UploadDirectoryPaginationListFilter{
		Cursor:    "",
		PageSize:  1_000_000_000, // Unlimited
		SortField: "sort_number",
		SortOrder: 1,
		// TenantID:  tid,
		Status: UploadDirectoryStatusActive,
	}
	//TODO: Impl. listing by IDs.
	return impl.ListByFilter(ctx, f)
}
