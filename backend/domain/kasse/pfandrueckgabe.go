package kasse

import (
	"maps"

	e "github.com/nicograef/jotti/backend/domain/event"
)

// PfandrueckgabeHistorieEintrag is one Pfandrückgabe replayed from its stream; TischID 0 is the
// Direktverkauf.
type PfandrueckgabeHistorieEintrag struct {
	VorgangHistorieEintrag
	RueckgabeID string
	TischID     int
}

func BuildPfandrueckgabeHistorieEintrag(events []e.Event) (PfandrueckgabeHistorieEintrag, error) {
	eintrag, data, err := replayVorgang(events, EventTypePfandZurueckgegebenV1, EventTypePfandrueckgabeStorniertV1,
		func(d PfandZurueckgegebenV1Data) vorgangKern {
			return vorgangKern{Positionen: d.Positionen, GesamtbetragCents: d.GesamtbetragCents, Kommentar: d.Kommentar}
		},
		func(d PfandrueckgabeStorniertV1Data) stornoKern {
			return stornoKern{StornierungID: d.StornierungID, Positionen: d.Positionen, GesamtStornierungCents: d.GesamtStornierungCents}
		},
	)
	if err != nil {
		return PfandrueckgabeHistorieEintrag{}, err
	}
	return PfandrueckgabeHistorieEintrag{VorgangHistorieEintrag: eintrag, RueckgabeID: data.RueckgabeID, TischID: data.TischID}, nil
}

// Verrechne stamps a settled Rückgabe with the time of the Zahlung or Verkauf it is written with.
// The shared timestamp is the pairing VerrechnungsPartner finds on a retry.
func Verrechne(rueckgabe e.Event, partner e.Event) e.Event {
	rueckgabe.Time = partner.Time
	return rueckgabe
}

// VerrechnungsPartner returns the event of typ among events that was written with the settled
// Rückgabe: same timestamp, same user.
func VerrechnungsPartner(rueckgabe e.Event, events []e.Event, typ EventType) (e.Event, bool) {
	for _, evt := range events {
		if evt.Type == string(typ) && evt.UserID == rueckgabe.UserID && evt.Time.Equal(rueckgabe.Time) {
			return evt, true
		}
	}
	return e.Event{}, false
}

// GleicheRueckgabe reports whether the stored Rückgabe books tischID and mengen, the quantity per
// variant.
func GleicheRueckgabe(rueckgabe e.Event, tischID int, mengen map[int]int) (bool, error) {
	data := PfandZurueckgegebenV1Data{}
	if err := e.ParseData(rueckgabe, &data, pfandZurueckgegebenV1DataSchema); err != nil {
		return false, err
	}
	return data.TischID == tischID && maps.Equal(mengenJeVariante(data.Positionen), mengen), nil
}

// GleicherVerkauf reports whether the stored Direktverkauf books mengen, the quantity per variant.
func GleicherVerkauf(verkauf e.Event, mengen map[int]int) (bool, error) {
	data := DirektverkaufGetaetigtV1Data{}
	if err := e.ParseData(verkauf, &data, direktverkaufGetaetigtV1DataSchema); err != nil {
		return false, err
	}
	return maps.Equal(mengenJeVariante(data.Positionen), mengen), nil
}

// GleicheZahlung reports whether the stored Zahlung pays exactly the referenced quantities.
func GleicheZahlung(zahlung e.Event, refs []PositionRef) (bool, error) {
	data := ZahlungKassiertV1Data{}
	if err := e.ParseData(zahlung, &data, zahlungKassiertV1DataSchema); err != nil {
		return false, err
	}
	gezahlt := make(map[string]int, len(data.Positionen))
	for _, p := range data.Positionen {
		gezahlt[p.PositionID] += p.Menge
	}
	angefragt := make(map[string]int, len(refs))
	for _, ref := range refs {
		angefragt[ref.PositionID] += ref.Menge
	}
	return maps.Equal(gezahlt, angefragt), nil
}

func mengenJeVariante(positionen []PositionEventData) map[int]int {
	mengen := make(map[int]int, len(positionen))
	for _, p := range positionen {
		mengen[p.VarianteID] += p.Menge
	}
	return mengen
}
