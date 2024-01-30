package httptransport

import (
	"log/slog"

	assistantthread_c "github.com/bartmika/databoutique-backend/internal/app/assistantthread/controller"
)

// Handler Creates http request handler
type Handler struct {
	Logger     *slog.Logger
	Controller assistantthread_c.AssistantThreadController
}

// NewHandler Constructor
func NewHandler(loggerp *slog.Logger, c assistantthread_c.AssistantThreadController) *Handler {
	return &Handler{
		Logger:     loggerp,
		Controller: c,
	}
}
