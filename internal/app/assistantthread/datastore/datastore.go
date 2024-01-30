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
	AssistantThreadStatusActive   = 1
	AssistantThreadStatusQueued   = 2
	AssistantThreadStatusError    = 3
	AssistantThreadStatusArchived = 4
)

type AssistantThread struct {
	// The assistant we are referencing in MongoDB
	AssistantID primitive.ObjectID `bson:"assistant_id" json:"assistant_id"`
	// The copy of the name of the assistant.
	AssistantName string `bson:"assistant_name" json:"assistant_name"`
	// The copy of the description of the assistant.
	AssistantDescription string `bson:"assistant_description" json:"assistant_description"`
	// The copy of the of the OpenAI `assistant_id` from the assistant record.
	OpenAIAssistantID string `bson:"openai_assistant_id" json:"openai_assistant_id"`
	// Variable controls how this thread is running in-app.
	Status int8 `bson:"status" json:"status"`
	// The unique id provided by OpenAI for our thread.
	OpenAIAssistantThreadID string `bson:"openai_assistantthread_id" json:"openai_assistantthread_id"`
	// The unique identifier used in-app powered by MongoDB.
	ID         primitive.ObjectID `bson:"_id" json:"id"`
	TenantID   primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	TenantName string             `bson:"tenant_name" json:"tenant_name,omitempty"`

	UserID          primitive.ObjectID `bson:"user_id" json:"user_id"`
	UserPublicID    uint64             `bson:"user_public_id" json:"user_public_id"`
	UserFirstName   string             `bson:"user_first_name" json:"user_first_name,omitempty"`
	UserLastName    string             `bson:"user_last_name" json:"user_last_name,omitempty"`
	UserName        string             `bson:"user_name" json:"user_name,omitempty"`
	UserLexicalName string             `bson:"user_lexical_name" json:"user_lexical_name,omitempty"`

	PublicID              uint64             `bson:"public_id" json:"public_id"`
	CreatedAt             time.Time          `bson:"created_at" json:"created_at"`
	CreatedByUserID       primitive.ObjectID `bson:"created_by_user_id" json:"created_by_user_id,omitempty"`
	CreatedByUserName     string             `bson:"created_by_user_name" json:"created_by_user_name"`
	CreatedFromIPAddress  string             `bson:"created_from_ip_address" json:"created_from_ip_address"`
	ModifiedAt            time.Time          `bson:"modified_at" json:"modified_at"`
	ModifiedByUserID      primitive.ObjectID `bson:"modified_by_user_id" json:"modified_by_user_id,omitempty"`
	ModifiedByUserName    string             `bson:"modified_by_user_name" json:"modified_by_user_name"`
	ModifiedFromIPAddress string             `bson:"modified_from_ip_address" json:"modified_from_ip_address"`
}

type AssistantThreadFileOption struct {
	ID           primitive.ObjectID `bson:"_id" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Description  string             `bson:"description" json:"description"`
	OpenAIFileID string             `bson:"openai_file_id" json:"openai_file_id"` // https://platform.openai.com/docs/assistantthreads/tools/supported-files
	Status       int8               `bson:"status" json:"status"`
}

type AssistantThreadListFilter struct {
	// Pagination related.
	Cursor    primitive.ObjectID
	PageSize  int64
	SortField string
	SortOrder int8 // 1=ascending | -1=descending

	// Filter related.
	TenantID        primitive.ObjectID
	Status          int8
	ExcludeArchived bool
	SearchText      string
}

type AssistantThreadListResult struct {
	Results     []*AssistantThread `json:"results"`
	NextCursor  primitive.ObjectID `json:"next_cursor"`
	HasNextPage bool               `json:"has_next_page"`
}

type AssistantThreadAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"name" json:"label"`
}

// AssistantThreadStorer Interface for user.
type AssistantThreadStorer interface {
	Create(ctx context.Context, m *AssistantThread) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*AssistantThread, error)
	GetByPublicID(ctx context.Context, oldID uint64) (*AssistantThread, error)
	GetByEmail(ctx context.Context, email string) (*AssistantThread, error)
	GetByVerificationCode(ctx context.Context, verificationCode string) (*AssistantThread, error)
	GetLatestByTenantID(ctx context.Context, tenantID primitive.ObjectID) (*AssistantThread, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *AssistantThread) error
	ListByFilter(ctx context.Context, f *AssistantThreadPaginationListFilter) (*AssistantThreadPaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *AssistantThreadListFilter) ([]*AssistantThreadAsSelectOption, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type AssistantThreadStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) AssistantThreadStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("assistant_threads")

	// // For debugging purposes only.
	// if _, err := uc.Indexes().DropAll(context.TODO()); err != nil {
	// 	loggerp.Error("failed deleting all indexes",
	// 		slog.Any("err", err))
	//
	// 	// It is important that we crash the app on startup to meet the
	// 	// requirements of `google/wire` framework.
	// 	log.Fatal(err)
	// }

	_, err := uc.Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "tenant_id", Value: 1}}},
		{Keys: bson.D{{Key: "public_id", Value: -1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "name", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
		{Keys: bson.D{
			{"name", "text"},
			{"description", "text"},
		}},
	})
	if err != nil {
		// It is important that we crash the app on startup to meet the
		// requirements of `google/wire` framework.
		log.Fatal(err)
	}

	s := &AssistantThreadStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
