package controller

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"

	executable_s "github.com/bartmika/databoutique-backend/internal/app/executable/datastore"
)

func (c *ExecutableControllerImpl) GetByID(ctx context.Context, id primitive.ObjectID) (*executable_s.Executable, error) {
	// Retrieve from our database the record for the specific id.
	m, err := c.ExecutableStorer.GetByID(ctx, id)
	if err != nil {
		c.Logger.Error("database get by id error", slog.Any("error", err))
		return nil, err
	}
	return m, err
}
