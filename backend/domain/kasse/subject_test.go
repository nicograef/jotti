//go:build unit

package kasse

import "testing"

func TestKassensitzungSubject(t *testing.T) {
	cases := []struct {
		znr  int
		want string
	}{
		{1, "kassensitzung-1"},
		{42, "kassensitzung-42"},
	}
	for _, tc := range cases {
		got := KassensitzungSubject(tc.znr)
		if got != tc.want {
			t.Errorf("KassensitzungSubject(%d) = %q, want %q", tc.znr, got, tc.want)
		}
	}
}

func TestTischSessionSubject(t *testing.T) {
	cases := []struct {
		znr     int
		tischID int
		want    string
	}{
		{1, 42, "kassensitzung-1/tisch-42"},
		{3, 100, "kassensitzung-3/tisch-100"},
	}
	for _, tc := range cases {
		got := TischSessionSubject(tc.znr, tc.tischID)
		if got != tc.want {
			t.Errorf("TischSessionSubject(%d, %d) = %q, want %q", tc.znr, tc.tischID, got, tc.want)
		}
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
