package controller

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"

	uploaddirectory_s "github.com/bartmika/databoutique-backend/internal/app/uploaddirectory/datastore"
)

func (c *UploadDirectoryControllerImpl) GetByID(ctx context.Context, id primitive.ObjectID) (*uploaddirectory_s.UploadDirectory, error) {
	// Retrieve from our database the record for the specific id.
	m, err := c.UploadDirectoryStorer.GetByID(ctx, id)
	if err != nil {
		c.Logger.Error("database get by id error", slog.Any("error", err))
		return nil, err
	}
	return m, err
}
