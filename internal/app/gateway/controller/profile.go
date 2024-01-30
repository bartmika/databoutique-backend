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

func (impl *GatewayControllerImpl) Profile(ctx context.Context) (*user_s.User, error) {
	// Extract from our session the following data.
	userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)

	// Lookup the user in our database, else return a `400 Bad Request` error.
	u, err := impl.UserStorer.GetByID(ctx, userID)
	if err != nil {
		impl.Logger.Error("database error", slog.Any("err", err))
		return nil, err
	}
	if u == nil {
		impl.Logger.Warn("user does not exist validation error")
		return nil, httperror.NewForBadRequestWithSingleField("id", "does not exist")
	}
	return u, nil
}

func (impl *GatewayControllerImpl) ProfileUpdate(ctx context.Context, nu *user_s.User) error {
	// Extract from our session the following data.
	userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)

	// Lookup the user in our database, else return a `400 Bad Request` error.
	ou, err := impl.UserStorer.GetByID(ctx, userID)
	if err != nil {
		impl.Logger.Error("database error", slog.Any("err", err))
		return err
	}
	if ou == nil {
		impl.Logger.Warn("user does not exist validation error")
		return httperror.NewForBadRequestWithSingleField("id", "does not exist")
	}

	hh, err := impl.HowHearAboutUsItemStorer.GetByID(ctx, nu.HowDidYouHearAboutUsID)
	if err != nil {
		impl.Logger.Error("fetching how hear error", slog.Any("error", err))
		return err
	}
	if hh == nil {
		impl.Logger.Error("how hear does not exist error", slog.Any("tagID", nu.HowDidYouHearAboutUsID))
		return httperror.NewForBadRequestWithSingleField("tags", nu.HowDidYouHearAboutUsID.Hex()+" how hear does not exist")
	}
	ou.HowDidYouHearAboutUsID = hh.ID
	ou.HowDidYouHearAboutUsText = hh.Text
	ou.IsHowDidYouHearAboutUsOther = nu.IsHowDidYouHearAboutUsOther
	ou.HowDidYouHearAboutUsOther = nu.HowDidYouHearAboutUsOther

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
	ou.AgreePromotionsEmail = nu.AgreePromotionsEmail
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
		return err
	}
	return nil
}

type ProfileChangePasswordRequestIDO struct {
	OldPassword      string `json:"old_password"`
	Password         string `json:"password"`
	PasswordRepeated string `json:"password_repeated"`
}

func ValidateProfileChangePassworRequest(dirtyData *ProfileChangePasswordRequestIDO) error {
	e := make(map[string]string)

	if dirtyData.OldPassword == "" {
		e["old_password"] = "missing value"
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

	if len(e) != 0 {
		return httperror.NewForBadRequest(&e)
	}
	return nil
}

func (impl *GatewayControllerImpl) ProfileChangePassword(ctx context.Context, req *ProfileChangePasswordRequestIDO) error {
	// Extract from our session the following data.
	userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)

	// Lookup the user in our database, else return a `400 Bad Request` error.
	u, err := impl.UserStorer.GetByID(ctx, userID)
	if err != nil {
		impl.Logger.Error("database error", slog.Any("err", err))
		return err
	}
	if u == nil {
		impl.Logger.Warn("user does not exist validation error")
		return httperror.NewForBadRequestWithSingleField("id", "does not exist")
	}

	if err := ValidateProfileChangePassworRequest(req); err != nil {
		impl.Logger.Warn("user validation failed", slog.Any("err", err))
		return err
	}

	// Verify the inputted password and hashed password match.
	if passwordMatch, _ := impl.Password.ComparePasswordAndHash(req.OldPassword, u.PasswordHash); passwordMatch == false {
		impl.Logger.Warn("password check validation error")
		return httperror.NewForBadRequestWithSingleField("old_password", "old password do not match with record of existing password")
	}

	passwordHash, err := impl.Password.GenerateHashFromPassword(req.Password)
	if err != nil {
		impl.Logger.Error("hashing error", slog.Any("error", err))
		return err
	}
	u.PasswordHash = passwordHash
	u.PasswordHashAlgorithm = impl.Password.AlgorithmName()
	if err := impl.UserStorer.UpdateByID(ctx, u); err != nil {
		impl.Logger.Error("user update by id error", slog.Any("error", err))
		return err
	}
	return nil
}
