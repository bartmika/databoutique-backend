package httptransport

import (
	"encoding/json"
	"net/http"

	sub_s "github.com/bartmika/databoutique-backend/internal/app/assistant/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func (h *Handler) ListAsSelectOptionByFilter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	f := &sub_s.AssistantListFilter{
		PageSize: 1_000_000,
		// LastID:    "",
		SortField:       "text",
		ExcludeArchived: true,
	}

	// Perform our database operation.
	m, err := h.Controller.ListAsSelectOptionByFilter(ctx, f)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	MarshalListAsSelectOptionResponse(m, w)
}

func MarshalListAsSelectOptionResponse(res []*sub_s.AssistantAsSelectOption, w http.ResponseWriter) {
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
