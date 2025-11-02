package api

// helper function for sending json responses
import (
	"encoding/json"
	"log"
	"net/http"
)

type ErrorResponse struct {
	Message string `json:"error"`
	Code    string `json:"code"`
}

func sendJSONResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func sendInternalServerError(w http.ResponseWriter) {
	http.Error(w, "Internal server error", http.StatusInternalServerError)
}

func sendBadRequestError(w http.ResponseWriter, response ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}

func sendUnauthorizedError(w http.ResponseWriter, response ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(response)
}

func sendMethodNotAllowedError(w http.ResponseWriter, response ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	json.NewEncoder(w).Encode(response)
}

// readJSONRequest reads JSON from the request body into the provided destination.
// Returns false if decoding fails.
func readJSONRequest[T any](w http.ResponseWriter, r *http.Request, dest *T) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Disallow unknown fields for strict matching

	err := decoder.Decode(dest)
	if err != nil {
		log.Printf("ERROR Failed to decode JSON request: %v", err)
		sendBadRequestError(w, ErrorResponse{
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
		sendMethodNotAllowedError(w, ErrorResponse{
			Message: "Method not allowed",
			Code:    "method_not_allowed",
		})
		return false
	}

	return true
}
