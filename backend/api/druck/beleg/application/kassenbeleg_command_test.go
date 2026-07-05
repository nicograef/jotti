//go:build unit

package application

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/betreiber"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
)

const testKassensitzungNr = 1

var testOpenKS = &kasse.Kassensitzung{
	ZNr:    testKassensitzungNr,
	Status: kasse.KassensitzungOffen,
}

var testActiveTisch = tisch.Tisch{
	ID:        1,
	Name:      "Tisch 1",
	Status:    tisch.ActiveStatus,
	CreatedAt: time.Now().UTC(),
	UpdatedAt: time.Now().UTC(),
}

type mockDruckstationRepo struct {
	konfig map[string]druckstation.Druckstation
	err    error
}

func (m *mockDruckstationRepo) GetKonfigurierteDruckstationen(_ context.Context) (map[string]druckstation.Druckstation, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.konfig, nil
}

// kassenbelegStationen ist die konfigurierte Kassenbeleg-Druckstation für die
// KassenbelegDrucken-Tests (Ziel-IP des Kassenbeleg-Druckers).
var kassenbelegStationen = map[string]druckstation.Druckstation{
	"kassenbeleg": {DruckerIP: "192.168.1.80"},
}

type mockDruckauftragRepo struct {
	enqueued []druckauftrag_repo.NeuerDruckauftrag
	err      error
}

func (m *mockDruckauftragRepo) EnqueueDruckauftraege(_ context.Context, auftraege []druckauftrag_repo.NeuerDruckauftrag) error {
	if m.err != nil {
		return m.err
	}
	m.enqueued = append(m.enqueued, auftraege...)
	return nil
}

type mockSettingsRepo struct {
	betreiber    betreiber.Betreiber
	betreiberErr error
}

func (m *mockSettingsRepo) GetBetreiber(_ context.Context) (betreiber.Betreiber, error) {
	if m.betreiberErr != nil {
		return betreiber.Betreiber{}, m.betreiberErr
	}
	return m.betreiber, nil
}

// mockTSEAuftragRepo liefert den Signaturauftrags-Stand je Event-ID, den
// aktiven Stoerungszeitraum und die Kassenidentitaet; Events ohne Eintrag gelten
// als nicht signaturpflichtig (db.ErrNotFound).
type mockTSEAuftragRepo struct {
	staende     map[int]tse.SignaturauftragStand
	stoerung    *tse.Stoerung
	kassenident tse.Kassenidentitaet
}

func (m *mockTSEAuftragRepo) GetSignaturauftragZuEvent(_ context.Context, eventID int) (tse.SignaturauftragStand, error) {
	if stand, ok := m.staende[eventID]; ok {
		return stand, nil
	}
	return tse.SignaturauftragStand{}, db.ErrNotFound
}

func (m *mockTSEAuftragRepo) GetAktiveTSEStoerung(_ context.Context) (*tse.Stoerung, error) {
	return m.stoerung, nil
}

func (m *mockTSEAuftragRepo) GetKassenidentitaet(_ context.Context) (tse.Kassenidentitaet, error) {
	return m.kassenident, nil
}

func TestKassenbelegDrucken_SuccessAndReprint(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	zahlungEvent, err := kasse.NewZahlungKassiertEvent(subject, 1, "Test User", []kasse.Position{
		{
			PositionID:   "11111111-1111-1111-1111-111111111111",
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk", Steuersatz: "regel",
			Einzelpreis: 350,
			Menge:       2,
		},
	}, 700, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	var eventData struct {
		ZahlungID string `json:"zahlungId"`
	}
	if err := json.Unmarshal(zahlungEvent.Data, &eventData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(zahlungEvent)

	auftragMock := &mockDruckauftragRepo{}
	settingsMock := &mockSettingsRepo{
		betreiber: betreiber.Betreiber{
			Vereinsname: "SV Musterstadt",
			Strasse:     "Musterstrasse 1",
			Plz:         "12345",
			Ort:         "Musterstadt",
			UpdatedAt:   time.Now(),
		},
	}

	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		BetreiberRepo:       settingsMock,
		DruckauftragRepo:    auftragMock,
		TSERepo:             &mockTSEAuftragRepo{},
	}

	status, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, ZahlungID: eventData.ZahlungID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != BelegStatusEingereiht {
		t.Fatalf("expected status eingereiht, got %q", status)
	}
	status, err = command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, ZahlungID: eventData.ZahlungID})
	if err != nil {
		t.Fatalf("expected no reprint error, got %v", err)
	}
	if status != BelegStatusEingereiht {
		t.Fatalf("expected reprint status eingereiht, got %q", status)
	}

	if len(auftragMock.enqueued) != 2 {
		t.Fatalf("expected 2 enqueued auftraege, got %d", len(auftragMock.enqueued))
	}
	if auftragMock.enqueued[0].BonArt != "kassenbeleg" {
		t.Fatalf("expected bon_art kassenbeleg, got %s", auftragMock.enqueued[0].BonArt)
	}
	if auftragMock.enqueued[0].ZielIP != "192.168.1.80" {
		t.Fatalf("expected ziel_ip 192.168.1.80, got %s", auftragMock.enqueued[0].ZielIP)
	}
	if !strings.HasPrefix(auftragMock.enqueued[0].Referenz, "zahlung-kassiert:") {
		t.Fatalf("expected zahlung-kassiert referenz, got %s", auftragMock.enqueued[0].Referenz)
	}
}

