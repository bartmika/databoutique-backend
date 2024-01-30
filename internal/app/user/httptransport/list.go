package httptransport

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/bartmika/timekit"
	sub_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	f := &sub_s.UserListFilter{
		Cursor:          primitive.NilObjectID,
		PageSize:        25,
		SortField:       "_id",
		SortOrder:       1, // 1=ascending | -1=descending
		ExcludeArchived: true,
	}

	// Here is where you extract url parameters.
	query := r.URL.Query()

	cursor := query.Get("cursor")
	if cursor != "" {
		cursor, err := primitive.ObjectIDFromHex(cursor)
		if err != nil {
			httperror.ResponseError(w, err)
			return
		}
		f.Cursor = cursor
	}

	pageSize := query.Get("page_size")
	if pageSize != "" {
		pageSize, _ := strconv.ParseInt(pageSize, 10, 64)
		if pageSize == 0 || pageSize > 250 {
			pageSize = 250
		}
		f.PageSize = pageSize
	}

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

	statusStr := query.Get("status")
	if statusStr != "" {
		status, _ := strconv.ParseInt(statusStr, 10, 64)
		f.Status = int8(status)
	}

	roleStr := query.Get("role")
	if roleStr != "" {
		role, _ := strconv.ParseInt(roleStr, 10, 64)
		f.Role = int8(role)
	}

	createdAtGTEStr := query.Get("created_at_gte")
	if createdAtGTEStr != "" {
		createdAtGTE, err := timekit.ParseJavaScriptTimeString(createdAtGTEStr)
		if err != nil {
			httperror.ResponseError(w, err)
			return
		}
		f.CreatedAtGTE = createdAtGTE
	}

	// Perform our database operation.
	m, err := h.Controller.ListByFilter(ctx, f)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	MarshalListResponse(m, w)
}

func MarshalListResponse(res *sub_s.UserListResult, w http.ResponseWriter) {
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
