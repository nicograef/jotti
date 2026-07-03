package dsfinvk

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/steuer"
)

// ErrKeineVorgaenge signalisiert eine Sitzung ohne abrechenbare Belege. Ein
// Archiv ohne einen einzigen Bon ist fachlich leer; der Aufrufer meldet das
// verständlich, statt ein defektes Archiv zu liefern.
var ErrKeineVorgaenge = errors.New("kassensitzung enthält keine vorgänge")

// SignaturNachladen liefert die nachsignierte TSE-Signatur eines Vorgangs über
// seine tx_id aus der Seitentabelle. Der reine Mapper kennt kein I/O; der
// Orchestrator reicht diese Lesefunktion herein. Ein nil-Ergebnis ohne Fehler
// bedeutet: noch nicht nachsigniert (der Vorgang erscheint dann nicht in
// transactions_tse.csv). Eine nil-Funktion bedeutet: keine Seitentabelle zur
// Hand, ein unsignierter Vorgang bleibt ohne Signatur.
type SignaturNachladen func(txID string) (*kasse.TSEData, error)

// Feste DSFinV-K-Ausprägungen und jotti-Kassenidentität.
const (
	bonTypBeleg      = "Beleg"        // Anhang B: abgeschlossener Kassenvorgang (Zahlung)
	bonTypBestellung = "AVBestellung" // Anhang B: Bestellung als anderer Vorgang, geldneutral
	gvTypUmsatz      = "Umsatz"       // Anhang C: realisierter Umsatz auf Positionsebene
	// GV-Typen, die ausschließlich den Kassenbestand betreffen (Anhang C). jotti
	// erfasst sie als BON_TYP "Beleg" mit einer einzigen nicht-steuerbaren Position.
	gvTypAnfangsbestand   = "Anfangsbestand"   // Bargeldbestand zu Sitzungsbeginn (Eröffnungs-Event)
	gvTypGeldtransit      = "Geldtransit"      // Bargeld-Ein-/Entnahme (z. B. zur Bank/Tresor)
	gvTypDifferenzSollIst = "DifferenzSollIst" // gebuchte Kassendifferenz aus dem Kassensturz
	zahlartBar            = "Bar"              // Anhang D: jotti kassiert ausschließlich bar
	refTypTransaktion     = "Transaktion"      // Anhang E: REF_TYP für eine Referenz innerhalb der DSFinV-K (Storno → Ursprung)
	tseReferenzID         = "1"                // eine TSS pro Kasse, im Abschluss als ID 1 referenziert
	tseFehlerAusfall      = "TSE-Ausfall"      // TSE_TA_FEHLER eines unsignierten Ausfall-Vorgangs (noch nicht nachsigniert)
	land                  = "DEU"              // ISO 3166 ALPHA-3
	basiswaehrung         = "EUR"              // ISO 4217
	tsePDEncoding         = "UTF-8"            // Encoding der ProcessData
	zertifikatChunk       = 1000               // max. Zeichen je TSE_ZERTIFIKAT-Feld (amtlich: zwei Felder)
	zertifikatSpalten     = 2                  // TSE_ZERTIFIKAT_I/_II — amtliches Schema der DSFinV-K
	defaultTSEZeitformat  = "unixTime"         // fiskaly liefert unixTime; Fallback ohne Stammdaten
	kasseBrand            = "jotti"
	kasseModell           = "jotti mPOS"
	kasseSoftware         = "jotti"
	kasseSWVersion        = "1.0"
)

// Archive hält die typisierten Zeilen-Kollektionen eines DSFinV-K-Exports, eine
// Table je CSV-Datei in kanonischer Reihenfolge. Für jotti gegenstandslose
// Tabellen (slaves.csv, pa.csv) fehlen und werden daher auch nicht in der
// index.xml deklariert.
type Archive struct {
	tables []Table
}

// Tables liefert die enthaltenen Tabellen in kanonischer Reihenfolge.
func (a Archive) Tables() []Table { return a.tables }

// beleg ist die belegbezogene Zwischensicht, aus der mehrere Tabellen abgeleitet
// werden (Bonkopf, dessen USt-/Zahlart-/Positions-Details und die TSE-Zeile).
type beleg struct {
	bonID            string
	bonNr            int
	bonTyp           string   // BON_TYP: "Beleg" (Zahlung) oder "AVBestellung" (offene Bestellung)
	gvTyp            string   // GV_TYP der Positionen: "Umsatz" oder ein Bargeld-GV-Typ
	zahlart          string   // ZAHLART_TYP: "Bar"; leer bei der geldneutralen AVBestellung
	abrechnungskreis string   // ABRECHNUNGSKREIS (Tischname); leer ohne Tischbezug (z. B. Direktverkauf)
	storno           bool     // negative Belegdarstellung (Warenrücknahme/Korrektur): kehrt das Vorzeichen um; kein Vorgangs-Storno, BON_STORNO bleibt 0
	barabfluss       bool     // mindert den Kassenbestand (Geldtransit-Entnahme, Kassenfehlbetrag); steuert das Vorzeichen wie storno
	geldneutral      bool     // AVBestellung (Bestellung/Korrektur/Umbuchung): TSE-gesichert und informativ in lines.csv, aber ohne Umsatz, USt, Zahlart und Kassenbestandswirkung
	nichtSteuerbar   bool     // Bargeldbewegung ohne USt-Bezug: eine einzige Position mit UST_SCHLUESSEL 5 statt Steueraufteilung
	artikeltext      string   // ARTIKELTEXT der synthetischen Position (nur nichtSteuerbar); sonst aus den Positionen
	refBonIDs        []string // REF_BON_ID je referenziertem Ursprungsbon (Warenrücknahme → Zahlung, Korrektur → Bestellung, Umbuchungs-Zugang → Abgang)
	start            string
	ende             string
	bedienerID       int
	bedienerName     string
	positionen       []kasse.PositionEventData
	bruttoCents      int
	tse              *kasse.TSEData
	tseTxID          string // tx_id zum Nachladen der Signatur aus der Seitentabelle (TSE-Ausfall, später nachsigniert)
	notiz            string
}

// sign liefert das Vorzeichen der Beträge eines Belegs: +1 im Regelfall, -1 für
// einen Negativ-Beleg (Warenrücknahme bezahlter bzw. Korrektur unbezahlter Positionen,
// DSFinV-K Tz. 4.2.5, „Vorzeichen umkehren“) und für einen Bar-Abfluss
// (Geldtransit-Entnahme, Kassenfehlbetrag).
// Beträge liegen im beleg stets als positive Magnitude vor; das Vorzeichen wird
// erst beim Serialisieren der Zeilen gesetzt (die Steueraufteilung rechnet
// ausschließlich mit nicht-negativen Beträgen).
func (b *beleg) sign() int {
	if b.storno || b.barabfluss {
		return -1
	}
	return 1
}

// ustBetrag ist die USt-Aufschlüsselung eines Belegs für einen Steuerschlüssel,
// als positive Magnitude (das Vorzeichen setzt der Aufrufer über beleg.sign()).
type ustBetrag struct {
	schluessel int
	brutto     int
	netto      int
	ust        int
}

// ustAufteilung liefert die USt-Aufschlüsselung des Belegs je Steuerschlüssel.
// Steuerbare Belege werden über die Steuermatrix ihrer Positionen aufgeteilt;
// eine nicht-steuerbare Bargeldbewegung ergibt eine einzige Zeile mit
// UST_SCHLUESSEL 5 (Netto = Brutto, keine USt). Bonkopf-USt (transactions_vat)
// und Kassenabschluss (businesscases) leiten sich aus derselben Quelle ab.
func (b *beleg) ustAufteilung() []ustBetrag {
	if b.nichtSteuerbar {
		return []ustBetrag{{schluessel: ustNichtSteuerbar, brutto: b.bruttoCents, netto: b.bruttoCents, ust: 0}}
	}
	matrix := steuer.Steuermatrix(steuermatrixPositionen(b.positionen))
	out := make([]ustBetrag, 0, len(matrix))
	for _, a := range matrix {
		out = append(out, ustBetrag{schluessel: ustSchluessel(a.Satz), brutto: a.Brutto, netto: a.Netto, ust: a.Steuer})
	}
	return out
}

