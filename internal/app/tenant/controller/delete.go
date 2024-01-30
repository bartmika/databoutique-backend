package controller

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"

	org_d "github.com/bartmika/databoutique-backend/internal/app/tenant/datastore"
	user_d "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func (impl *TenantControllerImpl) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	// Extract from our session the following data.
	userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	userRole := ctx.Value(constants.SessionUserRole).(int8)

	// Apply protection based on ownership and role.
	if userRole != user_d.UserRoleExecutive {
		impl.Logger.Error("authenticated user is not staff role error",
			slog.Any("role", userRole),
			slog.Any("userID", userID))
		return httperror.NewForForbiddenWithSingleField("message", "you role does not grant you access to this")
	}

	// Update the database.
	tenant, err := impl.GetByID(ctx, id)
	tenant.Status = org_d.TenantArchivedStatus
	if err != nil {
		impl.Logger.Error("database get by id error", slog.Any("error", err))
		return err
	}
	if tenant == nil {
		impl.Logger.Error("database returns nothing from get by id")
		return err
	}
	// Security: Prevent deletion of root user(s).
	if userRole == org_d.RootType {
		impl.Logger.Warn("root tenant cannot be deleted error")
		return httperror.NewForForbiddenWithSingleField("role", "root tenant cannot be deleted")
	}

	// Save to the database the modified tenant.
	if err := impl.TenantStorer.UpdateByID(ctx, tenant); err != nil {
		impl.Logger.Error("database update by id error", slog.Any("error", err))
		return err
	}
	return nil
}
