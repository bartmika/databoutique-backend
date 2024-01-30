package controller

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"

	uploaddirectory_s "github.com/bartmika/databoutique-backend/internal/app/uploaddirectory/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func (impl *UploadDirectoryControllerImpl) ArchiveByID(ctx context.Context, id primitive.ObjectID) (*uploaddirectory_s.UploadDirectory, error) {
	// // Extract from our session the following data.
	// userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)

	// Lookup the uploaddirectory in our database, else return a `400 Bad Request` error.
	ou, err := impl.UploadDirectoryStorer.GetByID(ctx, id)
	if err != nil {
		impl.Logger.Error("database error", slog.Any("err", err))
		return nil, err
	}
	if ou == nil {
		impl.Logger.Warn("uploaddirectory does not exist validation error")
		return nil, httperror.NewForBadRequestWithSingleField("id", "does not exist")
	}

	ou.Status = uploaddirectory_s.UploadDirectoryStatusArchived

	if err := impl.UploadDirectoryStorer.UpdateByID(ctx, ou); err != nil {
		impl.Logger.Error("uploaddirectory update by id error", slog.Any("error", err))
		return nil, err
	}
	return ou, nil
}
