package kasse

import (
	"fmt"
	"slices"

	e "github.com/nicograef/jotti/backend/domain/event"
)

type HistorieEintragArt string

const (
	HistorieEintragBestellung  HistorieEintragArt = "bestellung"
	HistorieEintragZahlung     HistorieEintragArt = "zahlung"
	HistorieEintragStornierung HistorieEintragArt = "stornierung"
	HistorieEintragAusgabe     HistorieEintragArt = "ausgabe"
	HistorieEintragAuszahlung  HistorieEintragArt = "auszahlung"
)

type HistorieEintrag struct {
	Art         HistorieEintragArt
	Bestellung  *Bestellung
	Zahlung     *Zahlung
	Stornierung *Stornierung
	Ausgabe     *Ausgabe
	Auszahlung  *Auszahlung

	// StornierbarePositionen and UmbuchbarePositionen are populated only for
	// Bestellung entries. They carry, per still-actionable position, the quantity
	// that remains: stornierbar = ordered − cancelled, umbuchbar = ordered −
	// cancelled − paid. Computed here so the backend stays the single source of
	// truth for this filtering (no client-side replay of the history).
	StornierbarePositionen []Position
	UmbuchbarePositionen   []Position
}

func GetHistorieFromEvents(events []e.Event) ([]HistorieEintrag, error) {
	history := []HistorieEintrag{}

	for _, event := range events {
		switch event.Type {
		case string(EventTypeBestellungAufgenommenV1):
			bestellung, err := buildBestellungFromEvent(event)
			if err != nil {
				return []HistorieEintrag{}, err
			}
			history = append(history, HistorieEintrag{Art: HistorieEintragBestellung, Bestellung: &bestellung})
		case string(EventTypeZahlungKassiertV1):
			zahlung, err := buildZahlungFromEvent(event)
			if err != nil {
				return []HistorieEintrag{}, err
			}
			history = append(history, HistorieEintrag{Art: HistorieEintragZahlung, Zahlung: &zahlung})
		case string(EventTypeStornierungErteiltV1):
			stornierung, err := buildStornierungFromEvent(event)
			if err != nil {
				return []HistorieEintrag{}, err
			}
			history = append(history, HistorieEintrag{Art: HistorieEintragStornierung, Stornierung: &stornierung})

		case string(EventTypeAusgabeBestaetigtV1):
			ausgabe, err := buildAusgabeFromEvent(event)
			if err != nil {
				return []HistorieEintrag{}, err
			}
			history = append(history, HistorieEintrag{Art: HistorieEintragAusgabe, Ausgabe: &ausgabe})

		case string(EventTypeAuszahlungGeleistetV1):
			auszahlung, err := buildAuszahlungFromEvent(event)
			if err != nil {
				return []HistorieEintrag{}, err
			}
			history = append(history, HistorieEintrag{Art: HistorieEintragAuszahlung, Auszahlung: &auszahlung})

		default:
			return []HistorieEintrag{}, fmt.Errorf("unknown event type: %s", event.Type)
		}
	}

	enrichBestellungenMitRestmengen(history)

	// reverse the order of the array so that the most recent event is first
	slices.Reverse(history)

	return history, nil
}

// enrichBestellungenMitRestmengen annotates each Bestellung entry with the
// positions that are still stornierbar (ordered − cancelled) and umbuchbar
// (ordered − cancelled − paid). Position IDs are unique per Bestellung, so the
// totals cancelled/paid for a position only ever apply to its own order.
func enrichBestellungenMitRestmengen(history []HistorieEintrag) {
	storniert := map[string]int{}
	bezahlt := map[string]int{}
	for _, eintrag := range history {
		switch eintrag.Art {
		case HistorieEintragStornierung:
			if eintrag.Stornierung != nil {
				for _, pos := range eintrag.Stornierung.Positionen {
					storniert[pos.PositionID] += pos.Menge
				}
			}
		case HistorieEintragZahlung:
			if eintrag.Zahlung != nil {
				for _, pos := range eintrag.Zahlung.Positionen {
					bezahlt[pos.PositionID] += pos.Menge
				}
			}
		}
	}

	for i := range history {
		if history[i].Art != HistorieEintragBestellung || history[i].Bestellung == nil {
			continue
		}
		positionen := history[i].Bestellung.Positionen
		history[i].StornierbarePositionen = restmengen(positionen, func(pos Position) int {
			return storniert[pos.PositionID]
		})
		history[i].UmbuchbarePositionen = restmengen(positionen, func(pos Position) int {
			return storniert[pos.PositionID] + bezahlt[pos.PositionID]
		})
	}
}

// restmengen returns the positions whose remaining quantity (menge minus the
// amount reported by abzug) is still positive, with menge set to that remainder.
func restmengen(positionen []Position, abzug func(Position) int) []Position {
	rest := []Position{}
	for _, pos := range positionen {
		verbleibend := pos.Menge - abzug(pos)
		if verbleibend <= 0 {
			continue
		}
		reduziert := pos
		reduziert.Menge = verbleibend
		rest = append(rest, reduziert)
	}
	return rest
}
