package controller

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/sashabaranov/go-openai"
	"go.mongodb.org/mongo-driver/mongo"

	program_s "github.com/bartmika/databoutique-backend/internal/app/program/datastore"
)

func (impl *ProgramControllerImpl) createProgramInBackgroundForOpenAI(prog *program_s.Program) error {
	ctx := context.Background()

	// Lock this program until OpenAI finishes executing.
	impl.Kmutex.Lockf("openai_program_%s", prog.ID.Hex())
	defer impl.Kmutex.Unlockf("openai_program_%s", prog.ID.Hex())

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

		////
		//// Get related records & connect to OpenAI.
		////

		creds, err := impl.TenantStorer.GetOpenAICredentialsByID(sessCtx, prog.TenantID)
		if err != nil {
			impl.Logger.Error("failed getting openai credentials",
				slog.Any("program_id", prog.ID),
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
		fileIDs := prog.GetOpenAIFileIDs()

		////
		//// Create assistant.
		////

		impl.Logger.Debug("beginning to create assistant...",
			slog.Any("program_id", prog.ID))

		pname := fmt.Sprintf("program_%s", prog.ID.Hex())
		assistant, err := client.CreateAssistant(context.Background(), openai.AssistantRequest{
			Name:         &pname,
			Model:        prog.Model,
			Instructions: &prog.Instructions,
			Tools:        []openai.AssistantTool{{Type: openai.AssistantToolTypeRetrieval}},
			FileIDs:      fileIDs,
		})
		if err != nil {
			impl.Logger.Error("failed creating assistant",
				slog.Any("program_id", prog.ID),
				slog.Any("name", assistant.Name),
				slog.Any("model", assistant.Model),
				slog.Any("instructions", assistant.Instructions),
				slog.Any("tools", assistant.Tools),
				slog.Any("file_ids", assistant.FileIDs),
				slog.Any("error", err))
			return nil, err
		}
		if isStructEmpty(assistant) {
			impl.Logger.Error("no openai assistant returned",
				slog.Any("program_id", prog.ID),
				slog.Any("assistant", assistant))
			return "", errors.New("no openai file returned")
		}

		prog.OpenAIAssistantID = assistant.ID

		impl.Logger.Debug("finished creating assistant",
			slog.Any("program_id", prog.ID),
			slog.Any("assistant_id", assistant.ID))

		// Set the status that this executive is active.
		prog.Status = program_s.ProgramStatusActive

		// Update the program.
		if err := impl.ProgramStorer.UpdateByID(sessCtx, prog); err != nil {
			impl.Logger.Error("failed updating program",
				slog.Any("program_id", prog.ID),
				slog.Any("error", err))
			return nil, err
		}

		////
		//// Exit our transaction successfully.
		////

		return prog, nil
	}

	// Start a transaction
	if _, err := session.WithTransaction(ctx, transactionFunc); err != nil {
		impl.Logger.Error("session failed error",
			slog.Any("error", err))
		return err
	}

	return nil
}

func isStructEmpty(s interface{}) bool {
	val := reflect.ValueOf(s)
	zeroVal := reflect.Zero(val.Type())
	return reflect.DeepEqual(val.Interface(), zeroVal.Interface())
}
