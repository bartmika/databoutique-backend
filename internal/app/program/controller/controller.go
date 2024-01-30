package controller

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	s3_storage "github.com/bartmika/databoutique-backend/internal/adapter/storage/s3"
	"github.com/bartmika/databoutique-backend/internal/adapter/templatedemailer"
	program_s "github.com/bartmika/databoutique-backend/internal/app/program/datastore"
	user_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config"
	"github.com/bartmika/databoutique-backend/internal/provider/kmutex"
	"github.com/bartmika/databoutique-backend/internal/provider/password"
	"github.com/bartmika/databoutique-backend/internal/provider/uuid"
)

// ProgramController Interface for program business logic controller.
type ProgramController interface {
	Create(ctx context.Context, requestData *ProgramCreateRequestIDO) (*program_s.Program, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*program_s.Program, error)
	UpdateByID(ctx context.Context, requestData *ProgramUpdateRequestIDO) (*program_s.Program, error)
	ListByFilter(ctx context.Context, f *program_s.ProgramPaginationListFilter) (*program_s.ProgramPaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *program_s.ProgramPaginationListFilter) ([]*program_s.ProgramAsSelectOption, error)
	PublicListAsSelectOptionByFilter(ctx context.Context, f *program_s.ProgramPaginationListFilter) ([]*program_s.ProgramAsSelectOption, error)
	ArchiveByID(ctx context.Context, id primitive.ObjectID) (*program_s.Program, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type ProgramControllerImpl struct {
	Config                   *config.Conf
	Logger                   *slog.Logger
	UUID                     uuid.Provider
	S3                       s3_storage.S3Storager
	Password                 password.Provider
	Kmutex                   kmutex.Provider
	DbClient                 *mongo.Client
	UserStorer               user_s.UserStorer
	ProgramStorer program_s.ProgramStorer
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
	program_s program_s.ProgramStorer,
) ProgramController {
	s := &ProgramControllerImpl{
		Config:                   appCfg,
		Logger:                   loggerp,
		UUID:                     uuidp,
		S3:                       s3,
		Password:                 passwordp,
		Kmutex:                   kmux,
		TemplatedEmailer:         temailer,
		DbClient:                 client,
		UserStorer:               usr_storer,
		ProgramStorer: program_s,
	}
	s.Logger.Debug("program controller initialization started...")
	s.Logger.Debug("program controller initialized")
	return s
}
