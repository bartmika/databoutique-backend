package httptransport

import (
	"encoding/json"
	"net/http"
	"strconv"

	sub_s "github.com/bartmika/databoutique-backend/internal/app/uploadfile/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	f := &sub_s.UploadFilePaginationListFilter{
		Cursor:    "",
		PageSize:  25,
		SortField: "created_at",
		SortOrder: sub_s.OrderDescending,
	}

	// Here is where you extract url parameters.
	query := r.URL.Query()

	cursor := query.Get("cursor")
	if cursor != "" {
		f.Cursor = cursor
	}

	pageSize := query.Get("page_size")
	if pageSize != "" {
		pageSize, _ := strconv.ParseInt(pageSize, 10, 64)
		if pageSize == 0 || pageSize > 250 {
			pageSize = 250
		}
		f.PageSize = pageSize
	}

	name := query.Get("name")
	if name != "" {
		f.Name = name
	}

	description := query.Get("description")
	if description != "" {
		f.Description = description
	}

	uploadDirectoryIDStr := query.Get("upload_directory_id")
	if uploadDirectoryIDStr != "" {
		uploadDirectoryID, err := primitive.ObjectIDFromHex(uploadDirectoryIDStr)
		if err != nil {
			httperror.ResponseError(w, err)
			return
		}
		f.UploadDirectoryID = uploadDirectoryID
	}

	m, err := h.Controller.ListByFilter(ctx, f)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	MarshalListResponse(m, w)
}

func MarshalListResponse(res *sub_s.UploadFilePaginationListResult, w http.ResponseWriter) {
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) ListAsSelectOptionByFilter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	f := &sub_s.UploadFilePaginationListFilter{
		Cursor:    "",
		PageSize:  1_000_000_000,
		SortField: "created_at",
		SortOrder: sub_s.OrderDescending,
	}

	m, err := h.Controller.ListAsSelectOptionByFilter(ctx, f)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	MarshalListAsSelectOptionResponse(m, w)
}

func MarshalListAsSelectOptionResponse(res []*sub_s.UploadFileAsSelectOption, w http.ResponseWriter) {
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
