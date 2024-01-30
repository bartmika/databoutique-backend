package controller

import (
	"context"

	user_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"
)

func (impl *UserControllerImpl) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	// Extract from our session the following data.
	userRole := ctx.Value(constants.SessionUserRole).(int8)

	// Apply filtering based on ownership and role.
	if userRole != user_s.UserRoleExecutive {
		return httperror.NewForForbiddenWithSingleField("message", "you do not have permission")
	}

	// STEP 1: Lookup the record or error.
	user, err := impl.UserStorer.GetByID(ctx, id)
	if err != nil {
		impl.Logger.Error("database get by id error", slog.Any("error", err))
		return err
	}
	if user == nil {
		impl.Logger.Error("database returns nothing from get by id")
		return err
	}

	// Security: Prevent deletion of root user(s).
	if user.Role == user_s.UserRoleExecutive {
		impl.Logger.Warn("root user(s) cannot be deleted error")
		return httperror.NewForForbiddenWithSingleField("role", "root user(s) cannot be deleted")
	}

	// STEP 2: Delete from database.
	if err := impl.UserStorer.DeleteByID(ctx, id); err != nil {
		impl.Logger.Error("database delete by id error", slog.Any("error", err))
		return err
	}
	return nil
}
