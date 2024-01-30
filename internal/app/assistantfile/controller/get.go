package controller

import (
	"context"
	"log/slog"
	"time"

	domain "github.com/bartmika/databoutique-backend/internal/app/assistantfile/datastore"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (c *AssistantFileControllerImpl) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.AssistantFile, error) {
	// // Extract from our session the following data.
	// userAssistantFileID := ctx.Value(constants.SessionUserAssistantFileID).(primitive.ObjectID)
	// userRole := ctx.Value(constants.SessionUserRole).(int8)
	//
	// If user is not administrator nor belongs to the assistantfile then error.
	// if userRole != user_d.UserRoleRoot && id != userAssistantFileID {
	// 	c.Logger.Error("authenticated user is not staff role nor belongs to the assistantfile error",
	// 		slog.Any("userRole", userRole),
	// 		slog.Any("userAssistantFileID", userAssistantFileID))
	// 	return nil, httperror.NewForForbiddenWithSingleField("message", "you do not belong to this assistantfile")
	// }

	// Retrieve from our database the record for the specific id.
	m, err := c.AssistantFileStorer.GetByID(ctx, id)
	if err != nil {
		c.Logger.Error("database get by id error", slog.Any("error", err))
		return nil, err
	}

	// Generate the URL.
	fileURL, err := c.S3.GetPresignedURL(ctx, m.ObjectKey, 5*time.Minute)
	if err != nil {
		c.Logger.Error("s3 failed get presigned url error", slog.Any("error", err))
		return nil, err
	}

	m.ObjectURL = fileURL
	return m, err
}
