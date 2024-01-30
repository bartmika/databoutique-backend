package controller

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"

	programcategory_s "github.com/bartmika/databoutique-backend/internal/app/programcategory/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func (impl *ProgramCategoryControllerImpl) ArchiveByID(ctx context.Context, id primitive.ObjectID) (*programcategory_s.ProgramCategory, error) {
	// // Extract from our session the following data.
	// userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)

	// Lookup the programcategory in our database, else return a `400 Bad Request` error.
	ou, err := impl.ProgramCategoryStorer.GetByID(ctx, id)
	if err != nil {
		impl.Logger.Error("database error", slog.Any("err", err))
		return nil, err
	}
	if ou == nil {
		impl.Logger.Warn("programcategory does not exist validation error")
		return nil, httperror.NewForBadRequestWithSingleField("id", "does not exist")
	}

	ou.Status = programcategory_s.ProgramCategoryStatusArchived

	if err := impl.ProgramCategoryStorer.UpdateByID(ctx, ou); err != nil {
		impl.Logger.Error("programcategory update by id error", slog.Any("error", err))
		return nil, err
	}
	return ou, nil
}
