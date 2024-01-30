package httptransport

import (
	"log/slog"

	gateway_c "github.com/bartmika/databoutique-backend/internal/app/gateway/controller"
)

// Handler Creates http request handler
type Handler struct {
	Logger     *slog.Logger
	Controller gateway_c.GatewayController
}

// NewHandler Constructor
func NewHandler(loggerp *slog.Logger, c gateway_c.GatewayController) *Handler {
	return &Handler{
		Logger:     loggerp,
		Controller: c,
	}
}
