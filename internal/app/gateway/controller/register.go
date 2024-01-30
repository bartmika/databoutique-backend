package controller

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
	_ "time/tzdata"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	gateway_s "github.com/bartmika/databoutique-backend/internal/app/gateway/datastore"
	user_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

type UserRegisterRequestIDO struct {
	TenantID                    primitive.ObjectID `json:"tenant_id,omitempty"`
	FirstName                   string             `json:"first_name"`
	LastName                    string             `json:"last_name"`
	Email                       string             `json:"email"`
	EmailRepeated               string             `json:"email_repeated"`
	IsOkToEmail                 bool               `json:"is_ok_to_email"`
	Password                    string             `json:"password"`
	PasswordRepeated            string             `json:"password_repeated"`
	Phone                       string             `json:"phone,omitempty"`
	PhoneType                   int8               `json:"phone_type"`
	PhoneExtension              string             `json:"phone_extension"`
	IsOkToText                  bool               `json:"is_ok_to_text"`
	OtherPhone                  string             `json:"other_phone"`
	OtherPhoneExtension         string             `json:"other_phone_extension"`
	OtherPhoneType              int8               `json:"other_phone_type"`
	Country                     string             `json:"country,omitempty"`
	Region                      string             `json:"region,omitempty"`
	City                        string             `json:"city,omitempty"`
	PostalCode                  string             `json:"postal_code,omitempty"`
	AddressLine1                string             `json:"address_line_1,omitempty"`
	AddressLine2                string             `json:"address_line_2,omitempty"`
	StoreLogo                   string             `json:"store_logo,omitempty"`
	HowDidYouHearAboutUsID      primitive.ObjectID `bson:"how_did_you_hear_about_us_id" json:"how_did_you_hear_about_us_id,omitempty"`
	HowDidYouHearAboutUsText    string             `bson:"how_did_you_hear_about_us_text" json:"how_did_you_hear_about_us_text,omitempty"`
	IsHowDidYouHearAboutUsOther bool               `bson:"is_how_did_you_hear_about_us_other" json:"is_how_did_you_hear_about_us_other,omitempty"`
	HowDidYouHearAboutUsOther   string             `bson:"how_did_you_hear_about_us_other" json:"how_did_you_hear_about_us_other,omitempty"`
	AgreeTOS                    bool               `bson:"agree_tos" json:"agree_tos,omitempty"`
	AgreePromotionsEmail        bool               `json:"agree_promotions_email,omitempty"`
	AgreeWaiver                 bool               `bson:"agree_waiver" json:"agree_waiver,omitempty"`
	HasShippingAddress          bool               `bson:"has_shipping_address" json:"has_shipping_address,omitempty"`
	ShippingName                string             `bson:"shipping_name" json:"shipping_name,omitempty"`
	ShippingPhone               string             `bson:"shipping_phone" json:"shipping_phone,omitempty"`
	ShippingCountry             string             `bson:"shipping_country" json:"shipping_country,omitempty"`
	ShippingRegion              string             `bson:"shipping_region" json:"shipping_region,omitempty"`
	ShippingCity                string             `bson:"shipping_city" json:"shipping_city,omitempty"`
	ShippingPostalCode          string             `bson:"shipping_postal_code" json:"shipping_postal_code,omitempty"`
	ShippingAddressLine1        string             `bson:"shipping_address_line1" json:"shipping_address_line1,omitempty"`
	ShippingAddressLine2        string             `bson:"shipping_address_line2" json:"shipping_address_line2,omitempty"`
	HasCouponPromotionalCode    int8               `json:"has_coupon_promotional_code,omitempty"`
	CouponPromotionalCode       string             `bson:"coupon_promotional_code" json:"coupon_promotional_code,omitempty"`
	BirthDate                   time.Time          `json:"birth_date"`
	JoinDate                    time.Time          `json:"join_date"`
	Gender                      int8               `bson:"gender" json:"gender"`
	GenderOther                 string             `bson:"gender_other" json:"gender_other"`
}

