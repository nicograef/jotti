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
		{"Erst-Upgrade: Daten, aber kein Marker", "", "v2", true, true},
		{"Erstinstallation: kein Marker, keine Daten", "", "v2", false, false},
		{"gleiche Version", "v1", "v1", true, false},
		{"Wechsel ohne Daten", "v1", "v2", false, false},
		{"Wechsel mit Daten", "v1", "v2", true, true},
		{"Dev-Build sichert nie (kein Marker)", "", "dev", true, false},
		{"Dev-Build sichert nie (mit Marker)", "v1", "dev", true, false},
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

func TestPlanBackupMirrorCopiesAndRotates(t *testing.T) {
	// Host hat schon zwei aeltere Dumps; der neue kommt hinzu. Bei keep 2
	// bleiben die neuesten zwei erhalten, der aelteste faellt weg.
	host := []string{"jotti-20260610-080000.sql", "jotti-20260611-060000.sql"}
	got := PlanBackupMirror("jotti-20260612-070000.sql", host, 2)
	want := MirrorPlan{
		Copy:   "jotti-20260612-070000.sql",
		Delete: []string{"jotti-20260610-080000.sql"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("PlanBackupMirror: got %+v, want %+v", got, want)
	}
}

func TestPlanBackupMirrorSkipsCopyWhenAlreadyOnHost(t *testing.T) {
	// Liegt der Dump schon auf dem Host (Wiederholung im selben Sekundentakt),
	// wird nicht erneut kopiert; die Gesamtmenge aendert sich nicht.
	host := []string{"jotti-20260611-060000.sql", "jotti-20260612-070000.sql"}
	got := PlanBackupMirror("jotti-20260612-070000.sql", host, 5)
	if got.Copy != "" {
		t.Fatalf("bereits gespiegelter Dump darf nicht erneut kopiert werden: got Copy %q", got.Copy)
	}
	if got.Delete != nil {
		t.Fatalf("unter dem Limit darf nichts geloescht werden: got %v", got.Delete)
	}
}

func TestPlanBackupMirrorKeepsWhenWithinLimit(t *testing.T) {
	// Leerer Host, erster Spiegel: kopieren, nichts loeschen.
	got := PlanBackupMirror("jotti-20260612-070000.sql", nil, 5)
	if got.Copy != "jotti-20260612-070000.sql" {
		t.Fatalf("erster Spiegel muss kopieren: got Copy %q", got.Copy)
	}
	if got.Delete != nil {
		t.Fatalf("unter dem Limit darf nichts geloescht werden: got %v", got.Delete)
	}
}

func TestPlanBackupMirrorGuardsAgainstZeroKeep(t *testing.T) {
	// keep <= 0 ist eine Fehlkonfiguration: kopieren ja, aber nie alles loeschen.
	host := []string{"jotti-20260610-080000.sql", "jotti-20260611-060000.sql"}
	got := PlanBackupMirror("jotti-20260612-070000.sql", host, 0)
	if got.Copy != "jotti-20260612-070000.sql" {
		t.Fatalf("Kopie muss auch bei keep <= 0 erfolgen: got Copy %q", got.Copy)
	}
	if got.Delete != nil {
		t.Fatalf("keep <= 0 darf nie Host-Backups loeschen: got %v", got.Delete)
	}
}
