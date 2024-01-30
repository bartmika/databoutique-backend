package controller

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	mg "github.com/bartmika/databoutique-backend/internal/adapter/emailer/mailgun"
	s3_storage "github.com/bartmika/databoutique-backend/internal/adapter/storage/s3"
	fileinfo_s "github.com/bartmika/databoutique-backend/internal/app/fileinfo/datastore"
	domain "github.com/bartmika/databoutique-backend/internal/app/fileinfo/datastore"
	tenant_s "github.com/bartmika/databoutique-backend/internal/app/tenant/datastore"
	user_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config"
	"github.com/bartmika/databoutique-backend/internal/provider/uuid"
)

// FileInfoController Interface for fileinfo business logic controller.
type FileInfoController interface {
	Create(ctx context.Context, req *FileInfoCreateRequestIDO) (*domain.FileInfo, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.FileInfo, error)
	UpdateByID(ctx context.Context, ns *FileInfoUpdateRequestIDO) (*domain.FileInfo, error)
	ListByFilter(ctx context.Context, f *domain.FileInfoPaginationListFilter) (*domain.FileInfoPaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *domain.FileInfoPaginationListFilter) ([]*domain.FileInfoAsSelectOption, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	PermanentlyDeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type FileInfoControllerImpl struct {
	Config         *config.Conf
	Logger         *slog.Logger
	UUID           uuid.Provider
	S3             s3_storage.S3Storager
	Emailer        mg.Emailer
	DbClient       *mongo.Client
	TenantStorer   tenant_s.TenantStorer
	FileInfoStorer fileinfo_s.FileInfoStorer
	UserStorer     user_s.UserStorer
}

func NewController(
	appCfg *config.Conf,
	loggerp *slog.Logger,
	uuidp uuid.Provider,
	s3 s3_storage.S3Storager,
	client *mongo.Client,
	emailer mg.Emailer,
	t_storer tenant_s.TenantStorer,
	org_storer fileinfo_s.FileInfoStorer,
	usr_storer user_s.UserStorer,
) FileInfoController {
	s := &FileInfoControllerImpl{
		Config:         appCfg,
		Logger:         loggerp,
		UUID:           uuidp,
		S3:             s3,
		Emailer:        emailer,
		DbClient:       client,
		TenantStorer:   t_storer,
		FileInfoStorer: org_storer,
		UserStorer:     usr_storer,
	}
	s.Logger.Debug("fileinfo controller initialization started...")
	s.Logger.Debug("fileinfo controller initialized")
	return s
}
