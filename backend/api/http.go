package api

import (
	"encoding/json"
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type errorResponse struct {
	Code    string `json:"code"`
	Details any    `json:"details,omitempty"`
}

func sendJSONResponse(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("Failed to encode JSON response")
	}
}

func SendResponse(w http.ResponseWriter, response any) {
	sendJSONResponse(w, response, http.StatusOK)
}

func SendEmptyResponse(w http.ResponseWriter) {
	sendJSONResponse(w, struct{}{}, http.StatusOK)
}

func SendClientError(w http.ResponseWriter, code string, details any) {
	sendJSONResponse(w, errorResponse{Code: code, Details: details}, http.StatusBadRequest)
}

func SendServerError(w http.ResponseWriter) {
	sendJSONResponse(w, errorResponse{Code: "internal_server_error"}, http.StatusInternalServerError)
}

// ReadAndValidateBody reads the JSON request body into the provided struct
// and validates it against the provided Zod schema.
func ReadAndValidateBody[T any](w http.ResponseWriter, r *http.Request, body *T, schema *z.StructSchema) bool {
	log := zerolog.Ctx(r.Context())

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Disallow unknown fields for strict matching

	err := decoder.Decode(body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode JSON request")
		SendClientError(w, "invalid_json", nil)
		return false
	}

	if err := schema.Validate(body); err != nil {
		issues := z.Issues.SanitizeMapAndCollect(err)
		log.Error().Interface("issues", issues).Msg("Invalid request body")
		SendClientError(w, "invalid_request_body", issues)
		return false
	}

	return true
}
