package controller

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"

	u_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

type DashboardResponseIDO struct {
	ClientsCount    int64 `json:"clients_count"`
	AssociatesCount int64 `json:"associates_count"`
	JobsCount       int64 `json:"jobs_count"`
	TasksCount      int64 `json:"tasks_count"`
}

func (impl *GatewayControllerImpl) getActiveClientsCount(ctx context.Context, tenantID primitive.ObjectID) (int64, error) {
	f := &u_s.UserListFilter{
		TenantID: tenantID,
		Role:     u_s.UserRoleCustomer,
		Status:   u_s.UserStatusActive,
	}
	count, err := impl.UserStorer.CountByFilter(ctx, f)
	if err != nil {
		impl.Logger.Error("database count error", slog.Any("err", err))
		return count, err
	}
	return count, nil
}

func (impl *GatewayControllerImpl) getActiveAssociatesCount(ctx context.Context, tenantID primitive.ObjectID) (int64, error) {
	f := &u_s.UserListFilter{
		Cursor:   primitive.NilObjectID,
		TenantID: tenantID,
		Role:     u_s.UserRoleAssociate,
		Status:   u_s.UserStatusActive,
	}
	count, err := impl.UserStorer.CountByFilter(ctx, f)
	if err != nil {
		impl.Logger.Error("database count error", slog.Any("err", err))
		return count, err
	}
	return count, nil
}

func (impl *GatewayControllerImpl) getActiveJobsCount(ctx context.Context, tenantID primitive.ObjectID) (int64, error) {
	return 58, nil
}

func (impl *GatewayControllerImpl) getActiveTasksCount(ctx context.Context, tenantID primitive.ObjectID) (int64, error) {
	return 809, nil
}

func (impl *GatewayControllerImpl) Dashboard(ctx context.Context) (*DashboardResponseIDO, error) {
	// Extract from our session the following data.
	tenantID, _ := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	userID, _ := ctx.Value(constants.SessionUserID).(primitive.ObjectID)

	// Lookup the user in our database, else return a `400 Bad Request` error.
	u, err := impl.UserStorer.GetByID(ctx, userID)
	if err != nil {
		impl.Logger.Error("database error", slog.Any("err", err))
		return nil, err
	}
	if u == nil {
		impl.Logger.Warn("user does not exist validation error")
		return nil, httperror.NewForBadRequestWithSingleField("id", "does not exist")
	}

	////
	//// Get counts.
	////

	clientsCount, err := impl.getActiveClientsCount(ctx, tenantID)
	if err != nil {
		impl.Logger.Error("database error", slog.Any("err", err))
		return nil, err
	}
	associatesCount, err := impl.getActiveAssociatesCount(ctx, tenantID)
	if err != nil {
		impl.Logger.Error("database error", slog.Any("err", err))
		return nil, err
	}
	jobsCount, err := impl.getActiveJobsCount(ctx, tenantID)
	if err != nil {
		impl.Logger.Error("database error", slog.Any("err", err))
		return nil, err
	}
	tasksCount, err := impl.getActiveTasksCount(ctx, tenantID)
	if err != nil {
		impl.Logger.Error("database error", slog.Any("err", err))
		return nil, err
	}

	////
	//// Return the response.
	////

	res := &DashboardResponseIDO{
		ClientsCount:    clientsCount,
		AssociatesCount: associatesCount,
		JobsCount:       jobsCount,
		TasksCount:      tasksCount,
	}
	return res, nil
}
