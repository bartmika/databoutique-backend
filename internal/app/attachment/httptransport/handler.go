package httptransport

import (
	attachment_c "github.com/bartmika/databoutique-backend/internal/app/attachment/controller"
)

// Handler Creates http request handler
type Handler struct {
	Controller attachment_c.AttachmentController
}

// NewHandler Constructor
func NewHandler(c attachment_c.AttachmentController) *Handler {
	return &Handler{
		Controller: c,
	}
}
