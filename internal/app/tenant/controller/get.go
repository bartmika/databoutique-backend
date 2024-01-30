package controller

import (
	"context"

	domain "github.com/bartmika/databoutique-backend/internal/app/tenant/datastore"
	user_d "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"
)

func (c *TenantControllerImpl) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Tenant, error) {
	// Extract from our session the following data.
	userTenantID := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	userRole := ctx.Value(constants.SessionUserRole).(int8)

	// If user is not administrator nor belongs to the Tenant then error.
	if userRole != user_d.UserRoleExecutive && id != userTenantID {
		c.Logger.Error("authenticated user is not staff role nor belongs to the Tenant error",
			slog.Any("userRole", userRole),
			slog.Any("userTenantID", userTenantID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "you do not belong to this Tenant")
	}

	// Retrieve from our database the record for the specific id.
	m, err := c.TenantStorer.GetByID(ctx, id)
	if err != nil {
		c.Logger.Error("database get by id error", slog.Any("error", err))
		return nil, err
	}
	return m, err
}
