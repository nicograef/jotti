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
	HistorieEintragUmbuchung   HistorieEintragArt = "umbuchung"
)

type HistorieEintrag struct {
	Art         HistorieEintragArt
	Bestellung  *Bestellung
	Zahlung     *Zahlung
	Stornierung *Stornierung
	Umbuchung   *Umbuchung

	// StornierbarePositionen and UmbuchbarePositionen are set only on entries that put
	// positions on the table (a Bestellung or the Zugang of a Umbuchung): per position the
	// quantity that remains, stornierbar = ordered − cancelled − moved away,
	// umbuchbar = stornierbar − paid.
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
				return nil, err
			}
			history = append(history, HistorieEintrag{Art: HistorieEintragBestellung, Bestellung: &bestellung})
		case string(EventTypeZahlungKassiertV1):
			zahlung, err := buildZahlungFromEvent(event)
			if err != nil {
				return nil, err
			}
			history = append(history, HistorieEintrag{Art: HistorieEintragZahlung, Zahlung: &zahlung})
		case string(EventTypeStornierungErteiltV1):
			stornierung, err := buildStornierungFromEvent(event)
			if err != nil {
				return nil, err
			}
			history = append(history, HistorieEintrag{Art: HistorieEintragStornierung, Stornierung: &stornierung})

		case string(EventTypeBestellungKorrigiertV1):
			korrektur, err := buildKorrekturFromEvent(event)
			if err != nil {
				return nil, err
			}
			history = append(history, HistorieEintrag{Art: HistorieEintragStornierung, Stornierung: &korrektur})

		case string(EventTypeBestellungUmgebuchtV1):
			umbuchung, err := buildUmbuchungFromEvent(event)
			if err != nil {
				return nil, err
			}
			history = append(history, HistorieEintrag{Art: HistorieEintragUmbuchung, Umbuchung: &umbuchung})

		default:
			return nil, fmt.Errorf("unknown event type: %s", event.Type)
		}
	}

	enrichBestellungenMitRestmengen(history)

	// Most recent first.
	slices.Reverse(history)

	return history, nil
}

// enrichBestellungenMitRestmengen fills StornierbarePositionen/UmbuchbarePositionen.
// Position IDs are unique per introducing entry, so the cancelled/paid/moved totals of a
// position only ever apply to its own source.
func enrichBestellungenMitRestmengen(history []HistorieEintrag) {
	storniert := map[string]int{}
	bezahlt := map[string]int{}
	umgebucht := map[string]int{}
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
		case HistorieEintragUmbuchung:
			// Nur der Abgang entfernt Positionen; der Zugang ist selbst eine Positionsquelle.
			if eintrag.Umbuchung != nil && !eintrag.Umbuchung.IstZugang() {
				for _, pos := range eintrag.Umbuchung.Positionen {
					umgebucht[pos.PositionID] += pos.Menge
				}
			}
		}
	}

	for i := range history {
		positionen, ok := positionsquelle(history[i])
		if !ok {
			continue
		}
		history[i].StornierbarePositionen = restmengen(positionen, func(pos Position) int {
			return storniert[pos.PositionID] + umgebucht[pos.PositionID]
		})
		history[i].UmbuchbarePositionen = restmengen(positionen, func(pos Position) int {
			return storniert[pos.PositionID] + umgebucht[pos.PositionID] + bezahlt[pos.PositionID]
		})
	}
}

func positionsquelle(eintrag HistorieEintrag) ([]Position, bool) {
	switch eintrag.Art {
	case HistorieEintragBestellung:
		if eintrag.Bestellung != nil {
			return eintrag.Bestellung.Positionen, true
		}
	case HistorieEintragUmbuchung:
		if eintrag.Umbuchung != nil && eintrag.Umbuchung.IstZugang() {
			return eintrag.Umbuchung.Positionen, true
		}
	}
	return nil, false
}

// restmengen returns the positions with a positive remainder, Menge set to that remainder.
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
