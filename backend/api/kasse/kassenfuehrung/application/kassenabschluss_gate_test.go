//go:build unit

package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
)

// Das Gate urteilt über dieselbe tse.DetermineSignaturstatus-Funktion wie der
// Beleg-Abruf (kein zweiter Zurechnungspfad): offen ohne Störung ist ausstehend
// (blockt), offen bei aktivem Störungszeitraum ist Ausfall (lässt durch),
// Endstatus ist Ausfall bzw. — bei fehlender Konfiguration — deutlich ausgewiesen.
func TestCheckSignaturGate(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	offen := tse.SignaturauftragStand{Status: tse.StatusOffen, ErstelltAm: now}
	tseFehler := &tse.Stoerung{Beginn: now.Add(-time.Minute), GrundArt: tse.StoerungGrundTSEFehler}
	keineKonfig := &tse.Stoerung{Beginn: now.Add(-time.Minute), GrundArt: tse.StoerungGrundKeineKonfiguration}

	cases := []struct {
		name             string
		staende          []tse.SignaturauftragStand
		stoerung         *tse.Stoerung
		wantAusstehend   int
		wantAusfallReste int
		wantOhneKonfig   int
	}{
		{"leere Queue lässt durch", nil, nil, 0, 0, 0},
		{"offen ohne Störung → ausstehend", []tse.SignaturauftragStand{offen}, nil, 1, 0, 0},
		{"offen bei TSE-Störung → Ausfall-Rest", []tse.SignaturauftragStand{offen}, tseFehler, 0, 1, 0},
		{"offen bei keine_konfiguration-Störung → ohne Konfiguration", []tse.SignaturauftragStand{offen}, keineKonfig, 0, 0, 1},
		{"fehlgeschlagen → Ausfall-Rest", []tse.SignaturauftragStand{{Status: tse.StatusFehlgeschlagen, ErstelltAm: now}}, nil, 0, 1, 0},
		{"tse_nicht_konfiguriert → ohne Konfiguration", []tse.SignaturauftragStand{{Status: tse.StatusTSENichtKonfiguriert, ErstelltAm: now}}, nil, 0, 0, 1},
		{
			"gemischt: ausstehend, Ausfall-Rest und ohne Konfiguration",
			[]tse.SignaturauftragStand{
				offen,
				{Status: tse.StatusFehlgeschlagen, ErstelltAm: now},
				{Status: tse.StatusTSENichtKonfiguriert, ErstelltAm: now},
			},
			nil, 1, 1, 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := Command{TSERepo: tseGateMock{staende: tc.staende, stoerung: tc.stoerung}}
			gate, err := cmd.checkSignaturGate(ctx, 1)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if gate.ausstehendAnzahl != tc.wantAusstehend {
				t.Errorf("ausstehendAnzahl = %d, want %d", gate.ausstehendAnzahl, tc.wantAusstehend)
			}
			if gate.ausfallResteAnzahl != tc.wantAusfallReste {
				t.Errorf("ausfallResteAnzahl = %d, want %d", gate.ausfallResteAnzahl, tc.wantAusfallReste)
			}
			if gate.ohneKonfigurationAnzahl != tc.wantOhneKonfig {
				t.Errorf("ohneKonfigurationAnzahl = %d, want %d", gate.ohneKonfigurationAnzahl, tc.wantOhneKonfig)
			}
		})
	}
}

