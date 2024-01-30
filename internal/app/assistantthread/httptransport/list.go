package httptransport

import (
	"encoding/json"
	"net/http"
	"strconv"

	assistantthread_s "github.com/bartmika/databoutique-backend/internal/app/assistantthread/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	f := &assistantthread_s.AssistantThreadPaginationListFilter{
		Cursor:    "",
		PageSize:  25,
		SortField: "text",
		SortOrder: assistantthread_s.OrderAscending,
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

	sortField := query.Get("sort_field")
	if sortField != "" {
		f.SortField = sortField
	}

	sortOrderStr := query.Get("sort_order")
	if sortOrderStr != "" {
		sortOrder, _ := strconv.ParseInt(sortOrderStr, 10, 64)
		if sortOrder != 1 && sortOrder != -1 {
			sortOrder = 1
		}
		f.SortOrder = int8(sortOrder)
	}

	searchKeyword := query.Get("search")
	if searchKeyword != "" {
		f.SearchText = searchKeyword
	}

	assistantID := query.Get("assistant_id")
	if assistantID != "" {
		assistantID, err := primitive.ObjectIDFromHex(assistantID)
		if err != nil {
			httperror.ResponseError(w, err)
			return
		}
		f.AssistantID = assistantID
	}

	m, err := h.Controller.ListByFilter(ctx, f)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	MarshalListResponse(m, w)
}

func MarshalListResponse(res *assistantthread_s.AssistantThreadPaginationListResult, w http.ResponseWriter) {
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