// Map transformiert Snapshot und Events einer Kassensitzung in das typisierte
// Archiv. Reine Funktion ohne I/O. Der Umsatz entsteht bei der Zahlung
// (Revenue-at-payment, DSFinV-K Tz. 2.7.2) und wird durch eine Warenrücknahme negativ
// gemindert: `zahlung-kassiert` (positiv) und `stornierung-erteilt` (negativer Bar-Beleg
// mit Referenz auf die Zahlung) sind die umsatzwirksamen Belege. `bestellung-aufgenommen`,
// `bestellung-korrigiert` und `bestellung-umgebucht` werden geldneutrale
// `AVBestellung`-Vorgänge (TSE-gesichert, informative Positionen, aber ohne
// Umsatz/USt/Zahlart); Korrektur und Umbuchungs-Zugang tragen zusätzlich eine Referenz
// auf den Ursprung. Hinzu kommen die Direktverkauf-Vorgänge sowie die
// Bargeldbewegungen (Anfangsbestand, Geldtransit, Kassendifferenz).
// Das Kassenabschlussmodul (businesscases, payment, cash_per_currency) aggregiert
// dieselben Belege je GV-Typ und Zahlart.
// signaturNachladen vereinigt während eines TSE-Ausfalls unsigniert persistierte
// Vorgänge mit ihrer später nachsignierten Signatur (Seitentabelle).
func Map(snapshot Snapshot, events []event.Event, signaturNachladen SignaturNachladen) (Archive, error) {
	erstellung := snapshot.Erstellung.UTC().Format(time.RFC3339)

	belege, err := belegeFromEvents(events, snapshot.Tischnamen)
	if err != nil {
		return Archive{}, err
	}
	if len(belege) == 0 {
		return Archive{}, ErrKeineVorgaenge
	}
	if err := nachsigniereBelege(belege, signaturNachladen); err != nil {
		return Archive{}, err
	}

	// Alle 20 amtlich deklarierten Dateien in der Reihenfolge der amtlichen
	// index.xml — nicht befüllte als Header-only-CSV.
	tables := []Table{
		buildCashpointclosing(snapshot, erstellung, belege),
		buildLocation(snapshot, erstellung),
		buildCashregister(snapshot, erstellung),
		headerOnlyTable("slaves.csv", "Stamm_Terminals", "Terminal-Kassen (in jotti nicht vorhanden)", slavesColumns),
		headerOnlyTable("pa.csv", "Stamm_Agenturen", "Agenturgeschäft (in jotti nicht vorhanden)", paColumns),
		buildTSE(snapshot, erstellung, belege),
		buildVat(snapshot, erstellung, belege),
		buildBusinesscases(snapshot, erstellung, belege),
		buildPayment(snapshot, erstellung, belege),
		buildCashPerCurrency(snapshot, erstellung, belege),
		buildTransactions(snapshot, erstellung, belege),
		buildDatapayment(snapshot, erstellung, belege),
		buildLines(snapshot, erstellung, belege),
		headerOnlyTable("itemamounts.csv", "Bonpos_Preisfindung", "Preisfindung je Position (in jotti nicht vorhanden)", itemamountsColumns),
		headerOnlyTable("subitems.csv", "Bonpos_Zusatzinfo", "Zusatzinformationen je Position (in jotti nicht vorhanden)", subitemsColumns),
		buildTransactionsTSE(snapshot, erstellung, belege),
		buildTransactionsVat(snapshot, erstellung, belege),
		buildLinesVat(snapshot, erstellung, belege),
		buildAllocationGroups(snapshot, erstellung, belege),
		buildReferences(snapshot, erstellung, belege),
	}

	return Archive{tables: tables}, nil
}