// checkSignaturGate meldet Anzahl und Alter des ältesten ausstehenden Auftrags
// für die 409-Antwort.
func TestCheckSignaturGate_AeltesterAusstehend(t *testing.T) {
	ctx := context.Background()
	alt := time.Now().Add(-45 * time.Second).UTC()
	jung := time.Now().Add(-5 * time.Second).UTC()
	cmd := Command{TSERepo: tseGateMock{staende: []tse.SignaturauftragStand{
		{Status: tse.StatusOffen, ErstelltAm: jung},
		{Status: tse.StatusOffen, ErstelltAm: alt},
	}}}

	gate, err := cmd.checkSignaturGate(ctx, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if gate.ausstehendAnzahl != 2 {
		t.Fatalf("expected 2 ausstehend, got %d", gate.ausstehendAnzahl)
	}
	if !gate.aeltesterAusstehend.Equal(alt) {
		t.Errorf("aeltesterAusstehend = %v, want %v (oldest)", gate.aeltesterAusstehend, alt)
	}
}

// Frischer offener Auftrag ohne Störung blockiert den Abschluss mit
// SignaturenAusstehendError (Anzahl + Alter); die Barriere wird nicht gesetzt
// und kein Event geschrieben — das Gate greift vor der Barriere.
func TestKasseAbschliessen_GateBlocktBeiAusstehend(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	journalMock.SetKassenbestand(50000)
	sitzungMock := kassensitzungen_repo.NewMock(testOpenKS, nil)
	erstellt := time.Now().Add(-20 * time.Second).UTC()
	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: sitzungMock,

		TSERepo: tseGateMock{staende: []tse.SignaturauftragStand{
			{Status: tse.StatusOffen, ErstelltAm: erstellt},
		}},
	}

	_, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000)
	var ausstehend *SignaturenAusstehendError
	if !errors.As(err, &ausstehend) {
		t.Fatalf("expected SignaturenAusstehendError, got %v", err)
	}
	if ausstehend.Anzahl != 1 {
		t.Errorf("expected Anzahl 1, got %d", ausstehend.Anzahl)
	}
	if !ausstehend.AeltesterErstelltAm.Equal(erstellt) {
		t.Errorf("expected AeltesterErstelltAm %v, got %v", erstellt, ausstehend.AeltesterErstelltAm)
	}
	if sitzungMock.WirdAbgeschlossenCalls != 0 {
		t.Errorf("expected barrier NOT set when gate blocks, got %d calls", sitzungMock.WirdAbgeschlossenCalls)
	}
	events, _ := journalMock.ReadEventsBySubject(ctx, kasse.KassensitzungSubject(testOpenKS.ZNr))
	if len(events) != 0 {
		t.Fatalf("expected no events written when gate blocks, got %d", len(events))
	}
}

// Ausfall-Reste (endgültig fehlgeschlagen sowie offen bei aktivem
// Störungszeitraum) lassen den Abschluss zu und werden in der Abschlussmeldung
// ausgewiesen; tse_nicht_konfiguriert blockiert nie und wird deutlich als „Tag
// ohne TSE" ausgewiesen.
func TestKasseAbschliessen_GateLaesstAusfallResteDurch(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	journalMock.SetKassenbestand(50000)
	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),

		TSERepo: tseGateMock{
			staende: []tse.SignaturauftragStand{
				{Status: tse.StatusFehlgeschlagen, ErstelltAm: now},
				{Status: tse.StatusOffen, ErstelltAm: now}, // offen bei aktivem Störungszeitraum → Ausfall
				{Status: tse.StatusTSENichtKonfiguriert, ErstelltAm: now},
			},
			stoerung: &tse.Stoerung{Beginn: now.Add(-time.Minute), GrundArt: tse.StoerungGrundTSEFehler},
		},
	}

	ergebnis, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000)
	if err != nil {
		t.Fatalf("expected Abschluss to pass with Ausfall-Reste, got %v", err)
	}
	if ergebnis.AusfallResteAnzahl != 2 {
		t.Errorf("expected AusfallResteAnzahl 2 (fehlgeschlagen + offen-bei-Störung), got %d", ergebnis.AusfallResteAnzahl)
	}
	if ergebnis.OhneKonfigurationAnzahl != 1 {
		t.Errorf("expected OhneKonfigurationAnzahl 1, got %d", ergebnis.OhneKonfigurationAnzahl)
	}
	events, _ := journalMock.ReadEventsBySubject(ctx, kasse.KassensitzungSubject(testOpenKS.ZNr))
	if len(events) != 2 {
		t.Fatalf("expected kassensturz + tagesabschluss written, got %d", len(events))
	}
}

// Ein Tag vollständig ohne TSE schließt durch: nur tse_nicht_konfiguriert-Reste,
// kein Block, in der Abschlussmeldung deutlich ausgewiesen.
func TestKasseAbschliessen_GateTagOhneTSE(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	journalMock.SetKassenbestand(50000)
	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),

		TSERepo: tseGateMock{staende: []tse.SignaturauftragStand{
			{Status: tse.StatusTSENichtKonfiguriert, ErstelltAm: now},
			{Status: tse.StatusTSENichtKonfiguriert, ErstelltAm: now},
		}},
	}

	ergebnis, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000)
	if err != nil {
		t.Fatalf("expected Abschluss ohne TSE to pass, got %v", err)
	}
	if ergebnis.OhneKonfigurationAnzahl != 2 {
		t.Errorf("expected OhneKonfigurationAnzahl 2, got %d", ergebnis.OhneKonfigurationAnzahl)
	}
	if ergebnis.AusfallResteAnzahl != 0 {
		t.Errorf("expected no AusfallReste, got %d", ergebnis.AusfallResteAnzahl)
	}
}
