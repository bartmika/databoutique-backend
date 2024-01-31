package httptransport

import (
	uploadfile_c "github.com/bartmika/databoutique-backend/internal/app/uploadfile/controller"
)

// Handler Creates http request handler
type Handler struct {
	Controller uploadfile_c.UploadFileController
}

// NewHandler Constructor
func NewHandler(c uploadfile_c.UploadFileController) *Handler {
	return &Handler{
		Controller: c,
	}
}
