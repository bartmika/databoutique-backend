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
	ProgramStatusActive                           = 1
	ProgramStatusArchived                         = 2
	ProgramBusinessFunctionCustomerDocumentReview = 1
	ProgramBusinessFunctionAdmintorDocumentReview = 2
)

type Program struct {
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
	SortNumber int8 `bson:"sort_number" json:"sort_number"`
	// The type of program being employed.
	BusinessFunction      int8               `bson:"business_function" json:"business_function"`
	ID                    primitive.ObjectID `bson:"_id" json:"id"`
	TenantID              primitive.ObjectID `bson:"tenant_id" json:"tenant_id"`
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

type ProgramListResult struct {
	Results     []*Program         `json:"results"`
	NextCursor  primitive.ObjectID `json:"next_cursor"`
	HasNextPage bool               `json:"has_next_page"`
}

type ProgramAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"text" json:"label"`
}

// ProgramStorer Interface for user.
type ProgramStorer interface {
	Create(ctx context.Context, m *Program) error
	CreateOrGetByID(ctx context.Context, hh *Program) (*Program, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*Program, error)
	GetByPublicID(ctx context.Context, oldID uint64) (*Program, error)
	GetByText(ctx context.Context, text string) (*Program, error)
	GetLatestByTenantID(ctx context.Context, tenantID primitive.ObjectID) (*Program, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *Program) error
	ListByFilter(ctx context.Context, f *ProgramPaginationListFilter) (*ProgramPaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *ProgramPaginationListFilter) ([]*ProgramAsSelectOption, error)
	ListByTenantID(ctx context.Context, tid primitive.ObjectID) (*ProgramPaginationListResult, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type ProgramStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) ProgramStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("programs")

	_, err := uc.Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "tenant_id", Value: 1}}},
		{Keys: bson.D{{Key: "public_id", Value: -1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{
			{"name", "text"},
			{"description", "text"},
			{"instructions", "text"},
		}},
	})
	if err != nil {
		// It is important that we crash the app on startup to meet the
		// requirements of `google/wire` framework.
		log.Fatal(err)
	}

	s := &ProgramStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
