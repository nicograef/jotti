//go:build unit

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	z "github.com/Oudwins/zog"
)

func TestSendJSONResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	data := map[string]string{"foo": "bar"}
	sendResponse(rec, data)
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

func TestReadJSONRequest_Success(t *testing.T) {
	rec := httptest.NewRecorder()
	type testStruct struct{ Foo string }
	body := bytes.NewBufferString(`{"Foo":"bar"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	var dest testStruct
	ok := readJSONRequest(rec, req, &dest)
	if !ok {
		t.Errorf("expected success reading JSON")
	}
	if dest.Foo != "bar" {
		t.Errorf("expected Foo=bar, got %s", dest.Foo)
	}
}

func TestReadJSONRequest_InvalidJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	type testStruct struct{ Foo string }
	body := bytes.NewBufferString(`{"Foo":}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	var dest testStruct
	ok := readJSONRequest(rec, req, &dest)
	if ok {
		t.Errorf("expected failure for invalid JSON")
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestValidateBody_Success(t *testing.T) {
	rec := httptest.NewRecorder()
	type testStruct struct {
		Foo string
		Bar int
	}
	schema := z.Struct(z.Shape{
		"Foo": z.String().Min(3),
		"Bar": z.Int().GT(5),
	})

	body := testStruct{Foo: "bar", Bar: 10}
	ok := validateBody(rec, &body, schema)

	if !ok {
		t.Errorf("expected body to be valid")
	}
	var resp errorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err == nil {
		t.Errorf("expected no response body, got %v", resp)
	}
}

func TestValidateBody_Invalid(t *testing.T) {
	rec := httptest.NewRecorder()
	type testStruct struct {
		Foo string
		Bar int
	}
	schema := z.Struct(z.Shape{
		"Foo": z.String().Min(3),
		"Bar": z.Int().GT(5),
	})

	body := testStruct{Foo: "ab", Bar: 2}
	ok := validateBody(rec, &body, schema)

	if ok {
		t.Errorf("expected body to be invalid")
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "number must be greater than 5") {
		t.Errorf("expected error message about Bar greater than, got %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "string must contain at least 3 character(s)") {
		t.Errorf("expected error message about Foo length, got %s", rec.Body.String())
	}
	var resp errorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if resp.Code != "invalid_request_body" {
		t.Errorf("expected code invalid_request_body, got %s", resp.Code)
	}
	if resp.Details == nil {
		t.Errorf("expected details in response, got nil")
	}
}

func TestValidateMethod(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ok := validateMethod(rec, req, http.MethodPost)
	if ok {
		t.Errorf("expected method not allowed")
	}
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/", nil)
	ok2 := validateMethod(rec2, req2, http.MethodPost)
	if !ok2 {
		t.Errorf("expected method allowed")
	}
}
