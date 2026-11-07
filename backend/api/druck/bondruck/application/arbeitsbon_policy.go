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
// Positionskommentar; den pfand-Schlüssel ignoriert sie.
// Keine Schema-Validierung nötig, da die Daten beim Event-Write validiert wurden.
type positionenMitKommentarEventData struct {
	Positionen []kasse.KommentiertePositionEventData `json:"positionen"`
	Kommentar  string                                `json:"kommentar"`
}

type arbeitsbonDaten struct {
	Positionen []kasse.Position
	Kommentar  string
}

// CreateArbeitsbonAuftraegeFromEvent routes the event's positions to the configured
// Druckstationen per Bonmodus; without a configured station no job is created.
// Routing rule: docs/handbuch.md §4.5.
func CreateArbeitsbonAuftraegeFromEvent(
	evt event.Event,
	druckstationen map[string]druckstation.Druckstation,
	tischName string,
) []druckstation.NeuerDruckauftrag {
	switch kasse.EventType(evt.Type) {
	case kasse.EventTypeBestellungAufgenommenV1:
		return createStationsAuftraege(evt, druckstationen, tischName, druckstation.EventReferenz(druckstation.ReferenzBestellung, evt.ID))
	case kasse.EventTypeDirektverkaufGetaetigtV1, kasse.EventTypeDirektverkaufGetaetigtV2:
		return createDirektverkaufAuftraege(evt, druckstationen)
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
) []druckstation.NeuerDruckauftrag {
	data, ok := unmarshalPositionenMitKommentar(evt)
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

	return createStationsAuftraegeFromData(evt, data, druckstationen, "Direktverkauf", referenz)
}

func createAbholbonAuftraege(
	evt event.Event,
	data arbeitsbonDaten,
	station druckstation.Druckstation,
	referenz string,
) []druckstation.NeuerDruckauftrag {
	abholbon := func(positionen []kasse.Position) druckstation.NeuerDruckauftrag {
		payload := escpos.FormatDirektverkaufAbholbon(positionen, evt.UserName, evt.Time, data.Kommentar)
		return druckstation.NeuerDruckauftrag{
			ZielIP:   station.DruckerIP,
			Payload:  base64.StdEncoding.EncodeToString(payload),
			BonArt:   druckstation.BonArtArbeitsbon,
			Referenz: referenz,
		}
	}

	switch station.Bonmodus {
	case druckstation.BonmodusProBestellung:
		return []druckstation.NeuerDruckauftrag{abholbon(data.Positionen)}

	case druckstation.BonmodusProStueck:
		var auftraege []druckstation.NeuerDruckauftrag
		for _, pos := range data.Positionen {
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
		auftraege := make([]druckstation.NeuerDruckauftrag, 0, len(data.Positionen))
		for _, pos := range data.Positionen {
			auftraege = append(auftraege, abholbon([]kasse.Position{pos}))
		}
		return auftraege
	}
}

func createStationsAuftraege(
	evt event.Event,
	druckstationen map[string]druckstation.Druckstation,
	kontextName string,
	referenz string,
) []druckstation.NeuerDruckauftrag {
	data, ok := unmarshalPositionenMitKommentar(evt)
	if !ok {
		return nil
	}

	return createStationsAuftraegeFromData(evt, data, druckstationen, kontextName, referenz)
}

func createStationsAuftraegeFromData(
	evt event.Event,
	data arbeitsbonDaten,
	druckstationen map[string]druckstation.Druckstation,
	kontextName string,
	referenz string,
) []druckstation.NeuerDruckauftrag {

	byKategorie := map[string][]kasse.Position{}
	for _, pos := range data.Positionen {
		byKategorie[pos.Kategorie] = append(byKategorie[pos.Kategorie], pos)
	}

	var auftraege []druckstation.NeuerDruckauftrag
	for kategorie, positionen := range byKategorie {
		konfig, ok := druckstationen[kategorie]
		if !ok || konfig.DruckerIP == "" {
			continue
		}

		if konfig.Bonmodus == druckstation.BonmodusProBestellung {
			payload := escpos.FormatSammelBon(
				positionen,
				kontextName,
				evt.UserName,
				evt.Time,
				data.Kommentar,
			)
			auftraege = append(auftraege, druckstation.NeuerDruckauftrag{
				ZielIP:   konfig.DruckerIP,
				Payload:  base64.StdEncoding.EncodeToString(payload),
				BonArt:   druckstation.BonArtArbeitsbon,
				Referenz: referenz,
			})
			continue
		}

		for _, pos := range positionen {
			payload := escpos.FormatPositionBon(
				pos,
				kontextName,
				evt.UserName,
				evt.Time,
				data.Kommentar,
			)
			auftraege = append(auftraege, druckstation.NeuerDruckauftrag{
				ZielIP:   konfig.DruckerIP,
				Payload:  base64.StdEncoding.EncodeToString(payload),
				BonArt:   druckstation.BonArtArbeitsbon,
				Referenz: referenz,
			})
		}
	}

	return auftraege
}

func unmarshalPositionenMitKommentar(evt event.Event) (arbeitsbonDaten, bool) {
	data, err := event.Decode[positionenMitKommentarEventData](evt)
	if err != nil {
		return arbeitsbonDaten{}, false
	}

	return arbeitsbonDaten{Positionen: kasse.PositionenMitKommentar(data.Positionen), Kommentar: data.Kommentar}, true
}
