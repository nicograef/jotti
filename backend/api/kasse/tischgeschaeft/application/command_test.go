package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/repository/repotest"
)

const testKassensitzungNr = 1

var testOpenKS = &kasse.Kassensitzung{
	ZNr:    testKassensitzungNr,
	Status: kasse.KassensitzungOffen,
}

func newTestCommand(tables []tisch.Tisch) Command {
	return newTestCommandWithEventMock(tables, repotest.NewKassenjournalRepo(nil, nil))
}

func newTestCommandWithEventMock(tables []tisch.Tisch, eventMock *repotest.KassenjournalRepo) Command {
	return Command{
		TischRepo:           repotest.NewTischRepo(tables, nil),
		EventRepo:           eventMock,
		ProduktRepo:         repotest.NewProduktRepo(nil, db.ErrNotFound),
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{},
	}
}

var testProduct = produkt.Produkt{
	ID:         1,
	Name:       "Cola",
	Kategorie:  produkt.GetraenkKategorie,
	Steuersatz: steuer.RegelSteuersatz,
	Status:     produkt.ActiveStatus,
	Varianten:  []produkt.Variante{},
	CreatedAt:  time.Now().UTC(),
	UpdatedAt:  time.Now().UTC(),
}

var testVariant = produkt.Variante{
	ID:         1,
	Name:       "Cola 0,5l",
	PreisCents: 350,
	Status:     produkt.ActiveStatus,
	CreatedAt:  time.Now().UTC(),
	UpdatedAt:  time.Now().UTC(),
}

var testActiveTisch = tisch.Tisch{
	ID:        1,
	Name:      "Tisch 1",
	Status:    tisch.ActiveStatus,
	CreatedAt: time.Now().UTC(),
	UpdatedAt: time.Now().UTC(),
}

var testInactiveTisch = tisch.Tisch{
	ID:        2,
	Name:      "Tisch 2",
	Status:    tisch.InactiveStatus,
	CreatedAt: time.Now().UTC(),
	UpdatedAt: time.Now().UTC(),
}

type mockDruckstationRepo struct {
	konfig        map[string]druckstation.Druckstation
	zusatzdrucker map[int]druckstation.Zusatzdrucker // keyed by VarianteID
	err           error
}

func (m *mockDruckstationRepo) GetKonfigurierteDruckstationen(_ context.Context) (map[string]druckstation.Druckstation, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.konfig, nil
}

// GetZusatzdruckerJeVariante answers only the requested Varianten, like the repository.
func (m *mockDruckstationRepo) GetZusatzdruckerJeVariante(_ context.Context, varianteIDs []int) (map[int]druckstation.Zusatzdrucker, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := map[int]druckstation.Zusatzdrucker{}
	for _, id := range varianteIDs {
		if zd, ok := m.zusatzdrucker[id]; ok {
			result[id] = zd
		}
	}
	return result, nil
}

