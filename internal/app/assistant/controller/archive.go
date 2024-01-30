package controller

import (
	"context"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"

	assistant_s "github.com/bartmika/databoutique-backend/internal/app/assistant/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func (impl *AssistantControllerImpl) ArchiveByID(ctx context.Context, id primitive.ObjectID) (*assistant_s.Assistant, error) {
	// // Extract from our session the following data.
	// userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)

	// Lookup the assistant in our database, else return a `400 Bad Request` error.
	ou, err := impl.AssistantStorer.GetByID(ctx, id)
	if err != nil {
		impl.Logger.Error("database error", slog.Any("err", err))
		return nil, err
	}
	if ou == nil {
		impl.Logger.Warn("assistant does not exist validation error")
		return nil, httperror.NewForBadRequestWithSingleField("id", "does not exist")
	}

	ou.Status = assistant_s.AssistantStatusArchived

	if err := impl.AssistantStorer.UpdateByID(ctx, ou); err != nil {
		impl.Logger.Error("assistant update by id error", slog.Any("error", err))
		return nil, err
	}
	return ou, nil
}
