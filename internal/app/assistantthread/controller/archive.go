package controller

import (
	"context"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"

	assistantthread_s "github.com/bartmika/databoutique-backend/internal/app/assistantthread/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func (impl *AssistantThreadControllerImpl) ArchiveByID(ctx context.Context, id primitive.ObjectID) (*assistantthread_s.AssistantThread, error) {
	// // Extract from our session the following data.
	// userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)

	// Lookup the assistantthread in our database, else return a `400 Bad Request` error.
	ou, err := impl.AssistantThreadStorer.GetByID(ctx, id)
	if err != nil {
		impl.Logger.Error("database error", slog.Any("err", err))
		return nil, err
	}
	if ou == nil {
		impl.Logger.Warn("assistantthread does not exist validation error")
		return nil, httperror.NewForBadRequestWithSingleField("id", "does not exist")
	}

	ou.Status = assistantthread_s.AssistantThreadStatusArchived

	if err := impl.AssistantThreadStorer.UpdateByID(ctx, ou); err != nil {
		impl.Logger.Error("assistantthread update by id error", slog.Any("error", err))
		return nil, err
	}
	return ou, nil
}
