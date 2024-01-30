package httptransport

import (
	"log/slog"

	program_c "github.com/bartmika/databoutique-backend/internal/app/program/controller"
)

// Handler Creates http request handler
type Handler struct {
	Logger     *slog.Logger
	Controller program_c.ProgramController
}

// NewHandler Constructor
func NewHandler(loggerp *slog.Logger, c program_c.ProgramController) *Handler {
	return &Handler{
		Logger:     loggerp,
		Controller: c,
	}
}
