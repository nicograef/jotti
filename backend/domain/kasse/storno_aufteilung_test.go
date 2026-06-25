//go:build unit

package kasse

import (
	"encoding/json"
	"testing"

	e "github.com/nicograef/jotti/backend/domain/event"
)

// zahlungIDFromEvent liest die generierte ZahlungID eines zahlung-kassiert-Events.
func zahlungIDFromEvent(t *testing.T, evt e.Event) string {
	t.Helper()
	var data ZahlungKassiertV1Data
	if err := json.Unmarshal(evt.Data, &data); err != nil {
		t.Fatalf("unmarshal zahlung data: %v", err)
	}
	return data.ZahlungID
}

func posIDFromOrder(t *testing.T, orderEvent e.Event) string {
	t.Helper()
	bestellung, err := buildBestellungFromEvent(orderEvent)
	if err != nil {
		t.Fatalf("build bestellung: %v", err)
	}
	return bestellung.Positionen[0].PositionID
}

func TestComputeStornoAufteilung_PureUnpaid(t *testing.T) {
	order := mustCreateOrderEvent(t, testSubject, 1, []Position{testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 2)})
	posID := posIDFromOrder(t, order)

	aufteilung, ok := ComputeStornoAufteilung([]e.Event{order}, []PositionRef{{PositionID: posID, Menge: 1}})
	if !ok {
		t.Fatal("expected ok=true")
	}
	if len(aufteilung.Warenruecknahmen) != 0 {
		t.Fatalf("expected no Warenrücknahmen, got %d", len(aufteilung.Warenruecknahmen))
	}
	if len(aufteilung.Korrektur) != 1 || aufteilung.Korrektur[0].Menge != 1 {
		t.Fatalf("expected korrektur menge 1, got %+v", aufteilung.Korrektur)
	}
	if aufteilung.KorrekturCents != 500 {
		t.Fatalf("expected KorrekturCents 500, got %d", aufteilung.KorrekturCents)
	}
}

func TestComputeStornoAufteilung_PurePaidSingleZahlung(t *testing.T) {
	order := mustCreateOrderEvent(t, testSubject, 1, []Position{testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 2)})
	posID := posIDFromOrder(t, order)
	pay := mustCreatePaymentEvent(t, testSubject, 1, positionsFromOrder(t, order, 2), 1000)
	zahlungID := zahlungIDFromEvent(t, pay)

	aufteilung, ok := ComputeStornoAufteilung([]e.Event{order, pay}, []PositionRef{{PositionID: posID, Menge: 1}})
	if !ok {
		t.Fatal("expected ok=true")
	}
	if len(aufteilung.Korrektur) != 0 {
		t.Fatalf("expected no Korrektur, got %+v", aufteilung.Korrektur)
	}
	if len(aufteilung.Warenruecknahmen) != 1 {
		t.Fatalf("expected 1 Warenrücknahme, got %d", len(aufteilung.Warenruecknahmen))
	}
	wr := aufteilung.Warenruecknahmen[0]
	if wr.ZahlungID != zahlungID {
		t.Fatalf("expected ZahlungID %q, got %q", zahlungID, wr.ZahlungID)
	}
	if wr.GesamtCents != 500 || len(wr.Positionen) != 1 || wr.Positionen[0].Menge != 1 {
		t.Fatalf("expected 1x500 Warenrücknahme, got %+v", wr)
	}
}

