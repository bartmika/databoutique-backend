package templatedemailer

import (
	"bytes"
	"context"
	"path"
	"text/template"

	"log/slog"
)

func (impl *templatedEmailer) SendNewUserTemporaryPasswordEmail(email, firstName, temporaryPassword string) error {
	impl.Logger.Debug("sending new user temporary password email...")

	// FOR TESTING PURPOSES ONLY.
	fp := path.Join("templates", "user_temporary_password.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		impl.Logger.Error("parsing error", slog.Any("error", err))
		return err
	}

	var processed bytes.Buffer

	// Render the HTML template with our data.
	data := struct {
		Email             string
		FirstName         string
		TemporaryPassword string
		LoginURL          string
	}{
		Email:             email,
		FirstName:         firstName,
		TemporaryPassword: temporaryPassword,
		LoginURL:          "https://" + impl.Emailer.GetDomainName() + "/login",
	}
	if err := tmpl.Execute(&processed, data); err != nil {
		impl.Logger.Error("template execution error", slog.Any("error", err))
		return err
	}
	body := processed.String() // DEVELOPERS NOTE: Convert our long sequence of data into a string.

	if err := impl.Emailer.Send(context.Background(), impl.Emailer.GetSenderEmail(), "Welcome to your new account", email, body); err != nil {
		impl.Logger.Error("sending error", slog.Any("error", err))
		return err
	}
	impl.Logger.Debug("new user temporary password email sent")
	return nil
}
