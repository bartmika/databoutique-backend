package httptransport

import (
	fileinfo_c "github.com/bartmika/databoutique-backend/internal/app/fileinfo/controller"
)

// Handler Creates http request handler
type Handler struct {
	Controller fileinfo_c.FileInfoController
}

// NewHandler Constructor
func NewHandler(c fileinfo_c.FileInfoController) *Handler {
	return &Handler{
		Controller: c,
	}
}