func TestComputeStornoAufteilung_PaidAcrossTwoZahlungenFIFO(t *testing.T) {
	order := mustCreateOrderEvent(t, testSubject, 1, []Position{testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 4)})
	posID := posIDFromOrder(t, order)
	payA := mustCreatePaymentEvent(t, testSubject, 1, positionsFromOrder(t, order, 3), 1500)
	payB := mustCreatePaymentEvent(t, testSubject, 1, positionsFromOrder(t, order, 1), 500)
	order.ID, payA.ID, payB.ID = 1, 2, 3

	aufteilung, ok := ComputeStornoAufteilung([]e.Event{order, payA, payB}, []PositionRef{{PositionID: posID, Menge: 4}})
	if !ok {
		t.Fatal("expected ok=true")
	}
	if len(aufteilung.Korrektur) != 0 {
		t.Fatalf("expected no Korrektur, got %+v", aufteilung.Korrektur)
	}
	if len(aufteilung.Warenruecknahmen) != 2 {
		t.Fatalf("expected 2 Warenrücknahmen (one per Zahlung), got %d", len(aufteilung.Warenruecknahmen))
	}
	// FIFO: die ältere Zahlung (payA, 3 Stück) zuerst.
	if aufteilung.Warenruecknahmen[0].ZahlungID != zahlungIDFromEvent(t, payA) || aufteilung.Warenruecknahmen[0].GesamtCents != 1500 {
		t.Fatalf("expected first Warenrücknahme payA/1500, got %+v", aufteilung.Warenruecknahmen[0])
	}
	if aufteilung.Warenruecknahmen[1].ZahlungID != zahlungIDFromEvent(t, payB) || aufteilung.Warenruecknahmen[1].GesamtCents != 500 {
		t.Fatalf("expected second Warenrücknahme payB/500, got %+v", aufteilung.Warenruecknahmen[1])
	}
}

func TestComputeStornoAufteilung_MixedPrefersKorrektur(t *testing.T) {
	// Bestellt 3, bezahlt 1 → 2 unbezahlt. Storno 3 = 2 Korrektur (unbezahlt zuerst) + 1 Warenrücknahme.
	order := mustCreateOrderEvent(t, testSubject, 1, []Position{testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 3)})
	posID := posIDFromOrder(t, order)
	pay := mustCreatePaymentEvent(t, testSubject, 1, positionsFromOrder(t, order, 1), 500)
	order.ID, pay.ID = 1, 2

	aufteilung, ok := ComputeStornoAufteilung([]e.Event{order, pay}, []PositionRef{{PositionID: posID, Menge: 3}})
	if !ok {
		t.Fatal("expected ok=true")
	}
	if len(aufteilung.Korrektur) != 1 || aufteilung.Korrektur[0].Menge != 2 || aufteilung.KorrekturCents != 1000 {
		t.Fatalf("expected korrektur 2x500, got %+v (%d cents)", aufteilung.Korrektur, aufteilung.KorrekturCents)
	}
	if len(aufteilung.Warenruecknahmen) != 1 || aufteilung.Warenruecknahmen[0].GesamtCents != 500 {
		t.Fatalf("expected 1 Warenrücknahme 1x500, got %+v", aufteilung.Warenruecknahmen)
	}
}

func TestComputeStornoAufteilung_ExceedsStornierbar(t *testing.T) {
	order := mustCreateOrderEvent(t, testSubject, 1, []Position{testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 2)})
	posID := posIDFromOrder(t, order)

	if _, ok := ComputeStornoAufteilung([]e.Event{order}, []PositionRef{{PositionID: posID, Menge: 3}}); ok {
		t.Fatal("expected ok=false for over-request")
	}
}

func TestComputeStornoAufteilung_UnknownPosition(t *testing.T) {
	order := mustCreateOrderEvent(t, testSubject, 1, []Position{testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 2)})

	if _, ok := ComputeStornoAufteilung([]e.Event{order}, []PositionRef{{PositionID: "00000000-0000-0000-0000-000000000099", Menge: 1}}); ok {
		t.Fatal("expected ok=false for unknown position")
	}
}

func TestComputeStornoAufteilung_AlreadyCorrectedNotStornierbar(t *testing.T) {
	order := mustCreateOrderEvent(t, testSubject, 1, []Position{testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 1)})
	posID := posIDFromOrder(t, order)
	korrektur := mustCreateKorrekturEvent(t, testSubject, 1, positionsFromOrder(t, order, 1), 500)
	order.ID, korrektur.ID = 1, 2

	if _, ok := ComputeStornoAufteilung([]e.Event{order, korrektur}, []PositionRef{{PositionID: posID, Menge: 1}}); ok {
		t.Fatal("expected ok=false: position already corrected")
	}
}
