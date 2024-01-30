package controller

import (
	"context"
	"errors"
	"log/slog"

	"github.com/sashabaranov/go-openai"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/bartmika/databoutique-backend/internal/config/constants"
)

func (impl *AssistantControllerImpl) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	// Extract from our session the following data.
	tenantID, _ := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)

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

		// STEP 1: Lookup the record or error.
		assistant, err := impl.GetByID(sessCtx, id)
		if err != nil {
			impl.Logger.Error("database get by id error", slog.Any("error", err))
			return nil, err
		}
		if assistant == nil {
			impl.Logger.Error("database returns nothing from get by id")
			return nil, err
		}

		// STEP 2: Delete from database.
		if err := impl.AssistantStorer.DeleteByID(sessCtx, id); err != nil {
			impl.Logger.Error("database delete by id error", slog.Any("error", err))
			return nil, err
		}

		// STEP 3: Delete from OpenAI.
		creds, err := impl.TenantStorer.GetOpenAICredentialsByID(sessCtx, tenantID)
		if err != nil {
			impl.Logger.Error("failed file upload to openai",
				slog.String("tenant_id", tenantID.Hex()),
				slog.Any("error", err))
			return nil, err
		}
		if creds == nil {
			return nil, errors.New("no openai credentials returned")
		}
		if err := impl.deleteOpanAI(sessCtx, assistant.OpenAIAssistantID, creds.APIKey, creds.OrgKey); err != nil {
			impl.Logger.Error("failed deleting assitant from openai", slog.Any("error", err))
			return nil, err
		}

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

func (impl *AssistantControllerImpl) deleteOpanAI(ctx context.Context, assitantID string, apikey string, orgKey string) error {
	impl.Logger.Debug("openai initializing...")
	client := openai.NewOrgClient(apikey, orgKey)
	impl.Logger.Debug("openai initialized")
	if _, err := client.DeleteAssistant(ctx, assitantID); err != nil {
		impl.Logger.Error("failed deleting open ai assitant", slog.Any("error", err))
		return err
	}
	impl.Logger.Debug("deleted openai assitant from assistant api", slog.Any("assistant_id", assitantID))
	return nil
}
