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

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/betreiber"
	e "github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
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
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	sitzungMock := kassensitzungen_repo.NewMock(ks, nil)
	return Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: sitzungMock,
		BetreiberRepo:       settingsMock{vereinsname: "TestVerein"},
		TSERepo:             tseGateMock{},
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
		BetreiberRepo:       settingsMock{betreiberErr: db.ErrNotFound},
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
		BetreiberRepo:       settingsMock{betreiberErr: db.ErrDatabase},
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

	err := cmd.GeldtransitBuchen(ctx, 1, "Admin", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", "einlage", 10000, "Wechselgeld nachgelegt")
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

		TSERepo: tseGateMock{},
	}

	_, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000)
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

		TSERepo: tseGateMock{},
	}

	_, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 49500) // Ist = 495 EUR, Differenz = 500
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

// stubCleaner ist eine Test-Doublette fuer druckauftragCleaner: zaehlt die
// Aufrufe und liefert einen konfigurierbaren Fehler, um die Best-effort-Semantik
// des Aufraeumens beim Tagesabschluss zu belegen.
type stubCleaner struct {
	calls int
	err   error
}

func (s *stubCleaner) DiscardAlleFehlgeschlagenen(context.Context) (int64, error) {
	s.calls++
	return 0, s.err
}

func TestKasseAbschliessen_RaeumtFehlgeschlageneDruckauftraegeAuf(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	journalMock.SetKassenbestand(50000) // Soll = Ist
	cleaner := &stubCleaner{}
	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		TSERepo:             tseGateMock{},
		DruckauftragRepo:    cleaner,
	}

	_, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cleaner.calls != 1 {
		t.Fatalf("expected DiscardAlleFehlgeschlagenen called once, got %d", cleaner.calls)
	}

	events, err := journalMock.ReadEventsBySubject(ctx, kasse.KassensitzungSubject(testOpenKS.ZNr))
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected two events (kassensturz + tagesabschluss), got %d", len(events))
	}
	if events[1].Type != string(kasse.EventTypeTagesabschlussErstelltV1) {
		t.Fatalf("expected second event tagesabschluss, got %q", events[1].Type)
	}
}

func TestKasseAbschliessen_CleanerFehlerBleibtBestEffort(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	journalMock.SetKassenbestand(50000) // Soll = Ist
	cleaner := &stubCleaner{err: fmt.Errorf("cleanup kaputt")}
	sitzungMock := kassensitzungen_repo.NewMock(testOpenKS, nil)
	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: sitzungMock,
		TSERepo:             tseGateMock{},
		DruckauftragRepo:    cleaner,
	}

	ergebnis, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000)
	if err != nil {
		t.Fatalf("expected Abschluss to stay successful despite cleaner error, got %v", err)
	}
	_ = ergebnis
	if cleaner.calls != 1 {
		t.Fatalf("expected DiscardAlleFehlgeschlagenen called once, got %d", cleaner.calls)
	}

	// Kern der Best-effort-Invariante: Der Cleaner-Fehler wird über eine lokale
	// Variable geschluckt, nicht über den benannten Return err. Sonst würde der
	// defer-Block die bereits geschlossene Sitzung fälschlich auf 'offen'
	// zurücksetzen. Kein Reset ist der Beleg, dass der Abschluss endgültig bleibt.
	if sitzungMock.OffenCalls != 0 {
		t.Fatalf("expected NO reset to offen after cleaner error (Abschluss ist endgueltig), got %d", sitzungMock.OffenCalls)
	}

	events, err := journalMock.ReadEventsBySubject(ctx, kasse.KassensitzungSubject(testOpenKS.ZNr))
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected two events (kassensturz + tagesabschluss) despite cleaner error, got %d", len(events))
	}
	if events[1].Type != string(kasse.EventTypeTagesabschlussErstelltV1) {
		t.Fatalf("expected second event tagesabschluss, got %q", events[1].Type)
	}
}

