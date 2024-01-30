package controller

import (
	"context"
	"fmt"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"

	user_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

type UserUpdateRequestIDO struct {
	ID               primitive.ObjectID `bson:"_id" json:"id"`
	TenantID         primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	FirstName        string             `json:"first_name"`
	LastName         string             `json:"last_name"`
	Email            string             `json:"email"`
	Password         string             `json:"password"`
	PasswordRepeated string             `json:"password_repeated"`
	Phone            string             `json:"phone,omitempty"`
	Country          string             `json:"country,omitempty"`
	Region           string             `json:"region,omitempty"`
	City             string             `json:"city,omitempty"`
	PostalCode       string             `json:"postal_code,omitempty"`
	AddressLine1     string             `json:"address_line1,omitempty"`
	AddressLine2     string             `json:"address_line2,omitempty"`
	// HowDidYouHearAboutUs      int8               `json:"how_did_you_hear_about_us,omitempty"`
	// HowDidYouHearAboutUsOther string             `json:"how_did_you_hear_about_us_other,omitempty"`
	AgreeTOS             bool   `json:"agree_tos,omitempty"`
	AgreePromotionsEmail bool   `json:"agree_promotions_email,omitempty"`
	Status               int8   `bson:"status" json:"status"`
	Role                 int8   `bson:"role" json:"role"`
	HasShippingAddress   bool   `bson:"has_shipping_address" json:"has_shipping_address,omitempty"`
	ShippingName         string `bson:"shipping_name" json:"shipping_name,omitempty"`
	ShippingPhone        string `bson:"shipping_phone" json:"shipping_phone,omitempty"`
	ShippingCountry      string `bson:"shipping_country" json:"shipping_country,omitempty"`
	ShippingRegion       string `bson:"shipping_region" json:"shipping_region,omitempty"`
	ShippingCity         string `bson:"shipping_city" json:"shipping_city,omitempty"`
	ShippingPostalCode   string `bson:"shipping_postal_code" json:"shipping_postal_code,omitempty"`
	ShippingAddressLine1 string `bson:"shipping_address_line1" json:"shipping_address_line1,omitempty"`
	ShippingAddressLine2 string `bson:"shipping_address_line2" json:"shipping_address_line2,omitempty"`
}

func (impl *UserControllerImpl) userFromUpdateRequest(requestData *UserUpdateRequestIDO) (*user_s.User, error) {
	passwordHash, err := impl.Password.GenerateHashFromPassword(requestData.Password)
	if err != nil {
		impl.Logger.Error("hashing error", slog.Any("error", err))
		return nil, err
	}

	return &user_s.User{
		ID:                    requestData.ID,
		TenantID:              requestData.TenantID,
		FirstName:             requestData.FirstName,
		LastName:              requestData.LastName,
		Email:                 requestData.Email,
		PasswordHash:          passwordHash,
		PasswordHashAlgorithm: impl.Password.AlgorithmName(),
		Phone:                 requestData.Phone,
		Country:               requestData.Country,
		Region:                requestData.Region,
		City:                  requestData.City,
		PostalCode:            requestData.PostalCode,
		AddressLine1:          requestData.AddressLine1,
		AddressLine2:          requestData.AddressLine2,
		// HowDidYouHearAboutUs:      requestData.HowDidYouHearAboutUs,
		// HowDidYouHearAboutUsOther: requestData.HowDidYouHearAboutUsOther,
		AgreeTOS:             requestData.AgreeTOS,
		AgreePromotionsEmail: requestData.AgreePromotionsEmail,
		Status:               requestData.Status,
		Role:                 requestData.Role,
		HasShippingAddress:   requestData.HasShippingAddress,
		ShippingName:         requestData.ShippingName,
		ShippingPhone:        requestData.ShippingPhone,
		ShippingCountry:      requestData.ShippingCountry,
		ShippingRegion:       requestData.ShippingRegion,
		ShippingCity:         requestData.ShippingCity,
		ShippingPostalCode:   requestData.ShippingPostalCode,
		ShippingAddressLine1: requestData.ShippingAddressLine1,
		ShippingAddressLine2: requestData.ShippingAddressLine2,
	}, nil
}

func (impl *UserControllerImpl) UpdateByID(ctx context.Context, requestData *UserUpdateRequestIDO) (*user_s.User, error) {
	nu, err := impl.userFromUpdateRequest(requestData)
	if err != nil {
		return nil, err
	}

	// Extract from our session the following data.
	userID, _ := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	userName, _ := ctx.Value(constants.SessionUserName).(string)

	// Extract from our session the following data.
	userRole := ctx.Value(constants.SessionUserRole).(int8)

	// Apply filtering based on ownership and role.
	if userRole != user_s.UserRoleExecutive {
		return nil, httperror.NewForForbiddenWithSingleField("message", "you do not have permission")
	}

	// Lookup the user in our database, else return a `400 Bad Request` error.
	ou, err := impl.UserStorer.GetByID(ctx, nu.ID)
	if err != nil {
		impl.Logger.Error("database error", slog.Any("err", err))
		return nil, err
	}
	if ou == nil {
		impl.Logger.Warn("user does not exist validation error")
		return nil, httperror.NewForBadRequestWithSingleField("id", "does not exist")
	}

	// Lookup the tenant in our database, else return a `400 Bad Request` error.
	o, err := impl.TenantStorer.GetByID(ctx, nu.TenantID)
	if err != nil {
		impl.Logger.Error("database error", slog.Any("err", err))
		return nil, err
	}
	if o == nil {
		impl.Logger.Warn("tenant does not exist exists validation error")
		return nil, httperror.NewForBadRequestWithSingleField("tenant_id", "tenant does not exist")
	}

	ou.TenantID = o.ID
	ou.FirstName = nu.FirstName
	ou.LastName = nu.LastName
	ou.Name = fmt.Sprintf("%s %s", nu.FirstName, nu.LastName)
	ou.LexicalName = fmt.Sprintf("%s, %s", nu.LastName, nu.FirstName)
	ou.Email = nu.Email
	ou.Phone = nu.Phone
	ou.Country = nu.Country
	ou.Region = nu.Region
	ou.City = nu.City
	ou.PostalCode = nu.PostalCode
	ou.AddressLine1 = nu.AddressLine1
	ou.AddressLine2 = nu.AddressLine2
	// ou.HowDidYouHearAboutUs = nu.HowDidYouHearAboutUs
	// ou.HowDidYouHearAboutUsOther = nu.HowDidYouHearAboutUsOther
	ou.AgreePromotionsEmail = nu.AgreePromotionsEmail
	ou.ModifiedByUserID = userID
	ou.ModifiedByUserName = userName
	ou.HasShippingAddress = nu.HasShippingAddress
	ou.ShippingName = nu.ShippingName
	ou.ShippingPhone = nu.ShippingPhone
	ou.ShippingCountry = nu.ShippingCountry
	ou.ShippingRegion = nu.ShippingRegion
	ou.ShippingCity = nu.ShippingCity
	ou.ShippingPostalCode = nu.ShippingPostalCode
	ou.ShippingAddressLine1 = nu.ShippingAddressLine1
	ou.ShippingAddressLine2 = nu.ShippingAddressLine2

	if err := impl.UserStorer.UpdateByID(ctx, ou); err != nil {
		impl.Logger.Error("user update by id error", slog.Any("error", err))
		return nil, err
	}
	return ou, nil
}
