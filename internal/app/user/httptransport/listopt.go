package httptransport

import (
	"encoding/json"
	"net/http"

	sub_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h *Handler) ListAsSelectOptions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	f := &sub_s.UserListFilter{
		// PageSize:  10,
		// LastID:    "",
		SortField:       "_id",
		ExcludeArchived: true,
	}

	// Here is where you extract url parameters.
	query := r.URL.Query()

	// Apply search text if it exists in url parameter.
	searchKeyword := query.Get("search")
	if searchKeyword != "" {
		f.SearchText = searchKeyword
	}

	// Apply filters it exists in url parameter.
	firstName := query.Get("first_name")
	if firstName != "" {
		f.FirstName = firstName
	}
	lastName := query.Get("first_name")
	if lastName != "" {
		f.LastName = lastName
	}
	email := query.Get("email")
	if email != "" {
		f.Email = email
	}
	phone := query.Get("phone")
	if phone != "" {
		f.Phone = phone
	}
	tenantID := query.Get("tenant_id")
	if tenantID != "" {
		tenantID, err := primitive.ObjectIDFromHex(tenantID)
		if err != nil {
			httperror.ResponseError(w, err)
			return
		}
		f.TenantID = tenantID
	}

	// Perform our database operation.
	m, err := h.Controller.ListAsSelectOptionByFilter(ctx, f)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	MarshalListAsSelectOptionResponse(m, w)
}

func MarshalListAsSelectOptionResponse(res []*sub_s.UserAsSelectOption, w http.ResponseWriter) {
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
