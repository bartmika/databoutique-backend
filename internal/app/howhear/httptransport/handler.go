package httptransport

import (
	"log/slog"

	howhear_c "github.com/bartmika/databoutique-backend/internal/app/howhear/controller"
)

// Handler Creates http request handler
type Handler struct {
	Logger     *slog.Logger
	Controller howhear_c.HowHearAboutUsItemController
}

// NewHandler Constructor
func NewHandler(loggerp *slog.Logger, c howhear_c.HowHearAboutUsItemController) *Handler {
	return &Handler{
		Logger:     loggerp,
		Controller: c,
	}
}
