package controller

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	s3_storage "github.com/bartmika/databoutique-backend/internal/adapter/storage/s3"
	"github.com/bartmika/databoutique-backend/internal/adapter/templatedemailer"
	executable_s "github.com/bartmika/databoutique-backend/internal/app/executable/datastore"
	user_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config"
	"github.com/bartmika/databoutique-backend/internal/provider/kmutex"
	"github.com/bartmika/databoutique-backend/internal/provider/password"
	"github.com/bartmika/databoutique-backend/internal/provider/uuid"
)

// ExecutableController Interface for executable business logic controller.
type ExecutableController interface {
	Create(ctx context.Context, requestData *ExecutableCreateRequestIDO) (*executable_s.Executable, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*executable_s.Executable, error)
	UpdateByID(ctx context.Context, requestData *ExecutableUpdateRequestIDO) (*executable_s.Executable, error)
	ListByFilter(ctx context.Context, f *executable_s.ExecutablePaginationListFilter) (*executable_s.ExecutablePaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *executable_s.ExecutablePaginationListFilter) ([]*executable_s.ExecutableAsSelectOption, error)
	PublicListAsSelectOptionByFilter(ctx context.Context, f *executable_s.ExecutablePaginationListFilter) ([]*executable_s.ExecutableAsSelectOption, error)
	ArchiveByID(ctx context.Context, id primitive.ObjectID) (*executable_s.Executable, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type ExecutableControllerImpl struct {
	Config                   *config.Conf
	Logger                   *slog.Logger
	UUID                     uuid.Provider
	S3                       s3_storage.S3Storager
	Password                 password.Provider
	Kmutex                   kmutex.Provider
	DbClient                 *mongo.Client
	UserStorer               user_s.UserStorer
	ExecutableStorer executable_s.ExecutableStorer
	TemplatedEmailer         templatedemailer.TemplatedEmailer
}

func NewController(
	appCfg *config.Conf,
	loggerp *slog.Logger,
	uuidp uuid.Provider,
	s3 s3_storage.S3Storager,
	passwordp password.Provider,
	kmux kmutex.Provider,
	temailer templatedemailer.TemplatedEmailer,
	client *mongo.Client,
	usr_storer user_s.UserStorer,
	executable_s executable_s.ExecutableStorer,
) ExecutableController {
	s := &ExecutableControllerImpl{
		Config:                   appCfg,
		Logger:                   loggerp,
		UUID:                     uuidp,
		S3:                       s3,
		Password:                 passwordp,
		Kmutex:                   kmux,
		TemplatedEmailer:         temailer,
		DbClient:                 client,
		UserStorer:               usr_storer,
		ExecutableStorer: executable_s,
	}
	s.Logger.Debug("executable controller initialization started...")
	s.Logger.Debug("executable controller initialized")
	return s
}
