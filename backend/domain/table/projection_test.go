//go:build unit

package table

import (
	"encoding/json"
	"testing"

	e "github.com/nicograef/jotti/backend/domain/event"
)

func TestApplyEvent_BestellungOnEmptyTable(t *testing.T) {
	products := []Position{
		testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 2),
	}
	orderEvent := mustCreateOrderEvent(t, 1, 1, products)
	orderEvent.ID = 1
	orderEvent.Version = 1

	state, err := ApplyEvent(TischState{}, orderEvent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if state.SaldoCents != 1000 {
		t.Fatalf("expected SaldoCents 1000, got %d", state.SaldoCents)
	}
	if len(state.UnbezahltePositionen) != 1 {
		t.Fatalf("expected 1 unbezahlte position, got %d", len(state.UnbezahltePositionen))
	}
	if state.UnbezahltePositionen[0].Menge != 2 {
		t.Fatalf("expected Menge 2, got %d", state.UnbezahltePositionen[0].Menge)
	}
	if len(state.AusstehendePositionen) != 1 {
		t.Fatalf("expected 1 ungelieferte position, got %d", len(state.AusstehendePositionen))
	}
	if state.AusstehendePositionen[0].Menge != 2 {
		t.Fatalf("expected Menge 2, got %d", state.AusstehendePositionen[0].Menge)
	}
	if state.GesamtZahlungenCents != 0 {
		t.Fatalf("expected GesamtZahlungenCents 0, got %d", state.GesamtZahlungenCents)
	}
	if state.LastEventID != 1 {
		t.Fatalf("expected LastEventID 1, got %d", state.LastEventID)
	}
	if state.LastEventVersion != 1 {
		t.Fatalf("expected LastEventVersion 1, got %d", state.LastEventVersion)
	}
}

