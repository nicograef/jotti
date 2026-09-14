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

// errorResponse is the uniform error body of the HTTP API. Code is a stable
// snake_case code the frontend maps to a German message. Details is parsed by
// the frontend in exactly two cases: "validation_error" carries zog issues as
// map[field][]message (see ReadAndValidateBody), "signaturen_ausstehend"
// (Kassenabschluss-Gate) a structured object with the number of pending
// signatures (see SendConflictDetails). Everywhere else it is at most a short
// English diagnostic for operators and logs — never localized, never parsed.
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

// ExtendWriteDeadline setzt die Schreibfrist der Verbindung auf jetzt + timeout
// und ersetzt damit für diesen Request die globale Frist des Servers
// (WriteTimeout: 10s, backend/app/app.go).
//
// Die Frist ist eine ABSOLUTE Zeit, keine Stoppuhr für den Schreibvorgang:
// net/http setzt sie beim Lesen der Request-Header auf jetzt + WriteTimeout, sie
// läuft also während der gesamten Handler-Laufzeit weiter. Ein Handler, der
// länger arbeitet als sein timeout, schreibt danach in eine abgelaufene Frist —
// die Arbeit war erfolgreich, das Ergebnis erreicht den Client nie. Bei der
// TSE-Einrichtung wären PUK und Admin-PIN damit verloren: Sie werden genau
// einmal ausgeliefert und nirgends persistiert.
//
// Ein langlaufender Handler ruft die Funktion deshalb ZWEIMAL auf: als erste
// Anweisung, damit auch alles, was vor der langen Arbeit antwortet (ungültiger
// Body, fachliche Ablehnung), dieses Budget statt der globalen 10 Sekunden
// bekommt; und unmittelbar vor dem Schreiben der Antwort, damit der
// Schreibvorgang ein eigenes Budget hat.
//
// Lässt sich die Frist nicht setzen (ResponseWriter ohne Unterstützung, z. B.
// httptest.ResponseRecorder), wird nur gewarnt: Die Verlängerung ist eine
// Verbesserung, kein Abbruchgrund.
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
