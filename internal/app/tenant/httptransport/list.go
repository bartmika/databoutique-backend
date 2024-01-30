package httptransport

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/bartmika/timekit"
	sub_s "github.com/bartmika/databoutique-backend/internal/app/tenant/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	f := &sub_s.TenantListFilter{
		Cursor:          primitive.NilObjectID,
		PageSize:        25,
		SortField:       "_id",
		SortOrder:       1, // 1=ascending | -1=descending
		ExcludeArchived: true,
	}

	// Here is where you extract url parameters.
	query := r.URL.Query()

	cursor := query.Get("cursor")
	if cursor != "" {
		cursor, err := primitive.ObjectIDFromHex(cursor)
		if err != nil {
			httperror.ResponseError(w, err)
			return
		}
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

	searchText := query.Get("search")
	if searchText != "" {
		f.SearchText = searchText
	}

	statusStr := query.Get("status")
	if statusStr != "" {
		status, _ := strconv.ParseInt(statusStr, 10, 64)
		f.Status = int8(status)
	}
	createdAtGTEStr := query.Get("created_at_gte")
	if createdAtGTEStr != "" {
		createdAtGTE, err := timekit.ParseJavaScriptTimeString(createdAtGTEStr)
		if err != nil {
			httperror.ResponseError(w, err)
			return
		}
		f.CreatedAtGTE = createdAtGTE
	}

	m, err := h.Controller.ListByFilter(ctx, f)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	MarshalListResponse(m, w)
}

func MarshalListResponse(res *sub_s.TenantListResult, w http.ResponseWriter) {
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) ListAsSelectOptionByFilter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	f := &sub_s.TenantListFilter{
		Cursor:          primitive.NilObjectID,
		PageSize:        6,
		SortField:       "_id",
		SortOrder:       1, // 1=ascending | -1=descending
		ExcludeArchived: true,
	}

	m, err := h.Controller.ListAsSelectOptionByFilter(ctx, f)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	MarshalListAsSelectOptionResponse(m, w)
}

func MarshalListAsSelectOptionResponse(res []*sub_s.TenantAsSelectOption, w http.ResponseWriter) {
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
