//go:build unit

package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	tseApp "github.com/nicograef/jotti/backend/api/tse/application"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/reporting"
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/rs/zerolog"
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
	tse          settings.TSEKonfiguration
	tseErr       error
}

func (m settingsMock) GetBetreiber(_ context.Context) (settings.Betreiber, error) {
	if m.betreiberErr != nil {
		return settings.Betreiber{}, m.betreiberErr
	}
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

type reportingMock struct {
	data reporting.ReportingData
	err  error
}

func (m reportingMock) GetReporting(_ context.Context, _ int) (reporting.ReportingData, error) {
	return m.data, m.err
}

func newTestCommand(ks *kasse.Kassensitzung) Command {
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	sitzungMock := kassensitzungen_repo.NewMock(ks, nil)
	return Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: sitzungMock,
		SettingsRepo:        settingsMock{vereinsname: "TestVerein"},
		ReportingRepo:       reportingMock{},
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
		KassenjournalRepo:   kassenjournal_repo.NewMock(nil, nil),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(nil, nil),
		SettingsRepo:        settingsMock{betreiberErr: db.ErrNotFound},
	}

	_, err := cmd.KassensitzungEroeffnen(ctx, 1, "Admin", "Vereinsfest 2026", 10000)
	if err != ErrBetreiberNichtKonfiguriert {
		t.Fatalf("expected ErrBetreiberNichtKonfiguriert, got %v", err)
	}
}

func TestKassensitzungEroeffnen_BetreiberDatabaseError(t *testing.T) {
	ctx := context.Background()
	cmd := Command{
		KassenjournalRepo:   kassenjournal_repo.NewMock(nil, nil),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(nil, nil),
		SettingsRepo:        settingsMock{betreiberErr: db.ErrDatabase},
	}

	_, err := cmd.KassensitzungEroeffnen(ctx, 1, "Admin", "Vereinsfest 2026", 10000)
	if err != ErrDatabase {
		t.Fatalf("expected ErrDatabase, got %v", err)
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

func TestKasseAbschliessen_OhneDifferenz(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	journalMock.SetKassenbestand(50000) // Soll = Ist
	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		ReportingRepo:       reportingMock{},
	}

	err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	events, err := journalMock.ReadEventsBySubject(ctx, kasse.KassensitzungSubject(testOpenKS.ZNr))
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected two events (kassensturz + tagesabschluss), got %d", len(events))
	}
	if events[0].Type != string(kasse.EventTypeKassensturzDurchgefuehrtV1) {
		t.Fatalf("expected first event kassensturz, got %q", events[0].Type)
	}
	if events[1].Type != string(kasse.EventTypeTagesabschlussErstelltV1) {
		t.Fatalf("expected second event tagesabschluss, got %q", events[1].Type)
	}
}

func TestKasseAbschliessen_MitDifferenz(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	journalMock.SetKassenbestand(50000) // Soll = 500 EUR
	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		ReportingRepo:       reportingMock{},
	}

	err := cmd.KasseAbschliessen(ctx, 1, "Admin", 49500) // Ist = 495 EUR, Differenz = 500
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	events, err := journalMock.ReadEventsBySubject(ctx, kasse.KassensitzungSubject(testOpenKS.ZNr))
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected three events (kassensturz + differenz + tagesabschluss), got %d", len(events))
	}
	if events[0].Type != string(kasse.EventTypeKassensturzDurchgefuehrtV1) {
		t.Fatalf("expected first event kassensturz, got %q", events[0].Type)
	}
	if events[1].Type != string(kasse.EventTypeDifferenzSollIstGebuchtV1) {
		t.Fatalf("expected second event differenz, got %q", events[1].Type)
	}
	if events[2].Type != string(kasse.EventTypeTagesabschlussErstelltV1) {
		t.Fatalf("expected third event tagesabschluss, got %q", events[2].Type)
	}
}

func TestKasseAbschliessen_TagesabschlussMitEchtenSummen(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	journalMock.SetKassenbestand(50000)
	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		ReportingRepo: reportingMock{data: reporting.ReportingData{
			Summary: reporting.Summary{
				GesamtUmsatzCents:        12345,
				GesamtStornierungenCents: 200,
				GeldtransitCents:         400,
			},
		}},
	}

	err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	events, err := journalMock.ReadEventsBySubject(ctx, kasse.KassensitzungSubject(testOpenKS.ZNr))
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	tagesabschluss := events[len(events)-1]

	var data kasse.TagesabschlussErstelltV1Data
	if err := json.Unmarshal(tagesabschluss.Data, &data); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}
	if data.UmsatzGesamtCents != 12345 {
		t.Errorf("expected UmsatzGesamtCents 12345, got %d", data.UmsatzGesamtCents)
	}
	if data.StornierungCents != 200 {
		t.Errorf("expected StornierungCents 200, got %d", data.StornierungCents)
	}
	if data.GeldtransitCents != 400 {
		t.Errorf("expected GeldtransitCents 400, got %d", data.GeldtransitCents)
	}
}

