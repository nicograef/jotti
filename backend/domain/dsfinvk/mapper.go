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

// Feste DSFinV-K-Ausprägungen und jotti-Kassenidentität.
const (
	bonTypBeleg       = "Beleg"                // Anhang B: abgeschlossener Kassenvorgang (Zahlung)
	bonTypBestellung  = "AVBestellung"         // Anhang B: Bestellung als anderer Vorgang, noch kein Umsatz
	gvTypUmsatz       = "Umsatz"               // Anhang C: realisierter Umsatz auf Positionsebene
	gvTypForderung    = "Forderungsentstehung" // Anhang C: offene Bestellung, Umsatz noch nicht realisiert
	zahlartBar        = "Bar"                  // Anhang D: jotti kassiert ausschließlich bar
	zahlartForderung  = "Forderungsentstehung" // Anhang D: Forderung statt Geldzufluss bei offener Bestellung
	refTypTransaktion = "Transaktion"          // Anhang E: REF_TYP für eine Referenz innerhalb der DSFinV-K (Storno → Ursprung)
	tseReferenzID     = "1"                    // eine TSS pro Kasse, im Abschluss als ID 1 referenziert
	land              = "DEU"                  // ISO 3166 ALPHA-3
	basiswaehrung     = "EUR"                  // ISO 4217
	tsePDEncoding     = "UTF-8"                // Encoding der ProcessData
	zertifikatChunk   = 1000                   // max. Zeichen je TSE_ZERTIFIKAT-Feld
	kasseBrand        = "jotti"
	kasseModell       = "jotti mPOS"
	kasseSoftware     = "jotti"
	kasseSWVersion    = "1.0"
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
	bonTyp           string // BON_TYP: "Beleg" (Zahlung) oder "AVBestellung" (offene Bestellung)
	gvTyp            string // GV_TYP der Positionen: "Umsatz" oder "Forderungsentstehung"
	zahlart          string // ZAHLART_TYP: "Bar" oder "Forderungsentstehung"
	abrechnungskreis string // ABRECHNUNGSKREIS (Tischname); leer ohne Tischbezug (z. B. Direktverkauf)
	storno           bool
	refBonIDs        []string // REF_BON_ID je Ursprungsbon (nur Storno); ein Eintrag je referenziertem Vorgang
	start            string
	ende             string
	bedienerID       int
	bedienerName     string
	positionen       []kasse.PositionEventData
	bruttoCents      int
	tse              *kasse.TSEData
	notiz            string
}

// sign liefert das Vorzeichen der Beträge eines Belegs: +1 im Regelfall, -1 für
// einen Storno-Beleg. Das GoBD-Radierverbot verlangt, Stornos als eigene
// Negativ-Datensätze auszuweisen (DSFinV-K Tz. 4.2.2, „Vorzeichen umkehren“),
// statt den Ursprungsbon zu verändern. Beträge liegen im beleg als positive
// Magnitude vor; das Vorzeichen wird erst beim Serialisieren der Zeilen gesetzt
// (die Steueraufteilung rechnet ausschließlich mit nicht-negativen Beträgen).
func (b *beleg) sign() int {
	if b.storno {
		return -1
	}
	return 1
}

// Map transformiert Snapshot und Events einer Kassensitzung in das typisierte
// Archiv. Reine Funktion ohne I/O. Belege entstehen aus `bestellung-aufgenommen`
// (Forderungsentstehung), `zahlung-kassiert` (Umsatzrealisierung), `stornierung-
// erteilt` (Negativ-Beleg mit Referenz auf den Ursprung) sowie aus den
// Direktverkauf-Vorgängen; weitere Vorgangstypen folgen in späteren Phasen.
func Map(snapshot Snapshot, events []event.Event) (Archive, error) {
	erstellung := snapshot.Erstellung.UTC().Format(time.RFC3339)

	belege, err := belegeFromEvents(events, snapshot.Tischnamen)
	if err != nil {
		return Archive{}, err
	}
	if len(belege) == 0 {
		return Archive{}, ErrKeineVorgaenge
	}

	tables := []Table{
		buildCashpointclosing(snapshot, erstellung, belege),
		buildLocation(snapshot, erstellung),
		buildCashregister(snapshot, erstellung),
		buildVat(snapshot, erstellung, belege),
		buildTSE(snapshot, erstellung, belege),
		buildTransactions(snapshot, erstellung, belege),
		buildAllocationGroups(snapshot, erstellung, belege),
		buildTransactionsVat(snapshot, erstellung, belege),
		buildDatapayment(snapshot, erstellung, belege),
		buildReferences(snapshot, erstellung, belege),
		buildLines(snapshot, erstellung, belege),
		buildLinesVat(snapshot, erstellung, belege),
		buildTransactionsTSE(snapshot, erstellung, belege),
	}

	return Archive{tables: tables}, nil
}

