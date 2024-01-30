package controller

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	assistantthread_s "github.com/bartmika/databoutique-backend/internal/app/assistantthread/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

type AssistantThreadUpdateRequestIDO struct {
	ID          primitive.ObjectID `bson:"id" json:"id"`
	AssistantID primitive.ObjectID `bson:"assistant_id" json:"assistant_id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	Message     string             `bson:"message" json:"message"`
}

func ValidateUpdateRequest(dirtyData *AssistantThreadUpdateRequestIDO) error {
	e := make(map[string]string)

	if dirtyData.ID.IsZero() {
		e["id"] = "missing value"
	}
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

func (impl *AssistantThreadControllerImpl) UpdateByID(ctx context.Context, requestData *AssistantThreadUpdateRequestIDO) (*assistantthread_s.AssistantThread, error) {
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

		// Lookup the assistantthread in our database, else return a `400 Bad Request` error.
		ou, err := impl.AssistantThreadStorer.GetByID(sessCtx, requestData.ID)
		if err != nil {
			impl.Logger.Error("database error", slog.Any("err", err))
			return nil, err
		}
		if ou == nil {
			impl.Logger.Warn("assistantthread does not exist validation error")
			return nil, httperror.NewForBadRequestWithSingleField("id", "does not exist")
		}

		//
		// Update base.
		//

		ou.TenantID = tid
		// ou.Name = requestData.Name
		// ou.Description = requestData.Description
		// ou.Instructions = requestData.Instructions
		// ou.Model = requestData.Model
		ou.ModifiedAt = time.Now()
		ou.ModifiedByUserID = userID
		ou.ModifiedByUserName = userName
		ou.ModifiedFromIPAddress = ipAddress

		if err := impl.AssistantThreadStorer.UpdateByID(sessCtx, ou); err != nil {
			impl.Logger.Error("assistantthread update by id error", slog.Any("error", err))
			return nil, err
		}

		// ////
		// //// Update related
		// ////
		//
		// ou.AssistantThreadFiles = make([]*assistantthread_s.AssistantThreadFileOption, 0)
		// for _, assistantthreadFileID := range requestData.AssistantThreadFileIDs {
		// 	// Step 1: Lookup original.
		// 	assistantthreadFile, err := impl.AssistantThreadFileStorer.GetByID(sessCtx, assistantthreadFileID)
		// 	if err != nil {
		// 		impl.Logger.Error("fetching assistantthread file error", slog.Any("error", err))
		// 		return nil, err
		// 	}
		// 	if assistantthreadFile == nil {
		// 		impl.Logger.Error("assistantthread file does not exist error", slog.Any("assistantthreadFileID", assistantthreadFileID))
		// 		return nil, httperror.NewForBadRequestWithSingleField("assistantthread_file_ids", assistantthreadFileID.Hex()+" assistantthread file id does not exist")
		// 	}
		// 	af := &assistantthread_s.AssistantThreadFileOption{
		// 		ID:           assistantthreadFileID,
		// 		Name:         assistantthreadFile.Name,
		// 		Description:  assistantthreadFile.Description,
		// 		OpenAIFileID: assistantthreadFile.OpenAIFileID,
		// 		Status:       assistantthreadFile.Status,
		// 	}
		// 	ou.AssistantThreadFiles = append(ou.AssistantThreadFiles, af)
		// }
		//
		// ////
		// //// Submit to OpenAI.
		// ////
		//
		// creds, err := impl.TenantStorer.GetOpenAICredentialsByID(sessCtx, tid)
		// if err != nil {
		// 	impl.Logger.Error("failed to get OpenAI credentials",
		// 		slog.String("tenant_id", tid.Hex()),
		// 		slog.Any("error", err))
		// 	return nil, err
		// }
		// if creds == nil {
		// 	return nil, errors.New("no openai credentials returned")
		// }
		//
		// impl.Logger.Debug("openai initializing...")
		// client := openai.NewOrgClient(creds.APIKey, creds.OrgKey)
		// impl.Logger.Debug("openai initialized")
		//
		// assistantthread, err := client.RetrieveAssistantThread(sessCtx, ou.OpenAIAssistantThreadID)
		// if err != nil {
		// 	impl.Logger.Error("failed to get OpenAI credentials",
		// 		slog.String("tenant_id", tid.Hex()),
		// 		slog.Any("error", err))
		// 	return nil, err
		// }
		// if isStructEmpty(assistantthread) {
		// 	impl.Logger.Error("no openai assistantthread returned")
		// 	return "", errors.New("no openai assistantthread returned")
		// }
		//
		// // Get all the files which we will pre-train the LLM.
		// afIDs := ou.GetAssistantThreadFileIDs()
		//
		// // Update the existing fields in OpenAI.
		// assistantthread.Name = &requestData.Name
		// assistantthread.Model = requestData.Model
		// assistantthread.Instructions = &requestData.Instructions
		// assistantthread.FileIDs = afIDs
		// modReq := openai.AssistantThreadRequest{
		// 	Model:        ou.Model,
		// 	Name:         &ou.Name,
		// 	Description:  &ou.Description,
		// 	Instructions: &ou.Instructions,
		// 	Tools:        []openai.AssistantThreadTool{{Type: openai.AssistantThreadToolTypeRetrieval}},
		// 	FileIDs:      afIDs,
		// 	// Metadata
		// }
		// if _, err := client.ModifyAssistantThread(sessCtx, assistantthread.ID, modReq); err != nil {
		// 	impl.Logger.Error("failed modify assistantthread",
		// 		slog.String("tenant_id", tid.Hex()),
		// 		slog.Any("error", err))
		// 	return nil, err
		// }

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

	return result.(*assistantthread_s.AssistantThread), nil
}
