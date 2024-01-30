package controller

import (
	"context"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"

	assistantthread_s "github.com/bartmika/databoutique-backend/internal/app/assistantthread/datastore"
)

func (c *AssistantThreadControllerImpl) GetByID(ctx context.Context, id primitive.ObjectID) (*assistantthread_s.AssistantThread, error) {
	// Retrieve from our database the record for the specific id.
	at, err := c.AssistantThreadStorer.GetByID(ctx, id)
	if err != nil {
		c.Logger.Error("database get by id error", slog.Any("error", err))
		return nil, err
	}
	return at, err
}
