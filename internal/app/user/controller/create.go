package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"

	user_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

type UserCreateRequestIDO struct {
	TenantID     primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	FirstName    string             `json:"first_name"`
	LastName     string             `json:"last_name"`
	Email        string             `json:"email"`
	Phone        string             `json:"phone,omitempty"`
	Country      string             `json:"country,omitempty"`
	Region       string             `json:"region,omitempty"`
	City         string             `json:"city,omitempty"`
	PostalCode   string             `json:"postal_code,omitempty"`
	AddressLine1 string             `json:"address_line1,omitempty"`
	AddressLine2 string             `json:"address_line2,omitempty"`
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

func (impl *UserControllerImpl) userFromCreateRequest(requestData *UserCreateRequestIDO) (*user_s.User, error) {
	return &user_s.User{
		TenantID:     requestData.TenantID,
		FirstName:    requestData.FirstName,
		LastName:     requestData.LastName,
		Email:        requestData.Email,
		Phone:        requestData.Phone,
		Country:      requestData.Country,
		Region:       requestData.Region,
		City:         requestData.City,
		PostalCode:   requestData.PostalCode,
		AddressLine1: requestData.AddressLine1,
		AddressLine2: requestData.AddressLine2,
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

func (impl *UserControllerImpl) Create(ctx context.Context, requestData *UserCreateRequestIDO) (*user_s.User, error) {
	m, err := impl.userFromCreateRequest(requestData)
	if err != nil {
		return nil, err
	}

	// Extract from our session the following data.
	tid, _ := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	userRole, _ := ctx.Value(constants.SessionUserRole).(int8)
	userID, _ := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	userName, _ := ctx.Value(constants.SessionUserName).(string)

	// Apply filtering based on ownership and role.
	if userRole != user_s.UserRoleExecutive {
		return nil, httperror.NewForForbiddenWithSingleField("message", "you do not have permission")
	}

	// DEVELOPERS NOTE:
	// Every submission needs to have a unique `public id` (PID)
	// generated. The following needs to happen to generate the unique PID:
	// 1. Make the `Create` function be `atomic` and thus lock this function.
	// 2. Count total records in system (for particular tenant).
	// 3. Generate PID.
	// 4. Apply the PID to the record.
	// 5. Unlock this `Create` function to be usable again by other calls after
	//    the function successfully submits the record into our system.
	impl.Kmutex.Lockf("create-user-by-tenant-%s", tid.Hex())
	defer impl.Kmutex.Unlockf("create-user-by-tenant-%s", tid.Hex())

	// Lookup the user in our database, else return a `400 Bad Request` error.
	u, err := impl.UserStorer.GetByEmail(ctx, m.Email)
	if err != nil {
		impl.Logger.Error("database error", slog.Any("err", err))
		return nil, err
	}
	if u != nil {
		impl.Logger.Warn("user already exists validation error")
		return nil, httperror.NewForBadRequestWithSingleField("email", "email is not unique")
	}

	// Lookup the tenant in our database, else return a `400 Bad Request` error.
	o, err := impl.TenantStorer.GetByID(ctx, m.TenantID)
	if err != nil {
		impl.Logger.Error("database error", slog.Any("err", err))
		return nil, err
	}
	if o == nil {
		impl.Logger.Warn("tenant does not exist exists validation error")
		return nil, httperror.NewForBadRequestWithSingleField("tenant_id", "tenant does not exist")
	}

	// Modify the user based on role.

	// Add defaults.
	m.Email = strings.ToLower(m.Email)
	m.TenantID = o.ID
	m.ID = primitive.NewObjectID()
	m.CreatedAt = time.Now()
	m.CreatedByUserID = userID
	m.CreatedByUserName = userName
	m.ModifiedAt = time.Now()
	m.ModifiedByUserID = userID
	m.ModifiedByUserName = userName
	m.Name = fmt.Sprintf("%s %s", m.FirstName, m.LastName)
	m.LexicalName = fmt.Sprintf("%s, %s", m.LastName, m.FirstName)
	m.WasEmailVerified = true

	// Generate a temporary password.
	temporaryPassword := primitive.NewObjectID().Hex()

	// Hash our password with the temporary password and attach to account.
	temporaryPasswordHash, err := impl.Password.GenerateHashFromPassword(temporaryPassword)
	if err != nil {
		impl.Logger.Error("hashing error", slog.Any("error", err))
		return nil, err
	}
	m.PasswordHashAlgorithm = impl.Password.AlgorithmName()
	m.PasswordHash = temporaryPasswordHash

	// Save to our database.
	if err := impl.UserStorer.Create(ctx, m); err != nil {
		impl.Logger.Error("database create error", slog.Any("error", err))
		return nil, err
	}

	// Send email to user of the new password.
	if err := impl.TemplatedEmailer.SendNewUserTemporaryPasswordEmail(m.Email, m.FirstName, temporaryPassword); err != nil {
		impl.Logger.Error("failed sending verification email with error", slog.Any("err", err))
		return nil, err
	}

	return m, nil
}
