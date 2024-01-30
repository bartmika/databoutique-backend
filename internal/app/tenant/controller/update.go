package controller

import (
	"context"
	"time"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"

	domain "github.com/bartmika/databoutique-backend/internal/app/tenant/datastore"
	user_d "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func validateUpdateRequest(dirtyData *domain.Tenant) error {
	e := make(map[string]string)

	// if dirtyData.ServiceType == 0 {
	// 	e["service_type"] = "missing value"
	// }
	if dirtyData.Name == "" {
		e["name"] = "missing value"
	}
	if dirtyData.Description == "" {
		e["description"] = "missing value"
	}
	if len(e) != 0 {
		return httperror.NewForBadRequest(&e)
	}
	return nil
}

func (c *TenantControllerImpl) UpdateByID(ctx context.Context, ns *domain.Tenant) (*domain.Tenant, error) {
	// Perform our validation and return validation error on any issues detected.
	if err := validateUpdateRequest(ns); err != nil {
		return nil, err
	}

	// Fetch the original Tenant.
	os, err := c.TenantStorer.GetByID(ctx, ns.ID)
	if err != nil {
		c.Logger.Error("database get by id error", slog.Any("error", err))
		return nil, err
	}
	if os == nil {
		c.Logger.Error("Tenant does not exist error",
			slog.Any("Tenant_id", ns.ID))
		return nil, httperror.NewForBadRequestWithSingleField("message", "Tenant does not exist")
	}

	// Extract from our session the following data.
	userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	userTenantID := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	userRole := ctx.Value(constants.SessionUserRole).(int8)
	userName := ctx.Value(constants.SessionUserName).(string)

	// If user is not administrator nor belongs to the Tenant then error.
	if userRole != user_d.UserRoleExecutive && os.ID != userTenantID {
		c.Logger.Error("authenticated user is not staff role nor belongs to the Tenant error",
			slog.Any("userRole", userRole),
			slog.Any("userTenantID", userTenantID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "you do not belong to this Tenant")
	}

	// Modify our original Tenant.
	os.ModifiedAt = time.Now()
	os.ModifiedByUserID = userID
	os.ModifiedByUserName = userName
	os.Status = ns.Status
	os.Name = ns.Name
	os.Description = ns.Description

	// Save to the database the modified Tenant.
	if err := c.TenantStorer.UpdateByID(ctx, os); err != nil {
		c.Logger.Error("database update by id error", slog.Any("error", err))
		return nil, err
	}

	return os, nil
}
