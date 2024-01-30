package httptransport

import (
	"log/slog"

	programcategory_c "github.com/bartmika/databoutique-backend/internal/app/programcategory/controller"
)

// Handler Creates http request handler
type Handler struct {
	Logger     *slog.Logger
	Controller programcategory_c.ProgramCategoryController
}

// NewHandler Constructor
func NewHandler(loggerp *slog.Logger, c programcategory_c.ProgramCategoryController) *Handler {
	return &Handler{
		Logger:     loggerp,
		Controller: c,
	}
}
