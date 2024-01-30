package controller

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	s3_storage "github.com/bartmika/databoutique-backend/internal/adapter/storage/s3"
	"github.com/bartmika/databoutique-backend/internal/adapter/templatedemailer"
	assistant_s "github.com/bartmika/databoutique-backend/internal/app/assistant/datastore"
	assistantfile "github.com/bartmika/databoutique-backend/internal/app/assistantfile/datastore"
	assistantmessage_s "github.com/bartmika/databoutique-backend/internal/app/assistantmessage/datastore"
	assistantthread_s "github.com/bartmika/databoutique-backend/internal/app/assistantthread/datastore"
	t_s "github.com/bartmika/databoutique-backend/internal/app/assistantthread/datastore"
	tenant_s "github.com/bartmika/databoutique-backend/internal/app/tenant/datastore"
	user_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config"
	"github.com/bartmika/databoutique-backend/internal/provider/kmutex"
	"github.com/bartmika/databoutique-backend/internal/provider/password"
	"github.com/bartmika/databoutique-backend/internal/provider/uuid"
)

// AssistantThreadController Interface for assistantthread business logic controller.
type AssistantThreadController interface {
	Create(ctx context.Context, requestData *AssistantThreadCreateRequestIDO) (*assistantthread_s.AssistantThread, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*assistantthread_s.AssistantThread, error)
	UpdateByID(ctx context.Context, nu *AssistantThreadUpdateRequestIDO) (*assistantthread_s.AssistantThread, error)
	ListByFilter(ctx context.Context, f *t_s.AssistantThreadPaginationListFilter) (*t_s.AssistantThreadPaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *assistantthread_s.AssistantThreadListFilter) ([]*assistantthread_s.AssistantThreadAsSelectOption, error)
	ArchiveByID(ctx context.Context, id primitive.ObjectID) (*assistantthread_s.AssistantThread, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type AssistantThreadControllerImpl struct {
	Config                 *config.Conf
	Logger                 *slog.Logger
	UUID                   uuid.Provider
	S3                     s3_storage.S3Storager
	Password               password.Provider
	Kmutex                 kmutex.Provider
	DbClient               *mongo.Client
	TenantStorer           tenant_s.TenantStorer
	UserStorer             user_s.UserStorer
	AssistantFileStorer    assistantfile.AssistantFileStorer
	AssistantStorer        assistant_s.AssistantStorer
	AssistantThreadStorer  t_s.AssistantThreadStorer
	AssistantMessageStorer assistantmessage_s.AssistantMessageStorer
	TemplatedEmailer       templatedemailer.TemplatedEmailer
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
	af_storer assistantfile.AssistantFileStorer,
	a_storer assistant_s.AssistantStorer,
	at_storer assistantthread_s.AssistantThreadStorer,
	am_storer assistantmessage_s.AssistantMessageStorer,
) AssistantThreadController {
	s := &AssistantThreadControllerImpl{
		Config:                 appCfg,
		Logger:                 loggerp,
		UUID:                   uuidp,
		S3:                     s3,
		Password:               passwordp,
		Kmutex:                 kmux,
		TemplatedEmailer:       temailer,
		DbClient:               client,
		TenantStorer:           t_storer,
		UserStorer:             usr_storer,
		AssistantFileStorer:    af_storer,
		AssistantStorer:        a_storer,
		AssistantThreadStorer:  at_storer,
		AssistantMessageStorer: am_storer,
	}
	s.Logger.Debug("assistant thread controller initialization started...")
	s.Logger.Debug("assistant thread controller initialized")
	return s
}
