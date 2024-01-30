package controller

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	s3_storage "github.com/bartmika/databoutique-backend/internal/adapter/storage/s3"
	"github.com/bartmika/databoutique-backend/internal/adapter/templatedemailer"
	assistant_s "github.com/bartmika/databoutique-backend/internal/app/assistant/datastore"
	t_s "github.com/bartmika/databoutique-backend/internal/app/assistant/datastore"
	assistantfile_s "github.com/bartmika/databoutique-backend/internal/app/assistantfile/datastore"
	tenant_s "github.com/bartmika/databoutique-backend/internal/app/tenant/datastore"
	user_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config"
	"github.com/bartmika/databoutique-backend/internal/provider/kmutex"
	"github.com/bartmika/databoutique-backend/internal/provider/password"
	"github.com/bartmika/databoutique-backend/internal/provider/uuid"
)

// AssistantController Interface for assistant business logic controller.
type AssistantController interface {
	Create(ctx context.Context, requestData *AssistantCreateRequestIDO) (*assistant_s.Assistant, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*assistant_s.Assistant, error)
	UpdateByID(ctx context.Context, nu *AssistantUpdateRequestIDO) (*assistant_s.Assistant, error)
	ListByFilter(ctx context.Context, f *t_s.AssistantPaginationListFilter) (*t_s.AssistantPaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *assistant_s.AssistantListFilter) ([]*assistant_s.AssistantAsSelectOption, error)
	ArchiveByID(ctx context.Context, id primitive.ObjectID) (*assistant_s.Assistant, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type AssistantControllerImpl struct {
	Config              *config.Conf
	Logger              *slog.Logger
	UUID                uuid.Provider
	S3                  s3_storage.S3Storager
	Password            password.Provider
	Kmutex              kmutex.Provider
	DbClient            *mongo.Client
	TenantStorer        tenant_s.TenantStorer
	UserStorer          user_s.UserStorer
	AssistantFileStorer assistantfile_s.AssistantFileStorer
	AssistantStorer     t_s.AssistantStorer
	TemplatedEmailer    templatedemailer.TemplatedEmailer
}

func NewController(
	appCfg *config.Conf,
	loggerp *slog.Logger,
	uuidp uuid.Provider,
	s3 s3_storage.S3Storager,
	passwordp password.Provider,
	kmux kmutex.Provider,
	client *mongo.Client,
	temailer templatedemailer.TemplatedEmailer,
	t_storer tenant_s.TenantStorer,
	usr_storer user_s.UserStorer,
	af_storer assistantfile_s.AssistantFileStorer,
	a_storer assistant_s.AssistantStorer,
) AssistantController {
	s := &AssistantControllerImpl{
		Config:              appCfg,
		Logger:              loggerp,
		UUID:                uuidp,
		S3:                  s3,
		Password:            passwordp,
		Kmutex:              kmux,
		TemplatedEmailer:    temailer,
		DbClient:            client,
		TenantStorer:        t_storer,
		UserStorer:          usr_storer,
		AssistantFileStorer: af_storer,
		AssistantStorer:     a_storer,
	}
	s.Logger.Debug("assistant controller initialization started...")
	s.Logger.Debug("assistant controller initialized")
	return s
}
