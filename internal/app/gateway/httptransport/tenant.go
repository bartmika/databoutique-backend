package httptransport

import (
	"context"
	"encoding/json"
	"net/http"

	gateway_c "github.com/bartmika/databoutique-backend/internal/app/gateway/controller"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func UnmarshalExecutiveVisitsTenantRequest(ctx context.Context, r *http.Request) (*gateway_c.ExecutiveVisitsTenantRequest, error) {
	// Initialize our array which will store all the results from the remote server.
	var requestData gateway_c.ExecutiveVisitsTenantRequest

	defer r.Body.Close()

	// Read the JSON string and convert it into our golang stuct else we need
	// to send a `400 Bad Request` errror message back to the client,
	err := json.NewDecoder(r.Body).Decode(&requestData) // [1]
	if err != nil {
		return nil, httperror.NewForSingleField(http.StatusBadRequest, "non_field_error", "payload structure is wrong")
	}

	// Return our result
	return &requestData, nil
}

func (h *Handler) ExecutiveVisitsTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	requestData, err := UnmarshalExecutiveVisitsTenantRequest(ctx, r)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	if err := h.Controller.ExecutiveVisitsTenant(ctx, requestData); err != nil {
		httperror.ResponseError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
