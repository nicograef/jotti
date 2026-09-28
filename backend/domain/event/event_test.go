package event

import (
	"encoding/json"
	"errors"
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
		expected error
	}{
		{"non-positive user ID", func(e *Event) { e.UserID = 0 }, ErrInvalidUserID},
		{"empty user name", func(e *Event) { e.UserName = "" }, ErrEmptyUserName},
		{"whitespace user name", func(e *Event) { e.UserName = "   " }, ErrEmptyUserName},
		{"short type", func(e *Event) { e.Type = "aaa" }, ErrInvalidType},
		{"zero time", func(e *Event) { e.Time = time.Time{} }, ErrZeroTime},
		{"short subject", func(e *Event) { e.Subject = "" }, ErrEmptySubject},
		{"zero version", func(e *Event) { e.Version = 0 }, ErrInvalidVersion},
		{"negative version", func(e *Event) { e.Version = -1 }, ErrInvalidVersion},
		{"nil data", func(e *Event) { e.Data = []byte{} }, ErrEmptyData},
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
			tc.mutate(e)
			if err := e.Validate(); !errors.Is(err, tc.expected) {
				t.Errorf("expected error %v, got %v", tc.expected, err)
			}
		})
	}
}
