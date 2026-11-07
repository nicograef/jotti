package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/db"
	e "github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/repository/repotest"
)

func assertSitzungStatus(t *testing.T, repo *repotest.KassensitzungenRepo, want kasse.KassensitzungStatus) {
	t.Helper()
	ks, err := repo.GetAktiveKassensitzung(context.Background())
	if err != nil || ks == nil {
		t.Fatalf("expected an active Kassensitzung, got %v (err %v)", ks, err)
	}
	if ks.Status != want {
		t.Errorf("Kassensitzung status = %q, want %q", ks.Status, want)
	}
}

func TestKasseAbschliessen_OhneDifferenz(t *testing.T) {
	ctx := context.Background()
	journalMock := repotest.NewKassenjournalRepo(nil, nil)
	journalMock.SetKassenbestand(50000) // Soll = Ist
	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
		TSERepo:             tseGateMock{},
		DruckauftragRepo:    &fakeDruckauftraege{},
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
		t.Errorf("expected first event kassensturz, got %q", events[0].Type)
	}
	if events[1].Type != string(kasse.EventTypeTagesabschlussErstelltV1) {
		t.Errorf("expected second event tagesabschluss, got %q", events[1].Type)
	}
}

func TestKasseAbschliessen_MitDifferenz(t *testing.T) {
	ctx := context.Background()
	journalMock := repotest.NewKassenjournalRepo(nil, nil)
	journalMock.SetKassenbestand(50000) // Soll = 500 EUR
	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
		TSERepo:             tseGateMock{},
		DruckauftragRepo:    &fakeDruckauftraege{},
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
		t.Errorf("expected first event kassensturz, got %q", events[0].Type)
	}
	if events[1].Type != string(kasse.EventTypeDifferenzSollIstGebuchtV1) {
		t.Errorf("expected second event differenz, got %q", events[1].Type)
	}
	if events[2].Type != string(kasse.EventTypeTagesabschlussErstelltV1) {
		t.Errorf("expected third event tagesabschluss, got %q", events[2].Type)
	}
}

// fakeDruckauftraege holds the number of failed Druckaufträge the Tagesabschluss discards.
// A set err makes the discard fail and leaves the count unchanged.
type fakeDruckauftraege struct {
	fehlgeschlagen int64
	err            error
	versuche       int
}

func (f *fakeDruckauftraege) DiscardAlleFehlgeschlagenen(context.Context) (int64, error) {
	f.versuche++
	if f.err != nil {
		return 0, f.err
	}
	verworfen := f.fehlgeschlagen
	f.fehlgeschlagen = 0
	return verworfen, nil
}

func TestKasseAbschliessen_RaeumtFehlgeschlageneDruckauftraegeAuf(t *testing.T) {
	ctx := context.Background()
	journalMock := repotest.NewKassenjournalRepo(nil, nil)
	journalMock.SetKassenbestand(50000) // Soll = Ist
	cleaner := &fakeDruckauftraege{fehlgeschlagen: 2}
	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
		TSERepo:             tseGateMock{},
		DruckauftragRepo:    cleaner,
	}

	_, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cleaner.fehlgeschlagen != 0 {
		t.Errorf("expected the failed Druckaufträge to be discarded, %d remain", cleaner.fehlgeschlagen)
	}

	events, err := journalMock.ReadEventsBySubject(ctx, kasse.KassensitzungSubject(testOpenKS.ZNr))
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected two events (kassensturz + tagesabschluss), got %d", len(events))
	}
	if events[1].Type != string(kasse.EventTypeTagesabschlussErstelltV1) {
		t.Errorf("expected second event tagesabschluss, got %q", events[1].Type)
	}
}

func TestKasseAbschliessen_CleanerFehlerBleibtBestEffort(t *testing.T) {
	ctx := context.Background()
	journalMock := repotest.NewKassenjournalRepo(nil, nil)
	journalMock.SetKassenbestand(50000) // Soll = Ist
	cleaner := &fakeDruckauftraege{fehlgeschlagen: 2, err: fmt.Errorf("cleanup kaputt")}
	sitzungMock := repotest.NewKassensitzungenRepo(testOpenKS, nil)
	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: sitzungMock,
		TSERepo:             tseGateMock{},
		DruckauftragRepo:    cleaner,
	}

	_, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000)
	if err != nil {
		t.Fatalf("expected Abschluss to stay successful despite cleaner error, got %v", err)
	}
	if cleaner.versuche != 1 {
		t.Errorf("expected one discard attempt, got %d", cleaner.versuche)
	}
	// The cleaner error must not reach the named return err. The mock status proves the defer reset did not run.
	assertSitzungStatus(t, sitzungMock, kasse.KassensitzungWirdAbgeschlossen)

	events, err := journalMock.ReadEventsBySubject(ctx, kasse.KassensitzungSubject(testOpenKS.ZNr))
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected two events (kassensturz + tagesabschluss) despite cleaner error, got %d", len(events))
	}
	if events[1].Type != string(kasse.EventTypeTagesabschlussErstelltV1) {
		t.Errorf("expected second event tagesabschluss, got %q", events[1].Type)
	}
}

