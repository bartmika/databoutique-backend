package controller

import (
	"context"
	"log/slog"
	"time"

	"github.com/sashabaranov/go-openai"

	am_s "github.com/bartmika/databoutique-backend/internal/app/assistantmessage/datastore"
)

// CreateOpenAIMessageInBackground function runs in background context to submit
// to OpenAI a `CreateMessage` API call and update our database with the latest
// response.
func CreateOpenAIMessageInBackground(
	logger *slog.Logger,
	amStorer am_s.AssistantMessageStorer,
	client *openai.Client,
	openAIAssistantID string,
	openAIThreadID string,
	message string,
	am *am_s.AssistantMessage,
) error {
	ctx := context.Background()
	var err error
	_, err = client.CreateMessage(ctx,
		openAIThreadID,
		openai.MessageRequest{
			Role:    string(openai.ThreadMessageRoleUser),
			Content: message,
		})
	if err != nil {
		logger.Error("failed created message from openai",
			slog.Any("error", err))
		return err
	}
	logger.Debug("submitted create message to openai")

	run, err := client.CreateRun(context.Background(), openAIThreadID, openai.RunRequest{
		AssistantID: openAIAssistantID,
	})
	if err != nil {
		logger.Error("failed executing run from openai",
			slog.Any("error", err))
		return err
	}
	logger.Debug("submitted create run to openai")

	// Continue to loop through the following code and polling openai every 25
	// seconds to see if the `CreateMessage` request has been executed for our
	// particular assistant.
	for run.Status != openai.RunStatusCompleted {
		time.Sleep(25 * time.Second) // Sleep for 25 seconds.

		// retrieve the status of the run
		run, err = client.RetrieveRun(ctx, openAIThreadID, run.ID)
		if err != nil {
			logger.Error("failed retrieving run from openai",
				slog.Any("error", err))
			return err
		}
	}
	logger.Debug("openai finished running")

	// The following code will fetch the latest messages for the particular
	// `thread_id` and return all the messages so far. Then extract most
	// recent message and save it into our system.

	msgs, err := client.ListMessage(context.Background(), openAIThreadID, nil, nil, nil, nil)
	if err != nil {
		logger.Error("failed listing message",
			slog.Any("error", err))
		return err
	}
	logger.Debug("fetched messages from openai")
	msg := msgs.Messages[0]

	// Update our record with the latest message.
	am.Status = am_s.AssistantMessageStatusActive
	am.Text = msg.Content[0].Text.Value
	am.ModifiedAt = time.Now()
	if err := amStorer.UpdateByID(ctx, am); err != nil {
		logger.Error("failed updating assistant message by id",
			slog.Any("error", err))
		return err
	}
	logger.Debug("updated message")

	return nil
}
