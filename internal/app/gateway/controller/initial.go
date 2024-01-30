package controller

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	hh_s "github.com/bartmika/databoutique-backend/internal/app/howhear/datastore"
	tenant_s "github.com/bartmika/databoutique-backend/internal/app/tenant/datastore"
	user_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
)

func (impl *GatewayControllerImpl) initializeAccounts(ctx context.Context) error {
	initialStore, err := impl.TenantStorer.GetByID(ctx, impl.Config.InitialAccount.AdminTenantID)
	if err != nil {
		return err
	}
	if initialStore == nil {
		impl.Logger.Debug("initializing accounts for first-time use...")

		// Use the user's provided time zone or default to UTC.
		location, _ := time.LoadLocation("UTC")

		//TODO: Implement saving into encrypted field for openai key.

		impl.Logger.Debug("initializing primary tenant")
		tenant := &tenant_s.Tenant{
			ID:           impl.Config.InitialAccount.AdminTenantID,
			Name:         impl.Config.InitialAccount.AdminTenantName,
			Timezone:     "UTC",
			Email:        impl.Config.InitialAccount.AdminEmail,
			CreatedAt:    time.Now().In(location),
			ModifiedAt:   time.Now().In(location),
			OpenAIAPIKey: impl.Config.InitialAccount.AdminTenantOpenAIAPIKey,
			OpenAIOrgKey: impl.Config.InitialAccount.AdminTenantOpenAIOrgKey,
		}
		if err := impl.TenantStorer.Create(ctx, tenant); err != nil {
			return err
		}

		impl.Logger.Debug("initializing primary executive administrator")

		passwordHash, err := impl.Password.GenerateHashFromPassword(impl.Config.InitialAccount.AdminPassword)
		if err != nil {
			impl.Logger.Error("hashing error", slog.Any("error", err))
			return err
		}

		admin := &user_s.User{
			ID:                    primitive.NewObjectID(),
			Email:                 impl.Config.InitialAccount.AdminEmail,
			FirstName:             "System",
			LastName:              "Administrator",
			Name:                  "System Administrator",
			LexicalName:           "Administrator, System",
			TenantName:            impl.Config.InitialAccount.AdminTenantName,
			TenantID:              tenant.ID,
			PasswordHash:          passwordHash,
			PasswordHashAlgorithm: impl.Password.AlgorithmName(),
			Role:                  user_s.UserRoleExecutive,
			WasEmailVerified:      true,
			CreatedAt:             time.Now().In(location),
			ModifiedAt:            time.Now().In(location),
			Status:                user_s.UserStatusActive,
			AgreeTOS:              true,
		}
		err = impl.UserStorer.Create(ctx, admin)
		if err != nil {
			impl.Logger.Error("database create error", slog.Any("error", err))
			return err
		}
		impl.Logger.Debug("executive user created.",
			slog.Any("_id", admin.ID),
			slog.String("name", admin.Name),
			slog.String("email", admin.Email),
			slog.String("password_hash_algorithm", admin.PasswordHashAlgorithm),
			slog.String("password_hash", admin.PasswordHash))
	}
	return nil
}

func (impl *GatewayControllerImpl) initializeHowHearAboutUsItems(ctx context.Context) error {
	initialStore, err := impl.TenantStorer.GetByID(ctx, impl.Config.InitialAccount.AdminTenantID)
	if err != nil {
		return err
	}

	oID1, _ := primitive.ObjectIDFromHex("65af37e16abda7495021e676")
	oID2, _ := primitive.ObjectIDFromHex("65aca72906b626524b3d1634")
	oID3, _ := primitive.ObjectIDFromHex("65ac92cba6a970bd138c6e2f")

	o1 := &hh_s.HowHearAboutUsItem{
		ID:         oID1,
		TenantID:   initialStore.ID,
		Text:       "Search Engine Results",
		SortNumber: 1,
		Status:     hh_s.HowHearAboutUsItemStatusActive,
		CreatedAt:  time.Now(),
		// CreatedByUserID       primitive.ObjectID `bson:"created_by_user_id" json:"created_by_user_id,omitempty"`
		// CreatedByUserName     string             `bson:"created_by_user_name" json:"created_by_user_name"`
		// CreatedFromIPAddress  string             `bson:"created_from_ip_address" json:"created_from_ip_address"`
		ModifiedAt: time.Now(),
		// ModifiedByUserID      primitive.ObjectID `bson:"modified_by_user_id" json:"modified_by_user_id,omitempty"`
		// ModifiedByUserName    string             `bson:"modified_by_user_name" json:"modified_by_user_name"`
		// ModifiedFromIPAddress string             `bson:"modified_from_ip_address" json:"modified_from_ip_address"`
	}
	if _, err := impl.HowHearAboutUsItemStorer.CreateOrGetByID(ctx, o1); err != nil {
		return err
	}

	o2 := &hh_s.HowHearAboutUsItem{
		ID:         oID2,
		TenantID:   initialStore.ID,
		Text:       "Friend",
		SortNumber: 2,
		Status:     hh_s.HowHearAboutUsItemStatusActive,
		CreatedAt:  time.Now(),
		// CreatedByUserID       primitive.ObjectID `bson:"created_by_user_id" json:"created_by_user_id,omitempty"`
		// CreatedByUserName     string             `bson:"created_by_user_name" json:"created_by_user_name"`
		// CreatedFromIPAddress  string             `bson:"created_from_ip_address" json:"created_from_ip_address"`
		ModifiedAt: time.Now(),
		// ModifiedByUserID      primitive.ObjectID `bson:"modified_by_user_id" json:"modified_by_user_id,omitempty"`
		// ModifiedByUserName    string             `bson:"modified_by_user_name" json:"modified_by_user_name"`
		// ModifiedFromIPAddress string             `bson:"modified_from_ip_address" json:"modified_from_ip_address"`
	}
	if _, err := impl.HowHearAboutUsItemStorer.CreateOrGetByID(ctx, o2); err != nil {
		return err
	}

	o3 := &hh_s.HowHearAboutUsItem{
		ID:         oID3,
		TenantID:   initialStore.ID,
		Text:       "Other",
		SortNumber: 99,
		Status:     hh_s.HowHearAboutUsItemStatusActive,
		CreatedAt:  time.Now(),
		// CreatedByUserID       primitive.ObjectID `bson:"created_by_user_id" json:"created_by_user_id,omitempty"`
		// CreatedByUserName     string             `bson:"created_by_user_name" json:"created_by_user_name"`
		// CreatedFromIPAddress  string             `bson:"created_from_ip_address" json:"created_from_ip_address"`
		ModifiedAt: time.Now(),
		// ModifiedByUserID      primitive.ObjectID `bson:"modified_by_user_id" json:"modified_by_user_id,omitempty"`
		// ModifiedByUserName    string             `bson:"modified_by_user_name" json:"modified_by_user_name"`
		// ModifiedFromIPAddress string             `bson:"modified_from_ip_address" json:"modified_from_ip_address"`
	}
	if _, err := impl.HowHearAboutUsItemStorer.CreateOrGetByID(ctx, o3); err != nil {
		return err
	}

	return nil
}
