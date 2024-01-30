package controller

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"

	program_s "github.com/bartmika/databoutique-backend/internal/app/program/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func (impl *ProgramControllerImpl) ArchiveByID(ctx context.Context, id primitive.ObjectID) (*program_s.Program, error) {
	// // Extract from our session the following data.
	// userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)

	// Lookup the program in our database, else return a `400 Bad Request` error.
	ou, err := impl.ProgramStorer.GetByID(ctx, id)
	if err != nil {
		impl.Logger.Error("database error", slog.Any("err", err))
		return nil, err
	}
	if ou == nil {
		impl.Logger.Warn("program does not exist validation error")
		return nil, httperror.NewForBadRequestWithSingleField("id", "does not exist")
	}

	ou.Status = program_s.ProgramStatusArchived

	if err := impl.ProgramStorer.UpdateByID(ctx, ou); err != nil {
		impl.Logger.Error("program update by id error", slog.Any("error", err))
		return nil, err
	}
	return ou, nil
}
