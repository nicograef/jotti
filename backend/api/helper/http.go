package helper

import (
	"encoding/json"
	"errors"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// errorResponse is the uniform error body of the HTTP API.
//
// Code is a stable, machine-readable error code (snake_case); the frontend
// maps it to a German user-facing message.
//
// Details is an optional diagnostic field. Its shape is part of the API
// contract in exactly two cases, which the frontend parses:
//   - code "validation_error": zog issues as map[field][]message
//     (see ReadAndValidateBody)
//   - code "signaturen_ausstehend" (Kassenabschluss-Gate): structured object
//     with the number of pending signatures and the age of the oldest
//     (see SendConflictDetails)
//
// Everywhere else, details is at most a short English diagnostic string for
// operators and logs — never localized, never parsed by clients.
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

// SendNotFound sends a 404 Not Found response with the given error code.
func SendNotFound(w http.ResponseWriter, code string) {
	SendJSONResponse(w, errorResponse{Code: code}, http.StatusNotFound)
}

// SendUnauthorized sends a 401 Unauthorized response with the given error code.
// The frontend logs the user out and redirects to the login page on 401.
func SendUnauthorized(w http.ResponseWriter, code string) {
	SendJSONResponse(w, errorResponse{Code: code}, http.StatusUnauthorized)
}

// SendForbidden sends a 403 Forbidden response with the given error code and
// an optional short English diagnostic. Used when the authenticated user's
// role lacks permission; the frontend keeps the session (auto-logout is bound
// to 401).
func SendForbidden(w http.ResponseWriter, code string, details any) {
	SendJSONResponse(w, errorResponse{Code: code, Details: details}, http.StatusForbidden)
}

// SendConflict sends a 409 Conflict response with the given error code.
func SendConflict(w http.ResponseWriter, code string) {
	SendJSONResponse(w, errorResponse{Code: code}, http.StatusConflict)
}

// SendTooManyRequests sends a 429 Too Many Requests response with the given error
// code — used e.g. by the per-account login throttle so the client sees a clear
// "throttled" signal instead of "invalid credentials".
func SendTooManyRequests(w http.ResponseWriter, code string) {
	SendJSONResponse(w, errorResponse{Code: code}, http.StatusTooManyRequests)
}

// SendConflictDetails sends a 409 Conflict response with an error code and
// structured details (e.g. the Kassenabschluss-Gate reports the number of
// pending signatures and the age of the oldest).
func SendConflictDetails(w http.ResponseWriter, code string, details any) {
	SendJSONResponse(w, errorResponse{Code: code, Details: details}, http.StatusConflict)
}

func SendServerError(w http.ResponseWriter) {
	SendJSONResponse(w, errorResponse{Code: "internal_server_error"}, http.StatusInternalServerError)
}

// ReadBody reads the JSON request body into the provided struct
func ReadBody[T any](w http.ResponseWriter, r *http.Request, body *T) bool {
	log := zerolog.Ctx(r.Context())

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Disallow unknown fields for strict matching

	err := decoder.Decode(body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			SendJSONResponse(w, errorResponse{Code: "request_too_large"}, http.StatusRequestEntityTooLarge)
			return false
		}
		log.Error().Err(err).Msg("Failed to decode JSON request")
		SendClientError(w, "invalid_json", nil)
		return false
	}

	return true
}

// ReadAndValidateBody reads the JSON request body and validates it against a zog struct schema.
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

// MapError maps a domain/application error to an HTTP error response.
// It checks the error against each entry in the provided error-to-code map.
// If a match is found, it sends a client error with the corresponding code.
// If no match is found, it sends a generic server error.
func MapError(w http.ResponseWriter, err error, codeMap map[error]string) {
	for target, code := range codeMap {
		if errors.Is(err, target) {
			SendClientError(w, code, nil)
			return
		}
	}
	SendServerError(w)
}
