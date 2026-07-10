//go:build unit

package kasse

import (
	"testing"

	e "github.com/nicograef/jotti/backend/domain/event"
)

func TestGetHistorieFromEvents_Empty(t *testing.T) {
	history, err := GetHistorieFromEvents([]e.Event{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(history) != 0 {
		t.Fatalf("expected empty history, got %d items", len(history))
	}
}

func TestGetHistorieFromEvents_ReturnsAllEventTypes(t *testing.T) {
	products := []Position{
		testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 1),
	}
	orderEvent := mustCreateOrderEvent(t, testSubject, 1, products)
	positions := positionsFromOrder(t, orderEvent, 1)

	events := []e.Event{
		orderEvent,
		mustCreatePaymentEvent(t, testSubject, 1, positions, 500),
		mustCreateCancelationEvent(t, testSubject, 1, testZahlungID, positions, 500),
	}

	history, err := GetHistorieFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(history) != 3 {
		t.Fatalf("expected 3 history items, got %d", len(history))
	}
}

func TestGetHistorieFromEvents_BestellungTraegtBestellerName(t *testing.T) {
	products := []Position{
		testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 1),
	}
	orderEvent, err := NewBestellungAufgenommenEvent(testSubject, 7, "Anna", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", products, "")
	if err != nil {
		t.Fatalf("failed to create order event: %v", err)
	}

	history, err := GetHistorieFromEvents([]e.Event{orderEvent})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(history) != 1 || history[0].Bestellung == nil {
		t.Fatalf("expected one Bestellung entry, got %+v", history)
	}
	if history[0].Bestellung.UserName != "Anna" {
		t.Errorf("expected besteller name Anna, got %q", history[0].Bestellung.UserName)
	}
}

func TestGetHistorieFromEvents_EnrichesBestellungMitRestmengen(t *testing.T) {
	products := []Position{
		testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 3),
	}
	orderEvent := mustCreateOrderEvent(t, testSubject, 1, products)
	eine := positionsFromOrder(t, orderEvent, 1)

	events := []e.Event{
		orderEvent,
		mustCreateCancelationEvent(t, testSubject, 1, testZahlungID, eine, 500),
		mustCreatePaymentEvent(t, testSubject, 1, eine, 500),
	}

	history, err := GetHistorieFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// History is reversed (most recent first); the Bestellung is the last entry.
	bestellung := history[len(history)-1]
	if bestellung.Art != HistorieEintragBestellung {
		t.Fatalf("expected last entry to be Bestellung, got %q", bestellung.Art)
	}

	// Ordered 3, cancelled 1 => 2 still stornierbar.
	if len(bestellung.StornierbarePositionen) != 1 || bestellung.StornierbarePositionen[0].Menge != 2 {
		t.Fatalf("expected stornierbar menge 2, got %+v", bestellung.StornierbarePositionen)
	}
	// Ordered 3, cancelled 1, paid 1 => 1 still umbuchbar.
	if len(bestellung.UmbuchbarePositionen) != 1 || bestellung.UmbuchbarePositionen[0].Menge != 1 {
		t.Fatalf("expected umbuchbar menge 1, got %+v", bestellung.UmbuchbarePositionen)
	}
}

func TestGetHistorieFromEvents_FullyConsumedBestellungHasNoRestmengen(t *testing.T) {
	products := []Position{
		testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 1),
	}
	orderEvent := mustCreateOrderEvent(t, testSubject, 1, products)
	positions := positionsFromOrder(t, orderEvent, 1)

	events := []e.Event{
		orderEvent,
		mustCreateCancelationEvent(t, testSubject, 1, testZahlungID, positions, 500),
	}

	history, err := GetHistorieFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	bestellung := history[len(history)-1]
	if len(bestellung.StornierbarePositionen) != 0 {
		t.Fatalf("expected no stornierbare positionen, got %+v", bestellung.StornierbarePositionen)
	}
	if len(bestellung.UmbuchbarePositionen) != 0 {
		t.Fatalf("expected no umbuchbare positionen, got %+v", bestellung.UmbuchbarePositionen)
	}
}

func TestBuildStornierung_BarRueckgabeLeitetStornoArtAusEventTypAb(t *testing.T) {
	products := []Position{
		testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 1),
	}
	orderEvent := mustCreateOrderEvent(t, testSubject, 1, products)
	positions := positionsFromOrder(t, orderEvent, 1)

	warenruecknahme := mustCreateCancelationEvent(t, testSubject, 1, testZahlungID, positions, 500)
	korrektur := mustCreateKorrekturEvent(t, testSubject, 1, positions, 500)

	storno, err := buildStornierungFromEvent(warenruecknahme)
	if err != nil {
		t.Fatalf("expected no error for warenruecknahme, got %v", err)
	}
	if !storno.BarRueckgabe {
		t.Error("expected BarRueckgabe true for stornierung-erteilt:v1 (Warenrücknahme), got false")
	}

	kor, err := buildKorrekturFromEvent(korrektur)
	if err != nil {
		t.Fatalf("expected no error for korrektur, got %v", err)
	}
	if kor.BarRueckgabe {
		t.Error("expected BarRueckgabe false for bestellung-korrigiert:v1 (geldneutrale Korrektur), got true")
	}
}

