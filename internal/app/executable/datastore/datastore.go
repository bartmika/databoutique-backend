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
	ExecutableStatusActive     = 1
	ExecutableStatusProcessing = 2
	ExecutableStatusArchived   = 3
)

type Executable struct {
	ID                      primitive.ObjectID    `bson:"_id" json:"id"`
	TenantID                primitive.ObjectID    `bson:"tenant_id" json:"tenant_id"`
	ProgramID               primitive.ObjectID    `bson:"program_id" json:"program_id"`
	ProgramName             string                `bson:"program_name" json:"program_name"`
	Question                string                `bson:"question" json:"question"`
	Status                  int8                  `bson:"status" json:"status"`
	PublicID                uint64                `bson:"public_id" json:"public_id"`
	CreatedAt               time.Time             `bson:"created_at" json:"created_at"`
	CreatedByUserID         primitive.ObjectID    `bson:"created_by_user_id" json:"created_by_user_id,omitempty"`
	CreatedByUserName       string                `bson:"created_by_user_name" json:"created_by_user_name"`
	CreatedFromIPAddress    string                `bson:"created_from_ip_address" json:"created_from_ip_address"`
	ModifiedAt              time.Time             `bson:"modified_at" json:"modified_at"`
	ModifiedByUserID        primitive.ObjectID    `bson:"modified_by_user_id" json:"modified_by_user_id,omitempty"`
	ModifiedByUserName      string                `bson:"modified_by_user_name" json:"modified_by_user_name"`
	ModifiedFromIPAddress   string                `bson:"modified_from_ip_address" json:"modified_from_ip_address"`
	Directories             []*UploadFolderOption `bson:"directories" json:"directories,omitempty"`
	OpenAIAssistantID       string                `bson:"openai_assistant_id" json:"openai_assistant_id"` // https://platform.openai.com/docs/assistants/tools/supported-files
	OpenAIAssistantThreadID string                `bson:"openai_assistant_thread_id" json:"openai_assistant_thread_id"`
	UserID                  primitive.ObjectID    `bson:"user_id" json:"user_id"`
	UserName                string                `bson:"user_name" json:"user_name"`
	UserLexicalName         string                `bson:"user_lexical_name" json:"user_lexical_name"`
	Messages                []*Message            `bson:"messages" json:"messages,omitempty"`
}

type UploadFolderOption struct {
	ID          primitive.ObjectID  `bson:"_id" json:"id"`
	Name        string              `bson:"name" json:"name"`
	Description string              `bson:"description" json:"description"`
	Status      int8                `bson:"status" json:"status"`
	Files       []*UploadFileOption `bson:"files" json:"files,omitempty"`
}

type UploadFileOption struct {
	ID           primitive.ObjectID `bson:"_id" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Description  string             `bson:"description" json:"description"`
	OpenAIFileID string             `bson:"openai_file_id" json:"openai_file_id"` // https://platform.openai.com/docs/assistants/tools/supported-files
	Status       int8               `bson:"status" json:"status"`
}

type Message struct {
	ID              primitive.ObjectID `bson:"_id" json:"id"`
	Content         string             `bson:"content" json:"content"`
	OpenAIMessageID string             `bson:"openai_message_id" json:"openai_message_id"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	Status          int8               `bson:"status" json:"status"`
	FromExecutable  bool               `bson:"from_executable" json:"from_executable"`
}

// GetOpenAIFileIDs function will iterate through all the assistant files
// and return the OpenAI file ID values.
func (a *Executable) GetOpenAIFileIDs() []string {
	if a.Directories == nil {
		return nil
	}
	ids := []string{}
	for _, dir := range a.Directories {
		for _, file := range dir.Files {
			ids = append(ids, file.OpenAIFileID)
		}
	}
	return ids
}

func (a *Executable) GetUploadDirectoryIDs() []primitive.ObjectID {
	if a.Directories == nil {
		return nil
	}
	ids := []primitive.ObjectID{}
	for _, dir := range a.Directories {
		ids = append(ids, dir.ID)
	}
	return ids
}

type ExecutableListResult struct {
	Results     []*Executable      `json:"results"`
	NextCursor  primitive.ObjectID `json:"next_cursor"`
	HasNextPage bool               `json:"has_next_page"`
}

type ExecutableAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"text" json:"label"`
}

// ExecutableStorer Interface for user.
type ExecutableStorer interface {
	Create(ctx context.Context, m *Executable) error
	CreateOrGetByID(ctx context.Context, hh *Executable) (*Executable, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*Executable, error)
	GetByPublicID(ctx context.Context, oldID uint64) (*Executable, error)
	GetByText(ctx context.Context, text string) (*Executable, error)
	GetLatestByTenantID(ctx context.Context, tenantID primitive.ObjectID) (*Executable, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *Executable) error
	ListByFilter(ctx context.Context, f *ExecutablePaginationListFilter) (*ExecutablePaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *ExecutablePaginationListFilter) ([]*ExecutableAsSelectOption, error)
	ListByTenantID(ctx context.Context, tid primitive.ObjectID) (*ExecutablePaginationListResult, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type ExecutableStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) ExecutableStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("executables")

	_, err := uc.Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "tenant_id", Value: 1}}},
		{Keys: bson.D{{Key: "public_id", Value: -1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{
			{"program_name", "text"},
		}},
	})
	if err != nil {
		// It is important that we crash the app on startup to meet the
		// requirements of `google/wire` framework.
		log.Fatal(err)
	}

	s := &ExecutableStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
