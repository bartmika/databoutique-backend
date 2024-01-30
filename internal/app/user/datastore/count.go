package datastore

import (
	"context"
	"time"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
)

func (impl UserStorerImpl) CountByFilter(ctx context.Context, f *UserListFilter) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	// Create the filter based on the cursor
	filter := bson.M{}

	// Add filter conditions to the filter
	if !f.TenantID.IsZero() {
		filter["tenant_id"] = f.TenantID
	}
	if f.Role > 0 {
		filter["role"] = f.Role
	}
	if f.Status != 0 {
		filter["status"] = f.Status
	}

	impl.Logger.Debug("counting w/ filter:",
		slog.Any("filter", filter))

	// Use the CountDocuments method to count the matching documents.
	count, err := impl.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	return count, nil
}
