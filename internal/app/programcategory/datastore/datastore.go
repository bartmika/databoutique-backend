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
	ProgramCategoryStatusActive   = 1
	ProgramCategoryStatusArchived = 2
)

type ProgramCategory struct {
	ID                    primitive.ObjectID `bson:"_id" json:"id"`
	TenantID              primitive.ObjectID `bson:"tenant_id" json:"tenant_id"`
	Name                  string             `bson:"name" json:"name"`
	Description           string             `bson:"description" json:"description"`
	SortNumber            int8               `bson:"sort_number" json:"sort_number"`
	Status                int8               `bson:"status" json:"status"`
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

type ProgramCategoryListResult struct {
	Results     []*ProgramCategory `json:"results"`
	NextCursor  primitive.ObjectID `json:"next_cursor"`
	HasNextPage bool               `json:"has_next_page"`
}

type ProgramCategoryAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"name" json:"label"`
}

// ProgramCategoryStorer Interface for user.
type ProgramCategoryStorer interface {
	Create(ctx context.Context, m *ProgramCategory) error
	CreateOrGetByID(ctx context.Context, hh *ProgramCategory) (*ProgramCategory, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*ProgramCategory, error)
	GetByPublicID(ctx context.Context, oldID uint64) (*ProgramCategory, error)
	GetByName(ctx context.Context, name string) (*ProgramCategory, error)
	GetLatestByTenantID(ctx context.Context, tenantID primitive.ObjectID) (*ProgramCategory, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *ProgramCategory) error
	ListByFilter(ctx context.Context, f *ProgramCategoryPaginationListFilter) (*ProgramCategoryPaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *ProgramCategoryPaginationListFilter) ([]*ProgramCategoryAsSelectOption, error)
	ListByTenantID(ctx context.Context, tid primitive.ObjectID) (*ProgramCategoryPaginationListResult, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type ProgramCategoryStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) ProgramCategoryStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("program_categories")

	_, err := uc.Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "tenant_id", Value: 1}}},
		{Keys: bson.D{{Key: "public_id", Value: -1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{
			{"name", "text"},
		}},
	})
	if err != nil {
		// It is important that we crash the app on startup to meet the
		// requirements of `google/wire` framework.
		log.Fatal(err)
	}

	s := &ProgramCategoryStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
