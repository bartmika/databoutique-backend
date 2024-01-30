package httptransport

import (
	"encoding/json"
	"net/http"

	way_c "github.com/bartmika/databoutique-backend/internal/app/gateway/controller"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	profile, err := h.Controller.Dashboard(ctx)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}
	MarshalDashboardResponse(profile, w)
}

func MarshalDashboardResponse(responseData *way_c.DashboardResponseIDO, w http.ResponseWriter) {
	if err := json.NewEncoder(w).Encode(&responseData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
