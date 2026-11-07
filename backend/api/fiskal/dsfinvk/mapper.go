package dsfinvk

import (
	"errors"
	"time"

	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
)

// ErrKeineVorgaenge marks a Kassensitzung without a single Bon; the caller
// reports it instead of shipping an empty archive.
var ErrKeineVorgaenge = errors.New("kassensitzung enthält keine vorgänge")

const (
	bonTypBeleg      = "Beleg"        // Anhang B: completed Kassenvorgang (payment)
	bonTypBestellung = "AVBestellung" // Anhang B: order as "anderer Vorgang", cash-neutral
	bonTypSonstige   = "AVSonstige"   // Anhang B: other "anderer Vorgang" (Tagesabschluss), cash-neutral
	gvTypUmsatz      = "Umsatz"       // Anhang C: realised revenue at line level
	// Anhang C: the Pfand line of a sale and its Warenrücknahme. Only the v2 sale events carry the
	// Pfand flag per position; v1 sale lines stay "Umsatz" (docs/decisions.md D20).
	gvTypPfand = "Pfand"
	// Anhang C: Pfand paid out to the guest, a negative Entgeltminderung at the Pfand
	// product's VAT rate (Tz. 4.2.5).
	gvTypPfandRueckzahlung = "PfandRueckzahlung"
	// Anhang C GV types that only move the cash balance; jotti records them as
	// BON_TYP "Beleg" with one non-taxable line (see docs/compliance.md §6.8).
	gvTypAnfangsbestand   = "Anfangsbestand"         // cash at session start (opening event)
	gvTypGeldtransit      = "Geldtransit"            // cash deposit or withdrawal (e.g. bank, safe)
	gvTypDifferenzSollIst = "DifferenzSollIst"       // booked cash difference from the Kassensturz
	zahlartBar            = "Bar"                    // Anhang D: jotti takes cash only
	refTypTransaktion     = "Transaktion"            // Anhang E: REF_TYP for a reference inside the DSFinV-K (Storno → origin)
	tseFehlerAusfall      = "TSE-Ausfall"            // TSE_TA_FEHLER of a Vorgang whose signing failed for good
	tseFehlerAusstehend   = "Signatur ausstehend"    // TSE_TA_FEHLER of a Vorgang still queued for signing
	tseFehlerKeineTSE     = "keine TSE konfiguriert" // TSE_TA_FEHLER of a Vorgang recorded without a TSE
	land                  = "DEU"                    // ISO 3166 ALPHA-3
	basiswaehrung         = "EUR"                    // ISO 4217
	tsePDEncoding         = "UTF-8"                  // encoding of the ProcessData
	zertifikatChunk       = 1000                     // max characters per TSE_ZERTIFIKAT field (the spec has two)
	zertifikatSpalten     = 2                        // TSE_ZERTIFIKAT_I/_II of the official DSFinV-K schema
	// maxLengthAbrechnungskreis is the official ABRECHNUNGSKREIS length in
	// allocation_groups.csv (index.xml), counted in characters.
	maxLengthAbrechnungskreis = 50
	defaultTSEZeitformat      = "unixTime" // fiskaly logs unixTime; fallback without TSE master data
	kasseBrand                = "jotti"
	kasseModell               = "jotti mPOS"
	kasseSoftware             = "jotti"
)

// Archive holds a DSFinV-K export as one Table per CSV file, in the order of
// the official index.xml.
type Archive struct {
	tables []Table
}

func (a Archive) Tables() []Table { return a.tables }

