package templatedemailer

import (
	"bytes"
	"context"
	"path"
	"text/template"

	"log/slog"
)

func (impl *templatedEmailer) SendVerificationEmail(email, verificationCode, firstName string) error {
	impl.Logger.Debug("sending email verification email...")

	// FOR TESTING PURPOSES ONLY.
	fp := path.Join("templates", "verification_email.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		impl.Logger.Error("parsing error", slog.Any("error", err))
		return err
	}

	var processed bytes.Buffer

	// Render the HTML template with our data.
	data := struct {
		Email            string
		VerificationLink string
		FirstName        string
	}{
		Email:            email,
		VerificationLink: "https://" + impl.Emailer.GetDomainName() + "/verify?q=" + verificationCode,
		FirstName:        firstName,
	}
	if err := tmpl.Execute(&processed, data); err != nil {
		impl.Logger.Error("template execution error", slog.Any("error", err))
		return err
	}
	body := processed.String() // DEVELOPERS NOTE: Convert our long sequence of data into a string.

	if err := impl.Emailer.Send(context.Background(), impl.Emailer.GetSenderEmail(), "Activate your CPS Retail Partner Account", email, body); err != nil {
		impl.Logger.Error("sending error", slog.Any("error", err))
		return err
	}
	impl.Logger.Debug("email verification email sent")
	return nil
}
