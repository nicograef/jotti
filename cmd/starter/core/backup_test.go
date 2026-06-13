package core

import (
	"reflect"
	"testing"
)

func TestShouldBackupOnlyOnVersionChangeWithData(t *testing.T) {
	cases := []struct {
		name        string
		lastVersion string
		current     string
		dataExists  bool
		want        bool
	}{
		{"Erststart ohne Marker", "", "v2", true, false},
		{"gleiche Version", "v1", "v1", true, false},
		{"Wechsel ohne Daten", "v1", "v2", false, false},
		{"Wechsel mit Daten", "v1", "v2", true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ShouldBackup(tc.lastVersion, tc.current, tc.dataExists); got != tc.want {
				t.Fatalf("ShouldBackup(%q, %q, %v): got %v, want %v",
					tc.lastVersion, tc.current, tc.dataExists, got, tc.want)
			}
		})
	}
}

func TestDumpsToDeleteKeepsNewest(t *testing.T) {
	// Bewusst unsortiert uebergeben — die Funktion sortiert chronologisch und
	// loescht die aeltesten ueber keep hinaus.
	names := []string{
		"jotti-20260613-090000.sql",
		"jotti-20260610-080000.sql",
		"jotti-20260612-070000.sql",
		"jotti-20260611-060000.sql",
	}
	got := DumpsToDelete(names, 2)
	want := []string{"jotti-20260610-080000.sql", "jotti-20260611-060000.sql"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DumpsToDelete: got %v, want %v", got, want)
	}
}

func TestDumpsToDeleteNothingWhenWithinLimit(t *testing.T) {
	names := []string{"jotti-20260611-060000.sql", "jotti-20260612-070000.sql"}
	if got := DumpsToDelete(names, 5); got != nil {
		t.Fatalf("unter dem Limit darf nichts geloescht werden: got %v", got)
	}
	if got := DumpsToDelete(names, 2); got != nil {
		t.Fatalf("genau am Limit darf nichts geloescht werden: got %v", got)
	}
}

func TestDumpsToDeleteGuardsAgainstZeroKeep(t *testing.T) {
	names := []string{"jotti-20260611-060000.sql", "jotti-20260612-070000.sql"}
	if got := DumpsToDelete(names, 0); got != nil {
		t.Fatalf("keep <= 0 darf nie alle Backups loeschen: got %v", got)
	}
}
