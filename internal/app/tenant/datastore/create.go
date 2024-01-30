package datastore

import (
	"context"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl TenantStorerImpl) Create(ctx context.Context, u *Tenant) error {
	// DEVELOPER NOTES:
	// According to mongodb documentaiton:
	//     Non-existent Databases and Collections
	//     If the necessary database and collection don't exist when you perform a write operation, the server implicitly creates them.
	//     Source: https://www.mongodb.com/docs/drivers/go/current/usage-examples/insertOne/

	if u.ID == primitive.NilObjectID {
		u.ID = primitive.NewObjectID()
		impl.Logger.Warn("database insert tenant not included id value, created id now.", slog.Any("id", u.ID))
	}

	// If `public_is` not explicitly set then we implicitly set it.
	if u.PublicID == 0 {
		publicID, err := impl.generatePublicID(ctx)
		if err != nil {
			return err
		}
		u.PublicID = publicID
	}

	_, err := impl.Collection.InsertOne(ctx, u)

	// check for errors in the insertion
	if err != nil {
		impl.Logger.Error("database insert error", slog.Any("error", err))
	}

	return nil
}

func (impl TenantStorerImpl) generatePublicID(ctx context.Context) (uint64, error) {
	var publicID uint64
	latest, err := impl.GetLatest(ctx)
	if err != nil {
		impl.Logger.Error("database get latest tenant by tenant id error",
			slog.Any("error", err))
		return 0, err
	}
	if latest == nil {
		impl.Logger.Debug("first tenant creation detected, setting publicID to value of 1")
		publicID = 1
	} else {
		publicID = latest.PublicID + 1
		impl.Logger.Debug("system generated new tenant publicID")
	}
	return publicID, nil
}
