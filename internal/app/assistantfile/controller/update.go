package controller

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	a_d "github.com/bartmika/databoutique-backend/internal/app/assistantfile/datastore"
	user_d "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

type AssistantFileUpdateRequestIDO struct {
	ID          primitive.ObjectID
	Name        string
	Description string
	FileName    string
	FileType    string
	File        multipart.File
}

func validateUpdateRequest(dirtyData *AssistantFileUpdateRequestIDO) error {
	e := make(map[string]string)

	if dirtyData.ID.IsZero() {
		e["id"] = "missing value"
	}
	if dirtyData.Name == "" {
		e["name"] = "missing value"
	}
	if dirtyData.Description == "" {
		e["description"] = "missing value"
	}
	if len(e) != 0 {
		return httperror.NewForBadRequest(&e)
	}
	return nil
}

func (impl *AssistantFileControllerImpl) UpdateByID(ctx context.Context, req *AssistantFileUpdateRequestIDO) (*a_d.AssistantFile, error) {
	if err := validateUpdateRequest(req); err != nil {
		return nil, err
	}

	// Fetch the original assistantfile.
	os, err := impl.AssistantFileStorer.GetByID(ctx, req.ID)
	if err != nil {
		impl.Logger.Error("database get by id error",
			slog.Any("error", err),
			slog.Any("assistantfile_id", req.ID))
		return nil, err
	}
	if os == nil {
		impl.Logger.Error("assistantfile does not exist error",
			slog.Any("assistantfile_id", req.ID))
		return nil, httperror.NewForBadRequestWithSingleField("message", "assistantfile does not exist")
	}

	// Extract from our session the following data.
	userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	userTenantID := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	userTenantName := ctx.Value(constants.SessionUserTenantName).(string)
	userRole := ctx.Value(constants.SessionUserRole).(int8)
	userName := ctx.Value(constants.SessionUserName).(string)

	// If user is not administrator nor belongs to the assistantfile then error.
	if userRole != user_d.UserRoleExecutive {
		impl.Logger.Error("authenticated user is not staff role nor belongs to the assistantfile error",
			slog.Any("userRole", userRole),
			slog.Any("userTenantID", userTenantID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "you do not belong to this assistantfile")
	}

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

		// Update the file if the user uploaded a new file.
		if req.File != nil {
			// The following code will choose the directory we will upload based on the image type.
			var directory string = "assistant-files"

			// Generate the key of our upload.
			objectKey := fmt.Sprintf("tenant/%v/%v/%v", userTenantID.Hex(), directory, req.FileName)

			// For debugging purposes only.
			impl.Logger.Debug("pre-upload meta",
				slog.String("FileName", req.FileName),
				slog.String("FileType", req.FileType),
				slog.String("Directory", directory),
				slog.String("ObjectKey", objectKey),
				slog.String("Name", req.Name),
				slog.String("Desc", req.Description),
				slog.Any("tenantID", userTenantID),
				slog.String("tenantName", userTenantName),
				slog.Any("userID", userID),
				slog.String("userName", userName),
			)

			// 	// Proceed to delete the physical files from AWS s3.
			// 	if err := impl.S3.DeleteByKeys(ctx, []string{os.ObjectKey}); err != nil {
			// 		impl.Logger.Warn("s3 delete by keys error", slog.Any("error", err))
			// 		// Do not return an error, simply continue this function as there might
			// 		// be a case were the file was removed on the s3 bucket by ourselves
			// 		// or some other reason.
			// 	}
			//
			// 	// The following code will choose the directory we will upload based on the image type.
			// 	var directory string = "assistant-files"
			//
			// 	// Generate the key of our upload.
			// 	objectKey := fmt.Sprintf("tenant/%v/%v/%v", tenantID.Hex(), directory, req.FileName)
			//
			// 	// go func(file multipart.File, objkey string) {
			// 	// 	impl.Logger.Debug("beginning private s3 image upload...")
			// 	// 	if err := impl.S3.UploadContentFromMulipart(context.Background(), objkey, file); err != nil {
			// 	// 		impl.Logger.Error("private s3 upload error", slog.Any("error", err))
			// 	// 		// Do not return an error, simply continue this function as there might
			// 	// 		// be a case were the file was removed on the s3 bucket by ourselves
			// 	// 		// or some other reason.
			// 	// 	}
			// 	// 	impl.Logger.Debug("Finished private s3 image upload")
			// 	// }(req.File, objectKey)
			//
			// 	// Update file.
			// 	os.ObjectKey = objectKey
			// 	os.Filename = req.FileName

			creds, err := impl.TenantStorer.GetOpenAICredentialsByID(ctx, userTenantID)
			if err != nil {
				impl.Logger.Error("failed file upload to openai",
					slog.String("tenant_id", userTenantID.Hex()),
					slog.Any("error", err))
				return nil, err
			}
			if creds == nil {
				return nil, errors.New("no openai credentials returned")
			}

			impl.Logger.Debug("beginning openai file upload...", slog.String("tenant_id", userTenantID.Hex()))
			fileID, err := impl.uploadContentFromMulipart(context.Background(), req.FileName, req.File, creds.APIKey, creds.OrgKey)
			if err != nil {
				impl.Logger.Error("failed file upload to openai",
					slog.String("tenant_id", userTenantID.Hex()),
					slog.Any("error", err))
				return nil, err
			}
			if fileID == "" {
				impl.Logger.Debug("no openai `file_id` returned",
					slog.String("tenant_id", userTenantID.Hex()))
				return nil, errors.New("failed file upload to openai as no `file_id` was returned")
			}
			impl.Logger.Debug("finished file upload to openai",
				slog.String("tenant_id", userTenantID.Hex()),
				slog.Any("file_id", fileID))
		}

		// Modify our original assistantfile.
		os.ModifiedAt = time.Now()
		os.ModifiedByUserID = userID
		os.ModifiedByUserName = userName
		os.Name = req.Name
		os.Description = req.Description

		// Save to the database the modified assistantfile.
		if err := impl.AssistantFileStorer.UpdateByID(ctx, os); err != nil {
			impl.Logger.Error("database update by id error", slog.Any("error", err))
			return nil, err
		}

		// go func(org *a_d.AssistantFile) {
		// 	impl.updateAssistantFileNameForAllUsers(ctx, org)
		// }(os)
		// go func(org *a_d.AssistantFile) {
		// 	impl.updateAssistantFileNameForAllComicSubmissions(ctx, org)
		// }(os)

		return os, nil
	}

	// Start a transaction
	result, err := session.WithTransaction(ctx, transactionFunc)
	if err != nil {
		impl.Logger.Error("session failed error",
			slog.Any("error", err))
		return nil, err
	}

	return result.(*a_d.AssistantFile), nil
}
