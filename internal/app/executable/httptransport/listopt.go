package httptransport

import (
	"encoding/json"
	"net/http"
	"strconv"

	executable_s "github.com/bartmika/databoutique-backend/internal/app/executable/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func (h *Handler) ListAsSelectOptionByFilter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	f := &executable_s.ExecutablePaginationListFilter{
		PageSize: 1_000_000,
		// LastID:    "",
		SortField: "sort_number",
		SortOrder: 1, // 1=ascending | -1=descending
		Status:    executable_s.ExecutableStatusActive,
	}

	// Here is where you extract url parameters.
	query := r.URL.Query()

	statusStr := query.Get("status")
	if statusStr != "" {
		status, _ := strconv.ParseInt(statusStr, 10, 64)
		f.Status = int8(status)
	}

	// Perform our database operation.
	m, err := h.Controller.ListAsSelectOptionByFilter(ctx, f)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	MarshalListAsSelectOptionResponse(m, w)
}

func MarshalListAsSelectOptionResponse(res []*executable_s.ExecutableAsSelectOption, w http.ResponseWriter) {
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) PublicListAsSelectOptions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	f := &executable_s.ExecutablePaginationListFilter{
		PageSize: 1_000_000,
		// LastID:    "",
		SortField: "sort_number",
		SortOrder: 1, // 1=ascending | -1=descending
		Status:    executable_s.ExecutableStatusActive,
	}

	// Here is where you extract url parameters.
	query := r.URL.Query()

	statusStr := query.Get("status")
	if statusStr != "" {
		status, _ := strconv.ParseInt(statusStr, 10, 64)
		f.Status = int8(status)
	}

	// Perform our database operation.
	m, err := h.Controller.PublicListAsSelectOptionByFilter(ctx, f)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	MarshalListAsSelectOptionResponse(m, w)
}
