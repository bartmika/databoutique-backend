package datastore

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (impl UploadFileStorerImpl) GetByID(ctx context.Context, id primitive.ObjectID) (*UploadFile, error) {
	filter := bson.M{"_id": id}

	var result UploadFile
	err := impl.Collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// This error means your query did not match any documents.
			return nil, nil
		}
		impl.Logger.Error("database get by id error", slog.Any("error", err))
		return nil, err
	}
	return &result, nil
}

func (impl UploadFileStorerImpl) GetOpenAIFileIDsInUploadDirectoryIDs(ctx context.Context, uploadDirectoryIDs []primitive.ObjectID) ([]string, error) {
	filter := bson.M{
		"upload_directory_id": bson.M{"$in": uploadDirectoryIDs},
	}

	cursor, err := impl.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var openAIFileIDs []string
	for cursor.Next(ctx) {
		var uploadFile UploadFile
		if err := cursor.Decode(&uploadFile); err != nil {
			return nil, err
		}
		openAIFileIDs = append(openAIFileIDs, uploadFile.OpenAIFileID)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return openAIFileIDs, nil
}
