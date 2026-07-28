//go:build unit

package helper

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSendJSONResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	data := map[string]string{"foo": "bar"}
	SendResponse(rec, data)
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected content-type application/json, got %s", ct)
	}
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if resp["foo"] != "bar" {
		t.Errorf("expected foo=bar, got %v", resp)
	}
}

func TestReadBody_Success(t *testing.T) {
	rec := httptest.NewRecorder()
	type testStruct struct {
		Foo string
		Bar int
	}

	body := bytes.NewBufferString(`{"Foo":"bar","Bar":10}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	var dest testStruct
	ok := ReadBody(rec, req, &dest)

	if !ok {
		t.Errorf("expected body to be valid")
	}
	if dest.Foo != "bar" {
		t.Errorf("expected Foo=bar, got %s", dest.Foo)
	}
	if dest.Bar != 10 {
		t.Errorf("expected Bar=10, got %d", dest.Bar)
	}
}

func TestReadBody_InvalidJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	type testStruct struct {
		Foo string
	}

	body := bytes.NewBufferString(`{"Foo":}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	var dest testStruct
	ok := ReadBody(rec, req, &dest)

	if ok {
		t.Errorf("expected failure for invalid JSON")
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	var resp errorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if resp.Code != "invalid_json" {
		t.Errorf("expected code invalid_json, got %s", resp.Code)
	}
}

func TestReadBody_TooLarge(t *testing.T) {
	rec := httptest.NewRecorder()
	type testStruct struct {
		Foo string
	}

	// Create a body larger than 1 MB
	largeBody := `{"Foo":"` + strings.Repeat("x", 1<<20+1) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(largeBody))
	var dest testStruct
	ok := ReadBody(rec, req, &dest)

	if ok {
		t.Errorf("expected failure for too-large body")
	}
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got %d", rec.Code)
	}
	var resp errorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if resp.Code != "request_too_large" {
		t.Errorf("expected code request_too_large, got %s", resp.Code)
	}
}

func TestReadBody_UnknownFields(t *testing.T) {
	rec := httptest.NewRecorder()
	type testStruct struct {
		Foo string
	}

	body := bytes.NewBufferString(`{"Foo":"bar","Unknown":"field"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	var dest testStruct
	ok := ReadBody(rec, req, &dest)

	if ok {
		t.Errorf("expected failure for unknown fields")
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// deadlineCapturingWriter implements the SetWriteDeadline interface
// http.ResponseController looks for, so the test can observe the deadline.
type deadlineCapturingWriter struct {
	*httptest.ResponseRecorder
	deadline time.Time
	set      bool
}

func (w *deadlineCapturingWriter) SetWriteDeadline(t time.Time) error {
	w.deadline = t
	w.set = true
	return nil
}

func TestExtendWriteDeadline(t *testing.T) {
	w := &deadlineCapturingWriter{ResponseRecorder: httptest.NewRecorder()}
	req := httptest.NewRequest(http.MethodPost, "/", nil)

	before := time.Now()
	ExtendWriteDeadline(w, req, 2*time.Minute)

	if !w.set {
		t.Fatal("expected SetWriteDeadline to be called")
	}
	if w.deadline.Before(before.Add(2 * time.Minute)) {
		t.Errorf("expected a deadline at least 2m in the future, got %s", w.deadline.Sub(before))
	}
}

// Unterstützt der ResponseWriter SetWriteDeadline nicht (wie
// httptest.ResponseRecorder), ist das kein Abbruchgrund: Der Handler läuft
// mit der globalen Frist weiter und liefert seine Antwort.
func TestExtendWriteDeadline_UnsupportedWriter(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)

	ExtendWriteDeadline(rec, req, 2*time.Minute)

	SendEmptyResponse(rec)
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 after a failed deadline extension, got %d", rec.Code)
	}
}
