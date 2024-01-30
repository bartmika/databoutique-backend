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
	AssistantMessageStatusActive   = 1
	AssistantMessageStatusQueued   = 2
	AssistantMessageStatusError    = 3
	AssistantMessageStatusArchived = 4
)

type AssistantMessage struct {
	ID                      primitive.ObjectID `bson:"_id" json:"id"`
	TenantID                primitive.ObjectID `bson:"tenant_id" json:"tenant_id"`
	AssistantID             primitive.ObjectID `bson:"assistant_id" json:"assistant_id"`
	AssistantThreadID       primitive.ObjectID `bson:"assistant_thread_id" json:"assistant_thread_id"`
	OpenAIAssistantID       string             `bson:"openai_assistant_id" json:"openai_assistant_id"`
	OpenAIAssistantThreadID string             `bson:"openai_assistant_thread_id" json:"openai_assistant_thread_id"`
	Text                    string             `bson:"text" json:"text"`
	Status                  int8               `bson:"status" json:"status"`
	PublicID                uint64             `bson:"public_id" json:"public_id"`
	FromAssistant           bool               `bson:"from_assistant" json:"from_assistant"`
	CreatedAt               time.Time          `bson:"created_at" json:"created_at"`
	CreatedByUserID         primitive.ObjectID `bson:"created_by_user_id" json:"created_by_user_id,omitempty"`
	CreatedByUserName       string             `bson:"created_by_user_name" json:"created_by_user_name"`
	CreatedFromIPAddress    string             `bson:"created_from_ip_address" json:"created_from_ip_address"`
	ModifiedAt              time.Time          `bson:"modified_at" json:"modified_at"`
	ModifiedByUserID        primitive.ObjectID `bson:"modified_by_user_id" json:"modified_by_user_id,omitempty"`
	ModifiedByUserName      string             `bson:"modified_by_user_name" json:"modified_by_user_name"`
	ModifiedFromIPAddress   string             `bson:"modified_from_ip_address" json:"modified_from_ip_address"`
}

type AssistantMessageListResult struct {
	Results     []*AssistantMessage `json:"results"`
	NextCursor  primitive.ObjectID  `json:"next_cursor"`
	HasNextPage bool                `json:"has_next_page"`
}

type AssistantMessageAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"text" json:"label"`
}

// AssistantMessageStorer Interface for user.
type AssistantMessageStorer interface {
	Create(ctx context.Context, m *AssistantMessage) error
	CreateOrGetByID(ctx context.Context, hh *AssistantMessage) (*AssistantMessage, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*AssistantMessage, error)
	GetByPublicID(ctx context.Context, oldID uint64) (*AssistantMessage, error)
	GetByText(ctx context.Context, text string) (*AssistantMessage, error)
	GetLatestByTenantID(ctx context.Context, tenantID primitive.ObjectID) (*AssistantMessage, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *AssistantMessage) error
	ListByFilter(ctx context.Context, f *AssistantMessagePaginationListFilter) (*AssistantMessagePaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *AssistantMessagePaginationListFilter) ([]*AssistantMessageAsSelectOption, error)
	ListByTenantID(ctx context.Context, tid primitive.ObjectID) (*AssistantMessagePaginationListResult, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type AssistantMessageStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) AssistantMessageStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("assistant_messages")

	_, err := uc.Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "tenant_id", Value: 1}}},
		{Keys: bson.D{{Key: "public_id", Value: -1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{
			{"text", "text"},
		}},
	})
	if err != nil {
		// It is important that we crash the app on startup to meet the
		// requirements of `google/wire` framework.
		log.Fatal(err)
	}

	s := &AssistantMessageStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
