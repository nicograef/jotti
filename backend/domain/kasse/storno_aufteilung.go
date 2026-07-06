package kasse

import (
	"encoding/json"

	e "github.com/nicograef/jotti/backend/domain/event"
)

// StornoWarenruecknahme ist der kassenwirksame Teil eines Stornos für genau eine
// begleichende Zahlung: die zurückzunehmenden (bezahlten) Positionen und ihr
// Gesamtbetrag. Je Eintrag entsteht ein stornierung-erteilt-Event mit genau dieser
// ZahlungID.
type StornoWarenruecknahme struct {
	ZahlungID   string
	Positionen  []Position
	GesamtCents int
}

// StornoAufteilung ist das Ergebnis des Storno-Routings einer „Stornieren"-Aktion:
// die geldneutrale Korrektur noch unbezahlter Positionen (ein bestellung-korrigiert,
// evtl. leer) und je betroffener Zahlung eine kassenwirksame Warenrücknahme
// (ein stornierung-erteilt, FIFO nach Zahlung — älteste zuerst).
type StornoAufteilung struct {
	Korrektur        []Position
	KorrekturCents   int
	Warenruecknahmen []StornoWarenruecknahme
}

// zahlungRest hält je Zahlung (in Reihenfolge ihres Auftretens, FIFO) die noch
// zurücknehmbaren bezahlten Mengen je PositionID sowie die der Zahlung zugeordneten
// Storno-Positionen während der Aufteilung.
type zahlungRest struct {
	id       string
	rest     map[string]int
	genommen []Position
	cents    int
}

// ComputeStornoAufteilung spielt die Events einer Tisch-Session nach und teilt eine
// Storno-Anforderung nach Bezahlstatus auf: unbezahlte Mengen werden geldneutral
// korrigiert, bezahlte Mengen werden ihren begleichenden Zahlungen FIFO (älteste
// Zahlung zuerst) zugeordnet und je Zahlung als Warenrücknahme zurückgenommen. Pro
// Position wird zuerst die unbezahlte Menge korrigiert, der Rest aus den Zahlungen
// genommen. Der zweite Rückgabewert ist false, wenn eine angeforderte Menge die noch
// stornierbare (bestellte, nicht bereits zurückgenommene) Menge übersteigt oder eine
// PositionID mehrfach referenziert wird (kein legitimer Client sendet Duplikate).
func ComputeStornoAufteilung(events []e.Event, refs []PositionRef) (StornoAufteilung, bool) {
	details := map[string]Position{}
	unbezahlt := map[string]int{}
	var zahlungen []*zahlungRest

	merke := func(positionen []PositionEventData) {
		for _, p := range positionen {
			details[p.PositionID] = PositionFromEventData(p)
		}
	}

	for _, evt := range events {
		switch evt.Type {
		case string(EventTypeBestellungAufgenommenV1):
			var data BestellungAufgenommenV1Data
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return StornoAufteilung{}, false
			}
			merke(data.Positionen)
			for _, p := range data.Positionen {
				unbezahlt[p.PositionID] += p.Menge
			}

		case string(EventTypeBestellungUmgebuchtV1):
			var data BestellungUmgebuchtV1Data
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return StornoAufteilung{}, false
			}
			tischID, err := ParseTischIDFromSubject(evt.Subject)
			if err != nil {
				return StornoAufteilung{}, false
			}
			if tischID == data.QuellTischID {
				for _, p := range data.Positionen {
					unbezahlt[p.PositionID] -= p.Menge
				}
			} else {
				merke(data.Positionen)
				for _, p := range data.Positionen {
					unbezahlt[p.PositionID] += p.Menge
				}
			}

		case string(EventTypeZahlungKassiertV1):
			var data ZahlungKassiertV1Data
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return StornoAufteilung{}, false
			}
			bucket := &zahlungRest{id: data.ZahlungID, rest: map[string]int{}}
			for _, p := range data.Positionen {
				unbezahlt[p.PositionID] -= p.Menge
				bucket.rest[p.PositionID] += p.Menge
			}
			zahlungen = append(zahlungen, bucket)

		case string(EventTypeBestellungKorrigiertV1):
			var data BestellungKorrigiertV1Data
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return StornoAufteilung{}, false
			}
			for _, p := range data.Positionen {
				unbezahlt[p.PositionID] -= p.Menge
			}

		case string(EventTypeStornierungErteiltV1):
			var data StornierungErteiltV1Data
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return StornoAufteilung{}, false
			}
			for _, z := range zahlungen {
				if z.id == data.ZahlungID {
					for _, p := range data.Positionen {
						z.rest[p.PositionID] -= p.Menge
					}
					break
				}
			}
		}
	}

	var aufteilung StornoAufteilung
	seen := make(map[string]bool, len(refs))
	for _, ref := range refs {
		if seen[ref.PositionID] {
			return StornoAufteilung{}, false
		}
		seen[ref.PositionID] = true

		det, ok := details[ref.PositionID]
		if !ok {
			return StornoAufteilung{}, false
		}
		offen := ref.Menge

		// 1. Unbezahlte Menge geldneutral korrigieren.
		if frei := unbezahlt[ref.PositionID]; frei > 0 {
			take := min(offen, frei)
			unbezahlt[ref.PositionID] -= take
			offen -= take
			aufteilung.Korrektur = append(aufteilung.Korrektur, mitMenge(det, take))
			aufteilung.KorrekturCents += det.EinzelpreisCents * take
		}

		// 2. Bezahlte Menge FIFO je Zahlung als Warenrücknahme zurücknehmen.
		for _, z := range zahlungen {
			if offen == 0 {
				break
			}
			frei := z.rest[ref.PositionID]
			if frei <= 0 {
				continue
			}
			take := min(offen, frei)
			z.rest[ref.PositionID] -= take
			offen -= take
			z.genommen = append(z.genommen, mitMenge(det, take))
			z.cents += det.EinzelpreisCents * take
		}

		if offen > 0 {
			return StornoAufteilung{}, false
		}
	}

	for _, z := range zahlungen {
		if len(z.genommen) == 0 {
			continue
		}
		aufteilung.Warenruecknahmen = append(aufteilung.Warenruecknahmen, StornoWarenruecknahme{
			ZahlungID:   z.id,
			Positionen:  z.genommen,
			GesamtCents: z.cents,
		})
	}

	return aufteilung, true
}

// mitMenge liefert eine Kopie der Positions-Vorlage mit der angegebenen Menge.
func mitMenge(vorlage Position, menge int) Position {
	vorlage.Menge = menge
	return vorlage
}
