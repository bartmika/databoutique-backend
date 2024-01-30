package httptransport

import (
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"

	folderinfo_s "github.com/bartmika/databoutique-backend/internal/app/folderinfo/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	res, err := h.Controller.GetByID(ctx, objectID)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	MarshalDetailResponse(res, w)
}

func MarshalDetailResponse(res *folderinfo_s.FolderInfo, w http.ResponseWriter) {
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
