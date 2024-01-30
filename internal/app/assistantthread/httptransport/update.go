package httptransport

import (
	"context"
	"encoding/json"
	"net/http"

	assistantthread_c "github.com/bartmika/databoutique-backend/internal/app/assistantthread/controller"
	assistantthread_s "github.com/bartmika/databoutique-backend/internal/app/assistantthread/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func UnmarshalUpdateRequest(ctx context.Context, r *http.Request) (*assistantthread_c.AssistantThreadUpdateRequestIDO, error) {
	// Initialize our array which will store all the results from the remote server.
	var requestData assistantthread_c.AssistantThreadUpdateRequestIDO

	defer r.Body.Close()

	// Read the JSON string and convert it into our golang stuct else we need
	// to send a `400 Bad Request` errror message back to the client,
	err := json.NewDecoder(r.Body).Decode(&requestData) // [1]
	if err != nil {
		return nil, httperror.NewForSingleField(http.StatusBadRequest, "non_field_error", "payload structure is wrong")
	}
	return &requestData, nil
}

func (h *Handler) UpdateByID(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	data, err := UnmarshalUpdateRequest(ctx, r)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	org, err := h.Controller.UpdateByID(ctx, data)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	MarshalUpdateResponse(org, w)
}

func MarshalUpdateResponse(res *assistantthread_s.AssistantThread, w http.ResponseWriter) {
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
