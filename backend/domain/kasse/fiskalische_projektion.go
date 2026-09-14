package kasse

import (
	"encoding/json"
	"fmt"

	e "github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/tse"
)

// FiskalischerVorgang ist processType und processData (DSFinV-K Anhang I) eines
// signaturpflichtigen Events — der Snapshot für den Signaturauftrag.
type FiskalischerVorgang struct {
	ProcessType string
	ProcessData string
}

// FiskalischeProjektion ist die einzige Stelle, die über Signaturpflicht entscheidet, und
// entscheidet datenabhängig: Die Sitzungseröffnung ist nur bei Anfangsbestand > 0 ein
// Geschäftsvorfall (Bareinlage, AEAO 2.2.3.6.1). Ein unbekannter Event-Typ ist ein Fehler,
// damit ein neuer Typ ohne Projektions-Eintrag nicht still unsigniert bleibt.
func FiskalischeProjektion(evt e.Event) (FiskalischerVorgang, bool, error) {
	switch EventType(evt.Type) {
	case EventTypeBestellungAufgenommenV1:
		data, err := parseProjektionsData[BestellungAufgenommenV1Data](evt)
		if err != nil {
			return FiskalischerVorgang{}, false, err
		}
		return bestellungVorgang(fromPositionenEventData(data.Positionen), 1)

	case EventTypeBestellungKorrigiertV1:
		data, err := parseProjektionsData[BestellungKorrigiertV1Data](evt)
		if err != nil {
			return FiskalischerVorgang{}, false, err
		}
		return bestellungVorgang(fromPositionenEventData(data.Positionen), -1)

	case EventTypeBestellungUmgebuchtV1:
		// Abgang (Quelltisch) negativ, Zugang (Zieltisch) positiv — sonst erschiene die Ware
		// TSE-seitig doppelt bestellt. Die Seite folgt aus dem Tisch des Subjects.
		data, err := parseProjektionsData[BestellungUmgebuchtV1Data](evt)
		if err != nil {
			return FiskalischerVorgang{}, false, err
		}
		tischID, err := ParseTischIDFromSubject(evt.Subject)
		if err != nil {
			return FiskalischerVorgang{}, false, fmt.Errorf("fiskalische projektion %s: %w", evt.Type, err)
		}
		faktor := 1
		if tischID == data.QuellTischID {
			faktor = -1
		}
		return bestellungVorgang(fromPositionenEventData(data.Positionen), faktor)

	case EventTypeZahlungKassiertV1:
		data, err := parseProjektionsData[ZahlungKassiertV1Data](evt)
		if err != nil {
			return FiskalischerVorgang{}, false, err
		}
		return kassenbelegVorgang(fromPositionenEventData(data.Positionen), data.GesamtZahlungCents, 1)

	case EventTypeStornierungErteiltV1:
		data, err := parseProjektionsData[StornierungErteiltV1Data](evt)
		if err != nil {
			return FiskalischerVorgang{}, false, err
		}
		return kassenbelegVorgang(fromPositionenEventData(data.Positionen), -data.GesamtStornierungCents, -1)

	case EventTypeDirektverkaufGetaetigtV1:
		data, err := parseProjektionsData[DirektverkaufGetaetigtV1Data](evt)
		if err != nil {
			return FiskalischerVorgang{}, false, err
		}
		return kassenbelegVorgang(fromPositionenEventData(data.Positionen), data.GesamtbetragCents, 1)

	case EventTypeDirektverkaufStorniertV1:
		data, err := parseProjektionsData[DirektverkaufStorniertV1Data](evt)
		if err != nil {
			return FiskalischerVorgang{}, false, err
		}
		return kassenbelegVorgang(fromPositionenEventData(data.Positionen), -data.GesamtStornierungCents, -1)

	case EventTypeKassensitzungEroeffnetV1:
		data, err := parseProjektionsData[KassensitzungEroeffnetV1Data](evt)
		if err != nil {
			return FiskalischerVorgang{}, false, err
		}
		if data.BetragCents == 0 {
			return FiskalischerVorgang{}, false, nil
		}
		return FiskalischerVorgang{
			ProcessType: tse.ProcessTypeKassenbelegV1,
			ProcessData: BuildEigenbelegProcessData(data.BetragCents),
		}, true, nil

	case EventTypeGeldtransitGebuchtV1:
		data, err := parseProjektionsData[GeldtransitGebuchtV1Data](evt)
		if err != nil {
			return FiskalischerVorgang{}, false, err
		}
		processData, err := BuildGeldtransitProcessData(data.Richtung, data.BetragCents)
		if err != nil {
			return FiskalischerVorgang{}, false, fmt.Errorf("fiskalische projektion %s: %w", evt.Type, err)
		}
		return FiskalischerVorgang{ProcessType: tse.ProcessTypeKassenbelegV1, ProcessData: processData}, true, nil

	case EventTypeDifferenzSollIstGebuchtV1:
		// BetragCents = Soll − Ist; die Bargeldbewegung ist Ist − Soll, daher negiert.
		data, err := parseProjektionsData[DifferenzSollIstGebuchtV1Data](evt)
		if err != nil {
			return FiskalischerVorgang{}, false, err
		}
		return FiskalischerVorgang{
			ProcessType: tse.ProcessTypeKassenbelegV1,
			ProcessData: BuildEigenbelegProcessData(-data.BetragCents),
		}, true, nil

	case EventTypeTagesabschlussErstelltV1:
		data, err := parseProjektionsData[TagesabschlussErstelltV1Data](evt)
		if err != nil {
			return FiskalischerVorgang{}, false, err
		}
		return FiskalischerVorgang{
			ProcessType: tse.ProcessTypeSonstigerVorgang,
			ProcessData: BuildTagesabschlussProcessData(data.ZNr, data.ZeitraumVon, data.ZeitraumBis),
		}, true, nil

	case EventTypeKassensturzDurchgefuehrtV1:
		return FiskalischerVorgang{}, false, nil

	default:
		return FiskalischerVorgang{}, false, fmt.Errorf("fiskalische projektion: unbekannter event-typ %q", evt.Type)
	}
}

func bestellungVorgang(positionen []Position, faktor int) (FiskalischerVorgang, bool, error) {
	processData, err := BuildBestellungProcessData(positionen, faktor)
	if err != nil {
		return FiskalischerVorgang{}, false, fmt.Errorf("fiskalische projektion bestellung: %w", err)
	}
	return FiskalischerVorgang{ProcessType: tse.ProcessTypeBestellungV1, ProcessData: processData}, true, nil
}

func kassenbelegVorgang(positionen []Position, zahlbetragCents int, faktor int) (FiskalischerVorgang, bool, error) {
	processData, err := BuildKassenbelegProcessData(positionen, zahlbetragCents, faktor)
	if err != nil {
		return FiskalischerVorgang{}, false, fmt.Errorf("fiskalische projektion kassenbeleg: %w", err)
	}
	return FiskalischerVorgang{ProcessType: tse.ProcessTypeKassenbelegV1, ProcessData: processData}, true, nil
}

func parseProjektionsData[T any](evt e.Event) (T, error) {
	var data T
	if err := json.Unmarshal(evt.Data, &data); err != nil {
		return data, fmt.Errorf("fiskalische projektion %s: event-daten parsen: %w", evt.Type, err)
	}
	return data, nil
}