func TestKassenbelegDrucken_ContainsSteuerkennzeichenUndSteuermatrix(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	zahlungEvent, err := kasse.NewZahlungKassiertEvent(subject, 1, "Test User", []kasse.Position{
		{
			PositionID:   "11111111-1111-1111-1111-111111111111",
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk", Steuersatz: "regel",
			Einzelpreis: 350,
			Menge:       2,
		},
		{
			PositionID:   "22222222-2222-2222-2222-222222222222",
			VarianteID:   2,
			ProduktName:  "Brezel",
			VarianteName: "normal",
			Kategorie:    "essen", Steuersatz: "ermaessigt",
			Einzelpreis: 300,
			Menge:       1,
		},
	}, 1000, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	var eventData struct {
		ZahlungID string `json:"zahlungId"`
	}
	if err := json.Unmarshal(zahlungEvent.Data, &eventData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(zahlungEvent)

	auftragMock := &mockDruckauftragRepo{}
	settingsMock := &mockSettingsRepo{
		betreiber: betreiber.Betreiber{
			Vereinsname: "SV Musterstadt",
			Strasse:     "Musterstrasse 1",
			Plz:         "12345",
			Ort:         "Musterstadt",
			UpdatedAt:   time.Now(),
		},
	}

	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		BetreiberRepo:       settingsMock,
		DruckauftragRepo:    auftragMock,
		TSERepo:             &mockTSEAuftragRepo{},
	}

	if _, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, ZahlungID: eventData.ZahlungID}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(auftragMock.enqueued) != 1 {
		t.Fatalf("expected exactly 1 enqueued auftrag, got %d", len(auftragMock.enqueued))
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}

	got := string(payload)
	checks := []string{
		"GESAMT: 10,00 EUR",
		"= 7,00 EUR (A)",
		"= 3,00 EUR (B)",
		"Steueraufteilung:",
		"A: Netto 5,88 EUR, Steuer 1,12 EUR, Brutto 7,00 EUR",
		"B: Netto 2,80 EUR, Steuer 0,20 EUR, Brutto 3,00 EUR",
	}

	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("kassenbeleg payload enthaelt %q nicht; got:\n%q", check, got)
		}
	}
}

func TestKassenbelegDrucken_MitSignaturAmAuftrag_ContainsTSEBlock(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	zahlungEvent, err := kasse.NewZahlungKassiertEvent(subject, 1, "Test User", []kasse.Position{
		{
			PositionID:   "11111111-1111-1111-1111-111111111111",
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk", Steuersatz: "regel",
			Einzelpreis: 350,
			Menge:       1,
		},
	}, 350, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	var eventData struct {
		ZahlungID string `json:"zahlungId"`
	}
	if err := json.Unmarshal(zahlungEvent.Data, &eventData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(zahlungEvent) // Event-ID 1

	// Die Signatur liegt am quittierten Auftrag — die einzige Signaturquelle.
	tseRepo := &mockTSEAuftragRepo{staende: map[int]tse.SignaturauftragStand{
		1: {Status: tse.StatusErledigt, ErstelltAm: time.Date(2026, 6, 10, 18, 0, 0, 0, time.UTC), Signatur: &tse.Signatur{
			TransaktionNummer: 3001,
			SignaturZaehler:   77,
			TSESeriennummer:   "SW-TSE-SN-0042",
			LogTimeStart:      time.Date(2026, 6, 10, 18, 0, 1, 0, time.UTC),
			LogTimeEnd:        time.Date(2026, 6, 10, 18, 0, 3, 0, time.UTC),
			Signatur:          "SIG-XYZ",
			QRCodeData:        "V0;XYZ",
		}},
	}}

	auftragMock := &mockDruckauftragRepo{}
	settingsMock := &mockSettingsRepo{
		betreiber: betreiber.Betreiber{
			Vereinsname: "SV Musterstadt",
			Strasse:     "Musterstrasse 1",
			Plz:         "12345",
			Ort:         "Musterstadt",
			UpdatedAt:   time.Now(),
		},
	}

	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		BetreiberRepo:       settingsMock,
		DruckauftragRepo:    auftragMock,
		TSERepo:             tseRepo,
	}

	status, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, ZahlungID: eventData.ZahlungID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != BelegStatusEingereiht {
		t.Fatalf("expected status eingereiht, got %q", status)
	}

	if len(auftragMock.enqueued) != 1 {
		t.Fatalf("expected exactly 1 enqueued auftrag, got %d", len(auftragMock.enqueued))
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}

	got := string(payload)
	checks := []string{
		"TSE-Daten:",
		"TSE-Transaktion: 3001",
		"Signaturzaehler: 77",
		"TSE-Seriennummer: SW-TSE-SN-0042",
		"TSE-Start: 10.06.2026 18:00:01",
		"TSE-Ende: 10.06.2026 18:00:03",
		"Signatur: SIG-XYZ",
	}

	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("kassenbeleg payload enthaelt %q nicht; got:\n%q", check, got)
		}
	}

	// Prompte Signatur: kein Nachsigniert-Vermerk (Kriterium: verspätet).
	if strings.Contains(got, "Nachsigniert") {
		t.Fatalf("expected no Nachsigniert-Vermerk on promptly signed beleg, got:\n%q", got)
	}
}

