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
	// OwnershipTypeTemporary indicates file has been uploaded and saved in our system but not assigned ownership to anything. As a result, if this fileinfo is not assigned within 24 hours then the crontab will delete this fileinfo record and the uploaded file.
	OwnershipTypeTemporary = 1
	OwnershipTypeUser      = 4
	OwnershipTypeTenant    = 5
	ContentTypeFile        = 6
	ContentTypeImage       = 7
)

type FileInfo struct {
	Name               string             `bson:"name" json:"name"`
	Description        string             `bson:"description" json:"description"`
	Filename           string             `bson:"filename" json:"filename"`
	ID                 primitive.ObjectID `bson:"_id" json:"id"`
	OpenAIFileID       string             `bson:"openai_file_id" json:"openai_file_id"` // https://platform.openai.com/docs/assistants/tools/supported-files
	CreatedAt          time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty"`
	CreatedByUserName  string             `bson:"created_by_user_name" json:"created_by_user_name"`
	CreatedByUserID    primitive.ObjectID `bson:"created_by_user_id" json:"created_by_user_id"`
	ModifiedAt         time.Time          `bson:"modified_at,omitempty" json:"modified_at,omitempty"`
	ModifiedByUserName string             `bson:"modified_by_user_name" json:"modified_by_user_name"`
	ModifiedByUserID   primitive.ObjectID `bson:"modified_by_user_id" json:"modified_by_user_id"`
	ObjectKey          string             `bson:"object_key" json:"object_key"`
	ObjectURL          string             `bson:"object_url" json:"object_url"`
	Status             int8               `bson:"status" json:"status"`
	ContentType        int8               `bson:"content_type" json:"content_type"`
	TenantID           primitive.ObjectID `bson:"tenant_id,omitempty" json:"tenant_id,omitempty"`
	TenantName         string             `bson:"tenant_name" json:"tenant_name"`
}

type FileInfoAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"name" json:"label"`
}

// FileInfoStorer Interface for fileinfo.
type FileInfoStorer interface {
	Create(ctx context.Context, m *FileInfo) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*FileInfo, error)
	UpdateByID(ctx context.Context, m *FileInfo) error
	ListByFilter(ctx context.Context, m *FileInfoPaginationListFilter) (*FileInfoPaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *FileInfoPaginationListFilter) ([]*FileInfoAsSelectOption, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	// //TODO: Add more...
}

type FileInfoStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) FileInfoStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("assistant_files")

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

	s := &FileInfoStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