// belegeFromEvents leitet die Belege aus den Events ab (nach `id` geordnet, daher
// liegt ein Ursprungsbon stets vor seinem Storno). Eine `bestellung-aufgenommen`
// wird zur geldneutralen `AVBestellung` (TSE-gesichert, noch kein Umsatz), eine
// `zahlung-kassiert` zur Umsatzrealisierung. Eine `stornierung-erteilt` ist die
// kassenwirksame Warenrücknahme bezahlter Positionen: ein negativer Bar-Beleg mit
// Referenz auf die Zahlung. Eine `bestellung-korrigiert` ist die geldneutrale
// Stornierung unbezahlter Positionen und verweist auf die Ursprungsbestellung.
// Direktverkäufe sind eigenständige Barbelege ohne Tischbezug, ihr Storno ein
// Negativ-Beleg mit Referenz auf den Ursprungsverkauf. BON_NR wird fortlaufend
// vergeben; BON_ID ist die jeweilige Vorgangs-ID.
func belegeFromEvents(events []event.Event, tischnamen map[int]string) ([]beleg, error) {
	var belege []beleg
	bonNr := 0
	// herkunft bildet jede PositionID auf die BON_ID ihrer Bestellung ab.
	// PositionIDs sind je Bestellung eindeutig, daher löst der Tisch-Storno seine
	// Ursprungsbestellung eindeutig über seine Positionen auf.
	herkunft := map[string]string{}

	for _, ev := range events {
		switch ev.Type {
		case string(kasse.EventTypeBestellungAufgenommenV1):
			var data kasse.BestellungAufgenommenV1Data
			if err := json.Unmarshal(ev.Data, &data); err != nil {
				return nil, fmt.Errorf("unmarshal bestellung-aufgenommen (event %d): %w", ev.ID, err)
			}
			for _, p := range data.Positionen {
				herkunft[p.PositionID] = data.BestellungID
			}
			bonNr++
			belege = append(belege, beleg{
				bonID:            data.BestellungID,
				bonNr:            bonNr,
				bonTyp:           bonTypBestellung,
				gvTyp:            gvTypUmsatz,
				geldneutral:      true,
				abrechnungskreis: abrechnungskreis(ev.Subject, tischnamen),
				start:            zeit(ev),
				ende:             zeit(ev),
				bedienerID:       ev.UserID,
				bedienerName:     ev.UserName,
				positionen:       data.Positionen,
				bruttoCents:      data.GesamtPreisCents,
				tse:              data.TSEData,
				tseTxID:          data.TSETxID,
				notiz:            data.Kommentar,
			})

		case string(kasse.EventTypeZahlungKassiertV1):
			var data kasse.ZahlungKassiertV1Data
			if err := json.Unmarshal(ev.Data, &data); err != nil {
				return nil, fmt.Errorf("unmarshal zahlung-kassiert (event %d): %w", ev.ID, err)
			}
			bonNr++
			belege = append(belege, beleg{
				bonID:            data.ZahlungID,
				bonNr:            bonNr,
				bonTyp:           bonTypBeleg,
				gvTyp:            gvTypUmsatz,
				zahlart:          zahlartBar,
				abrechnungskreis: abrechnungskreis(ev.Subject, tischnamen),
				start:            zeit(ev),
				ende:             zeit(ev),
				bedienerID:       ev.UserID,
				bedienerName:     ev.UserName,
				positionen:       data.Positionen,
				bruttoCents:      data.GesamtZahlungCents,
				tse:              data.TSEData,
				tseTxID:          data.TSETxID,
				notiz:            data.Kommentar,
			})

		case string(kasse.EventTypeStornierungErteiltV1):
			var data kasse.StornierungErteiltV1Data
			if err := json.Unmarshal(ev.Data, &data); err != nil {
				return nil, fmt.Errorf("unmarshal stornierung-erteilt (event %d): %w", ev.ID, err)
			}
			// Kassenwirksame Warenrücknahme bezahlter Positionen: negativer Umsatz am
			// Ursprungssteuersatz mit Bar-Rückgabe (DSFinV-K Tz. 4.2.5), Referenz auf die
			// begleichende Zahlung (Tz. 4.2.2). Negative Belegdarstellung, kein
			// Vorgangs-Storno (BON_STORNO bleibt 0).
			bonNr++
			belege = append(belege, beleg{
				bonID:            data.StornierungID,
				bonNr:            bonNr,
				bonTyp:           bonTypBeleg,
				gvTyp:            gvTypUmsatz,
				zahlart:          zahlartBar,
				storno:           true,
				abrechnungskreis: abrechnungskreis(ev.Subject, tischnamen),
				refBonIDs:        []string{data.ZahlungID},
				start:            zeit(ev),
				ende:             zeit(ev),
				bedienerID:       ev.UserID,
				bedienerName:     ev.UserName,
				positionen:       data.Positionen,
				bruttoCents:      data.GesamtStornierungCents,
				tse:              data.TSEData,
				tseTxID:          data.TSETxID,
				notiz:            data.Kommentar,
			})

		case string(kasse.EventTypeBestellungKorrigiertV1):
			var data kasse.BestellungKorrigiertV1Data
			if err := json.Unmarshal(ev.Data, &data); err != nil {
				return nil, fmt.Errorf("unmarshal bestellung-korrigiert (event %d): %w", ev.ID, err)
			}
			// Geldneutrale Stornierung unbezahlter Positionen: AVBestellung ohne Umsatz,
			// Zahlart oder Kassenbestandswirkung; verweist auf die Ursprungsbestellung.
			bonNr++
			belege = append(belege, beleg{
				bonID:            data.KorrekturID,
				bonNr:            bonNr,
				bonTyp:           bonTypBestellung,
				gvTyp:            gvTypUmsatz,
				geldneutral:      true,
				abrechnungskreis: abrechnungskreis(ev.Subject, tischnamen),
				storno:           true,
				refBonIDs:        ursprungsbons(data.Positionen, herkunft),
				start:            zeit(ev),
				ende:             zeit(ev),
				bedienerID:       ev.UserID,
				bedienerName:     ev.UserName,
				positionen:       data.Positionen,
				bruttoCents:      data.GesamtCents,
				tse:              data.TSEData,
				tseTxID:          data.TSETxID,
				notiz:            data.Kommentar,
			})

		case string(kasse.EventTypeBestellungUmgebuchtV1):
			var data kasse.BestellungUmgebuchtV1Data
			if err := json.Unmarshal(ev.Data, &data); err != nil {
				return nil, fmt.Errorf("unmarshal bestellung-umgebucht (event %d): %w", ev.ID, err)
			}
			tischID, err := kasse.ParseTischIDFromSubject(ev.Subject)
			if err != nil {
				return nil, fmt.Errorf("parse tisch from bestellung-umgebucht (event %d): %w", ev.ID, err)
			}
			// Eine Umbuchung ist geldneutral (AVBestellung, kein Umsatz/Zahlart/Kassen-
			// bestand). Quelle und Ziel teilen sich die UmbuchungID: der Abgang trägt sie
			// als BON_ID, der Zugang referenziert sie — so sind beide Seiten verknüpft.
			bon := beleg{
				bonNr:            0, // unten gesetzt
				bonTyp:           bonTypBestellung,
				gvTyp:            gvTypUmsatz,
				geldneutral:      true,
				abrechnungskreis: abrechnungskreis(ev.Subject, tischnamen),
				start:            zeit(ev),
				ende:             zeit(ev),
				bedienerID:       ev.UserID,
				bedienerName:     ev.UserName,
				positionen:       data.Positionen,
				bruttoCents:      data.GesamtCents,
				tse:              data.TSEData,
				tseTxID:          data.TSETxID,
				notiz:            data.Kommentar,
			}
			if tischID == data.QuellTischID {
				bon.bonID = data.UmbuchungID
			} else {
				bon.bonID = fmt.Sprintf("umbuchung-%d", ev.ID)
				bon.refBonIDs = []string{data.UmbuchungID}
				for _, p := range data.Positionen {
					herkunft[p.PositionID] = bon.bonID
				}
			}
			bonNr++
			bon.bonNr = bonNr
			belege = append(belege, bon)

		case string(kasse.EventTypeDirektverkaufGetaetigtV1):
			var data kasse.DirektverkaufGetaetigtV1Data
			if err := json.Unmarshal(ev.Data, &data); err != nil {
				return nil, fmt.Errorf("unmarshal direktverkauf-getaetigt (event %d): %w", ev.ID, err)
			}
			bonNr++
			belege = append(belege, beleg{
				bonID:        data.VerkaufID,
				bonNr:        bonNr,
				bonTyp:       bonTypBeleg,
				gvTyp:        gvTypUmsatz,
				zahlart:      zahlartBar,
				start:        zeit(ev),
				ende:         zeit(ev),
				bedienerID:   ev.UserID,
				bedienerName: ev.UserName,
				positionen:   data.Positionen,
				bruttoCents:  data.GesamtbetragCents,
				tse:          data.TSEData,
				tseTxID:      data.TSETxID,
				notiz:        data.Kommentar,
			})

		case string(kasse.EventTypeDirektverkaufStorniertV1):
			var data kasse.DirektverkaufStorniertV1Data
			if err := json.Unmarshal(ev.Data, &data); err != nil {
				return nil, fmt.Errorf("unmarshal direktverkauf-storniert (event %d): %w", ev.ID, err)
			}
			bonNr++
			belege = append(belege, beleg{
				bonID:        data.StornierungID,
				bonNr:        bonNr,
				bonTyp:       bonTypBeleg,
				gvTyp:        gvTypUmsatz,
				zahlart:      zahlartBar,
				storno:       true,
				refBonIDs:    []string{data.VerkaufID},
				start:        zeit(ev),
				ende:         zeit(ev),
				bedienerID:   ev.UserID,
				bedienerName: ev.UserName,
				positionen:   data.Positionen,
				bruttoCents:  data.GesamtStornierungCents,
				tse:          data.TSEData,
				tseTxID:      data.TSETxID,
				notiz:        data.Kommentar,
			})

		case string(kasse.EventTypeKassensitzungEroeffnetV1):
			var data kasse.KassensitzungEroeffnetV1Data
			if err := json.Unmarshal(ev.Data, &data); err != nil {
				return nil, fmt.Errorf("unmarshal kassensitzung-eroeffnet (event %d): %w", ev.ID, err)
			}
			// Ohne Bargeld zu Sitzungsbeginn gibt es keinen Anfangsbestand zu
			// dokumentieren (Anhang C: „Erfassung nicht zwingend erforderlich“).
			if data.BetragCents == 0 {
				continue
			}
			bonNr++
			belege = append(belege, geldbewegung(ev, fmt.Sprintf("anfangsbestand-%d", ev.ID), bonNr, gvTypAnfangsbestand, data.BetragCents, false, data.Bezeichnung, data.TSETxID, data.TSEData))

		case string(kasse.EventTypeGeldtransitGebuchtV1):
			var data kasse.GeldtransitGebuchtV1Data
			if err := json.Unmarshal(ev.Data, &data); err != nil {
				return nil, fmt.Errorf("unmarshal geldtransit-gebucht (event %d): %w", ev.ID, err)
			}
			bonNr++
			belege = append(belege, geldbewegung(ev, data.BewegungID, bonNr, gvTypGeldtransit, data.BetragCents, data.Richtung == "entnahme", data.Kommentar, data.TSETxID, data.TSEData))

		case string(kasse.EventTypeDifferenzSollIstGebuchtV1):
			var data kasse.DifferenzSollIstGebuchtV1Data
			if err := json.Unmarshal(ev.Data, &data); err != nil {
				return nil, fmt.Errorf("unmarshal differenz-soll-ist-gebucht (event %d): %w", ev.ID, err)
			}
			// BetragCents = Soll − Ist: ein positiver Wert ist ein Fehlbetrag
			// (Bargeld fehlt, Bestand mindern), ein negativer ein Überschuss.
			bonNr++
			belege = append(belege, geldbewegung(ev, fmt.Sprintf("differenz-soll-ist-%d", ev.ID), bonNr, gvTypDifferenzSollIst, abs(data.BetragCents), data.BetragCents > 0, "", data.TSETxID, data.TSEData))
		}
	}

	return belege, nil
}

