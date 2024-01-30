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
	AssistantStatusActive   = 1
	AssistantStatusArchived = 2
)

type Assistant struct {
	// The name of the assistant to be available to administators in-app and on OpenAI dashboard.
	Name string `bson:"name" json:"name"`
	// Description to include for this assistant, used only by app
	Description string `bson:"description" json:"description"`
	// The OpenAI prompt to utilize for this assistant.
	Instructions string `bson:"instructions" json:"instructions"`
	// Model holds the OpenAI completion model to utilize, to see list then visit: https://github.com/sashabaranov/go-openai/blob/eff8dc1118ea82a1b50ee316608e24d83df74d6b/completion.go
	Model string `bson:"model" json:"model"`
	// Variable controls how this assistant is running in-app.
	Status int8 `bson:"status" json:"status"`
	// The unique identifier used in-app powered by MongoDB.
	ID                    primitive.ObjectID     `bson:"_id" json:"id"`
	TenantID              primitive.ObjectID     `bson:"tenant_id" json:"tenant_id,omitempty"`
	TenantName            string                 `bson:"tenant_name" json:"tenant_name,omitempty"`
	PublicID              uint64                 `bson:"public_id" json:"public_id"`
	CreatedAt             time.Time              `bson:"created_at" json:"created_at"`
	CreatedByUserID       primitive.ObjectID     `bson:"created_by_user_id" json:"created_by_user_id,omitempty"`
	CreatedByUserName     string                 `bson:"created_by_user_name" json:"created_by_user_name"`
	CreatedFromIPAddress  string                 `bson:"created_from_ip_address" json:"created_from_ip_address"`
	ModifiedAt            time.Time              `bson:"modified_at" json:"modified_at"`
	ModifiedByUserID      primitive.ObjectID     `bson:"modified_by_user_id" json:"modified_by_user_id,omitempty"`
	ModifiedByUserName    string                 `bson:"modified_by_user_name" json:"modified_by_user_name"`
	ModifiedFromIPAddress string                 `bson:"modified_from_ip_address" json:"modified_from_ip_address"`
	AssistantFiles        []*AssistantFileOption `bson:"assistant_files" json:"assistant_files,omitempty"`
	OpenAIAssistantID     string                 `bson:"openai_assistant_id" json:"openai_assistant_id"` // https://platform.openai.com/docs/assistants/tools/supported-files
}

// GetAssistantFileIDs function will iterate through all the assistant files
// and return the file ID values.
func (a *Assistant) GetAssistantFileIDs() []string {
	if a.AssistantFiles == nil {
		return nil
	}
	ids := []string{}
	for _, af := range a.AssistantFiles {
		ids = append(ids, af.OpenAIFileID)
	}
	return ids
}

type AssistantFileOption struct {
	ID           primitive.ObjectID `bson:"_id" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Description  string             `bson:"description" json:"description"`
	OpenAIFileID string             `bson:"openai_file_id" json:"openai_file_id"` // https://platform.openai.com/docs/assistants/tools/supported-files
	Status       int8               `bson:"status" json:"status"`
}

type AssistantListFilter struct {
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

type AssistantListResult struct {
	Results     []*Assistant       `json:"results"`
	NextCursor  primitive.ObjectID `json:"next_cursor"`
	HasNextPage bool               `json:"has_next_page"`
}

type AssistantAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"name" json:"label"`
}

// AssistantStorer Interface for user.
type AssistantStorer interface {
	Create(ctx context.Context, m *Assistant) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*Assistant, error)
	GetByPublicID(ctx context.Context, oldID uint64) (*Assistant, error)
	GetByEmail(ctx context.Context, email string) (*Assistant, error)
	GetByVerificationCode(ctx context.Context, verificationCode string) (*Assistant, error)
	GetLatestByTenantID(ctx context.Context, tenantID primitive.ObjectID) (*Assistant, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *Assistant) error
	ListByFilter(ctx context.Context, f *AssistantPaginationListFilter) (*AssistantPaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *AssistantListFilter) ([]*AssistantAsSelectOption, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type AssistantStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) AssistantStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("assistants")

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

	s := &AssistantStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
