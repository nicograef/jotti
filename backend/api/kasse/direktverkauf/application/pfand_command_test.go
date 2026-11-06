package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/repository/repotest"
)

func newPfandCommand(eventRepo eventRepo) Command {
	pfandProdukt := produkt.Produkt{ID: 2, Name: "Pfand", Kategorie: produkt.SonstigesKategorie, Steuersatz: steuer.RegelSteuersatz, Status: produkt.ActiveStatus, Pfand: true}
	produkte := repotest.NewProduktRepo([]produkt.Produkt{testProduct, pfandProdukt}, nil)
	produkte.AddVariante(testProduct.ID, testVariant)
	produkte.AddVariante(pfandProdukt.ID, produkt.Variante{ID: 2, Name: "Glas", PreisCents: 200, Status: produkt.ActiveStatus})
	return Command{
		EventRepo:           eventRepo,
		ProduktRepo:         produkte,
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
	}
}

func TestPfandZurueckgeben_KeinPfandProdukt_KeinEvent(t *testing.T) {
	spy := &spyEventRepo{}
	cmd := newPfandCommand(spy)

	inputs := []enrichment.PositionInput{
		{ProduktID: 2, VarianteID: 2, Menge: 1},
		{ProduktID: testProduct.ID, VarianteID: testVariant.ID, Menge: 1},
	}
	err := cmd.PfandZurueckgeben(context.Background(), 1, "Test", uuid.New().String(), 0, inputs, "")
	if !errors.Is(err, enrichment.ErrKeinPfandProdukt) {
		t.Fatalf("erwartet ErrKeinPfandProdukt, bekam %v", err)
	}
	if len(spy.written) != 0 {
		t.Errorf("erwartet kein Event, geschrieben: %d", len(spy.written))
	}
}

func TestPfandZurueckgeben_SchreibtVersion1InEigenenStream(t *testing.T) {
	spy := &spyEventRepo{}
	cmd := newPfandCommand(spy)
	rueckgabeID := uuid.New().String()

	inputs := []enrichment.PositionInput{{ProduktID: 2, VarianteID: 2, Menge: 2}}
	if err := cmd.PfandZurueckgeben(context.Background(), 1, "Test", rueckgabeID, 0, inputs, "zwei Gläser"); err != nil {
		t.Fatalf("PfandZurueckgeben: %v", err)
	}

	if len(spy.written) != 1 {
		t.Fatalf("erwartet 1 Event, geschrieben: %d", len(spy.written))
	}
	w := spy.written[0]
	if w.streamType != kasse.StreamTypePfandrueckgabe {
		t.Errorf("streamType = %q, erwartet %q", w.streamType, kasse.StreamTypePfandrueckgabe)
	}
	if w.event.Subject != "kassensitzung-1/pfandrueckgabe-"+rueckgabeID {
		t.Errorf("subject = %q", w.event.Subject)
	}
	if w.event.Version != 1 {
		t.Errorf("version = %d, erwartet 1", w.event.Version)
	}
	var data kasse.PfandZurueckgegebenV1Data
	if err := json.Unmarshal(w.event.Data, &data); err != nil {
		t.Fatalf("event-daten: %v", err)
	}
	if data.GesamtbetragCents != 400 || data.TischID != 0 || data.Kommentar != "zwei Gläser" {
		t.Errorf("data = %+v", data)
	}
}
