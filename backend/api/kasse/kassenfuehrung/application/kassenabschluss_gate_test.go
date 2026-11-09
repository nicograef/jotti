package application

import (
	"context"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/repotest"
)

// The gate judges via tse.DetermineSignaturstatus like the Beleg path (docs/handbuch.md §3.13):
// offen without Störung is ausstehend and blocks; offen in an active Störungszeitraum and the end
// states pass as Ausfall, or as flagged when no TSE is configured.
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

// Ein Tag vollständig ohne TSE schließt durch: nur tse_nicht_konfiguriert-Reste,
// kein Block, in der Abschlussmeldung deutlich ausgewiesen.
func TestKasseAbschliessen_GateTagOhneTSE(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	journalMock := repotest.NewKassenjournalRepo(nil, nil)
	journalMock.SetKassenbestand(50000)
	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
		DruckauftragRepo:    &fakeDruckauftraege{},
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