func TestBestellungAufnehmen_KasseNichtGeoeffnet(t *testing.T) {
	ctx := context.Background()
	productMock := repotest.NewProduktRepo([]produkt.Produkt{testProduct}, nil)
	productMock.AddVariante(testProduct.ID, testVariant)
	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	command := Command{
		TischRepo:           repotest.NewTischRepo([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		ProduktRepo:         productMock,
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(nil, nil), // no open KS
	}

	inputs := []enrichment.PositionInput{
		{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 1},
	}

	err := command.BestellungAufnehmen(ctx, 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", 1, inputs, "")
	if !errors.Is(err, ErrKasseNichtGeoeffnet) {
		t.Errorf("expected ErrKasseNichtGeoeffnet, got %v", err)
	}
}

func TestBestellungAufnehmen_WithOCC(t *testing.T) {
	ctx := context.Background()
	productMock := repotest.NewProduktRepo([]produkt.Produkt{testProduct}, nil)
	productMock.AddVariante(testProduct.ID, testVariant)
	command := newTestCommand([]tisch.Tisch{testActiveTisch})
	command.ProduktRepo = productMock

	inputs := []enrichment.PositionInput{
		{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 2},
	}

	err := command.BestellungAufnehmen(ctx, 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", 1, inputs, "Testkommentar")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestBestellungAufnehmen_EnqueueArbeitsbonDruckauftraege(t *testing.T) {
	ctx := context.Background()
	productMock := repotest.NewProduktRepo([]produkt.Produkt{testProduct}, nil)
	productMock.AddVariante(testProduct.ID, testVariant)

	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	stationMock := &mockDruckstationRepo{konfig: map[string]druckstation.Druckstation{
		"getraenk": {DruckerIP: "192.168.1.50", Bonmodus: "pro_position"},
	}}

	command := newTestCommandWithEventMock([]tisch.Tisch{testActiveTisch}, eventMock)
	command.ProduktRepo = productMock
	command.DruckstationRepo = stationMock

	inputs := []enrichment.PositionInput{{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 2}}

	err := command.BestellungAufnehmen(ctx, 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", 1, inputs, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	enqueued := eventMock.CapturedDruckauftraege()
	if len(enqueued) != 1 {
		t.Fatalf("expected 1 enqueued druckauftrag, got %d", len(enqueued))
	}
	if enqueued[0].BonArt != "arbeitsbon" {
		t.Errorf("expected BonArt arbeitsbon, got %s", enqueued[0].BonArt)
	}
	if enqueued[0].ZielIP != "192.168.1.50" {
		t.Errorf("expected ZielIP 192.168.1.50, got %s", enqueued[0].ZielIP)
	}
	if enqueued[0].Payload == "" {
		t.Error("expected non-empty payload")
	}
}

func TestBestellungAufnehmen_DrucktAmZusatzdrucker(t *testing.T) {
	ctx := context.Background()
	productMock := repotest.NewProduktRepo([]produkt.Produkt{testProduct}, nil)
	productMock.AddVariante(testProduct.ID, testVariant)

	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	command := newTestCommandWithEventMock([]tisch.Tisch{testActiveTisch}, eventMock)
	command.ProduktRepo = productMock
	command.DruckstationRepo = &mockDruckstationRepo{
		konfig: map[string]druckstation.Druckstation{
			"getraenk": {DruckerIP: "192.168.1.50", Bonmodus: "pro_position"},
		},
		zusatzdrucker: map[int]druckstation.Zusatzdrucker{
			testVariant.ID: {ID: 1, Name: "Bar", DruckerIP: "192.168.1.61", Bonmodus: "pro_position"},
		},
	}

	inputs := []enrichment.PositionInput{{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 2}}
	if err := command.BestellungAufnehmen(ctx, 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaab", 1, inputs, ""); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	enqueued := eventMock.CapturedDruckauftraege()
	if len(enqueued) != 1 {
		t.Fatalf("expected 1 enqueued druckauftrag, got %d", len(enqueued))
	}
	if enqueued[0].ZielIP != "192.168.1.61" {
		t.Errorf("expected ZielIP of the Zusatzdrucker 192.168.1.61, got %s", enqueued[0].ZielIP)
	}
}

// bestellung-aufgenommen:v1 trägt kein Pfand-Kennzeichen: Der Command reicht es aus den Stammdaten an die
// Arbeitsbon-Policy, damit das Pfand-Glas keinen eigenen Bon an der Getränke-Station druckt.
func TestBestellungAufnehmen_PfandOhneArbeitsbon(t *testing.T) {
	ctx := context.Background()
	glas := produkt.Produkt{ID: 2, Name: "Pfand Glas", Kategorie: produkt.GetraenkKategorie, Steuersatz: steuer.RegelSteuersatz, Pfand: true, Status: produkt.ActiveStatus}
	glasVariante := produkt.Variante{ID: 9, Name: "Glas", PreisCents: 200, Status: produkt.ActiveStatus}
	productMock := repotest.NewProduktRepo([]produkt.Produkt{testProduct, glas}, nil)
	productMock.AddVariante(testProduct.ID, testVariant)
	productMock.AddVariante(glas.ID, glasVariante)

	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	command := newTestCommandWithEventMock([]tisch.Tisch{testActiveTisch}, eventMock)
	command.ProduktRepo = productMock
	command.DruckstationRepo = &mockDruckstationRepo{konfig: map[string]druckstation.Druckstation{
		"getraenk": {DruckerIP: "192.168.1.50", Bonmodus: "pro_position"},
	}}

	inputs := []enrichment.PositionInput{
		{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 2},
		{ProduktID: glas.ID, VarianteID: glasVariante.ID, Menge: 2},
	}
	if err := command.BestellungAufnehmen(ctx, 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaac", 1, inputs, ""); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	enqueued := eventMock.CapturedDruckauftraege()
	if len(enqueued) != 1 {
		t.Fatalf("expected 1 enqueued druckauftrag (Cola only), got %d", len(enqueued))
	}
	if enqueued[0].ZielIP != "192.168.1.50" {
		t.Errorf("expected ZielIP 192.168.1.50, got %s", enqueued[0].ZielIP)
	}
}

func TestBestellungAufnehmen_SchreibtPositionskommentar(t *testing.T) {
	ctx := context.Background()
	productMock := repotest.NewProduktRepo([]produkt.Produkt{testProduct}, nil)
	productMock.AddVariante(testProduct.ID, testVariant)
	zweiteVariante := testVariant
	zweiteVariante.ID = 2
	zweiteVariante.Name = "Cola 0,3l"
	productMock.AddVariante(testProduct.ID, zweiteVariante)
	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	command := newTestCommandWithEventMock([]tisch.Tisch{testActiveTisch}, eventMock)
	command.ProduktRepo = productMock

	inputs := []enrichment.PositionInput{
		{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 1, Kommentar: "ohne Eis"},
		{ProduktID: testProduct.ID, VarianteID: zweiteVariante.ID, Menge: 1},
	}
	if err := command.BestellungAufnehmen(ctx, 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", 1, inputs, ""); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	events, err := eventMock.ReadKassensitzungEvents(ctx, testKassensitzungNr)
	if err != nil || len(events) != 1 {
		t.Fatalf("expected one written event, got %d (err %v)", len(events), err)
	}
	var data struct {
		Positionen []map[string]json.RawMessage `json:"positionen"`
	}
	if err := json.Unmarshal(events[0].Data, &data); err != nil {
		t.Fatalf("unmarshal event data: %v", err)
	}
	if got := string(data.Positionen[0]["kommentar"]); got != `"ohne Eis"` {
		t.Errorf("positionen[0].kommentar = %s, want \"ohne Eis\"", got)
	}
	if _, ok := data.Positionen[1]["kommentar"]; ok {
		t.Errorf("positionen[1] carries a kommentar key, want it absent: %v", data.Positionen[1])
	}
}

func TestBestellungAufnehmen_Conflict(t *testing.T) {
	ctx := context.Background()
	productMock := repotest.NewProduktRepo([]produkt.Produkt{testProduct}, nil)
	productMock.AddVariante(testProduct.ID, testVariant)
	eventMock := repotest.NewKassenjournalRepoWithWriteErr(nil, db.ErrAlreadyExists)
	command := newTestCommandWithEventMock([]tisch.Tisch{testActiveTisch}, eventMock)
	command.ProduktRepo = productMock

	inputs := []enrichment.PositionInput{
		{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 1},
	}

	err := command.BestellungAufnehmen(ctx, 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", 1, inputs, "")
	if !errors.Is(err, ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}
}

func TestBestellungAufnehmen_InactiveTisch(t *testing.T) {
	ctx := context.Background()
	productMock := repotest.NewProduktRepo([]produkt.Produkt{testProduct}, nil)
	productMock.AddVariante(testProduct.ID, testVariant)
	command := newTestCommand([]tisch.Tisch{testInactiveTisch})
	command.ProduktRepo = productMock

	inputs := []enrichment.PositionInput{
		{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 1},
	}

	err := command.BestellungAufnehmen(ctx, 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", testInactiveTisch.ID, inputs, "")
	if !errors.Is(err, ErrTischNotActive) {
		t.Errorf("expected ErrTischNotActive, got %v", err)
	}
}

// Ein deaktivierte (inactive) Variante taucht im Bestell-Menü nicht auf, kann
// aber per direktem POST referenziert werden — der Command lehnt sie ab.
func TestBestellungAufnehmen_InactiveVariante(t *testing.T) {
	ctx := context.Background()
	inactiveVariant := produkt.Variante{
		ID:         2,
		Name:       "Cola 0,5l",
		PreisCents: 350,
		Status:     produkt.InactiveStatus,
	}
	productMock := repotest.NewProduktRepo([]produkt.Produkt{testProduct}, nil)
	productMock.AddVariante(testProduct.ID, inactiveVariant)
	command := newTestCommand([]tisch.Tisch{testActiveTisch})
	command.ProduktRepo = productMock

	inputs := []enrichment.PositionInput{
		{ProduktID: testProduct.ID, VarianteID: inactiveVariant.ID, Menge: 1},
	}

	err := command.BestellungAufnehmen(ctx, 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", testActiveTisch.ID, inputs, "")
	if !errors.Is(err, enrichment.ErrVarianteNichtAktiv) {
		t.Errorf("expected ErrVarianteNichtAktiv, got %v", err)
	}
}

// duplikatTestSession liefert eine Tisch-Session mit einer Position (Menge 3),
// unbezahlt, für die Duplikat-Ablehnungstests.
func duplikatTestSession() kasse.TischSession {
	pos := kasse.Position{PositionID: "pos-1", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 3}
	return kasse.TischSession{
		SaldoCents:           1050,
		UnbezahltePositionen: []kasse.Position{pos},
	}
}

// Duplikate sind per se ungültig — auch wenn die Summe der Mengen (1+1) die
// verfügbare Menge (3) nicht übersteigt.
var duplikatRefs = []kasse.PositionRef{
	{PositionID: "pos-1", Menge: 1},
	{PositionID: "pos-1", Menge: 1},
}
