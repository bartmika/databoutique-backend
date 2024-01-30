package datastore

import (
	"context"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (impl TenantStorerImpl) GetByID(ctx context.Context, id primitive.ObjectID) (*Tenant, error) {
	filter := bson.M{"_id": id}

	var result Tenant
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

func (impl TenantStorerImpl) GetByPublicID(ctx context.Context, oldID uint64) (*Tenant, error) {
	filter := bson.M{"public_id": oldID}

	var result Tenant
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

func (impl TenantStorerImpl) GetBySchemaName(ctx context.Context, schemaName string) (*Tenant, error) {
	filter := bson.M{"schema_name": schemaName}

	var result Tenant
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

func (impl TenantStorerImpl) GetLatest(ctx context.Context) (*Tenant, error) {
	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{"public_id", -1}}).SetLimit(1)

	var order Tenant
	cursor, err := impl.Collection.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	if cursor.Next(context.Background()) {
		err := cursor.Decode(&order)
		if err != nil {
			return nil, err
		}
		return &order, nil
	}

	return nil, nil
}

func (impl TenantStorerImpl) GetOpenAICredentialsByID(ctx context.Context, id primitive.ObjectID) (*TenantOpenAICredentials, error) {
	filter := bson.M{"_id": id}

	// DEVELOPERS NOTE/TODO: IMPLEMENT SECURITY HERE FOR RETURNING UNENCRYPTED DATA.

	var result TenantOpenAICredentials
	err := impl.Collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// This error means your query did not match any documents.
			return nil, nil
		}
		impl.Logger.Error("database get openai credentials by id error", slog.Any("error", err))
		return nil, err
	}
	return &result, nil
}
