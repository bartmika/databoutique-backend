package datastore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (impl UserStorerImpl) UpsertByID(ctx context.Context, user *User) error {
	opts := options.Update().SetUpsert(true) // Use upsert option

	filter := bson.M{"_id": user.ID}

	update := bson.M{"$set": user}

	_, err := impl.Collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}

	return nil
}

func (impl UserStorerImpl) UpsertByEmail(ctx context.Context, user *User) error {
	opts := options.Update().SetUpsert(true) // Use upsert option

	filter := bson.M{"email": user.Email}

	update := bson.M{"$set": user}

	_, err := impl.Collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}

	return nil
}
