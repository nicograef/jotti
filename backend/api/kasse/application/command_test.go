//go:build unit

package application

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	tseApp "github.com/nicograef/jotti/backend/api/tse/application"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
)

var testOpenKS = &kasse.Kassensitzung{
	ZNr:       1,
	Status:    kasse.KassensitzungOffen,
	CreatedAt: time.Now().UTC(),
	UpdatedAt: time.Now().UTC(),
}

type settingsMock struct {
	vereinsname string
	tse         settings.TSEKonfiguration
	tseErr      error
}

func (m settingsMock) GetBetreiber(_ context.Context) (settings.Betreiber, error) {
	return settings.Betreiber{
		Vereinsname: m.vereinsname,
		Strasse:     "Teststraße 1",
		Plz:         "12345",
		Ort:         "Teststadt",
		UpdatedAt:   time.Now(),
	}, nil
}

func (m settingsMock) GetTSEKonfiguration(_ context.Context) (settings.TSEKonfiguration, error) {
	if m.tseErr != nil {
		return settings.TSEKonfiguration{}, m.tseErr
	}
	if !m.tse.IstKonfiguriert() {
		return settings.TSEKonfiguration{}, db.ErrNotFound
	}
	return m.tse, nil
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

func TestTagesabschlussErstellen_DirektverkaufBlocksNever(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMock(nil, nil)

	// Kassensturz is still mandatory.
	ksSubject := kasse.KassensitzungSubject(testOpenKS.ZNr)
	kassensturzEvt, _ := kasse.NewKassensturzDurchgefuehrtEvent(ksSubject, 1, "Admin", 50000, 50000, 0)
	journalMock.AddEvent(kassensturzEvt)

	// Direktverkauf events are in separate streams and must not influence the
	// Tisch-Saldo-Sperre (no tisch session projection / no saldo to settle).
	dvSubject := kasse.DirektverkaufSubject(testOpenKS.ZNr, "verkauf-123")
	dvEvt, _ := event.New(1, "Admin", string(kasse.EventTypeDirektverkaufGetaetigtV1), dvSubject, map[string]any{
		"verkaufId":         "verkauf-123",
		"gesamtbetragCents": 900,
		"positionen":        []map[string]any{},
		"kommentar":         "",
	})
	journalMock.AddEvent(dvEvt)

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

func TestGeldtransitBuchen_MitTSE_DatenImEvent(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	start := time.Date(2026, 6, 10, 21, 10, 1, 0, time.UTC)
	end := time.Date(2026, 6, 10, 21, 10, 2, 0, time.UTC)

	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		SettingsRepo:        settingsMock{vereinsname: "TestVerein"},
		TSESignierer: tseApp.Signierer{
			SettingsRepo: settingsMock{tse: settings.TSEKonfiguration{
				ApiKey:    "api-key",
				ApiSecret: "api-secret",
				TssID:     "tss-1",
				ClientID:  "client-1",
				UpdatedAt: time.Now(),
			}},
			NewTSEClient: func(_ tse.Credentials) (tse.TSEClient, error) {
				return tse.FakeClient{
					StartResponse:  tse.StartResult{TransactionNumber: 41, LogTime: start, SerialNumberTSE: "TSE-SN-1", SignatureCounter: 40},
					FinishResponse: tse.FinishResult{TransactionNumber: 41, LogTimeStart: start, LogTimeEnd: end, LogTime: end, SignatureCounter: 41, SerialNumberTSE: "TSE-SN-1", Signature: "SIG-GELDTRANSIT"},
				}, nil
			},
		},
	}

	err := cmd.GeldtransitBuchen(ctx, 1, "Admin", "einlage", 1000, "Wechselgeld")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	events, err := journalMock.ReadEventsBySubject(ctx, kasse.KassensitzungSubject(testOpenKS.ZNr))
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected one event, got %d", len(events))
	}

	var data kasse.GeldtransitGebuchtV1Data
	if err := json.Unmarshal(events[0].Data, &data); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}
	if data.TSEData == nil {
		t.Fatal("expected TSE data in geldtransit event")
	}
	if data.TSEData.ProcessType != "Kassenbeleg-V1" {
		t.Fatalf("expected process type Kassenbeleg-V1, got %q", data.TSEData.ProcessType)
	}
}

