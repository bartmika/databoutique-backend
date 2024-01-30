package controller

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"

	assistant_s "github.com/bartmika/databoutique-backend/internal/app/assistant/datastore"
)

func (c *AssistantControllerImpl) GetByID(ctx context.Context, id primitive.ObjectID) (*assistant_s.Assistant, error) {
	// Retrieve from our database the record for the specific id.
	m, err := c.AssistantStorer.GetByID(ctx, id)
	if err != nil {
		c.Logger.Error("database get by id error", slog.Any("error", err))
		return nil, err
	}
	return m, err
}