// nachsigniereBelege vereinigt die TSE-Signaturen je Beleg (Vorgang): liegt im
// Event-Payload keine Signatur vor (während eines TSE-Ausfalls unsigniert
// persistiert), wird sie über die tx_id aus der Seitentabelle nachgeladen. So
// erscheinen auch nachsignierte Vorgänge vollständig in transactions_tse.csv.
// Belege mit Signatur im Event-Payload bleiben unberührt — die Seitentabelle ist
// reiner Fallback und wird für sie nicht gelesen. Ein Vorgang, der weder im Event
// noch in der Seitentabelle signiert ist (noch nicht nachsigniert), bleibt ohne
// Signatur.
func nachsigniereBelege(belege []beleg, signaturNachladen SignaturNachladen) error {
	if signaturNachladen == nil {
		return nil
	}
	for bi := range belege {
		b := &belege[bi]
		if b.tse != nil || b.tseTxID == "" {
			continue
		}
		nachgeladen, err := signaturNachladen(b.tseTxID)
		if err != nil {
			return fmt.Errorf("tse-signatur nachladen (tx %s): %w", b.tseTxID, err)
		}
		b.tse = nachgeladen
	}
	return nil
}

// geldbewegung baut einen Beleg für eine nicht-steuerbare Bargeldbewegung
// (Anfangsbestand, Geldtransit, Kassendifferenz). Solche Vorgänge sind nach
// DSFinV-K Anhang C BON_TYP "Beleg" mit einer einzigen Position
// (ARTIKELTEXT = GV-Typ, UST_SCHLUESSEL 5). betragCents ist die positive
// Magnitude; das Vorzeichen ergibt sich aus barabfluss. Sie liegen alle auf
// Kassensitzungsebene (kein Tischbezug), daher bleibt der Abrechnungskreis leer.
func geldbewegung(ev event.Event, bonID string, bonNr int, gvTyp string, betragCents int, barabfluss bool, notiz string, tseTxID string, tse *kasse.TSEData) beleg {
	return beleg{
		bonID:          bonID,
		bonNr:          bonNr,
		bonTyp:         bonTypBeleg,
		gvTyp:          gvTyp,
		zahlart:        zahlartBar,
		barabfluss:     barabfluss,
		nichtSteuerbar: true,
		artikeltext:    gvTyp,
		start:          zeit(ev),
		ende:           zeit(ev),
		bedienerID:     ev.UserID,
		bedienerName:   ev.UserName,
		bruttoCents:    betragCents,
		tse:            tse,
		tseTxID:        tseTxID,
		notiz:          notiz,
	}
}

// ursprungsbons liefert die BON_IDs der Bestellungen, aus denen die stornierten
// Positionen stammen — in Reihenfolge des ersten Auftretens und ohne Duplikate.
// Ein Tisch-Storno betrifft meist genau eine Bestellung; über mehrere Bestell-
// runden hinweg kann er mehrere referenzieren (eine references-Zeile je Ursprung).
func ursprungsbons(positionen []kasse.PositionEventData, herkunft map[string]string) []string {
	var bons []string
	gesehen := map[string]bool{}
	for _, p := range positionen {
		bon, ok := herkunft[p.PositionID]
		if !ok || gesehen[bon] {
			continue
		}
		gesehen[bon] = true
		bons = append(bons, bon)
	}
	return bons
}

// zeit formatiert den Event-Zeitstempel als ISO-8601-UTC für BON_START/BON_ENDE.
func zeit(ev event.Event) string { return ev.Time.UTC().Format(time.RFC3339) }

// Erstellungszeitpunkt liefert den Z_ERSTELLUNG-Zeitpunkt der Sitzung: bei einer
// abgeschlossenen Sitzung die Zeit des `tagesabschluss-erstellt`-Events, sonst
// fallback (der Exportzeitpunkt einer noch offenen Sitzung). Der Orchestrator
// ruft die Funktion mit den geladenen Events und reicht das Ergebnis als
// Snapshot.Erstellung weiter.
func Erstellungszeitpunkt(events []event.Event, fallback time.Time) time.Time {
	for _, ev := range events {
		if ev.Type == string(kasse.EventTypeTagesabschlussErstelltV1) {
			return ev.Time
		}
	}
	return fallback
}

// abrechnungskreis leitet den ABRECHNUNGSKREIS aus dem Subject ab: jede
// Tisch-Session ist ein Abrechnungskreis (F-06). Der Name stammt aus den
// Tisch-Stammdaten; fehlt er (gelöschter Tisch), wird "Tisch N" synthetisiert.
// Subjects ohne Tischbezug (Direktverkauf) tragen keinen Abrechnungskreis.
func abrechnungskreis(subject string, tischnamen map[int]string) string {
	tischID, err := kasse.ParseTischIDFromSubject(subject)
	if err != nil {
		return ""
	}
	if name, ok := tischnamen[tischID]; ok {
		return name
	}
	return fmt.Sprintf("Tisch %d", tischID)
}

// --- Stammdatenmodul ---

var cashpointclosingColumns = []column{
	alpha("Z_KASSE_ID"), alpha("Z_ERSTELLUNG"), num("Z_NR", 0),
	alpha("Z_BUCHUNGSTAG"), alpha("TAXONOMIE_VERSION"),
	alpha("Z_START_ID"), alpha("Z_ENDE_ID"),
	alpha("NAME"), alpha("STRASSE"), alpha("PLZ"), alpha("ORT"), alpha("LAND"),
	alpha("STNR"), alpha("USTID"),
	num("Z_SE_ZAHLUNGEN", 2), num("Z_SE_BARZAHLUNGEN", 2),
}

func buildCashpointclosing(s Snapshot, erstellung string, belege []beleg) Table {
	bar := barbestand(belege)
	buchungstag := s.Erstellung.UTC().Format(time.DateOnly)

	record := []string{
		s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
		buchungstag, Version,
		belege[0].bonID, belege[len(belege)-1].bonID,
		s.Betreiber.Vereinsname, s.Betreiber.Strasse, s.Betreiber.Plz, s.Betreiber.Ort, land,
		ptr(s.Betreiber.Steuernummer), ptr(s.Betreiber.UstID),
		formatAmount(bar), formatAmount(bar),
	}

	return Table{
		File:        "cashpointclosing.csv",
		LogicalName: "Stamm_Abschluss",
		Description: "Metadaten zum Kassenabschluss (Z-Bon)",
		Columns:     cashpointclosingColumns,
		Records:     [][]string{record},
	}
}

