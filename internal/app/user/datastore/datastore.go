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
	UserStatusActive   = 1
	UserStatusArchived = 2

	UserRoleExecutive      = 1
	UserRoleManagement     = 2
	UserRoleFrontlineStaff = 3
	UserRoleStaff          = 3
	UserRoleAssociate      = 4
	UserRoleCustomer       = 5

	UserHasCouponPromotionalCodeUnassigned = 0
	UserHasCouponPromotionalCodeYes        = 1
	UserHasCouponPromotionalCodeNo         = 2
)

type User struct {
	ID                          primitive.ObjectID `bson:"_id" json:"id"`
	Email                       string             `bson:"email" json:"email"`
	FirstName                   string             `bson:"first_name" json:"first_name"`
	LastName                    string             `bson:"last_name" json:"last_name"`
	Name                        string             `bson:"name" json:"name"`
	LexicalName                 string             `bson:"lexical_name" json:"lexical_name"`
	TenantName                  string             `bson:"tenant_name" json:"tenant_name"`
	TenantType                  int8               `bson:"tenant_type" json:"tenant_type"`
	TenantID                    primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	PasswordHashAlgorithm       string             `bson:"password_hash_algorithm" json:"password_hash_algorithm,omitempty"`
	PasswordHash                string             `bson:"password_hash" json:"password_hash,omitempty"`
	Role                        int8               `bson:"role" json:"role"`
	HasStaffRole                bool               `bson:"has_staff_role" json:"has_staff_role"`
	WasEmailVerified            bool               `bson:"was_email_verified" json:"was_email_verified"`
	EmailVerificationCode       string             `bson:"email_verification_code,omitempty" json:"email_verification_code,omitempty"`
	EmailVerificationExpiry     time.Time          `bson:"email_verification_expiry,omitempty" json:"email_verification_expiry,omitempty"`
	Phone                       string             `bson:"phone" json:"phone,omitempty"`
	Country                     string             `bson:"country" json:"country,omitempty"`
	Region                      string             `bson:"region" json:"region,omitempty"`
	City                        string             `bson:"city" json:"city,omitempty"`
	PostalCode                  string             `bson:"postal_code" json:"postal_code,omitempty"`
	AddressLine1                string             `bson:"address_line1" json:"address_line1,omitempty"`
	AddressLine2                string             `bson:"address_line2" json:"address_line2,omitempty"`
	HasShippingAddress          bool               `bson:"has_shipping_address" json:"has_shipping_address,omitempty"`
	ShippingName                string             `bson:"shipping_name" json:"shipping_name,omitempty"`
	ShippingPhone               string             `bson:"shipping_phone" json:"shipping_phone,omitempty"`
	ShippingCountry             string             `bson:"shipping_country" json:"shipping_country,omitempty"`
	ShippingRegion              string             `bson:"shipping_region" json:"shipping_region,omitempty"`
	ShippingCity                string             `bson:"shipping_city" json:"shipping_city,omitempty"`
	ShippingPostalCode          string             `bson:"shipping_postal_code" json:"shipping_postal_code,omitempty"`
	ShippingAddressLine1        string             `bson:"shipping_address_line1" json:"shipping_address_line1,omitempty"`
	ShippingAddressLine2        string             `bson:"shipping_address_line2" json:"shipping_address_line2,omitempty"`
	HowDidYouHearAboutUsID      primitive.ObjectID `bson:"how_did_you_hear_about_us_id" json:"how_did_you_hear_about_us_id,omitempty"`
	HowDidYouHearAboutUsText    string             `bson:"how_did_you_hear_about_us_text" json:"how_did_you_hear_about_us_text,omitempty"`
	IsHowDidYouHearAboutUsOther bool               `bson:"is_how_did_you_hear_about_us_other" json:"is_how_did_you_hear_about_us_other,omitempty"`
	HowDidYouHearAboutUsOther   string             `bson:"how_did_you_hear_about_us_other" json:"how_did_you_hear_about_us_other,omitempty"`
	AgreeTOS                    bool               `bson:"agree_tos" json:"agree_tos,omitempty"`
	TOSVersion                  string             `bson:"tos_version" json:"tos_version,omitempty"`
	TOSText                     string             `bson:"tos_text" json:"tos_text,omitempty"`
	TOSAgreedOn                 time.Time          `bson:"tos_agreed_on" json:"tos_agreed_on,omitempty"`
	PrivacyVersion              string             `bson:"privacy_version" json:"privacy_version,omitempty"`
	PrivacyText                 string             `bson:"privacy_text" json:"privacy_text,omitempty"`
	PrivacyAgreedOn             time.Time          `bson:"privacy_agreed_on" json:"privacy_agreed_on,omitempty"`
	AgreePromotionsEmail        bool               `bson:"agree_promotions_email" json:"agree_promotions_email,omitempty"`
	AgreeWaiver                 bool               `bson:"agree_waiver" json:"agree_waiver,omitempty"`
	WaiverText                  string             `bson:"waiver_text" json:"waiver_text,omitempty"`
	WaiverAgreedOn              time.Time          `bson:"waiver_agreed_on" json:"waiver_agreed_on,omitempty"`
	CreatedAt                   time.Time          `bson:"created_at" json:"created_at"`
	CreatedByUserID             primitive.ObjectID `bson:"created_by_user_id" json:"created_by_user_id,omitempty"`
	CreatedByUserName           string             `bson:"created_by_user_name" json:"created_by_user_name"`
	CreatedFromIPAddress        string             `bson:"created_from_ip_address" json:"created_from_ip_address"`
	ModifiedAt                  time.Time          `bson:"modified_at" json:"modified_at"`
	ModifiedByUserID            primitive.ObjectID `bson:"modified_by_user_id" json:"modified_by_user_id,omitempty"`
	ModifiedByUserName          string             `bson:"modified_by_user_name" json:"modified_by_user_name"`
	ModifiedFromIPAddress       string             `bson:"modified_from_ip_address" json:"modified_from_ip_address"`
	Status                      int8               `bson:"status" json:"status"`
	Comments                    []*UserComment     `bson:"comments" json:"comments"`
	Salt                        string             `bson:"salt" json:"salt,omitempty"`
	JoinedTime                  time.Time          `bson:"joined_time" json:"joined_time,omitempty"`
	PrAccessCode                string             `bson:"pr_access_code" json:"pr_access_code,omitempty"`
	PrExpiryTime                time.Time          `bson:"pr_expiry_time" json:"pr_expiry_time,omitempty"`
	PublicID                    uint64             `bson:"public_id" json:"public_id,omitempty"`
	Timezone                    string             `bson:"timezone" json:"timezone,omitempty"`
	// AccessToken       string             `bson:"access_token" json:"access_token,omitempty"`
	// RefreshToken      string             `bson:"refresh_token" json:"refresh_token,omitempty"`
	HasCouponPromotionalCode int8                 `bson:"has_coupon_promotional_code" json:"has_coupon_promotional_code"`
	CouponID                 primitive.ObjectID   `bson:"coupon_id" json:"coupon_id,omitempty"`
	CouponPromotionalCode    string               `bson:"coupon_promotional_code" json:"coupon_promotional_code,omitempty"`
	CouponName               string               `bson:"coupon_name" json:"coupon_name"`
	CouponDescription        string               `bson:"coupon_description" json:"coupon_description"`
	CouponHasExpirationDate  bool                 `bson:"coupon_has_expiration_date" json:"coupon_has_expiration_date"`
	CouponExpirationDate     time.Time            `bson:"coupon_expiration_date,omitempty" json:"coupon_expiration_date,omitempty"`
	CouponClaimedAt          time.Time            `bson:"coupon_claimed_at" json:"coupon_claimed_at,omitempty"`
	Coupons                  []*UserClaimedCoupon `bson:"coupons" json:"coupons,omitempty"`

	// The name of the payment processor we are using to handle payments with
	// this particular member.
	PaymentProcessorName string `bson:"payment_processor_name" json:"payment_processor_name"`

	// The unique identifier used by the payment processor which has a somesort of
	// copy of this member's details saved and we can reference that customer on
	// the payment processor using this `customer_id`.
	PaymentProcessorCustomerID string `bson:"payment_processor_customer_id" json:"payment_processor_customer_id"`
}

