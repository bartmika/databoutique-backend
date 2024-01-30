package controller

import (
	"context"

	domain "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	user_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"
)

func (c *UserControllerImpl) GetByID(ctx context.Context, id primitive.ObjectID) (*user_s.User, error) {
	// Extract from our session the following data.
	userRole := ctx.Value(constants.SessionUserRole).(int8)

	// Apply filtering based on ownership and role.
	if userRole != user_s.UserRoleExecutive {
		return nil, httperror.NewForForbiddenWithSingleField("message", "you do not have permission")
	}

	// Retrieve from our database the record for the specific id.
	m, err := c.UserStorer.GetByID(ctx, id)
	if err != nil {
		c.Logger.Error("database get by id error", slog.Any("error", err))
		return nil, err
	}
	return m, err
}

// CreateInitialRootAdmin function creates the initial root administrator if not previously created.
func (c *UserControllerImpl) GetUserBySessionUUID(ctx context.Context, sessionUUID string) (*domain.User, error) {
	panic("TODO: IMPLEMENT")
	return nil, nil
}