var locationColumns = []column{
	alpha("Z_KASSE_ID"), alpha("Z_ERSTELLUNG"), num("Z_NR", 0),
	alpha("LOC_NAME"), alpha("LOC_STRASSE"), alpha("LOC_PLZ"), alpha("LOC_ORT"),
	alpha("LOC_LAND"), alpha("LOC_USTID"),
}

func buildLocation(s Snapshot, erstellung string) Table {
	record := []string{
		s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
		s.Betreiber.Vereinsname, s.Betreiber.Strasse, s.Betreiber.Plz, s.Betreiber.Ort,
		land, ptr(s.Betreiber.UstID),
	}

	return Table{
		File:        "location.csv",
		LogicalName: "Stamm_Orte",
		Description: "Betriebsstätte des Betreibers",
		Columns:     locationColumns,
		Records:     [][]string{record},
	}
}

var cashregisterColumns = []column{
	alpha("Z_KASSE_ID"), alpha("Z_ERSTELLUNG"), num("Z_NR", 0),
	alpha("KASSE_BRAND"), alpha("KASSE_MODELL"), alpha("KASSE_SERIENNR"),
	alpha("KASSE_SW_BRAND"), alpha("KASSE_SW_VERSION"),
	alpha("KASSE_BASISWAEH_CODE"), alpha("KEINE_UST_ZUORDNUNG"),
}

func buildCashregister(s Snapshot, erstellung string) Table {
	record := []string{
		s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
		kasseBrand, kasseModell, s.KasseSeriennummer,
		kasseSoftware, kasseSWVersion,
		basiswaehrung, "",
	}

	return Table{
		File:        "cashregister.csv",
		LogicalName: "Stamm_Kassen",
		Description: "Seriennummer, Software-Typ und -Version der Kasse",
		Columns:     cashregisterColumns,
		Records:     [][]string{record},
	}
}

var vatColumns = []column{
	alpha("Z_KASSE_ID"), alpha("Z_ERSTELLUNG"), num("Z_NR", 0),
	num("UST_SCHLUESSEL", 0), num("UST_SATZ", 2), alpha("UST_BESCHR"),
}

// buildVat deklariert die in der Sitzung tatsächlich verwendeten Steuersätze,
// aufsteigend nach Umsatzsteuerschlüssel. Nicht-steuerbare Bargeldbewegungen
// führen den Schlüssel 5 (0 %).
func buildVat(s Snapshot, erstellung string, _ []beleg) Table {
	// Die DSFinV-K-Anlage 2 definiert die USt-Schlüssel 1-7 fest; die vat.csv
	// führt alle vordefinierten Schlüssel auf (nicht nur die in der Sitzung
	// verwendeten), wie es Prüfsoftware erwartet. UST_SATZ je Schlüssel ist
	// amtlich vorgegeben.
	amtlicheSchluessel := [][2]string{
		{"19,00", "Allgemeiner Steuersatz"},
		{"7,00", "Ermäßigter Steuersatz"},
		{"10,70", "Durchschnittsatz (§ 24 Abs. 1 Nr. 3 UStG)"},
		{"5,50", "Durchschnittsatz (§ 24 Abs. 1 Nr. 1 UStG)"},
		{"0,00", "Nicht Steuerbar"},
		{"0,00", "Umsatzsteuerfrei"},
		{"0,00", "UmsatzsteuerNichtErmittelbar"},
	}

	records := make([][]string, 0, len(amtlicheSchluessel))
	for i, eintrag := range amtlicheSchluessel {
		records = append(records, []string{
			s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
			itoa(i + 1), eintrag[0], eintrag[1],
		})
	}

	return Table{
		File:        "vat.csv",
		LogicalName: "Stamm_USt",
		Description: "Verwendete Umsatzsteuersätze",
		Columns:     vatColumns,
		Records:     records,
	}
}

var tseColumns = []column{
	alpha("Z_KASSE_ID"), alpha("Z_ERSTELLUNG"), num("Z_NR", 0),
	num("TSE_ID", 0), alpha("TSE_SERIAL"), alpha("TSE_SIG_ALGO"),
	alpha("TSE_ZEITFORMAT"), alpha("TSE_PD_ENCODING"), alpha("TSE_PUBLIC_KEY"),
	alpha("TSE_ZERTIFIKAT_I"), alpha("TSE_ZERTIFIKAT_II"),
}

func buildTSE(s Snapshot, erstellung string, belege []beleg) Table {
	// TSE_ZEITFORMAT deklariert das Log-Time-Format der TSE selbst (fiskaly:
	// unixTime) und stammt aus den beim Setup gespeicherten TSE-Stammdaten.
	// TSE_TA_START/ENDE sind davon unabhängig amtlich als ISO 8601 vorgegeben.
	zeitformat := s.TSEStammdaten.LogTimeFormat
	if zeitformat == "" {
		zeitformat = defaultTSEZeitformat
	}

	record := []string{
		s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
		tseReferenzID, tseSerial(belege), s.TSEStammdaten.SignaturAlgorithmus,
		zeitformat, tsePDEncoding, s.TSEStammdaten.PublicKey,
	}
	for i := 0; i < zertifikatSpalten; i++ {
		record = append(record, certChunk(s.TSEStammdaten.Zertifikat, i))
	}

	return Table{
		File:        "tse.csv",
		LogicalName: "Stamm_TSE",
		Description: "Stammdaten der technischen Sicherheitseinrichtung",
		Columns:     tseColumns,
		Records:     [][]string{record},
	}
}

// --- Einzelaufzeichnungsmodul ---

var transactionsColumns = []column{
	alpha("Z_KASSE_ID"), alpha("Z_ERSTELLUNG"), num("Z_NR", 0),
	alpha("BON_ID"), num("BON_NR", 0), alpha("BON_TYP"), alpha("BON_NAME"),
	alpha("TERMINAL_ID"), alpha("BON_STORNO"), alpha("BON_START"), alpha("BON_ENDE"),
	alpha("BEDIENER_ID"), alpha("BEDIENER_NAME"), num("UMS_BRUTTO", 2),
	alpha("KUNDE_NAME"), alpha("KUNDE_ID"), alpha("KUNDE_TYP"), alpha("KUNDE_STRASSE"),
	alpha("KUNDE_PLZ"), alpha("KUNDE_ORT"), alpha("KUNDE_LAND"), alpha("KUNDE_USTID"),
	alpha("BON_NOTIZ"),
}

func buildTransactions(s Snapshot, erstellung string, belege []beleg) Table {
	records := make([][]string, 0, len(belege))
	for bi := range belege {
		b := &belege[bi]
		// Die geldneutrale AVBestellung trägt keinen Umsatz (UMS_BRUTTO = 0.00);
		// ihr Bruttobetrag erscheint nur informativ auf Positionsebene (lines.csv).
		umsBrutto := b.sign() * b.bruttoCents
		if b.geldneutral {
			umsBrutto = 0
		}
		// BON_STORNO kennzeichnet die vollständige Aufhebung eines ganzen Belegs. jotti
		// nimmt nur Teilmengen negativ zurück (Warenrücknahme, DSFinV-K Tz. 4.2.5) bzw.
		// korrigiert geldneutral; beides ist kein Vorgangs-Storno, daher bleibt
		// BON_STORNO stets 0 (das negative Vorzeichen trägt b.sign()).
		records = append(records, []string{
			s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
			b.bonID, itoa(b.bonNr), b.bonTyp, "",
			"", storno(false), b.start, b.ende,
			itoa(b.bedienerID), b.bedienerName, formatAmount(umsBrutto),
			"", "", "", "",
			"", "", "", "",
			b.notiz,
		})
	}

	return Table{
		File:        "transactions.csv",
		LogicalName: "Bonkopf",
		Description: "Ein Datensatz je Kassenbon",
		Columns:     transactionsColumns,
		Records:     records,
	}
}

