package controller

import (
	"context"
	"time"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"

	s_d "github.com/bartmika/databoutique-backend/internal/app/tenant/datastore"
	user_d "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func (c *TenantControllerImpl) Create(ctx context.Context, m *s_d.Tenant) (*s_d.Tenant, error) {
	// Extract from our session the following data.
	userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	userName := ctx.Value(constants.SessionUserName).(string)
	userRole := ctx.Value(constants.SessionUserRole).(int8)

	// Apply protection based on ownership and role.
	if userRole != user_d.UserRoleExecutive {
		c.Logger.Error("authenticated user is not staff role error",
			slog.Any("role", userRole),
			slog.Any("userID", userID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "you role does not grant you access to this")
	}

	c.Kmutex.Lock("create-tenant")
	defer c.Kmutex.Unlock("create-tenant")

	// Add defaults.
	m.ID = primitive.NewObjectID()
	m.CreatedByUserID = userID
	m.CreatedByUserName = userName
	m.CreatedAt = time.Now()
	m.ModifiedByUserID = userID
	m.ModifiedByUserName = userName
	m.ModifiedAt = time.Now()

	// Save to our database.
	err := c.TenantStorer.Create(ctx, m)
	if err != nil {
		c.Logger.Error("database create error", slog.Any("error", err))
		return nil, err
	}

	return m, nil
}