func TestKasseAbschliessen_TischSaldoSperre(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	journalMock.SetKassenbestand(50000)

	// A tisch session with non-zero saldo must block the Kassenabschluss before any event is written.
	tischSubject := kasse.TischSessionSubject(testOpenKS.ZNr, 42)
	journalMock.SetTischSession(tischSubject, kasse.TischSession{
		TischID:         42,
		KassensitzungNr: testOpenKS.ZNr,
		SaldoCents:      350,
	})

	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		ReportingRepo:       reportingMock{},
	}

	err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000)
	if err != ErrTischeSaldoOffen {
		t.Fatalf("expected ErrTischeSaldoOffen, got %v", err)
	}

	events, err := journalMock.ReadEventsBySubject(ctx, kasse.KassensitzungSubject(testOpenKS.ZNr))
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected no events written when saldo blocks, got %d", len(events))
	}
}

func TestKasseAbschliessen_KasseNichtGeoeffnet(t *testing.T) {
	ctx := context.Background()
	cmd := newTestCommand(nil)

	err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000)
	if err != ErrKasseNichtGeoeffnet {
		t.Fatalf("expected ErrKasseNichtGeoeffnet, got %v", err)
	}
}

// Phase 1 des Abschlusses setzt die Barriere ('wird_abgeschlossen') als ersten Schritt und
// setzt sie bei Erfolg nicht zurück.
func TestKasseAbschliessen_SetztBarriere(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	journalMock.SetKassenbestand(50000)
	sitzungMock := kassensitzungen_repo.NewMock(testOpenKS, nil)
	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: sitzungMock,
		ReportingRepo:       reportingMock{},
	}

	if err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if sitzungMock.WirdAbgeschlossenCalls != 1 {
		t.Fatalf("expected barrier to be set once, got %d", sitzungMock.WirdAbgeschlossenCalls)
	}
	if sitzungMock.OffenCalls != 0 {
		t.Fatalf("expected no reset on success, got %d", sitzungMock.OffenCalls)
	}
}

// Schlägt der Abschluss nach dem Statuswechsel fehl, wird die Sitzung best effort auf 'offen'
// zurückgesetzt, damit sie nicht im Zwischenstatus hängen bleibt.
func TestKasseAbschliessen_FehlerSetztStatusZurueck(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	journalMock.SetKassenbestand(50000)
	sitzungMock := kassensitzungen_repo.NewMock(testOpenKS, nil)
	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: sitzungMock,
		ReportingRepo:       reportingMock{err: db.ErrDatabase}, // Reporting scheitert nach der Barriere
	}

	if err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000); err == nil {
		t.Fatal("expected an error, got nil")
	}
	if sitzungMock.WirdAbgeschlossenCalls != 1 {
		t.Fatalf("expected barrier to be set once, got %d", sitzungMock.WirdAbgeschlossenCalls)
	}
	if sitzungMock.OffenCalls != 1 {
		t.Fatalf("expected status reset to offen after error, got %d", sitzungMock.OffenCalls)
	}
}

// Ein Versionskonflikt bedeutet einen konkurrierenden zweiten Abschluss — die unterlegene
// Instanz darf die Barriere nicht unter dem gewinnenden Abschluss wegräumen.
func TestKasseAbschliessen_KonfliktSetztStatusNichtZurueck(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMockWithWriteErr(nil, db.ErrAlreadyExists)
	journalMock.SetKassenbestand(50000)
	sitzungMock := kassensitzungen_repo.NewMock(testOpenKS, nil)
	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: sitzungMock,
		ReportingRepo:       reportingMock{},
	}

	if err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000); err != ErrKonflikt {
		t.Fatalf("expected ErrKonflikt, got %v", err)
	}
	if sitzungMock.OffenCalls != 0 {
		t.Fatalf("expected NO reset on conflict, got %d", sitzungMock.OffenCalls)
	}
}

// Ein Deadlock (40P01) beim Event-Write wird wie ein Konflikt behandelt (409 statt 500);
// die Barriere bleibt stehen, ein erneuter Aufruf setzt den Abschluss fort.
func TestKasseAbschliessen_DeadlockMapsToKonflikt(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMockWithWriteErr(nil, db.ErrConflict)
	journalMock.SetKassenbestand(50000)
	sitzungMock := kassensitzungen_repo.NewMock(testOpenKS, nil)
	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: sitzungMock,
		ReportingRepo:       reportingMock{},
	}

	if err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000); err != ErrKonflikt {
		t.Fatalf("expected ErrKonflikt, got %v", err)
	}
}

