package api

import (
	"encoding/json"
	"log"
	"net/http"
)

type errorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

func sendJSONResponse(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("ERROR Failed to encode JSON response: %v", err)
	}
}

func sendResponse(w http.ResponseWriter, data any) {
	sendJSONResponse(w, data, http.StatusOK)
}

func sendInternalServerError(w http.ResponseWriter) {
	response := errorResponse{
		Message: "Internal server error",
		Code:    "internal_server_error",
	}
	sendJSONResponse(w, response, http.StatusInternalServerError)
}

func sendBadRequestError(w http.ResponseWriter, response errorResponse) {
	sendJSONResponse(w, response, http.StatusBadRequest)
}

func sendNotFoundError(w http.ResponseWriter, response errorResponse) {
	sendJSONResponse(w, response, http.StatusNotFound)
}

func sendUnauthorizedError(w http.ResponseWriter, response errorResponse) {
	sendJSONResponse(w, response, http.StatusUnauthorized)
}

func sendForbiddenError(w http.ResponseWriter, response errorResponse) {
	sendJSONResponse(w, response, http.StatusForbidden)
}

func sendMethodNotAllowedError(w http.ResponseWriter, response errorResponse) {
	sendJSONResponse(w, response, http.StatusMethodNotAllowed)
}

// readJSONRequest reads JSON from the request body into the provided destination.
// Returns false if decoding fails.
func readJSONRequest[T any](w http.ResponseWriter, r *http.Request, dest *T) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Disallow unknown fields for strict matching

	err := decoder.Decode(dest)
	if err != nil {
		log.Printf("ERROR Failed to decode JSON request: %v", err)
		sendBadRequestError(w, errorResponse{
			Message: "Invalid JSON request",
			Code:    "invalid_json",
		})
		return false
	}

	return true
}

func validateMethod(w http.ResponseWriter, r *http.Request, expectedMethod string) bool {
	if r.Method != expectedMethod {
		log.Printf("WARN Invalid method %s, expected %s", r.Method, expectedMethod)
		sendMethodNotAllowedError(w, errorResponse{
			Message: "Method not allowed",
			Code:    "method_not_allowed",
		})
		return false
	}

	return true
}
