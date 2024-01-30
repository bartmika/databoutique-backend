package controller

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"

	program_s "github.com/bartmika/databoutique-backend/internal/app/program/datastore"
)

func (c *ProgramControllerImpl) GetByID(ctx context.Context, id primitive.ObjectID) (*program_s.Program, error) {
	// Retrieve from our database the record for the specific id.
	m, err := c.ProgramStorer.GetByID(ctx, id)
	if err != nil {
		c.Logger.Error("database get by id error", slog.Any("error", err))
		return nil, err
	}
	return m, err
}
