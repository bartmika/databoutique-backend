package httptransport

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	gateway_c "github.com/bartmika/databoutique-backend/internal/app/gateway/controller"
	user_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	profile, err := h.Controller.Profile(ctx)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}
	MarshalProfileResponse(profile, w)
}

func MarshalProfileResponse(responseData *user_s.User, w http.ResponseWriter) {
	if err := json.NewEncoder(w).Encode(&responseData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type ProfileUpdateRequestIDO struct {
	FirstName                   string             `bson:"first_name" json:"first_name"`
	LastName                    string             `bson:"last_name" json:"last_name"`
	Email                       string             `json:"email"`
	Phone                       string             `bson:"phone,omitempty" json:"phone,omitempty"`
	Country                     string             `bson:"country,omitempty" json:"country,omitempty"`
	Region                      string             `bson:"region,omitempty" json:"region,omitempty"`
	City                        string             `bson:"city,omitempty" json:"city,omitempty"`
	PostalCode                  string             `bson:"postal_code,omitempty" json:"postal_code,omitempty"`
	AddressLine1                string             `bson:"address_line1,omitempty" json:"address_line1,omitempty"`
	AddressLine2                string             `bson:"address_line2,omitempty" json:"address_line2,omitempty"`
	HowDidYouHearAboutUsID      primitive.ObjectID `bson:"how_did_you_hear_about_us_id" json:"how_did_you_hear_about_us_id,omitempty"`
	IsHowDidYouHearAboutUsOther bool               `bson:"is_how_did_you_hear_about_us_other" json:"is_how_did_you_hear_about_us_other,omitempty"`
	HowDidYouHearAboutUsOther   string             `bson:"how_did_you_hear_about_us_other" json:"how_did_you_hear_about_us_other,omitempty"`
	AgreePromotionsEmail        bool               `bson:"agree_promotions_email,omitempty" json:"agree_promotions_email,omitempty"`
	HasShippingAddress          bool               `bson:"has_shipping_address" json:"has_shipping_address,omitempty"`
	ShippingName                string             `bson:"shipping_name" json:"shipping_name,omitempty"`
	ShippingPhone               string             `bson:"shipping_phone" json:"shipping_phone,omitempty"`
	ShippingCountry             string             `bson:"shipping_country" json:"shipping_country,omitempty"`
	ShippingRegion              string             `bson:"shipping_region" json:"shipping_region,omitempty"`
	ShippingCity                string             `bson:"shipping_city" json:"shipping_city,omitempty"`
	ShippingPostalCode          string             `bson:"shipping_postal_code" json:"shipping_postal_code,omitempty"`
	ShippingAddressLine1        string             `bson:"shipping_address_line1" json:"shipping_address_line1,omitempty"`
	ShippingAddressLine2        string             `bson:"shipping_address_line2" json:"shipping_address_line2,omitempty"`
}

func UnmarshalProfileUpdateRequest(ctx context.Context, r *http.Request) (*user_s.User, error) {
	// Initialize our array which will store all the results from the remote server.
	var requestData ProfileUpdateRequestIDO

	defer r.Body.Close()

	// Read the JSON string and convert it into our golang stuct else we need
	// to send a `400 Bad Request` errror message back to the client,
	err := json.NewDecoder(r.Body).Decode(&requestData) // [1]
	if err != nil {
		return nil, httperror.NewForSingleField(http.StatusBadRequest, "non_field_error", "payload structure is wrong")
	}

	// Defensive Code: For security purposes we need to remove all whitespaces from the email and lower the characters.
	requestData.Email = strings.ToLower(requestData.Email)
	requestData.Email = strings.ReplaceAll(requestData.Email, " ", "")

	// Perform our validation and return validation error on any issues detected.
	if err = ValidateProfileUpdateRequest(&requestData); err != nil {
		return nil, err
	}

	// Convert to the user collection.
	return &user_s.User{
		FirstName:                   requestData.FirstName,
		LastName:                    requestData.LastName,
		Email:                       requestData.Email,
		Phone:                       requestData.Phone,
		Country:                     requestData.Country,
		Region:                      requestData.Region,
		City:                        requestData.City,
		PostalCode:                  requestData.PostalCode,
		AddressLine1:                requestData.AddressLine1,
		AddressLine2:                requestData.AddressLine2,
		HowDidYouHearAboutUsID:      requestData.HowDidYouHearAboutUsID,
		IsHowDidYouHearAboutUsOther: requestData.IsHowDidYouHearAboutUsOther,
		HowDidYouHearAboutUsOther:   requestData.HowDidYouHearAboutUsOther,
		AgreePromotionsEmail:        requestData.AgreePromotionsEmail,
		HasShippingAddress:          requestData.HasShippingAddress,
		ShippingName:                requestData.ShippingName,
		ShippingPhone:               requestData.ShippingPhone,
		ShippingCountry:             requestData.ShippingCountry,
		ShippingRegion:              requestData.ShippingRegion,
		ShippingCity:                requestData.ShippingCity,
		ShippingPostalCode:          requestData.ShippingPostalCode,
		ShippingAddressLine1:        requestData.ShippingAddressLine1,
		ShippingAddressLine2:        requestData.ShippingAddressLine2,
	}, nil
}

func ValidateProfileUpdateRequest(dirtyData *ProfileUpdateRequestIDO) error {
	e := make(map[string]string)

	if dirtyData.FirstName == "" {
		e["first_name"] = "missing value"
	}
	if dirtyData.LastName == "" {
		e["last_name"] = "missing value"
	}
	if dirtyData.Email == "" {
		e["email"] = "missing value"
	}
	if len(dirtyData.Email) > 255 {
		e["email"] = "too long"
	}
	if dirtyData.Phone == "" {
		e["phone"] = "missing value"
	}
	if dirtyData.Country == "" {
		e["country"] = "missing value"
	}
	if dirtyData.Region == "" {
		e["region"] = "missing value"
	}
	if dirtyData.City == "" {
		e["city"] = "missing value"
	}
	if dirtyData.PostalCode == "" {
		e["postal_code"] = "missing value"
	}
	if dirtyData.AddressLine1 == "" {
		e["address_line1"] = "missing value"
	}

	// The following logic will enforce shipping address input validation.
	if dirtyData.HasShippingAddress == true {
		if dirtyData.ShippingName == "" {
			e["shipping_name"] = "missing value"
		}
		if dirtyData.ShippingPhone == "" {
			e["shipping_phone"] = "missing value"
		}
		if dirtyData.ShippingCountry == "" {
			e["shipping_country"] = "missing value"
		}
		if dirtyData.ShippingRegion == "" {
			e["shipping_region"] = "missing value"
		}
		if dirtyData.ShippingCity == "" {
			e["shipping_city"] = "missing value"
		}
		if dirtyData.ShippingPostalCode == "" {
			e["shipping_postal_code"] = "missing value"
		}
		if dirtyData.ShippingAddressLine1 == "" {
			e["shipping_address_line1"] = "missing value"
		}
	}

	if len(e) != 0 {
		return httperror.NewForBadRequest(&e)
	}
	return nil
}

func (h *Handler) ProfileUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := UnmarshalProfileUpdateRequest(ctx, r)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	if err := h.Controller.ProfileUpdate(ctx, data); err != nil {
		httperror.ResponseError(w, err)
		return
	}

	// Get the request
	h.Profile(w, r)
}

func UnmarshalProfileChangePasswordRequest(ctx context.Context, r *http.Request) (*gateway_c.ProfileChangePasswordRequestIDO, error) {
	// Initialize our array which will store all the results from the remote server.
	var requestData gateway_c.ProfileChangePasswordRequestIDO

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

func (h *Handler) ProfileChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := UnmarshalProfileChangePasswordRequest(ctx, r)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	if err := h.Controller.ProfileChangePassword(ctx, data); err != nil {
		httperror.ResponseError(w, err)
		return
	}

	// Get the request
	h.Profile(w, r)
}
