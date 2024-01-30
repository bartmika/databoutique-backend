package controller

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"

	user_s "github.com/bartmika/databoutique-backend/internal/app/user/datastore"
	"github.com/bartmika/databoutique-backend/internal/config/constants"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func (c *UserControllerImpl) CreateComment(ctx context.Context, customerID primitive.ObjectID, content string) (*user_s.User, error) {
	// Fetch the original customer.
	s, err := c.UserStorer.GetByID(ctx, customerID)
	if err != nil {
		c.Logger.Error("database get by id error", slog.Any("error", err))
		return nil, err
	}
	if s == nil {
		c.Logger.Error("user does not exist error",
			slog.Any("userid", customerID))
		return nil, httperror.NewForBadRequestWithSingleField("message", "user does not exist")
	}

	// Create our comment.
	comment := &user_s.UserComment{
		ID:               primitive.NewObjectID(),
		Content:          content,
		TenantID:         ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID),
		CreatedByUserID:  ctx.Value(constants.SessionUserID).(primitive.ObjectID),
		CreatedByName:    ctx.Value(constants.SessionUserName).(string),
		CreatedAt:        time.Now(),
		ModifiedByUserID: ctx.Value(constants.SessionUserID).(primitive.ObjectID),
		ModifiedByName:   ctx.Value(constants.SessionUserName).(string),
		ModifiedAt:       time.Now(),
	}

	// Add our comment to the comments.
	s.ModifiedByUserID = ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	s.ModifiedAt = time.Now()
	s.Comments = append(s.Comments, comment)

	// Save to the database the modified customer.
	if err := c.UserStorer.UpdateByID(ctx, s); err != nil {
		c.Logger.Error("database update by id error", slog.Any("error", err))
		return nil, err
	}

	return s, nil
}