func TestKassenbelegDrucken_Tischzahlung_WithErsteBestellungKlartext(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	zahlungEvent, err := kasse.NewZahlungKassiertEvent(subject, 1, "Test User", []kasse.Position{
		{
			PositionID:   "11111111-1111-4111-8111-111111111111",
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk", Steuersatz: "regel",
			Einzelpreis: 350,
			Menge:       1,
		},
	}, 350, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	var eventData struct {
		ZahlungID string `json:"zahlungId"`
	}
	if err := json.Unmarshal(zahlungEvent.Data, &eventData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}

	ersteBestellung := time.Date(2026, 5, 1, 18, 1, 0, 0, time.UTC)

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(zahlungEvent)
	eventMock.SetTischSession(subject, kasse.TischSession{
		Subject:                subject,
		TischID:                testActiveTisch.ID,
		KassensitzungNr:        testKassensitzungNr,
		ErsteBestellungLogTime: &ersteBestellung,
	})

	auftragMock := &mockDruckauftragRepo{}
	settingsMock := &mockSettingsRepo{
		betreiber: betreiber.Betreiber{
			Vereinsname: "SV Musterstadt",
			Strasse:     "Musterstrasse 1",
			Plz:         "12345",
			Ort:         "Musterstadt",
			UpdatedAt:   time.Now(),
		},
	}

	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		BetreiberRepo:       settingsMock,
		DruckauftragRepo:    auftragMock,
		TSERepo:             &mockTSEAuftragRepo{},
	}

	if _, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, ZahlungID: eventData.ZahlungID}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}

	got := string(payload)
	if !strings.Contains(got, "Erste Bestellung: 01.05.2026 18:01:00") {
		t.Fatalf("expected first order klartext in table receipt, got:\n%q", got)
	}
}