var allocationGroupsColumns = []column{
	alpha("Z_KASSE_ID"), alpha("Z_ERSTELLUNG"), num("Z_NR", 0),
	alpha("BON_ID"), alpha("ABRECHNUNGSKREIS"),
}

// buildAllocationGroups ordnet jeden Bon mit Tischbezug seinem ABRECHNUNGSKREIS
// (Tischname) zu (F-06). Belege ohne Abrechnungskreis (Direktverkauf) bleiben
// außen vor; jede TSE-Transaktion ist über ihren Bon einem Kreis zugeordnet.
func buildAllocationGroups(s Snapshot, erstellung string, belege []beleg) Table {
	var records [][]string
	for bi := range belege {
		b := &belege[bi]
		if b.abrechnungskreis == "" {
			continue
		}
		records = append(records, []string{
			s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
			b.bonID, b.abrechnungskreis,
		})
	}

	return Table{
		File:        "allocation_groups.csv",
		LogicalName: "Bonkopf_AbrKreis",
		Description: "Zuordnung Bon zu Abrechnungskreis (Tisch)",
		Columns:     allocationGroupsColumns,
		Records:     records,
	}
}

var transactionsVatColumns = []column{
	alpha("Z_KASSE_ID"), alpha("Z_ERSTELLUNG"), num("Z_NR", 0),
	alpha("BON_ID"), num("UST_SCHLUESSEL", 0),
	num("BON_BRUTTO", 2), num("BON_NETTO", 2), num("BON_UST", 2),
}

func buildTransactionsVat(s Snapshot, erstellung string, belege []beleg) Table {
	var records [][]string
	for bi := range belege {
		b := &belege[bi]
		if b.geldneutral {
			continue
		}
		for _, z := range b.ustAufteilung() {
			records = append(records, []string{
				s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
				b.bonID, itoa(z.schluessel),
				formatAmount(b.sign() * z.brutto), formatAmount(b.sign() * z.netto), formatAmount(b.sign() * z.ust),
			})
		}
	}

	return Table{
		File:        "transactions_vat.csv",
		LogicalName: "Bonkopf_USt",
		Description: "USt-Aufschlüsselung je Bon",
		Columns:     transactionsVatColumns,
		Records:     records,
	}
}

var datapaymentColumns = []column{
	alpha("Z_KASSE_ID"), alpha("Z_ERSTELLUNG"), num("Z_NR", 0),
	alpha("BON_ID"), alpha("ZAHLART_TYP"), alpha("ZAHLART_NAME"),
	alpha("ZAHLWAEH_CODE"), num("ZAHLWAEH_BETRAG", 2), num("BASISWAEH_BETRAG", 2),
}

func buildDatapayment(s Snapshot, erstellung string, belege []beleg) Table {
	var records [][]string
	for bi := range belege {
		b := &belege[bi]
		if b.geldneutral {
			continue
		}
		records = append(records, []string{
			s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
			b.bonID, b.zahlart, b.zahlart,
			basiswaehrung, formatAmount(b.sign() * b.bruttoCents), formatAmount(b.sign() * b.bruttoCents),
		})
	}

	return Table{
		File:        "datapayment.csv",
		LogicalName: "Bonkopf_Zahlarten",
		Description: "Zahlarten je Bon",
		Columns:     datapaymentColumns,
		Records:     records,
	}
}

var referencesColumns = []column{
	alpha("Z_KASSE_ID"), alpha("Z_ERSTELLUNG"), num("Z_NR", 0),
	alpha("BON_ID"), alpha("POS_ZEILE"), alpha("REF_TYP"), alpha("REF_NAME"),
	alpha("REF_DATUM"), alpha("REF_Z_KASSE_ID"), num("REF_Z_NR", 0), alpha("REF_BON_ID"),
}

// buildReferences verkettet referenzierende Belege mit ihrem Ursprungsvorgang: den
// Storno mit dem stornierten Beleg (Radierverbot, DSFinV-K Tz. 4.2.2) und den
// Umbuchungs-Zugang mit dem zugehörigen Abgang. REF_TYP "Transaktion" verweist
// innerhalb der DSFinV-K; da Ursprung und referenzierender Beleg in derselben
// Sitzung liegen, sind REF_DATUM, REF_Z_KASSE_ID und REF_Z_NR die Abschlusswerte
// dieser Sitzung. POS_ZEILE bleibt leer (Verweis aus dem Bonkopf, nicht aus einer
// Position).
func buildReferences(s Snapshot, erstellung string, belege []beleg) Table {
	var records [][]string
	for bi := range belege {
		b := &belege[bi]
		for _, refBonID := range b.refBonIDs {
			records = append(records, []string{
				s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
				b.bonID, "", refTypTransaktion, "",
				erstellung, s.KasseSeriennummer, itoa(s.KassensitzungNr), refBonID,
			})
		}
	}

	return Table{
		File:        "references.csv",
		LogicalName: "Bon_Referenzen",
		Description: "Referenzen auf andere Bons (Storno/Umbuchung → Ursprung)",
		Columns:     referencesColumns,
		Records:     records,
	}
}

var linesColumns = []column{
	alpha("Z_KASSE_ID"), alpha("Z_ERSTELLUNG"), num("Z_NR", 0),
	alpha("BON_ID"), alpha("POS_ZEILE"), alpha("GUTSCHEIN_NR"), alpha("ARTIKELTEXT"),
	alpha("POS_TERMINAL_ID"), alpha("GV_TYP"), alpha("GV_NAME"), alpha("INHAUS"),
	alpha("P_STORNO"), num("AGENTUR_ID", 0), alpha("ART_NR"), alpha("GTIN"),
	alpha("WARENGR_ID"), alpha("WARENGR"), num("MENGE", 3), num("FAKTOR", 3),
	alpha("EINHEIT"), num("STK_BR", 2),
}

func buildLines(s Snapshot, erstellung string, belege []beleg) Table {
	var records [][]string
	for bi := range belege {
		b := &belege[bi]
		if b.nichtSteuerbar {
			// Bargeldbewegung: eine synthetische Position (ARTIKELTEXT = GV-Typ),
			// Menge ±1 trägt das Vorzeichen, der Stückpreis die positive Magnitude.
			records = append(records, []string{
				s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
				b.bonID, "1", "", b.artikeltext,
				"", b.gvTyp, "", "",
				storno(false), "0", "", "",
				"", "", formatQuantity(b.sign()), "",
				"", formatAmount(b.bruttoCents),
			})
			continue
		}
		// Geldneutrale AVBestellungen tragen keinen Geschäftsvorfall-Typ: Ein
		// GV_TYP "Umsatz" auf ihren Positionen würde bei einer Aggregation der
		// Bonpos je GV_TYP mehr Umsatz ausweisen, als der Kassenabschluss kennt.
		posGvTyp := b.gvTyp
		if b.geldneutral {
			posGvTyp = ""
		}
		for i, p := range b.positionen {
			records = append(records, []string{
				s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
				b.bonID, itoa(i + 1), "", positionText(p),
				"", posGvTyp, "", "",
				storno(false), "0", itoa(p.VarianteID), "",
				p.Kategorie, p.Kategorie, formatQuantity(b.sign() * p.Menge), "",
				"", formatAmount(p.Einzelpreis),
			})
		}
	}

	return Table{
		File:        "lines.csv",
		LogicalName: "Bonpos",
		Description: "Artikelzeilen je Bon",
		Columns:     linesColumns,
		Records:     records,
	}
}

