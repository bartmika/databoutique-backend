package controller

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"

	user_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func (impl *UserControllerImpl) ArchiveByID(ctx context.Context, id primitive.ObjectID) (*user_s.User, error) {
	// Extract from our session the following data.
	userRole := ctx.Value(constants.SessionUserRole).(int8)

	// Apply filtering based on ownership and role.
	if userRole != user_s.UserRoleExecutive {
		return nil, httperror.NewForForbiddenWithSingleField("message", "you do not have permission")
	}

	// Lookup the user in our database, else return a `400 Bad Request` error.
	ou, err := impl.UserStorer.GetByID(ctx, id)
	if err != nil {
		impl.Logger.Error("database error", slog.Any("err", err))
		return nil, err
	}
	if ou == nil {
		impl.Logger.Warn("user does not exist validation error")
		return nil, httperror.NewForBadRequestWithSingleField("id", "does not exist")
	}

	// Security: Prevent deletion of root user(s).
	if ou.Role == user_s.UserRoleExecutive {
		impl.Logger.Warn("root user(s) cannot be deleted error")
		return nil, httperror.NewForForbiddenWithSingleField("role", "root user(s) cannot be deleted")
	}

	ou.ModifiedAt = time.Now()
	ou.Status = user_s.UserStatusArchived

	if err := impl.UserStorer.UpdateByID(ctx, ou); err != nil {
		impl.Logger.Error("user update by id error", slog.Any("error", err))
		return nil, err
	}
	return ou, nil
}
