package controller

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	am_s "github.com/bartmika/databoutique-backend/internal/app/assistantmessage/datastore"
	at_s "github.com/bartmika/databoutique-backend/internal/app/assistantthread/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
	"github.com/sashabaranov/go-openai"
)

type AssistantMessageCreateRequestIDO struct {
	AssistantThreadID primitive.ObjectID `bson:"assistant_thread_id" json:"assistant_thread_id"`
	Text              string             `bson:"text" json:"text"`
}

func (impl *AssistantMessageControllerImpl) validateCreateRequest(ctx context.Context, dirtyData *AssistantMessageCreateRequestIDO) error {
	e := make(map[string]string)

	if dirtyData.AssistantThreadID.IsZero() {
		e["assistant_thread_id"] = "missing value"
	}
	if dirtyData.Text == "" {
		e["text"] = "missing value"
	}

	if len(e) != 0 {
		return httperror.NewForBadRequest(&e)
	}
	return nil
}

func (impl *AssistantMessageControllerImpl) Create(ctx context.Context, requestData *AssistantMessageCreateRequestIDO) (*am_s.AssistantMessage, error) {
	//
	// Get variables from our user authenticated session.
	//

	tid, _ := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	// role, _ := ctx.Value(constants.SessionUserRole).(int8)
	uid, _ := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	uname, _ := ctx.Value(constants.SessionUserName).(string)
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
	impl.Kmutex.Lockf("create-assistant-message-by-tenant-%s", tid.Hex())
	defer impl.Kmutex.Unlockf("create-assistant-message-by-tenant-%s", tid.Hex())

	//
	// Perform our validation and return validation error on any issues detected.
	//

	if err := impl.validateCreateRequest(ctx, requestData); err != nil {
		impl.Logger.Error("validation error", slog.Any("error", err))
		return nil, err
	}

	// switch role { //TODO: TECHDEBT.
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
		// Fetch our thread.

		at, err := impl.AssistantThreadStorer.GetByID(sessCtx, requestData.AssistantThreadID)
		if err != nil {
			impl.Logger.Error("failed getting assistant thread by id",
				slog.Any("error", err))
			return nil, err
		}
		if at == nil {
			err := fmt.Errorf("no assistant thread found with id: %s", requestData.AssistantThreadID.Hex())
			impl.Logger.Error("", slog.Any("error", err))
			return nil, err
		}

		// Fetch our OpenAI credentials which we will use to make our API call.

		creds, err := impl.TenantStorer.GetOpenAICredentialsByID(sessCtx, tid)
		if err != nil {
			impl.Logger.Error("failed getting openai credentials",
				slog.String("tenant_id", tid.Hex()),
				slog.Any("error", err))
			return nil, err
		}
		if creds == nil {
			return nil, errors.New("no openai credentials returned")
		}

		client := openai.NewOrgClient(creds.APIKey, creds.OrgKey)

		// The next lines of code will involve two steps:
		// Step 1: Create our question message of what the user asked.
		// Step 2: Create the empty assistant response message and submit into
		//         openAI to process in the background and return to us when
		//         it is completed.

		// This is our question message.
		am1 := &am_s.AssistantMessage{
			ID:                      primitive.NewObjectID(),
			TenantID:                tid,
			AssistantID:             at.AssistantID,
			AssistantThreadID:       at.ID,
			OpenAIAssistantID:       at.OpenAIAssistantID,
			OpenAIAssistantThreadID: at.OpenAIAssistantThreadID,
			Text:                    requestData.Text,
			FromAssistant:           false,
			Status:                  at_s.AssistantThreadStatusActive,
			CreatedAt:               time.Now(),
			CreatedByUserID:         uid,
			CreatedByUserName:       uname,
			CreatedFromIPAddress:    ipAddress,
			ModifiedAt:              time.Now(),
			ModifiedByUserID:        uid,
			ModifiedByUserName:      uname,
			ModifiedFromIPAddress:   ipAddress,
		}

		// Save to our database.
		if err := impl.AssistantMessageStorer.Create(sessCtx, am1); err != nil {
			impl.Logger.Error("failed creating question message", slog.Any("error", err))
			return nil, err
		}

		// This is our assistant response message. This message will remain
		// empty and then openAI will update the text response.
		am2 := &am_s.AssistantMessage{
			ID:                      primitive.NewObjectID(),
			TenantID:                tid,
			AssistantID:             at.AssistantID,
			AssistantThreadID:       at.ID,
			OpenAIAssistantID:       at.OpenAIAssistantID,
			OpenAIAssistantThreadID: at.OpenAIAssistantThreadID,
			Text:                    "",
			FromAssistant:           true,
			Status:                  at_s.AssistantThreadStatusQueued,
			CreatedAt:               time.Now(),
			CreatedByUserID:         uid,
			CreatedByUserName:       uname,
			CreatedFromIPAddress:    ipAddress,
			ModifiedAt:              time.Now(),
			ModifiedByUserID:        uid,
			ModifiedByUserName:      uname,
			ModifiedFromIPAddress:   ipAddress,
		}

		// Save to our database.
		if err := impl.AssistantMessageStorer.Create(sessCtx, am2); err != nil {
			impl.Logger.Error("failed creating assistant response message", slog.Any("error", err))
			return nil, err
		}

		// Submit the following into the background of this web-application.
		// This function will run independently of this function call.
		go func(lg *slog.Logger, amStorer am_s.AssistantMessageStorer, c *openai.Client, openAIAssistantID string, openAIAssistantThreadID string, text string, res *am_s.AssistantMessage) {
			if err := CreateOpenAIMessageInBackground(lg, amStorer, c, openAIAssistantID, openAIAssistantThreadID, text, res); err != nil {
				impl.Logger.Error("failed polling openai", slog.Any("error", err))
			}
		}(impl.Logger, impl.AssistantMessageStorer, client, at.OpenAIAssistantID, at.OpenAIAssistantThreadID, requestData.Text, am2)

		return am1, nil
	}

	// Start a transaction
	result, err := session.WithTransaction(ctx, transactionFunc)
	if err != nil {
		impl.Logger.Error("session failed error",
			slog.Any("error", err))
		return nil, err
	}

	return result.(*am_s.AssistantMessage), nil
}

