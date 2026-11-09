package kasse

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

func TestGetHistorieFromEvents_Empty(t *testing.T) {
	history, err := GetHistorieFromEvents([]e.Event{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(history) != 0 {
		t.Errorf("expected empty history, got %d items", len(history))
	}
}

func TestGetHistorieFromEvents_ReturnsAllEventTypes(t *testing.T) {
	products := []Position{
		testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 1),
	}
	orderEvent := mustCreateBestellungEvent(t, testSubject, 1, products)
	positions := positionenAusBestellung(t, orderEvent, 1)

	zahlungV1, err := e.New(1, "TestUser", string(EventTypeZahlungKassiertV1), testSubject, ZahlungKassiertV1Data{
		ZahlungID:          testZahlungID,
		Positionen:         toPositionenEventData(positions),
		GesamtZahlungCents: 500,
	})
	if err != nil {
		t.Fatalf("zahlung v1 event: %v", err)
	}
	umbuchung, _, err := NewBestellungUmgebuchtEvents(1, uuid.NewString(), 42, 43, 1, "TestUser", positions, 500, "Umbuchung zu Tisch 43", "Umbuchung von Tisch 42", "")
	if err != nil {
		t.Fatalf("umbuchung events: %v", err)
	}

	cases := []struct {
		event e.Event
		art   HistorieEintragArt
	}{
		{orderEvent, HistorieEintragBestellung},
		{zahlungV1, HistorieEintragZahlung},
		{mustCreateZahlungEvent(t, testSubject, 1, positions, 500), HistorieEintragZahlung},
		{mustCreateStornierungEvent(t, testSubject, 1, testZahlungID, positions, 500), HistorieEintragStornierung},
		{mustCreateKorrekturEvent(t, testSubject, 1, positions, 500), HistorieEintragStornierung},
		{umbuchung, HistorieEintragUmbuchung},
	}
	for _, c := range cases {
		t.Run(c.event.Type, func(t *testing.T) {
			history, err := GetHistorieFromEvents([]e.Event{c.event})
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if len(history) != 1 {
				t.Fatalf("expected 1 history item, got %d", len(history))
			}
			if history[0].Art != c.art {
				t.Errorf("Art: got %q, want %q", history[0].Art, c.art)
			}
		})
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

func TestGetHistorieFromEvents_Positionskommentar(t *testing.T) {
	products := []Position{
		testPosition(1, "Zwiebelkuchen", "Stück", "essen", 400, 1),
		testPosition(2, "Beer", "Pils 0.5l", "getraenk", 500, 1),
	}
	products[0].Kommentar = "ohne Kümmel"
	orderEvent := mustCreateBestellungEvent(t, testSubject, 1, products)

	history, err := GetHistorieFromEvents([]e.Event{orderEvent})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(history) != 1 || history[0].Bestellung == nil {
		t.Fatalf("expected one Bestellung entry, got %+v", history)
	}
	positionen := history[0].Bestellung.Positionen
	if positionen[0].Kommentar != "ohne Kümmel" {
		t.Errorf("Positionen[0].Kommentar = %q, want %q", positionen[0].Kommentar, "ohne Kümmel")
	}
	if positionen[1].Kommentar != "" {
		t.Errorf("Positionen[1].Kommentar = %q, want empty", positionen[1].Kommentar)
	}
}

func TestGetHistorieFromEvents_EnrichesBestellungMitRestmengen(t *testing.T) {
	tests := []struct {
		name            string
		bestellt        int
		folge           func(t *testing.T, eine []Position) []e.Event
		wantStornierbar int
		wantUmbuchbar   int
	}{
		{
			// Die Warenrücknahme hebt die Zahlung auf; die zwei unbezahlten bleiben umbuchbar.
			name:     "bestellt 3, bezahlt 1, Warenrücknahme 1",
			bestellt: 3,
			folge: func(t *testing.T, eine []Position) []e.Event {
				return []e.Event{
					mustCreateZahlungEvent(t, testSubject, 1, eine, 500),
					mustCreateStornierungEvent(t, testSubject, 1, testZahlungID, eine, 500),
				}
			},
			wantStornierbar: 2,
			wantUmbuchbar:   2,
		},
		{
			name:     "bestellt 1, bezahlt 1, Warenrücknahme 1",
			bestellt: 1,
			folge: func(t *testing.T, eine []Position) []e.Event {
				return []e.Event{
					mustCreateZahlungEvent(t, testSubject, 1, eine, 500),
					mustCreateStornierungEvent(t, testSubject, 1, testZahlungID, eine, 500),
				}
			},
			wantStornierbar: 0,
			wantUmbuchbar:   0,
		},
		{
			name:     "bestellt 3, Korrektur 1, bezahlt 1",
			bestellt: 3,
			folge: func(t *testing.T, eine []Position) []e.Event {
				return []e.Event{
					mustCreateKorrekturEvent(t, testSubject, 1, eine, 500),
					mustCreateZahlungEvent(t, testSubject, 1, eine, 500),
				}
			},
			wantStornierbar: 2,
			wantUmbuchbar:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			products := []Position{
				testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, tt.bestellt),
			}
			orderEvent := mustCreateBestellungEvent(t, testSubject, 1, products)
			eine := positionenAusBestellung(t, orderEvent, 1)

			history, err := GetHistorieFromEvents(append([]e.Event{orderEvent}, tt.folge(t, eine)...))
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			// History is reversed (most recent first); the Bestellung is the last entry.
			bestellung := history[len(history)-1]
			if bestellung.Art != HistorieEintragBestellung {
				t.Fatalf("expected last entry to be Bestellung, got %q", bestellung.Art)
			}
			if got := summeMenge(bestellung.StornierbarePositionen); got != tt.wantStornierbar {
				t.Errorf("stornierbar = %d, want %d (%+v)", got, tt.wantStornierbar, bestellung.StornierbarePositionen)
			}
			if got := summeMenge(bestellung.UmbuchbarePositionen); got != tt.wantUmbuchbar {
				t.Errorf("umbuchbar = %d, want %d (%+v)", got, tt.wantUmbuchbar, bestellung.UmbuchbarePositionen)
			}
		})
	}
}

func summeMenge(positionen []Position) int {
	summe := 0
	for _, pos := range positionen {
		summe += pos.Menge
	}
	return summe
}

func TestBuildStornierung_BarRueckgabeLeitetStornoArtAusEventTypAb(t *testing.T) {
	products := []Position{
		testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 1),
	}
	orderEvent := mustCreateBestellungEvent(t, testSubject, 1, products)
	positions := positionenAusBestellung(t, orderEvent, 1)

	warenruecknahme := mustCreateStornierungEvent(t, testSubject, 1, testZahlungID, positions, 500)
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
	orderEvent := mustCreateBestellungEvent(t, testSubject, 1, products)
	positions := positionenAusBestellung(t, orderEvent, 1)

	events := []e.Event{
		orderEvent,
		mustCreateZahlungEvent(t, testSubject, 1, positions, 500),
	}

	history, err := GetHistorieFromEvents(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if history[0].Art != HistorieEintragZahlung || history[0].Zahlung == nil {
		t.Fatalf("expected first item to be Zahlung, got kind %q", history[0].Art)
	}
	if history[1].Art != HistorieEintragBestellung || history[1].Bestellung == nil {
		t.Errorf("expected second item to be Bestellung, got kind %q", history[1].Art)
	}
}

func TestGetHistorieFromEvents_UmbuchungAbgangReduziertRestmengen(t *testing.T) {
	const zNr, quellTischID, zielTischID = 1, 42, 7
	quellSubject := TischSessionSubject(zNr, quellTischID)

	products := []Position{
		testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 3),
	}
	orderEvent := mustCreateBestellungEvent(t, quellSubject, 1, products)
	eine := positionenAusBestellung(t, orderEvent, 1)

	quellEvent, _, err := NewBestellungUmgebuchtEvents(zNr, uuid.NewString(), quellTischID, zielTischID, 1, "TestUser", eine, 500, "Umbuchung auf Tisch Ziel", "Umbuchung von Tisch Quelle", "")
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
		t.Error("expected the source entry to be an Abgang, not a Zugang")
	}

	bestellung := history[len(history)-1]
	// Ordered 3, moved away 1 => 2 still stornierbar and umbuchbar.
	if len(bestellung.StornierbarePositionen) != 1 || bestellung.StornierbarePositionen[0].Menge != 2 {
		t.Errorf("expected stornierbar menge 2, got %+v", bestellung.StornierbarePositionen)
	}
	if len(bestellung.UmbuchbarePositionen) != 1 || bestellung.UmbuchbarePositionen[0].Menge != 2 {
		t.Errorf("expected umbuchbar menge 2, got %+v", bestellung.UmbuchbarePositionen)
	}
}

