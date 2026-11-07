package kasse

import (
	"maps"
	"strconv"

	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

// AbgeleiteteID derives the ID of the index-th event a client key books. A Tisch-Storno writes a
// Korrektur and one Warenrücknahme per Zahlung; each needs its own BON_ID, and a retry with the same
// key derives the same IDs, which the BON_ID index then rejects.
func AbgeleiteteID(schluessel uuid.UUID, index int) string {
	return uuid.NewSHA1(schluessel, []byte(strconv.Itoa(index))).String()
}

// GebuchteZahlung finds the Zahlung with zahlungID in a Tisch stream.
func GebuchteZahlung(events []e.Event, zahlungID string) (e.Event, bool, error) {
	for _, evt := range events {
		if t := EventType(evt.Type); t != EventTypeZahlungKassiertV1 && t != EventTypeZahlungKassiertV2 {
			continue
		}
		data, err := e.Decode[ZahlungKassiertV1Data](evt)
		if err != nil {
			return e.Event{}, false, err
		}
		if data.ZahlungID == zahlungID {
			return evt, true, nil
		}
	}
	return e.Event{}, false, nil
}

// GebuchteTischStornierung collects the events a Tisch-Storno with schluessel booked in a Tisch
// stream, in derivation order; none means the key is not booked there.
func GebuchteTischStornierung(events []e.Event, schluessel uuid.UUID) ([]e.Event, error) {
	nachID := map[string]e.Event{}
	for _, evt := range events {
		t := EventType(evt.Type)
		if t != EventTypeStornierungErteiltV1 && t != EventTypeBestellungKorrigiertV1 {
			continue
		}
		var id string
		if t == EventTypeStornierungErteiltV1 {
			data, err := e.Decode[StornierungErteiltV1Data](evt)
			if err != nil {
				return nil, err
			}
			id = data.StornierungID
		} else {
			data, err := e.Decode[BestellungKorrigiertV1Data](evt)
			if err != nil {
				return nil, err
			}
			id = data.KorrekturID
		}
		nachID[id] = evt
	}

	var gebucht []e.Event
	for i := 0; ; i++ {
		evt, ok := nachID[AbgeleiteteID(schluessel, i)]
		if !ok {
			return gebucht, nil
		}
		gebucht = append(gebucht, evt)
	}
}

// GleicheTischStornierung reports whether the booked Korrektur and Warenrücknahmen together cancel
// exactly the referenced quantities.
func GleicheTischStornierung(gebucht []e.Event, refs []PositionRef) (bool, error) {
	mengen := map[string]int{}
	for _, evt := range gebucht {
		var positionen []PositionEventData
		if EventType(evt.Type) == EventTypeStornierungErteiltV1 {
			data, err := e.Decode[StornierungErteiltV1Data](evt)
			if err != nil {
				return false, err
			}
			positionen = data.Positionen
		} else {
			data, err := e.Decode[BestellungKorrigiertV1Data](evt)
			if err != nil {
				return false, err
			}
			positionen = data.Positionen
		}
		for id, menge := range mengenJePosition(positionen) {
			mengen[id] += menge
		}
	}
	return maps.Equal(mengen, mengenJeRef(refs)), nil
}

// GebuchteUmbuchung finds the Abgang of the Umbuchung with umbuchungID in the source table's stream.
func GebuchteUmbuchung(quellEvents []e.Event, umbuchungID string) (e.Event, bool, error) {
	for _, evt := range quellEvents {
		if EventType(evt.Type) != EventTypeBestellungUmgebuchtV1 {
			continue
		}
		data, err := e.Decode[BestellungUmgebuchtV1Data](evt)
		if err != nil {
			return e.Event{}, false, err
		}
		if data.UmbuchungID == umbuchungID {
			return evt, true, nil
		}
	}
	return e.Event{}, false, nil
}

// GleicheUmbuchung reports whether the booked Abgang moves exactly the referenced quantities between
// the same two tables.
func GleicheUmbuchung(abgang e.Event, quellTischID, zielTischID int, refs []PositionRef) (bool, error) {
	data, err := e.Decode[BestellungUmgebuchtV1Data](abgang)
	if err != nil {
		return false, err
	}
	return data.QuellTischID == quellTischID && data.ZielTischID == zielTischID &&
		maps.Equal(mengenJePosition(data.Positionen), mengenJeRef(refs)), nil
}

// GebuchteDirektverkaufStornierung finds the Stornierung with stornierungID in a Verkauf stream.
func GebuchteDirektverkaufStornierung(events []e.Event, stornierungID string) (e.Event, bool, error) {
	for _, evt := range events {
		if EventType(evt.Type) != EventTypeDirektverkaufStorniertV1 {
			continue
		}
		data, err := e.Decode[DirektverkaufStorniertV1Data](evt)
		if err != nil {
			return e.Event{}, false, err
		}
		if data.StornierungID == stornierungID {
			return evt, true, nil
		}
	}
	return e.Event{}, false, nil
}

// GleicheDirektverkaufStornierung reports whether the booked Stornierung cancels exactly the
// referenced quantities.
func GleicheDirektverkaufStornierung(stornierung e.Event, refs []PositionRef) (bool, error) {
	data, err := e.Decode[DirektverkaufStorniertV1Data](stornierung)
	if err != nil {
		return false, err
	}
	return maps.Equal(mengenJePosition(data.Positionen), mengenJeRef(refs)), nil
}
