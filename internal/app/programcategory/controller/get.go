package controller

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"

	programcategory_s "github.com/bartmika/databoutique-backend/internal/app/programcategory/datastore"
)

func (c *ProgramCategoryControllerImpl) GetByID(ctx context.Context, id primitive.ObjectID) (*programcategory_s.ProgramCategory, error) {
	// Retrieve from our database the record for the specific id.
	m, err := c.ProgramCategoryStorer.GetByID(ctx, id)
	if err != nil {
		c.Logger.Error("database get by id error", slog.Any("error", err))
		return nil, err
	}
	return m, err
}