var linesVatColumns = []column{
	alpha("Z_KASSE_ID"), alpha("Z_ERSTELLUNG"), num("Z_NR", 0),
	alpha("BON_ID"), alpha("POS_ZEILE"), num("UST_SCHLUESSEL", 0),
	num("POS_BRUTTO", 2), num("POS_NETTO", 2), num("POS_UST", 2),
}

func buildLinesVat(s Snapshot, erstellung string, belege []beleg) Table {
	var records [][]string
	for bi := range belege {
		b := &belege[bi]
		if b.geldneutral {
			continue
		}
		if b.nichtSteuerbar {
			records = append(records, []string{
				s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
				b.bonID, "1", itoa(ustNichtSteuerbar),
				formatAmount(b.sign() * b.bruttoCents), formatAmount(b.sign() * b.bruttoCents), formatAmount(0),
			})
			continue
		}
		for i, p := range b.positionen {
			brutto := p.Einzelpreis * p.Menge
			for _, aufteilung := range steuer.Aufteilen(brutto, steuer.Steuersatz(p.Steuersatz)) {
				records = append(records, []string{
					s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
					b.bonID, itoa(i + 1), itoa(ustSchluessel(aufteilung.Satz)),
					formatAmount(b.sign() * aufteilung.Brutto), formatAmount(b.sign() * aufteilung.Netto), formatAmount(b.sign() * aufteilung.Steuer),
				})
			}
		}
	}

	return Table{
		File:        "lines_vat.csv",
		LogicalName: "Bonpos_USt",
		Description: "USt-Aufschlüsselung je Artikelzeile",
		Columns:     linesVatColumns,
		Records:     records,
	}
}

// Amtlich deklarierte, in jotti nicht befüllte Tabellen: Sie werden als
// Header-only-CSV mitgeliefert, weil die amtliche index.xml alle 20 Dateien
// deklariert und Prüfsoftware deren Existenz erwartet. jotti hat keine
// Terminal-Kassen (slaves), kein Agenturgeschäft (pa), keine Preisfindung
// (itemamounts) und keine Positions-Zusatzinfos wie Pfand (subitems).
var slavesColumns = []column{
	alpha("Z_KASSE_ID"), alpha("Z_ERSTELLUNG"), num("Z_NR", 0),
	alpha("TERMINAL_ID"), alpha("TERMINAL_BRAND"), alpha("TERMINAL_MODELL"),
	alpha("TERMINAL_SERIENNR"), alpha("TERMINAL_SW_BRAND"), alpha("TERMINAL_SW_VERSION"),
}

var paColumns = []column{
	alpha("Z_KASSE_ID"), alpha("Z_ERSTELLUNG"), num("Z_NR", 0),
	num("AGENTUR_ID", 0), alpha("AGENTUR_NAME"), alpha("AGENTUR_STRASSE"),
	alpha("AGENTUR_PLZ"), alpha("AGENTUR_ORT"), alpha("AGENTUR_LAND"),
	alpha("AGENTUR_STNR"), alpha("AGENTUR_USTID"),
}

var itemamountsColumns = []column{
	alpha("Z_KASSE_ID"), alpha("Z_ERSTELLUNG"), num("Z_NR", 0),
	alpha("BON_ID"), num("POS_ZEILE", 0), alpha("TYP"),
	num("UST_SCHLUESSEL", 0), num("PF_BRUTTO", 5), num("PF_NETTO", 5), num("PF_UST", 5),
}

var subitemsColumns = []column{
	alpha("Z_KASSE_ID"), alpha("Z_ERSTELLUNG"), num("Z_NR", 0),
	alpha("BON_ID"), num("POS_ZEILE", 0), alpha("ZI_ART_NR"),
	alpha("ZI_GTIN"), alpha("ZI_NAME"), alpha("ZI_WARENGR_ID"),
	alpha("ZI_WARENGR"), num("ZI_MENGE", 3), num("ZI_FAKTOR", 3),
	alpha("ZI_EINHEIT"), num("ZI_UST_SCHLUESSEL", 0),
	num("ZI_BASISPREIS_BRUTTO", 5), num("ZI_BASISPREIS_NETTO", 5), num("ZI_BASISPREIS_UST", 5),
}

func headerOnlyTable(file, logicalName, description string, columns []column) Table {
	return Table{File: file, LogicalName: logicalName, Description: description, Columns: columns, Records: nil}
}

var transactionsTSEColumns = []column{
	alpha("Z_KASSE_ID"), alpha("Z_ERSTELLUNG"), num("Z_NR", 0),
	alpha("BON_ID"), num("TSE_ID", 0), num("TSE_TANR", 0),
	alpha("TSE_TA_START"), alpha("TSE_TA_ENDE"), alpha("TSE_TA_VORGANGSART"),
	num("TSE_TA_SIGZ", 0), alpha("TSE_TA_SIG"), alpha("TSE_TA_FEHLER"),
	alpha("TSE_VORGANGSDATEN"),
}

func buildTransactionsTSE(s Snapshot, erstellung string, belege []beleg) Table {
	var records [][]string
	for bi := range belege {
		b := &belege[bi]
		switch {
		case b.tse != nil:
			records = append(records, []string{
				s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
				b.bonID, tseReferenzID, itoa(b.tse.TransactionNumber),
				b.tse.LogTimeStart, b.tse.LogTimeEnd, b.tse.ProcessType,
				itoa(b.tse.SignatureCounter), b.tse.Signature, "",
				b.tse.QRCodeData,
			})
		case b.tseTxID != "":
			// TSE-Ausfall ohne Nachsignierung — für jede Vorgangsart: Ein Beleg
			// mit tx-ID, aber ohne Signatur (weder im Event noch nachsigniert)
			// ist ein (noch) unsignierter, TSE-pflichtiger Vorgang. Statt zu
			// fehlen trägt er eine Fehlerzeile — TSE_TA_FEHLER gesetzt, alle
			// Transaktionsfelder leer (es gab keine abgeschlossene
			// TSE-Transaktion) —, damit jeder Bonkopf eine TSE-Zeile hat.
			// Ohne tx-ID wurde nie signiert (keine TSE konfiguriert): keine Zeile.
			records = append(records, []string{
				s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
				b.bonID, tseReferenzID, "",
				"", "", "",
				"", "", tseFehlerAusfall,
				"",
			})
		}
	}

	return Table{
		File:        "transactions_tse.csv",
		LogicalName: "TSE_Transaktionen",
		Description: "TSE-Transaktionsdaten je Bon",
		Columns:     transactionsTSEColumns,
		Records:     records,
	}
}

// --- Kassenabschlussmodul ---

var businesscasesColumns = []column{
	alpha("Z_KASSE_ID"), alpha("Z_ERSTELLUNG"), num("Z_NR", 0),
	alpha("GV_TYP"), alpha("GV_NAME"), num("AGENTUR_ID", 0), num("UST_SCHLUESSEL", 0),
	num("Z_UMS_BRUTTO", 2), num("Z_UMS_NETTO", 2), num("Z_UST", 2),
}

// gvTypReihenfolge ordnet die Geschäftsvorfalltypen für eine stabile Ausgabe der
// businesscases.csv (Umsatz vor den Bargeldbewegungen).
var gvTypReihenfolge = map[string]int{
	gvTypUmsatz:           0,
	gvTypAnfangsbestand:   1,
	gvTypGeldtransit:      2,
	gvTypDifferenzSollIst: 3,
}

// gvUstSchluessel ist der Aggregationsschlüssel der businesscases.csv: ein
// Geschäftsvorfalltyp je Umsatzsteuersatz.
type gvUstSchluessel struct {
	gvTyp      string
	schluessel int
}

