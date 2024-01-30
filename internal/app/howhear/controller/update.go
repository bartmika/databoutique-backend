package controller

import (
	"context"
	"time"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	howhear_s "github.com/bartmika/databoutique-backend/internal/app/howhear/datastore"
	u_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

type HowHearAboutUsItemUpdateRequestIDO struct {
	ID             primitive.ObjectID `bson:"id" json:"id"`
	Text           string             `bson:"text" json:"text"`
	SortNumber     int8               `bson:"sort_number" json:"sort_number"`
	IsForAssociate bool               `bson:"is_for_associate" json:"is_for_associate"`
	IsForCustomer  bool               `bson:"is_for_customer" json:"is_for_customer"`
	IsForStaff     bool               `bson:"is_for_staff" json:"is_for_staff"`
}

func (impl *HowHearAboutUsItemControllerImpl) validateUpdateRequest(ctx context.Context, dirtyData *HowHearAboutUsItemUpdateRequestIDO) error {
	e := make(map[string]string)

	if dirtyData.Text == "" {
		e["text"] = "missing value"
	}
	if dirtyData.SortNumber == 0 {
		e["sort_number"] = "missing value"
	}

	if len(e) != 0 {
		return httperror.NewForBadRequest(&e)
	}
	return nil
}

func (impl *HowHearAboutUsItemControllerImpl) UpdateByID(ctx context.Context, requestData *HowHearAboutUsItemUpdateRequestIDO) (*howhear_s.HowHearAboutUsItem, error) {
	//
	// Perform our validation and return validation error on any issues detected.
	//

	if err := impl.validateUpdateRequest(ctx, requestData); err != nil {
		impl.Logger.Error("validation error", slog.Any("error", err))
		return nil, err
	}

	// Get variables from our user authenticated session.
	//

	tid, _ := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	role, _ := ctx.Value(constants.SessionUserRole).(int8)
	userID, _ := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	userName, _ := ctx.Value(constants.SessionUserName).(string)
	ipAddress, _ := ctx.Value(constants.SessionIPAddress).(string)

	switch role {
	case u_s.UserRoleExecutive, u_s.UserRoleManagement, u_s.UserRoleFrontlineStaff:
		break
	default:
		impl.Logger.Error("you do not have permission to create a client")
		return nil, httperror.NewForForbiddenWithSingleField("message", "you do not have permission to create a client")
	}

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

		////
		//// Get data.
		////

		// Lookup the howhear in our database, else return a `400 Bad Request` error.
		hh, err := impl.HowHearAboutUsItemStorer.GetByID(sessCtx, requestData.ID)
		if err != nil {
			impl.Logger.Error("database error", slog.Any("err", err))
			return nil, err
		}
		if hh == nil {
			impl.Logger.Warn("howhear does not exist validation error")
			return nil, httperror.NewForBadRequestWithSingleField("id", "does not exist")
		}

		////
		//// Update primary record.
		////

		// Base
		hh.TenantID = tid
		hh.ModifiedAt = time.Now()
		hh.ModifiedByUserID = userID
		hh.ModifiedByUserName = userName
		hh.ModifiedFromIPAddress = ipAddress

		// Content
		hh.Text = requestData.Text
		hh.SortNumber = requestData.SortNumber
		hh.IsForAssociate = requestData.IsForAssociate
		hh.IsForCustomer = requestData.IsForCustomer
		hh.IsForStaff = requestData.IsForStaff

		if err := impl.HowHearAboutUsItemStorer.UpdateByID(sessCtx, hh); err != nil {
			impl.Logger.Error("howhear update by id error", slog.Any("error", err))
			return nil, err
		}

		////
		//// Update related records.
		////

		//

		////
		//// Exit our transaction successfully.
		////

		return hh, nil
	}

	// Start a transaction
	result, err := session.WithTransaction(ctx, transactionFunc)
	if err != nil {
		impl.Logger.Error("session failed error",
			slog.Any("error", err))
		return nil, err
	}

	return result.(*howhear_s.HowHearAboutUsItem), nil
}
