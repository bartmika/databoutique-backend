package httptransport

import (
	"log/slog"

	assistantmessage_c "github.com/bartmika/databoutique-backend/internal/app/assistantmessage/controller"
)

// Handler Creates http request handler
type Handler struct {
	Logger     *slog.Logger
	Controller assistantmessage_c.AssistantMessageController
}

// NewHandler Constructor
func NewHandler(loggerp *slog.Logger, c assistantmessage_c.AssistantMessageController) *Handler {
	return &Handler{
		Logger:     loggerp,
		Controller: c,
	}
}
