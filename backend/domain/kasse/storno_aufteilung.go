package kasse

import (
	"encoding/json"

	e "github.com/nicograef/jotti/backend/domain/event"
)

// StornoWarenruecknahme ist der kassenwirksame Teil eines Stornos für genau eine Zahlung:
// die zurückzunehmenden Positionen und ihr Gesamtbetrag. Je Eintrag entsteht ein
// stornierung-erteilt-Event mit dieser ZahlungID.
type StornoWarenruecknahme struct {
	ZahlungID   string
	Positionen  []Position
	GesamtCents int
}

// StornoAufteilung ist das Ergebnis des Storno-Routings: die geldneutrale Korrektur
// unbezahlter Positionen (ein bestellung-korrigiert, evtl. leer) und je betroffener
// Zahlung eine Warenrücknahme (ein stornierung-erteilt, FIFO — älteste Zahlung zuerst).
type StornoAufteilung struct {
	Korrektur        []Position
	KorrekturCents   int
	Warenruecknahmen []StornoWarenruecknahme
}

// zahlungRest hält je Zahlung (FIFO in der Reihenfolge ihres Auftretens) die noch
// zurücknehmbaren bezahlten Mengen je PositionID und die zugeordneten Storno-Positionen.
type zahlungRest struct {
	id       string
	rest     map[string]int
	genommen []Position
	cents    int
}

// ComputeStornoAufteilung spielt die Events einer Tisch-Session nach und teilt eine
// Storno-Anforderung nach Bezahlstatus auf: je Position zuerst die unbezahlte Menge
// (geldneutrale Korrektur), der Rest FIFO aus den begleichenden Zahlungen. Der zweite
// Rückgabewert ist false, wenn eine Menge die noch stornierbare übersteigt oder eine
// PositionID mehrfach referenziert wird.
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

		default:
			// Unbekannter Event-Typ: verweigern statt raten — eine stille Fehlaufteilung wäre schlimmer.
			return StornoAufteilung{}, false
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

		if frei := unbezahlt[ref.PositionID]; frei > 0 {
			take := min(offen, frei)
			unbezahlt[ref.PositionID] -= take
			offen -= take
			aufteilung.Korrektur = append(aufteilung.Korrektur, mitMenge(det, take))
			aufteilung.KorrekturCents += det.EinzelpreisCents * take
		}

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

func mitMenge(vorlage Position, menge int) Position {
	vorlage.Menge = menge
	return vorlage
}
