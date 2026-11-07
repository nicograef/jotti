package application

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/betreiber"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/internal/logging"
	"github.com/nicograef/jotti/backend/repository/repotest"
)

var testOpenKS = &kasse.Kassensitzung{
	ZNr:       1,
	Status:    kasse.KassensitzungOffen,
	CreatedAt: time.Now().UTC(),
	UpdatedAt: time.Now().UTC(),
}

type settingsMock struct {
	vereinsname  string
	betreiberErr error
}

func (m settingsMock) GetBetreiber(_ context.Context) (betreiber.Betreiber, error) {
	if m.betreiberErr != nil {
		return betreiber.Betreiber{}, m.betreiberErr
	}
	return betreiber.Betreiber{
		Vereinsname: m.vereinsname,
		Strasse:     "Teststraße 1",
		Plz:         "12345",
		Ort:         "Teststadt",
		UpdatedAt:   time.Now(),
	}, nil
}

// tseGateMock speist das Kassenabschluss-Gate: die noch nicht erledigten
// Signatur-Stände der Sitzung und der aktive Störungszeitraum (Nullwert lässt
// das Gate durch) sowie die TSE-Konfiguration für die Eröffnungs-Warnung.
type tseGateMock struct {
	staende  []tse.SignaturauftragStand
	stoerung *tse.Stoerung
	err      error
	tse      tse.Konfiguration
	tseErr   error
}

func (m tseGateMock) GetOffeneSignaturauftragStaendeFuerKassensitzung(_ context.Context, _ int) ([]tse.SignaturauftragStand, error) {
	return m.staende, m.err
}

func (m tseGateMock) GetAktiveTSEStoerung(_ context.Context) (*tse.Stoerung, error) {
	return m.stoerung, nil
}

func (m tseGateMock) GetTSEKonfiguration(_ context.Context) (tse.Konfiguration, error) {
	if m.tseErr != nil {
		return tse.Konfiguration{}, m.tseErr
	}
	if !m.tse.IstKonfiguriert() {
		return tse.Konfiguration{}, db.ErrNotFound
	}
	return m.tse, nil
}

func newTestCommand(ks *kasse.Kassensitzung) Command {
	journalMock := repotest.NewKassenjournalRepo(nil, nil)
	sitzungMock := repotest.NewKassensitzungenRepo(ks, nil)
	return Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: sitzungMock,
		BetreiberRepo:       settingsMock{vereinsname: "TestVerein"},
		TSERepo:             tseGateMock{},
		DruckauftragRepo:    &fakeDruckauftraege{},
	}
}

func TestKassensitzungEroeffnen(t *testing.T) {
	ctx := context.Background()
	cmd := newTestCommand(nil) // no open KS

	zNr, err := cmd.KassensitzungEroeffnen(ctx, 1, "Admin", "Vereinsfest 2026", 10000)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if zNr < 1 {
		t.Errorf("expected z_nr >= 1, got %d", zNr)
	}
}

func TestKassensitzungEroeffnen_BetreiberNichtKonfiguriert(t *testing.T) {
	ctx := context.Background()
	cmd := Command{
		KassenjournalRepo:   repotest.NewKassenjournalRepo(nil, nil),
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(nil, nil),
		BetreiberRepo:       settingsMock{betreiberErr: db.ErrNotFound},
	}

	_, err := cmd.KassensitzungEroeffnen(ctx, 1, "Admin", "Vereinsfest 2026", 10000)
	if !errors.Is(err, ErrBetreiberNichtKonfiguriert) {
		t.Errorf("expected ErrBetreiberNichtKonfiguriert, got %v", err)
	}
}

func TestKassensitzungEroeffnen_BetreiberDatabaseError(t *testing.T) {
	ctx := context.Background()
	cmd := Command{
		KassenjournalRepo:   repotest.NewKassenjournalRepo(nil, nil),
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(nil, nil),
		BetreiberRepo:       settingsMock{betreiberErr: db.ErrDatabase},
	}

	_, err := cmd.KassensitzungEroeffnen(ctx, 1, "Admin", "Vereinsfest 2026", 10000)
	if !errors.Is(err, ErrDatabase) {
		t.Errorf("expected ErrDatabase, got %v", err)
	}
}

func TestKassensitzungEroeffnen_AlreadyOpen(t *testing.T) {
	ctx := context.Background()
	cmd := newTestCommand(testOpenKS)

	_, err := cmd.KassensitzungEroeffnen(ctx, 1, "Admin", "Vereinsfest 2026", 10000)
	if !errors.Is(err, ErrKasseAlreadyOpen) {
		t.Errorf("expected ErrKasseAlreadyOpen, got %v", err)
	}
}

// A concurrent open that passed the active-session check hits idx_kassensitzungen_eine_aktiv.
func TestKassensitzungEroeffnen_ParalleleEroeffnung(t *testing.T) {
	ctx := context.Background()
	cmd := newTestCommand(nil)
	cmd.KassenjournalRepo = repotest.NewKassenjournalRepo(nil, db.ErrAlreadyExists)

	_, err := cmd.KassensitzungEroeffnen(ctx, 1, "Admin", "Vereinsfest 2026", 10000)
	if !errors.Is(err, ErrKasseAlreadyOpen) {
		t.Errorf("expected ErrKasseAlreadyOpen, got %v", err)
	}
}

func TestGeldtransitBuchen(t *testing.T) {
	ctx := context.Background()
	cmd := newTestCommand(testOpenKS)

	err := cmd.GeldtransitBuchen(ctx, 1, "Admin", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "einlage", 10000, "Wechselgeld nachgelegt")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// Eine Buchung in eine Sitzung im Zwischenstatus wird mit ErrKasseWirdAbgeschlossen abgelehnt.
func TestGeldtransitBuchen_WirdAbgeschlossen(t *testing.T) {
	ctx := context.Background()
	imAbschluss := &kasse.Kassensitzung{ZNr: 1, Status: kasse.KassensitzungWirdAbgeschlossen, CreatedAt: time.Now().UTC()}
	cmd := newTestCommand(imAbschluss)

	err := cmd.GeldtransitBuchen(ctx, 1, "Admin", "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", "einlage", 1000, "Wechselgeld")
	if !errors.Is(err, ErrKasseWirdAbgeschlossen) {
		t.Fatalf("expected ErrKasseWirdAbgeschlossen, got %v", err)
	}
}

// Eröffnen ohne konfigurierte TSE wird nicht gesperrt, aber im Log vermerkt.
func TestKassensitzungEroeffnen_OhneTSE_LoggtWarnung(t *testing.T) {
	var logbuf bytes.Buffer
	ctx := logging.WithContext(context.Background(), logging.New(&logbuf))
	cmd := newTestCommand(nil) // settingsMock ohne TSE-Konfiguration

	zNr, err := cmd.KassensitzungEroeffnen(ctx, 1, "Admin", "Vereinsfest 2026", 10000)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var warnLine string
	for line := range strings.SplitSeq(strings.TrimSpace(logbuf.String()), "\n") {
		if strings.Contains(line, `"level":"WARN"`) && strings.Contains(line, "ohne TSE-Konfiguration") {
			warnLine = line
		}
	}
	if warnLine == "" {
		t.Fatalf("expected warning about missing TSE-Konfiguration, got logs: %s", logbuf.String())
	}
	if !strings.Contains(warnLine, fmt.Sprintf(`"z_nr":%d`, zNr)) {
		t.Errorf("expected z_nr %d in warning, got: %s", zNr, warnLine)
	}
}
