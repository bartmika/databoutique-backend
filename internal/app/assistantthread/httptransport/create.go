package httptransport

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	assistantthread_c "github.com/bartmika/databoutique-backend/internal/app/assistantthread/controller"
	assistantthread_s "github.com/bartmika/databoutique-backend/internal/app/assistantthread/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func UnmarshalCreateRequest(ctx context.Context, r *http.Request) (*assistantthread_c.AssistantThreadCreateRequestIDO, error) {
	// Initialize our array which will store all the results from the remote server.
	var requestData assistantthread_c.AssistantThreadCreateRequestIDO

	defer r.Body.Close()

	// Read the JSON string and convert it into our golang stuct else we need
	// to send a `400 Bad Request` errror message back to the client,
	err := json.NewDecoder(r.Body).Decode(&requestData) // [1]
	if err != nil {
		log.Println(err)
		return nil, httperror.NewForSingleField(http.StatusBadRequest, "non_field_error", "payload structure is wrong")
	}

	return &requestData, nil
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := UnmarshalCreateRequest(ctx, r)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	res, err := h.Controller.Create(ctx, data)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	MarshalCreateResponse(res, w)
}

func MarshalCreateResponse(res *assistantthread_s.AssistantThread, w http.ResponseWriter) {
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
