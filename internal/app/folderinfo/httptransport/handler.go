package httptransport

import (
	"log/slog"

	folderinfo_c "github.com/bartmika/databoutique-backend/internal/app/folderinfo/controller"
)

// Handler Creates http request handler
type Handler struct {
	Logger     *slog.Logger
	Controller folderinfo_c.FolderInfoController
}

// NewHandler Constructor
func NewHandler(loggerp *slog.Logger, c folderinfo_c.FolderInfoController) *Handler {
	return &Handler{
		Logger:     loggerp,
		Controller: c,
	}
}
