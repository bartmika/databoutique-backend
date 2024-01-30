//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/bartmika/databoutique-backend/internal/config"

	"github.com/bartmika/databoutique-backend/internal/adapter/cache/mongodbcache"
	"github.com/bartmika/databoutique-backend/internal/adapter/emailer/mailgun"
	s3_storage "github.com/bartmika/databoutique-backend/internal/adapter/storage/s3"
	"github.com/bartmika/databoutique-backend/internal/adapter/templatedemailer"

	"github.com/bartmika/databoutique-backend/internal/provider/jwt"
	"github.com/bartmika/databoutique-backend/internal/provider/kmutex"
	"github.com/bartmika/databoutique-backend/internal/provider/logger"
	"github.com/bartmika/databoutique-backend/internal/provider/mongodb"
	"github.com/bartmika/databoutique-backend/internal/provider/password"

	"github.com/bartmika/databoutique-backend/internal/provider/time"
	"github.com/bartmika/databoutique-backend/internal/provider/uuid"

	ds_assistant "github.com/bartmika/databoutique-backend/internal/app/assistant/datastore"
	ds_assistantfile "github.com/bartmika/databoutique-backend/internal/app/assistantfile/datastore"
	ds_assistantmessage "github.com/bartmika/databoutique-backend/internal/app/assistantmessage/datastore"
	ds_assistantthread "github.com/bartmika/databoutique-backend/internal/app/assistantthread/datastore"
	ds_attachment "github.com/bartmika/databoutique-backend/internal/app/attachment/datastore"
	ds_howhear "github.com/bartmika/databoutique-backend/internal/app/howhear/datastore"
	ds_tenant "github.com/bartmika/databoutique-backend/internal/app/tenant/datastore"
	ds_user "github.com/bartmika/databoutique-backend/internal/app/user/datastore"

	uc_assistant "github.com/bartmika/databoutique-backend/internal/app/assistant/controller"
	uc_assistantfile "github.com/bartmika/databoutique-backend/internal/app/assistantfile/controller"
	uc_assistantmessage "github.com/bartmika/databoutique-backend/internal/app/assistantmessage/controller"
	uc_assistantthread "github.com/bartmika/databoutique-backend/internal/app/assistantthread/controller"
	uc_attachment "github.com/bartmika/databoutique-backend/internal/app/attachment/controller"
	uc_gateway "github.com/bartmika/databoutique-backend/internal/app/gateway/controller"
	uc_howhear "github.com/bartmika/databoutique-backend/internal/app/howhear/controller"
	uc_tenant "github.com/bartmika/databoutique-backend/internal/app/tenant/controller"
	uc_user "github.com/bartmika/databoutique-backend/internal/app/user/controller"

	http_assistant "github.com/bartmika/databoutique-backend/internal/app/assistant/httptransport"
	http_assistantfile "github.com/bartmika/databoutique-backend/internal/app/assistantfile/httptransport"
	http_assistantmessage "github.com/bartmika/databoutique-backend/internal/app/assistantmessage/httptransport"
	http_assistantthread "github.com/bartmika/databoutique-backend/internal/app/assistantthread/httptransport"
	http_attachment "github.com/bartmika/databoutique-backend/internal/app/attachment/httptransport"
	http_gate "github.com/bartmika/databoutique-backend/internal/app/gateway/httptransport"
	http_howhear "github.com/bartmika/databoutique-backend/internal/app/howhear/httptransport"
	http_tenant "github.com/bartmika/databoutique-backend/internal/app/tenant/httptransport"
	http_user "github.com/bartmika/databoutique-backend/internal/app/user/httptransport"

	http "github.com/bartmika/databoutique-backend/internal/inputport/httptransport"
	http_middleware "github.com/bartmika/databoutique-backend/internal/inputport/httptransport/middleware"
)

func InitializeEvent() Application {
	// Our application is dependent on the following Golang packages. We need to
	// provide them to Google wire so it can sort out the dependency injection
	// at compile time.
	wire.Build(
		// CONFIGURATION SECTION
		config.New,

		// PROVIDERS SECTION
		logger.NewProvider,
		uuid.NewProvider,
		time.NewProvider,
		jwt.NewProvider,
		password.NewProvider,
		kmutex.NewProvider,
		mongodb.NewProvider,

		// TODO
		mailgun.NewEmailer,
		templatedemailer.NewTemplatedEmailer,
		mongodbcache.NewCache,
		s3_storage.NewStorage,

		// ADAPTERS SECTION

		// DATASTORE
		ds_tenant.NewDatastore,
		ds_user.NewDatastore,
		ds_howhear.NewDatastore,
		ds_attachment.NewDatastore,
		ds_assistantfile.NewDatastore,
		ds_assistant.NewDatastore,
		ds_assistantthread.NewDatastore,
		ds_assistantmessage.NewDatastore,

		// USECASE
		uc_tenant.NewController,
		uc_gateway.NewController,
		uc_user.NewController,
		uc_howhear.NewController,
		uc_attachment.NewController,
		uc_assistantfile.NewController,
		uc_assistant.NewController,
		uc_assistantthread.NewController,
		uc_assistantmessage.NewController,

		// HTTP TRANSPORT SECTION
		http_tenant.NewHandler,
		http_gate.NewHandler,
		http_user.NewHandler,
		http_howhear.NewHandler,
		http_attachment.NewHandler,
		http_assistantfile.NewHandler,
		http_assistant.NewHandler,
		http_assistantthread.NewHandler,
		http_assistantmessage.NewHandler,

		// INPUT PORT SECTION
		http_middleware.NewMiddleware,
		http.NewInputPort,

		// APP
		NewApplication)
	return Application{}
}