type UserRegisterResponseIDO struct {
	User                   *user_s.User `json:"user"`
	AccessToken            string       `json:"access_token"`
	AccessTokenExpiryTime  time.Time    `json:"access_token_expiry_time"`
	RefreshToken           string       `json:"refresh_token"`
	RefreshTokenExpiryTime time.Time    `json:"refresh_token_expiry_time"`
}

func (impl *GatewayControllerImpl) UserRegister(ctx context.Context, req *UserRegisterRequestIDO) (*gateway_s.LoginResponseIDO, error) {
	// Defensive Code: For security purposes we need to remove all whitespaces from the email and lower the characters.
	req.Email = strings.ToLower(req.Email)
	req.Password = strings.ReplaceAll(req.Password, " ", "")

	if err := validateUserRegisterRequest(req); err != nil {
		return nil, err
	}

	impl.Kmutex.Lockf("REGISTRATION-WITH-EMAIL-%v", req.Email)
	defer impl.Kmutex.Unlockf("REGISTRATION-WITH-EMAIL-%v", req.Email)

	////
	//// Start the transaction.
	////

	session, err := impl.DbClient.StartSession()
	if err != nil {
		impl.Logger.Error("start session error",
			slog.Any("error", err))
		return nil, err
	}
	defer session.EndSession(ctx)

	// Define a transaction function with a series of operations
	transactionFunc := func(sessCtx mongo.SessionContext) (interface{}, error) {

		// Lookup the user in our database, else return a `400 Bad Request` error.
		u, err := impl.UserStorer.GetByEmail(sessCtx, req.Email)
		if err != nil {
			impl.Logger.Error("database error",
				slog.Any("err", err),
				slog.String("Email", req.Email))
			return nil, err
		}
		if u != nil {
			impl.Logger.Warn("user already exists validation error",
				slog.String("Email", req.Email))
			return nil, httperror.NewForBadRequestWithSingleField("email", "email is not unique")
		}

		// Create our user.
		u, err = impl.createUserForRequest(sessCtx, req)
		if err != nil {
			return nil, err
		}

		// // Send our verification email.
		// if err := impl.TemplatedEmailer.SendMemberVerificationEmail(u.Email, u.EmailVerificationCode, u.FirstName, u.TenantName); err != nil {
		// 	impl.Logger.Error("failed sending verification email with error from registration",
		// 		slog.Any("err", err),
		// 		slog.String("Email", u.Email),
		// 		slog.Any("UserID", u.ID))
		// 	// Do not send error message to user nor abort the registration process.
		// 	// Just simply log an error message and continue.
		// }
		return nil, nil
	}

	// Start a transaction
	if _, err := session.WithTransaction(ctx, transactionFunc); err != nil {
		impl.Logger.Error("session failed error",
			slog.Any("error", err))
		return nil, err
	}

	return impl.Login(ctx, req.Email, req.Password)
}

