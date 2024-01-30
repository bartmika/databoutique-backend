package httptransport

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/rs/cors"

	assistant "github.com/bartmika/databoutique-backend/internal/app/assistant/httptransport"
	assistantfile "github.com/bartmika/databoutique-backend/internal/app/assistantfile/httptransport"
	assistantmessage "github.com/bartmika/databoutique-backend/internal/app/assistantmessage/httptransport"
	assistantthread "github.com/bartmika/databoutique-backend/internal/app/assistantthread/httptransport"
	attachment "github.com/bartmika/databoutique-backend/internal/app/attachment/httptransport"
	gateway "github.com/bartmika/databoutique-backend/internal/app/gateway/httptransport"
	howhear "github.com/bartmika/databoutique-backend/internal/app/howhear/httptransport"
	program "github.com/bartmika/databoutique-backend/internal/app/program/httptransport"
	programcategory "github.com/bartmika/databoutique-backend/internal/app/programcategory/httptransport"
	tenant "github.com/bartmika/databoutique-backend/internal/app/tenant/httptransport"
	user "github.com/bartmika/databoutique-backend/internal/app/user/httptransport"
	"github.com/bartmika/databoutique-backend/internal/config"
	"github.com/bartmika/databoutique-backend/internal/inputport/httptransport/middleware"
)

type InputPortServer interface {
	Run()
	Shutdown()
}

type httpTransportInputPort struct {
	Config           *config.Conf
	Logger           *slog.Logger
	Server           *http.Server
	Middleware       middleware.Middleware
	Tenant           *tenant.Handler
	Gateway          *gateway.Handler
	User             *user.Handler
	HowHear          *howhear.Handler
	Attachment       *attachment.Handler
	AssistantFile    *assistantfile.Handler
	Assistant        *assistant.Handler
	AssistantThread  *assistantthread.Handler
	AssistantMessage *assistantmessage.Handler
	ProgramCategory  *programcategory.Handler
	Program          *program.Handler
}

func NewInputPort(
	configp *config.Conf,
	loggerp *slog.Logger,
	mid middleware.Middleware,
	org *tenant.Handler,
	gate *gateway.Handler,
	user *user.Handler,
	howhear *howhear.Handler,
	att *attachment.Handler,
	af *assistantfile.Handler,
	assistant *assistant.Handler,
	at *assistantthread.Handler,
	am *assistantmessage.Handler,
	pc *programcategory.Handler,
	prog *program.Handler,
) InputPortServer {
	// Initialize the ServeMux.
	mux := http.NewServeMux()

	// cors.Default() setup the middleware with default options being
	// all origins accepted with simple methods (GET, POST). See
	// documentation via `https://github.com/rs/cors` for more options.
	handler := cors.AllowAll().Handler(mux)

	// Bind the HTTP server to the assigned address and port.
	addr := fmt.Sprintf("%s:%s", configp.AppServer.IP, configp.AppServer.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	// Create our HTTP server controller.
	p := &httpTransportInputPort{
		Config:           configp,
		Logger:           loggerp,
		Middleware:       mid,
		Tenant:           org,
		Gateway:          gate,
		User:             user,
		HowHear:          howhear,
		Attachment:       att,
		AssistantFile:    af,
		Assistant:        assistant,
		AssistantThread:  at,
		AssistantMessage: am,
		ProgramCategory:  pc,
		Program:          prog,
		Server:           srv,
	}

	// Attach the HTTP server controller to the ServerMux.
	mux.HandleFunc("/", mid.Attach(p.HandleRequests))

	return p
}

func (port *httpTransportInputPort) Run() {
	port.Logger.Info("HTTP server running")
	if err := port.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		port.Logger.Error("listen failed", slog.Any("error", err))

		// DEVELOPERS NOTE: We terminate app here b/c dependency injection not allowed to fail, so fail here at startup of app.
		panic("failed running")
	}
}