type UserComment struct {
	ID               primitive.ObjectID `bson:"_id" json:"id"`
	TenantID         primitive.ObjectID `bson:"tenant_id" json:"tenant_id"`
	CreatedAt        time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty"`
	CreatedByUserID  primitive.ObjectID `bson:"created_by_user_id" json:"created_by_user_id"`
	CreatedByName    string             `bson:"created_by_name" json:"created_by_name"`
	ModifiedAt       time.Time          `bson:"modified_at,omitempty" json:"modified_at,omitempty"`
	ModifiedByUserID primitive.ObjectID `bson:"modified_by_user_id" json:"modified_by_user_id"`
	ModifiedByName   string             `bson:"modified_by_name" json:"modified_by_name"`
	Content          string             `bson:"content" json:"content"`
}

type UserClaimedCoupon struct {
	ID                 primitive.ObjectID `bson:"_id" json:"id"`
	Name               string             `bson:"name" json:"name"`
	Description        string             `bson:"description" json:"description"`
	Status             int8               `bson:"status" json:"status"`
	BusinessFunction   int8               `bson:"business_function" json:"business_function"`
	ClassLimit         int                `bson:"class_limit" json:"class_limit"`
	DaysLimit          int                `bson:"days_limit" json:"days_limit"`
	DiscountType       int8               `bson:"discount_type" json:"discount_type"`
	PercentageDiscount int                `bson:"percentage_discount" json:"percentage_discount"`
	HasExpirationDate  bool               `bson:"has_expiration_date" json:"has_expiration_date"`
	ExpirationDate     time.Time          `bson:"expiration_date,omitempty" json:"expiration_date,omitempty"`
	ClaimedAt          time.Time          `bson:"claimed_at,omitempty" json:"claimed_at,omitempty"`
	PromotionalCode    string             `bson:"promotional_code" json:"promotional_code"`

	// The ID of the payment processor we are using to handle payments with
	// this particular coupon.
	PaymentProcessorID int8 `bson:"payment_processor_id" json:"payment_processor_id"`

	// The unique identifier used by the payment processor which has a somesort of
	// copy of this coupons's details saved and we can reference that couopn on
	// the payment processor using this `coupon_id`.
	PaymentProcessorCouponID string `bson:"payment_processor_coupon_id" json:"payment_processor_coupon_id"`
}

