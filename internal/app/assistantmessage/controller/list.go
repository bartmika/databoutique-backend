package controller

import (
	"context"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"

	assistantmessage_s "github.com/bartmika/databoutique-backend/internal/app/assistantmessage/datastore"
	t_s "github.com/bartmika/databoutique-backend/internal/app/assistantmessage/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
)

func (c *AssistantMessageControllerImpl) ListByFilter(ctx context.Context, f *t_s.AssistantMessagePaginationListFilter) (*t_s.AssistantMessagePaginationListResult, error) {
	// // Extract from our session the following data.
	tenantID := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)

	// Apply filtering based on ownership and role.
	f.TenantID = tenantID // Manditory

	c.Logger.Debug("listing using filter options:",
		slog.Any("Cursor", f.Cursor),
		slog.Int64("PageSize", f.PageSize),
		slog.String("SortField", f.SortField),
		// slog.Int("SortOrder", int(f.SortOrder)),
		// slog.Any("TenantID", f.TenantID),
		// slog.Any("Type", f.Type),
		// slog.Any("Status", f.Status),
		// slog.Bool("ExcludeArchived", f.ExcludeArchived),
		// slog.String("SearchText", f.SearchText),
		// slog.Any("FirstName", f.FirstName),
		// slog.Any("LastName", f.LastName),
		// slog.Any("Email", f.Email),
		// slog.Any("Phone", f.Phone),
		// slog.Time("CreatedAtGTE", f.CreatedAtGTE)
	)

	m, err := c.AssistantMessageStorer.ListByFilter(ctx, f)
	if err != nil {
		c.Logger.Error("database list by filter error", slog.Any("error", err))
		return nil, err
	}
	return m, err
}

// func (c *AssistantMessageControllerImpl) LiteListByFilter(ctx context.Context, f *t_s.AssistantMessagePaginationListFilter) (*t_s.AssistantMessageLiteListResult, error) {
// 	// // Extract from our session the following data.
// 	tenantID := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
//
// 	// Apply filtering based on ownership and role.
// 	f.TenantID = tenantID // Manditory
//
// 	c.Logger.Debug("listing using filter options:",
// 		slog.Any("Cursor", f.Cursor),
// 		slog.Int64("PageSize", f.PageSize),
// 		slog.String("SortField", f.SortField),
// 		slog.Int("SortOrder", int(f.SortOrder)),
// 		slog.Any("TenantID", f.TenantID),
// 	)
//
// 	m, err := c.AssistantMessageStorer.LiteListByFilter(ctx, f)
// 	if err != nil {
// 		c.Logger.Error("database list by filter error", slog.Any("error", err))
// 		return nil, err
// 	}
// 	return m, err
// }

func (c *AssistantMessageControllerImpl) ListAsSelectOptionByFilter(ctx context.Context, f *assistantmessage_s.AssistantMessagePaginationListFilter) ([]*assistantmessage_s.AssistantMessageAsSelectOption, error) {
	// // Extract from our session the following data.
	tenantID := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)

	// Apply filtering based on ownership and role.
	f.TenantID = tenantID // Manditory

	c.Logger.Debug("listing using filter options:",
		slog.Any("Cursor", f.Cursor),
		slog.Int64("PageSize", f.PageSize),
		slog.String("SortField", f.SortField),
		slog.Int("SortOrder", int(f.SortOrder)),
		slog.Any("TenantID", f.TenantID),
	)

	// Filtering the database.
	m, err := c.AssistantMessageStorer.ListAsSelectOptionByFilter(ctx, f)
	if err != nil {
		c.Logger.Error("database list by filter error", slog.Any("error", err))
		return nil, err
	}
	return m, err
}

func (c *AssistantMessageControllerImpl) PublicListAsSelectOptionByFilter(ctx context.Context, f *assistantmessage_s.AssistantMessagePaginationListFilter) ([]*assistantmessage_s.AssistantMessageAsSelectOption, error) {

	// // If unspecified the tenant then auto-assign the default tenant in our app.
	// if tenant == nil {
	// 	tenant, err = impl.TenantStorer.GetByID(sessCtx, impl.Config.InitialAccount.AdminTenantID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	c.Logger.Debug("listing using filter options:",
		slog.Any("Cursor", f.Cursor),
		slog.Int64("PageSize", f.PageSize),
		slog.String("SortField", f.SortField),
		slog.Int("SortOrder", int(f.SortOrder)),
		slog.Any("TenantID", f.TenantID),
	)

	// Filtering the database.
	m, err := c.AssistantMessageStorer.ListAsSelectOptionByFilter(ctx, f)
	if err != nil {
		c.Logger.Error("database list by filter error", slog.Any("error", err))
		return nil, err
	}
	return m, err
}