// The three Tagesabschluss sums come from the session's journal events. The mock also returns the
// Kassensturz events written in the same run, since they are committed when read.
func TestKasseAbschliessen_TagesabschlussMitEchtenSummen(t *testing.T) {
	ctx := context.Background()
	journalMock := repotest.NewKassenjournalRepo(nil, nil)
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
	transitRaw, _ := json.Marshal(map[string]any{"richtung": "einlage", "betragCents": 400})
	journalMock.AddEvent(e.Event{
		UserID:   1,
		UserName: "Test",
		Type:     string(kasse.EventTypeGeldtransitGebuchtV1),
		Subject:  kasse.KassensitzungSubject(testOpenKS.ZNr),
		Data:     transitRaw,
	})

	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
		TSERepo:             tseGateMock{},
		DruckauftragRepo:    &fakeDruckauftraege{},
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
	journalMock := repotest.NewKassenjournalRepo(nil, nil)
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
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
		TSERepo:             tseGateMock{},
		DruckauftragRepo:    &fakeDruckauftraege{},
	}

	_, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000)
	if !errors.Is(err, ErrTischeSaldoOffen) {
		t.Errorf("expected ErrTischeSaldoOffen, got %v", err)
	}

	events, err := journalMock.ReadEventsBySubject(ctx, kasse.KassensitzungSubject(testOpenKS.ZNr))
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected no events written when saldo blocks, got %d", len(events))
	}
}

func TestKasseAbschliessen_KasseNichtGeoeffnet(t *testing.T) {
	ctx := context.Background()
	cmd := newTestCommand(nil)

	_, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000)
	if !errors.Is(err, ErrKasseNichtGeoeffnet) {
		t.Errorf("expected ErrKasseNichtGeoeffnet, got %v", err)
	}
}

// Der Abschluss setzt die Barriere ('wird_abgeschlossen') als ersten Schritt und
// setzt sie bei Erfolg nicht zurück.
func TestKasseAbschliessen_SetztBarriere(t *testing.T) {
	ctx := context.Background()
	journalMock := repotest.NewKassenjournalRepo(nil, nil)
	journalMock.SetKassenbestand(50000)
	sitzungMock := repotest.NewKassensitzungenRepo(testOpenKS, nil)
	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: sitzungMock,
		TSERepo:             tseGateMock{},
		DruckauftragRepo:    &fakeDruckauftraege{},
	}

	if _, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	assertSitzungStatus(t, sitzungMock, kasse.KassensitzungWirdAbgeschlossen)
}

// Schlägt der Abschluss nach dem Statuswechsel fehl, wird die Sitzung best effort auf 'offen'
// zurückgesetzt, damit sie nicht im Zwischenstatus hängen bleibt.
func TestKasseAbschliessen_FehlerSetztStatusZurueck(t *testing.T) {
	ctx := context.Background()
	journalMock := repotest.NewKassenjournalRepo(nil, nil)
	journalMock.SetKassenbestand(50000)
	// ReadKassensitzungEvents scheitert nach der Barriere und den Kassensturz-Writes.
	journalMock.SetReadKassensitzungEventsErr(db.ErrDatabase)
	sitzungMock := repotest.NewKassensitzungenRepo(testOpenKS, nil)
	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: sitzungMock,
		TSERepo:             tseGateMock{},
		DruckauftragRepo:    &fakeDruckauftraege{},
	}

	if _, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000); err == nil {
		t.Error("expected an error, got nil")
	}
	assertSitzungStatus(t, sitzungMock, kasse.KassensitzungOffen)
	want := []kasse.KassensitzungStatus{kasse.KassensitzungWirdAbgeschlossen, kasse.KassensitzungOffen}
	if !slices.Equal(sitzungMock.Statuswechsel, want) {
		t.Errorf("expected barrier set then reset %v, got %v", want, sitzungMock.Statuswechsel)
	}
}

// Ein Versionskonflikt bedeutet einen konkurrierenden zweiten Abschluss — die unterlegene
// Instanz darf die Barriere nicht unter dem gewinnenden Abschluss wegräumen.
func TestKasseAbschliessen_KonfliktSetztStatusNichtZurueck(t *testing.T) {
	ctx := context.Background()
	journalMock := repotest.NewKassenjournalRepoWithWriteErr(nil, db.ErrAlreadyExists)
	journalMock.SetKassenbestand(50000)
	sitzungMock := repotest.NewKassensitzungenRepo(testOpenKS, nil)
	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: sitzungMock,
		TSERepo:             tseGateMock{},
		DruckauftragRepo:    &fakeDruckauftraege{},
	}

	if _, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000); !errors.Is(err, ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}
	assertSitzungStatus(t, sitzungMock, kasse.KassensitzungWirdAbgeschlossen)
}

