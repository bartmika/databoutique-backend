package httptransport

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	a_c "github.com/bartmika/databoutique-backend/internal/app/uploadfile/controller"
	sub_c "github.com/bartmika/databoutique-backend/internal/app/uploadfile/controller"
	sub_s "github.com/bartmika/databoutique-backend/internal/app/uploadfile/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func UnmarshalUpdateRequest(ctx context.Context, r *http.Request) (*sub_c.UploadFileUpdateRequestIDO, error) {
	defer r.Body.Close()

	// Parse the multipart form data
	err := r.ParseMultipartForm(32 << 20) // Limit the maximum memory used for parsing to 32MB
	if err != nil {
		log.Println("UnmarshalUpdateRequest:ParseMultipartForm:err:", err)
		return nil, err
	}

	// Get the values of form fields
	id := r.FormValue("id")
	name := r.FormValue("name")
	description := r.FormValue("description")

	// Get the uploaded file from the request
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Println("UnmarshalUpdateRequest:FormFile:err:", err)
		// return nil, err, http.StatusInternalServerError
	}

	aid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("UnmarshalUpdateRequest: primitive.ObjectIDFromHex:err:", err)
	}

	uploadDirectoryID := primitive.NilObjectID
	uploadDirectoryIDStr := r.FormValue("upload_directory_id")
	if uploadDirectoryIDStr != "" {
		uploadDirectoryID, err = primitive.ObjectIDFromHex(uploadDirectoryIDStr)
		if err != nil {
			log.Println("UnmarshalCmsImageCreateRequest:missing:uploadDirectoryID")
			// return nil, err, http.StatusInternalServerError
		}
	}

	// Initialize our array which will store all the results from the remote server.
	requestData := &a_c.UploadFileUpdateRequestIDO{
		ID:                aid,
		Name:              name,
		Description:       description,
		UploadDirectoryID: uploadDirectoryID,
	}

	if header != nil {
		// Extract filename and filetype from the file header
		requestData.FileName = header.Filename
		requestData.FileType = header.Header.Get("Content-Type")
		requestData.File = file
	}
	return requestData, nil
}

func (h *Handler) UpdateByID(w http.ResponseWriter, r *http.Request, idStr string) {
	ctx := r.Context()

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	data, err := UnmarshalUpdateRequest(ctx, r)
	data.ID = id
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}
	uploadfile, err := h.Controller.UpdateByID(ctx, data)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	MarshalUpdateResponse(uploadfile, w)
}

func MarshalUpdateResponse(res *sub_s.UploadFile, w http.ResponseWriter) {
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
