package controller

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	"github.com/sashabaranov/go-openai"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	am_c "github.com/bartmika/databoutique-backend/internal/app/assistantmessage/controller"
	am_s "github.com/bartmika/databoutique-backend/internal/app/assistantmessage/datastore"
	at_s "github.com/bartmika/databoutique-backend/internal/app/assistantthread/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

type AssistantThreadCreateRequestIDO struct {
	AssistantID primitive.ObjectID `bson:"assistant_id" json:"assistant_id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	Message     string             `bson:"message" json:"message"`
}

func ValidateCreateRequest(dirtyData *AssistantThreadCreateRequestIDO) error {
	e := make(map[string]string)

	if dirtyData.AssistantID.IsZero() {
		e["assistant_id"] = "missing value"
	}
	if dirtyData.UserID.IsZero() {
		e["user_id"] = "missing value"
	}
	if dirtyData.Message == "" {
		e["message"] = "missing value"
	}
	if len(e) != 0 {
		return httperror.NewForBadRequest(&e)
	}
	return nil
}

func (impl *AssistantThreadControllerImpl) Create(ctx context.Context, requestData *AssistantThreadCreateRequestIDO) (*at_s.AssistantThread, error) {

	//
	// Get variables from our user authenticated session.
	//

	tid, _ := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	tname, _ := ctx.Value(constants.SessionUserTenantName).(string)
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
	impl.Kmutex.Lockf("create-assistant-thread-by-tenant-%s", tid.Hex())
	defer impl.Kmutex.Unlockf("create-assistant-thread-by-tenant-%s", tid.Hex())

	//
	// Perform our validation and return validation error on any issues detected.
	//

	if err := ValidateCreateRequest(requestData); err != nil {
		return nil, err
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

		// Fetch our related records: assistant and user, if either do not
		// exist then we will error.

		a, err := impl.AssistantStorer.GetByID(sessCtx, requestData.AssistantID)
		if err != nil {
			impl.Logger.Error("failed getting assistant by id",
				slog.Any("error", err))
			return nil, err
		}
		if a == nil {
			err := fmt.Errorf("no assistant found with id: %s", requestData.AssistantID.Hex())
			impl.Logger.Error("", slog.Any("error", err))
			return nil, err
		}

		u, err := impl.UserStorer.GetByID(sessCtx, requestData.UserID)
		if err != nil {
			impl.Logger.Error("failed getting user by id",
				slog.Any("error", err))
			return nil, err
		}
		if u == nil {
			err := fmt.Errorf("no user found with id: %s", requestData.UserID.Hex())
			impl.Logger.Error("", slog.Any("error", err))
			return nil, err
		}

		// Create our thread for this particular user in our web-app.

		at := &at_s.AssistantThread{
			AssistantID:             a.ID,
			AssistantName:           a.Name,
			AssistantDescription:    a.Description,
			OpenAIAssistantID:       a.OpenAIAssistantID,
			Status:                  at_s.AssistantThreadStatusActive,
			OpenAIAssistantThreadID: "-",
			ID:                      primitive.NewObjectID(),
			TenantID:                tid,
			TenantName:              tname,
			UserID:                  u.ID,
			UserPublicID:            u.PublicID,
			UserFirstName:           u.FirstName,
			UserLastName:            u.LastName,
			UserName:                u.Name,
			UserLexicalName:         u.LexicalName,
			CreatedAt:               time.Now(),
			CreatedByUserID:         u.ID,
			CreatedFromIPAddress:    ipAddress,
			ModifiedAt:              time.Now(),
			ModifiedByUserID:        u.ID,
			ModifiedByUserName:      u.Name,
			ModifiedFromIPAddress:   ipAddress,
		}

		// Save to our database.
		if err := impl.AssistantThreadStorer.Create(sessCtx, at); err != nil {
			impl.Logger.Error("database create error", slog.Any("error", err))
			return nil, err
		}

		// The following lines of code will request openai to create us a
		// assistant thread. Once openai returns a `thread_id` for us to
		// use, we will save it in our app for further use. To begin all of
		// this, start by getting our tenant's credentials.

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

		// Create a thread in OpenAI.
		thread, err := client.CreateThread(sessCtx, openai.ThreadRequest{})
		if err != nil {
			impl.Logger.Error("failed creating assistantthread",
				slog.String("tenant_id", tid.Hex()),
				slog.Any("error", err))
			return nil, err
		}
		if isStructEmpty(thread) {
			impl.Logger.Error("no openai assistant thread returned", slog.Any("assistant_thread", thread))
			return "", errors.New("no openai assistant thread returned")
		}

		impl.Logger.Debug("create openai thread",
			slog.String("thread_id", thread.ID),
			slog.String("tenant_id", tid.Hex()))

		at.OpenAIAssistantThreadID = thread.ID
		if err := impl.AssistantThreadStorer.UpdateByID(sessCtx, at); err != nil {
			impl.Logger.Error("database update error", slog.Any("error", err))
			return nil, err
		}

		// The next lines of code will involve two steps:
		// Step 1: Create our first initial message of what the user asked.
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
			Text:                    requestData.Message,
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
			impl.Logger.Error("failed creating initial message", slog.Any("error", err))
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
			if err := am_c.CreateOpenAIMessageInBackground(lg, amStorer, c, openAIAssistantID, openAIAssistantThreadID, text, res); err != nil {
				impl.Logger.Error("failed polling openai", slog.Any("error", err))
			}
		}(impl.Logger, impl.AssistantMessageStorer, client, at.OpenAIAssistantID, at.OpenAIAssistantThreadID, requestData.Message, am2)

		// ////
		// //// Exit our transaction successfully.
		// ////

		return at, nil
	}

	// Start a transaction
	result, err := session.WithTransaction(ctx, transactionFunc)
	if err != nil {
		impl.Logger.Error("session failed error",
			slog.Any("error", err))
		return nil, err
	}

	return result.(*at_s.AssistantThread), nil
}

func isStructEmpty(s interface{}) bool {
	val := reflect.ValueOf(s)
	zeroVal := reflect.Zero(val.Type())
	return reflect.DeepEqual(val.Interface(), zeroVal.Interface())
}