// Der Beleg-Abruf antwortet sofort mit dem Signaturstatus: Solange der Auftrag
// nicht quittiert ist, entsteht kein Druckauftrag (ausstehend, die UI fasst
// nach); nach der Quittierung liefert derselbe Aufruf den Beleg mit dem
// TSE-Abschnitt aus den Signaturspalten des Auftrags.
func TestKassenbelegDrucken_AusstehendDannEingereiht(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	zahlungEvent, err := kasse.NewZahlungKassiertEvent(subject, 1, "Test User", []kasse.Position{{
		PositionID:   "11111111-1111-4111-8111-111111111111",
		VarianteID:   1,
		ProduktName:  "Cola",
		VarianteName: "0,5l",
		Kategorie:    "getraenk",
		Steuersatz:   "regel",
		Einzelpreis:  350,
		Menge:        1,
	}}, 350, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	var eventData struct {
		ZahlungID string `json:"zahlungId"`
	}
	if err := json.Unmarshal(zahlungEvent.Data, &eventData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(zahlungEvent) // Event-ID 1

	// Auftrag existiert, aber der Worker hat noch nicht quittiert.
	tseRepo := &mockTSEAuftragRepo{staende: map[int]tse.SignaturauftragStand{
		1: {Status: tse.StatusOffen, ErstelltAm: time.Now().UTC()},
	}}

	auftragMock := &mockDruckauftragRepo{}
	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		BetreiberRepo:       belegTestSettingsMock(),
		DruckauftragRepo:    auftragMock,
		TSERepo:             tseRepo,
	}

	status, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, ZahlungID: eventData.ZahlungID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != BelegStatusAusstehend {
		t.Fatalf("expected status ausstehend, got %q", status)
	}
	if len(auftragMock.enqueued) != 0 {
		t.Fatalf("expected no druckauftrag while signature is pending, got %d", len(auftragMock.enqueued))
	}

	// Der Worker quittiert — der nächste Abruf liefert den Beleg mit Signatur.
	tseRepo.staende[1] = tse.SignaturauftragStand{Status: tse.StatusErledigt, ErstelltAm: time.Date(2026, 6, 10, 18, 10, 0, 0, time.UTC), Signatur: &tse.Signatur{
		TransaktionNummer: 3002,
		SignaturZaehler:   78,
		TSESeriennummer:   "SW-TSE-SN-0043",
		LogTimeStart:      time.Date(2026, 6, 10, 18, 10, 1, 0, time.UTC),
		LogTimeEnd:        time.Date(2026, 6, 10, 18, 10, 3, 0, time.UTC),
		Signatur:          "SIG-NACHGEHOLT",
		QRCodeData:        "V0;NACHGEHOLT",
	}}

	status, err = command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, ZahlungID: eventData.ZahlungID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != BelegStatusEingereiht {
		t.Fatalf("expected status eingereiht, got %q", status)
	}
	if len(auftragMock.enqueued) != 1 {
		t.Fatalf("expected exactly 1 enqueued auftrag, got %d", len(auftragMock.enqueued))
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}

	got := string(payload)
	if !strings.Contains(got, "TSE-Daten:") {
		t.Fatalf("expected TSE block from auftrag signature, got:\n%q", got)
	}
	if !strings.Contains(got, "SIG-NACHGEHOLT") {
		t.Fatalf("expected auftrag signature in payload, got:\n%q", got)
	}
}

// belegZahlungFixture liefert einen Event-Mock mit einer kassierten Zahlung
// (Event-ID 1) und deren zahlungId — Fixture der Signaturstatus-Belegtests.
func belegZahlungFixture(t *testing.T) (*kassenjournal_repo.MockRepo, string) {
	t.Helper()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	zahlungEvent, err := kasse.NewZahlungKassiertEvent(subject, 1, "Test User", []kasse.Position{{
		PositionID:   "11111111-1111-4111-8111-111111111111",
		VarianteID:   1,
		ProduktName:  "Cola",
		VarianteName: "0,5l",
		Kategorie:    "getraenk",
		Steuersatz:   "regel",
		Einzelpreis:  350,
		Menge:        1,
	}}, 350, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	var eventData struct {
		ZahlungID string `json:"zahlungId"`
	}
	if err := json.Unmarshal(zahlungEvent.Data, &eventData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(zahlungEvent) // Event-ID 1
	return eventMock, eventData.ZahlungID
}

func belegTestCommand(eventMock *kassenjournal_repo.MockRepo, tseRepo *mockTSEAuftragRepo, auftragMock *mockDruckauftragRepo) Command {
	return Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		BetreiberRepo:       belegTestSettingsMock(),
		DruckauftragRepo:    auftragMock,
		TSERepo:             tseRepo,
	}
}

// Verspätete Signatur (später als rund eine Minute nach Auftragserstellung):
// Der Beleg trägt das Nachsigniert-Kennzeichen — auch beim Nachdruck, der über
// denselben Weg läuft.
func TestKassenbelegDrucken_VerspaeteteSignatur_TraegtNachsigniertVermerk(t *testing.T) {
	ctx := context.Background()
	eventMock, zahlungID := belegZahlungFixture(t)

	// Auftrag um 18:00:00 erstellt, Signatur erst gegen 18:07 quittiert.
	tseRepo := &mockTSEAuftragRepo{staende: map[int]tse.SignaturauftragStand{
		1: {Status: tse.StatusErledigt, ErstelltAm: time.Date(2026, 6, 10, 18, 0, 0, 0, time.UTC), Signatur: &tse.Signatur{
			TransaktionNummer: 3005,
			SignaturZaehler:   80,
			TSESeriennummer:   "SW-TSE-SN-0043",
			LogTimeStart:      time.Date(2026, 6, 10, 18, 7, 1, 0, time.UTC),
			LogTimeEnd:        time.Date(2026, 6, 10, 18, 7, 3, 0, time.UTC),
			Signatur:          "SIG-VERSPAETET",
			QRCodeData:        "V0;VERSPAETET",
		}},
	}}

	auftragMock := &mockDruckauftragRepo{}
	command := belegTestCommand(eventMock, tseRepo, auftragMock)

	status, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, ZahlungID: zahlungID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != BelegStatusEingereiht {
		t.Fatalf("expected status eingereiht, got %q", status)
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}

	got := string(payload)
	if !strings.Contains(got, "TSE-Daten:") {
		t.Fatalf("expected TSE block on nachsignierter beleg, got:\n%q", got)
	}
	if !strings.Contains(got, "Nachsigniert am 10.06.2026 18:07:03") {
		t.Fatalf("expected Nachsigniert-Vermerk, got:\n%q", got)
	}
}

// Endstatus des Auftrags (hier: fehlgeschlagen) ist dokumentierter Ausfall:
// Der Beleg entsteht ohne TSE-Daten, weist den Ausfall aber aus.
func TestKassenbelegDrucken_AusfallEndstatus_BelegMitAusfallvermerk(t *testing.T) {
	ctx := context.Background()
	eventMock, zahlungID := belegZahlungFixture(t)

	tseRepo := &mockTSEAuftragRepo{staende: map[int]tse.SignaturauftragStand{
		1: {Status: tse.StatusFehlgeschlagen, ErstelltAm: time.Now().UTC().Add(-2 * time.Minute)},
	}}

	auftragMock := &mockDruckauftragRepo{}
	command := belegTestCommand(eventMock, tseRepo, auftragMock)

	status, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, ZahlungID: zahlungID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != BelegStatusEingereiht {
		t.Fatalf("expected status eingereiht, got %q", status)
	}
	if len(auftragMock.enqueued) != 1 {
		t.Fatalf("expected exactly 1 enqueued auftrag, got %d", len(auftragMock.enqueued))
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}

	got := string(payload)
	if strings.Contains(got, "TSE-Daten:") {
		t.Fatalf("expected no TSE block on ausfall beleg, got:\n%q", got)
	}
	if !strings.Contains(got, "TSE-Hinweis:") {
		t.Fatalf("expected Ausfallvermerk on beleg, got:\n%q", got)
	}
}

// Ein offener Auftrag während eines aktiven Störungszeitraums ist Ausfall
// (nicht ausstehend): Der Beleg entsteht sofort und weist den Ausfall aus —
// auch in der Aufholphase, solange der Rückstands-Zeitraum aktiv ist.
func TestKassenbelegDrucken_OffenBeiAktiverStoerung_BelegMitAusfallvermerk(t *testing.T) {
	ctx := context.Background()
	eventMock, zahlungID := belegZahlungFixture(t)

	tseRepo := &mockTSEAuftragRepo{
		staende: map[int]tse.SignaturauftragStand{
			1: {Status: tse.StatusOffen, ErstelltAm: time.Now().UTC().Add(-3 * time.Minute)},
		},
		stoerung: &tse.Stoerung{
			Beginn:     time.Now().UTC().Add(-time.Minute),
			GrundArt:   tse.StoerungGrundRueckstand,
			Fehlertext: "Signaturaufträge im Rückstand",
		},
	}

	auftragMock := &mockDruckauftragRepo{}
	command := belegTestCommand(eventMock, tseRepo, auftragMock)

	status, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, ZahlungID: zahlungID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != BelegStatusEingereiht {
		t.Fatalf("expected status eingereiht, got %q", status)
	}
	if len(auftragMock.enqueued) != 1 {
		t.Fatalf("expected exactly 1 enqueued auftrag, got %d", len(auftragMock.enqueued))
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}
	if !strings.Contains(string(payload), "TSE-Hinweis:") {
		t.Fatalf("expected Ausfallvermerk on beleg, got:\n%q", string(payload))
	}
}

func TestKassenbelegDrucken_ZahlungNichtGefunden(t *testing.T) {
	ctx := context.Background()
	command := Command{
		EventRepo:           kassenjournal_repo.NewMock(nil, nil),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		BetreiberRepo:       &mockSettingsRepo{},
		DruckstationRepo:    &mockDruckstationRepo{},
		DruckauftragRepo:    &mockDruckauftragRepo{},
		TSERepo:             &mockTSEAuftragRepo{},
	}

	_, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, ZahlungID: "11111111-1111-1111-1111-111111111111"})
	if err != ErrZahlungNichtGefunden {
		t.Fatalf("expected ErrZahlungNichtGefunden, got %v", err)
	}
}

