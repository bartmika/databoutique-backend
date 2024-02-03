package controller

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	executable_s "github.com/bartmika/databoutique-backend/internal/app/executable/datastore"
	program_s "github.com/bartmika/databoutique-backend/internal/app/program/datastore"
	uploaddirectory_s "github.com/bartmika/databoutique-backend/internal/app/uploaddirectory/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

type ExecutableCreateRequestIDO struct {
	ProgramID          primitive.ObjectID   `bson:"program_id" json:"program_id"`
	UserID             primitive.ObjectID   `bson:"user_id" json:"user_id"`
	UploadDirectoryIDs []primitive.ObjectID `bson:"upload_directory_ids" json:"upload_directory_ids"`
	Question           string               `bson:"question" json:"question"`
}

func (impl *ExecutableControllerImpl) validateCreateRequest(ctx context.Context, dirtyData *ExecutableCreateRequestIDO) error {
	e := make(map[string]string)

	if dirtyData.ProgramID.IsZero() {
		e["program_id"] = "missing value"
	}
	if dirtyData.UserID.IsZero() {
		e["user_id"] = "missing value"
	}
	if dirtyData.Question == "" {
		e["question"] = "missing value"
	}
	if len(e) != 0 {
		return httperror.NewForBadRequest(&e)
	}
	return nil
}

