package datastore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
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

func (impl UploadDirectoryStorerImpl) ListByIDs(ctx context.Context, ids []primitive.ObjectID) (*UploadDirectoryPaginationListResult, error) {
	filter := bson.M{
		"_id": bson.M{"$in": ids},
	}

	cursor, err := impl.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var uploadFiles []*UploadDirectory
	for cursor.Next(ctx) {
		var uploadFile *UploadDirectory
		if err := cursor.Decode(&uploadFile); err != nil {
			return nil, err
		}
		uploadFiles = append(uploadFiles, uploadFile)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return &UploadDirectoryPaginationListResult{
		Results:     uploadFiles,
		NextCursor:  "",
		HasNextPage: false,
	}, nil
}