func TestApplyEvent_ZahlungReducesSaldoAndUnbezahlt(t *testing.T) {
	products := []Position{
		testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 2),
	}
	orderEvent := mustCreateOrderEvent(t, 1, 1, products)
	orderEvent.ID = 1
	orderEvent.Version = 1

	state, err := ApplyEvent(TischState{}, orderEvent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	positions := positionsFromOrder(t, orderEvent, 1)
	paymentEvent := mustCreatePaymentEvent(t, 1, 1, positions, 500)
	paymentEvent.ID = 2
	paymentEvent.Version = 2

	state, err = ApplyEvent(state, paymentEvent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if state.SaldoCents != 500 {
		t.Fatalf("expected SaldoCents 500, got %d", state.SaldoCents)
	}
	if state.GesamtZahlungenCents != 500 {
		t.Fatalf("expected GesamtZahlungenCents 500, got %d", state.GesamtZahlungenCents)
	}
	if len(state.UnbezahltePositionen) != 1 {
		t.Fatalf("expected 1 unbezahlte position, got %d", len(state.UnbezahltePositionen))
	}
	if state.UnbezahltePositionen[0].Menge != 1 {
		t.Fatalf("expected Menge 1, got %d", state.UnbezahltePositionen[0].Menge)
	}
	if state.LastEventID != 2 {
		t.Fatalf("expected LastEventID 2, got %d", state.LastEventID)
	}
}

func TestApplyEvent_StornierungReducesSaldoAndUnbezahlt(t *testing.T) {
	products := []Position{
		testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 2),
	}
	orderEvent := mustCreateOrderEvent(t, 1, 1, products)
	orderEvent.ID = 1
	orderEvent.Version = 1

	state, err := ApplyEvent(TischState{}, orderEvent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	positions := positionsFromOrder(t, orderEvent, 1)
	cancelEvent := mustCreateCancelationEvent(t, 1, 1, positions, 500)
	cancelEvent.ID = 2
	cancelEvent.Version = 2

	state, err = ApplyEvent(state, cancelEvent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if state.SaldoCents != 500 {
		t.Fatalf("expected SaldoCents 500, got %d", state.SaldoCents)
	}
	if len(state.UnbezahltePositionen) != 1 {
		t.Fatalf("expected 1 unbezahlte position, got %d", len(state.UnbezahltePositionen))
	}
	if state.UnbezahltePositionen[0].Menge != 1 {
		t.Fatalf("expected Menge 1, got %d", state.UnbezahltePositionen[0].Menge)
	}
	// Stornierung also reduces ausstehende
	if len(state.AusstehendePositionen) != 1 {
		t.Fatalf("expected 1 ungelieferte position, got %d", len(state.AusstehendePositionen))
	}
	if state.AusstehendePositionen[0].Menge != 1 {
		t.Fatalf("expected Menge 1, got %d", state.AusstehendePositionen[0].Menge)
	}
}

func TestApplyEvent_AusgabeReducesOnlyAusstehend(t *testing.T) {
	products := []Position{
		testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 2),
	}
	orderEvent := mustCreateOrderEvent(t, 1, 1, products)
	orderEvent.ID = 1
	orderEvent.Version = 1

	state, err := ApplyEvent(TischState{}, orderEvent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	positions := positionsFromOrder(t, orderEvent, 2)
	deliveryEvent := mustCreateDeliveryEvent(t, 1, 1, positions)
	deliveryEvent.ID = 2
	deliveryEvent.Version = 2

	state, err = ApplyEvent(state, deliveryEvent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Saldo unchanged
	if state.SaldoCents != 1000 {
		t.Fatalf("expected SaldoCents 1000, got %d", state.SaldoCents)
	}
	// Unbezahlt unchanged
	if len(state.UnbezahltePositionen) != 1 {
		t.Fatalf("expected 1 unbezahlte position, got %d", len(state.UnbezahltePositionen))
	}
	if state.UnbezahltePositionen[0].Menge != 2 {
		t.Fatalf("expected Menge 2, got %d", state.UnbezahltePositionen[0].Menge)
	}
	// Ungeliefert reduced to 0
	if len(state.AusstehendePositionen) != 0 {
		t.Fatalf("expected 0 ungelieferte positionen, got %d", len(state.AusstehendePositionen))
	}
}

func TestApplyEvent_MultipleEventsSequentially(t *testing.T) {
	products := []Position{
		testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 3),
		testPosition(2, "Wurst", "Bratwurst", "essen", 400, 2),
	}
	orderEvent := mustCreateOrderEvent(t, 1, 1, products)
	orderEvent.ID = 1
	orderEvent.Version = 1

	// Apply order: saldo = 500*3 + 400*2 = 2300
	state, err := ApplyEvent(TischState{}, orderEvent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.SaldoCents != 2300 {
		t.Fatalf("expected SaldoCents 2300, got %d", state.SaldoCents)
	}

	// Deliver all beer
	bestellung, err := buildBestellungFromEvent(orderEvent)
	if err != nil {
		t.Fatalf("failed to build bestellung: %v", err)
	}
	beerPos := []Position{bestellung.Positionen[0]}
	beerPos[0].Menge = 3
	deliveryEvent := mustCreateDeliveryEvent(t, 1, 1, beerPos)
	deliveryEvent.ID = 2
	deliveryEvent.Version = 2

	state, err = ApplyEvent(state, deliveryEvent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// Saldo unchanged
	if state.SaldoCents != 2300 {
		t.Fatalf("expected SaldoCents 2300, got %d", state.SaldoCents)
	}
	// 2 ausstehende (only wurst remaining)
	if len(state.AusstehendePositionen) != 1 {
		t.Fatalf("expected 1 ungelieferte position, got %d", len(state.AusstehendePositionen))
	}

	// Pay for 1 beer (500)
	beerPayPos := []Position{bestellung.Positionen[0]}
	beerPayPos[0].Menge = 1
	paymentEvent := mustCreatePaymentEvent(t, 1, 1, beerPayPos, 500)
	paymentEvent.ID = 3
	paymentEvent.Version = 3

	state, err = ApplyEvent(state, paymentEvent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.SaldoCents != 1800 {
		t.Fatalf("expected SaldoCents 1800, got %d", state.SaldoCents)
	}
	if state.GesamtZahlungenCents != 500 {
		t.Fatalf("expected GesamtZahlungenCents 500, got %d", state.GesamtZahlungenCents)
	}

	// Cancel 1 wurst (400)
	wurstPos := []Position{bestellung.Positionen[1]}
	wurstPos[0].Menge = 1
	cancelEvent := mustCreateCancelationEvent(t, 1, 1, wurstPos, 400)
	cancelEvent.ID = 4
	cancelEvent.Version = 4

	state, err = ApplyEvent(state, cancelEvent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.SaldoCents != 1400 {
		t.Fatalf("expected SaldoCents 1400, got %d", state.SaldoCents)
	}
	if state.LastEventID != 4 {
		t.Fatalf("expected LastEventID 4, got %d", state.LastEventID)
	}
	if state.LastEventVersion != 4 {
		t.Fatalf("expected LastEventVersion 4, got %d", state.LastEventVersion)
	}
}

func TestApplyEvent_UnknownEventType_ReturnsError(t *testing.T) {
	evt := e.Event{
		ID:      1,
		Type:    "unknown.event:v1",
		Data:    json.RawMessage(`{}`),
		Version: 1,
	}

	_, err := ApplyEvent(TischState{}, evt)
	if err == nil {
		t.Fatal("expected error for unknown event type, got nil")
	}
}

func TestApplyEvent_StornierungAfterPayment_NegativeSaldo(t *testing.T) {
	products := []Position{
		testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 2),
	}
	orderEvent := mustCreateOrderEvent(t, 1, 1, products)
	orderEvent.ID = 1
	orderEvent.Version = 1

	state, err := ApplyEvent(TischState{}, orderEvent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Pay for all 2 beers
	payPositions := positionsFromOrder(t, orderEvent, 2)
	payEvent := mustCreatePaymentEvent(t, 1, 1, payPositions, 1000)
	payEvent.ID = 2
	payEvent.Version = 2

	state, err = ApplyEvent(state, payEvent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.SaldoCents != 0 {
		t.Fatalf("expected SaldoCents 0, got %d", state.SaldoCents)
	}

	// Cancel 1 beer after payment — should result in negative saldo
	cancelPositions := positionsFromOrder(t, orderEvent, 1)
	cancelEvent := mustCreateCancelationEvent(t, 1, 1, cancelPositions, 500)
	cancelEvent.ID = 3
	cancelEvent.Version = 3

	state, err = ApplyEvent(state, cancelEvent)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if state.SaldoCents != -500 {
		t.Fatalf("expected SaldoCents -500, got %d", state.SaldoCents)
	}
	// Unbezahlt was already empty (paid), stays empty
	if len(state.UnbezahltePositionen) != 0 {
		t.Fatalf("expected 0 unbezahlte positionen, got %d", len(state.UnbezahltePositionen))
	}
}
