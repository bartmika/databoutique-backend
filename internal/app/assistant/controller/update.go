package controller

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	assistant_s "github.com/bartmika/databoutique-backend/internal/app/assistant/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
	"github.com/sashabaranov/go-openai"
)

type AssistantUpdateRequestIDO struct {
	ID               primitive.ObjectID   `bson:"id" json:"id"`
	Name             string               `bson:"name" json:"name"`
	Instructions     string               `bson:"instructions" json:"instructions"`
	Model            string               `bson:"model" json:"model"`
	Description      string               `bson:"description" json:"description"`
	AssistantFileIDs []primitive.ObjectID `bson:"assistant_file_ids" json:"assistant_file_ids"`
}

func ValidateUpdateRequest(dirtyData *AssistantUpdateRequestIDO) error {
	e := make(map[string]string)

	if dirtyData.ID.IsZero() {
		e["id"] = "missing value"
	}
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

func (impl *AssistantControllerImpl) UpdateByID(ctx context.Context, requestData *AssistantUpdateRequestIDO) (*assistant_s.Assistant, error) {
	//
	// Perform our validation and return validation error on any issues detected.
	//

	if err := ValidateUpdateRequest(requestData); err != nil {
		return nil, err
	}

	//
	// Get variables from our user authenticated session.
	//

	tid, _ := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	// role, _ := ctx.Value(constants.SessionUserRole).(int8)
	userID, _ := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	userName, _ := ctx.Value(constants.SessionUserName).(string)
	ipAddress, _ := ctx.Value(constants.SessionIPAddress).(string)

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

		// Lookup the assistant in our database, else return a `400 Bad Request` error.
		ou, err := impl.AssistantStorer.GetByID(sessCtx, requestData.ID)
		if err != nil {
			impl.Logger.Error("database error", slog.Any("err", err))
			return nil, err
		}
		if ou == nil {
			impl.Logger.Warn("assistant does not exist validation error")
			return nil, httperror.NewForBadRequestWithSingleField("id", "does not exist")
		}

		//
		// Update base.
		//

		ou.TenantID = tid
		ou.Name = requestData.Name
		ou.Description = requestData.Description
		ou.Instructions = requestData.Instructions
		ou.Model = requestData.Model
		ou.ModifiedAt = time.Now()
		ou.ModifiedByUserID = userID
		ou.ModifiedByUserName = userName
		ou.ModifiedFromIPAddress = ipAddress

		if err := impl.AssistantStorer.UpdateByID(sessCtx, ou); err != nil {
			impl.Logger.Error("assistant update by id error", slog.Any("error", err))
			return nil, err
		}

		////
		//// Update related
		////

		ou.AssistantFiles = make([]*assistant_s.AssistantFileOption, 0)
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
			ou.AssistantFiles = append(ou.AssistantFiles, af)
		}

		////
		//// Submit to OpenAI.
		////

		creds, err := impl.TenantStorer.GetOpenAICredentialsByID(sessCtx, tid)
		if err != nil {
			impl.Logger.Error("failed to get OpenAI credentials",
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

		assistant, err := client.RetrieveAssistant(sessCtx, ou.OpenAIAssistantID)
		if err != nil {
			impl.Logger.Error("failed to get OpenAI credentials",
				slog.String("tenant_id", tid.Hex()),
				slog.Any("error", err))
			return nil, err
		}
		if isStructEmpty(assistant) {
			impl.Logger.Error("no openai assistant returned")
			return "", errors.New("no openai assistant returned")
		}

		// Get all the files which we will pre-train the LLM.
		afIDs := ou.GetAssistantFileIDs()

		// Update the existing fields in OpenAI.
		assistant.Name = &requestData.Name
		assistant.Model = requestData.Model
		assistant.Instructions = &requestData.Instructions
		assistant.FileIDs = afIDs
		modReq := openai.AssistantRequest{
			Model:        ou.Model,
			Name:         &ou.Name,
			Description:  &ou.Description,
			Instructions: &ou.Instructions,
			Tools:        []openai.AssistantTool{{Type: openai.AssistantToolTypeRetrieval}},
			FileIDs:      afIDs,
			// Metadata
		}
		if _, err := client.ModifyAssistant(sessCtx, assistant.ID, modReq); err != nil {
			impl.Logger.Error("failed modify assistant",
				slog.String("tenant_id", tid.Hex()),
				slog.Any("error", err))
			return nil, err
		}

		////
		//// Exit our transaction successfully.
		////

		return ou, nil
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
