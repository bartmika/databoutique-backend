package controller

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"

	user_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
)

type ExecutiveVisitsTenantRequest struct {
	TenantID primitive.ObjectID `json:"tenant_id,omitempty"`
}

func (impl *GatewayControllerImpl) ExecutiveVisitsTenant(ctx context.Context, req *ExecutiveVisitsTenantRequest) error {
	////
	//// Extract the `sessionID` so we can process it.
	////

	sessionID := ctx.Value(constants.SessionID).(string)
	userRole := ctx.Value(constants.SessionUserRole).(int8)

	if userRole != user_s.UserRoleExecutive {
		impl.Logger.Error("not executive error", slog.Int("role", int(userRole)))
		return errors.New("user is not executive")
	}

	////
	//// Lookup in our in-memory the user record for the `sessionID` or error.
	////

	uBin, err := impl.Cache.Get(ctx, sessionID)
	if err != nil {
		impl.Logger.Error("in-memory set error", slog.Any("err", err))
		return err
	}

	var u *user_s.User
	err = json.Unmarshal(uBin, &u)
	if err != nil {
		impl.Logger.Error("unmarshal error", slog.Any("err", err))
		return err
	}

	////
	//// Set the user's logged in session to point to specific tenant.
	////

	// Set the tenant id of the user's authenticated session.
	u.TenantID = req.TenantID

	// Set expiry duration.
	expiry := 14 * 24 * time.Hour

	// Start our session using an access and refresh token.
	newSessionUUID := impl.UUID.NewUUID()

	// Save the session.
	if err := impl.Cache.SetWithExpiry(ctx, newSessionUUID, uBin, expiry); err != nil {
		impl.Logger.Error("cache set with expiry error", slog.Any("err", err))
		return err
	}

	return nil
}
