package controller

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"

	howhear_s "github.com/bartmika/databoutique-backend/internal/app/howhear/datastore"
)

func (c *HowHearAboutUsItemControllerImpl) GetByID(ctx context.Context, id primitive.ObjectID) (*howhear_s.HowHearAboutUsItem, error) {
	// Retrieve from our database the record for the specific id.
	m, err := c.HowHearAboutUsItemStorer.GetByID(ctx, id)
	if err != nil {
		c.Logger.Error("database get by id error", slog.Any("error", err))
		return nil, err
	}
	return m, err
}
