//go:build unit

package application

import (
	"context"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
)

var testOpenKS = &kasse.Kassensitzung{
	ZNr:       1,
	Status:    kasse.KassensitzungOffen,
	CreatedAt: time.Now().UTC(),
	UpdatedAt: time.Now().UTC(),
}

type settingsMock struct{ vereinsname string }

func (m settingsMock) GetBetreiber(_ context.Context) (settings.Betreiber, error) {
	return settings.Betreiber{Vereinsname: m.vereinsname}, nil
}

func newTestCommand(ks *kasse.Kassensitzung) Command {
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	sitzungMock := kassensitzungen_repo.NewMock(ks, nil)
	return Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: sitzungMock,
		SettingsRepo:        settingsMock{vereinsname: "TestVerein"},
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

func TestKassensitzungEroeffnen_AlreadyOpen(t *testing.T) {
	ctx := context.Background()
	cmd := newTestCommand(testOpenKS)

	_, err := cmd.KassensitzungEroeffnen(ctx, 1, "Admin", "Vereinsfest 2026", 10000)
	if err != ErrKasseAlreadyOpen {
		t.Fatalf("expected ErrKasseAlreadyOpen, got %v", err)
	}
}

func TestGeldtransitBuchen(t *testing.T) {
	ctx := context.Background()
	cmd := newTestCommand(testOpenKS)

	err := cmd.GeldtransitBuchen(ctx, 1, "Admin", "einlage", 10000, "Wechselgeld nachgelegt")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestKassensturzDurchfuehren(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	journalMock.SetKassenbestand(50000) // Soll = 500 EUR
	cmd := Command{KassenjournalRepo: journalMock, KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil)}

	err := cmd.KassensturzDurchfuehren(ctx, 1, "Admin", 49500) // Ist = 495 EUR, Differenz = 500
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestKassensturzDurchfuehren_NoDifference(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	journalMock.SetKassenbestand(50000) // Soll = 500 EUR
	cmd := Command{KassenjournalRepo: journalMock, KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil)}

	// Ist matches Soll exactly — no differenz event should be created
	err := cmd.KassensturzDurchfuehren(ctx, 1, "Admin", 50000)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestKassensturzDurchfuehren_KasseNichtGeoeffnet(t *testing.T) {
	ctx := context.Background()
	cmd := newTestCommand(nil)

	err := cmd.KassensturzDurchfuehren(ctx, 1, "Admin", 50000)
	if err != ErrKasseNichtGeoeffnet {
		t.Fatalf("expected ErrKasseNichtGeoeffnet, got %v", err)
	}
}

func TestTagesabschlussErstellen(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMock(nil, nil)

	// Add a kassensturz event to satisfy the prerequisite
	subject := kasse.KassensitzungSubject(testOpenKS.ZNr)
	kassensturzEvt, _ := kasse.NewKassensturzDurchgefuehrtEvent(subject, 1, "Admin", 50000, 50000, 0)
	journalMock.AddEvent(kassensturzEvt)

	cmd := Command{KassenjournalRepo: journalMock, KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil)}

	err := cmd.TagesabschlussErstellen(ctx, 1, "Admin")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestTagesabschlussErstellen_KassensturzRequired(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	// No kassensturz event added
	cmd := Command{KassenjournalRepo: journalMock, KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil)}

	err := cmd.TagesabschlussErstellen(ctx, 1, "Admin")
	if err != ErrKassensturzErforderlich {
		t.Fatalf("expected ErrKassensturzErforderlich, got %v", err)
	}
}

func TestTagesabschlussErstellen_TischSaldoSperre(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMock(nil, nil)

	// Add kassensturz event to satisfy that prerequisite
	subject := kasse.KassensitzungSubject(testOpenKS.ZNr)
	kassensturzEvt, _ := kasse.NewKassensturzDurchgefuehrtEvent(subject, 1, "Admin", 50000, 50000, 0)
	journalMock.AddEvent(kassensturzEvt)

	// Add a tisch session with non-zero saldo
	tischSubject := kasse.TischSessionSubject(testOpenKS.ZNr, 42)
	journalMock.SetTischSession(tischSubject, kasse.TischSession{
		TischID:         42,
		KassensitzungNr: testOpenKS.ZNr,
		SaldoCents:      350, // non-zero saldo
	})

	cmd := Command{KassenjournalRepo: journalMock, KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil)}

	err := cmd.TagesabschlussErstellen(ctx, 1, "Admin")
	if err != ErrTischeSaldoOffen {
		t.Fatalf("expected ErrTischeSaldoOffen, got %v", err)
	}
}

func TestTagesabschlussErstellen_KasseNichtGeoeffnet(t *testing.T) {
	ctx := context.Background()
	cmd := newTestCommand(nil)

	err := cmd.TagesabschlussErstellen(ctx, 1, "Admin")
	if err != ErrKasseNichtGeoeffnet {
		t.Fatalf("expected ErrKasseNichtGeoeffnet, got %v", err)
	}
}