// buildBusinesscases aggregiert die Sitzung je Geschäftsvorfalltyp und
// Steuersatz (DSFinV-K Anhang C). Die Summen entstehen aus denselben Belegen wie
// das Einzelaufzeichnungsmodul, daher gleicht sich die Tagessumme gegen die
// Einzelbons ab.
func buildBusinesscases(s Snapshot, erstellung string, belege []beleg) Table {
	summen := map[gvUstSchluessel]ustBetrag{}
	for bi := range belege {
		b := &belege[bi]
		if b.geldneutral {
			continue
		}
		for _, z := range b.ustAufteilung() {
			key := gvUstSchluessel{gvTyp: b.gvTyp, schluessel: z.schluessel}
			cur := summen[key]
			cur.brutto += b.sign() * z.brutto
			cur.netto += b.sign() * z.netto
			cur.ust += b.sign() * z.ust
			summen[key] = cur
		}
	}

	keys := make([]gvUstSchluessel, 0, len(summen))
	for k := range summen {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		oi, oj := ordnung(gvTypReihenfolge, keys[i].gvTyp), ordnung(gvTypReihenfolge, keys[j].gvTyp)
		if oi != oj {
			return oi < oj
		}
		if keys[i].gvTyp != keys[j].gvTyp {
			return keys[i].gvTyp < keys[j].gvTyp
		}
		return keys[i].schluessel < keys[j].schluessel
	})

	records := make([][]string, 0, len(keys))
	for _, k := range keys {
		summe := summen[k]
		records = append(records, []string{
			s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
			k.gvTyp, "", "0", itoa(k.schluessel),
			formatAmount(summe.brutto), formatAmount(summe.netto), formatAmount(summe.ust),
		})
	}

	return Table{
		File:        "businesscases.csv",
		LogicalName: "Z_GV_Typ",
		Description: "Aggregierte Beträge je Geschäftsvorfalltyp und Steuersatz",
		Columns:     businesscasesColumns,
		Records:     records,
	}
}

var paymentColumns = []column{
	alpha("Z_KASSE_ID"), alpha("Z_ERSTELLUNG"), num("Z_NR", 0),
	alpha("ZAHLART_TYP"), alpha("ZAHLART_NAME"), num("Z_ZAHLART_BETRAG", 2),
}

// zahlartReihenfolge ordnet die Zahlarten der payment.csv. jotti kassiert
// ausschließlich bar; die Map hält die Sortierung offen für künftige Zahlarten.
var zahlartReihenfolge = map[string]int{
	zahlartBar: 0,
}

// buildPayment aggregiert die Beträge je Zahlart (DSFinV-K Anhang D). jotti
// kennt nur Bar; die geldneutrale AVBestellung trägt keine Zahlart bei.
func buildPayment(s Snapshot, erstellung string, belege []beleg) Table {
	summen := map[string]int{}
	for bi := range belege {
		b := &belege[bi]
		if b.geldneutral {
			continue
		}
		summen[b.zahlart] += b.sign() * b.bruttoCents
	}

	zahlarten := make([]string, 0, len(summen))
	for z := range summen {
		zahlarten = append(zahlarten, z)
	}
	sort.Slice(zahlarten, func(i, j int) bool {
		oi, oj := ordnung(zahlartReihenfolge, zahlarten[i]), ordnung(zahlartReihenfolge, zahlarten[j])
		if oi != oj {
			return oi < oj
		}
		return zahlarten[i] < zahlarten[j]
	})

	records := make([][]string, 0, len(zahlarten))
	for _, z := range zahlarten {
		records = append(records, []string{
			s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
			z, z, formatAmount(summen[z]),
		})
	}

	return Table{
		File:        "payment.csv",
		LogicalName: "Z_Zahlart",
		Description: "Aggregierte Summen je Zahlart",
		Columns:     paymentColumns,
		Records:     records,
	}
}

var cashPerCurrencyColumns = []column{
	alpha("Z_KASSE_ID"), alpha("Z_ERSTELLUNG"), num("Z_NR", 0),
	alpha("ZAHLART_WAEH"), num("ZAHLART_BETRAG_WAEH", 2),
}

// buildCashPerCurrency weist den Bargeldbestand zum Abschluss je Währung aus.
// jotti rechnet ausschließlich in EUR; der Bestand ergibt sich aus allen baren
// Belegen (Anfangsbestand, Einnahmen, Geldtransit, Kassendifferenz).
func buildCashPerCurrency(s Snapshot, erstellung string, belege []beleg) Table {
	record := []string{
		s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
		basiswaehrung, formatAmount(barbestand(belege)),
	}

	return Table{
		File:        "cash_per_currency.csv",
		LogicalName: "Z_Waehrungen",
		Description: "Bargeldbestand je Währung zum Abschluss",
		Columns:     cashPerCurrencyColumns,
		Records:     [][]string{record},
	}
}

// --- Hilfsfunktionen ---

// barbestand summiert die baren Belege (vorzeichenbehaftet): Bareinnahmen und
// Anfangsbestand mehren, Geldtransit-Entnahmen und Warenrücknahmen mindern
// den Bestand. Die geldneutrale AVBestellung trägt keine Zahlart "Bar" und bewegt
// daher kein Bargeld. Quelle für Z_SE_(BAR)ZAHLUNGEN und den Bargeldbestand der
// cash_per_currency.csv.
func barbestand(belege []beleg) int {
	bar := 0
	for bi := range belege {
		b := &belege[bi]
		if b.zahlart == zahlartBar {
			bar += b.sign() * b.bruttoCents
		}
	}
	return bar
}

// ordnung liefert die Sortierposition eines Schlüssels; unbekannte Werte landen
// hinter den bekannten und werden untereinander alphabetisch sortiert.
func ordnung(reihenfolge map[string]int, key string) int {
	if v, ok := reihenfolge[key]; ok {
		return v
	}
	return len(reihenfolge) + 1
}

func steuermatrixPositionen(positionen []kasse.PositionEventData) []steuer.SteuermatrixPosition {
	out := make([]steuer.SteuermatrixPosition, len(positionen))
	for i, p := range positionen {
		out[i] = steuer.SteuermatrixPosition{
			Brutto:     p.Einzelpreis * p.Menge,
			Steuersatz: steuer.Steuersatz(p.Steuersatz),
		}
	}
	return out
}

// tseSerial liefert die TSS-Seriennummer aus dem ersten signierten Beleg; alle
// Belege einer Sitzung nutzen dieselbe TSS.
func tseSerial(belege []beleg) string {
	for bi := range belege {
		b := &belege[bi]
		if b.tse != nil {
			return b.tse.SerialNumberTSE
		}
	}
	return ""
}

// certChunk liefert den index-ten 1000-Zeichen-Block des base64-Zertifikats
// (für TSE_ZERTIFIKAT_I und _II). Base64 ist ASCII, daher ist Byte-Slicing sicher.
//
// Passt das Zertifikat nicht in die zwei amtlichen Felder (> 2000 Zeichen, z. B.
// eine ganze Kette), bleiben beide Felder leer statt ein abgeschnittenes — und
// damit wertloses — Zertifikat zu exportieren: Das vollständige Zertifikat liegt
// in den TSE-Stammdaten (DB-Backup) und im TSE-Export des Anbieters vor.
func certChunk(cert string, index int) string {
	if len(cert) > zertifikatSpalten*zertifikatChunk {
		return ""
	}
	start := index * zertifikatChunk
	if start >= len(cert) {
		return ""
	}
	end := start + zertifikatChunk
	if end > len(cert) {
		end = len(cert)
	}
	return cert[start:end]
}

// positionText ist die kanonische Artikelbezeichnung: Produkt- und Variantenname
// mit einem Leerzeichen verbunden.
func positionText(p kasse.PositionEventData) string {
	return kasse.PositionFromEventData(p).Bezeichnung()
}
