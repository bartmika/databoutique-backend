package controller

import (
	"context"
	"errors"
	"log/slog"

	"github.com/sashabaranov/go-openai"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	program_s "github.com/bartmika/databoutique-backend/internal/app/program/datastore"
)

func (impl *ExecutableControllerImpl) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	//
	// Get variables from our user authenticated session.
	//

	// tid, _ := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	// role, _ := ctx.Value(constants.SessionUserRole).(int8)
	// userID, _ := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	// userName, _ := ctx.Value(constants.SessionUserName).(string)
	// ipAddress, _ := ctx.Value(constants.SessionIPAddress).(string)

	// Keep data consistent.
	impl.Kmutex.Lockf("executable_%s", id.Hex())
	defer impl.Kmutex.Unlockf("executable_%s", id.Hex())

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
		return err
	}
	defer session.EndSession(ctx)

	// Define a transaction function with a series of operations
	transactionFunc := func(sessCtx mongo.SessionContext) (interface{}, error) {
		impl.Logger.Debug("begging to delete executable...",
			slog.String("executable_id", id.Hex()))

		////
		//// STEP 1: Lookup the record or error.
		////

		exec, err := impl.ExecutableStorer.GetByID(sessCtx, id)
		if err != nil {
			impl.Logger.Error("database get by id error", slog.Any("error", err))
			return nil, err
		}
		if exec == nil {
			impl.Logger.Error("database returns nothing from get by id")
			return nil, err
		}

		////
		//// STEP 2: Remove from OpenAI.
		////

		// --- Get credentials --- //

		creds, err := impl.TenantStorer.GetOpenAICredentialsByID(sessCtx, exec.TenantID)
		if err != nil {
			impl.Logger.Error("failed getting openai credentials",
				slog.Any("executable_id", exec.ID),
				slog.Any("error", err))
			return nil, err
		}
		if creds == nil {
			return nil, errors.New("no openai credentials returned")
		}

		impl.Logger.Debug("openai initializing...")
		client := openai.NewOrgClient(creds.APIKey, creds.OrgKey)
		impl.Logger.Debug("openai initialized")

		// --- Assistant --- //

		// Only delete the assistant from OpenAI if the program this executable
		// is based on is customer document review.
		if exec.ProgramBusinessFunction == program_s.ProgramBusinessFunctionCustomerDocumentReview {
			impl.Logger.Debug("beginning to delete assistant from openai...",
				slog.Any("executable_id", exec.ID))

			if _, err := client.DeleteAssistant(sessCtx, exec.OpenAIAssistantID); err != nil {
				impl.Logger.Error("failed deleting assistant from openai",
					slog.String("assistant_id", exec.OpenAIAssistantID),
					slog.String("executable_id", id.Hex()),
					slog.Any("error", err))
				return nil, err
			}

			impl.Logger.Debug("delete assistant from openai",
				slog.String("assistant_id", id.Hex()),
				slog.String("executable_id", exec.ID.Hex()))
		} else {
			impl.Logger.Debug("skipped deleting assistant from openai",
				slog.String("executable_id", exec.ID.Hex()))
		}

		// --- Threads --- //

		impl.Logger.Debug("beginning to delete thread(s) from openai...",
			slog.String("thread_id", exec.OpenAIAssistantThreadID))

		if _, err := client.DeleteThread(sessCtx, exec.OpenAIAssistantThreadID); err != nil {
			impl.Logger.Error("failed deleting thread from openai",
				slog.String("executable_id", exec.ID.Hex()),
				slog.String("thread_id", exec.OpenAIAssistantThreadID),
				slog.Any("error", err))
			return nil, err
		}

		impl.Logger.Debug("delete thread from openai",
			slog.String("executable_id", exec.ID.Hex()),
			slog.String("thread_id", exec.OpenAIAssistantThreadID))

		////
		//// STEP 2: Delete from database.
		////

		if err := impl.ExecutableStorer.DeleteByID(sessCtx, id); err != nil {
			impl.Logger.Error("database delete by id error", slog.Any("error", err))
			return nil, err
		}

		impl.Logger.Debug("delete executable",
			slog.String("executable_id", exec.ID.Hex()))

		return nil, nil
	}

	// Start a transaction
	if _, err := session.WithTransaction(ctx, transactionFunc); err != nil {
		impl.Logger.Error("session failed error",
			slog.Any("error", err))
		return err
	}

	return nil
}
