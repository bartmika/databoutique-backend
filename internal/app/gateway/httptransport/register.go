package httptransport

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	gateway_c "github.com/bartmika/databoutique-backend/internal/app/gateway/controller"
	"github.com/bartmika/databoutique-backend/internal/utils/httperror"
)

func (h *Handler) unmarshalUserRegisterRequest(ctx context.Context, r *http.Request) (*gateway_c.UserRegisterRequestIDO, error) {
	// Initialize our array which will store all the results from the remote server.
	var requestData gateway_c.UserRegisterRequestIDO

	defer r.Body.Close()

	var rawJSON bytes.Buffer
	teeReader := io.TeeReader(r.Body, &rawJSON) // TeeReader allows you to read the JSON and capture it

	// Read the JSON string and convert it into our golang stuct else we need
	// to send a `400 Bad Request` errror message back to the client,
	err := json.NewDecoder(teeReader).Decode(&requestData) // [1]
	if err != nil {
		h.Logger.Error("decoding error",
			slog.Any("err", err),
			slog.String("json", rawJSON.String()))
		return nil, httperror.NewForSingleField(http.StatusBadRequest, "non_field_error", "payload structure is wrong")
	}

	// Defensive Code: For security purposes we need to remove all whitespaces from the email and lower the characters.
	requestData.Email = strings.ToLower(requestData.Email)
	requestData.Email = strings.ReplaceAll(requestData.Email, " ", "")
	requestData.EmailRepeated = strings.ToLower(requestData.EmailRepeated)
	requestData.EmailRepeated = strings.ReplaceAll(requestData.EmailRepeated, " ", "")

	// // BirthDate
	//
	// if requestData.BirthDate != "" {
	// 	birthDateDT, err := iso8601.ParseString(requestData.BirthDate)
	// 	if err != nil {
	// 		h.Logger.Error("iso8601 parsing error",
	// 			slog.Any("err", err),
	// 			slog.String("BirthDate", requestData.BirthDate),
	// 			slog.String("json", rawJSON.String()),
	// 		)
	// 		return nil, httperror.NewForSingleField(http.StatusBadRequest, "birth_date", "payload structure is wrong")
	// 	}
	// 	requestData.BirthDateDT = birthDateDT
	// }

	return &requestData, nil
}

func (h *Handler) UserRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := h.unmarshalUserRegisterRequest(ctx, r)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	res, err := h.Controller.UserRegister(ctx, data)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}
	MarshalLoginResponse(res, w)
}
