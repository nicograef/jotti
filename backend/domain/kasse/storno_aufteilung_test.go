package kasse

import (
	"encoding/json"
	"testing"

	e "github.com/nicograef/jotti/backend/domain/event"
)

func zahlungIDFromEvent(t *testing.T, evt e.Event) string {
	t.Helper()
	var data ZahlungKassiertV1Data
	if err := json.Unmarshal(evt.Data, &data); err != nil {
		t.Fatalf("unmarshal zahlung data: %v", err)
	}
	return data.ZahlungID
}

func positionIDAusBestellung(t *testing.T, orderEvent e.Event) string {
	t.Helper()
	bestellung, err := buildBestellungFromEvent(orderEvent)
	if err != nil {
		t.Fatalf("build bestellung: %v", err)
	}
	return bestellung.Positionen[0].PositionID
}

func TestComputeStornoAufteilung_PureUnpaid(t *testing.T) {
	order := mustCreateBestellungEvent(t, testSubject, 1, []Position{testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 2)})
	posID := positionIDAusBestellung(t, order)

	aufteilung, ok := ComputeStornoAufteilung([]e.Event{order}, []PositionRef{{PositionID: posID, Menge: 1}})
	if !ok {
		t.Fatal("expected ok=true")
	}
	if len(aufteilung.Warenruecknahmen) != 0 {
		t.Errorf("expected no Warenrücknahmen, got %d", len(aufteilung.Warenruecknahmen))
	}
	if len(aufteilung.Korrektur) != 1 || aufteilung.Korrektur[0].Menge != 1 {
		t.Errorf("expected korrektur menge 1, got %+v", aufteilung.Korrektur)
	}
	if aufteilung.KorrekturCents != 500 {
		t.Errorf("expected KorrekturCents 500, got %d", aufteilung.KorrekturCents)
	}
}

func TestComputeStornoAufteilung_DuplikatRefs(t *testing.T) {
	order := mustCreateBestellungEvent(t, testSubject, 1, []Position{testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 3)})
	posID := positionIDAusBestellung(t, order)

	// Duplikate sind per se ungültig, auch wenn 1+1 die stornierbare Menge 3 nicht übersteigt.
	refs := []PositionRef{{PositionID: posID, Menge: 1}, {PositionID: posID, Menge: 1}}
	if _, ok := ComputeStornoAufteilung([]e.Event{order}, refs); ok {
		t.Error("expected ok=false for duplicate refs")
	}
}

func TestComputeStornoAufteilung_PurePaidSingleZahlung(t *testing.T) {
	order := mustCreateBestellungEvent(t, testSubject, 1, []Position{testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 2)})
	posID := positionIDAusBestellung(t, order)
	pay := mustCreateZahlungEvent(t, testSubject, 1, positionenAusBestellung(t, order, 2), 1000)
	zahlungID := zahlungIDFromEvent(t, pay)

	aufteilung, ok := ComputeStornoAufteilung([]e.Event{order, pay}, []PositionRef{{PositionID: posID, Menge: 1}})
	if !ok {
		t.Fatal("expected ok=true")
	}
	if len(aufteilung.Korrektur) != 0 {
		t.Errorf("expected no Korrektur, got %+v", aufteilung.Korrektur)
	}
	if len(aufteilung.Warenruecknahmen) != 1 {
		t.Fatalf("expected 1 Warenrücknahme, got %d", len(aufteilung.Warenruecknahmen))
	}
	wr := aufteilung.Warenruecknahmen[0]
	if wr.ZahlungID != zahlungID {
		t.Errorf("expected ZahlungID %q, got %q", zahlungID, wr.ZahlungID)
	}
	if wr.GesamtCents != 500 || len(wr.Positionen) != 1 || wr.Positionen[0].Menge != 1 {
		t.Errorf("expected 1x500 Warenrücknahme, got %+v", wr)
	}
}