func (impl *AssistantMessageControllerImpl) pollInBackground(client *openai.Client, run openai.Run, at *am_s.AssistantMessage, questionMessage string) error {
	// var err error
	// ctx := context.Background()
	// for run.Status != openai.RunStatusCompleted {
	//
	// 	// Sleep for 25 seconds.
	// 	time.Sleep(25 * time.Second)
	//
	// 	// retrieve the status of the run
	// 	run, err = client.RetrieveRun(ctx, at.OpenAIAssistantThreadID, run.ID)
	// 	if err != nil {
	// 		impl.Logger.Error("failed retrieving run",
	// 			slog.Any("error", err))
	// 		return err
	// 	}
	// }
	//
	// at.Status = assistantthread_s.AssistantThreadStatusActive
	// if err := impl.AssistantThreadStorer.UpdateByID(ctx, at); err != nil {
	// 	impl.Logger.Error("database create error", slog.Any("error", err))
	// 	return err
	// }
	//
	// impl.Logger.Debug("finished running assistant thread",
	// 	slog.Any("status", assistantthread_s.AssistantThreadStatusActive))
	//
	// impl.Logger.Debug("fetcheing list messages for assistant thread", slog.Any("assistant_thread_id", at.ID))
	//
	// msgs, err := client.ListMessage(context.Background(), at.OpenAIAssistantThreadID, nil, nil, nil, nil)
	// if err != nil {
	// 	impl.Logger.Error("failed listing message",
	// 		slog.Any("error", err))
	// 	return err
	// }
	// msg := msgs.Messages[0]
	//
	// atm := &assistantthread_s.AssistantThreadMessage{
	// 	ID:        primitive.NewObjectID(),
	// 	Question:  questionMessage,
	// 	Answer:    msg.Content[0].Text.Value,
	// 	CreatedAt: time.Now(),
	// }
	//
	// at.AssistantThreadMessages = append(at.AssistantThreadMessages, atm)
	// if err := impl.AssistantThreadStorer.UpdateByID(ctx, at); err != nil {
	// 	impl.Logger.Error("database create error", slog.Any("error", err))
	// 	return err
	// }
	return nil
}
