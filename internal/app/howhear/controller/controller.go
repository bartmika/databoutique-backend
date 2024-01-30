package controller

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	s3_storage "github.com/bartmika/databoutique-backend/internal/adapter/storage/s3"
	"github.com/bartmika/databoutique-backend/internal/adapter/templatedemailer"
	howhear_s "github.com/bartmika/databoutique-backend/internal/app/howhear/datastore"
	user_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config"
	"github.com/bartmika/databoutique-backend/internal/provider/kmutex"
	"github.com/bartmika/databoutique-backend/internal/provider/password"
	"github.com/bartmika/databoutique-backend/internal/provider/uuid"
)

// HowHearAboutUsItemController Interface for howhear business logic controller.
type HowHearAboutUsItemController interface {
	Create(ctx context.Context, requestData *HowHearAboutUsItemCreateRequestIDO) (*howhear_s.HowHearAboutUsItem, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*howhear_s.HowHearAboutUsItem, error)
	UpdateByID(ctx context.Context, requestData *HowHearAboutUsItemUpdateRequestIDO) (*howhear_s.HowHearAboutUsItem, error)
	ListByFilter(ctx context.Context, f *howhear_s.HowHearAboutUsItemPaginationListFilter) (*howhear_s.HowHearAboutUsItemPaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *howhear_s.HowHearAboutUsItemPaginationListFilter) ([]*howhear_s.HowHearAboutUsItemAsSelectOption, error)
	PublicListAsSelectOptionByFilter(ctx context.Context, f *howhear_s.HowHearAboutUsItemPaginationListFilter) ([]*howhear_s.HowHearAboutUsItemAsSelectOption, error)
	ArchiveByID(ctx context.Context, id primitive.ObjectID) (*howhear_s.HowHearAboutUsItem, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type HowHearAboutUsItemControllerImpl struct {
	Config                   *config.Conf
	Logger                   *slog.Logger
	UUID                     uuid.Provider
	S3                       s3_storage.S3Storager
	Password                 password.Provider
	Kmutex                   kmutex.Provider
	DbClient                 *mongo.Client
	UserStorer               user_s.UserStorer
	HowHearAboutUsItemStorer howhear_s.HowHearAboutUsItemStorer
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
	howhear_s howhear_s.HowHearAboutUsItemStorer,
) HowHearAboutUsItemController {
	s := &HowHearAboutUsItemControllerImpl{
		Config:                   appCfg,
		Logger:                   loggerp,
		UUID:                     uuidp,
		S3:                       s3,
		Password:                 passwordp,
		Kmutex:                   kmux,
		TemplatedEmailer:         temailer,
		DbClient:                 client,
		UserStorer:               usr_storer,
		HowHearAboutUsItemStorer: howhear_s,
	}
	s.Logger.Debug("howhear controller initialization started...")
	s.Logger.Debug("howhear controller initialized")
	return s
}
