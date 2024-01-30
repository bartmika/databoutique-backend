package httptransport

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	executable_c "github.com/bartmika/databoutique-backend/internal/app/executable/controller"
	executable_s "github.com/bartmika/databoutique-backend/internal/app/executable/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func UnmarshalCreateRequest(ctx context.Context, r *http.Request) (*executable_c.ExecutableCreateRequestIDO, error) {
	// Initialize our array which will store all the results from the remote server.
	var requestData executable_c.ExecutableCreateRequestIDO

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

func MarshalCreateResponse(res *executable_s.Executable, w http.ResponseWriter) {
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
