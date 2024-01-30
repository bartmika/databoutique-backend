package controller

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"
)

func (impl *ProgramCategoryControllerImpl) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	// STEP 1: Lookup the record or error.
	programcategory, err := impl.GetByID(ctx, id)
	if err != nil {
		impl.Logger.Error("database get by id error", slog.Any("error", err))
		return err
	}
	if programcategory == nil {
		impl.Logger.Error("database returns nothing from get by id")
		return err
	}

	// STEP 2: Delete from database.
	if err := impl.ProgramCategoryStorer.DeleteByID(ctx, id); err != nil {
		impl.Logger.Error("database delete by id error", slog.Any("error", err))
		return err
	}
	return nil
}
