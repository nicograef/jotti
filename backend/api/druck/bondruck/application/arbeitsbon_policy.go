package application

import (
	"encoding/base64"

	"github.com/nicograef/jotti/backend/api/druck/bondruck/application/escpos"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
)

// positionenMitKommentarEventData spiegelt die benötigten Felder von
// bestellung-aufgenommen:v1 und direktverkauf-getaetigt in ihrer Event-Form, samt
// Positionskommentar und Pfand-Kennzeichen; Events ohne pfand-Schlüssel lesen sich als false.
// Keine Schema-Validierung nötig, da die Daten beim Event-Write validiert wurden.
type positionenMitKommentarEventData struct {
	Positionen []kasse.KommentiertePositionV2EventData `json:"positionen"`
	Kommentar  string                                  `json:"kommentar"`
}

type arbeitsbonDaten struct {
	Positionen []kasse.Position
	Kommentar  string
}

// CreateArbeitsbonAuftraegeFromEvent routes the event's positions to their Druckziel: the
// Zusatzdrucker of the Produkt (keyed by VarianteID) before the Druckstation of the Kategorie.
// A target without a printer IP creates no job; an assigned Zusatzdrucker never falls back to the
// Kategorie. Pfand positions print on no Arbeitsbon: a position counts as Pfand when its variant is in
// pfandVarianten (the Stammdaten, as bestellung-aufgenommen:v1 carries no flag) or its event flag is set.
// Routing rule: docs/handbuch.md §4.5.
func CreateArbeitsbonAuftraegeFromEvent(
	evt event.Event,
	druckstationen map[string]druckstation.Druckstation,
	zusatzdrucker map[int]druckstation.Zusatzdrucker,
	pfandVarianten map[int]bool,
	tischName string,
) []druckstation.NeuerDruckauftrag {
	switch kasse.EventType(evt.Type) {
	case kasse.EventTypeBestellungAufgenommenV1:
		return createStationsAuftraege(evt, druckstationen, zusatzdrucker, pfandVarianten, tischName, druckstation.EventReferenz(druckstation.ReferenzBestellung, evt.ID))
	case kasse.EventTypeDirektverkaufGetaetigtV1, kasse.EventTypeDirektverkaufGetaetigtV2:
		return createDirektverkaufAuftraege(evt, druckstationen, zusatzdrucker, pfandVarianten)
	case kasse.EventTypeZahlungKassiertV1, kasse.EventTypeZahlungKassiertV2, kasse.EventTypeStornierungErteiltV1,
		kasse.EventTypeBestellungKorrigiertV1, kasse.EventTypeBestellungUmgebuchtV1,
		kasse.EventTypeDirektverkaufStorniertV1, kasse.EventTypePfandZurueckgegebenV1, kasse.EventTypePfandrueckgabeStorniertV1,
		kasse.EventTypeKassensitzungEroeffnetV1, kasse.EventTypeGeldtransitGebuchtV1, kasse.EventTypeKassensturzDurchgefuehrtV1,
		kasse.EventTypeDifferenzSollIstGebuchtV1, kasse.EventTypeTagesabschlussErstelltV1:
		return nil
	default:
		return nil
	}
}

func createDirektverkaufAuftraege(
	evt event.Event,
	druckstationen map[string]druckstation.Druckstation,
	zusatzdrucker map[int]druckstation.Zusatzdrucker,
	pfandVarianten map[int]bool,
) []druckstation.NeuerDruckauftrag {
	data, ok := unmarshalPositionenOhnePfand(evt, pfandVarianten)
	if !ok {
		return nil
	}

	referenz := druckstation.EventReferenz(druckstation.ReferenzDirektverkauf, evt.ID)

	abholbon, ok := druckstationen[string(druckstation.KategorieAbholbon)]
	if ok && abholbon.Bonmodus == druckstation.BonmodusKeinBon {
		return nil
	}
	if ok && abholbon.DruckerIP != "" {
		return createAbholbonAuftraege(evt, data, abholbon, referenz)
	}

	return createStationsAuftraegeFromData(evt, data, druckstationen, zusatzdrucker, "Direktverkauf", referenz)
}

func createAbholbonAuftraege(
	evt event.Event,
	data arbeitsbonDaten,
	station druckstation.Druckstation,
	referenz string,
) []druckstation.NeuerDruckauftrag {
	positionen := data.Positionen
	if len(positionen) == 0 {
		return nil
	}

	abholbon := func(bonPositionen []kasse.Position) druckstation.NeuerDruckauftrag {
		payload := escpos.FormatDirektverkaufAbholbon(bonPositionen, evt.UserName, evt.Time, data.Kommentar)
		return druckstation.NeuerDruckauftrag{
			ZielIP:   station.DruckerIP,
			Payload:  base64.StdEncoding.EncodeToString(payload),
			BonArt:   druckstation.BonArtArbeitsbon,
			Referenz: referenz,
		}
	}

	switch station.Bonmodus {
	case druckstation.BonmodusProBestellung:
		return []druckstation.NeuerDruckauftrag{abholbon(positionen)}

	case druckstation.BonmodusProStueck:
		var auftraege []druckstation.NeuerDruckauftrag
		for _, pos := range positionen {
			einheit := pos
			einheit.Menge = 1
			for range pos.Menge {
				auftraege = append(auftraege, abholbon([]kasse.Position{einheit}))
			}
		}
		return auftraege

	case druckstation.BonmodusKeinBon:
		return nil

	case druckstation.BonmodusProPosition:
		fallthrough
	default: // Rückfall für unbekannte Werte
		auftraege := make([]druckstation.NeuerDruckauftrag, 0, len(positionen))
		for _, pos := range positionen {
			auftraege = append(auftraege, abholbon([]kasse.Position{pos}))
		}
		return auftraege
	}
}