// TestKasseAbschliessen_TagesabschlussMitEchtenSummen prüft, dass die drei Summen
// im tagesabschluss-erstellt-Event aus den Journal-Events der Kassensitzung berechnet
// werden (und nicht mehr aus einem separaten Reporting-Repository).
// Der Journal-Mock liefert dabei auch die im selben Vorgang geschriebenen
// Kassensturz-Events, da sie zum Lesezeitpunkt committed sind.
func TestKasseAbschliessen_TagesabschlussMitEchtenSummen(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	journalMock.SetKassenbestand(50000)

	// Pre-load events that ComputeAbschlussSummen should aggregate.
	// zahlung: umsatz +12345
	zahlungRaw, _ := json.Marshal(map[string]int{"gesamtZahlungCents": 12345})
	journalMock.AddEvent(e.Event{
		UserID:   1,
		UserName: "Test",
		Type:     string(kasse.EventTypeZahlungKassiertV1),
		Subject:  kasse.TischSessionSubject(testOpenKS.ZNr, 99),
		Data:     zahlungRaw,
	})
	// korrektur (geldneutral): stornierungen +200, umsatz unverändert
	korrekturRaw, _ := json.Marshal(map[string]int{"gesamtCents": 200})
	journalMock.AddEvent(e.Event{
		UserID:   1,
		UserName: "Test",
		Type:     string(kasse.EventTypeBestellungKorrigiertV1),
		Subject:  kasse.TischSessionSubject(testOpenKS.ZNr, 99),
		Data:     korrekturRaw,
	})
	// geldtransit einlage: geldtransit +400
	transitRaw, _ := json.Marshal(map[string]interface{}{"richtung": "einlage", "betragCents": 400})
	journalMock.AddEvent(e.Event{
		UserID:   1,
		UserName: "Test",
		Type:     string(kasse.EventTypeGeldtransitGebuchtV1),
		Subject:  kasse.KassensitzungSubject(testOpenKS.ZNr),
		Data:     transitRaw,
	})

	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		TSERepo:             tseGateMock{},
	}

	_, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	allEvents, err := journalMock.ReadKassensitzungEvents(ctx, testOpenKS.ZNr)
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	tagesabschluss := allEvents[len(allEvents)-1]

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

		TSERepo: tseGateMock{},
	}

	_, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000)
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

	_, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000)
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

		TSERepo: tseGateMock{},
	}

	if _, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000); err != nil {
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
	// ReadKassensitzungEvents scheitert nach der Barriere und den Kassensturz-Writes.
	journalMock.SetReadKassensitzungEventsErr(db.ErrDatabase)
	sitzungMock := kassensitzungen_repo.NewMock(testOpenKS, nil)
	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: sitzungMock,
		TSERepo:             tseGateMock{},
	}

	if _, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000); err == nil {
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

		TSERepo: tseGateMock{},
	}

	if _, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000); err != ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
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

		TSERepo: tseGateMock{},
	}

	if _, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000); err != ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
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

		TSERepo: tseGateMock{},
	}

	if _, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000); err != nil {
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

// Wiederanlauf nach Teilfehler: Steht der Kassensturz eines abgebrochenen früheren
// Versuchs bereits im Journal, wird Schritt 1 übersprungen — es entsteht kein zweites
// kassensturz-Event, und die Differenz rechnet gegen den dort dokumentierten Ist-Bestand.
func TestKasseAbschliessen_WiederanlaufSchreibtKeinenZweitenKassensturz(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	// Die Differenzbuchung des ersten Versuchs kam nicht durch: Soll steht noch auf 50000.
	journalMock.SetKassenbestand(50000)

	sturzRaw, err := json.Marshal(kasse.KassensturzDurchgefuehrtV1Data{
		SollBestandCents: 50000,
		IstBestandCents:  49500,
		DifferenzCents:   500,
		DurchgefuehrtVon: 1,
	})
	if err != nil {
		t.Fatalf("marshal kassensturz data: %v", err)
	}
	journalMock.AddEvent(e.Event{
		UserID:   1,
		UserName: "Admin",
		Type:     string(kasse.EventTypeKassensturzDurchgefuehrtV1),
		Subject:  kasse.KassensitzungSubject(testOpenKS.ZNr),
		Version:  1,
		Data:     sturzRaw,
	})

	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		TSERepo:             tseGateMock{},
	}

	if _, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 49500); err != nil {
		t.Fatalf("expected retry to succeed, got %v", err)
	}

	events, err := journalMock.ReadEventsBySubject(ctx, kasse.KassensitzungSubject(testOpenKS.ZNr))
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected three events (vorhandener kassensturz + differenz + tagesabschluss), got %d", len(events))
	}
	if events[0].Type != string(kasse.EventTypeKassensturzDurchgefuehrtV1) {
		t.Fatalf("expected first event kassensturz, got %q", events[0].Type)
	}
	if events[1].Type != string(kasse.EventTypeDifferenzSollIstGebuchtV1) {
		t.Fatalf("expected second event differenz, got %q", events[1].Type)
	}
	var differenzData kasse.DifferenzSollIstGebuchtV1Data
	if err := json.Unmarshal(events[1].Data, &differenzData); err != nil {
		t.Fatalf("unmarshal differenz data: %v", err)
	}
	if differenzData.BetragCents != 500 {
		t.Errorf("expected differenz 500 gegen den dokumentierten Ist-Bestand, got %d", differenzData.BetragCents)
	}
	if events[2].Type != string(kasse.EventTypeTagesabschlussErstelltV1) {
		t.Fatalf("expected third event tagesabschluss, got %q", events[2].Type)
	}
}