func (impl *GatewayControllerImpl) createUserForRequest(sessCtx mongo.SessionContext, req *UserRegisterRequestIDO) (*user_s.User, error) {
	// Hash the password for security purposes.
	passwordHash, err := impl.Password.GenerateHashFromPassword(req.Password)
	if err != nil {
		impl.Logger.Error("hashing error", slog.Any("error", err))
		return nil, err
	}

	// Make shipping and billing same if not selected.
	if req.HasShippingAddress == false {
		req.ShippingCity = req.City
		req.ShippingCountry = req.Country
		req.ShippingAddressLine1 = req.AddressLine1
		req.ShippingAddressLine2 = req.AddressLine2
		req.ShippingPostalCode = req.PostalCode
		req.ShippingRegion = req.Region
	}

	// // Create an account with our payment processor.
	// var paymentProcessorCustomerID *string
	// paymentProcessorCustomerID, err = impl.PaymentProcessor.CreateCustomer(
	// 	fmt.Sprintf("%s %s", req.FirstName, req.LastName),
	// 	req.Email,
	// 	"", // description...
	// 	fmt.Sprintf("%s %s Shipping Address", req.FirstName, req.LastName),
	// 	req.Phone,
	// 	// req.ShippingCity, req.ShippingCountry, req.ShippingAddressLine1, req.ShippingAddressLine2, req.ShippingPostalCode, req.ShippingRegion, // Shipping
	// 	"", "", "", "", "", "",
	// 	req.City, req.Country, req.AddressLine1, req.AddressLine2, req.PostalCode, req.Region, // Billing
	// )
	// if err != nil {
	// 	impl.Logger.Error("creating customer from payment processor error", slog.Any("error", err))
	// 	return nil, err
	// }

	// Generate the unique identifier used for MongoDB.
	userID := primitive.NewObjectID()

	// Use the user's provided time zone or default to UTC.
	location, _ := time.LoadLocation("UTC")

	//
	// Extract the tenant.
	//

	tenant, err := impl.TenantStorer.GetByID(sessCtx, req.TenantID)
	if err != nil {
		return nil, err
	}

	// If unspecified the tenant then auto-assign the default tenant in our app.
	if tenant == nil {
		tenant, err = impl.TenantStorer.GetByID(sessCtx, impl.Config.InitialAccount.AdminTenantID)
		if err != nil {
			return nil, err
		}
	}

	//
	// Create our user
	//

	u := &user_s.User{
		TenantID:                tenant.ID,
		TenantName:              tenant.Name,
		ID:                      userID,
		FirstName:               req.FirstName,
		LastName:                req.LastName,
		Name:                    fmt.Sprintf("%s %s", req.FirstName, req.LastName),
		LexicalName:             fmt.Sprintf("%s, %s", req.LastName, req.FirstName),
		Email:                   req.Email,
		PasswordHash:            passwordHash,
		PasswordHashAlgorithm:   impl.Password.AlgorithmName(),
		Role:                    user_s.UserRoleCustomer,
		Phone:                   req.Phone,
		Country:                 req.Country,
		Region:                  req.Region,
		City:                    req.City,
		PostalCode:              req.PostalCode,
		AddressLine1:            req.AddressLine1,
		AgreeTOS:                req.AgreeTOS,
		TOSVersion:              "January, 2024",
		TOSText:                 "XXX",
		TOSAgreedOn:             time.Now().In(location),
		PrivacyVersion:          "January, 2024",
		PrivacyText:             "yyy",
		PrivacyAgreedOn:         time.Now().In(location),
		AgreePromotionsEmail:    req.AgreePromotionsEmail,
		AgreeWaiver:             req.AgreeWaiver,
		WaiverText:              "",
		WaiverAgreedOn:          time.Now().In(location),
		CreatedByUserID:         userID,
		CreatedAt:               time.Now().In(location),
		CreatedByUserName:       fmt.Sprintf("%s %s", req.FirstName, req.LastName),
		ModifiedByUserID:        userID,
		ModifiedAt:              time.Now().In(location),
		ModifiedByUserName:      fmt.Sprintf("%s %s", req.FirstName, req.LastName),
		WasEmailVerified:        false,
		EmailVerificationCode:   impl.UUID.NewUUID(),
		EmailVerificationExpiry: time.Now().Add(72 * time.Hour),
		Status:                  user_s.UserStatusActive,
		// PaymentProcessorName:       b.PaymentProcessorName, // Attach the required payment process
		// PaymentProcessorCustomerID: *paymentProcessorCustomerID,
		HasShippingAddress:       req.HasShippingAddress,
		ShippingName:             req.ShippingName,
		ShippingPhone:            req.ShippingPhone,
		ShippingCountry:          req.ShippingCountry,
		ShippingRegion:           req.ShippingRegion,
		ShippingCity:             req.ShippingCity,
		ShippingPostalCode:       req.ShippingPostalCode,
		ShippingAddressLine1:     req.ShippingAddressLine1,
		ShippingAddressLine2:     req.ShippingAddressLine2,
		HasCouponPromotionalCode: user_s.UserHasCouponPromotionalCodeNo,
		Coupons:                  make([]*user_s.UserClaimedCoupon, 0),
	}

	//
	// Extract the how did you hear about us.
	//

	hh, err := impl.HowHearAboutUsItemStorer.GetByID(sessCtx, req.HowDidYouHearAboutUsID)
	if err != nil {
		impl.Logger.Error("fetching how hear error", slog.Any("error", err))
		return nil, err
	}
	if hh == nil {
		impl.Logger.Error("how hear does not exist error", slog.Any("tagID", req.HowDidYouHearAboutUsID))
		return nil, httperror.NewForBadRequestWithSingleField("tags", req.HowDidYouHearAboutUsID.Hex()+" how hear does not exist")
	}
	u.HowDidYouHearAboutUsID = hh.ID
	u.HowDidYouHearAboutUsText = hh.Text
	u.IsHowDidYouHearAboutUsOther = req.IsHowDidYouHearAboutUsOther
	u.HowDidYouHearAboutUsOther = req.HowDidYouHearAboutUsOther

	err = impl.UserStorer.Create(sessCtx, u)
	if err != nil {
		impl.Logger.Error("database create error",
			slog.String("user_email", u.Email),
			slog.Any("error", err))
		return nil, err
	}
	impl.Logger.Info("User created.",
		slog.Any("tenant_id", u.TenantID),
		slog.Any("user_id", u.ID),
		slog.String("user_full_name", u.Name),
		slog.String("user_email", u.Email),
		slog.String("user_password_hash_algorithm", u.PasswordHashAlgorithm))

	return u, nil
}