func createStationsAuftraege(
	evt event.Event,
	druckstationen map[string]druckstation.Druckstation,
	zusatzdrucker map[int]druckstation.Zusatzdrucker,
	pfandVarianten map[int]bool,
	kontextName string,
	referenz string,
) []druckstation.NeuerDruckauftrag {
	data, ok := unmarshalPositionenOhnePfand(evt, pfandVarianten)
	if !ok {
		return nil
	}

	return createStationsAuftraegeFromData(evt, data, druckstationen, zusatzdrucker, kontextName, referenz)
}

// druckziel is a printer the positions of one Bestellung go to: a Zusatzdrucker or the Druckstation of a Kategorie.
type druckziel struct {
	druckerIP  string
	bonmodus   druckstation.Bonmodus
	positionen []kasse.Position
}

// gruppiereNachDruckziel keeps the order of each target's first position, so the job list is deterministic.
func gruppiereNachDruckziel(
	positionen []kasse.Position,
	druckstationen map[string]druckstation.Druckstation,
	zusatzdrucker map[int]druckstation.Zusatzdrucker,
) []*druckziel {
	// zusatzdruckerID 0 means the Kategorie's Druckstation; the column's IDs start at 1.
	type zielKey struct {
		zusatzdruckerID int
		kategorie       string
	}

	var ziele []*druckziel
	byKey := map[zielKey]*druckziel{}
	for _, pos := range positionen {
		var key zielKey
		var ziel druckziel
		if zd, ok := zusatzdrucker[pos.VarianteID]; ok {
			key = zielKey{zusatzdruckerID: zd.ID}
			ziel = druckziel{druckerIP: zd.DruckerIP, bonmodus: zd.Bonmodus}
		} else {
			key = zielKey{kategorie: pos.Kategorie}
			station := druckstationen[pos.Kategorie]
			ziel = druckziel{druckerIP: station.DruckerIP, bonmodus: station.Bonmodus}
		}

		gruppe, ok := byKey[key]
		if !ok {
			gruppe = &ziel
			byKey[key] = gruppe
			ziele = append(ziele, gruppe)
		}
		gruppe.positionen = append(gruppe.positionen, pos)
	}
	return ziele
}

func createStationsAuftraegeFromData(
	evt event.Event,
	data arbeitsbonDaten,
	druckstationen map[string]druckstation.Druckstation,
	zusatzdrucker map[int]druckstation.Zusatzdrucker,
	kontextName string,
	referenz string,
) []druckstation.NeuerDruckauftrag {
	auftrag := func(ziel *druckziel, payload []byte) druckstation.NeuerDruckauftrag {
		return druckstation.NeuerDruckauftrag{
			ZielIP:   ziel.druckerIP,
			Payload:  base64.StdEncoding.EncodeToString(payload),
			BonArt:   druckstation.BonArtArbeitsbon,
			Referenz: referenz,
		}
	}

	var auftraege []druckstation.NeuerDruckauftrag
	for _, ziel := range gruppiereNachDruckziel(data.Positionen, druckstationen, zusatzdrucker) {
		if ziel.druckerIP == "" {
			continue
		}

		if ziel.bonmodus == druckstation.BonmodusProBestellung {
			payload := escpos.FormatSammelBon(ziel.positionen, kontextName, evt.UserName, evt.Time, data.Kommentar)
			auftraege = append(auftraege, auftrag(ziel, payload))
			continue
		}

		for _, pos := range ziel.positionen {
			payload := escpos.FormatPositionBon(pos, kontextName, evt.UserName, evt.Time, data.Kommentar)
			auftraege = append(auftraege, auftrag(ziel, payload))
		}
	}

	return auftraege
}

// unmarshalPositionenOhnePfand drops the Pfand positions: Pfand is money, not work for kitchen or bar, so no
// Arbeitsbon or Abholbon lists it (D26).
func unmarshalPositionenOhnePfand(evt event.Event, pfandVarianten map[int]bool) (arbeitsbonDaten, bool) {
	data, err := event.Decode[positionenMitKommentarEventData](evt)
	if err != nil {
		return arbeitsbonDaten{}, false
	}

	alle := kasse.PositionenV2MitKommentar(data.Positionen)
	positionen := make([]kasse.Position, 0, len(alle))
	for _, pos := range alle {
		if !pos.Pfand && !pfandVarianten[pos.VarianteID] {
			positionen = append(positionen, pos)
		}
	}
	return arbeitsbonDaten{Positionen: positionen, Kommentar: data.Kommentar}, true
}

// PfandVarianten marks the variants of the enriched positions that belong to a Pfand-Produkt, the form
// CreateArbeitsbonAuftraegeFromEvent takes.
func PfandVarianten(positionen []kasse.Position) map[int]bool {
	pfand := map[int]bool{}
	for _, pos := range positionen {
		if pos.Pfand {
			pfand[pos.VarianteID] = true
		}
	}
	return pfand
}
