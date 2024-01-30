package controller

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"

	executable_s "github.com/bartmika/databoutique-backend/internal/app/executable/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func (impl *ExecutableControllerImpl) ArchiveByID(ctx context.Context, id primitive.ObjectID) (*executable_s.Executable, error) {
	// // Extract from our session the following data.
	// userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)

	// Lookup the executable in our database, else return a `400 Bad Request` error.
	ou, err := impl.ExecutableStorer.GetByID(ctx, id)
	if err != nil {
		impl.Logger.Error("database error", slog.Any("err", err))
		return nil, err
	}
	if ou == nil {
		impl.Logger.Warn("executable does not exist validation error")
		return nil, httperror.NewForBadRequestWithSingleField("id", "does not exist")
	}

	ou.Status = executable_s.ExecutableStatusArchived

	if err := impl.ExecutableStorer.UpdateByID(ctx, ou); err != nil {
		impl.Logger.Error("executable update by id error", slog.Any("error", err))
		return nil, err
	}
	return ou, nil
}