func validateUserRegisterRequest(dirtyData *UserRegisterRequestIDO) error {
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
	if dirtyData.EmailRepeated == "" {
		e["email_repeated"] = "missing value"
	}
	if len(dirtyData.EmailRepeated) > 255 {
		e["email_repeated"] = "too long"
	}
	if dirtyData.Email != dirtyData.EmailRepeated {
		e["email"] = "does not match email repeated"
		e["email_repeated"] = "does not match email"
	}
	if dirtyData.Password == "" {
		e["password"] = "missing value"
	}
	if dirtyData.PasswordRepeated == "" {
		e["password_repeated"] = "missing value"
	}
	if dirtyData.PasswordRepeated != dirtyData.Password {
		e["password"] = "does not match"
		e["password_repeated"] = "does not match"
	}
	if dirtyData.Phone == "" {
		e["phone"] = "missing value"
	}
	if dirtyData.PhoneType == 0 {
		e["phone_type"] = "missing value"
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
	if dirtyData.Password == "" {
		e["password"] = "missing value"
	}
	if dirtyData.AddressLine1 == "" {
		e["address_line_1"] = "missing value"
	}
	// if dirtyData.HowDidYouHearAboutUs == 0 {
	// 	e["how_did_you_hear_about_us"] = "missing value"
	// }
	if dirtyData.AgreeTOS == false {
		e["agree_tos"] = "you must agree to the terms before proceeding"
	}
	if dirtyData.AgreeWaiver == false {
		e["agree_waiver"] = "you must agree to the waiver before proceeding"
	}

	if dirtyData.HowDidYouHearAboutUsID.IsZero() {
		e["how_did_you_hear_about_us_id"] = "missing value"
	} else {
		if dirtyData.IsHowDidYouHearAboutUsOther {
			if dirtyData.HowDidYouHearAboutUsOther == "" {
				e["how_did_you_hear_about_us_other"] = "missing value"
			}
		}
	}
	if dirtyData.Gender == 1 && dirtyData.GenderOther == "" {
		e["gender_other"] = "missing value"
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
	if dirtyData.HasCouponPromotionalCode == 0 {
		e["has_coupon_promotional_code"] = "missing value"
	}
	if dirtyData.HasCouponPromotionalCode == 1 && dirtyData.CouponPromotionalCode == "" {
		e["coupon_promotional_code"] = "missing value"
	}
	if len(e) != 0 {
		return httperror.NewForBadRequest(&e)
	}
	return nil
}
