package kasse

import (
	"encoding/json"
	"fmt"

	e "github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/tse"
)

// FiskalischerVorgang ist das Ergebnis der fiskalischen Projektion eines
// signaturpflichtigen Events: processType und processData (DSFinV-K Anhang I)
// als Snapshot fuer den Signaturauftrag.
type FiskalischerVorgang struct {
	ProcessType string
	ProcessData string
}

// FiskalischeProjektion bildet ein Event auf (signaturpflichtig, processType,
// processData) ab. Sie ist die einzige Stelle, die ueber Signaturpflicht
// entscheidet, und auch datenabhaengig: Die Sitzungseroeffnung ist nur bei
// Anfangsbestand > 0 ein Geschaeftsvorfall (Bareinlage, AEAO 2.2.3.6.1).
// Unbekannte Event-Typen sind ein Fehler, damit ein neuer Event-Typ ohne
// Projektions-Eintrag nicht still unsigniert bleibt.
func FiskalischeProjektion(evt e.Event) (FiskalischerVorgang, bool, error) {
	switch EventType(evt.Type) {
	case EventTypeBestellungAufgenommenV1:
		data, err := parseProjektionsData[BestellungAufgenommenV1Data](evt)
		if err != nil {
			return FiskalischerVorgang{}, false, err
		}
		return bestellungVorgang(fromPositionenEventData(data.Positionen), 1)

	case EventTypeBestellungKorrigiertV1:
		// Geldneutrale Korrektur: negative Mengen (Anhang I), damit die Ruecknahme
		// TSE-seitig von einer Neubestellung unterscheidbar ist.
		data, err := parseProjektionsData[BestellungKorrigiertV1Data](evt)
		if err != nil {
			return FiskalischerVorgang{}, false, err
		}
		return bestellungVorgang(fromPositionenEventData(data.Positionen), -1)

	case EventTypeBestellungUmgebuchtV1:
		// Der Abgang vom Quelltisch wird mit negativen Mengen signiert, der Zugang
		// auf dem Zieltisch mit positiven — sonst erschiene die Ware TSE-seitig
		// doppelt bestellt. Die Seite ergibt sich aus dem Tisch des Subjects.
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
		// BetragCents = Soll − Ist. Die tatsaechliche Bargeldbewegung ist Ist − Soll:
		// ein Fehlbetrag (Soll > Ist) mindert den Bestand, ein Ueberschuss mehrt ihn.
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
