package httptransport

import (
	"encoding/json"
	"net/http"
	"strconv"

	assistantmessage_s "github.com/bartmika/databoutique-backend/internal/app/assistantmessage/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	f := &assistantmessage_s.AssistantMessagePaginationListFilter{
		Cursor:    "",
		PageSize:  25,
		SortField: "sort_number",
		SortOrder: 1, // 1=ascending | -1=descending
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

	assistantThreadID := query.Get("assistant_thread_id")
	if assistantThreadID != "" {
		assistantThreadID, err := primitive.ObjectIDFromHex(assistantThreadID)
		if err != nil {
			httperror.ResponseError(w, err)
			return
		}
		f.AssistantThreadID = assistantThreadID
	} else {
		httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("assistant_thread_id", "missing url query parameter"))
		return
	}

	m, err := h.Controller.ListByFilter(ctx, f)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	MarshalListResponse(m, w)
}

func MarshalListResponse(res *assistantmessage_s.AssistantMessagePaginationListResult, w http.ResponseWriter) {
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
