package controller

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"

	program_s "github.com/bartmika/databoutique-backend/internal/app/program/datastore"
)

func (impl *ProgramControllerImpl) GetByID(ctx context.Context, id primitive.ObjectID) (*program_s.Program, error) {
	impl.Kmutex.Lockf("openai_program_%s", id.Hex())
	defer impl.Kmutex.Unlockf("openai_program_%s", id.Hex())

	// Retrieve from our database the record for the specific id.
	m, err := impl.ProgramStorer.GetByID(ctx, id)
	if err != nil {
		impl.Logger.Error("database get by id error", slog.Any("error", err))
		return nil, err
	}
	return m, err
}
