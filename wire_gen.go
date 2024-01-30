// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/bartmika/databoutique-backend/internal/adapter/cache/mongodbcache"
	"github.com/bartmika/databoutique-backend/internal/adapter/emailer/mailgun"
	"github.com/bartmika/databoutique-backend/internal/adapter/storage/s3"
	"github.com/bartmika/databoutique-backend/internal/adapter/templatedemailer"
	controller7 "github.com/bartmika/databoutique-backend/internal/app/assistant/controller"
	datastore6 "github.com/bartmika/databoutique-backend/internal/app/assistant/datastore"
	httptransport7 "github.com/bartmika/databoutique-backend/internal/app/assistant/httptransport"
	controller6 "github.com/bartmika/databoutique-backend/internal/app/assistantfile/controller"
	datastore5 "github.com/bartmika/databoutique-backend/internal/app/assistantfile/datastore"
	httptransport6 "github.com/bartmika/databoutique-backend/internal/app/assistantfile/httptransport"
	controller9 "github.com/bartmika/databoutique-backend/internal/app/assistantmessage/controller"
	datastore8 "github.com/bartmika/databoutique-backend/internal/app/assistantmessage/datastore"
	httptransport9 "github.com/bartmika/databoutique-backend/internal/app/assistantmessage/httptransport"
	controller8 "github.com/bartmika/databoutique-backend/internal/app/assistantthread/controller"
	datastore7 "github.com/bartmika/databoutique-backend/internal/app/assistantthread/datastore"
	httptransport8 "github.com/bartmika/databoutique-backend/internal/app/assistantthread/httptransport"
	controller5 "github.com/bartmika/databoutique-backend/internal/app/attachment/controller"
	datastore4 "github.com/bartmika/databoutique-backend/internal/app/attachment/datastore"
	httptransport5 "github.com/bartmika/databoutique-backend/internal/app/attachment/httptransport"
	"github.com/bartmika/databoutique-backend/internal/app/gateway/controller"
	httptransport2 "github.com/bartmika/databoutique-backend/internal/app/gateway/httptransport"
	controller4 "github.com/bartmika/databoutique-backend/internal/app/howhear/controller"
	datastore3 "github.com/bartmika/databoutique-backend/internal/app/howhear/datastore"
	httptransport4 "github.com/bartmika/databoutique-backend/internal/app/howhear/httptransport"
	controller11 "github.com/bartmika/databoutique-backend/internal/app/program/controller"
	datastore10 "github.com/bartmika/databoutique-backend/internal/app/program/datastore"
	httptransport11 "github.com/bartmika/databoutique-backend/internal/app/program/httptransport"
	controller10 "github.com/bartmika/databoutique-backend/internal/app/programcategory/controller"
	datastore9 "github.com/bartmika/databoutique-backend/internal/app/programcategory/datastore"
	httptransport10 "github.com/bartmika/databoutique-backend/internal/app/programcategory/httptransport"
	controller2 "github.com/bartmika/databoutique-backend/internal/app/tenant/controller"
	datastore2 "github.com/bartmika/databoutique-backend/internal/app/tenant/datastore"
	"github.com/bartmika/databoutique-backend/internal/app/tenant/httptransport"
	controller3 "github.com/bartmika/databoutique-backend/internal/app/user/controller"
	"github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	httptransport3 "github.com/bartmika/databoutique-backend/internal/app/user/httptransport"
	"github.com/bartmika/databoutique-backend/internal/config"
	httptransport12 "github.com/bartmika/databoutique-backend/internal/inputport/httptransport"
	"github.com/bartmika/databoutique-backend/internal/inputport/httptransport/middleware"
	"github.com/bartmika/databoutique-backend/internal/provider/jwt"
	"github.com/bartmika/databoutique-backend/internal/provider/kmutex"
	"github.com/bartmika/databoutique-backend/internal/provider/logger"
	"github.com/bartmika/databoutique-backend/internal/provider/mongodb"
	"github.com/bartmika/databoutique-backend/internal/provider/password"
	"github.com/bartmika/databoutique-backend/internal/provider/time"
	"github.com/bartmika/databoutique-backend/internal/provider/uuid"
)

import (
	_ "go.uber.org/automaxprocs"
	_ "time/tzdata"
)

// Injectors from wire.go:

