package datastore

import (
	"context"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
)

func (impl AssistantStorerImpl) CheckIfExistsByEmail(ctx context.Context, email string) (bool, error) {
	filter := bson.D{{"email", email}}
	count, err := impl.Collection.CountDocuments(ctx, filter)
	if err != nil {
		impl.Logger.Error("database check if exists by email error", slog.Any("error", err))
		return false, err
	}
	return count >= 1, nil
}