// Wiederanlauf: Steht die Sitzung nach einem Abbruch noch auf 'wird_abgeschlossen', schließt
// ein erneuter Aufruf sie erfolgreich ab.
func TestKasseAbschliessen_WiederanlaufImZwischenstatus(t *testing.T) {
	ctx := context.Background()
	imAbschluss := &kasse.Kassensitzung{ZNr: 1, Status: kasse.KassensitzungWirdAbgeschlossen, CreatedAt: time.Now().UTC()}
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	journalMock.SetKassenbestand(50000)
	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(imAbschluss, nil),
		ReportingRepo:       reportingMock{},
	}

	if err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000); err != nil {
		t.Fatalf("expected resume to succeed, got %v", err)
	}

	events, err := journalMock.ReadEventsBySubject(ctx, kasse.KassensitzungSubject(imAbschluss.ZNr))
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected kassensturz + tagesabschluss, got %d", len(events))
	}
}

// Eine Buchung in eine Sitzung im Zwischenstatus wird mit ErrKasseWirdAbgeschlossen abgelehnt.
func TestGeldtransitBuchen_WirdAbgeschlossen(t *testing.T) {
	ctx := context.Background()
	imAbschluss := &kasse.Kassensitzung{ZNr: 1, Status: kasse.KassensitzungWirdAbgeschlossen, CreatedAt: time.Now().UTC()}
	cmd := newTestCommand(imAbschluss)

	err := cmd.GeldtransitBuchen(ctx, 1, "Admin", "einlage", 1000, "Wechselgeld")
	if err != ErrKasseWirdAbgeschlossen {
		t.Fatalf("expected ErrKasseWirdAbgeschlossen, got %v", err)
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

func TestKasseAbschliessen_MitTSE_SigniertDifferenzUndTagesabschluss(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	journalMock.SetKassenbestand(50000)
	start := time.Date(2026, 6, 10, 21, 20, 1, 0, time.UTC)
	end := time.Date(2026, 6, 10, 21, 20, 2, 0, time.UTC)

	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		SettingsRepo:        settingsMock{vereinsname: "TestVerein"},
		ReportingRepo:       reportingMock{},
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
					FinishResponse: tse.FinishResult{TransactionNumber: 51, LogTimeStart: start, LogTimeEnd: end, LogTime: end, SignatureCounter: 51, SerialNumberTSE: "TSE-SN-1", Signature: "SIG"},
				}, nil
			},
		},
	}

	// Ist != Soll erzwingt die Differenzbuchung, sodass beide signierungspflichtigen Events entstehen.
	err := cmd.KasseAbschliessen(ctx, 1, "Admin", 49500)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	events, err := journalMock.ReadEventsBySubject(ctx, kasse.KassensitzungSubject(testOpenKS.ZNr))
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected three events (kassensturz + differenz + tagesabschluss), got %d", len(events))
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

	var tagesabschluss kasse.TagesabschlussErstelltV1Data
	if err := json.Unmarshal(events[2].Data, &tagesabschluss); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}
	if tagesabschluss.TSEData == nil {
		t.Fatal("expected TSE data in tagesabschluss event")
	}
	if tagesabschluss.TSEData.ProcessType != "SonstigerVorgang" {
		t.Fatalf("expected process type SonstigerVorgang, got %q", tagesabschluss.TSEData.ProcessType)
	}
}

// Der Betriebstag ist das Wandkalenderdatum in Europe/Berlin — auch kurz nach
// Mitternacht (Sommer- wie Winterzeit), wo UTC noch der Vortag ist.
func TestBetriebstag_NachMitternachtOrtszeit(t *testing.T) {
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}

	cases := []struct {
		name string
		now  time.Time
		want string
	}{
		{"Sommerzeit 00:30 Berlin (22:30 UTC Vortag)", time.Date(2026, 7, 4, 22, 30, 0, 0, time.UTC), "2026-07-05"},
		{"Winterzeit 00:30 Berlin (23:30 UTC Vortag)", time.Date(2026, 1, 9, 23, 30, 0, 0, time.UTC), "2026-01-10"},
		{"mittags", time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC), "2026-07-04"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := betriebstag(tc.now, berlin).Format("2006-01-02"); got != tc.want {
				t.Errorf("betriebstag = %s, want %s", got, tc.want)
			}
		})
	}
}