// Map turns a Kassensitzung's snapshot and events into the archive; revenue arises
// at payment (DSFinV-K Tz. 2.7.2, see docs/compliance.md §6.8). signaturen maps
// event ID to signature state, and an event without entry needs no signature.
func Map(snapshot Snapshot, events []event.Event, signaturen map[int]tse.EventSignatur) (Archive, error) {
	erstellung := snapshot.Erstellung.UTC().Format(time.RFC3339)

	belege, err := belegeFromEvents(events, snapshot.Tischnamen, signaturen)
	if err != nil {
		return Archive{}, err
	}
	if len(belege) == 0 {
		return Archive{}, ErrKeineVorgaenge
	}
	tses := tseVerzeichnisAus(snapshot, belege)

	// All 20 files the official index.xml declares, in its order; unfilled ones as
	// header-only CSV.
	tables := []Table{
		buildCashpointclosing(snapshot, erstellung, belege),
		buildLocation(snapshot, erstellung),
		buildCashregister(snapshot, erstellung),
		headerOnlyTable("slaves.csv", "Stamm_Terminals", "Terminal-Kassen (in jotti nicht vorhanden)", slavesColumns),
		headerOnlyTable("pa.csv", "Stamm_Agenturen", "Agenturgeschäft (in jotti nicht vorhanden)", paColumns),
		buildTSE(snapshot, erstellung, tses),
		buildVat(snapshot, erstellung),
		buildBusinesscases(snapshot, erstellung, belege),
		buildPayment(snapshot, erstellung, belege),
		buildCashPerCurrency(snapshot, erstellung, belege),
		buildTransactions(snapshot, erstellung, belege),
		buildDatapayment(snapshot, erstellung, belege),
		buildLines(snapshot, erstellung, belege),
		headerOnlyTable("itemamounts.csv", "Bonpos_Preisfindung", "Preisfindung je Position (in jotti nicht vorhanden)", itemamountsColumns),
		headerOnlyTable("subitems.csv", "Bonpos_Zusatzinfo", "Zusatzinformationen je Position (in jotti nicht vorhanden)", subitemsColumns),
		buildTransactionsTSE(snapshot, erstellung, belege, tses),
		buildTransactionsVat(snapshot, erstellung, belege),
		buildLinesVat(snapshot, erstellung, belege),
		buildAllocationGroups(snapshot, erstellung, belege),
		buildReferences(snapshot, erstellung, belege),
	}

	return Archive{tables: tables}, nil
}

// zeit formats the event time as ISO 8601 UTC for BON_START/BON_ENDE.
func zeit(ev event.Event) string { return ev.Time.UTC().Format(time.RFC3339) }

// isoZeit formats a TSE logTime for TSE_TA_START/ENDE, which the spec requires as
// ISO 8601 with milliseconds. fiskaly logs whole seconds, so milliseconds are .000.
func isoZeit(t time.Time) string { return t.UTC().Format("2006-01-02T15:04:05.000Z07:00") }

// Erstellungszeitpunkt returns Z_ERSTELLUNG: the tagesabschluss-erstellt time of a
// closed session, else fallback (the export time of an open session).
func Erstellungszeitpunkt(events []event.Event, fallback time.Time) time.Time {
	for _, ev := range events {
		if kasse.EventType(ev.Type) == kasse.EventTypeTagesabschlussErstelltV1 {
			return ev.Time
		}
	}
	return fallback
}

// truncateRunes cuts wert to maxLength runes, as the official field lengths count
// characters; only stored values that skipped write validation (TEXT columns) exceed
// them. Steuernummer and USt-IdNr. are never cut: a truncated number is a wrong one.
func truncateRunes(wert string, maxLength int) string {
	runen := []rune(wert)
	if len(runen) <= maxLength {
		return wert
	}

	return string(runen[:maxLength])
}

// barbestand sums the Bar Belege with sign: sales and Anfangsbestand add, withdrawals,
// Warenrücknahmen and Pfand returns subtract, AVBestellungen carry no Bar. It feeds
// Z_SE_(BAR)ZAHLUNGEN and cash_per_currency.csv.
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

// ordnung returns a key's sort position; unknown keys sort after the known ones,
// alphabetically among themselves.
func ordnung(reihenfolge map[string]int, key string) int {
	if v, ok := reihenfolge[key]; ok {
		return v
	}
	return len(reihenfolge) + 1
}

// ZertifikatZuLang reports whether the TSE certificate exceeds the two official
// TSE_ZERTIFIKAT fields and is exported empty. The caller logs a warning; the archive
// stays valid, as TSE master data and the vendor export hold the full certificate.
func ZertifikatZuLang(cert string) bool {
	return len(cert) > zertifikatSpalten*zertifikatChunk
}

// certChunk returns the index-th 1000-character block of the base64 certificate (ASCII,
// so byte slicing is safe). A certificate longer than both fields (e.g. a whole chain)
// leaves both empty, since a truncated one is worthless.
func certChunk(cert string, index int) string {
	if ZertifikatZuLang(cert) {
		return ""
	}
	start := index * zertifikatChunk
	if start >= len(cert) {
		return ""
	}
	end := min(start+zertifikatChunk, len(cert))
	return cert[start:end]
}

// bonName returns BON_NAME: the spec requires it for the AVSonstige Bon, which jotti
// fills with "Tagesabschluss"; all other Bons leave it empty.
func bonName(b *beleg) string {
	if b.bonTyp == bonTypSonstige {
		return "Tagesabschluss"
	}
	return ""
}

func positionText(p kasse.PositionEventData) string {
	return kasse.PositionFromEventData(p).Bezeichnung()
}