func TestGetHistorieFromEvents_UmbuchungZugangIstStornierbar(t *testing.T) {
	const zNr, quellTischID, zielTischID = 1, 42, 7
	zielSubject := TischSessionSubject(zNr, zielTischID)

	positionen := []Position{testPosition(1, "Beer", "Pils 0.5l", "getraenk", 500, 2)}
	// PositionIDs werden im Quell-Event vergeben; für den Zugang reicht eine ID.
	positionen[0].PositionID = "11111111-1111-4111-8111-111111111111"

	_, zielEvent, err := NewBestellungUmgebuchtEvents(zNr, uuid.NewString(), quellTischID, zielTischID, 1, "TestUser", positionen, 1000, "Umbuchung auf Tisch Ziel", "Umbuchung von Tisch Quelle", "")
	if err != nil {
		t.Fatalf("failed to create umbuchung events: %v", err)
	}
	if zielEvent.Subject != zielSubject {
		t.Errorf("expected ziel subject %q, got %q", zielSubject, zielEvent.Subject)
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
		t.Error("expected the target entry to be a Zugang")
	}
	// Der Zugang bringt Positionen auf den Tisch: voll stornierbar/umbuchbar.
	if len(zugang.StornierbarePositionen) != 1 || zugang.StornierbarePositionen[0].Menge != 2 {
		t.Errorf("expected stornierbar menge 2, got %+v", zugang.StornierbarePositionen)
	}
	if len(zugang.UmbuchbarePositionen) != 1 || zugang.UmbuchbarePositionen[0].Menge != 2 {
		t.Errorf("expected umbuchbar menge 2, got %+v", zugang.UmbuchbarePositionen)
	}
}

