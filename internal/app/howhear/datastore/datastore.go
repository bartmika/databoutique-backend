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
	HowHearAboutUsItemStatusActive   = 1
	HowHearAboutUsItemStatusArchived = 2
)

type HowHearAboutUsItem struct {
	ID                    primitive.ObjectID `bson:"_id" json:"id"`
	TenantID              primitive.ObjectID `bson:"tenant_id" json:"tenant_id"`
	Text                  string             `bson:"text" json:"text"`
	SortNumber            int8               `bson:"sort_number" json:"sort_number"`
	IsForAssociate        bool               `bson:"is_for_associate" json:"is_for_associate"`
	IsForCustomer         bool               `bson:"is_for_customer" json:"is_for_customer"`
	IsForStaff            bool               `bson:"is_for_staff" json:"is_for_staff"`
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

type HowHearAboutUsItemListResult struct {
	Results     []*HowHearAboutUsItem `json:"results"`
	NextCursor  primitive.ObjectID    `json:"next_cursor"`
	HasNextPage bool                  `json:"has_next_page"`
}

type HowHearAboutUsItemAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"text" json:"label"`
}

// HowHearAboutUsItemStorer Interface for user.
type HowHearAboutUsItemStorer interface {
	Create(ctx context.Context, m *HowHearAboutUsItem) error
	CreateOrGetByID(ctx context.Context, hh *HowHearAboutUsItem) (*HowHearAboutUsItem, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*HowHearAboutUsItem, error)
	GetByPublicID(ctx context.Context, oldID uint64) (*HowHearAboutUsItem, error)
	GetByText(ctx context.Context, text string) (*HowHearAboutUsItem, error)
	GetLatestByTenantID(ctx context.Context, tenantID primitive.ObjectID) (*HowHearAboutUsItem, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *HowHearAboutUsItem) error
	ListByFilter(ctx context.Context, f *HowHearAboutUsItemPaginationListFilter) (*HowHearAboutUsItemPaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *HowHearAboutUsItemPaginationListFilter) ([]*HowHearAboutUsItemAsSelectOption, error)
	ListByTenantID(ctx context.Context, tid primitive.ObjectID) (*HowHearAboutUsItemPaginationListResult, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type HowHearAboutUsItemStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) HowHearAboutUsItemStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("how_hear_about_us_items")

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

	s := &HowHearAboutUsItemStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
