package httptransport

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"

	sub_s "github.com/bartmika/databoutique-backend/internal/app/tenant/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

type TenantOperationCreateCommentRequest struct {
	TenantID primitive.ObjectID `bson:"tenant_id" json:"tenant_id"`
	Content  string             `bson:"content" json:"content"`
}

func UnmarshalOperationCreateCommentRequest(ctx context.Context, r *http.Request) (*TenantOperationCreateCommentRequest, error) {
	// Initialize our array which will store all the results from the remote server.
	var requestData TenantOperationCreateCommentRequest

	defer r.Body.Close()

	// Read the JSON string and convert it into our golang stuct else we need
	// to send a `400 Bad Request` errror message back to the client,
	err := json.NewDecoder(r.Body).Decode(&requestData) // [1]
	if err != nil {
		log.Println("UnmarshalOperationCreateCommentRequest | NewDecoder/Decode | err:", err)
		return nil, httperror.NewForSingleField(http.StatusBadRequest, "non_field_error", "payload structure is wrong")
	}

	// Perform our validation and return validation error on any issues detected.
	if err := ValidateOperationCreateCommentRequest(&requestData); err != nil {
		return nil, err
	}
	return &requestData, nil
}

func ValidateOperationCreateCommentRequest(dirtyData *TenantOperationCreateCommentRequest) error {
	e := make(map[string]string)

	if dirtyData.TenantID.IsZero() {
		e["tenant_id"] = "missing value"
	}

	if dirtyData.Content == "" {
		e["content"] = "missing value"
	}

	if len(e) != 0 {
		return httperror.NewForBadRequest(&e)
	}
	return nil
}

func (h *Handler) OperationCreateComment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	reqData, err := UnmarshalOperationCreateCommentRequest(ctx, r)
	if err != nil {
		log.Println("OperationCreateComment | UnmarshalOperationCreateCommentRequest | err:", err)
		httperror.ResponseError(w, err)
		return
	}
	log.Println("reqData.TenantID -->", reqData.TenantID)
	data, err := h.Controller.CreateComment(ctx, reqData.TenantID, reqData.Content)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	MarshalOperationCreateCommentResponse(data, w)
}

func MarshalOperationCreateCommentResponse(res *sub_s.Tenant, w http.ResponseWriter) {
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
