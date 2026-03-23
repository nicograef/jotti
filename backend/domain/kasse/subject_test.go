//go:build unit

package kasse

import "testing"

func TestKassensitzungSubject(t *testing.T) {
	got := KassensitzungSubject(1)
	want := "kassensitzung-1"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestKassensitzungSubject_LargeNumber(t *testing.T) {
	got := KassensitzungSubject(42)
	want := "kassensitzung-42"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestTischSessionSubject(t *testing.T) {
	got := TischSessionSubject(1, 42)
	want := "kassensitzung-1/tisch-42"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestTischSessionSubject_LargeNumbers(t *testing.T) {
	got := TischSessionSubject(3, 100)
	want := "kassensitzung-3/tisch-100"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestParseTischIDFromSubject_Valid(t *testing.T) {
	id, err := ParseTischIDFromSubject("kassensitzung-1/tisch-42")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 42 {
		t.Fatalf("expected 42, got %d", id)
	}
}

func TestParseTischIDFromSubject_InvalidFormat(t *testing.T) {
	_, err := ParseTischIDFromSubject("invalid-subject")
	if err == nil {
		t.Fatal("expected error for invalid format, got nil")
	}
}

func TestParseTischIDFromSubject_InvalidID(t *testing.T) {
	_, err := ParseTischIDFromSubject("kassensitzung-1/tisch-abc")
	if err == nil {
		t.Fatal("expected error for invalid ID, got nil")
	}
}

func TestParseZNrFromSubject_KassensitzungSubject(t *testing.T) {
	nr, err := ParseZNrFromSubject("kassensitzung-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if nr != 1 {
		t.Fatalf("expected 1, got %d", nr)
	}
}

func TestParseZNrFromSubject_TischSessionSubject(t *testing.T) {
	nr, err := ParseZNrFromSubject("kassensitzung-3/tisch-42")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if nr != 3 {
		t.Fatalf("expected 3, got %d", nr)
	}
}

func TestParseZNrFromSubject_InvalidFormat(t *testing.T) {
	_, err := ParseZNrFromSubject("invalid-format")
	if err == nil {
		t.Fatal("expected error for invalid format, got nil")
	}
}

func TestParseZNrFromSubject_InvalidNumber(t *testing.T) {
	_, err := ParseZNrFromSubject("kassensitzung-abc")
	if err == nil {
		t.Fatal("expected error for invalid number, got nil")
	}
}
