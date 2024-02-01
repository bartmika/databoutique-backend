package controller

import (
	"context"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"

	executable_s "github.com/bartmika/databoutique-backend/internal/app/executable/datastore"
)

func (impl *ExecutableControllerImpl) GetByID(ctx context.Context, id primitive.ObjectID) (*executable_s.Executable, error) {
	// Keep data consistent.
	impl.Kmutex.Lockf("executable_%s", id.Hex())
	defer impl.Kmutex.Unlockf("executable_%s", id.Hex())

	// Retrieve from our database the record for the specific id.
	m, err := impl.ExecutableStorer.GetByID(ctx, id)
	if err != nil {
		impl.Logger.Error("failed getting executable",
			slog.String("id", id.Hex()),
			slog.Any("error", err))
		return nil, err
	}
	return m, err
}
