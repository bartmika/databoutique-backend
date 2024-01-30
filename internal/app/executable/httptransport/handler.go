package httptransport

import (
	"log/slog"

	executable_c "github.com/bartmika/databoutique-backend/internal/app/executable/controller"
)

// Handler Creates http request handler
type Handler struct {
	Logger     *slog.Logger
	Controller executable_c.ExecutableController
}

// NewHandler Constructor
func NewHandler(loggerp *slog.Logger, c executable_c.ExecutableController) *Handler {
	return &Handler{
		Logger:     loggerp,
		Controller: c,
	}
}
