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

func TestReadAndValidateBody_Success(t *testing.T) {
	rec := httptest.NewRecorder()
	type testStruct struct {
		Foo string
		Bar int
	}
	schema := z.Struct(z.Shape{
		"Foo": z.String().Min(3),
		"Bar": z.Int().GT(5),
	})

	body := bytes.NewBufferString(`{"Foo":"bar","Bar":10}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	var dest testStruct
	ok := ReadAndValidateBody(rec, req, &dest, schema)

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

func TestReadAndValidateBody_InvalidJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	type testStruct struct {
		Foo string
	}
	schema := z.Struct(z.Shape{
		"Foo": z.String(),
	})

	body := bytes.NewBufferString(`{"Foo":}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	var dest testStruct
	ok := ReadAndValidateBody(rec, req, &dest, schema)

	if ok {
		t.Errorf("expected failure for invalid JSON")
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	var resp ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if resp.Code != "invalid_json" {
		t.Errorf("expected code invalid_json, got %s", resp.Code)
	}
}

func TestReadAndValidateBody_ValidationFailure(t *testing.T) {
	rec := httptest.NewRecorder()
	type testStruct struct {
		Foo string
		Bar int
	}
	schema := z.Struct(z.Shape{
		"Foo": z.String().Min(3),
		"Bar": z.Int().GT(5),
	})

	body := bytes.NewBufferString(`{"Foo":"ab","Bar":2}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	var dest testStruct
	ok := ReadAndValidateBody(rec, req, &dest, schema)

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
	var resp ErrorResponse
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

func TestReadAndValidateBody_UnknownFields(t *testing.T) {
	rec := httptest.NewRecorder()
	type testStruct struct {
		Foo string
	}
	schema := z.Struct(z.Shape{
		"Foo": z.String(),
	})

	body := bytes.NewBufferString(`{"Foo":"bar","Unknown":"field"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	var dest testStruct
	ok := ReadAndValidateBody(rec, req, &dest, schema)

	if ok {
		t.Errorf("expected failure for unknown fields")
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
