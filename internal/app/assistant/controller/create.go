package controller

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	assistant_s "github.com/bartmika/databoutique-backend/internal/app/assistant/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
	"github.com/sashabaranov/go-openai"
)

type AssistantCreateRequestIDO struct {
	Name             string               `bson:"name" json:"name"`
	Description      string               `bson:"description" json:"description"`
	Instructions     string               `bson:"instructions" json:"instructions"`
	Model            string               `bson:"model" json:"model"`
	AssistantFileIDs []primitive.ObjectID `bson:"assistant_file_ids" json:"assistant_file_ids"`
}

func ValidateCreateRequest(dirtyData *AssistantCreateRequestIDO) error {
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

	if len(e) != 0 {
		return httperror.NewForBadRequest(&e)
	}
	return nil
}

func (impl *AssistantControllerImpl) Create(ctx context.Context, requestData *AssistantCreateRequestIDO) (*assistant_s.Assistant, error) {

	//
	// Get variables from our user authenticated session.
	//

	tid, _ := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	// role, _ := ctx.Value(constants.SessionUserRole).(int8)
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
	impl.Kmutex.Lockf("create-assistant-by-tenant-%s", tid.Hex())
	defer impl.Kmutex.Unlockf("create-assistant-by-tenant-%s", tid.Hex())

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
		m := &assistant_s.Assistant{
			Name:         requestData.Name,
			Description:  requestData.Description,
			Instructions: requestData.Instructions,
			Model:        requestData.Model,
			Status:       assistant_s.AssistantStatusActive,
		}

		// Add defaults.
		m.TenantID = tid
		m.ID = primitive.NewObjectID()
		m.CreatedAt = time.Now()
		m.CreatedByUserID = userID
		m.CreatedByUserName = userName
		m.CreatedFromIPAddress = ipAddress
		m.ModifiedAt = time.Now()
		m.ModifiedByUserID = userID
		m.ModifiedByUserName = userName
		m.ModifiedFromIPAddress = ipAddress

		// Save to our database.
		if err := impl.AssistantStorer.Create(sessCtx, m); err != nil {
			impl.Logger.Error("database create error", slog.Any("error", err))
			return nil, err
		}

		////
		//// Update related
		////

		m.AssistantFiles = make([]*assistant_s.AssistantFileOption, 0)
		for _, assistantFileID := range requestData.AssistantFileIDs {
			// Step 1: Lookup original.
			assistantFile, err := impl.AssistantFileStorer.GetByID(sessCtx, assistantFileID)
			if err != nil {
				impl.Logger.Error("fetching assistant file error", slog.Any("error", err))
				return nil, err
			}
			if assistantFile == nil {
				impl.Logger.Error("assistant file does not exist error", slog.Any("assistantFileID", assistantFileID))
				return nil, httperror.NewForBadRequestWithSingleField("assistant_file_ids", assistantFileID.Hex()+" assistant file id does not exist")
			}
			af := &assistant_s.AssistantFileOption{
				ID:           assistantFileID,
				Name:         assistantFile.Name,
				Description:  assistantFile.Description,
				OpenAIFileID: assistantFile.OpenAIFileID,
				Status:       assistantFile.Status,
			}
			m.AssistantFiles = append(m.AssistantFiles, af)
		}

		// Save to our database.
		if err := impl.AssistantStorer.UpdateByID(sessCtx, m); err != nil {
			impl.Logger.Error("database update error", slog.Any("error", err))
			return nil, err
		}

		////
		//// Submit to OpenAI.
		////

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

		impl.Logger.Debug("openai initializing...")
		client := openai.NewOrgClient(creds.APIKey, creds.OrgKey)
		impl.Logger.Debug("openai initialized")

		// Get all the files which we will pre-train the LLM.
		afIDs := m.GetAssistantFileIDs()

		impl.Logger.Debug("beginning to create assistant...",
			slog.String("tenant_id", tid.Hex()),
			slog.Any("name", m.Name),
			slog.Any("model", m.Model),
			slog.Any("instructions", m.Instructions),
			slog.Any("tools", []openai.AssistantTool{{Type: openai.AssistantToolTypeRetrieval}}),
			slog.Any("file_ids", afIDs),
			slog.String("tenant_id", tid.Hex()))

		// Create an assistant.
		assistant, err := client.CreateAssistant(context.Background(), openai.AssistantRequest{
			Name:         &m.Name,
			Model:        m.Model,
			Instructions: &m.Instructions,
			Tools:        []openai.AssistantTool{{Type: openai.AssistantToolTypeRetrieval}},
			FileIDs:      afIDs,
		})
		if err != nil {
			impl.Logger.Error("failed creating assistant",
				slog.String("tenant_id", tid.Hex()),
				slog.Any("name", assistant.Name),
				slog.Any("model", assistant.Model),
				slog.Any("instructions", assistant.Instructions),
				slog.Any("tools", assistant.Tools),
				slog.Any("file_ids", assistant.FileIDs),
				slog.Any("error", err))
			return nil, err
		}
		if isStructEmpty(assistant) {
			impl.Logger.Error("no openai assistant returned", slog.Any("assistant", assistant))
			return "", errors.New("no openai file returned")
		}

		impl.Logger.Debug("finished creating assistant",
			slog.String("tenant_id", tid.Hex()),
			slog.Any("assistant_id", assistant.ID))

		m.OpenAIAssistantID = assistant.ID
		if err := impl.AssistantStorer.UpdateByID(sessCtx, m); err != nil {
			impl.Logger.Error("database update error", slog.Any("error", err))
			return nil, err
		}

		////
		//// Exit our transaction successfully.
		////

		return m, nil
	}

	// Start a transaction
	result, err := session.WithTransaction(ctx, transactionFunc)
	if err != nil {
		impl.Logger.Error("session failed error",
			slog.Any("error", err))
		return nil, err
	}

	return result.(*assistant_s.Assistant), nil
}

func isStructEmpty(s interface{}) bool {
	val := reflect.ValueOf(s)
	zeroVal := reflect.Zero(val.Type())
	return reflect.DeepEqual(val.Interface(), zeroVal.Interface())
}
