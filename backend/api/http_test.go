//go:build unit

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
