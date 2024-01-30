package datastore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl HowHearAboutUsItemStorerImpl) ListByTenantID(ctx context.Context, tid primitive.ObjectID) (*HowHearAboutUsItemPaginationListResult, error) {
	f := &HowHearAboutUsItemPaginationListFilter{
		Cursor:    "",
		PageSize:  1_000_000_000, // Unlimited
		SortField: "sort_number",
		SortOrder: 1,
		TenantID:  tid,
		Status:    HowHearAboutUsItemStatusActive,
	}
	return impl.ListByFilter(ctx, f)
}
