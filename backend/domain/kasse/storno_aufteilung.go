package kasse

import (
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

// ComputeStornoAufteilung replays a Tisch-Session and splits a Storno request by payment status:
// per position the unpaid quantity first (Korrektur), the rest FIFO from the settling payments.
// It returns false when a quantity exceeds what remains stornierbar or a PositionID repeats.
func ComputeStornoAufteilung(events []e.Event, refs []PositionRef) (StornoAufteilung, bool) {
	bestand, ok := stornoBestandAus(events)
	if !ok {
		return StornoAufteilung{}, false
	}
	return bestand.teileAuf(refs)
}

// stornoBestand is what a Tisch-Session still holds for a Storno: per PositionID the position
// details and the unpaid quantity, and per Zahlung (FIFO) the quantities it still settles.
type stornoBestand struct {
	details   map[string]Position
	unbezahlt map[string]int
	zahlungen []*zahlungRest
}

func (b *stornoBestand) merke(positionen []PositionEventData) {
	for _, p := range positionen {
		b.details[p.PositionID] = PositionFromEventData(p)
	}
}

func (b *stornoBestand) aendereUnbezahlt(positionen []PositionEventData, vorzeichen int) {
	for _, p := range positionen {
		b.unbezahlt[p.PositionID] += vorzeichen * p.Menge
	}
}

// stornoBestandAus replays the stream; false means an unreadable or foreign event.
func stornoBestandAus(events []e.Event) (*stornoBestand, bool) {
	b := &stornoBestand{details: map[string]Position{}, unbezahlt: map[string]int{}}
	for _, evt := range events {
		if !b.apply(evt) {
			return nil, false
		}
	}
	return b, true
}

func (b *stornoBestand) apply(evt e.Event) bool {
	switch EventType(evt.Type) {
	case EventTypeBestellungAufgenommenV1:
		data, err := e.Decode[BestellungAufgenommenV1Data](evt)
		if err != nil {
			return false
		}
		b.merke(data.Positionen)
		b.aendereUnbezahlt(data.Positionen, 1)

	case EventTypeBestellungUmgebuchtV1:
		data, err := e.Decode[BestellungUmgebuchtV1Data](evt)
		if err != nil {
			return false
		}
		tischID, err := ParseTischIDFromSubject(evt.Subject)
		if err != nil {
			return false
		}
		if tischID == data.QuellTischID {
			b.aendereUnbezahlt(data.Positionen, -1)
		} else {
			b.merke(data.Positionen)
			b.aendereUnbezahlt(data.Positionen, 1)
		}

	case EventTypeZahlungKassiertV1, EventTypeZahlungKassiertV2:
		data, err := e.Decode[ZahlungKassiertV1Data](evt)
		if err != nil {
			return false
		}
		bucket := &zahlungRest{id: data.ZahlungID, rest: map[string]int{}}
		for _, p := range data.Positionen {
			bucket.rest[p.PositionID] += p.Menge
		}
		b.aendereUnbezahlt(data.Positionen, -1)
		b.zahlungen = append(b.zahlungen, bucket)

	case EventTypeBestellungKorrigiertV1:
		data, err := e.Decode[BestellungKorrigiertV1Data](evt)
		if err != nil {
			return false
		}
		b.aendereUnbezahlt(data.Positionen, -1)

	case EventTypeStornierungErteiltV1:
		data, err := e.Decode[StornierungErteiltV1Data](evt)
		if err != nil {
			return false
		}
		for _, z := range b.zahlungen {
			if z.id == data.ZahlungID {
				for _, p := range data.Positionen {
					z.rest[p.PositionID] -= p.Menge
				}
				break
			}
		}

	case EventTypeKassensitzungEroeffnetV1, EventTypeGeldtransitGebuchtV1, EventTypeKassensturzDurchgefuehrtV1,
		EventTypeDifferenzSollIstGebuchtV1, EventTypeTagesabschlussErstelltV1,
		EventTypeDirektverkaufGetaetigtV1, EventTypeDirektverkaufGetaetigtV2, EventTypeDirektverkaufStorniertV1,
		EventTypePfandZurueckgegebenV1, EventTypePfandrueckgabeStorniertV1:
		return false

	default:
		// Unbekannter Event-Typ: verweigern statt raten — eine stille Fehlaufteilung wäre schlimmer.
		return false
	}
	return true
}

// teileAuf allocates refs: per position the unpaid quantity first, the rest FIFO across the
// payments. It consumes the Bestand and returns false on a repeated, unknown or excess position.
func (b *stornoBestand) teileAuf(refs []PositionRef) (StornoAufteilung, bool) {
	var aufteilung StornoAufteilung
	seen := make(map[string]bool, len(refs))
	for _, ref := range refs {
		if seen[ref.PositionID] {
			return StornoAufteilung{}, false
		}
		seen[ref.PositionID] = true

		det, ok := b.details[ref.PositionID]
		if !ok {
			return StornoAufteilung{}, false
		}
		offen := ref.Menge

		if frei := b.unbezahlt[ref.PositionID]; frei > 0 {
			take := min(offen, frei)
			b.unbezahlt[ref.PositionID] -= take
			offen -= take
			aufteilung.Korrektur = append(aufteilung.Korrektur, mitMenge(det, take))
			aufteilung.KorrekturCents += det.EinzelpreisCents * take
		}

		for _, z := range b.zahlungen {
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

	for _, z := range b.zahlungen {
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
