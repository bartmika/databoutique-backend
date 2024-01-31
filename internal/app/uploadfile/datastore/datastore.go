package datastore

import (
	"context"
	"log"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	c "github.com/bartmika/databoutique-backend/internal/config"
)

const (
	StatusActive   = 1
	StatusError    = 2
	StatusArchived = 3
	// OwnershipTypeTemporary indicates file has been uploaded and saved in our system but not assigned ownership to anything. As a result, if this uploadfile is not assigned within 24 hours then the crontab will delete this uploadfile record and the uploaded file.
	OwnershipTypeTemporary = 1
	OwnershipTypeUser      = 4
	OwnershipTypeTenant    = 5
	ContentTypeFile        = 6
	ContentTypeImage       = 7
)

type UploadFile struct {
	Name                string             `bson:"name" json:"name"`
	Description         string             `bson:"description" json:"description"`
	Filename            string             `bson:"filename" json:"filename"`
	ID                  primitive.ObjectID `bson:"_id" json:"id"`
	OpenAIFileID        string             `bson:"openai_file_id" json:"openai_file_id"` // https://platform.openai.com/docs/assistants/tools/supported-files
	CreatedAt           time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty"`
	CreatedByUserName   string             `bson:"created_by_user_name" json:"created_by_user_name"`
	CreatedByUserID     primitive.ObjectID `bson:"created_by_user_id" json:"created_by_user_id"`
	ModifiedAt          time.Time          `bson:"modified_at,omitempty" json:"modified_at,omitempty"`
	ModifiedByUserName  string             `bson:"modified_by_user_name" json:"modified_by_user_name"`
	ModifiedByUserID    primitive.ObjectID `bson:"modified_by_user_id" json:"modified_by_user_id"`
	ObjectKey           string             `bson:"object_key" json:"object_key"`
	ObjectURL           string             `bson:"object_url" json:"object_url"`
	Status              int8               `bson:"status" json:"status"`
	ContentType         int8               `bson:"content_type" json:"content_type"`
	TenantID            primitive.ObjectID `bson:"tenant_id,omitempty" json:"tenant_id,omitempty"`
	TenantName          string             `bson:"tenant_name" json:"tenant_name"`
	UploadDirectoryID   primitive.ObjectID `bson:"upload_directory_id,omitempty" json:"upload_directory_id,omitempty"`
	UploadDirectoryName string             `bson:"upload_directory_name" json:"upload_directory_name"`
	UserID              primitive.ObjectID `bson:"user_id" json:"user_id"`
	UserName            string             `bson:"user_name" json:"user_name"`
	UserLexicalName     string             `bson:"user_lexical_name" json:"user_lexical_name"`
}

type UploadFileAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"name" json:"label"`
}

// UploadFileStorer Interface for uploadfile.
type UploadFileStorer interface {
	Create(ctx context.Context, m *UploadFile) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*UploadFile, error)
	UpdateByID(ctx context.Context, m *UploadFile) error
	ListByFilter(ctx context.Context, m *UploadFilePaginationListFilter) (*UploadFilePaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *UploadFilePaginationListFilter) ([]*UploadFileAsSelectOption, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	// //TODO: Add more...
}

type UploadFileStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) UploadFileStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("upload_files")

	// The following few lines of code will create the index for our app for this
	// colleciton.
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{"tenant_name", "text"},
			{"name", "text"},
			{"description", "text"},
			{"filename", "text"},
		},
	}
	_, err := uc.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		// It is important that we crash the app on startup to meet the
		// requirements of `google/wire` framework.
		log.Fatal(err)
	}

	s := &UploadFileStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
