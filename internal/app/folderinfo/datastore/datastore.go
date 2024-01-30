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
	FolderInfoStatusActive   = 1
	FolderInfoStatusArchived = 2
)

type FolderInfo struct {
	ID                    primitive.ObjectID `bson:"_id" json:"id"`
	TenantID              primitive.ObjectID `bson:"tenant_id" json:"tenant_id"`
	Text                  string             `bson:"text" json:"text"`
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

type FolderInfoListResult struct {
	Results     []*FolderInfo `json:"results"`
	NextCursor  primitive.ObjectID `json:"next_cursor"`
	HasNextPage bool               `json:"has_next_page"`
}

type FolderInfoAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"text" json:"label"`
}

// FolderInfoStorer Interface for user.
type FolderInfoStorer interface {
	Create(ctx context.Context, m *FolderInfo) error
	CreateOrGetByID(ctx context.Context, hh *FolderInfo) (*FolderInfo, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*FolderInfo, error)
	GetByPublicID(ctx context.Context, oldID uint64) (*FolderInfo, error)
	GetByText(ctx context.Context, text string) (*FolderInfo, error)
	GetLatestByTenantID(ctx context.Context, tenantID primitive.ObjectID) (*FolderInfo, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *FolderInfo) error
	ListByFilter(ctx context.Context, f *FolderInfoPaginationListFilter) (*FolderInfoPaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *FolderInfoPaginationListFilter) ([]*FolderInfoAsSelectOption, error)
	ListByTenantID(ctx context.Context, tid primitive.ObjectID) (*FolderInfoPaginationListResult, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type FolderInfoStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) FolderInfoStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("program_categories")

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

	s := &FolderInfoStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
