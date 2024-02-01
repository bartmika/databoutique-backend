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
	UploadDirectoryStatusActive   = 1
	UploadDirectoryStatusArchived = 2
)

type UploadDirectory struct {
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

type UploadDirectoryListResult struct {
	Results     []*UploadDirectory `json:"results"`
	NextCursor  primitive.ObjectID `json:"next_cursor"`
	HasNextPage bool               `json:"has_next_page"`
}

type UploadDirectoryAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"name" json:"label"`
}

// UploadDirectoryStorer Interface for user.
type UploadDirectoryStorer interface {
	Create(ctx context.Context, m *UploadDirectory) error
	CreateOrGetByID(ctx context.Context, hh *UploadDirectory) (*UploadDirectory, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*UploadDirectory, error)
	GetByPublicID(ctx context.Context, oldID uint64) (*UploadDirectory, error)
	GetByText(ctx context.Context, text string) (*UploadDirectory, error)
	GetLatestByTenantID(ctx context.Context, tenantID primitive.ObjectID) (*UploadDirectory, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *UploadDirectory) error
	ListByFilter(ctx context.Context, f *UploadDirectoryPaginationListFilter) (*UploadDirectoryPaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *UploadDirectoryPaginationListFilter) ([]*UploadDirectoryAsSelectOption, error)
	ListByTenantID(ctx context.Context, tid primitive.ObjectID) (*UploadDirectoryPaginationListResult, error)
	ListByIDs(ctx context.Context, ids []primitive.ObjectID) (*UploadDirectoryPaginationListResult, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type UploadDirectoryStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) UploadDirectoryStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("upload_directories")
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

	s := &UploadDirectoryStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
