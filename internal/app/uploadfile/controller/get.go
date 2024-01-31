package controller

import (
	"context"
	"log/slog"
	"time"

	domain "github.com/bartmika/databoutique-backend/internal/app/uploadfile/datastore"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (c *UploadFileControllerImpl) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.UploadFile, error) {
	// // Extract from our session the following data.
	// userUploadFileID := ctx.Value(constants.SessionUserUploadFileID).(primitive.ObjectID)
	// userRole := ctx.Value(constants.SessionUserRole).(int8)
	//
	// If user is not administrator nor belongs to the uploadfile then error.
	// if userRole != user_d.UserRoleRoot && id != userUploadFileID {
	// 	c.Logger.Error("authenticated user is not staff role nor belongs to the uploadfile error",
	// 		slog.Any("userRole", userRole),
	// 		slog.Any("userUploadFileID", userUploadFileID))
	// 	return nil, httperror.NewForForbiddenWithSingleField("message", "you do not belong to this uploadfile")
	// }

	// Retrieve from our database the record for the specific id.
	m, err := c.UploadFileStorer.GetByID(ctx, id)
	if err != nil {
		c.Logger.Error("database get by id error", slog.Any("error", err))
		return nil, err
	}

	// Generate the URL if it exists.
	if m.ObjectKey != "" {
		fileURL, err := c.S3.GetPresignedURL(ctx, m.ObjectKey, 5*time.Minute)
		if err != nil {
			c.Logger.Error("s3 failed get presigned url error", slog.Any("error", err))
			return nil, err
		}
		m.ObjectURL = fileURL
	}

	return m, err
}
