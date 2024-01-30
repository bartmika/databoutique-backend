package controller

import (
	"context"

	"log/slog"

	domain "github.com/bartmika/databoutique-backend/internal/app/assistantfile/datastore"
	user_d "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (c *AssistantFileControllerImpl) ListByFilter(ctx context.Context, f *domain.AssistantFilePaginationListFilter) (*domain.AssistantFilePaginationListResult, error) {
	// Extract from our session the following data.
	orgID := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	userRole := ctx.Value(constants.SessionUserRole).(int8)

	// Apply protection based on ownership and role.
	if userRole != user_d.UserRoleExecutive {
		f.TenantID = orgID // Force tenant tenancy restrictions.
	}

	c.Logger.Debug("fetching assistant files now...", slog.Any("userID", userID))

	aa, err := c.AssistantFileStorer.ListByFilter(ctx, f)
	if err != nil {
		c.Logger.Error("database list by filter error", slog.Any("error", err))
		return nil, err
	}
	c.Logger.Debug("fetched assistant files", slog.Any("aa", aa))

	// for _, a := range aa.Results {
	// 	// Generate the URL.
	// 	fileURL, err := c.S3.GetPresignedURL(ctx, a.ObjectKey, 5*time.Minute)
	// 	if err != nil {
	// 		c.Logger.Error("s3 failed get presigned url error", slog.Any("error", err))
	// 		return nil, err
	// 	}
	// 	a.ObjectURL = fileURL
	// }
	return aa, err
}

func (c *AssistantFileControllerImpl) ListAsSelectOptionByFilter(ctx context.Context, f *domain.AssistantFilePaginationListFilter) ([]*domain.AssistantFileAsSelectOption, error) {
	// Extract from our session the following data.
	userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	userRole := ctx.Value(constants.SessionUserRole).(int8)

	// Apply protection based on ownership and role.
	if userRole != user_d.UserRoleExecutive {
		c.Logger.Error("authenticated user is not staff role error",
			slog.Any("role", userRole),
			slog.Any("userID", userID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "you role does not grant you access to this")
	}

	c.Logger.Debug("fetching assistant files now...", slog.Any("userID", userID))

	m, err := c.AssistantFileStorer.ListAsSelectOptionByFilter(ctx, f)
	if err != nil {
		c.Logger.Error("database list by filter error", slog.Any("error", err))
		return nil, err
	}
	c.Logger.Debug("fetched assistant files", slog.Any("m", m))
	return m, err
}
