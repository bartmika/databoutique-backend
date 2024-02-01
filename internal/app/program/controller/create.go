package controller

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	program_s "github.com/bartmika/databoutique-backend/internal/app/program/datastore"
	u_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

type ProgramCreateRequestIDO struct {
	Name             string `bson:"name" json:"name"`
	Description      string `bson:"description" json:"description"`
	Instructions     string `bson:"instructions" json:"instructions"`
	Model            string `bson:"model" json:"model"`
	BusinessFunction int8   `bson:"business_function" json:"business_function"`
}

func (impl *ProgramControllerImpl) validateCreateRequest(ctx context.Context, dirtyData *ProgramCreateRequestIDO) error {
	e := make(map[string]string)

	if dirtyData.Name == "" {
		e["name"] = "missing value"
	}
	if dirtyData.Description == "" {
		e["description"] = "missing value"
	}
	if dirtyData.Instructions == "" {
		e["instructions"] = "missing value"
	}
	if dirtyData.Model == "" {
		e["model"] = "missing value"
	}
	if dirtyData.BusinessFunction == 0 {
		e["business_function"] = "missing value"
	}

	if len(e) != 0 {
		return httperror.NewForBadRequest(&e)
	}
	return nil
}

func (impl *ProgramControllerImpl) Create(ctx context.Context, requestData *ProgramCreateRequestIDO) (*program_s.Program, error) {
	//
	// Get variables from our user authenticated session.
	//

	tid, _ := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	role, _ := ctx.Value(constants.SessionUserRole).(int8)
	userID, _ := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	userName, _ := ctx.Value(constants.SessionUserName).(string)
	ipAddress, _ := ctx.Value(constants.SessionIPAddress).(string)

	// DEVELOPERS NOTE:
	// Every submission needs to have a unique `public id` (PID)
	// generated. The following needs to happen to generate the unique PID:
	// 1. Make the `Create` function be `atomic` and thus lock this function.
	// 2. Count total records in system (for particular tenant).
	// 3. Generate PID.
	// 4. Apply the PID to the record.
	// 5. Unlock this `Create` function to be usable again by other calls after
	//    the function successfully submits the record into our system.
	impl.Kmutex.Lockf("create-how-hear-about-us-item-by-tenant-%s", tid.Hex())
	defer impl.Kmutex.Unlockf("create-how-hear-about-us-item-by-tenant-%s", tid.Hex())

	//
	// Perform our validation and return validation error on any issues detected.
	//

	if err := impl.validateCreateRequest(ctx, requestData); err != nil {
		impl.Logger.Error("validation error", slog.Any("error", err))
		return nil, err
	}

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
		//// Create the record.
		////

		prog := &program_s.Program{
			Name:             requestData.Name,
			Description:      requestData.Description,
			Instructions:     requestData.Instructions,
			Model:            requestData.Model,
			Status:           program_s.ProgramStatusActive,
			BusinessFunction: requestData.BusinessFunction,
		}

		// Add defaults.
		prog.TenantID = tid
		prog.ID = primitive.NewObjectID()
		prog.CreatedAt = time.Now()
		prog.CreatedByUserID = userID
		prog.CreatedByUserName = userName
		prog.CreatedFromIPAddress = ipAddress
		prog.ModifiedAt = time.Now()
		prog.ModifiedByUserID = userID
		prog.ModifiedByUserName = userName
		prog.ModifiedFromIPAddress = ipAddress

		// Save to our database.
		if err := impl.ProgramStorer.Create(sessCtx, prog); err != nil {
			impl.Logger.Error("database create error", slog.Any("error", err))
			return nil, err
		}

		////
		//// OpenAI
		////

		//TODO

		////
		//// Exit our transaction successfully.
		////

		return prog, nil
	}

	// Start a transaction
	result, err := session.WithTransaction(ctx, transactionFunc)
	if err != nil {
		impl.Logger.Error("session failed error",
			slog.Any("error", err))
		return nil, err
	}

	return result.(*program_s.Program), nil
}
