package helper

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// errorResponse is the uniform error body: Code is a stable snake_case key, Details is structured only for two codes.
// See docs/handbuch.md §6.2 (Fehlerformat).
type errorResponse struct {
	Code    string `json:"code"`
	Details any    `json:"details,omitempty"`
}

func SendJSONResponse(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("Failed to encode JSON response")
	}
}

func SendResponse(w http.ResponseWriter, response any) {
	SendJSONResponse(w, response, http.StatusOK)
}

func SendEmptyResponse(w http.ResponseWriter) {
	SendJSONResponse(w, struct{}{}, http.StatusOK)
}

func SendClientError(w http.ResponseWriter, code string, details any) {
	SendJSONResponse(w, errorResponse{Code: code, Details: details}, http.StatusBadRequest)
}

func SendConflictError(w http.ResponseWriter) {
	SendConflict(w, "conflict")
}

func SendNotFound(w http.ResponseWriter, code string) {
	SendJSONResponse(w, errorResponse{Code: code}, http.StatusNotFound)
}

// The frontend logs the user out and redirects to the login page on 401.
func SendUnauthorized(w http.ResponseWriter, code string) {
	SendJSONResponse(w, errorResponse{Code: code}, http.StatusUnauthorized)
}

// SendForbidden is for an authenticated user whose role lacks permission; the
// frontend keeps the session (auto-logout is bound to 401).
func SendForbidden(w http.ResponseWriter, code string, details any) {
	SendJSONResponse(w, errorResponse{Code: code, Details: details}, http.StatusForbidden)
}

func SendConflict(w http.ResponseWriter, code string) {
	SendJSONResponse(w, errorResponse{Code: code}, http.StatusConflict)
}

func SendTooManyRequests(w http.ResponseWriter, code string) {
	SendJSONResponse(w, errorResponse{Code: code}, http.StatusTooManyRequests)
}

func SendConflictDetails(w http.ResponseWriter, code string, details any) {
	SendJSONResponse(w, errorResponse{Code: code, Details: details}, http.StatusConflict)
}

func SendServerError(w http.ResponseWriter) {
	SendJSONResponse(w, errorResponse{Code: "internal_server_error"}, http.StatusInternalServerError)
}

func ReadBody[T any](w http.ResponseWriter, r *http.Request, body *T) bool {
	log := zerolog.Ctx(r.Context())

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(body)
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			SendJSONResponse(w, errorResponse{Code: "request_too_large"}, http.StatusRequestEntityTooLarge)
			return false
		}
		log.Error().Err(err).Msg("Failed to decode JSON request")
		SendClientError(w, "invalid_json", nil)
		return false
	}

	return true
}

func ReadAndValidateBody[T any](w http.ResponseWriter, r *http.Request, body *T, schema *z.StructSchema) bool {
	if !ReadBody(w, r, body) {
		return false
	}
	if errs := schema.Validate(body); errs != nil {
		issues := z.Issues.FlattenAndCollect(errs)
		SendClientError(w, "validation_error", issues)
		return false
	}
	return true
}

// ExtendWriteDeadline sets the connection's absolute write deadline to now + timeout, replacing the server's WriteTimeout for this request.
// A long-running handler calls it first and again right before writing the response; see docs/handbuch.md §6.2.
func ExtendWriteDeadline(w http.ResponseWriter, r *http.Request, timeout time.Duration) {
	if err := http.NewResponseController(w).SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		zerolog.Ctx(r.Context()).Warn().Err(err).Dur("timeout", timeout).Msg("Failed to extend write deadline; falling back to server default")
	}
}

type ErrorCode struct {
	Err  error
	Code string
}

// MapError answers 400 with the code of the first entry the error matches
// (errors.Is) — the order of codes decides — and 500 without a match.
func MapError(w http.ResponseWriter, err error, codes []ErrorCode) {
	for _, entry := range codes {
		if errors.Is(err, entry.Err) {
			SendClientError(w, entry.Code, nil)
			return
		}
	}
	SendServerError(w)
}