func TestKassenbelegDrucken_KassenbelegDruckerNichtKonfiguriert(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	zahlungEvent, err := kasse.NewZahlungKassiertEvent(subject, 1, "Test User", []kasse.Position{
		{
			PositionID:   "11111111-1111-1111-1111-111111111111",
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk", Steuersatz: "regel",
			Einzelpreis: 350,
			Menge:       1,
		},
	}, 350, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	var eventData struct {
		ZahlungID string `json:"zahlungId"`
	}
	if err := json.Unmarshal(zahlungEvent.Data, &eventData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(zahlungEvent)

	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		BetreiberRepo:       &mockSettingsRepo{},
		DruckstationRepo:    &mockDruckstationRepo{},
		DruckauftragRepo:    &mockDruckauftragRepo{},
		TSERepo:             &mockTSEAuftragRepo{},
	}

	if _, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, ZahlungID: eventData.ZahlungID}); err != ErrKassenbelegDruckerNichtKonfiguriert {
		t.Fatalf("expected ErrKassenbelegDruckerNichtKonfiguriert, got %v", err)
	}
}

func TestKassenbelegDrucken_Direktverkauf_ExactlyOneAuftrag(t *testing.T) {
	ctx := context.Background()
	verkaufID := uuid.New().String()
	subject := kasse.DirektverkaufSubject(testKassensitzungNr, verkaufID)

	verkaufEvent, err := kasse.NewDirektverkaufGetaetigtEvent(subject, verkaufID, 1, "Test User", []kasse.Position{
		{
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk", Steuersatz: "regel",
			Einzelpreis: 350,
			Menge:       2,
		},
	}, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(verkaufEvent)

	auftragMock := &mockDruckauftragRepo{}
	settingsMock := &mockSettingsRepo{
		betreiber: betreiber.Betreiber{
			Vereinsname: "SV Musterstadt",
			Strasse:     "Musterstrasse 1",
			Plz:         "12345",
			Ort:         "Musterstadt",
			UpdatedAt:   time.Now(),
		},
	}

	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		BetreiberRepo:       settingsMock,
		DruckauftragRepo:    auftragMock,
		TSERepo:             &mockTSEAuftragRepo{},
	}

	if _, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{VerkaufID: verkaufID}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(auftragMock.enqueued) != 1 {
		t.Fatalf("expected exactly 1 enqueued auftrag, got %d", len(auftragMock.enqueued))
	}
	if auftragMock.enqueued[0].BonArt != "kassenbeleg" {
		t.Fatalf("expected bon_art kassenbeleg, got %s", auftragMock.enqueued[0].BonArt)
	}
	if !strings.HasPrefix(auftragMock.enqueued[0].Referenz, "direktverkauf-getaetigt:") {
		t.Fatalf("expected direktverkauf-getaetigt referenz, got %s", auftragMock.enqueued[0].Referenz)
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}
	if strings.Contains(string(payload), "Erste Bestellung:") {
		t.Fatalf("Direktverkauf-Beleg darf keinen Durchbedienen-Klarschriftzeitpunkt enthalten, got:\n%q", string(payload))
	}
}

