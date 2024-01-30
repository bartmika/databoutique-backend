package controller

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/bartmika/databoutique-backend/internal/adapter/templatedemailer"
	tenant_s "github.com/bartmika/databoutique-backend/internal/app/tenant/datastore"
	user_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config"
	"github.com/bartmika/databoutique-backend/internal/provider/kmutex"
	"github.com/bartmika/databoutique-backend/internal/provider/password"
	"github.com/bartmika/databoutique-backend/internal/provider/uuid"
)

// UserController Interface for user business logic controller.
type UserController interface {
	Create(ctx context.Context, requestData *UserCreateRequestIDO) (*user_s.User, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*user_s.User, error)
	GetUserBySessionUUID(ctx context.Context, sessionUUID string) (*user_s.User, error)
	ArchiveByID(ctx context.Context, id primitive.ObjectID) (*user_s.User, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	ListByFilter(ctx context.Context, f *user_s.UserListFilter) (*user_s.UserListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *user_s.UserListFilter) ([]*user_s.UserAsSelectOption, error)
	CountByFilter(ctx context.Context, f *user_s.UserListFilter) (*UserCountResult, error)
	UpdateByID(ctx context.Context, request *UserUpdateRequestIDO) (*user_s.User, error)
	CreateComment(ctx context.Context, customerID primitive.ObjectID, content string) (*user_s.User, error)
}

type UserControllerImpl struct {
	Config           *config.Conf
	Logger           *slog.Logger
	UUID             uuid.Provider
	Password         password.Provider
	Kmutex           kmutex.Provider
	DbClient         *mongo.Client
	TenantStorer     tenant_s.TenantStorer
	UserStorer       user_s.UserStorer
	TemplatedEmailer templatedemailer.TemplatedEmailer
}

func NewController(
	appCfg *config.Conf,
	loggerp *slog.Logger,
	uuidp uuid.Provider,
	passwordp password.Provider,
	kmux kmutex.Provider,
	client *mongo.Client,
	org_storer tenant_s.TenantStorer,
	usr_storer user_s.UserStorer,
	temailer templatedemailer.TemplatedEmailer,
) UserController {
	s := &UserControllerImpl{
		Config:           appCfg,
		Logger:           loggerp,
		UUID:             uuidp,
		Password:         passwordp,
		Kmutex:           kmux,
		DbClient:         client,
		TenantStorer:     org_storer,
		UserStorer:       usr_storer,
		TemplatedEmailer: temailer,
	}
	s.Logger.Debug("user controller initialization started...")

	s.Logger.Debug("user controller initialized")
	return s
}