// Ein Deadlock (40P01) beim Event-Write wird wie ein Konflikt behandelt (409 statt 500);
// die Barriere bleibt stehen, ein erneuter Aufruf setzt den Abschluss fort.
func TestKasseAbschliessen_DeadlockMapsToKonflikt(t *testing.T) {
	ctx := context.Background()
	journalMock := repotest.NewKassenjournalRepoWithWriteErr(nil, db.ErrConflict)
	journalMock.SetKassenbestand(50000)
	sitzungMock := repotest.NewKassensitzungenRepo(testOpenKS, nil)
	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: sitzungMock,
		TSERepo:             tseGateMock{},
		DruckauftragRepo:    &fakeDruckauftraege{},
	}

	if _, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000); !errors.Is(err, ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}
}

// Wiederanlauf: Steht die Sitzung nach einem Abbruch noch auf 'wird_abgeschlossen', schließt
// ein erneuter Aufruf sie erfolgreich ab.
func TestKasseAbschliessen_WiederanlaufImZwischenstatus(t *testing.T) {
	ctx := context.Background()
	imAbschluss := &kasse.Kassensitzung{ZNr: 1, Status: kasse.KassensitzungWirdAbgeschlossen, CreatedAt: time.Now().UTC()}
	journalMock := repotest.NewKassenjournalRepo(nil, nil)
	journalMock.SetKassenbestand(50000)
	cmd := Command{
		KassenjournalRepo:   journalMock,
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(imAbschluss, nil),
		TSERepo:             tseGateMock{},
		DruckauftragRepo:    &fakeDruckauftraege{},
	}

	if _, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000); err != nil {
		t.Fatalf("expected resume to succeed, got %v", err)
	}

	events, err := journalMock.ReadEventsBySubject(ctx, kasse.KassensitzungSubject(imAbschluss.ZNr))
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected kassensturz + tagesabschluss, got %d", len(events))
	}
}

// Wiederanlauf nach Teilfehler: Steht der Kassensturz eines abgebrochenen vorherigen
// Versuchs bereits im Journal, wird Schritt 1 übersprungen — es entsteht kein zweites
// kassensturz-Event, und die Differenz rechnet gegen den dort dokumentierten Ist-Bestand.
func TestKasseAbschliessen_WiederanlaufSchreibtKeinenZweitenKassensturz(t *testing.T) {
	ctx := context.Background()
	journalMock := repotest.NewKassenjournalRepo(nil, nil)
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
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
		TSERepo:             tseGateMock{},
		DruckauftragRepo:    &fakeDruckauftraege{},
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
		t.Errorf("expected first event kassensturz, got %q", events[0].Type)
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
		t.Errorf("expected third event tagesabschluss, got %q", events[2].Type)
	}
}

// Wiederanlauf mit Zwischenbuchung: Steht nach dem protokollierten Kassensturz eine echte
// Buchung (geldtransit-gebucht:v1) im Stream, bricht der Abschluss mit ErrBuchungenNachKassensturz
// ab, statt den veralteten Ist-Bestand zu übernehmen — es wird kein Abschluss-Event geschrieben.
func TestKasseAbschliessen_WiederanlaufMitZwischenbuchungBrichtAb(t *testing.T) {
	ctx := context.Background()
	journalMock := repotest.NewKassenjournalRepo(nil, nil)
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
	transitRaw, err := json.Marshal(map[string]any{"richtung": "einlage", "betragCents": 400})
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
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
		TSERepo:             tseGateMock{},
		DruckauftragRepo:    &fakeDruckauftraege{},
	}

	if _, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 49500); !errors.Is(err, ErrBuchungenNachKassensturz) {
		t.Errorf("expected ErrBuchungenNachKassensturz, got %v", err)
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

// An unparsable sum-relevant event aborts the Tagesabschluss without writing its event: wrong Z-Bon
// sums are worse than a blocked Abschluss. Reachable only with a corrupt store.
func TestKasseAbschliessen_KorruptesEventBrichtAbschlussAb(t *testing.T) {
	ctx := context.Background()
	journalMock := repotest.NewKassenjournalRepo(nil, nil)
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
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
		TSERepo:             tseGateMock{},
		DruckauftragRepo:    &fakeDruckauftraege{},
	}

	_, err := cmd.KasseAbschliessen(ctx, 1, "Admin", 50000)
	if err == nil {
		t.Error("expected error for corrupt event, got nil")
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
