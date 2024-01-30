package controller

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	s3_storage "github.com/bartmika/databoutique-backend/internal/adapter/storage/s3"
	"github.com/bartmika/databoutique-backend/internal/adapter/templatedemailer"
	uploaddirectory_s "github.com/bartmika/databoutique-backend/internal/app/uploaddirectory/datastore"
	user_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config"
	"github.com/bartmika/databoutique-backend/internal/provider/kmutex"
	"github.com/bartmika/databoutique-backend/internal/provider/password"
	"github.com/bartmika/databoutique-backend/internal/provider/uuid"
)

// UploadDirectoryController Interface for uploaddirectory business logic controller.
type UploadDirectoryController interface {
	Create(ctx context.Context, requestData *UploadDirectoryCreateRequestIDO) (*uploaddirectory_s.UploadDirectory, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*uploaddirectory_s.UploadDirectory, error)
	UpdateByID(ctx context.Context, requestData *UploadDirectoryUpdateRequestIDO) (*uploaddirectory_s.UploadDirectory, error)
	ListByFilter(ctx context.Context, f *uploaddirectory_s.UploadDirectoryPaginationListFilter) (*uploaddirectory_s.UploadDirectoryPaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *uploaddirectory_s.UploadDirectoryPaginationListFilter) ([]*uploaddirectory_s.UploadDirectoryAsSelectOption, error)
	PublicListAsSelectOptionByFilter(ctx context.Context, f *uploaddirectory_s.UploadDirectoryPaginationListFilter) ([]*uploaddirectory_s.UploadDirectoryAsSelectOption, error)
	ArchiveByID(ctx context.Context, id primitive.ObjectID) (*uploaddirectory_s.UploadDirectory, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type UploadDirectoryControllerImpl struct {
	Config                   *config.Conf
	Logger                   *slog.Logger
	UUID                     uuid.Provider
	S3                       s3_storage.S3Storager
	Password                 password.Provider
	Kmutex                   kmutex.Provider
	DbClient                 *mongo.Client
	UserStorer               user_s.UserStorer
	UploadDirectoryStorer uploaddirectory_s.UploadDirectoryStorer
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
	uploaddirectory_s uploaddirectory_s.UploadDirectoryStorer,
) UploadDirectoryController {
	s := &UploadDirectoryControllerImpl{
		Config:                   appCfg,
		Logger:                   loggerp,
		UUID:                     uuidp,
		S3:                       s3,
		Password:                 passwordp,
		Kmutex:                   kmux,
		TemplatedEmailer:         temailer,
		DbClient:                 client,
		UserStorer:               usr_storer,
		UploadDirectoryStorer: uploaddirectory_s,
	}
	s.Logger.Debug("uploaddirectory controller initialization started...")
	s.Logger.Debug("uploaddirectory controller initialized")
	return s
}
