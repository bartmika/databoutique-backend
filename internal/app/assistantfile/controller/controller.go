package controller

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	mg "github.com/bartmika/databoutique-backend/internal/adapter/emailer/mailgun"
	s3_storage "github.com/bartmika/databoutique-backend/internal/adapter/storage/s3"
	assistantfile_s "github.com/bartmika/databoutique-backend/internal/app/assistantfile/datastore"
	domain "github.com/bartmika/databoutique-backend/internal/app/assistantfile/datastore"
	tenant_s "github.com/bartmika/databoutique-backend/internal/app/tenant/datastore"
	user_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config"
	"github.com/bartmika/databoutique-backend/internal/provider/uuid"
)

// AssistantFileController Interface for assistantfile business logic controller.
type AssistantFileController interface {
	Create(ctx context.Context, req *AssistantFileCreateRequestIDO) (*domain.AssistantFile, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.AssistantFile, error)
	UpdateByID(ctx context.Context, ns *AssistantFileUpdateRequestIDO) (*domain.AssistantFile, error)
	ListByFilter(ctx context.Context, f *domain.AssistantFilePaginationListFilter) (*domain.AssistantFilePaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *domain.AssistantFilePaginationListFilter) ([]*domain.AssistantFileAsSelectOption, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	PermanentlyDeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type AssistantFileControllerImpl struct {
	Config              *config.Conf
	Logger              *slog.Logger
	UUID                uuid.Provider
	S3                  s3_storage.S3Storager
	Emailer             mg.Emailer
	DbClient            *mongo.Client
	TenantStorer        tenant_s.TenantStorer
	AssistantFileStorer assistantfile_s.AssistantFileStorer
	UserStorer          user_s.UserStorer
}

func NewController(
	appCfg *config.Conf,
	loggerp *slog.Logger,
	uuidp uuid.Provider,
	s3 s3_storage.S3Storager,
	client *mongo.Client,
	emailer mg.Emailer,
	t_storer tenant_s.TenantStorer,
	org_storer assistantfile_s.AssistantFileStorer,
	usr_storer user_s.UserStorer,
) AssistantFileController {
	s := &AssistantFileControllerImpl{
		Config:              appCfg,
		Logger:              loggerp,
		UUID:                uuidp,
		S3:                  s3,
		Emailer:             emailer,
		DbClient:            client,
		TenantStorer:        t_storer,
		AssistantFileStorer: org_storer,
		UserStorer:          usr_storer,
	}
	s.Logger.Debug("assistantfile controller initialization started...")
	s.Logger.Debug("assistantfile controller initialized")
	return s
}
