package controller

import (
	"context"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"

	assistantmessage_s "github.com/bartmika/databoutique-backend/internal/app/assistantmessage/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func (impl *AssistantMessageControllerImpl) ArchiveByID(ctx context.Context, id primitive.ObjectID) (*assistantmessage_s.AssistantMessage, error) {
	// // Extract from our session the following data.
	// userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)

	// Lookup the assistantmessage in our database, else return a `400 Bad Request` error.
	ou, err := impl.AssistantMessageStorer.GetByID(ctx, id)
	if err != nil {
		impl.Logger.Error("database error", slog.Any("err", err))
		return nil, err
	}
	if ou == nil {
		impl.Logger.Warn("assistantmessage does not exist validation error")
		return nil, httperror.NewForBadRequestWithSingleField("id", "does not exist")
	}

	ou.Status = assistantmessage_s.AssistantMessageStatusArchived

	if err := impl.AssistantMessageStorer.UpdateByID(ctx, ou); err != nil {
		impl.Logger.Error("assistantmessage update by id error", slog.Any("error", err))
		return nil, err
	}
	return ou, nil
}
