package controller

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	s3_storage "github.com/bartmika/databoutique-backend/internal/adapter/storage/s3"
	"github.com/bartmika/databoutique-backend/internal/adapter/templatedemailer"
	folderinfo_s "github.com/bartmika/databoutique-backend/internal/app/folderinfo/datastore"
	user_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config"
	"github.com/bartmika/databoutique-backend/internal/provider/kmutex"
	"github.com/bartmika/databoutique-backend/internal/provider/password"
	"github.com/bartmika/databoutique-backend/internal/provider/uuid"
)

// FolderInfoController Interface for folderinfo business logic controller.
type FolderInfoController interface {
	Create(ctx context.Context, requestData *FolderInfoCreateRequestIDO) (*folderinfo_s.FolderInfo, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*folderinfo_s.FolderInfo, error)
	UpdateByID(ctx context.Context, requestData *FolderInfoUpdateRequestIDO) (*folderinfo_s.FolderInfo, error)
	ListByFilter(ctx context.Context, f *folderinfo_s.FolderInfoPaginationListFilter) (*folderinfo_s.FolderInfoPaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *folderinfo_s.FolderInfoPaginationListFilter) ([]*folderinfo_s.FolderInfoAsSelectOption, error)
	PublicListAsSelectOptionByFilter(ctx context.Context, f *folderinfo_s.FolderInfoPaginationListFilter) ([]*folderinfo_s.FolderInfoAsSelectOption, error)
	ArchiveByID(ctx context.Context, id primitive.ObjectID) (*folderinfo_s.FolderInfo, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type FolderInfoControllerImpl struct {
	Config                   *config.Conf
	Logger                   *slog.Logger
	UUID                     uuid.Provider
	S3                       s3_storage.S3Storager
	Password                 password.Provider
	Kmutex                   kmutex.Provider
	DbClient                 *mongo.Client
	UserStorer               user_s.UserStorer
	FolderInfoStorer folderinfo_s.FolderInfoStorer
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
	folderinfo_s folderinfo_s.FolderInfoStorer,
) FolderInfoController {
	s := &FolderInfoControllerImpl{
		Config:                   appCfg,
		Logger:                   loggerp,
		UUID:                     uuidp,
		S3:                       s3,
		Password:                 passwordp,
		Kmutex:                   kmux,
		TemplatedEmailer:         temailer,
		DbClient:                 client,
		UserStorer:               usr_storer,
		FolderInfoStorer: folderinfo_s,
	}
	s.Logger.Debug("folderinfo controller initialization started...")
	s.Logger.Debug("folderinfo controller initialized")
	return s
}
