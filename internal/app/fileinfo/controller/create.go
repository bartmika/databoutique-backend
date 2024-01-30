package controller

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log/slog"
	"mime/multipart"
	"reflect"
	"time"

	"github.com/sashabaranov/go-openai"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	a_d "github.com/bartmika/databoutique-backend/internal/app/fileinfo/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

// Special:
// https://freedium.cfd/https://story.tomasen.org/openais-assistant-api-in-go-a-practical-guide-4b9e7243ebff

type FileInfoCreateRequestIDO struct {
	Name        string
	Description string
	FileName    string
	FileType    string
	File        multipart.File
}

func validateCreateRequest(dirtyData *FileInfoCreateRequestIDO) error {
	e := make(map[string]string)

	if dirtyData.Name == "" {
		e["name"] = "missing value"
	}
	if dirtyData.Description == "" {
		e["description"] = "missing value"
	}
	if dirtyData.FileName == "" {
		e["file"] = "missing value"
	}
	if len(e) != 0 {
		return httperror.NewForBadRequest(&e)
	}
	return nil
}

func (impl *FileInfoControllerImpl) Create(ctx context.Context, req *FileInfoCreateRequestIDO) (*a_d.FileInfo, error) {
	// Extract from our session the following data.
	tenantID, _ := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	tenantName, _ := ctx.Value(constants.SessionUserTenantName).(string)
	userID, _ := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	userName, _ := ctx.Value(constants.SessionUserName).(string)

	if err := validateCreateRequest(req); err != nil {
		return nil, err
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
		// The following code will choose the directory we will upload based on the image type.
		var directory string = "assistant-files"

		// Generate the key of our upload.
		objectKey := fmt.Sprintf("tenant/%v/%v/%v", tenantID.Hex(), directory, req.FileName)

		// For debugging purposes only.
		impl.Logger.Debug("pre-upload meta",
			slog.String("FileName", req.FileName),
			slog.String("FileType", req.FileType),
			slog.String("Directory", directory),
			slog.String("ObjectKey", objectKey),
			slog.String("Name", req.Name),
			slog.String("Desc", req.Description),
			slog.Any("tenantID", tenantID),
			slog.String("tenantName", tenantName),
			slog.Any("userID", userID),
			slog.String("userName", userName),
		)

		// go func(file multipart.File, objkey string) {
		// 	impl.Logger.Debug("beginning private s3 file upload...")
		// 	if err := impl.S3.UploadContentFromMulipart(context.Background(), objkey, file); err != nil {
		// 		impl.Logger.Error("private s3 file upload error", slog.Any("error", err))
		// 		// Do not return an error, simply continue this function as there might
		// 		// be a case were the file was removed on the s3 bucket by ourselves
		// 		// or some other reason.
		// 	}
		// 	impl.Logger.Debug("Finished private s3 file upload")
		// }(req.File, objectKey)

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

		impl.Logger.Debug("beginning openai file upload...", slog.String("tenant_id", tenantID.Hex()))
		fileID, err := impl.uploadContentFromMulipart(context.Background(), req.FileName, req.File, creds.APIKey, creds.OrgKey)
		if err != nil {
			impl.Logger.Error("failed file upload to openai",
				slog.String("tenant_id", tenantID.Hex()),
				slog.Any("error", err))
			return nil, err
		}
		if fileID == "" {
			impl.Logger.Debug("no openai `file_id` returned",
				slog.String("tenant_id", tenantID.Hex()))
			return nil, errors.New("failed file upload to openai as no `file_id` was returned")
		}
		impl.Logger.Debug("finished file upload to openai",
			slog.String("tenant_id", tenantID.Hex()),
			slog.Any("file_id", fileID))

		// Create our meta record in the database.
		res := &a_d.FileInfo{
			TenantID:           tenantID,
			TenantName:         tenantName,
			ID:                 primitive.NewObjectID(),
			CreatedAt:          time.Now(),
			CreatedByUserName:  userName,
			CreatedByUserID:    userID,
			ModifiedAt:         time.Now(),
			ModifiedByUserName: userName,
			ModifiedByUserID:   userID,
			Name:               req.Name,
			Description:        req.Description,
			Filename:           req.FileName,
			ObjectKey:          objectKey,
			ObjectURL:          "",
			Status:             a_d.StatusActive,
			OpenAIFileID:       fileID,
		}
		if err := impl.FileInfoStorer.Create(sessCtx, res); err != nil {
			impl.Logger.Error("assistant file create error",
				slog.String("tenant_id", tenantID.Hex()),
				slog.Any("error", err))
			return nil, err
		}
		return res, nil
	}

	// Start a transaction
	result, err := session.WithTransaction(ctx, transactionFunc)
	if err != nil {
		impl.Logger.Error("session failed error",
			slog.Any("error", err))
		return nil, err
	}

	return result.(*a_d.FileInfo), nil
}

func isStructEmpty(s interface{}) bool {
	val := reflect.ValueOf(s)
	zeroVal := reflect.Zero(val.Type())
	return reflect.DeepEqual(val.Interface(), zeroVal.Interface())
}

func (impl *FileInfoControllerImpl) uploadContentFromMulipart(ctx context.Context, filename string, file multipart.File, apikey string, orgKey string) (string, error) {
	impl.Logger.Debug("openai initializing...")
	client := openai.NewOrgClient(apikey, orgKey)
	impl.Logger.Debug("openai initialized")

	// Read the contents of the file into a byte slice
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		impl.Logger.Error("failed converting multipart file into file bytes", slog.Any("error", err))
		return "", err
	}

	openAIFile, err := client.CreateFileBytes(context.Background(), openai.FileBytesRequest{
		Name:    filename,
		Bytes:   fileBytes,
		Purpose: openai.PurposeAssistants,
	})
	if err != nil {
		impl.Logger.Error("failed uploaded openai file", slog.Any("error", err))
		return "", err
	}
	if isStructEmpty(openAIFile) {
		impl.Logger.Error("no openai file returned")
		return "", errors.New("no openai file returned")
	}

	return openAIFile.ID, nil
}
