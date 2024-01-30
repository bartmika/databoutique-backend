package templatedemailer

import (
	mg "github.com/bartmika/databoutique-backend/internal/adapter/emailer/mailgun"
	"log/slog"

	c "github.com/bartmika/databoutique-backend/internal/config"
	"github.com/bartmika/databoutique-backend/internal/provider/uuid"
)

// TemplatedEmailer Is adapter for responsive HTML email templates sender.
type TemplatedEmailer interface {
	SendNewUserTemporaryPasswordEmail(email, firstName, temporaryPassword string) error
	SendVerificationEmail(email, verificationCode, firstName string) error
	SendForgotPasswordEmail(email, verificationCode, firstName string) error
}

type templatedEmailer struct {
	UUID    uuid.Provider
	Logger  *slog.Logger
	Emailer mg.Emailer
}

func NewTemplatedEmailer(cfg *c.Conf, logger *slog.Logger, uuidp uuid.Provider, emailer mg.Emailer) TemplatedEmailer {
	// Defensive code: Make sure we have access to the file before proceeding any further with the code.
	logger.Debug("templated emailer initializing...")
	logger.Debug("templated emailer initialized")

	return &templatedEmailer{
		UUID:    uuidp,
		Logger:  logger,
		Emailer: emailer,
	}
}
