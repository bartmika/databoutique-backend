package controller

import (
	"context"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"

	assistantthread_s "github.com/bartmika/databoutique-backend/internal/app/assistantthread/datastore"
	t_s "github.com/bartmika/databoutique-backend/internal/app/assistantthread/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
)

func (c *AssistantThreadControllerImpl) ListByFilter(ctx context.Context, f *t_s.AssistantThreadPaginationListFilter) (*t_s.AssistantThreadPaginationListResult, error) {
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

	m, err := c.AssistantThreadStorer.ListByFilter(ctx, f)
	if err != nil {
		c.Logger.Error("database list by filter error", slog.Any("error", err))
		return nil, err
	}
	return m, err
}

// func (c *AssistantThreadControllerImpl) LiteListByFilter(ctx context.Context, f *t_s.AssistantThreadListFilter) (*t_s.AssistantThreadLiteListResult, error) {
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
// 	m, err := c.AssistantThreadStorer.LiteListByFilter(ctx, f)
// 	if err != nil {
// 		c.Logger.Error("database list by filter error", slog.Any("error", err))
// 		return nil, err
// 	}
// 	return m, err
// }

func (c *AssistantThreadControllerImpl) ListAsSelectOptionByFilter(ctx context.Context, f *assistantthread_s.AssistantThreadListFilter) ([]*assistantthread_s.AssistantThreadAsSelectOption, error) {
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
	m, err := c.AssistantThreadStorer.ListAsSelectOptionByFilter(ctx, f)
	if err != nil {
		c.Logger.Error("database list by filter error", slog.Any("error", err))
		return nil, err
	}
	return m, err
}
