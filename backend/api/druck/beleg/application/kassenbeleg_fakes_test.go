package application

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/betreiber"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/repotest"
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

var kassenbelegStationen = map[string]druckstation.Druckstation{
	"kassenbeleg": {DruckerIP: "192.168.1.80"},
}

type mockDruckauftragRepo struct {
	enqueued []druckstation.NeuerDruckauftrag
	err      error
}

func (m *mockDruckauftragRepo) EnqueueDruckauftraege(_ context.Context, auftraege []druckstation.NeuerDruckauftrag) error {
	if m.err != nil {
		return m.err
	}
	m.enqueued = append(m.enqueued, auftraege...)
	return nil
}

type mockBetreiberRepo struct {
	betreiber    betreiber.Betreiber
	betreiberErr error
}

func (m *mockBetreiberRepo) GetBetreiber(_ context.Context) (betreiber.Betreiber, error) {
	if m.betreiberErr != nil {
		return betreiber.Betreiber{}, m.betreiberErr
	}
	return m.betreiber, nil
}

// mockTSEAuftragRepo liefert den Signaturauftrags-Stand je Event-ID, den
// aktiven Störungszeitraum und die Kassenidentitaet; Events ohne Eintrag gelten
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

// belegZahlungFixture liefert einen Event-Mock mit einer kassierten Zahlung
// (Event-ID 1) und deren zahlungId — Fixture der Signaturstatus-Belegtests.
func belegZahlungFixture(t *testing.T) (*repotest.KassenjournalRepo, string) {
	t.Helper()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	zahlungEvent, err := kasse.NewZahlungKassiertEvent(subject, 1, "Test User", []kasse.Position{{
		PositionID:       "11111111-1111-4111-8111-111111111111",
		VarianteID:       1,
		ProduktName:      "Cola",
		VarianteName:     "0,5l",
		Kategorie:        "getraenk",
		Steuersatz:       "regel",
		EinzelpreisCents: 350,
		Menge:            1,
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

	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	eventMock.AddEvent(zahlungEvent) // Event-ID 1
	return eventMock, eventData.ZahlungID
}

func belegTestCommand(eventMock *repotest.KassenjournalRepo, tseRepo *mockTSEAuftragRepo, auftragMock *mockDruckauftragRepo) Command {
	return Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		BetreiberRepo:       belegTestBetreiberMock(),
		DruckauftragRepo:    auftragMock,
		TSERepo:             tseRepo,
	}
}

func belegTestBetreiberMock() *mockBetreiberRepo {
	return &mockBetreiberRepo{
		betreiber: betreiber.Betreiber{
			Vereinsname: "SV Musterstadt",
			Strasse:     "Musterstrasse 1",
			Plz:         "12345",
			Ort:         "Musterstadt",
			UpdatedAt:   time.Now(),
		},
	}
}

// pfandBelegFixture stores a Rückgabe of two glasses (event ID 1) and the cancellation of one (event ID 2).
func pfandBelegFixture(t *testing.T) (cmd Command, auftragMock *mockDruckauftragRepo, rueckgabeID, stornierungID string) {
	t.Helper()
	rueckgabeID = uuid.New().String()
	stornierungID = uuid.New().String()
	subject := kasse.PfandrueckgabeSubject(testKassensitzungNr, rueckgabeID)

	rueckgabe, err := kasse.NewPfandZurueckgegebenEvent(subject, rueckgabeID, 0, 1, "Test User", []kasse.Position{{
		VarianteID: 2, ProduktName: "Pfand", VarianteName: "Glas", Kategorie: "sonstiges", Steuersatz: "regel", EinzelpreisCents: 200, Menge: 2,
	}}, "")
	if err != nil {
		t.Fatalf("rueckgabe event: %v", err)
	}
	var rueckgabeData kasse.PfandZurueckgegebenV1Data
	if err := json.Unmarshal(rueckgabe.Data, &rueckgabeData); err != nil {
		t.Fatalf("rueckgabe data: %v", err)
	}
	glas := kasse.PositionFromEventData(rueckgabeData.Positionen[0])
	glas.Menge = 1
	storno, err := kasse.NewPfandrueckgabeStorniertEvent(subject, stornierungID, rueckgabeID, 0, 2, "Leitung", []kasse.Position{glas}, 200, "Falsch gezählt")
	if err != nil {
		t.Fatalf("storno event: %v", err)
	}

	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	eventMock.AddEvent(rueckgabe)
	eventMock.AddEvent(storno)

	auftragMock = &mockDruckauftragRepo{}
	cmd = Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		BetreiberRepo:       belegTestBetreiberMock(),
		DruckauftragRepo:    auftragMock,
		TSERepo:             &mockTSEAuftragRepo{},
	}
	return cmd, auftragMock, rueckgabeID, stornierungID
}

func pfandBelegPayload(t *testing.T, auftragMock *mockDruckauftragRepo) string {
	t.Helper()
	if len(auftragMock.enqueued) != 1 {
		t.Fatalf("erwartet 1 Druckauftrag, eingereiht: %d", len(auftragMock.enqueued))
	}
	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("payload: %v", err)
	}
	return string(payload)
}