// A stored position outside today's master-data rules (price > 99999, name < 3 characters) must
// stay readable: positionSchema has its own frozen bounds, decoupled from produkt.
func TestGetHistorieFromEvents_GespeichertePositionAusserhalbStammdatenregeln(t *testing.T) {
	cases := []struct {
		name        string
		produktName string
		preis       int
	}{
		{"Preis über Stammdaten-Obergrenze", "Fass Bier", 100000},
		{"Name kürzer als Stammdaten-Minimum", "Ei", 50},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			data := `{"bestellungId":"11111111-1111-4111-8111-111111111111","positionen":[{"positionId":"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa","varianteId":7,"produktName":"` +
				c.produktName + `","varianteName":"30l","kategorie":"getraenk","steuersatz":"regel","einzelpreisCents":` +
				strconv.Itoa(c.preis) + `,"menge":1}],"gesamtPreisCents":` + strconv.Itoa(c.preis) + `,"kommentar":""}`
			stored := e.Event{
				ID: 1, UserID: 1, UserName: "TestUser", Type: string(EventTypeBestellungAufgenommenV1),
				Time: time.Date(2026, 7, 7, 18, 0, 0, 0, time.UTC), Subject: testSubject, Version: 1,
				Data: json.RawMessage(data),
			}

			history, err := GetHistorieFromEvents([]e.Event{stored})
			if err != nil {
				t.Fatalf("stored event unreadable: %v", err)
			}
			if len(history) != 1 || history[0].Bestellung == nil {
				t.Fatalf("expected one Bestellung entry, got %+v", history)
			}
			pos := history[0].Bestellung.Positionen[0]
			if pos.EinzelpreisCents != c.preis || pos.ProduktName != c.produktName {
				t.Errorf("position: got %d/%q, want %d/%q", pos.EinzelpreisCents, pos.ProduktName, c.preis, c.produktName)
			}
		})
	}
}
