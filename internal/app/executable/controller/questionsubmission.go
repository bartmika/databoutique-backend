package controller

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	executable_s "github.com/bartmika/databoutique-backend/internal/app/executable/datastore"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

type QuestionSubmissionOperationRequestIDO struct {
	ExecutableID primitive.ObjectID `bson:"executable_id" json:"executable_id"`
	Content      string             `bson:"content" json:"content"`
}

func (impl *ExecutableControllerImpl) validateQuestionSubmissionOperationRequest(ctx context.Context, dirtyData *QuestionSubmissionOperationRequestIDO) error {
	e := make(map[string]string)

	if dirtyData.ExecutableID.IsZero() {
		e["executable_id"] = "missing value"
	}
	if dirtyData.Content == "" {
		e["content"] = "missing value"
	}
	if len(e) != 0 {
		return httperror.NewForBadRequest(&e)
	}
	return nil
}

func (impl *ExecutableControllerImpl) QuestionSubmissionOperation(ctx context.Context, requestData *QuestionSubmissionOperationRequestIDO) (*executable_s.Executable, error) {
	//
	// Get variables from our user authenticated session.
	//

	// tid, _ := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	// // role, _ := ctx.Value(constants.SessionUserRole).(int8)
	// userID, _ := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	// userName, _ := ctx.Value(constants.SessionUserName).(string)
	// ipAddress, _ := ctx.Value(constants.SessionIPAddress).(string)

	if err := impl.validateQuestionSubmissionOperationRequest(ctx, requestData); err != nil {
		impl.Logger.Error("validation error", slog.Any("error", err))
		return nil, err
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
	impl.Kmutex.Lockf("executable_%s", requestData.ExecutableID.Hex())
	defer impl.Kmutex.Unlockf("executable_%s", requestData.ExecutableID.Hex())

	//
	// Perform our validation and return validation error on any issues detected.
	//

	// switch role {
	// case u_s.UserRoleExecutive, u_s.UserRoleManagement, u_s.UserRoleFrontlineStaff:
	// 	break
	// default:
	// 	impl.Logger.Error("you do not have permission to create a client")
	// 	return nil, httperror.NewForForbiddenWithSingleField("message", "you do not have permission to create a client")
	// }

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

		exec, err := impl.ExecutableStorer.GetByID(sessCtx, requestData.ExecutableID)
		if err != nil {
			impl.Logger.Error("failed getting user",
				slog.Any("error", err))
			return nil, err
		}
		if exec == nil {
			err := fmt.Errorf("executable does not exist for id: %v", requestData.ExecutableID.Hex())
			impl.Logger.Error("executable does not exist", slog.Any("error", err))
			return nil, err
		}

		////
		//// Create database records.
		////

		msg1 := &executable_s.Message{
			ID:              primitive.NewObjectID(),
			Content:         requestData.Content,
			OpenAIMessageID: "",
			CreatedAt:       time.Now(),
			Status:          executable_s.ExecutableStatusActive,
			FromExecutable:  false,
		}
		msg2 := &executable_s.Message{
			ID:              primitive.NewObjectID(),
			Content:         requestData.Content,
			OpenAIMessageID: "",
			CreatedAt:       time.Now(),
			Status:          executable_s.ExecutableStatusProcessing,
			FromExecutable:  true,
		}
		exec.Messages = append(exec.Messages, msg1)
		exec.Messages = append(exec.Messages, msg2)
		exec.Status = executable_s.ExecutableStatusProcessing

		// Save to our database.
		if err := impl.ExecutableStorer.UpdateByID(sessCtx, exec); err != nil {
			impl.Logger.Error("database create error",
				slog.Any("error", err))
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
		if err := impl.processQuestionSubmissionInBackgroundForOpenAI(ex); err != nil {
			impl.Logger.Error("failed submitting to openai", slog.Any("error", err))
		}
	}(exec)

	return exec, nil
}