func TestComputeStornoAufteilung_PaidAcrossTwoZahlungenFIFO(t *testing.T) {
	order := mustCreateBestellungEvent(t, testSubject, 1, []Position{testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 4)})
	posID := positionIDAusBestellung(t, order)
	payA := mustCreateZahlungEvent(t, testSubject, 1, positionenAusBestellung(t, order, 3), 1500)
	payB := mustCreateZahlungEvent(t, testSubject, 1, positionenAusBestellung(t, order, 1), 500)
	order.ID, payA.ID, payB.ID = 1, 2, 3

	aufteilung, ok := ComputeStornoAufteilung([]e.Event{order, payA, payB}, []PositionRef{{PositionID: posID, Menge: 4}})
	if !ok {
		t.Fatal("expected ok=true")
	}
	if len(aufteilung.Korrektur) != 0 {
		t.Errorf("expected no Korrektur, got %+v", aufteilung.Korrektur)
	}
	if len(aufteilung.Warenruecknahmen) != 2 {
		t.Fatalf("expected 2 Warenrücknahmen (one per Zahlung), got %d", len(aufteilung.Warenruecknahmen))
	}
	// FIFO: die ältere Zahlung (payA, 3 Stück) zuerst.
	if aufteilung.Warenruecknahmen[0].ZahlungID != zahlungIDFromEvent(t, payA) || aufteilung.Warenruecknahmen[0].GesamtCents != 1500 {
		t.Errorf("expected first Warenrücknahme payA/1500, got %+v", aufteilung.Warenruecknahmen[0])
	}
	if aufteilung.Warenruecknahmen[1].ZahlungID != zahlungIDFromEvent(t, payB) || aufteilung.Warenruecknahmen[1].GesamtCents != 500 {
		t.Errorf("expected second Warenrücknahme payB/500, got %+v", aufteilung.Warenruecknahmen[1])
	}
}

func TestComputeStornoAufteilung_MixedPrefersKorrektur(t *testing.T) {
	// Bestellt 3, bezahlt 1 → 2 unbezahlt. Storno 3 = 2 Korrektur (unbezahlt zuerst) + 1 Warenrücknahme.
	order := mustCreateBestellungEvent(t, testSubject, 1, []Position{testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 3)})
	posID := positionIDAusBestellung(t, order)
	pay := mustCreateZahlungEvent(t, testSubject, 1, positionenAusBestellung(t, order, 1), 500)
	order.ID, pay.ID = 1, 2

	aufteilung, ok := ComputeStornoAufteilung([]e.Event{order, pay}, []PositionRef{{PositionID: posID, Menge: 3}})
	if !ok {
		t.Fatal("expected ok=true")
	}
	if len(aufteilung.Korrektur) != 1 || aufteilung.Korrektur[0].Menge != 2 || aufteilung.KorrekturCents != 1000 {
		t.Errorf("expected korrektur 2x500, got %+v (%d cents)", aufteilung.Korrektur, aufteilung.KorrekturCents)
	}
	if len(aufteilung.Warenruecknahmen) != 1 || aufteilung.Warenruecknahmen[0].GesamtCents != 500 {
		t.Errorf("expected 1 Warenrücknahme 1x500, got %+v", aufteilung.Warenruecknahmen)
	}
}

func TestComputeStornoAufteilung_ExceedsStornierbar(t *testing.T) {
	order := mustCreateBestellungEvent(t, testSubject, 1, []Position{testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 2)})
	posID := positionIDAusBestellung(t, order)

	if _, ok := ComputeStornoAufteilung([]e.Event{order}, []PositionRef{{PositionID: posID, Menge: 3}}); ok {
		t.Error("expected ok=false for over-request")
	}
}

func TestComputeStornoAufteilung_UnknownPosition(t *testing.T) {
	order := mustCreateBestellungEvent(t, testSubject, 1, []Position{testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 2)})

	if _, ok := ComputeStornoAufteilung([]e.Event{order}, []PositionRef{{PositionID: "00000000-0000-0000-0000-000000000099", Menge: 1}}); ok {
		t.Error("expected ok=false for unknown position")
	}
}

func TestComputeStornoAufteilung_AlreadyCorrectedNotStornierbar(t *testing.T) {
	order := mustCreateBestellungEvent(t, testSubject, 1, []Position{testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 1)})
	posID := positionIDAusBestellung(t, order)
	korrektur := mustCreateKorrekturEvent(t, testSubject, 1, positionenAusBestellung(t, order, 1), 500)
	order.ID, korrektur.ID = 1, 2

	if _, ok := ComputeStornoAufteilung([]e.Event{order, korrektur}, []PositionRef{{PositionID: posID, Menge: 1}}); ok {
		t.Error("expected ok=false: position already corrected")
	}
}

