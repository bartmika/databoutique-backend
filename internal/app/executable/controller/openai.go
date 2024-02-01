package controller

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	executable_s "github.com/bartmika/databoutique-backend/internal/app/executable/datastore"
	"github.com/sashabaranov/go-openai"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (impl *ExecutableControllerImpl) createExecutableInBackgroundForOpenAI(exec *executable_s.Executable) error {
	ctx := context.Background()

	// Lock this executable until OpenAI finishes executing.
	impl.Kmutex.Lockf("openai_executable_%s", exec.ID.Hex())
	defer impl.Kmutex.Unlockf("openai_executable_%s", exec.ID.Hex())

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

		p, err := impl.ProgramStorer.GetByID(sessCtx, exec.ProgramID)
		if err != nil {
			impl.Logger.Error("failed getting program",
				slog.Any("executable_id", exec.ID),
				slog.Any("error", err))
			return nil, err
		}
		if p == nil {
			err := fmt.Errorf("program does not exist for id: %v", exec.ProgramID.Hex())
			impl.Logger.Error("program does not exist", slog.Any("error", err))
			return nil, err
		}

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

		// Get all the files which we will pre-train the LLM.
		fileIDs := exec.GetOpenAIFileIDs()

		////
		//// Create assistant.
		////

		impl.Logger.Debug("beginning to create assistant...",
			slog.Any("executable_id", exec.ID))

		aname := fmt.Sprintf("program_%s_executable_%s", exec.ProgramID.Hex(), exec.ID.Hex())
		assistant, err := client.CreateAssistant(context.Background(), openai.AssistantRequest{
			Name:         &aname,
			Model:        p.Model,
			Instructions: &p.Instructions,
			Tools:        []openai.AssistantTool{{Type: openai.AssistantToolTypeRetrieval}},
			FileIDs:      fileIDs,
		})
		if err != nil {
			impl.Logger.Error("failed creating assistant",
				slog.Any("executable_id", exec.ID),
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
				slog.Any("executable_id", exec.ID),
				slog.Any("assistant", assistant))
			return "", errors.New("no openai file returned")
		}

		exec.OpenAIAssistantID = assistant.ID

		impl.Logger.Debug("finished creating assistant",
			slog.Any("executable_id", exec.ID),
			slog.Any("assistant_id", assistant.ID))

		////
		//// Create assistant thread.
		////

		// Create a thread in OpenAI.
		thread, err := client.CreateThread(sessCtx, openai.ThreadRequest{})
		if err != nil {
			impl.Logger.Error("failed creating assistant thread",
				slog.Any("executable_id", exec.ID),
				slog.Any("error", err))
			return nil, err
		}
		if isStructEmpty(thread) {
			impl.Logger.Error("no openai assistant thread returned",
				slog.Any("executable_id", exec.ID),
				slog.Any("assistant_thread", thread))
			return "", errors.New("no openai assistant thread returned")
		}

		exec.OpenAIAssistantThreadID = thread.ID

		impl.Logger.Debug("create openai thread",
			slog.String("thread_id", thread.ID),
			slog.Any("executable_id", exec.ID))

		////
		//// Create assistant message.
		////

		// --- Create message --- //

		_, err = client.CreateMessage(ctx,
			exec.OpenAIAssistantThreadID,
			openai.MessageRequest{
				Role:    string(openai.ThreadMessageRoleUser),
				Content: exec.Question,
			})
		if err != nil {
			impl.Logger.Error("failed created message from openai",
				slog.String("thread_id", thread.ID),
				slog.Any("error", err))
			return nil, err
		}
		impl.Logger.Debug("submitted create message to openai")

		// --- Run message creation --- //

		run, err := client.CreateRun(context.Background(), exec.OpenAIAssistantThreadID, openai.RunRequest{
			AssistantID: exec.OpenAIAssistantID,
		})
		if err != nil {
			impl.Logger.Error("failed executing run from openai",
				slog.Any("error", err))
			return nil, err
		}
		impl.Logger.Debug("openai running message processing...")

		// --- Poll in foreground for completion by openai --- //

		// Continue to loop through the following code and polling openai every 25
		// seconds to see if the `CreateMessage` request has been executed for our
		// particular assistant.
		for run.Status != openai.RunStatusCompleted {
			time.Sleep(25 * time.Second) // Sleep for 25 seconds.

			// retrieve the status of the run
			run, err = client.RetrieveRun(ctx, exec.OpenAIAssistantThreadID, run.ID)
			if err != nil {
				impl.Logger.Error("failed retrieving run from openai",
					slog.Any("error", err))
				return nil, err
			}
		}
		impl.Logger.Debug("openai finished running for message processing")

		// --- Get message list --- //

		// The following code will fetch the latest messages for the particular
		// `thread_id` and return all the messages so far. Then extract most
		// recent message and save it into our system.
		impl.Logger.Debug("fetching recent messages from openai...")

		msgs, err := client.ListMessage(context.Background(), exec.OpenAIAssistantThreadID, nil, nil, nil, nil)
		if err != nil {
			impl.Logger.Error("failed listing message",
				slog.Any("error", err))
			return nil, err
		}
		impl.Logger.Debug("received recent messages from openai")
		msg := msgs.Messages[0]

		// Create the message.
		message := &executable_s.Message{}
		message.ID = primitive.NewObjectID()
		message.OpenAIMessageID = msg.ID
		message.Content = msg.Content[0].Text.Value
		message.CreatedAt = time.Now()
		message.Status = executable_s.ExecutableStatusActive
		message.FromExecutable = true

		// Save the message to the executable.
		exec.Messages = append(exec.Messages, message)

		////
		//// Update database record.
		////

		// Set the status that this executive is active.
		exec.Status = executable_s.ExecutableStatusActive

		// Update the executable.
		if err := impl.ExecutableStorer.UpdateByID(sessCtx, exec); err != nil {
			impl.Logger.Error("failed updating executable",
				slog.Any("executable_id", exec.ID),
				slog.Any("error", err))
			return nil, err
		}

		impl.Logger.Debug("updated executable with latest message")

		////
		//// Exit our transaction successfully.
		////

		return exec, nil
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

func (impl *ExecutableControllerImpl) processQuestionSubmissionInBackgroundForOpenAI(exec *executable_s.Executable) error {
	ctx := context.Background()

	// Lock this executable until OpenAI finishes executing.
	impl.Kmutex.Lockf("openai_executable_%s", exec.ID.Hex())
	defer impl.Kmutex.Unlockf("openai_executable_%s", exec.ID.Hex())

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

		p, err := impl.ProgramStorer.GetByID(sessCtx, exec.ProgramID)
		if err != nil {
			impl.Logger.Error("failed getting program",
				slog.Any("executable_id", exec.ID),
				slog.Any("error", err))
			return nil, err
		}
		if p == nil {
			err := fmt.Errorf("program does not exist for id: %v", exec.ProgramID.Hex())
			impl.Logger.Error("program does not exist", slog.Any("error", err))
			return nil, err
		}

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

		////
		//// Create assistant message.
		////
		// --- Find the pending question --- //

		var pendingMessage *executable_s.Message
		for _, message := range exec.Messages {
			if message.Status == executable_s.ExecutableStatusProcessing {
				pendingMessage = message
				break
			}
		}

		// Defensive code.
		if pendingMessage == nil {
			err := fmt.Errorf("could not find pending message in executable ID: %v", exec.ID.Hex())
			impl.Logger.Error("no pending messages",
				slog.Any("error", err))
			return nil, err
		}

		// --- Create message --- //

		_, err = client.CreateMessage(ctx,
			exec.OpenAIAssistantThreadID,
			openai.MessageRequest{
				Role:    string(openai.ThreadMessageRoleUser),
				Content: pendingMessage.Content,
			})
		if err != nil {
			impl.Logger.Error("failed created message from openai",
				slog.String("thread_id", exec.OpenAIAssistantThreadID),
				slog.Any("error", err))
			return nil, err
		}
		impl.Logger.Debug("submitted create message to openai")

		// --- Run message creation --- //

		run, err := client.CreateRun(context.Background(), exec.OpenAIAssistantThreadID, openai.RunRequest{
			AssistantID: exec.OpenAIAssistantID,
		})
		if err != nil {
			impl.Logger.Error("failed executing run from openai",
				slog.Any("error", err))
			return nil, err
		}
		impl.Logger.Debug("openai running message processing...")

		// --- Poll in foreground for completion by openai --- //

		// Continue to loop through the following code and polling openai every 25
		// seconds to see if the `CreateMessage` request has been executed for our
		// particular assistant.
		for run.Status != openai.RunStatusCompleted {
			time.Sleep(25 * time.Second) // Sleep for 25 seconds.

			// retrieve the status of the run
			run, err = client.RetrieveRun(ctx, exec.OpenAIAssistantThreadID, run.ID)
			if err != nil {
				impl.Logger.Error("failed retrieving run from openai",
					slog.Any("error", err))
				return nil, err
			}
		}
		impl.Logger.Debug("openai finished running for message processing")

		// --- Get message list --- //

		// The following code will fetch the latest messages for the particular
		// `thread_id` and return all the messages so far. Then extract most
		// recent message and save it into our system.
		impl.Logger.Debug("fetching recent messages from openai...")

		msgs, err := client.ListMessage(context.Background(), exec.OpenAIAssistantThreadID, nil, nil, nil, nil)
		if err != nil {
			impl.Logger.Error("failed listing message",
				slog.Any("error", err))
			return nil, err
		}
		impl.Logger.Debug("received recent messages from openai")
		msg := msgs.Messages[0]

		// Populate the pending message contents from OpenAI.
		pendingMessage.OpenAIMessageID = msg.ID
		pendingMessage.Content = msg.Content[0].Text.Value
		pendingMessage.CreatedAt = time.Now()
		pendingMessage.Status = executable_s.ExecutableStatusActive
		pendingMessage.FromExecutable = true

		////
		//// Update database record.
		////

		// Set the status that this executive is active.
		exec.Status = executable_s.ExecutableStatusActive

		// Update the executable.
		if err := impl.ExecutableStorer.UpdateByID(sessCtx, exec); err != nil {
			impl.Logger.Error("failed updating executable",
				slog.Any("executable_id", exec.ID),
				slog.Any("error", err))
			return nil, err
		}

		impl.Logger.Debug("updated executable with latest message")

		////
		//// Exit our transaction successfully.
		////

		return exec, nil
	}

	// Start a transaction
	if _, err := session.WithTransaction(ctx, transactionFunc); err != nil {
		impl.Logger.Error("session failed error",
			slog.Any("error", err))
		return err
	}

	return nil
}
