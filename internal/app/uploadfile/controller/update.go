package controller

import (
	"context"
	"errors"
	"log/slog"
	"mime/multipart"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	a_d "github.com/bartmika/databoutique-backend/internal/app/uploadfile/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

type UploadFileUpdateRequestIDO struct {
	ID                primitive.ObjectID
	Name              string
	Description       string
	FileName          string
	FileType          string
	File              multipart.File
	UploadDirectoryID primitive.ObjectID
}

func validateUpdateRequest(dirtyData *UploadFileUpdateRequestIDO) error {
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
	if dirtyData.UploadDirectoryID.IsZero() {
		e["upload_directory_id"] = "missing value"
	}
	if len(e) != 0 {
		return httperror.NewForBadRequest(&e)
	}
	return nil
}

func (impl *UploadFileControllerImpl) UpdateByID(ctx context.Context, req *UploadFileUpdateRequestIDO) (*a_d.UploadFile, error) {
	if err := validateUpdateRequest(req); err != nil {
		return nil, err
	}

	// Fetch the original uploadfile.
	os, err := impl.UploadFileStorer.GetByID(ctx, req.ID)
	if err != nil {
		impl.Logger.Error("database get by id error",
			slog.Any("error", err),
			slog.Any("uploadfile_id", req.ID))
		return nil, err
	}
	if os == nil {
		impl.Logger.Error("uploadfile does not exist error",
			slog.Any("uploadfile_id", req.ID))
		return nil, httperror.NewForBadRequestWithSingleField("message", "uploadfile does not exist")
	}

	// Extract from our session the following data.
	tenantID, _ := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	userTenantID := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	userTenantName := ctx.Value(constants.SessionUserTenantName).(string)
	// userRole := ctx.Value(constants.SessionUserRole).(int8)
	userName := ctx.Value(constants.SessionUserName).(string)
	userLexicalName, _ := ctx.Value(constants.SessionUserLexicalName).(string)

	// // If user is not administrator nor belongs to the uploadfile then error.
	// if userRole != user_d.UserRoleExecutive {
	// 	impl.Logger.Error("authenticated user is not staff role nor belongs to the uploadfile error",
	// 		slog.Any("userRole", userRole),
	// 		slog.Any("userTenantID", userTenantID))
	// 	return nil, httperror.NewForForbiddenWithSingleField("message", "you do not belong to this uploadfile")
	// }

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

		uploadDirectory, err := impl.UploadDirectoryStorer.GetByID(sessCtx, req.UploadDirectoryID)
		if err != nil {
			impl.Logger.Error("failed getting upload directory",
				slog.String("tenant_id", tenantID.Hex()),
				slog.String("upload_directory_id", req.UploadDirectoryID.Hex()),
				slog.Any("error", err))
			return nil, err
		}
		if uploadDirectory == nil {
			return nil, errors.New("upload directory does not exist")
		}

		// Update the file if the user uploaded a new file.
		if req.File != nil {
			// For debugging purposes only.
			impl.Logger.Debug("pre-upload meta",
				slog.String("FileName", req.FileName),
				slog.String("FileType", req.FileType),
				slog.String("Name", req.Name),
				slog.String("Desc", req.Description),
				slog.Any("tenantID", userTenantID),
				slog.String("tenantName", userTenantName),
				slog.Any("userID", userID),
				slog.String("userName", userName),
			)

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

		// Modify our original uploadfile.
		os.ModifiedAt = time.Now()
		os.ModifiedByUserID = userID
		os.ModifiedByUserName = userName
		os.UploadDirectoryID = uploadDirectory.ID
		os.UploadDirectoryName = uploadDirectory.Name
		os.Name = req.Name
		os.Description = req.Description
		os.UserID = userID
		os.UserName = userName
		os.UserLexicalName = userLexicalName

		// Save to the database the modified uploadfile.
		if err := impl.UploadFileStorer.UpdateByID(ctx, os); err != nil {
			impl.Logger.Error("database update by id error", slog.Any("error", err))
			return nil, err
		}

		// go func(org *a_d.UploadFile) {
		// 	impl.updateUploadFileNameForAllUsers(ctx, org)
		// }(os)
		// go func(org *a_d.UploadFile) {
		// 	impl.updateUploadFileNameForAllComicSubmissions(ctx, org)
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

	return result.(*a_d.UploadFile), nil
}
