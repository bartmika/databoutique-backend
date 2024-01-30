package controller

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	s3_storage "github.com/bartmika/databoutique-backend/internal/adapter/storage/s3"
	"github.com/bartmika/databoutique-backend/internal/adapter/templatedemailer"
	programcategory_s "github.com/bartmika/databoutique-backend/internal/app/programcategory/datastore"
	user_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config"
	"github.com/bartmika/databoutique-backend/internal/provider/kmutex"
	"github.com/bartmika/databoutique-backend/internal/provider/password"
	"github.com/bartmika/databoutique-backend/internal/provider/uuid"
)

// ProgramCategoryController Interface for programcategory business logic controller.
type ProgramCategoryController interface {
	Create(ctx context.Context, requestData *ProgramCategoryCreateRequestIDO) (*programcategory_s.ProgramCategory, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*programcategory_s.ProgramCategory, error)
	UpdateByID(ctx context.Context, requestData *ProgramCategoryUpdateRequestIDO) (*programcategory_s.ProgramCategory, error)
	ListByFilter(ctx context.Context, f *programcategory_s.ProgramCategoryPaginationListFilter) (*programcategory_s.ProgramCategoryPaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *programcategory_s.ProgramCategoryPaginationListFilter) ([]*programcategory_s.ProgramCategoryAsSelectOption, error)
	PublicListAsSelectOptionByFilter(ctx context.Context, f *programcategory_s.ProgramCategoryPaginationListFilter) ([]*programcategory_s.ProgramCategoryAsSelectOption, error)
	ArchiveByID(ctx context.Context, id primitive.ObjectID) (*programcategory_s.ProgramCategory, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type ProgramCategoryControllerImpl struct {
	Config                   *config.Conf
	Logger                   *slog.Logger
	UUID                     uuid.Provider
	S3                       s3_storage.S3Storager
	Password                 password.Provider
	Kmutex                   kmutex.Provider
	DbClient                 *mongo.Client
	UserStorer               user_s.UserStorer
	ProgramCategoryStorer programcategory_s.ProgramCategoryStorer
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
	programcategory_s programcategory_s.ProgramCategoryStorer,
) ProgramCategoryController {
	s := &ProgramCategoryControllerImpl{
		Config:                   appCfg,
		Logger:                   loggerp,
		UUID:                     uuidp,
		S3:                       s3,
		Password:                 passwordp,
		Kmutex:                   kmux,
		TemplatedEmailer:         temailer,
		DbClient:                 client,
		UserStorer:               usr_storer,
		ProgramCategoryStorer: programcategory_s,
	}
	s.Logger.Debug("programcategory controller initialization started...")
	s.Logger.Debug("programcategory controller initialized")
	return s
}
