package controller

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	executable_s "github.com/bartmika/databoutique-backend/internal/app/executable/datastore"
	u_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

type ExecutableCreateRequestIDO struct {
	ProgramID          primitive.ObjectID   `bson:"program_id" json:"program_id"`
	Question           string               `bson:"question" json:"question"`
	UserID             primitive.ObjectID   `bson:"user_id" json:"user_id"`
	UploadDirectoryIDs []primitive.ObjectID `bson:"upload_directory_ids" json:"upload_directory_ids"`
}

func (impl *ExecutableControllerImpl) validateCreateRequest(ctx context.Context, dirtyData *ExecutableCreateRequestIDO) error {
	e := make(map[string]string)

	if dirtyData.ProgramID.IsZero() {
		e["program_id"] = "missing value"
	}
	if dirtyData.Question == "" {
		e["question"] = "missing value"
	}
	if dirtyData.UserID.IsZero() {
		e["user_id"] = "missing value"
	}
	if len(dirtyData.UploadDirectoryIDs) == 0 {
		e["upload_directory_ids"] = "missing value"
	}

	if len(e) != 0 {
		return httperror.NewForBadRequest(&e)
	}
	return nil
}

func (impl *ExecutableControllerImpl) Create(ctx context.Context, requestData *ExecutableCreateRequestIDO) (*executable_s.Executable, error) {
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
	impl.Kmutex.Lockf("create-program-by-tenant-%s", tid.Hex())
	defer impl.Kmutex.Unlockf("create-program-by-tenant-%s", tid.Hex())

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
		//// Get related data.
		////

		u, err := impl.UserStorer.GetByID(sessCtx, requestData.UserID)
		if err != nil {
			impl.Logger.Error("failed getting user",
				slog.Any("error", err))
			return nil, err
		}
		if u == nil {
			err := fmt.Errorf("user does not exist for id: %v", requestData.UserID.Hex())
			impl.Logger.Error("user does not exist", slog.Any("error", err))
			return nil, err
		}
		p, err := impl.ProgramStorer.GetByID(sessCtx, requestData.ProgramID)
		if err != nil {
			impl.Logger.Error("failed getting program",
				slog.Any("error", err))
			return nil, err
		}
		if p == nil {
			err := fmt.Errorf("program does not exist for id: %v", requestData.ProgramID.Hex())
			impl.Logger.Error("program does not exist", slog.Any("error", err))
			return nil, err
		}

		//TODO: Get folder and files.

		////
		//// Create record.
		////

		exec := &executable_s.Executable{}

		// Add defaults.
		exec.TenantID = tid
		exec.ID = primitive.NewObjectID()
		exec.CreatedAt = time.Now()
		exec.CreatedByUserID = userID
		exec.CreatedByUserName = userName
		exec.CreatedFromIPAddress = ipAddress
		exec.ModifiedAt = time.Now()
		exec.ModifiedByUserID = userID
		exec.ModifiedByUserName = userName
		exec.ModifiedFromIPAddress = ipAddress

		// Add base.
		exec.ProgramID = p.ID
		exec.ProgramName = p.Name
		exec.Question = requestData.Question
		exec.Status = executable_s.ExecutableStatusProcessing
		exec.UploadDirectories = make([]*executable_s.UploadFolderOption, 0)
		exec.OpenAIAssistantID = ""
		exec.UserID = u.ID
		exec.UserName = u.Name
		exec.UserLexicalName = u.LexicalName

		// Add related.
		//TODO: Impl.

		// Save to our database.
		if err := impl.ExecutableStorer.Create(sessCtx, exec); err != nil {
			impl.Logger.Error("database create error", slog.Any("error", err))
			return nil, err
		}

		////
		//// Exit our transaction successfully.
		////

		return exec, nil
	}

	// Start a transaction
	result, err := session.WithTransaction(ctx, transactionFunc)
	if err != nil {
		impl.Logger.Error("session failed error",
			slog.Any("error", err))
		return nil, err
	}

	// Convert from MongoDB transaction format into our data format.
	exec := result.(*executable_s.Executable)

	////
	//// Execute in background calling OpenAI API.
	////

	// Submit the following into the background of this web-application.
	// This function will run independently of this function call.
	go func(ex *executable_s.Executable) {
		if err := impl.CreateExecutableInBackgroundForOpenAI(ex); err != nil {
			impl.Logger.Error("failed polling openai", slog.Any("error", err))
		}
	}(exec)

	return exec, nil
}