// Wiederanlauf mit Zwischenbuchung: Steht nach dem protokollierten Kassensturz eine echte
// Buchung (geldtransit-gebucht:v1) im Stream, bricht der Abschluss mit ErrBuchungenNachKassensturz
// ab, statt den veralteten Ist-Bestand zu übernehmen — es wird kein Abschluss-Event geschrieben.
func TestKasseAbschliessen_WiederanlaufMitZwischenbuchungBrichtAb(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	journalMock.SetKassenbestand(50000)

	sturzRaw, err := json.Marshal(kasse.KassensturzDurchgefuehrtV1Data{
		SollBestandCents: 50000,
		IstBestandCents:  49500,
		DifferenzCents:   500,
		DurchgefuehrtVon: 1,
	})
	if err != nil {
		t.Fatalf("marshal kassensturz data: %v", err)
	}
	journalMock.AddEvent(e.Event{
		UserID:   1,
		UserName: "Admin",
		Type:     string(kasse.EventTypeKassensturzDurchgefuehrtV1),
		Subject:  kasse.KassensitzungSubject(testOpenKS.ZNr),
		Version:  1,
		Data:     sturzRaw,
	})
	// Zwischenbuchung nach dem Kassensturz: ein Geldtransit im Kassensitzungs-Stream.
	transitRaw, err := json.Marshal(map[string]interface{}{"richtung": "einlage", "betragCents": 400})
	if err != nil {
		t.Fatalf("marshal geldtransit data: %v", err)
	}
	journalMock.AddEvent(e.Event{
		UserID:   1,
		UserName: "Admin",
		Type:     string(kasse.EventTypeGeldtransitGebuchtV1),
		Subject:  kasse.KassensitzungSubject(testOpenKS.ZNr),
		Version:  2,
		Data:     transitRaw,
	})

	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		TSERepo:             tseGateMock{},
	}

	if _, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 49500); err != ErrBuchungenNachKassensturz {
		t.Fatalf("expected ErrBuchungenNachKassensturz, got %v", err)
	}

	events, err := journalMock.ReadEventsBySubject(ctx, kasse.KassensitzungSubject(testOpenKS.ZNr))
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	for _, evt := range events {
		if evt.Type == string(kasse.EventTypeDifferenzSollIstGebuchtV1) || evt.Type == string(kasse.EventTypeTagesabschlussErstelltV1) {
			t.Errorf("no closing event must be written on conflict, got %q", evt.Type)
		}
	}
}

// Eine Buchung in eine Sitzung im Zwischenstatus wird mit ErrKasseWirdAbgeschlossen abgelehnt.
func TestGeldtransitBuchen_WirdAbgeschlossen(t *testing.T) {
	ctx := context.Background()
	imAbschluss := &kasse.Kassensitzung{ZNr: 1, Status: kasse.KassensitzungWirdAbgeschlossen, CreatedAt: time.Now().UTC()}
	cmd := newTestCommand(imAbschluss)

	err := cmd.GeldtransitBuchen(ctx, 1, "Admin", "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", "einlage", 1000, "Wechselgeld")
	if err != ErrKasseWirdAbgeschlossen {
		t.Fatalf("expected ErrKasseWirdAbgeschlossen, got %v", err)
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

	tseKonfiguration := tse.Konfiguration{
		ApiKey:    "api-key",
		ApiSecret: "api-secret",
		TssID:     "tss-1",
		ClientID:  "client-1",
		UpdatedAt: time.Now(),
	}
	cmd := Command{
		KassenjournalRepo:   kassenjournal_repo.NewMock(nil, nil),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(nil, nil),
		BetreiberRepo:       settingsMock{vereinsname: "TestVerein"},
		TSERepo:             tseGateMock{tse: tseKonfiguration},
	}

	_, err := cmd.KassensitzungEroeffnen(ctx, 1, "Admin", "Vereinsfest 2026", 10000)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if strings.Contains(logbuf.String(), "ohne TSE-Konfiguration") {
		t.Errorf("expected no TSE warning, got logs: %s", logbuf.String())
	}
}

// TestKasseAbschliessen_KorruptesEventBrichtAbschlussAb belegt, dass ein nicht
// parsebares summen-wirksames Event den Tagesabschluss abbricht und kein
// tagesabschluss-erstellt-Event schreibt. Inkorrekte Summen im Z-Bon sind
// schlimmer als ein blockierter Abschluss; praktisch nur bei einem korrupten
// Store erreichbar.
func TestKasseAbschliessen_KorruptesEventBrichtAbschlussAb(t *testing.T) {
	ctx := context.Background()
	journalMock := kassenjournal_repo.NewMock(nil, nil)
	journalMock.SetKassenbestand(50000)

	// Korruptes zahlung-kassiert-Event mit ungültigem JSON.
	journalMock.AddEvent(e.Event{
		UserID:   1,
		UserName: "Test",
		Type:     string(kasse.EventTypeZahlungKassiertV1),
		Subject:  kasse.TischSessionSubject(testOpenKS.ZNr, 99),
		Data:     json.RawMessage(`{"fehlerhaft": true`), // ungültiges JSON
	})

	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		TSERepo:             tseGateMock{},
	}

	_, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000)
	if err == nil {
		t.Fatal("expected error for corrupt event, got nil")
	}

	// Kein tagesabschluss-Event darf trotz des Fehlers geschrieben worden sein.
	allEvents, readErr := journalMock.ReadKassensitzungEvents(ctx, testOpenKS.ZNr)
	if readErr != nil {
		t.Fatalf("expected no read error, got %v", readErr)
	}
	for _, evt := range allEvents {
		if evt.Type == string(kasse.EventTypeTagesabschlussErstelltV1) {
			t.Errorf("tagesabschluss-Event wurde trotz korruptem Event geschrieben")
		}
	}
}
