//go:build unit

package event

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNew_Success(t *testing.T) {
	data := map[string]any{"k": "v"}
	e, err := New(123, "TestUser", "com.example.event:v1", "table:123", data)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if e.UserID != 123 {
		t.Errorf("unexpected user ID: %d", e.UserID)
	}
	if e.UserName != "TestUser" {
		t.Errorf("unexpected user name: %s", e.UserName)
	}
	if e.Type != "com.example.event:v1" {
		t.Errorf("unexpected type: %s", e.Type)
	}
	if e.Subject != "table:123" {
		t.Errorf("unexpected subject: %s", e.Subject)
	}
	if e.Data == nil {
		t.Errorf("expected data to be set")
	}
	if time.Since(e.Time) > time.Minute {
		t.Errorf("unexpected event time: %v", e.Time)
	}
}

func TestValidate_Errors(t *testing.T) {
	cases := []struct {
		name     string
		mutate   func(*Event)
		expected string
	}{
		{"non-positive user ID", func(e *Event) { e.UserID = 0 }, "user ID must be a positive integer"},
		{"empty user name", func(e *Event) { e.UserName = "" }, "user name must be a non-empty string"},
		{"whitespace user name", func(e *Event) { e.UserName = "   " }, "user name must be a non-empty string"},
		{"short type", func(e *Event) { e.Type = "aaa" }, "event type must be at least 5 characters long"},
		{"zero time", func(e *Event) { e.Time = time.Time{} }, "event time cannot be zero"},
		{"short subject", func(e *Event) { e.Subject = "" }, "event subject must be a non-empty string"},
		{"zero version", func(e *Event) { e.Version = 0 }, "event version must be >= 1"},
		{"negative version", func(e *Event) { e.Version = -1 }, "event version must be >= 1"},
		{"nil data", func(e *Event) { e.Data = []byte{} }, "event data cannot be empty"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := &Event{
				UserID:   123,
				UserName: "TestUser",
				Type:     "com.example.event:v1",
				Time:     time.Now().UTC(),
				Subject:  "table:123",
				Version:  1,
				Data:     json.RawMessage(`{"k": "v"}`),
			}
			// mutate to make invalid
			tc.mutate(e)
			if err := e.Validate(); err == nil || err.Error() != tc.expected {
				t.Fatalf("expected error %q, got %v", tc.expected, err)
			}
		})
	}
}
