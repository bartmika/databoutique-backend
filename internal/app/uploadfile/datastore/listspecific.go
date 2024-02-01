package datastore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl UploadFileStorerImpl) ListByUploadDirectoryIDs(ctx context.Context, uploadDirectoryIDs []primitive.ObjectID) (*UploadFilePaginationListResult, error) {
	filter := bson.M{
		"upload_directory_id": bson.M{"$in": uploadDirectoryIDs},
	}

	cursor, err := impl.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var uploadFiles []*UploadFile
	for cursor.Next(ctx) {
		var uploadFile *UploadFile
		if err := cursor.Decode(&uploadFile); err != nil {
			return nil, err
		}
		uploadFiles = append(uploadFiles, uploadFile)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return &UploadFilePaginationListResult{
		Results:     uploadFiles,
		NextCursor:  "",
		HasNextPage: false,
	}, nil
}

func (impl UploadFileStorerImpl) ListByUploadDirectoryID(ctx context.Context, uploadDirectoryID primitive.ObjectID) (*UploadFilePaginationListResult, error) {
	return impl.ListByUploadDirectoryIDs(ctx, []primitive.ObjectID{uploadDirectoryID})
}