func TestComputeStornoAufteilung_NachStornierung(t *testing.T) {
	// Bestellt 2, beide bezahlt, 1 bereits storniert → nur noch 1 aus dieser Zahlung stornierbar.
	order := mustCreateBestellungEvent(t, testSubject, 1, []Position{testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 2)})
	posID := positionIDAusBestellung(t, order)
	pay := mustCreateZahlungEvent(t, testSubject, 1, positionenAusBestellung(t, order, 2), 1000)
	zahlungID := zahlungIDFromEvent(t, pay)
	storno := mustCreateStornierungEvent(t, testSubject, 1, zahlungID, positionenAusBestellung(t, order, 1), 500)
	events := []e.Event{order, pay, storno}

	if _, ok := ComputeStornoAufteilung(events, []PositionRef{{PositionID: posID, Menge: 2}}); ok {
		t.Error("Menge 2: expected ok=false, one piece is already storniert")
	}

	aufteilung, ok := ComputeStornoAufteilung(events, []PositionRef{{PositionID: posID, Menge: 1}})
	if !ok {
		t.Fatal("Menge 1: expected ok=true")
	}
	if len(aufteilung.Korrektur) != 0 || aufteilung.KorrekturCents != 0 {
		t.Errorf("expected no Korrektur, got %+v (%d cents)", aufteilung.Korrektur, aufteilung.KorrekturCents)
	}
	if len(aufteilung.Warenruecknahmen) != 1 {
		t.Fatalf("expected 1 Warenrücknahme, got %d", len(aufteilung.Warenruecknahmen))
	}
	wr := aufteilung.Warenruecknahmen[0]
	if wr.ZahlungID != zahlungID || wr.GesamtCents != 500 || len(wr.Positionen) != 1 || wr.Positionen[0].Menge != 1 {
		t.Errorf("expected Warenrücknahme %s 1x500, got %+v", zahlungID, wr)
	}
}

func TestComputeStornoAufteilung_NachUmbuchung(t *testing.T) {
	quellSubject, zielSubject := TischSessionSubject(1, 1), TischSessionSubject(1, 2)
	order := mustCreateBestellungEvent(t, quellSubject, 1, []Position{testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 2)})
	posID := positionIDAusBestellung(t, order)
	quelle, ziel, err := NewBestellungUmgebuchtEvents(1, "55555555-5555-4555-8555-555555555555", 1, 2, 1, "TestUser",
		positionenAusBestellung(t, order, 1), 500, "Umbuchung zu Tisch 2", "Umbuchung von Tisch 1", "")
	if err != nil {
		t.Fatalf("umbuchung events: %v", err)
	}
	if quelle.Subject != quellSubject || ziel.Subject != zielSubject {
		t.Fatalf("subjects: got %q/%q", quelle.Subject, ziel.Subject)
	}
	var zielData BestellungUmgebuchtV1Data
	if err := json.Unmarshal(ziel.Data, &zielData); err != nil {
		t.Fatalf("unmarshal ziel data: %v", err)
	}
	zielPosID := zielData.Positionen[0].PositionID

	cases := []struct {
		name   string
		events []e.Event
		ref    PositionRef
		ok     bool
	}{
		{"Quelle Menge 2 übersteigt den Rest", []e.Event{order, quelle}, PositionRef{PositionID: posID, Menge: 2}, false},
		{"Quelle Menge 1", []e.Event{order, quelle}, PositionRef{PositionID: posID, Menge: 1}, true},
		{"Ziel Zugang Menge 1", []e.Event{ziel}, PositionRef{PositionID: zielPosID, Menge: 1}, true},
		{"Ziel Zugang Menge 2 übersteigt den Zugang", []e.Event{ziel}, PositionRef{PositionID: zielPosID, Menge: 2}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			aufteilung, ok := ComputeStornoAufteilung(c.events, []PositionRef{c.ref})
			if ok != c.ok {
				t.Fatalf("ok: got %v, want %v", ok, c.ok)
			}
			if !c.ok {
				return
			}
			if len(aufteilung.Warenruecknahmen) != 0 {
				t.Errorf("expected no Warenrücknahmen, got %+v", aufteilung.Warenruecknahmen)
			}
			if len(aufteilung.Korrektur) != 1 || aufteilung.Korrektur[0].PositionID != c.ref.PositionID ||
				aufteilung.Korrektur[0].Menge != 1 || aufteilung.KorrekturCents != 500 {
				t.Errorf("expected Korrektur %s 1x500, got %+v (%d cents)", c.ref.PositionID, aufteilung.Korrektur, aufteilung.KorrekturCents)
			}
		})
	}
}

func TestComputeStornoAufteilung_FremdesEventImStream(t *testing.T) {
	order := mustCreateBestellungEvent(t, testSubject, 1, []Position{testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 2)})
	posID := positionIDAusBestellung(t, order)
	eroeffnet, err := NewKassensitzungEroeffnetEvent(KassensitzungSubject(1), 1, "TestUser", "2026-07-07", "Maihock 2026", 5000)
	if err != nil {
		t.Fatalf("eroeffnet event: %v", err)
	}

	if _, ok := ComputeStornoAufteilung([]e.Event{order, eroeffnet}, []PositionRef{{PositionID: posID, Menge: 1}}); ok {
		t.Error("expected ok=false for a kassensitzung-eroeffnet:v1 event in a Tisch stream")
	}
}