func TestGetHistorieFromEvents_ReversesOrder(t *testing.T) {
	products := []Position{
		testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 1),
	}
	orderEvent := mustCreateOrderEvent(t, testSubject, 1, products)
	positions := positionsFromOrder(t, orderEvent, 1)

	events := []e.Event{
		orderEvent,
		mustCreatePaymentEvent(t, testSubject, 1, positions, 500),
	}

	history, err := GetHistorieFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// First item should be payment (last event), second should be order (first event)
	if history[0].Art != HistorieEintragZahlung || history[0].Zahlung == nil {
		t.Fatalf("expected first item to be Zahlung, got kind %q", history[0].Art)
	}
	if history[1].Art != HistorieEintragBestellung || history[1].Bestellung == nil {
		t.Fatalf("expected second item to be Bestellung, got kind %q", history[1].Art)
	}
}

func TestGetHistorieFromEvents_UmbuchungAbgangReduziertRestmengen(t *testing.T) {
	const zNr, quellTischID, zielTischID = 1, 42, 7
	quellSubject := TischSessionSubject(zNr, quellTischID)

	products := []Position{
		testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 3),
	}
	orderEvent := mustCreateOrderEvent(t, quellSubject, 1, products)
	eine := positionsFromOrder(t, orderEvent, 1)

	quellEvent, _, err := NewBestellungUmgebuchtEvents(zNr, quellTischID, zielTischID, 1, "TestUser", eine, 500, "Umbuchung auf Tisch Ziel", "Umbuchung von Tisch Quelle")
	if err != nil {
		t.Fatalf("failed to create umbuchung events: %v", err)
	}

	history, err := GetHistorieFromEvents([]e.Event{orderEvent, quellEvent})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Most recent first: the umbuchung (abgang) is at index 0, the Bestellung last.
	if history[0].Art != HistorieEintragUmbuchung || history[0].Umbuchung == nil {
		t.Fatalf("expected first entry to be Umbuchung, got %q", history[0].Art)
	}
	if history[0].Umbuchung.IstZugang() {
		t.Fatal("expected the source entry to be an Abgang, not a Zugang")
	}

	bestellung := history[len(history)-1]
	// Ordered 3, moved away 1 => 2 still stornierbar and umbuchbar.
	if len(bestellung.StornierbarePositionen) != 1 || bestellung.StornierbarePositionen[0].Menge != 2 {
		t.Fatalf("expected stornierbar menge 2, got %+v", bestellung.StornierbarePositionen)
	}
	if len(bestellung.UmbuchbarePositionen) != 1 || bestellung.UmbuchbarePositionen[0].Menge != 2 {
		t.Fatalf("expected umbuchbar menge 2, got %+v", bestellung.UmbuchbarePositionen)
	}
}

func TestGetHistorieFromEvents_UmbuchungZugangIstStornierbar(t *testing.T) {
	const zNr, quellTischID, zielTischID = 1, 42, 7
	zielSubject := TischSessionSubject(zNr, zielTischID)

	positionen := []Position{testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 2)}
	// PositionIDs werden im Quell-Event vergeben; für den Zugang reicht eine ID.
	positionen[0].PositionID = "11111111-1111-4111-8111-111111111111"

	_, zielEvent, err := NewBestellungUmgebuchtEvents(zNr, quellTischID, zielTischID, 1, "TestUser", positionen, 1000, "Umbuchung auf Tisch Ziel", "Umbuchung von Tisch Quelle")
	if err != nil {
		t.Fatalf("failed to create umbuchung events: %v", err)
	}
	if zielEvent.Subject != zielSubject {
		t.Fatalf("expected ziel subject %q, got %q", zielSubject, zielEvent.Subject)
	}

	history, err := GetHistorieFromEvents([]e.Event{zielEvent})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(history) != 1 || history[0].Umbuchung == nil {
		t.Fatalf("expected one Umbuchung entry, got %+v", history)
	}
	zugang := history[0]
	if !zugang.Umbuchung.IstZugang() {
		t.Fatal("expected the target entry to be a Zugang")
	}
	// Der Zugang bringt Positionen auf den Tisch: voll stornierbar/umbuchbar.
	if len(zugang.StornierbarePositionen) != 1 || zugang.StornierbarePositionen[0].Menge != 2 {
		t.Fatalf("expected stornierbar menge 2, got %+v", zugang.StornierbarePositionen)
	}
	if len(zugang.UmbuchbarePositionen) != 1 || zugang.UmbuchbarePositionen[0].Menge != 2 {
		t.Fatalf("expected umbuchbar menge 2, got %+v", zugang.UmbuchbarePositionen)
	}
}