// belegeFromEvents leitet die Belege aus den Events ab (nach `id` geordnet, daher
// liegt ein Ursprungsbon stets vor seinem Storno). Eine `bestellung-aufgenommen`
// wird zur Forderungsentstehung (Umsatz noch nicht realisiert), eine `zahlung-
// kassiert` zur Umsatzrealisierung, die die Forderung in bar auflöst. Eine
// `stornierung-erteilt` ist ein Negativ-Beleg, der die Forderung des Ursprungs
// zurücknimmt; Direktverkäufe sind eigenständige Barbelege ohne Tischbezug, ihr
// Storno ein Negativ-Beleg mit Referenz auf den Ursprungsverkauf. BON_NR wird
// fortlaufend vergeben; BON_ID ist die jeweilige Vorgangs-ID.
func belegeFromEvents(events []event.Event, tischnamen map[int]string) ([]beleg, error) {
	var belege []beleg
	bonNr := 0
	// herkunft bildet jede PositionID auf die BON_ID ihrer Bestellung ab.
	// PositionIDs sind je Bestellung eindeutig, daher löst der Tisch-Storno seinen
	// Ursprungsbon (Forderungsentstehung) eindeutig über seine Positionen auf.
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
				gvTyp:            gvTypForderung,
				zahlart:          zahlartForderung,
				abrechnungskreis: abrechnungskreis(ev.Subject, tischnamen),
				start:            zeit(ev),
				ende:             zeit(ev),
				bedienerID:       ev.UserID,
				bedienerName:     ev.UserName,
				positionen:       data.Positionen,
				bruttoCents:      data.GesamtPreisCents,
				tse:              data.TSEData,
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
				notiz:            data.Kommentar,
			})

		case string(kasse.EventTypeStornierungErteiltV1):
			var data kasse.StornierungErteiltV1Data
			if err := json.Unmarshal(ev.Data, &data); err != nil {
				return nil, fmt.Errorf("unmarshal stornierung-erteilt (event %d): %w", ev.ID, err)
			}
			bonNr++
			belege = append(belege, beleg{
				bonID:            data.StornierungID,
				bonNr:            bonNr,
				bonTyp:           bonTypBestellung,
				gvTyp:            gvTypForderung,
				zahlart:          zahlartForderung,
				abrechnungskreis: abrechnungskreis(ev.Subject, tischnamen),
				storno:           true,
				refBonIDs:        ursprungsbons(data.Positionen, herkunft),
				start:            zeit(ev),
				ende:             zeit(ev),
				bedienerID:       ev.UserID,
				bedienerName:     ev.UserName,
				positionen:       data.Positionen,
				bruttoCents:      data.GesamtStornierungCents,
				tse:              data.TSEData,
				notiz:            data.Kommentar,
			})

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
				notiz:        data.Kommentar,
			})
		}
	}

	return belege, nil
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
	// Nur Barzahlungen fließen in den Kassenbestand; offene Bestellungen
	// (Forderungsentstehung) bewegen kein Geld und bleiben außen vor. Ein
	// Bar-Storno (Direktverkauf) mindert den Bestand über sein negatives Vorzeichen.
	bar := 0
	for bi := range belege {
		b := &belege[bi]
		if b.zahlart == zahlartBar {
			bar += b.sign() * b.bruttoCents
		}
	}

	record := []string{
		s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
		"", Version,
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
// aufsteigend nach Umsatzsteuerschlüssel.
func buildVat(s Snapshot, erstellung string, belege []beleg) Table {
	prozentJeSchluessel := map[int]int{}
	for bi := range belege {
		b := &belege[bi]
		for _, aufteilung := range steuer.Steuermatrix(steuermatrixPositionen(b.positionen)) {
			prozentJeSchluessel[ustSchluessel(aufteilung.Satz)] = aufteilung.Satz.Prozent()
		}
	}

	schluessel := make([]int, 0, len(prozentJeSchluessel))
	for k := range prozentJeSchluessel {
		schluessel = append(schluessel, k)
	}
	sort.Ints(schluessel)

	records := make([][]string, 0, len(schluessel))
	for _, k := range schluessel {
		records = append(records, []string{
			s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
			itoa(k), formatPercent(prozentJeSchluessel[k]), ustBeschreibung(k),
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
	record := []string{
		s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
		tseReferenzID, tseSerial(belege), s.TSEStammdaten.SignaturAlgorithmus,
		s.TSEStammdaten.LogTimeFormat, tsePDEncoding, s.TSEStammdaten.PublicKey,
		certChunk(s.TSEStammdaten.Zertifikat, 0), certChunk(s.TSEStammdaten.Zertifikat, 1),
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
		records = append(records, []string{
			s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
			b.bonID, itoa(b.bonNr), b.bonTyp, "",
			"", storno(b.storno), b.start, b.ende,
			itoa(b.bedienerID), b.bedienerName, formatAmount(b.sign() * b.bruttoCents),
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
		for _, aufteilung := range steuer.Steuermatrix(steuermatrixPositionen(b.positionen)) {
			records = append(records, []string{
				s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
				b.bonID, itoa(ustSchluessel(aufteilung.Satz)),
				formatAmount(b.sign() * aufteilung.Brutto), formatAmount(b.sign() * aufteilung.Netto), formatAmount(b.sign() * aufteilung.Steuer),
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
	records := make([][]string, 0, len(belege))
	for bi := range belege {
		b := &belege[bi]
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

// buildReferences verkettet jeden Storno-Beleg mit seinem Ursprungsvorgang
// (Radierverbot, DSFinV-K Tz. 4.2.2). REF_TYP "Transaktion" verweist innerhalb
// der DSFinV-K; da Ursprung und Storno in derselben Sitzung liegen, sind
// REF_DATUM, REF_Z_KASSE_ID und REF_Z_NR die Abschlusswerte dieser Sitzung.
// POS_ZEILE bleibt leer (Verweis aus dem Bonkopf, nicht aus einer Position).
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
		Description: "Referenzen auf andere Bons (Storno → Ursprung)",
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
		for i, p := range b.positionen {
			records = append(records, []string{
				s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
				b.bonID, itoa(i + 1), "", positionText(p),
				"", b.gvTyp, "", "",
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
		if b.tse == nil {
			continue
		}
		records = append(records, []string{
			s.KasseSeriennummer, erstellung, itoa(s.KassensitzungNr),
			b.bonID, tseReferenzID, itoa(b.tse.TransactionNumber),
			b.tse.LogTimeStart, b.tse.LogTimeEnd, b.tse.ProcessType,
			itoa(b.tse.SignatureCounter), b.tse.Signature, "",
			b.tse.QRCodeData,
		})
	}

	return Table{
		File:        "transactions_tse.csv",
		LogicalName: "TSE_Transaktionen",
		Description: "TSE-Transaktionsdaten je Bon",
		Columns:     transactionsTSEColumns,
		Records:     records,
	}
}

// --- Hilfsfunktionen ---

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
// (für TSE_ZERTIFIKAT_I, _II …). Base64 ist ASCII, daher ist Byte-Slicing sicher.
func certChunk(cert string, index int) string {
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
	return kasse.Position(p).Bezeichnung()
}