type UserListFilter struct {
	// Pagination related.
	Cursor    primitive.ObjectID
	PageSize  int64
	SortField string
	SortOrder int8 // 1=ascending | -1=descending

	// Filter related.
	TenantID        primitive.ObjectID
	Role            int8
	Status          int8
	ExcludeArchived bool
	SearchText      string
	FirstName       string
	LastName        string
	Email           string
	Phone           string
	CreatedAtGTE    time.Time
}

type UserListResult struct {
	Results     []*User            `json:"results"`
	NextCursor  primitive.ObjectID `json:"next_cursor"`
	HasNextPage bool               `json:"has_next_page"`
}

type UserAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"name" json:"label"`
}

// UserStorer Interface for user.
type UserStorer interface {
	Create(ctx context.Context, m *User) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*User, error)
	GetByPublicID(ctx context.Context, oldID uint64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByVerificationCode(ctx context.Context, verificationCode string) (*User, error)
	GetLatestByTenantID(ctx context.Context, tenantID primitive.ObjectID) (*User, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *User) error
	UpsertByID(ctx context.Context, m *User) error
	UpsertByEmail(ctx context.Context, m *User) error
	ListByFilter(ctx context.Context, f *UserListFilter) (*UserListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *UserListFilter) ([]*UserAsSelectOption, error)
	ListAllExecutives(ctx context.Context) (*UserListResult, error)
	ListAllStaffForTenantID(ctx context.Context, tenantID primitive.ObjectID) (*UserListResult, error)
	CountByFilter(ctx context.Context, f *UserListFilter) (int64, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type UserStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) UserStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("users")

	_, err := uc.Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "tenant_id", Value: 1}}},
		{Keys: bson.D{{Key: "email", Value: 1}}},
		{Keys: bson.D{{Key: "last_name", Value: 1}}},
		{Keys: bson.D{{Key: "name", Value: 1}}},
		{Keys: bson.D{{Key: "lexical_name", Value: 1}}},
		{Keys: bson.D{{Key: "public_id", Value: -1}}},
		{Keys: bson.D{{Key: "joined_time", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "type", Value: 1}}},
		{Keys: bson.D{
			{"name", "text"},
			{"lexical_name", "text"},
			{"email", "text"},
			{"phone", "text"},
			{"country", "text"},
			{"region", "text"},
			{"city", "text"},
			{"postal_code", "text"},
			{"address_line1", "text"},
			{"description", "text"},
		}},
	})
	if err != nil {
		// It is important that we crash the app on startup to meet the
		// requirements of `google/wire` framework.
		log.Fatal(err)
	}

	s := &UserStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
