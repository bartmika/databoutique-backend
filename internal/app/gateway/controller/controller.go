package controller

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/bartmika/databoutique-backend/internal/adapter/cache/mongodbcache"
	"github.com/bartmika/databoutique-backend/internal/adapter/templatedemailer"
	gateway_s "github.com/bartmika/databoutique-backend/internal/app/gateway/datastore"
	howhear_s "github.com/bartmika/databoutique-backend/internal/app/howhear/datastore"
	tenant_s "github.com/bartmika/databoutique-backend/internal/app/tenant/datastore"
	user_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config"
	"github.com/bartmika/databoutique-backend/internal/provider/jwt"
	"github.com/bartmika/databoutique-backend/internal/provider/kmutex"
	"github.com/bartmika/databoutique-backend/internal/provider/password"
	"github.com/bartmika/databoutique-backend/internal/provider/uuid"
)

type GatewayController interface {
	UserRegister(ctx context.Context, req *UserRegisterRequestIDO) (*gateway_s.LoginResponseIDO, error)
	Login(ctx context.Context, email, password string) (*gateway_s.LoginResponseIDO, error)
	GetUserBySessionID(ctx context.Context, sessionID string) (*user_s.User, error)
	RefreshToken(ctx context.Context, value string) (*user_s.User, string, time.Time, string, time.Time, error)
	Logout(ctx context.Context) error
	ForgotPassword(ctx context.Context, email string) error
	PasswordReset(ctx context.Context, code string, password string) error
	Profile(ctx context.Context) (*user_s.User, error)
	ProfileUpdate(ctx context.Context, nu *user_s.User) error
	ProfileChangePassword(ctx context.Context, req *ProfileChangePasswordRequestIDO) error
	ExecutiveVisitsTenant(ctx context.Context, req *ExecutiveVisitsTenantRequest) error
	Dashboard(ctx context.Context) (*DashboardResponseIDO, error)
}

type GatewayControllerImpl struct {
	Config                   *config.Conf
	Logger                   *slog.Logger
	UUID                     uuid.Provider
	JWT                      jwt.Provider
	Password                 password.Provider
	Kmutex                   kmutex.Provider
	DbClient                 *mongo.Client
	Cache                    mongodbcache.Cacher
	TemplatedEmailer         templatedemailer.TemplatedEmailer
	UserStorer               user_s.UserStorer
	TenantStorer             tenant_s.TenantStorer
	HowHearAboutUsItemStorer howhear_s.HowHearAboutUsItemStorer
}

func NewController(
	appCfg *config.Conf,
	loggerp *slog.Logger,
	uuidp uuid.Provider,
	jwtp jwt.Provider,
	passwordp password.Provider,
	kmux kmutex.Provider,
	cache mongodbcache.Cacher,
	te templatedemailer.TemplatedEmailer,
	client *mongo.Client,
	usr_storer user_s.UserStorer,
	org_storer tenant_s.TenantStorer,
	howhear_s howhear_s.HowHearAboutUsItemStorer,
) GatewayController {
	// loggerp.Debug("gateway controller initialization started...") // For debugging purposes only.
	s := &GatewayControllerImpl{
		Config:                   appCfg,
		Logger:                   loggerp,
		UUID:                     uuidp,
		JWT:                      jwtp,
		Kmutex:                   kmux,
		Password:                 passwordp,
		DbClient:                 client,
		Cache:                    cache,
		TemplatedEmailer:         te,
		UserStorer:               usr_storer,
		TenantStorer:             org_storer,
		HowHearAboutUsItemStorer: howhear_s,
	}
	// s.Logger.Debug("gateway controller initialized")
	if err := s.initializeAccounts(context.Background()); err != nil {
		log.Fatalf("failed initializing accounts %v", err)
	}
	if err := s.initializeHowHearAboutUsItems(context.Background()); err != nil {
		log.Fatalf("failed initializing accounts %v", err)
	}
	return s
}

func (impl *GatewayControllerImpl) GetUserBySessionID(ctx context.Context, sessionID string) (*user_s.User, error) {
	userBytes, err := impl.Cache.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if userBytes == nil {
		impl.Logger.Warn("record not found")
		return nil, errors.New("record not found")
	}
	var user user_s.User
	err = json.Unmarshal(userBytes, &user)
	if err != nil {
		impl.Logger.Error("unmarshalling failed", slog.Any("err", err))
		return nil, err
	}
	return &user, nil
}
