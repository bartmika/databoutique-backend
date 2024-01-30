package httptransport

import (
	"log/slog"

	assistant_c "github.com/bartmika/databoutique-backend/internal/app/assistant/controller"
)

// Handler Creates http request handler
type Handler struct {
	Logger     *slog.Logger
	Controller assistant_c.AssistantController
}

// NewHandler Constructor
func NewHandler(loggerp *slog.Logger, c assistant_c.AssistantController) *Handler {
	return &Handler{
		Logger:     loggerp,
		Controller: c,
	}
}
