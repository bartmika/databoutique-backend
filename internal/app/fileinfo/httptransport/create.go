package httptransport

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	a_c "github.com/bartmika/databoutique-backend/internal/app/fileinfo/controller"
	a_s "github.com/bartmika/databoutique-backend/internal/app/fileinfo/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func UnmarshalCreateRequest(ctx context.Context, r *http.Request) (*a_c.FileInfoCreateRequestIDO, error) {
	defer r.Body.Close()

	// Parse the multipart form data
	err := r.ParseMultipartForm(32 << 20) // Limit the maximum memory used for parsing to 32MB
	if err != nil {
		log.Println("UnmarshalCreateRequest:ParseMultipartForm:err:", err)
		return nil, err
	}

	// Get the values of form fields
	name := r.FormValue("name")
	description := r.FormValue("description")

	// Get the uploaded file from the request
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Println("UnmarshalCmsImageCreateRequest:FormFile:err:", err)
		// return nil, err, http.StatusInternalServerError
	}

	// Initialize our array which will store all the results from the remote server.
	requestData := &a_c.FileInfoCreateRequestIDO{
		Name:        name,
		Description: description,
	}

	if header != nil {
		// Extract filename and filetype from the file header
		requestData.FileName = header.Filename
		requestData.FileType = header.Header.Get("Content-Type")
		requestData.File = file
	}
	return requestData, nil
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := UnmarshalCreateRequest(ctx, r)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	fileinfo, err := h.Controller.Create(ctx, data)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	MarshalCreateResponse(fileinfo, w)
}

func MarshalCreateResponse(res *a_s.FileInfo, w http.ResponseWriter) {
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
