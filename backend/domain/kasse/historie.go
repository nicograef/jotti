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
	HistorieEintragAusgabe     HistorieEintragArt = "ausgabe"
)

type HistorieEintrag struct {
	Art         HistorieEintragArt
	Bestellung  *Bestellung
	Zahlung     *Zahlung
	Stornierung *Stornierung
	Umbuchung   *Umbuchung
	Ausgabe     *Ausgabe

	// StornierbarePositionen and UmbuchbarePositionen are populated for the entries
	// that introduce positions onto the table — a Bestellung or the incoming side
	// (Zugang) of a Umbuchung. They carry, per still-actionable position, the
	// quantity that remains: stornierbar = ordered − cancelled − moved away,
	// umbuchbar = stornierbar − paid. Computed here so the backend stays the single
	// source of truth for this filtering (no client-side replay of the history).
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
			// Die geldneutrale Korrektur erscheint in Historie und UI ebenfalls als
			// „Stornierung", trägt aber BarRueckgabe = false und wird darüber sichtbar
			// von der kassenwirksamen Warenrücknahme unterschieden.
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

		case string(EventTypeAusgabeBestaetigtV1):
			ausgabe, err := buildAusgabeFromEvent(event)
			if err != nil {
				return nil, err
			}
			history = append(history, HistorieEintrag{Art: HistorieEintragAusgabe, Ausgabe: &ausgabe})

		default:
			return nil, fmt.Errorf("unknown event type: %s", event.Type)
		}
	}

	enrichBestellungenMitRestmengen(history)

	// reverse the order of the array so that the most recent event is first
	slices.Reverse(history)

	return history, nil
}

// enrichBestellungenMitRestmengen annotates every position-introducing entry (a
// Bestellung or the incoming side of a Umbuchung) with the positions that are still
// stornierbar (ordered − cancelled − moved away) and umbuchbar (stornierbar − paid).
// Position IDs are unique per introducing entry, so the totals cancelled/paid/moved
// for a position only ever apply to its own source.
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
			// Nur der Abgang entfernt Positionen vom Tisch; der Zugang ist selbst eine
			// Positionsquelle und wird unten angereichert.
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

// positionsquelle liefert die Positionen eines Eintrags, der Positionen auf den
// Tisch bringt (eine Bestellung oder der Zugang einer Umbuchung), und ob der
// Eintrag eine solche Quelle ist.
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
