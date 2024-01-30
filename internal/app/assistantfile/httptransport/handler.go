package httptransport

import (
	assistantfile_c "github.com/bartmika/databoutique-backend/internal/app/assistantfile/controller"
)

// Handler Creates http request handler
type Handler struct {
	Controller assistantfile_c.AssistantFileController
}

// NewHandler Constructor
func NewHandler(c assistantfile_c.AssistantFileController) *Handler {
	return &Handler{
		Controller: c,
	}
}
