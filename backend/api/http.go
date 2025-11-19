package api

import (
	"encoding/json"
	"log"
	"net/http"

	z "github.com/Oudwins/zog"
)

type ErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
	Details any    `json:"details,omitempty"`
}

func sendJSONResponse(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("ERROR Failed to encode JSON response: %v", err)
	}
}

func SendResponse(w http.ResponseWriter, data any) {
	sendJSONResponse(w, data, http.StatusOK)
}

func SendEmptyResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func SendInternalServerError(w http.ResponseWriter) {
	response := ErrorResponse{
		Message: "Internal server error",
		Code:    "internal_server_error",
	}
	sendJSONResponse(w, response, http.StatusInternalServerError)
}

func SendBadRequestError(w http.ResponseWriter, response ErrorResponse) {
	sendJSONResponse(w, response, http.StatusBadRequest)
}

func SendNotFoundError(w http.ResponseWriter, response ErrorResponse) {
	sendJSONResponse(w, response, http.StatusNotFound)
}

func SendUnauthorizedError(w http.ResponseWriter, response ErrorResponse) {
	sendJSONResponse(w, response, http.StatusUnauthorized)
}

func SendForbiddenError(w http.ResponseWriter, response ErrorResponse) {
	sendJSONResponse(w, response, http.StatusForbidden)
}

func SendMethodNotAllowedError(w http.ResponseWriter, response ErrorResponse) {
	sendJSONResponse(w, response, http.StatusMethodNotAllowed)
}

// ReadJSONRequest reads JSON from the request body into the provided destination.
// Returns false if decoding fails.
func ReadJSONRequest[T any](w http.ResponseWriter, r *http.Request, dest *T) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Disallow unknown fields for strict matching

	err := decoder.Decode(dest)
	if err != nil {
		log.Printf("ERROR Failed to decode JSON request: %v", err)
		SendBadRequestError(w, ErrorResponse{
			Message: "Invalid JSON request",
			Code:    "invalid_json",
		})
		return false
	}

	return true
}

func ValidateBody[T any](w http.ResponseWriter, body *T, schema *z.StructSchema) bool {
	if err := schema.Validate(body); err != nil {
		issues := z.Issues.SanitizeMapAndCollect(err)
		log.Printf("ERROR Invalid request body: %v", issues)
		SendBadRequestError(w, ErrorResponse{
			Message: "Invalid request body",
			Code:    "invalid_request_body",
			Details: issues,
		})
		return false
	}
	return true
}

func ValidateMethod(w http.ResponseWriter, r *http.Request, expectedMethod string) bool {
	if r.Method != expectedMethod {
		log.Printf("ERROR Invalid method %s, expected %s", r.Method, expectedMethod)
		SendMethodNotAllowedError(w, ErrorResponse{
			Message: "Method not allowed",
			Code:    "method_not_allowed",
		})
		return false
	}

	return true
}
