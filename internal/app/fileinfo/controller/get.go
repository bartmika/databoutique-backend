package controller

import (
	"context"
	"log/slog"
	"time"

	domain "github.com/bartmika/databoutique-backend/internal/app/fileinfo/datastore"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (c *FileInfoControllerImpl) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.FileInfo, error) {
	// // Extract from our session the following data.
	// userFileInfoID := ctx.Value(constants.SessionUserFileInfoID).(primitive.ObjectID)
	// userRole := ctx.Value(constants.SessionUserRole).(int8)
	//
	// If user is not administrator nor belongs to the fileinfo then error.
	// if userRole != user_d.UserRoleRoot && id != userFileInfoID {
	// 	c.Logger.Error("authenticated user is not staff role nor belongs to the fileinfo error",
	// 		slog.Any("userRole", userRole),
	// 		slog.Any("userFileInfoID", userFileInfoID))
	// 	return nil, httperror.NewForForbiddenWithSingleField("message", "you do not belong to this fileinfo")
	// }

	// Retrieve from our database the record for the specific id.
	m, err := c.FileInfoStorer.GetByID(ctx, id)
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