func (impl *ExecutableControllerImpl) Create(ctx context.Context, requestData *ExecutableCreateRequestIDO) (*executable_s.Executable, error) {
	//
	// Get variables from our user authenticated session.
	//

	tid, _ := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	// role, _ := ctx.Value(constants.SessionUserRole).(int8)
	userID, _ := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	userName, _ := ctx.Value(constants.SessionUserName).(string)
	ipAddress, _ := ctx.Value(constants.SessionIPAddress).(string)

	// DEVELOPERS NOTE:
	// Every submission needs to have a unique `public id` (PID)
	// generated. The following needs to happen to generate the unique PID:
	// 1. Make the `Create` function be `atomic` and thus lock this function.
	// 2. Count total records in system (for particular tenant).
	// 3. Generate PID.
	// 4. Apply the PID to the record.
	// 5. Unlock this `Create` function to be usable again by other calls after
	//    the function successfully submits the record into our system.
	impl.Kmutex.Lockf("create-executable-by-tenant-%s", tid.Hex())
	defer impl.Kmutex.Unlockf("create-executable-by-tenant-%s", tid.Hex())

	//
	// Perform our validation and return validation error on any issues detected.
	//

	if err := impl.validateCreateRequest(ctx, requestData); err != nil {
		impl.Logger.Error("validation error", slog.Any("error", err))
		return nil, err
	}

	// switch role {
	// case u_s.UserRoleExecutive, u_s.UserRoleManagement, u_s.UserRoleFrontlineStaff:
	// 	break
	// default:
	// 	impl.Logger.Error("you do not have permission to create a client")
	// 	return nil, httperror.NewForForbiddenWithSingleField("message", "you do not have permission to create a client")
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

		////
		//// Get related data.
		////

		u, err := impl.UserStorer.GetByID(sessCtx, requestData.UserID)
		if err != nil {
			impl.Logger.Error("failed getting user",
				slog.Any("error", err))
			return nil, err
		}
		if u == nil {
			err := fmt.Errorf("user does not exist for id: %v", requestData.UserID.Hex())
			impl.Logger.Error("user does not exist", slog.Any("error", err))
			return nil, err
		}
		p, err := impl.ProgramStorer.GetByID(sessCtx, requestData.ProgramID)
		if err != nil {
			impl.Logger.Error("failed getting program",
				slog.Any("error", err))
			return nil, err
		}
		if p == nil {
			err := fmt.Errorf("program does not exist for id: %v", requestData.ProgramID.Hex())
			impl.Logger.Error("program does not exist", slog.Any("error", err))
			return nil, err
		}

		// Handle the two cases, either the customer provides the files or we
		// use the admin files.
		var uploadFolders *uploaddirectory_s.UploadDirectoryPaginationListResult
		if p.BusinessFunction == program_s.ProgramBusinessFunctionCustomerDocumentReview {
			uploadFolders, err = impl.UploadDirectoryStorer.ListByIDs(sessCtx, requestData.UploadDirectoryIDs)
			if err != nil {
				impl.Logger.Error("failed getting folders",
					slog.Any("upload_directory_ids", requestData.UploadDirectoryIDs),
					slog.Any("error", err))
				return nil, err
			}

			// DEFENSIVE CODE: If the program is `customer document review` and
			// there are no documents provided by the customer then we will
			// generate a 400 bad request.
			if len(uploadFolders.Results) == 0 {
				return nil, httperror.NewForSingleField(http.StatusBadRequest, "upload_directory_ids", "missing value")
			}
		}
		if p.BusinessFunction == program_s.ProgramBusinessFunctionAdmintorDocumentReview {
			uploadFolderIDs := p.GetUploadDirectoryIDs()
			uploadFolders, err = impl.UploadDirectoryStorer.ListByIDs(sessCtx, uploadFolderIDs)
			if err != nil {
				impl.Logger.Error("failed getting folders",
					slog.Any("upload_directory_ids", requestData.UploadDirectoryIDs),
					slog.Any("error", err))
				return nil, err
			}
		}

		////
		//// Create record.
		////

		exec := &executable_s.Executable{}

		// Add defaults.
		exec.TenantID = tid
		exec.ID = primitive.NewObjectID()
		exec.CreatedAt = time.Now()
		exec.CreatedByUserID = userID
		exec.CreatedByUserName = userName
		exec.CreatedFromIPAddress = ipAddress
		exec.ModifiedAt = time.Now()
		exec.ModifiedByUserID = userID
		exec.ModifiedByUserName = userName
		exec.ModifiedFromIPAddress = ipAddress

		// Add base.
		exec.ProgramID = p.ID
		exec.ProgramName = p.Name
		exec.Question = requestData.Question
		exec.Status = executable_s.ExecutableStatusProcessing
		exec.OpenAIAssistantID = ""
		exec.UserID = u.ID
		exec.UserName = u.Name
		exec.UserLexicalName = u.LexicalName

		if len(uploadFolders.Results) > 0 {
			// Add related.
			// 1. Iterate through all the selected folders and make a copy of them
			// 2. Iterate through all the files within the selected folders and
			//    make a copy of them.
			// 3. Save the copy to our executable.
			// 4. Add initial question into our messages list.
			exec.Directories = make([]*executable_s.UploadFolderOption, 0)
			for _, folder := range uploadFolders.Results {
				dir := &executable_s.UploadFolderOption{
					ID:          folder.ID,
					Name:        folder.Name,
					Description: folder.Description,
					Status:      folder.Status,
					Files:       make([]*executable_s.UploadFileOption, 0),
				}
				dirfiles, err := impl.UploadFileStorer.ListByUploadDirectoryID(ctx, folder.ID)
				if err != nil {
					impl.Logger.Error("failed getting files within folder",
						slog.Any("upload_directory_id", folder.ID),
						slog.Any("error", err))
					return nil, err
				}
				for _, dirfile := range dirfiles.Results {
					file := &executable_s.UploadFileOption{
						ID:           dirfile.ID,
						Name:         dirfile.Name,
						Description:  dirfile.Description,
						OpenAIFileID: dirfile.OpenAIFileID,
						Status:       dirfile.Status,
					}
					dir.Files = append(dir.Files, file)
				}
				exec.Directories = append(exec.Directories, dir)
			}
		}

		msg := &executable_s.Message{
			ID:              primitive.NewObjectID(),
			Content:         requestData.Question,
			OpenAIMessageID: "",
			CreatedAt:       time.Now(),
			Status:          executable_s.ExecutableStatusActive,
			FromExecutable:  false,
		}
		exec.Messages = append(exec.Messages, msg)

		// Save to our database.
		if err := impl.ExecutableStorer.Create(sessCtx, exec); err != nil {
			impl.Logger.Error("database create error",
				slog.Any("error", err))
			return nil, err
		}

		////
		//// Exit our transaction successfully.
		////

		return exec, nil
	}

	// Start a transaction
	result, err := session.WithTransaction(ctx, transactionFunc)
	if err != nil {
		impl.Logger.Error("session failed error",
			slog.Any("error", err))
		return nil, err
	}

	// Convert from MongoDB transaction format into our data format.
	exec := result.(*executable_s.Executable)

	////
	//// Execute in background calling OpenAI API.
	////

	// Submit the following into the background of this web-application.
	// This function will run independently of this function call.
	go func(ex *executable_s.Executable) {
		if err := impl.createExecutableInBackgroundForOpenAI(ex); err != nil {
			impl.Logger.Error("failed submitting to openai", slog.Any("error", err))
		}
	}(exec)

	return exec, nil
}
