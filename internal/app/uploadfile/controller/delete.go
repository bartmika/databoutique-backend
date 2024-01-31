package controller

import (
	"context"
	"errors"
	"log/slog"

	"github.com/sashabaranov/go-openai"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	attch_d "github.com/bartmika/databoutique-backend/internal/app/uploadfile/datastore"
	user_d "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func (impl *UploadFileControllerImpl) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
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

		// Update the database.
		uploadfile, err := impl.GetByID(sessCtx, id)
		uploadfile.Status = attch_d.StatusArchived
		if err != nil {
			impl.Logger.Error("database get by id error", slog.Any("error", err))
			return nil, err
		}
		if uploadfile == nil {
			impl.Logger.Error("database returns nothing from get by id")
			return nil, err
		}
		// // Security: Prevent deletion of root user(s).
		// if uploadfile.Type == attch_d.RootType {
		// 	impl.Logger.Warn("root uploadfile cannot be deleted error")
		// 	return httperror.NewForForbiddenWithSingleField("role", "root uploadfile cannot be deleted")
		// }

		// Save to the database the modified uploadfile.
		if err := impl.UploadFileStorer.UpdateByID(sessCtx, uploadfile); err != nil {
			impl.Logger.Error("database update by id error", slog.Any("error", err))
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

func (impl *UploadFileControllerImpl) PermanentlyDeleteByID(ctx context.Context, id primitive.ObjectID) error {
	// Extract from our session the following data.
	tenantID, _ := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	userID, _ := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	userRole, _ := ctx.Value(constants.SessionUserRole).(int8)

	// Apply protection based on ownership and role.
	if userRole != user_d.UserRoleExecutive {
		impl.Logger.Error("authenticated user is not staff role error",
			slog.Any("role", userRole),
			slog.Any("userID", userID))
		return httperror.NewForForbiddenWithSingleField("message", "you role does not grant you access to this")
	}

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

		// Update the database.
		uploadfile, err := impl.GetByID(sessCtx, id)
		if err != nil {
			impl.Logger.Error("database get by id error", slog.Any("error", err))
			return nil, err
		}
		if uploadfile == nil {
			impl.Logger.Error("database returns nothing from get by id")
			return nil, errors.New("does not exist")
		}

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

		// // Proceed to delete the physical files from AWS s3.
		// if err := impl.S3.DeleteByKeys(ctx, []string{uploadfile.ObjectKey}); err != nil {
		// 	impl.Logger.Warn("s3 delete by keys error", slog.Any("error", err))
		// 	// Do not return an error, simply continue this function as there might
		// 	// be a case were the file was removed on the s3 bucket by ourselves
		// 	// or some other reason.
		// }
		// impl.Logger.Debug("deleted from s3", slog.Any("uploadfile_id", id))

		if err := impl.UploadFileStorer.DeleteByID(sessCtx, uploadfile.ID); err != nil {
			impl.Logger.Error("database delete by id error", slog.Any("error", err))
			return nil, err
		}
		impl.Logger.Debug("deleted from database", slog.Any("uploadfile_id", id))

		if err := impl.deleteOpanAIFile(sessCtx, uploadfile.OpenAIFileID, creds.APIKey, creds.OrgKey); err != nil {
			impl.Logger.Error("failed deleting file from openai", slog.Any("error", err))
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

func (impl *UploadFileControllerImpl) deleteOpanAIFile(ctx context.Context, fileID string, apikey string, orgKey string) error {
	impl.Logger.Debug("openai initializing...")
	client := openai.NewOrgClient(apikey, orgKey)
	impl.Logger.Debug("openai initialized")
	if err := client.DeleteFile(ctx, fileID); err != nil {
		impl.Logger.Error("failed deleting open ai file", slog.Any("error", err))
		return err
	}
	impl.Logger.Debug("deleted openai file from assistant api", slog.Any("uploadfile_id", fileID))
	return nil
}
