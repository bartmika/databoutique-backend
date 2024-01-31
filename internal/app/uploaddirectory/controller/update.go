package controller

import (
	"context"
	"time"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	uploaddirectory_s "github.com/bartmika/databoutique-backend/internal/app/uploaddirectory/datastore"
	u_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

type UploadDirectoryUpdateRequestIDO struct {
	ID          primitive.ObjectID `bson:"id" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	SortNumber  int8               `bson:"sort_number" json:"sort_number"`
}

func (impl *UploadDirectoryControllerImpl) validateUpdateRequest(ctx context.Context, dirtyData *UploadDirectoryUpdateRequestIDO) error {
	e := make(map[string]string)

	if dirtyData.Name == "" {
		e["name"] = "missing value"
	}

	if len(e) != 0 {
		return httperror.NewForBadRequest(&e)
	}
	return nil
}

func (impl *UploadDirectoryControllerImpl) UpdateByID(ctx context.Context, requestData *UploadDirectoryUpdateRequestIDO) (*uploaddirectory_s.UploadDirectory, error) {
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

		// Lookup the uploaddirectory in our database, else return a `400 Bad Request` error.
		hh, err := impl.UploadDirectoryStorer.GetByID(sessCtx, requestData.ID)
		if err != nil {
			impl.Logger.Error("database error", slog.Any("err", err))
			return nil, err
		}
		if hh == nil {
			impl.Logger.Warn("uploaddirectory does not exist validation error")
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
		hh.Name = requestData.Name
		hh.Description = requestData.Description
		hh.SortNumber = requestData.SortNumber

		if err := impl.UploadDirectoryStorer.UpdateByID(sessCtx, hh); err != nil {
			impl.Logger.Error("uploaddirectory update by id error", slog.Any("error", err))
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

	return result.(*uploaddirectory_s.UploadDirectory), nil
}
