package api

// helper function for sending json responses
import (
	"encoding/json"
	"log"
	"net/http"
)

func sendJSONResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

// readJSONRequest reads JSON from the request body into the provided destination.
// Returns false if decoding fails.
func readJSONRequest[T any](w http.ResponseWriter, r *http.Request, dest *T) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Disallow unknown fields for strict matching

	err := decoder.Decode(dest)
	if err != nil {
		log.Printf("ERROR Failed to decode JSON request: %v", err)
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return false
	}

	return true
}

func validateMethod(w http.ResponseWriter, r *http.Request, expectedMethod string) bool {
	if r.Method != expectedMethod {
		log.Printf("WARN Invalid method %s, expected %s", r.Method, expectedMethod)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return false
	}

	return true
}
