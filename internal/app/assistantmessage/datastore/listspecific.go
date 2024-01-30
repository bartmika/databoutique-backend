package datastore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl AssistantMessageStorerImpl) ListByTenantID(ctx context.Context, tid primitive.ObjectID) (*AssistantMessagePaginationListResult, error) {
	f := &AssistantMessagePaginationListFilter{
		Cursor:    "",
		PageSize:  1_000_000_000, // Unlimited
		SortField: "sort_number",
		SortOrder: 1,
		TenantID:  tid,
		Status:    AssistantMessageStatusActive,
	}
	return impl.ListByFilter(ctx, f)
}