func TestKassenbelegDrucken_Direktverkauf_NichtGefunden(t *testing.T) {
	ctx := context.Background()
	command := Command{
		EventRepo:           kassenjournal_repo.NewMock(nil, nil),
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		BetreiberRepo:       &mockSettingsRepo{},
		DruckstationRepo:    &mockDruckstationRepo{},
		DruckauftragRepo:    &mockDruckauftragRepo{},
		TSERepo:             &mockTSEAuftragRepo{},
	}

	_, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{VerkaufID: uuid.New().String()})
	if err != ErrVerkaufNichtGefunden {
		t.Fatalf("expected ErrVerkaufNichtGefunden, got %v", err)
	}
}

func TestKassenbelegDrucken_Direktverkauf_KassenbelegDruckerNichtKonfiguriert(t *testing.T) {
	ctx := context.Background()
	verkaufID := uuid.New().String()
	subject := kasse.DirektverkaufSubject(testKassensitzungNr, verkaufID)

	verkaufEvent, err := kasse.NewDirektverkaufGetaetigtEvent(subject, verkaufID, 1, "Test User", []kasse.Position{
		{
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk", Steuersatz: "regel",
			Einzelpreis: 350,
			Menge:       1,
		},
	}, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(verkaufEvent)

	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		BetreiberRepo:       &mockSettingsRepo{},
		DruckstationRepo:    &mockDruckstationRepo{},
		DruckauftragRepo:    &mockDruckauftragRepo{},
		TSERepo:             &mockTSEAuftragRepo{},
	}

	if _, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{VerkaufID: verkaufID}); err != ErrKassenbelegDruckerNichtKonfiguriert {
		t.Fatalf("expected ErrKassenbelegDruckerNichtKonfiguriert, got %v", err)
	}
}

func belegTestSettingsMock() *mockSettingsRepo {
	return &mockSettingsRepo{
		betreiber: betreiber.Betreiber{
			Vereinsname: "SV Musterstadt",
			Strasse:     "Musterstrasse 1",
			Plz:         "12345",
			Ort:         "Musterstadt",
			UpdatedAt:   time.Now(),
		},
	}
}

func TestKassenbelegDrucken_Direktverkauf_MitSignaturAmAuftrag(t *testing.T) {
	ctx := context.Background()
	verkaufID := uuid.New().String()
	subject := kasse.DirektverkaufSubject(testKassensitzungNr, verkaufID)

	verkaufEvent, err := kasse.NewDirektverkaufGetaetigtEvent(subject, verkaufID, 1, "Test User", []kasse.Position{{
		VarianteID:   1,
		ProduktName:  "Cola",
		VarianteName: "0,5l",
		Kategorie:    "getraenk", Steuersatz: "regel",
		Einzelpreis: 350,
		Menge:       2,
	}}, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(verkaufEvent) // Event-ID 1

	tseRepo := &mockTSEAuftragRepo{staende: map[int]tse.SignaturauftragStand{
		1: {Status: tse.StatusErledigt, ErstelltAm: time.Date(2026, 6, 10, 19, 0, 0, 0, time.UTC), Signatur: &tse.Signatur{
			TransaktionNummer: 4001,
			SignaturZaehler:   99,
			TSESeriennummer:   "SW-TSE-SN-0044",
			LogTimeStart:      time.Date(2026, 6, 10, 19, 0, 1, 0, time.UTC),
			LogTimeEnd:        time.Date(2026, 6, 10, 19, 0, 3, 0, time.UTC),
			Signatur:          "SIG-DIREKTVERKAUF",
			QRCodeData:        "V0;DIREKTVERKAUF",
		}},
	}}

	auftragMock := &mockDruckauftragRepo{}
	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		BetreiberRepo:       belegTestSettingsMock(),
		DruckauftragRepo:    auftragMock,
		TSERepo:             tseRepo,
	}

	if _, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{VerkaufID: verkaufID}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}

	got := string(payload)
	for _, check := range []string{"TSE-Daten:", "TSE-Transaktion: 4001", "Signaturzaehler: 99", "TSE-Seriennummer: SW-TSE-SN-0044", "SIG-DIREKTVERKAUF", "V0;DIREKTVERKAUF"} {
		if !strings.Contains(got, check) {
			t.Fatalf("expected %q in direktverkauf receipt, got:\n%q", check, got)
		}
	}
}

