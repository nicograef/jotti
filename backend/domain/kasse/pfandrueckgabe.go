package kasse

import (
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
