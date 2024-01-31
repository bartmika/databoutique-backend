package controller

import (
	"context"
	"errors"
	"io/ioutil"
	"log/slog"
	"mime/multipart"
	"reflect"
	"time"

	"github.com/sashabaranov/go-openai"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	a_d "github.com/bartmika/databoutique-backend/internal/app/uploadfile/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

// Special:
// https://freedium.cfd/https://story.tomasen.org/openais-assistant-api-in-go-a-practical-guide-4b9e7243ebff

type UploadFileCreateRequestIDO struct {
	Name              string
	Description       string
	FileName          string
	FileType          string
	File              multipart.File
	UploadDirectoryID primitive.ObjectID
}

func validateCreateRequest(dirtyData *UploadFileCreateRequestIDO) error {
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
	if dirtyData.UploadDirectoryID.IsZero() {
		e["upload_directory_id"] = "missing value"
	}
	if len(e) != 0 {
		return httperror.NewForBadRequest(&e)
	}
	return nil
}

func (impl *UploadFileControllerImpl) Create(ctx context.Context, req *UploadFileCreateRequestIDO) (*a_d.UploadFile, error) {
	// Extract from our session the following data.
	tenantID, _ := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	tenantName, _ := ctx.Value(constants.SessionUserTenantName).(string)
	userID, _ := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	userName, _ := ctx.Value(constants.SessionUserName).(string)
	userLexicalName, _ := ctx.Value(constants.SessionUserLexicalName).(string)

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

		// For debugging purposes only.
		impl.Logger.Debug("pre-upload meta",
			slog.String("FileName", req.FileName),
			slog.String("FileType", req.FileType),
			slog.String("Name", req.Name),
			slog.String("Desc", req.Description),
			slog.Any("tenantID", tenantID),
			slog.String("tenantName", tenantName),
			slog.Any("userID", userID),
			slog.String("userName", userName),
		)

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
		res := &a_d.UploadFile{
			UploadDirectoryID:   uploadDirectory.ID,
			UploadDirectoryName: uploadDirectory.Name,
			TenantID:            tenantID,
			TenantName:          tenantName,
			ID:                  primitive.NewObjectID(),
			CreatedAt:           time.Now(),
			CreatedByUserName:   userName,
			CreatedByUserID:     userID,
			ModifiedAt:          time.Now(),
			ModifiedByUserName:  userName,
			ModifiedByUserID:    userID,
			Name:                req.Name,
			Description:         req.Description,
			Filename:            req.FileName,
			ObjectKey:           "",
			ObjectURL:           "",
			Status:              a_d.StatusActive,
			OpenAIFileID:        fileID,
			UserID:              userID,
			UserName:            userName,
			UserLexicalName:     userLexicalName,
		}
		if err := impl.UploadFileStorer.Create(sessCtx, res); err != nil {
			impl.Logger.Error("upload file create error",
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

	return result.(*a_d.UploadFile), nil
}

func isStructEmpty(s interface{}) bool {
	val := reflect.ValueOf(s)
	zeroVal := reflect.Zero(val.Type())
	return reflect.DeepEqual(val.Interface(), zeroVal.Interface())
}

func (impl *UploadFileControllerImpl) uploadContentFromMulipart(ctx context.Context, filename string, file multipart.File, apikey string, orgKey string) (string, error) {
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
