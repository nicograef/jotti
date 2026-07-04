package helper

import (
	"encoding/json"
	"errors"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

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

// SendConflict sends a 409 Conflict response with the given error code.
func SendConflict(w http.ResponseWriter, code string) {
	SendJSONResponse(w, errorResponse{Code: code}, http.StatusConflict)
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
