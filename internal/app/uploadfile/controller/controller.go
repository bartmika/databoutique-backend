package controller

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	mg "github.com/bartmika/databoutique-backend/internal/adapter/emailer/mailgun"
	s3_storage "github.com/bartmika/databoutique-backend/internal/adapter/storage/s3"
	tenant_s "github.com/bartmika/databoutique-backend/internal/app/tenant/datastore"
	uploaddirectory_s "github.com/bartmika/databoutique-backend/internal/app/uploaddirectory/datastore"
	uploadfile_ds "github.com/bartmika/databoutique-backend/internal/app/uploadfile/datastore"
	user_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config"
	"github.com/bartmika/databoutique-backend/internal/provider/uuid"
)

// UploadFileController Interface for uploadfile business logic controller.
type UploadFileController interface {
	Create(ctx context.Context, req *UploadFileCreateRequestIDO) (*uploadfile_ds.UploadFile, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*uploadfile_ds.UploadFile, error)
	UpdateByID(ctx context.Context, ns *UploadFileUpdateRequestIDO) (*uploadfile_ds.UploadFile, error)
	ListByFilter(ctx context.Context, f *uploadfile_ds.UploadFilePaginationListFilter) (*uploadfile_ds.UploadFilePaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *uploadfile_ds.UploadFilePaginationListFilter) ([]*uploadfile_ds.UploadFileAsSelectOption, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	PermanentlyDeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type UploadFileControllerImpl struct {
	Config                *config.Conf
	Logger                *slog.Logger
	UUID                  uuid.Provider
	S3                    s3_storage.S3Storager
	Emailer               mg.Emailer
	DbClient              *mongo.Client
	TenantStorer          tenant_s.TenantStorer
	UploadFileStorer      uploadfile_ds.UploadFileStorer
	UploadDirectoryStorer uploaddirectory_s.UploadDirectoryStorer
	UserStorer            user_s.UserStorer
}

func NewController(
	appCfg *config.Conf,
	loggerp *slog.Logger,
	uuidp uuid.Provider,
	s3 s3_storage.S3Storager,
	client *mongo.Client,
	emailer mg.Emailer,
	t_storer tenant_s.TenantStorer,
	uploaddirectory_s uploaddirectory_s.UploadDirectoryStorer,
	org_storer uploadfile_ds.UploadFileStorer,
	usr_storer user_s.UserStorer,
) UploadFileController {
	s := &UploadFileControllerImpl{
		Config:                appCfg,
		Logger:                loggerp,
		UUID:                  uuidp,
		S3:                    s3,
		Emailer:               emailer,
		DbClient:              client,
		TenantStorer:          t_storer,
		UploadDirectoryStorer: uploaddirectory_s,
		UploadFileStorer:      org_storer,
		UserStorer:            usr_storer,
	}
	s.Logger.Debug("uploadfile controller initialization started...")
	s.Logger.Debug("uploadfile controller initialized")
	return s
}