func InitializeEvent() Application {
	slogLogger := logger.NewProvider()
	conf := config.New()
	provider := uuid.NewProvider()
	timeProvider := time.NewProvider()
	jwtProvider := jwt.NewProvider(conf)
	passwordProvider := password.NewProvider()
	kmutexProvider := kmutex.NewProvider()
	client := mongodb.NewProvider(conf, slogLogger)
	cacher := mongodbcache.NewCache(conf, slogLogger, client)
	emailer := mailgun.NewEmailer(conf, slogLogger, provider)
	templatedEmailer := templatedemailer.NewTemplatedEmailer(conf, slogLogger, provider, emailer)
	userStorer := datastore.NewDatastore(conf, slogLogger, client)
	tenantStorer := datastore2.NewDatastore(conf, slogLogger, client)
	howHearAboutUsItemStorer := datastore3.NewDatastore(conf, slogLogger, client)
	gatewayController := controller.NewController(conf, slogLogger, provider, jwtProvider, passwordProvider, kmutexProvider, cacher, templatedEmailer, client, userStorer, tenantStorer, howHearAboutUsItemStorer)
	middlewareMiddleware := middleware.NewMiddleware(conf, slogLogger, provider, timeProvider, jwtProvider, gatewayController)
	s3Storager := s3.NewStorage(conf, slogLogger, provider)
	tenantController := controller2.NewController(conf, slogLogger, provider, kmutexProvider, s3Storager, emailer, client, tenantStorer)
	handler := httptransport.NewHandler(slogLogger, tenantController)
	httptransportHandler := httptransport2.NewHandler(slogLogger, gatewayController)
	userController := controller3.NewController(conf, slogLogger, provider, passwordProvider, kmutexProvider, client, tenantStorer, userStorer, templatedEmailer)
	handler2 := httptransport3.NewHandler(slogLogger, userController)
	howHearAboutUsItemController := controller4.NewController(conf, slogLogger, provider, s3Storager, passwordProvider, kmutexProvider, templatedEmailer, client, userStorer, howHearAboutUsItemStorer)
	handler3 := httptransport4.NewHandler(slogLogger, howHearAboutUsItemController)
	attachmentStorer := datastore4.NewDatastore(conf, slogLogger, client)
	attachmentController := controller5.NewController(conf, slogLogger, provider, s3Storager, client, emailer, attachmentStorer, userStorer)
	handler4 := httptransport5.NewHandler(attachmentController)
	assistantFileStorer := datastore5.NewDatastore(conf, slogLogger, client)
	assistantFileController := controller6.NewController(conf, slogLogger, provider, s3Storager, client, emailer, tenantStorer, assistantFileStorer, userStorer)
	handler5 := httptransport6.NewHandler(assistantFileController)
	assistantStorer := datastore6.NewDatastore(conf, slogLogger, client)
	assistantController := controller7.NewController(conf, slogLogger, provider, s3Storager, passwordProvider, kmutexProvider, client, templatedEmailer, tenantStorer, userStorer, assistantFileStorer, assistantStorer)
	handler6 := httptransport7.NewHandler(slogLogger, assistantController)
	assistantThreadStorer := datastore7.NewDatastore(conf, slogLogger, client)
	assistantMessageStorer := datastore8.NewDatastore(conf, slogLogger, client)
	assistantThreadController := controller8.NewController(conf, slogLogger, provider, s3Storager, passwordProvider, kmutexProvider, client, templatedEmailer, tenantStorer, userStorer, assistantFileStorer, assistantStorer, assistantThreadStorer, assistantMessageStorer)
	handler7 := httptransport8.NewHandler(slogLogger, assistantThreadController)
	assistantMessageController := controller9.NewController(conf, slogLogger, provider, s3Storager, passwordProvider, kmutexProvider, templatedEmailer, client, tenantStorer, userStorer, assistantFileStorer, assistantStorer, assistantThreadStorer, assistantMessageStorer)
	handler8 := httptransport9.NewHandler(slogLogger, assistantMessageController)
	programCategoryStorer := datastore9.NewDatastore(conf, slogLogger, client)
	programCategoryController := controller10.NewController(conf, slogLogger, provider, s3Storager, passwordProvider, kmutexProvider, templatedEmailer, client, userStorer, programCategoryStorer)
	handler9 := httptransport10.NewHandler(slogLogger, programCategoryController)
	programStorer := datastore10.NewDatastore(conf, slogLogger, client)
	programController := controller11.NewController(conf, slogLogger, provider, s3Storager, passwordProvider, kmutexProvider, templatedEmailer, client, userStorer, programStorer)
	handler10 := httptransport11.NewHandler(slogLogger, programController)
	inputPortServer := httptransport12.NewInputPort(conf, slogLogger, middlewareMiddleware, handler, httptransportHandler, handler2, handler3, handler4, handler5, handler6, handler7, handler8, handler9, handler10)
	application := NewApplication(slogLogger, inputPortServer)
	return application
}