func (port *httpTransportInputPort) Shutdown() {
	port.Logger.Info("HTTP server shutdown")
}

func (port *httpTransportInputPort) HandleRequests(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get our URL paths which are slash-seperated.
	ctx := r.Context()
	p := ctx.Value("url_split").([]string)
	n := len(p)
	port.Logger.Debug("Handling request",
		slog.Int("n", n),
		slog.String("m", r.Method),
		slog.Any("p", p),
	)

	switch {
	// --- GATEWAY & PROFILE --- //
	case n == 3 && p[1] == "v1" && p[2] == "health-check" && r.Method == http.MethodGet:
		port.Gateway.HealthCheck(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "version" && r.Method == http.MethodGet:
		port.Gateway.Version(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "greeting" && r.Method == http.MethodPost:
		port.Gateway.Greet(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "login" && r.Method == http.MethodPost:
		port.Gateway.Login(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "register" && r.Method == http.MethodPost:
		port.Gateway.UserRegister(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "refresh-token" && r.Method == http.MethodPost:
		port.Gateway.RefreshToken(w, r)
	// case n == 3 && p[1] == "v1" && p[2] == "verify" && r.Method == http.MethodPost:
	// 	port.Gateway.Verify(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "logout" && r.Method == http.MethodPost:
		port.Gateway.Logout(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "profile" && r.Method == http.MethodGet:
		port.Gateway.Profile(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "profile" && r.Method == http.MethodPut:
		port.Gateway.ProfileUpdate(w, r)
	case n == 4 && p[1] == "v1" && p[2] == "profile" && p[3] == "change-password" && r.Method == http.MethodPut:
		port.Gateway.ProfileChangePassword(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "forgot-password" && r.Method == http.MethodPost:
		port.Gateway.ForgotPassword(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "password-reset" && r.Method == http.MethodPost:
		port.Gateway.PasswordReset(w, r)
		// case n == 3 && p[1] == "v1" && p[2] == "profile" && r.Method == http.MethodGet:
	case n == 3 && p[1] == "v1" && p[2] == "executive-visit-tenant" && r.Method == http.MethodPost:
		port.Gateway.ExecutiveVisitsTenant(w, r)

	// // --- DASHBOARD --- //
	// case n == 3 && p[1] == "v1" && p[2] == "dashboard" && r.Method == http.MethodGet:
	// 	port.Dashboard.Dashboard(w, r)
	// 	// ...

	// --- ORGANIZATION --- //
	case n == 3 && p[1] == "v1" && p[2] == "tenants" && r.Method == http.MethodGet:
		port.Tenant.List(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "tenants" && r.Method == http.MethodPost:
		port.Tenant.Create(w, r)
	case n == 4 && p[1] == "v1" && p[2] == "tenant" && r.Method == http.MethodGet:
		port.Tenant.GetByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "tenant" && r.Method == http.MethodPut:
		port.Tenant.UpdateByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "tenant" && r.Method == http.MethodDelete:
		port.Tenant.DeleteByID(w, r, p[3])
	case n == 5 && p[1] == "v1" && p[2] == "tenants" && p[3] == "operation" && p[4] == "create-comment" && r.Method == http.MethodPost:
		port.Tenant.OperationCreateComment(w, r)
	case n == 4 && p[1] == "v1" && p[2] == "tenants" && p[3] == "select-options" && r.Method == http.MethodGet:
		port.Tenant.ListAsSelectOptionByFilter(w, r)

	// --- PROGRAM CATEGORY --- //
	case n == 3 && p[1] == "v1" && p[2] == "program-categories" && r.Method == http.MethodGet:
		port.ProgramCategory.List(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "program-categories" && r.Method == http.MethodPost:
		port.ProgramCategory.Create(w, r)
	case n == 4 && p[1] == "v1" && p[2] == "program-category" && r.Method == http.MethodGet:
		port.ProgramCategory.GetByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "program-category" && r.Method == http.MethodPut:
		port.ProgramCategory.UpdateByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "program-category" && r.Method == http.MethodDelete:
		port.ProgramCategory.DeleteByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "program-categories" && p[3] == "select-options" && r.Method == http.MethodGet:
		port.ProgramCategory.ListAsSelectOptionByFilter(w, r)

	// --- PROGRAM --- //
	case n == 3 && p[1] == "v1" && p[2] == "programs" && r.Method == http.MethodGet:
		port.Program.List(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "programs" && r.Method == http.MethodPost:
		port.Program.Create(w, r)
	case n == 4 && p[1] == "v1" && p[2] == "program" && r.Method == http.MethodGet:
		port.Program.GetByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "program" && r.Method == http.MethodPut:
		port.Program.UpdateByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "program" && r.Method == http.MethodDelete:
		port.Program.DeleteByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "programs" && p[3] == "select-options" && r.Method == http.MethodGet:
		port.Program.ListAsSelectOptionByFilter(w, r)

	// --- ASSISTANT FILE --- //
	case n == 3 && p[1] == "v1" && p[2] == "assistant-files" && r.Method == http.MethodGet:
		port.AssistantFile.List(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "assistant-files" && r.Method == http.MethodPost:
		port.AssistantFile.Create(w, r)
	case n == 4 && p[1] == "v1" && p[2] == "assistant-file" && r.Method == http.MethodGet:
		port.AssistantFile.GetByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "assistant-file" && r.Method == http.MethodPut:
		port.AssistantFile.UpdateByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "assistant-file" && r.Method == http.MethodDelete:
		port.AssistantFile.DeleteByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "assistant-files" && p[3] == "select-options" && r.Method == http.MethodGet:
		port.AssistantFile.ListAsSelectOptionByFilter(w, r)

	// --- ASSISTANT --- //
	case n == 3 && p[1] == "v1" && p[2] == "assistants" && r.Method == http.MethodGet:
		port.Assistant.List(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "assistants" && r.Method == http.MethodPost:
		port.Assistant.Create(w, r)
	case n == 4 && p[1] == "v1" && p[2] == "assistant" && r.Method == http.MethodGet:
		port.Assistant.GetByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "assistant" && r.Method == http.MethodPut:
		port.Assistant.UpdateByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "assistant" && r.Method == http.MethodDelete:
		port.Assistant.DeleteByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "assistants" && p[3] == "select-options" && r.Method == http.MethodGet:
		port.Assistant.ListAsSelectOptionByFilter(w, r)

	// --- ASSISTANT THREAD --- //
	case n == 3 && p[1] == "v1" && p[2] == "assistant-threads" && r.Method == http.MethodGet:
		port.AssistantThread.List(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "assistant-threads" && r.Method == http.MethodPost:
		port.AssistantThread.Create(w, r)
	case n == 4 && p[1] == "v1" && p[2] == "assistant-thread" && r.Method == http.MethodGet:
		port.AssistantThread.GetByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "assistant-thread" && r.Method == http.MethodPut:
		port.AssistantThread.UpdateByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "assistant-thread" && r.Method == http.MethodDelete:
		port.AssistantThread.DeleteByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "assistant-threads" && p[3] == "select-options" && r.Method == http.MethodGet:
		port.AssistantThread.ListAsSelectOptionByFilter(w, r)

	// --- ASSISTANT MESSAGE --- //
	case n == 3 && p[1] == "v1" && p[2] == "assistant-messages" && r.Method == http.MethodGet:
		port.AssistantMessage.List(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "assistant-messages" && r.Method == http.MethodPost:
		port.AssistantMessage.Create(w, r)
	case n == 4 && p[1] == "v1" && p[2] == "assistant-message" && r.Method == http.MethodGet:
		port.AssistantMessage.GetByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "assistant-message" && r.Method == http.MethodPut:
		port.AssistantMessage.UpdateByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "assistant-message" && r.Method == http.MethodDelete:
		port.AssistantMessage.DeleteByID(w, r, p[3])
	// case n == 4 && p[1] == "v1" && p[2] == "assistant-messages" && p[3] == "select-options" && r.Method == http.MethodGet:
	// 	port.AssistantMessage.ListAsSelectOptionByFilter(w, r)

	// // --- HOW HEAR --- //
	// case n == 3 && p[1] == "v1" && p[2] == "how-hear-about-us-items" && r.Method == http.MethodGet:
	// 	port.HowHear.List(w, r)
	// case n == 3 && p[1] == "v1" && p[2] == "how-hear-about-us-items" && r.Method == http.MethodPost:
	// 	port.HowHear.Create(w, r)
	// case n == 4 && p[1] == "v1" && p[2] == "how-hear-about-us-item" && r.Method == http.MethodGet:
	// 	port.HowHear.GetByID(w, r, p[3])
	// case n == 4 && p[1] == "v1" && p[2] == "how-hear-about-us-item" && r.Method == http.MethodPut:
	// 	port.HowHear.UpdateByID(w, r, p[3])
	// case n == 4 && p[1] == "v1" && p[2] == "how-hear-about-us-item" && r.Method == http.MethodDelete:
	// 	port.HowHear.DeleteByID(w, r, p[3])
	// // case n == 5 && p[1] == "v1" && p[2] == "users" && p[3] == "operation" && p[4] == "create-comment" && r.Method == http.MethodPost:
	// // 	port.Tag.OperationCreateComment(w, r)
	// case n == 4 && p[1] == "v1" && p[2] == "how-hear-about-us-items" && p[3] == "select-options" && r.Method == http.MethodGet:
	// 	port.HowHear.ListAsSelectOptions(w, r)
	case n == 4 && p[1] == "v1" && p[2] == "select-options" && p[3] == "how-hear-about-us-items" && r.Method == http.MethodGet:
		port.HowHear.PublicListAsSelectOptions(w, r)

	// --- USERS --- //
	case n == 3 && p[1] == "v1" && p[2] == "users" && r.Method == http.MethodGet:
		port.User.List(w, r)
	case n == 4 && p[1] == "v1" && p[2] == "users" && p[3] == "count" && r.Method == http.MethodGet:
		port.User.Count(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "users" && r.Method == http.MethodPost:
		port.User.Create(w, r)
	case n == 4 && p[1] == "v1" && p[2] == "user" && r.Method == http.MethodGet:
		port.User.GetByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "user" && r.Method == http.MethodPut:
		port.User.UpdateByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "user" && r.Method == http.MethodDelete:
		port.User.DeleteByID(w, r, p[3])
	case n == 5 && p[1] == "v1" && p[2] == "users" && p[3] == "operation" && p[4] == "create-comment" && r.Method == http.MethodPost:
		port.User.OperationCreateComment(w, r)
	case n == 4 && p[1] == "v1" && p[2] == "users" && p[3] == "select-options" && r.Method == http.MethodGet:
		port.User.ListAsSelectOptions(w, r)

	// --- ATTACHMENTS --- //
	case n == 3 && p[1] == "v1" && p[2] == "attachments" && r.Method == http.MethodGet:
		port.Attachment.List(w, r)
	case n == 3 && p[1] == "v1" && p[2] == "attachments" && r.Method == http.MethodPost:
		port.Attachment.Create(w, r)
	case n == 4 && p[1] == "v1" && p[2] == "attachment" && r.Method == http.MethodGet:
		port.Attachment.GetByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "attachment" && r.Method == http.MethodPut:
		port.Attachment.UpdateByID(w, r, p[3])
	case n == 4 && p[1] == "v1" && p[2] == "attachment" && r.Method == http.MethodDelete:
		port.Attachment.DeleteByID(w, r, p[3])

	// --- CATCH ALL: D.N.E. ---
	default:
		http.NotFound(w, r)
	}
}