// Der Anfangsbestand (> 0) ist eine Bareinlage und wird wie Geldtransit als
// Kassenbeleg-V1-Eigenbeleg signiert.
func TestKassensitzungEroeffnen_MitTSE_SigniertAnfangsbestand(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	start := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 10, 9, 0, 1, 0, time.UTC)

	cmd := Command{
		KassenjournalRepo:   journalMock,
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
				return tse.FakeClient{
					StartResponse:  tse.StartResult{TransactionNumber: 60, LogTime: start, SerialNumberTSE: "TSE-SN-1", SignatureCounter: 59},
					FinishResponse: tse.FinishResult{TransactionNumber: 60, LogTimeStart: start, LogTimeEnd: end, LogTime: end, SignatureCounter: 60, SerialNumberTSE: "TSE-SN-1", Signature: "SIG"},
				}, nil
			},
		},
	}

	zNr, err := cmd.KassensitzungEroeffnen(ctx, 1, "Admin", "Vereinsfest", 10000)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	events, err := journalMock.ReadEventsBySubject(ctx, kasse.KassensitzungSubject(zNr))
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected one event, got %d", len(events))
	}

	var data kasse.KassensitzungEroeffnetV1Data
	if err := json.Unmarshal(events[0].Data, &data); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}
	if data.TSEData == nil {
		t.Fatal("expected TSE data in eroeffnet event (Anfangsbestand ist eine Bareinlage)")
	}
	if data.TSEData.ProcessType != "Kassenbeleg-V1" {
		t.Fatalf("expected process type Kassenbeleg-V1, got %q", data.TSEData.ProcessType)
	}
}

// Ohne Anfangsbestand gibt es keinen Geschäftsvorfall — keine TSE-Transaktion.
// (Das Event-Schema lehnt betragCents = 0 derzeit ohnehin ab; der Test dokumentiert,
// dass der Signierpfad in diesem Fall nie erreicht wird.)
func TestKassensitzungEroeffnen_OhneAnfangsbestand_WirdNichtSigniert(t *testing.T) {
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

	_, err := cmd.KassensitzungEroeffnen(ctx, 1, "Admin", "Vereinsfest", 0)
	if err == nil {
		t.Fatal("expected validation error for betragCents = 0 (Schema verlangt Anfangsbestand)")
	}
	if tseClientCalled {
		t.Fatal("expected TSE client to not be created for kassensitzung-eroeffnet without Anfangsbestand")
	}
}

// Eröffnen ohne konfigurierte TSE wird nicht gesperrt, aber im Log vermerkt (F6).
func TestKassensitzungEroeffnen_OhneTSE_LoggtWarnung(t *testing.T) {
	var logbuf bytes.Buffer
	ctx := zerolog.New(&logbuf).WithContext(context.Background())
	cmd := newTestCommand(nil) // settingsMock ohne TSE-Konfiguration

	zNr, err := cmd.KassensitzungEroeffnen(ctx, 1, "Admin", "Vereinsfest 2026", 10000)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var warnLine string
	for _, line := range strings.Split(strings.TrimSpace(logbuf.String()), "\n") {
		if strings.Contains(line, `"level":"warn"`) && strings.Contains(line, "ohne TSE-Konfiguration") {
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

// Mit konfigurierter TSE bleibt das Eröffnen warnungsfrei.
func TestKassensitzungEroeffnen_MitTSE_KeineWarnung(t *testing.T) {
	var logbuf bytes.Buffer
	ctx := zerolog.New(&logbuf).WithContext(context.Background())
	start := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 10, 9, 0, 1, 0, time.UTC)

	tseSettings := settingsMock{vereinsname: "TestVerein", tse: settings.TSEKonfiguration{
		ApiKey:    "api-key",
		ApiSecret: "api-secret",
		TssID:     "tss-1",
		ClientID:  "client-1",
		UpdatedAt: time.Now(),
	}}
	cmd := Command{
		KassenjournalRepo:   kassenjournal_repo.NewMock(nil, nil),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(nil, nil),
		SettingsRepo:        tseSettings,
		TSESignierer: tseApp.Signierer{
			SettingsRepo: tseSettings,
			NewTSEClient: func(_ tse.Credentials) (tse.TSEClient, error) {
				return tse.FakeClient{
					StartResponse:  tse.StartResult{TransactionNumber: 60, LogTime: start, SerialNumberTSE: "TSE-SN-1", SignatureCounter: 59},
					FinishResponse: tse.FinishResult{TransactionNumber: 60, LogTimeStart: start, LogTimeEnd: end, LogTime: end, SignatureCounter: 60, SerialNumberTSE: "TSE-SN-1", Signature: "SIG"},
				}, nil
			},
		},
	}

	_, err := cmd.KassensitzungEroeffnen(ctx, 1, "Admin", "Vereinsfest 2026", 10000)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if strings.Contains(logbuf.String(), "ohne TSE-Konfiguration") {
		t.Errorf("expected no TSE warning, got logs: %s", logbuf.String())
	}
}