// Solange der Signaturauftrag des Direktverkaufs nicht quittiert ist, antwortet
// der Beleg-Abruf mit ausstehend und legt keinen Druckauftrag an.
func TestKassenbelegDrucken_Direktverkauf_SignaturAusstehend_KeinDruckauftrag(t *testing.T) {
	ctx := context.Background()
	verkaufID := uuid.New().String()
	subject := kasse.DirektverkaufSubject(testKassensitzungNr, verkaufID)

	verkaufEvent, err := kasse.NewDirektverkaufGetaetigtEvent(subject, verkaufID, 1, "Test User", []kasse.Position{{
		VarianteID:   1,
		ProduktName:  "Cola",
		VarianteName: "0,5l",
		Kategorie:    "getraenk", Steuersatz: "regel",
		Einzelpreis: 350,
		Menge:       1,
	}}, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(verkaufEvent) // Event-ID 1

	tseRepo := &mockTSEAuftragRepo{staende: map[int]tse.SignaturauftragStand{
		1: {Status: tse.StatusOffen, ErstelltAm: time.Now().UTC()},
	}}

	auftragMock := &mockDruckauftragRepo{}
	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		BetreiberRepo:       belegTestSettingsMock(),
		DruckauftragRepo:    auftragMock,
		TSERepo:             tseRepo,
	}

	status, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{VerkaufID: verkaufID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != BelegStatusAusstehend {
		t.Fatalf("expected status ausstehend, got %q", status)
	}
	if len(auftragMock.enqueued) != 0 {
		t.Fatalf("expected no druckauftrag while signature is pending, got %d", len(auftragMock.enqueued))
	}
}

func TestKassenbelegDrucken_DirektverkaufStorno_DruckbarAlsStornobeleg(t *testing.T) {
	ctx := context.Background()
	verkaufID := uuid.New().String()
	subject := kasse.DirektverkaufSubject(testKassensitzungNr, verkaufID)

	verkaufEvent, err := kasse.NewDirektverkaufGetaetigtEvent(subject, verkaufID, 1, "Test User", []kasse.Position{{
		VarianteID:   1,
		ProduktName:  "Cola",
		VarianteName: "0,5l",
		Kategorie:    "getraenk", Steuersatz: "regel",
		Einzelpreis: 350,
		Menge:       2,
	}}, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	stornoEvent, err := kasse.NewDirektverkaufStorniertEvent(subject, verkaufID, 2, "Leitung", []kasse.Position{{
		PositionID:   "11111111-1111-4111-8111-111111111111",
		VarianteID:   1,
		ProduktName:  "Cola",
		VarianteName: "0,5l",
		Kategorie:    "getraenk", Steuersatz: "regel",
		Einzelpreis: 350,
		Menge:       2,
	}}, 700, "Rückgabe")
	if err != nil {
		t.Fatalf("expected no storno event error, got %v", err)
	}

	var stornoData kasse.DirektverkaufStorniertV1Data
	if err := json.Unmarshal(stornoEvent.Data, &stornoData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(verkaufEvent) // Event-ID 1
	eventMock.AddEvent(stornoEvent)  // Event-ID 2

	// Die Signatur des Storno-Vorgangs liegt am quittierten Auftrag.
	tseRepo := &mockTSEAuftragRepo{staende: map[int]tse.SignaturauftragStand{
		2: {Status: tse.StatusErledigt, ErstelltAm: time.Date(2026, 6, 10, 19, 20, 0, 0, time.UTC), Signatur: &tse.Signatur{
			TransaktionNummer: 4003,
			SignaturZaehler:   101,
			TSESeriennummer:   "SW-TSE-SN-0044",
			LogTimeStart:      time.Date(2026, 6, 10, 19, 20, 1, 0, time.UTC),
			LogTimeEnd:        time.Date(2026, 6, 10, 19, 20, 3, 0, time.UTC),
			Signatur:          "SIG-STORNO",
			QRCodeData:        "V0;STORNO",
		}},
	}}

	auftragMock := &mockDruckauftragRepo{}
	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		BetreiberRepo:       belegTestSettingsMock(),
		DruckauftragRepo:    auftragMock,
		TSERepo:             tseRepo,
	}

	if _, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{VerkaufID: verkaufID, StornierungID: stornoData.StornierungID}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(auftragMock.enqueued) != 1 {
		t.Fatalf("expected exactly 1 enqueued auftrag, got %d", len(auftragMock.enqueued))
	}
	if !strings.HasPrefix(auftragMock.enqueued[0].Referenz, "direktverkauf-storniert:") {
		t.Fatalf("expected direktverkauf-storniert referenz, got %s", auftragMock.enqueued[0].Referenz)
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}

	got := string(payload)
	for _, check := range []string{"STORNOBELEG", "Storno zu Bon-Nr: 1", "GESAMT: -7,00 EUR", "-3,50 x 2 = -7,00 EUR", "SIG-STORNO"} {
		if !strings.Contains(got, check) {
			t.Fatalf("expected %q in stornobeleg, got:\n%q", check, got)
		}
	}
}

func TestKassenbelegDrucken_TischStorno_DruckbarAlsStornobeleg(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	orderEvent, _ := kasse.NewBestellungAufgenommenEvent(subject, 1, "Test User", []kasse.Position{{
		VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 350, Menge: 2,
	}}, "")
	var orderData kasse.BestellungAufgenommenV1Data
	if err := json.Unmarshal(orderEvent.Data, &orderData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}
	posID := orderData.Positionen[0].PositionID

	paymentEvent, _ := kasse.NewZahlungKassiertEvent(subject, 1, "Test User", []kasse.Position{{
		PositionID: posID, VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 350, Menge: 2,
	}}, 700, "")
	var paymentData kasse.ZahlungKassiertV1Data
	if err := json.Unmarshal(paymentEvent.Data, &paymentData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}

	stornoEvent, err := kasse.NewStornierungErteiltEvent(subject, 2, "Leitung", paymentData.ZahlungID, []kasse.Position{{
		PositionID: posID, VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 350, Menge: 2,
	}}, 700, "Rückgabe")
	if err != nil {
		t.Fatalf("expected no storno event error, got %v", err)
	}
	var stornoData kasse.StornierungErteiltV1Data
	if err := json.Unmarshal(stornoEvent.Data, &stornoData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(orderEvent)   // ID 1
	eventMock.AddEvent(paymentEvent) // ID 2
	eventMock.AddEvent(stornoEvent)  // ID 3

	// Die Signatur des Storno-Vorgangs liegt am quittierten Auftrag.
	tseRepo := &mockTSEAuftragRepo{staende: map[int]tse.SignaturauftragStand{
		3: {Status: tse.StatusErledigt, ErstelltAm: time.Date(2026, 6, 10, 19, 20, 0, 0, time.UTC), Signatur: &tse.Signatur{
			TransaktionNummer: 4003,
			SignaturZaehler:   101,
			TSESeriennummer:   "SW-TSE-SN-0044",
			LogTimeStart:      time.Date(2026, 6, 10, 19, 20, 1, 0, time.UTC),
			LogTimeEnd:        time.Date(2026, 6, 10, 19, 20, 3, 0, time.UTC),
			Signatur:          "SIG-STORNO",
			QRCodeData:        "V0;STORNO",
		}},
	}}

	auftragMock := &mockDruckauftragRepo{}
	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		BetreiberRepo:       belegTestSettingsMock(),
		DruckauftragRepo:    auftragMock,
		TSERepo:             tseRepo,
	}

	if _, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, StornierungID: stornoData.StornierungID}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(auftragMock.enqueued) != 1 {
		t.Fatalf("expected exactly 1 enqueued auftrag, got %d", len(auftragMock.enqueued))
	}
	if !strings.HasPrefix(auftragMock.enqueued[0].Referenz, "stornierung-erteilt:") {
		t.Fatalf("expected stornierung-erteilt referenz, got %s", auftragMock.enqueued[0].Referenz)
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}

	got := string(payload)
	// Referenz auf den Ursprungs-Zahlungsbeleg (Event-ID 2) und negativer Betrag.
	for _, check := range []string{"STORNOBELEG", "Storno zu Bon-Nr: 2", "GESAMT: -7,00 EUR", "SIG-STORNO"} {
		if !strings.Contains(got, check) {
			t.Fatalf("expected %q in stornobeleg, got:\n%q", check, got)
		}
	}
}

func TestKassenbelegDrucken_DirektverkaufStorno_NichtGefunden(t *testing.T) {
	ctx := context.Background()
	verkaufID := uuid.New().String()
	subject := kasse.DirektverkaufSubject(testKassensitzungNr, verkaufID)

	verkaufEvent, err := kasse.NewDirektverkaufGetaetigtEvent(subject, verkaufID, 1, "Test User", []kasse.Position{{
		VarianteID:   1,
		ProduktName:  "Cola",
		VarianteName: "0,5l",
		Kategorie:    "getraenk", Steuersatz: "regel",
		Einzelpreis: 350,
		Menge:       1,
	}}, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	eventMock := kassenjournal_repo.NewMock(nil, nil)
	eventMock.AddEvent(verkaufEvent)

	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: kassensitzungen_repo.NewMock(testOpenKS, nil),
		BetreiberRepo:       belegTestSettingsMock(),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		DruckauftragRepo:    &mockDruckauftragRepo{},
		TSERepo:             &mockTSEAuftragRepo{},
	}

	_, err = command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{VerkaufID: verkaufID, StornierungID: uuid.New().String()})
	if err != ErrStornierungNichtGefunden {
		t.Fatalf("expected ErrStornierungNichtGefunden, got %v", err)
	}
}
