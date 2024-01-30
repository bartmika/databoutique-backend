package controller

import (
	"context"
	"errors"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"

	attch_d "github.com/bartmika/databoutique-backend/internal/app/attachment/datastore"
	user_d "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func (impl *AttachmentControllerImpl) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	// Extract from our session the following data.
	userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	userRole := ctx.Value(constants.SessionUserRole).(int8)

	// Apply protection based on ownership and role.
	if userRole != user_d.UserRoleExecutive {
		impl.Logger.Error("authenticated user is not staff role error",
			slog.Any("role", userRole),
			slog.Any("userID", userID))
		return httperror.NewForForbiddenWithSingleField("message", "you role does not grant you access to this")
	}

	// Update the database.
	attachment, err := impl.GetByID(ctx, id)
	attachment.Status = attch_d.StatusArchived
	if err != nil {
		impl.Logger.Error("database get by id error", slog.Any("error", err))
		return err
	}
	if attachment == nil {
		impl.Logger.Error("database returns nothing from get by id")
		return err
	}
	// // Security: Prevent deletion of root user(s).
	// if attachment.Type == attch_d.RootType {
	// 	impl.Logger.Warn("root attachment cannot be deleted error")
	// 	return httperror.NewForForbiddenWithSingleField("role", "root attachment cannot be deleted")
	// }

	// Save to the database the modified attachment.
	if err := impl.AttachmentStorer.UpdateByID(ctx, attachment); err != nil {
		impl.Logger.Error("database update by id error", slog.Any("error", err))
		return err
	}
	return nil
}

func (impl *AttachmentControllerImpl) PermanentlyDeleteByID(ctx context.Context, id primitive.ObjectID) error {
	// Extract from our session the following data.
	userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	userRole := ctx.Value(constants.SessionUserRole).(int8)

	// Apply protection based on ownership and role.
	if userRole != user_d.UserRoleExecutive {
		impl.Logger.Error("authenticated user is not staff role error",
			slog.Any("role", userRole),
			slog.Any("userID", userID))
		return httperror.NewForForbiddenWithSingleField("message", "you role does not grant you access to this")
	}

	// Update the database.
	attachment, err := impl.GetByID(ctx, id)
	if err != nil {
		impl.Logger.Error("database get by id error", slog.Any("error", err))
		return err
	}
	if attachment == nil {
		impl.Logger.Error("database returns nothing from get by id")
		return errors.New("does not exist")
	}

	// Proceed to delete the physical files from AWS s3.
	if err := impl.S3.DeleteByKeys(ctx, []string{attachment.ObjectKey}); err != nil {
		impl.Logger.Warn("s3 delete by keys error", slog.Any("error", err))
		// Do not return an error, simply continue this function as there might
		// be a case were the file was removed on the s3 bucket by ourselves
		// or some other reason.
	}
	impl.Logger.Debug("deleted from s3", slog.Any("attachment_id", id))

	if err := impl.AttachmentStorer.DeleteByID(ctx, attachment.ID); err != nil {
		impl.Logger.Error("database delete by id error", slog.Any("error", err))
		return err
	}
	impl.Logger.Debug("deleted from database", slog.Any("attachment_id", id))

	// Update exercise.
	if !attachment.OwnershipID.IsZero() {

	}
	return nil
}
