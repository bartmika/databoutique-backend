package controller

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"

	folderinfo_s "github.com/bartmika/databoutique-backend/internal/app/folderinfo/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func (impl *FolderInfoControllerImpl) ArchiveByID(ctx context.Context, id primitive.ObjectID) (*folderinfo_s.FolderInfo, error) {
	// // Extract from our session the following data.
	// userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)

	// Lookup the folderinfo in our database, else return a `400 Bad Request` error.
	ou, err := impl.FolderInfoStorer.GetByID(ctx, id)
	if err != nil {
		impl.Logger.Error("database error", slog.Any("err", err))
		return nil, err
	}
	if ou == nil {
		impl.Logger.Warn("folderinfo does not exist validation error")
		return nil, httperror.NewForBadRequestWithSingleField("id", "does not exist")
	}

	ou.Status = folderinfo_s.FolderInfoStatusArchived

	if err := impl.FolderInfoStorer.UpdateByID(ctx, ou); err != nil {
		impl.Logger.Error("folderinfo update by id error", slog.Any("error", err))
		return nil, err
	}
	return ou, nil
}
