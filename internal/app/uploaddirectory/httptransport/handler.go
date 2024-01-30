package httptransport

import (
	"log/slog"

	uploaddirectory_c "github.com/bartmika/databoutique-backend/internal/app/uploaddirectory/controller"
)

// Handler Creates http request handler
type Handler struct {
	Logger     *slog.Logger
	Controller uploaddirectory_c.UploadDirectoryController
}

// NewHandler Constructor
func NewHandler(loggerp *slog.Logger, c uploaddirectory_c.UploadDirectoryController) *Handler {
	return &Handler{
		Logger:     loggerp,
		Controller: c,
	}
}
