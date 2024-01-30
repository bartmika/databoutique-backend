package controller

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"

	howhear_s "github.com/bartmika/databoutique-backend/internal/app/howhear/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func (impl *HowHearAboutUsItemControllerImpl) ArchiveByID(ctx context.Context, id primitive.ObjectID) (*howhear_s.HowHearAboutUsItem, error) {
	// // Extract from our session the following data.
	// userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)

	// Lookup the howhear in our database, else return a `400 Bad Request` error.
	ou, err := impl.HowHearAboutUsItemStorer.GetByID(ctx, id)
	if err != nil {
		impl.Logger.Error("database error", slog.Any("err", err))
		return nil, err
	}
	if ou == nil {
		impl.Logger.Warn("howhear does not exist validation error")
		return nil, httperror.NewForBadRequestWithSingleField("id", "does not exist")
	}

	ou.Status = howhear_s.HowHearAboutUsItemStatusArchived

	if err := impl.HowHearAboutUsItemStorer.UpdateByID(ctx, ou); err != nil {
		impl.Logger.Error("howhear update by id error", slog.Any("error", err))
		return nil, err
	}
	return ou, nil
}
