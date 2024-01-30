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
	tenant_s "github.com/bartmika/databoutique-backend/internal/app/tenant/datastore"
	user_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config"
	"github.com/bartmika/databoutique-backend/internal/provider/kmutex"
	"github.com/bartmika/databoutique-backend/internal/provider/password"
	"github.com/bartmika/databoutique-backend/internal/provider/uuid"
)

// AssistantMessageController Interface for assistantmessage business logic controller.
type AssistantMessageController interface {
	Create(ctx context.Context, requestData *AssistantMessageCreateRequestIDO) (*assistantmessage_s.AssistantMessage, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*assistantmessage_s.AssistantMessage, error)
	UpdateByID(ctx context.Context, requestData *AssistantMessageUpdateRequestIDO) (*assistantmessage_s.AssistantMessage, error)
	ListByFilter(ctx context.Context, f *assistantmessage_s.AssistantMessagePaginationListFilter) (*assistantmessage_s.AssistantMessagePaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *assistantmessage_s.AssistantMessagePaginationListFilter) ([]*assistantmessage_s.AssistantMessageAsSelectOption, error)
	PublicListAsSelectOptionByFilter(ctx context.Context, f *assistantmessage_s.AssistantMessagePaginationListFilter) ([]*assistantmessage_s.AssistantMessageAsSelectOption, error)
	ArchiveByID(ctx context.Context, id primitive.ObjectID) (*assistantmessage_s.AssistantMessage, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type AssistantMessageControllerImpl struct {
	Config                 *config.Conf
	Logger                 *slog.Logger
	UUID                   uuid.Provider
	S3                     s3_storage.S3Storager
	Password               password.Provider
	Kmutex                 kmutex.Provider
	DbClient               *mongo.Client
	TenantStorer           tenant_s.TenantStorer
	UserStorer             user_s.UserStorer
	TemplatedEmailer       templatedemailer.TemplatedEmailer
	AssistantFileStorer    assistantfile.AssistantFileStorer
	AssistantStorer        assistant_s.AssistantStorer
	AssistantThreadStorer  assistantthread_s.AssistantThreadStorer
	AssistantMessageStorer assistantmessage_s.AssistantMessageStorer
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
	t_storer tenant_s.TenantStorer,
	usr_storer user_s.UserStorer,
	af_storer assistantfile.AssistantFileStorer,
	a_storer assistant_s.AssistantStorer,
	at_storer assistantthread_s.AssistantThreadStorer,
	am_storer assistantmessage_s.AssistantMessageStorer,
) AssistantMessageController {
	s := &AssistantMessageControllerImpl{
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
	s.Logger.Debug("assistant message controller initialization started...")
	s.Logger.Debug("assistant message controller initialized")
	return s
}
