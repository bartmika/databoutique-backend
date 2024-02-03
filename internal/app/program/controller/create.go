package controller

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	program_s "github.com/bartmika/databoutique-backend/internal/app/program/datastore"
	u_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

type ProgramCreateRequestIDO struct {
	Name               string               `bson:"name" json:"name"`
	Description        string               `bson:"description" json:"description"`
	Instructions       string               `bson:"instructions" json:"instructions"`
	Model              string               `bson:"model" json:"model"`
	BusinessFunction   int8                 `bson:"business_function" json:"business_function"`
	UploadDirectoryIDs []primitive.ObjectID `bson:"upload_directory_ids" json:"upload_directory_ids"`
}

func (impl *ProgramControllerImpl) validateCreateRequest(ctx context.Context, dirtyData *ProgramCreateRequestIDO) error {
	e := make(map[string]string)

	if dirtyData.Name == "" {
		e["name"] = "missing value"
	}
	if dirtyData.Description == "" {
		e["description"] = "missing value"
	}
	if dirtyData.Instructions == "" {
		e["instructions"] = "missing value"
	}
	if dirtyData.Model == "" {
		e["model"] = "missing value"
	}
	if dirtyData.BusinessFunction == 0 {
		e["business_function"] = "missing value"
	}
	if len(dirtyData.UploadDirectoryIDs) == 0 && dirtyData.BusinessFunction == 2 {
		e["upload_directory_ids"] = "missing value"
	}

	if len(e) != 0 {
		return httperror.NewForBadRequest(&e)
	}
	return nil
}

func (impl *ProgramControllerImpl) Create(ctx context.Context, requestData *ProgramCreateRequestIDO) (*program_s.Program, error) {
	//
	// Get variables from our user authenticated session.
	//

	tid, _ := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	role, _ := ctx.Value(constants.SessionUserRole).(int8)
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
	impl.Kmutex.Lockf("program-by-tenant-%s", tid.Hex())
	defer impl.Kmutex.Unlockf("program-by-tenant-%s", tid.Hex())

	//
	// Perform our validation and return validation error on any issues detected.
	//

	if err := impl.validateCreateRequest(ctx, requestData); err != nil {
		impl.Logger.Error("validation error", slog.Any("error", err))
		return nil, err
	}

	switch role {
	case u_s.UserRoleExecutive, u_s.UserRoleManagement, u_s.UserRoleFrontlineStaff:
		break
	default:
		impl.Logger.Error("you do not have permission to create a client")
		return nil, httperror.NewForForbiddenWithSingleField("message", "you do not have permission to create a client")
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

		////
		//// Get related records.
		////

		uploadFolders, err := impl.UploadDirectoryStorer.ListByIDs(sessCtx, requestData.UploadDirectoryIDs)
		if err != nil {
			impl.Logger.Error("failed getting folders",
				slog.Any("upload_directory_ids", requestData.UploadDirectoryIDs),
				slog.Any("error", err))
			return nil, err
		}

		////
		//// Create the record.
		////

		prog := &program_s.Program{
			Name:             requestData.Name,
			Description:      requestData.Description,
			Instructions:     requestData.Instructions,
			Model:            requestData.Model,
			Status:           program_s.ProgramStatusActive,
			BusinessFunction: requestData.BusinessFunction,
		}

		// Add defaults.
		prog.TenantID = tid
		prog.ID = primitive.NewObjectID()
		prog.CreatedAt = time.Now()
		prog.CreatedByUserID = userID
		prog.CreatedByUserName = userName
		prog.CreatedFromIPAddress = ipAddress
		prog.ModifiedAt = time.Now()
		prog.ModifiedByUserID = userID
		prog.ModifiedByUserName = userName
		prog.ModifiedFromIPAddress = ipAddress

		// Add related.
		if prog.BusinessFunction == program_s.ProgramBusinessFunctionAdmintorDocumentReview && len(uploadFolders.Results) > 0 {
			// 1. Iterate through all the selected folders and make a copy of them
			// 2. Iterate through all the files within the selected folders and
			//    make a copy of them.
			// 3. Save the copy to our program.
			prog.Directories = make([]*program_s.UploadFolderOption, 0)
			for _, folder := range uploadFolders.Results {
				dir := &program_s.UploadFolderOption{
					ID:          folder.ID,
					Name:        folder.Name,
					Description: folder.Description,
					Status:      folder.Status,
					Files:       make([]*program_s.UploadFileOption, 0),
				}
				dirfiles, err := impl.UploadFileStorer.ListByUploadDirectoryID(ctx, folder.ID)
				if err != nil {
					impl.Logger.Error("failed getting files within folder",
						slog.Any("upload_directory_id", folder.ID),
						slog.Any("error", err))
					return nil, err
				}
				for _, dirfile := range dirfiles.Results {
					file := &program_s.UploadFileOption{
						ID:           dirfile.ID,
						Name:         dirfile.Name,
						Description:  dirfile.Description,
						OpenAIFileID: dirfile.OpenAIFileID,
						Status:       dirfile.Status,
					}
					dir.Files = append(dir.Files, file)
				}
				prog.Directories = append(prog.Directories, dir)
			}
		}

		// Save to our database.
		if err := impl.ProgramStorer.Create(sessCtx, prog); err != nil {
			impl.Logger.Error("database create error", slog.Any("error", err))
			return nil, err
		}

		////
		//// Exit our transaction successfully.
		////

		return prog, nil
	}

	// Start a transaction
	result, err := session.WithTransaction(ctx, transactionFunc)
	if err != nil {
		impl.Logger.Error("session failed error",
			slog.Any("error", err))
		return nil, err
	}

	// Convert from MongoDB transaction format into our data format.
	prog := result.(*program_s.Program)

	////
	//// Execute in background calling OpenAI API.
	////

	// Submit the following into the background of this web-application.
	// This function will run independently of this function call.
	go func(p *program_s.Program) {
		if p.BusinessFunction == program_s.ProgramBusinessFunctionAdmintorDocumentReview {
			if err := impl.createProgramInBackgroundForOpenAI(p); err != nil {
				impl.Logger.Error("failed submitting to openai", slog.Any("error", err))
			}
		}
	}(prog)

	return prog, nil
}