func TestKassensturzDurchfuehren_MitTSE_SigniertNurDifferenzEvent(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	journalMock.SetKassenbestand(50000)
	start := time.Date(2026, 6, 10, 21, 20, 1, 0, time.UTC)
	end := time.Date(2026, 6, 10, 21, 20, 2, 0, time.UTC)

	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		SettingsRepo:        settingsMock{vereinsname: "TestVerein"},
		TSESignierer: tseApp.Signierer{
			SettingsRepo: settingsMock{tse: settings.TSEKonfiguration{
				ApiKey:    "api-key",
				ApiSecret: "api-secret",
				TssID:     "tss-1",
				ClientID:  "client-1",
				UpdatedAt: time.Now(),
			}},
			NewTSEClient: func(_ tse.Credentials) (tse.TSEClient, error) {
				return tse.FakeClient{
					StartResponse:  tse.StartResult{TransactionNumber: 51, LogTime: start, SerialNumberTSE: "TSE-SN-1", SignatureCounter: 50},
					FinishResponse: tse.FinishResult{TransactionNumber: 51, LogTimeStart: start, LogTimeEnd: end, LogTime: end, SignatureCounter: 51, SerialNumberTSE: "TSE-SN-1", Signature: "SIG-DIFFERENZ"},
				}, nil
			},
		},
	}

	err := cmd.KassensturzDurchfuehren(ctx, 1, "Admin", 49500)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	events, err := journalMock.ReadEventsBySubject(ctx, kasse.KassensitzungSubject(testOpenKS.ZNr))
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected two events (kassensturz + differenz), got %d", len(events))
	}
	if events[0].Type != string(kasse.EventTypeKassensturzDurchgefuehrtV1) {
		t.Fatalf("expected first event kassensturz, got %q", events[0].Type)
	}
	if events[1].Type != string(kasse.EventTypeDifferenzSollIstGebuchtV1) {
		t.Fatalf("expected second event differenz, got %q", events[1].Type)
	}

	var diff kasse.DifferenzSollIstGebuchtV1Data
	if err := json.Unmarshal(events[1].Data, &diff); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}
	if diff.TSEData == nil {
		t.Fatal("expected TSE data in differenz event")
	}
	if diff.TSEData.ProcessType != "Kassenbeleg-V1" {
		t.Fatalf("expected process type Kassenbeleg-V1, got %q", diff.TSEData.ProcessType)
	}
}

func TestTagesabschlussErstellen_MitTSE_DatenImEvent(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMock(nil, nil)

	ksSubject := kasse.KassensitzungSubject(testOpenKS.ZNr)
	kassensturzEvt, _ := kasse.NewKassensturzDurchgefuehrtEvent(ksSubject, 1, "Admin", 50000, 50000, 0)
	journalMock.AddEvent(kassensturzEvt)

	start := time.Date(2026, 6, 10, 21, 30, 1, 0, time.UTC)
	end := time.Date(2026, 6, 10, 21, 30, 2, 0, time.UTC)

	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		SettingsRepo:        settingsMock{vereinsname: "TestVerein"},
		TSESignierer: tseApp.Signierer{
			SettingsRepo: settingsMock{tse: settings.TSEKonfiguration{
				ApiKey:    "api-key",
				ApiSecret: "api-secret",
				TssID:     "tss-1",
				ClientID:  "client-1",
				UpdatedAt: time.Now(),
			}},
			NewTSEClient: func(_ tse.Credentials) (tse.TSEClient, error) {
				return tse.FakeClient{
					StartResponse:  tse.StartResult{TransactionNumber: 61, LogTime: start, SerialNumberTSE: "TSE-SN-1", SignatureCounter: 60},
					FinishResponse: tse.FinishResult{TransactionNumber: 61, LogTimeStart: start, LogTimeEnd: end, LogTime: end, SignatureCounter: 61, SerialNumberTSE: "TSE-SN-1", Signature: "SIG-TAGESABSCHLUSS"},
				}, nil
			},
		},
	}

	err := cmd.TagesabschlussErstellen(ctx, 1, "Admin")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	events, err := journalMock.ReadEventsBySubject(ctx, ksSubject)
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected two events, got %d", len(events))
	}
	if events[1].Type != string(kasse.EventTypeTagesabschlussErstelltV1) {
		t.Fatalf("expected second event tagesabschluss, got %q", events[1].Type)
	}

	var data kasse.TagesabschlussErstelltV1Data
	if err := json.Unmarshal(events[1].Data, &data); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}
	if data.TSEData == nil {
		t.Fatal("expected TSE data in tagesabschluss event")
	}
	if data.TSEData.ProcessType != "SonstigerVorgang" {
		t.Fatalf("expected process type SonstigerVorgang, got %q", data.TSEData.ProcessType)
	}
}

func TestKassensitzungEroeffnen_MitTSEKonfiguration_WirdNichtSigniert(t *testing.T) {
	ctx := context.Background()
	tseClientCalled := false

	cmd := Command{
		KassenjournalRepo:   kassenjournal_repo.NewMock(nil, nil),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(nil, nil),
		SettingsRepo:        settingsMock{vereinsname: "TestVerein"},
		TSESignierer: tseApp.Signierer{
			SettingsRepo: settingsMock{tse: settings.TSEKonfiguration{
				ApiKey:    "api-key",
				ApiSecret: "api-secret",
				TssID:     "tss-1",
				ClientID:  "client-1",
				UpdatedAt: time.Now(),
			}},
			NewTSEClient: func(_ tse.Credentials) (tse.TSEClient, error) {
				tseClientCalled = true
				return tse.FakeClient{}, nil
			},
		},
	}

	_, err := cmd.KassensitzungEroeffnen(ctx, 1, "Admin", "Vereinsfest", 10000)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tseClientCalled {
		t.Fatal("expected TSE client to not be created for kassensitzung-eroeffnet")
	}
}
