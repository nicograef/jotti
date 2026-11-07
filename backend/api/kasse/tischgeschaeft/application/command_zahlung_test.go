package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/repository/repotest"
)

func TestZahlungKassieren_NonOrderedPosition(t *testing.T) {
	ctx := context.Background()
	// No order events exist — paying a non-existent position should fail
	command := newTestCommand([]tisch.Tisch{testActiveTisch})

	fakeRefs := []kasse.PositionRef{
		{PositionID: "00000000-0000-0000-0000-000000000001", Menge: 1},
	}

	err := command.ZahlungKassieren(ctx, 1, "Test User", testActiveTisch.ID, fakeRefs, "", nil)
	if !errors.Is(err, ErrPositionNichtBezahlbar) {
		t.Errorf("expected ErrPositionNichtBezahlbar, got %v", err)
	}
}

func TestZahlungKassieren_DoublePayment(t *testing.T) {
	ctx := context.Background()
	// After order + payment, unbezahlt is empty
	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)
	eventMock.SetTischSession(subject, kasse.TischSession{
		SaldoCents:           0,
		UnbezahltePositionen: []kasse.Position{},
		GesamtZahlungenCents: 350,
	})
	command := Command{
		TischRepo:           repotest.NewTischRepo([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		ProduktRepo:         repotest.NewProduktRepo(nil, nil),
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
	}

	refs := []kasse.PositionRef{
		{PositionID: "00000000-0000-0000-0000-000000000001", Menge: 1},
	}

	// Try to pay again — should fail
	err := command.ZahlungKassieren(ctx, 1, "Test User", testActiveTisch.ID, refs, "", nil)
	if !errors.Is(err, ErrPositionNichtBezahlbar) {
		t.Errorf("expected ErrPositionNichtBezahlbar, got %v", err)
	}
}

// Read–validate–write race: the command validated against LastEventVersion 1 while a parallel Zahlung
// committed version 2. UNIQUE(subject, version) must fail the write: 409, not a double payment.
func TestZahlungKassieren_KonfliktBeiParallelemCommit(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	konkurrierendesEvent, err := kasse.NewZahlungKassiertEvent(subject, 2, "Andere Servicekraft",
		[]kasse.Position{{PositionID: "22222222-2222-4222-8222-222222222222", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 1}},
		350, "")
	if err != nil {
		t.Fatalf("failed to create concurrent event: %v", err)
	}
	konkurrierendesEvent.ID = 5
	konkurrierendesEvent.Version = 2

	eventMock := repotest.NewKassenjournalRepo([]event.Event{konkurrierendesEvent}, nil)
	// Die Projektion ist der Stand VOR dem parallelen Commit: Version 1, Position offen.
	eventMock.SetTischSession(subject, kasse.TischSession{
		SaldoCents: 350,
		UnbezahltePositionen: []kasse.Position{{
			PositionID:       "22222222-2222-4222-8222-222222222222",
			VarianteID:       1,
			ProduktName:      "Cola",
			VarianteName:     "0,5l",
			Kategorie:        "getraenk",
			Steuersatz:       "regel",
			EinzelpreisCents: 350,
			Menge:            1,
		}},
		LastEventVersion: 1,
	})

	command := Command{
		TischRepo:           repotest.NewTischRepo([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		ProduktRepo:         repotest.NewProduktRepo(nil, nil),
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
	}

	err = command.ZahlungKassieren(ctx, 1, "Test User", testActiveTisch.ID,
		[]kasse.PositionRef{{PositionID: "22222222-2222-4222-8222-222222222222", Menge: 1}}, "", nil)
	if !errors.Is(err, ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}
}

// Ohne parallelen Commit schreibt die Zahlung mit der Version des gelesenen Zustands + 1.
func TestZahlungKassieren_VersionAusGelesenerProjektion(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	eventMock.SetTischSession(subject, kasse.TischSession{
		SaldoCents: 350,
		UnbezahltePositionen: []kasse.Position{{
			PositionID:       "22222222-2222-4222-8222-222222222222",
			VarianteID:       1,
			ProduktName:      "Cola",
			VarianteName:     "0,5l",
			Kategorie:        "getraenk",
			Steuersatz:       "regel",
			EinzelpreisCents: 350,
			Menge:            1,
		}},
		LastEventVersion: 3,
	})

	command := Command{
		TischRepo:           repotest.NewTischRepo([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		ProduktRepo:         repotest.NewProduktRepo(nil, nil),
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
	}

	err := command.ZahlungKassieren(ctx, 1, "Test User", testActiveTisch.ID,
		[]kasse.PositionRef{{PositionID: "22222222-2222-4222-8222-222222222222", Menge: 1}}, "", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	events, err := eventMock.ReadEventsBySubject(ctx, subject)
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected exactly one event, got %d", len(events))
	}
	if events[0].Version != 4 {
		t.Errorf("expected version 4 (LastEventVersion 3 + 1), got %d", events[0].Version)
	}
}

// The Zahlung is the Beleg: it freezes the Pfand flag the product carries at payment, since the
// Bestellung event has none. The DSFinV-K export books the flagged line as GV_TYP Pfand.
func TestZahlungKassieren_FriertPfandKennzeichenEin(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	eventMock.SetTischSession(subject, kasse.TischSession{
		SaldoCents: 550,
		UnbezahltePositionen: []kasse.Position{
			{PositionID: "22222222-2222-4222-8222-222222222222", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 1},
			{PositionID: "33333333-3333-4333-8333-333333333333", VarianteID: 19, ProduktName: "Pfand", VarianteName: "Glas", Kategorie: "sonstiges", Steuersatz: "regel", EinzelpreisCents: 200, Menge: 1},
		},
		LastEventVersion: 1,
	})
	produkte := repotest.NewProduktRepo([]produkt.Produkt{testProduct, {ID: 9, Name: "Pfand", Pfand: true}}, nil)
	produkte.AddVariante(1, testVariant)
	produkte.AddVariante(9, produkt.Variante{ID: 19, Name: "Glas", PreisCents: 200, Status: produkt.ActiveStatus})

	command := Command{
		TischRepo:           repotest.NewTischRepo([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		ProduktRepo:         produkte,
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
	}

	err := command.ZahlungKassieren(ctx, 1, "Test User", testActiveTisch.ID, []kasse.PositionRef{
		{PositionID: "22222222-2222-4222-8222-222222222222", Menge: 1},
		{PositionID: "33333333-3333-4333-8333-333333333333", Menge: 1},
	}, "", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	events, err := eventMock.ReadEventsBySubject(ctx, subject)
	if err != nil || len(events) != 1 {
		t.Fatalf("expected one event, got %d (err %v)", len(events), err)
	}
	if events[0].Type != "zahlung-kassiert:v2" {
		t.Fatalf("type = %s, want zahlung-kassiert:v2", events[0].Type)
	}
	var data kasse.ZahlungKassiertV2Data
	if err := json.Unmarshal(events[0].Data, &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got := map[int]bool{}
	for _, p := range data.Positionen {
		got[p.VarianteID] = p.Pfand
	}
	if len(got) != 2 || got[1] || !got[19] {
		t.Errorf("pfand per variante = %v, want map[1:false 19:true]", got)
	}
}

// verspaeteteRueckgabeRepo hides the Rückgabe from the first lookup: a parallel attempt with the
// same rueckgabeId commits between the pre-check and the projection read.
type verspaeteteRueckgabeRepo struct {
	*repotest.KassenjournalRepo
	lookups int
}

func (r *verspaeteteRueckgabeRepo) GetPfandrueckgabe(ctx context.Context, rueckgabeID string) (event.Event, bool, error) {
	r.lookups++
	if r.lookups == 1 {
		return event.Event{}, false, nil
	}
	return r.KassenjournalRepo.GetPfandrueckgabe(ctx, rueckgabeID)
}

func TestZahlungKassieren_ParallelerVersuchMitPfand_Idempotent(t *testing.T) {
	ctx := context.Background()
	const rueckgabeID = "33333333-3333-4333-8333-333333333333"
	const positionID = "22222222-2222-4222-8222-222222222222"
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	zahlung, err := kasse.NewZahlungKassiertEvent(subject, 1, "Test User",
		[]kasse.Position{{PositionID: positionID, VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 1}},
		350, "")
	if err != nil {
		t.Fatalf("failed to create zahlung: %v", err)
	}
	zahlung.ID, zahlung.Version = 5, 2
	rueckgabe, err := kasse.NewPfandZurueckgegebenEvent(kasse.PfandrueckgabeSubject(testKassensitzungNr, rueckgabeID), rueckgabeID, testActiveTisch.ID, 1, "Test User",
		[]kasse.Position{{VarianteID: 7, ProduktName: "Pfand", VarianteName: "Glas", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 200, Menge: 2}}, "")
	if err != nil {
		t.Fatalf("failed to create rueckgabe: %v", err)
	}
	rueckgabe = kasse.Verrechne(rueckgabe, zahlung)
	rueckgabe.ID, rueckgabe.Version = 6, 1

	eventMock := &verspaeteteRueckgabeRepo{KassenjournalRepo: repotest.NewKassenjournalRepo([]event.Event{zahlung, rueckgabe}, nil)}
	// Die Projektion ist der Stand NACH dem parallelen Commit: die Position ist bezahlt.
	eventMock.SetTischSession(subject, kasse.TischSession{LastEventVersion: 2})

	command := Command{
		TischRepo:           repotest.NewTischRepo([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		ProduktRepo:         repotest.NewProduktRepo(nil, nil),
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
	}

	pfand := &enrichment.PfandEingabe{
		RueckgabeID: rueckgabeID,
		Positionen:  []enrichment.PositionInput{{ProduktID: 7, VarianteID: 7, Menge: 2}},
	}
	err = command.ZahlungKassieren(ctx, 1, "Test User", testActiveTisch.ID,
		[]kasse.PositionRef{{PositionID: positionID, Menge: 1}}, "", pfand)
	if err != nil {
		t.Fatalf("expected idempotent success, got %v", err)
	}
}

// A Pfand position of a product without the Pfand flag rejects the whole settled payment.
func TestZahlungKassieren_UngueltigesPfand_KeinEvent(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	eventMock.SetTischSession(subject, kasse.TischSession{
		SaldoCents: 350,
		UnbezahltePositionen: []kasse.Position{{
			PositionID:       "22222222-2222-4222-8222-222222222222",
			VarianteID:       testVariant.ID,
			ProduktName:      "Cola",
			VarianteName:     "0,5l",
			Kategorie:        "getraenk",
			Steuersatz:       "regel",
			EinzelpreisCents: 350,
			Menge:            1,
		}},
		LastEventVersion: 1,
	})

	productMock := repotest.NewProduktRepo([]produkt.Produkt{testProduct}, nil)
	productMock.AddVariante(testProduct.ID, testVariant)

	command := Command{
		TischRepo:           repotest.NewTischRepo([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		ProduktRepo:         productMock,
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
	}

	pfand := &enrichment.PfandEingabe{
		RueckgabeID: "33333333-3333-4333-8333-333333333333",
		Positionen:  []enrichment.PositionInput{{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 1}},
	}
	err := command.ZahlungKassieren(ctx, 1, "Test User", testActiveTisch.ID,
		[]kasse.PositionRef{{PositionID: "22222222-2222-4222-8222-222222222222", Menge: 1}}, "", pfand)
	if !errors.Is(err, enrichment.ErrKeinPfandProdukt) {
		t.Fatalf("expected ErrKeinPfandProdukt, got %v", err)
	}

	events, err := eventMock.ReadKassensitzungEvents(ctx, testKassensitzungNr)
	if err != nil {
		t.Fatalf("expected no read error, got %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected no event, got %d", len(events))
	}
}

func TestZahlungKassieren_ExceedsAvailableMenge(t *testing.T) {
	ctx := context.Background()
	// State has 1 position with Menge 1
	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)
	eventMock.SetTischSession(subject, kasse.TischSession{
		SaldoCents: 350,
		UnbezahltePositionen: []kasse.Position{
			{PositionID: "pos-1", VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 1},
		},
	})
	command := Command{
		TischRepo:           repotest.NewTischRepo([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		ProduktRepo:         repotest.NewProduktRepo(nil, nil),
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
	}

	// Try to pay for Menge 2 when only 1 was ordered
	refs := []kasse.PositionRef{
		{PositionID: "pos-1", Menge: 2},
	}

	err := command.ZahlungKassieren(ctx, 1, "Test User", testActiveTisch.ID, refs, "", nil)
	if !errors.Is(err, ErrPositionNichtBezahlbar) {
		t.Errorf("expected ErrPositionNichtBezahlbar, got %v", err)
	}
}

func TestZahlungKassieren_DuplikatPositionRefs(t *testing.T) {
	ctx := context.Background()
	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	eventMock.SetTischSession(kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID), duplikatTestSession())
	command := Command{
		TischRepo:           repotest.NewTischRepo([]tisch.Tisch{testActiveTisch}, nil),
		EventRepo:           eventMock,
		ProduktRepo:         repotest.NewProduktRepo(nil, nil),
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
	}

	err := command.ZahlungKassieren(ctx, 1, "Test User", testActiveTisch.ID, duplikatRefs, "", nil)
	if !errors.Is(err, ErrPositionNichtBezahlbar) {
		t.Errorf("expected ErrPositionNichtBezahlbar, got %v", err)
	}
}
