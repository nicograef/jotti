//go:build unit

package dsfinvk

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/betreiber"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
)

const (
	testSerial = "11111111-2222-3333-4444-555555555555"
	testBonID  = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
)

// testSignatur baut die quittierte Signatur eines Signaturauftrags, wie sie der
// Export je Event-ID erhaelt.
func testSignatur(t *testing.T, txNr, sigZaehler int, start, ende, signatur string) *tse.Signatur {
	t.Helper()
	logStart, err := time.Parse(time.RFC3339, start)
	if err != nil {
		t.Fatalf("logTimeStart parsen: %v", err)
	}
	logEnd, err := time.Parse(time.RFC3339, ende)
	if err != nil {
		t.Fatalf("logTimeEnd parsen: %v", err)
	}
	return &tse.Signatur{
		TransaktionNummer: txNr,
		SignaturZaehler:   sigZaehler,
		TSESeriennummer:   "abc123serial",
		LogTimeStart:      logStart,
		LogTimeEnd:        logEnd,
		Signatur:          signatur,
		QRCodeData:        "V0;qr",
	}
}

const testSWVersion = "dev"
const testTSESerial = "TSE-STAMMDATEN-SERIAL"

func testSnapshot() Snapshot {
	steuernummer := "12345/67890"
	return Snapshot{
		KasseSeriennummer: testSerial,
		Erstellung:        time.Date(2026, 6, 16, 14, 30, 0, 0, time.UTC),
		KassensitzungNr:   3,
		Betreiber: betreiber.Betreiber{
			Vereinsname:  "TSV Beispiel",
			Strasse:      "Hauptstr. 1",
			Plz:          "12345",
			Ort:          "Musterdorf",
			Steuernummer: &steuernummer,
		},
		TSEStammdaten: tse.Stammdaten{
			Seriennummer:        testTSESerial,
			SignaturAlgorithmus: "ecdsa-plain-SHA256",
			PublicKey:           "PUBKEY==",
			Zertifikat:          "CERTBASE64",
			LogTimeFormat:       "unixTime",
		},
		SoftwareVersion: testSWVersion,
	}
}

// barverkaufEvent baut die einfachste fiskalische Zahlung: ein Bier (19 %) und
// eine Brezel (7 %), bar kassiert.
func barverkaufEvent(t *testing.T) event.Event {
	t.Helper()

	data := kasse.ZahlungKassiertV1Data{
		ZahlungID:          testBonID,
		GesamtZahlungCents: 1050,
		Positionen: []kasse.PositionEventData{
			{PositionID: "p1", VarianteID: 101, ProduktName: "Bier", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 450, Menge: 2},
			{PositionID: "p2", VarianteID: 202, ProduktName: "Brezel", VarianteName: "", Kategorie: "essen", Steuersatz: "ermaessigt", EinzelpreisCents: 150, Menge: 1},
		},
	}

	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal zahlung data: %v", err)
	}

	return event.Event{
		ID:       1,
		UserID:   7,
		UserName: "anna",
		Type:     string(kasse.EventTypeZahlungKassiertV1),
		Time:     time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC),
		Subject:  kasse.TischSessionSubject(3, 42),
		Version:  1,
		Data:     raw,
	}
}

// barverkaufSignaturen ist der Signaturauftrags-Stand zum barverkaufEvent:
// quittierte Kassenbeleg-V1-Signatur.
func barverkaufSignaturen(t *testing.T) map[int]tse.EventSignatur {
	t.Helper()
	return map[int]tse.EventSignatur{
		1: {ProcessType: "Kassenbeleg-V1", Signatur: testSignatur(t, 4711, 12, "2026-06-16T12:00:00Z", "2026-06-16T12:00:01Z", "SIGBASE64==")},
	}
}

func TestMapBarverkaufGoldenRows(t *testing.T) {
	archive, err := Map(testSnapshot(), []event.Event{barverkaufEvent(t)}, barverkaufSignaturen(t))
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}

	const erstellung = "2026-06-16T14:30:00Z"

	wantRecords := map[string][][]string{
		"transactions_vat.csv": {
			{testSerial, erstellung, "3", testBonID, "1", "9,00", "7,56", "1,44"},
			{testSerial, erstellung, "3", testBonID, "2", "1,50", "1,40", "0,10"},
		},
		"lines.csv": {
			{testSerial, erstellung, "3", testBonID, "1", "", "Bier 0,5l", "", "Umsatz", "", "", "0", "0", "101", "", "getraenk", "getraenk", "2,000", "", "", "4,50"},
			{testSerial, erstellung, "3", testBonID, "2", "", "Brezel", "", "Umsatz", "", "", "0", "0", "202", "", "essen", "essen", "1,000", "", "", "1,50"},
		},
		"lines_vat.csv": {
			{testSerial, erstellung, "3", testBonID, "1", "1", "9,00", "7,56", "1,44"},
			{testSerial, erstellung, "3", testBonID, "2", "2", "1,50", "1,40", "0,10"},
		},
		"transactions_tse.csv": {
			{testSerial, erstellung, "3", testBonID, "1", "4711", "2026-06-16T12:00:00.000Z", "2026-06-16T12:00:01.000Z", "Kassenbeleg-V1", "12", "SIGBASE64==", "", ""},
		},
		"vat.csv": {
			{testSerial, erstellung, "3", "1", "19,00", "Allgemeiner Steuersatz"},
			{testSerial, erstellung, "3", "2", "7,00", "Ermäßigter Steuersatz"},
			{testSerial, erstellung, "3", "3", "10,70", "Durchschnittsatz (§ 24 Abs. 1 Nr. 3 UStG)"},
			{testSerial, erstellung, "3", "4", "5,50", "Durchschnittsatz (§ 24 Abs. 1 Nr. 1 UStG)"},
			{testSerial, erstellung, "3", "5", "0,00", "Nicht Steuerbar"},
			{testSerial, erstellung, "3", "6", "0,00", "Umsatzsteuerfrei"},
			{testSerial, erstellung, "3", "7", "0,00", "UmsatzsteuerNichtErmittelbar"},
		},
		"tse.csv": {
			{testSerial, erstellung, "3", "1", testTSESerial, "ecdsa-plain-SHA256", "unixTime", "UTF-8", "PUBKEY==", "CERTBASE64", ""},
		},
		"location.csv": {
			{testSerial, erstellung, "3", "TSV Beispiel", "Hauptstr. 1", "12345", "Musterdorf", "DEU", ""},
		},
		"cashregister.csv": {
			{testSerial, erstellung, "3", "jotti", "jotti mPOS", testSerial, "jotti", testSWVersion, "EUR", ""},
		},
		"cashpointclosing.csv": {
			{testSerial, erstellung, "3", "", "2.4", testBonID, testBonID, "TSV Beispiel", "Hauptstr. 1", "12345", "Musterdorf", "DEU", "12345/67890", "", "10,50", "10,50"},
		},
	}

	for file, want := range wantRecords {
		table := tableByFile(t, archive, file)
		if !reflect.DeepEqual(table.Records, want) {
			t.Errorf("%s records =\n%#v\nwant\n%#v", file, table.Records, want)
		}
	}

	// Bonkopf ist breit; die fiskalisch tragenden Felder gezielt prüfen.
	transactions := tableByFile(t, archive, "transactions.csv")
	if len(transactions.Records) != 1 {
		t.Fatalf("transactions: got %d rows, want 1", len(transactions.Records))
	}
	checks := map[string]string{
		"Z_KASSE_ID":    testSerial,
		"Z_ERSTELLUNG":  erstellung,
		"Z_NR":          "3",
		"BON_ID":        testBonID,
		"BON_NR":        "1",
		"BON_TYP":       "Beleg",
		"BON_STORNO":    "0",
		"BON_START":     "2026-06-16T12:00:00Z",
		"BON_ENDE":      "2026-06-16T12:00:00Z",
		"BEDIENER_ID":   "7",
		"BEDIENER_NAME": "anna",
		"UMS_BRUTTO":    "10,50",
	}
	for name, want := range checks {
		if got := field(t, transactions, 0, name); got != want {
			t.Errorf("transactions %s = %q, want %q", name, got, want)
		}
	}
}

// TestMapZertifikatChunking prüft die amtlichen zwei TSE_ZERTIFIKAT-Felder:
// Ein Zertifikat bis 2.000 Zeichen wird vollständig auf I/II verteilt; ein
// längeres wird NICHT abgeschnitten, sondern beide Felder bleiben leer (das
// vollständige Zertifikat liegt in den TSE-Stammdaten und im TSE-Export vor).
func TestMapZertifikatChunking(t *testing.T) {
	passt := strings.Repeat("A", 1500)
	snap := testSnapshot()
	snap.TSEStammdaten.Zertifikat = passt

	archive, err := Map(snap, []event.Event{barverkaufEvent(t)}, barverkaufSignaturen(t))
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}
	tse := tableByFile(t, archive, "tse.csv")
	gotCert := field(t, tse, 0, "TSE_ZERTIFIKAT_I") + field(t, tse, 0, "TSE_ZERTIFIKAT_II")
	if gotCert != passt {
		t.Errorf("zusammengesetztes Zertifikat = %d Zeichen, want %d", len(gotCert), len(passt))
	}
	if ZertifikatZuLang(passt) {
		t.Error("Zertifikat bis 2000 Zeichen darf nicht als zu lang gelten")
	}

	zuLang := strings.Repeat("B", 2500)
	snap.TSEStammdaten.Zertifikat = zuLang
	// Der Aufrufer (Export) loggt hierauf eine Warnung; die Felder bleiben leer.
	if !ZertifikatZuLang(zuLang) {
		t.Error("Zertifikat > 2000 Zeichen muss als zu lang gelten (Log-Warnung im Export)")
	}
	archive, err = Map(snap, []event.Event{barverkaufEvent(t)}, barverkaufSignaturen(t))
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}
	tse = tableByFile(t, archive, "tse.csv")
	if field(t, tse, 0, "TSE_ZERTIFIKAT_I") != "" || field(t, tse, 0, "TSE_ZERTIFIKAT_II") != "" {
		t.Error("Zertifikat > 2000 Zeichen darf nicht abgeschnitten exportiert werden — Felder müssen leer bleiben")
	}
}

const (
	bestellungBonID = "11111111-1111-1111-1111-111111111111"
	zahlungBonID    = "22222222-2222-2222-2222-222222222222"
)

// bestellungEvent baut eine offene Bestellung (geldneutrale AVBestellung): ein
// Bier, am Tisch 42 aufgenommen.
func bestellungEvent(t *testing.T) event.Event {
	t.Helper()

	data := kasse.BestellungAufgenommenV1Data{
		BestellungID:     bestellungBonID,
		GesamtPreisCents: 450,
		Positionen: []kasse.PositionEventData{
			{PositionID: "p1", VarianteID: 101, ProduktName: "Bier", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 450, Menge: 1},
		},
	}

	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal bestellung data: %v", err)
	}

	return event.Event{
		ID:       1,
		UserID:   7,
		UserName: "anna",
		Type:     string(kasse.EventTypeBestellungAufgenommenV1),
		Time:     time.Date(2026, 6, 16, 11, 0, 0, 0, time.UTC),
		Subject:  kasse.TischSessionSubject(3, 42),
		Version:  1,
		Data:     raw,
	}
}

// zahlungEvent baut die spätere Barzahlung desselben Tisches: der einzige
// umsatzwirksame Beleg (Revenue-at-payment).
func zahlungEvent(t *testing.T) event.Event {
	t.Helper()

	data := kasse.ZahlungKassiertV1Data{
		ZahlungID:          zahlungBonID,
		GesamtZahlungCents: 450,
		Positionen: []kasse.PositionEventData{
			{PositionID: "p1", VarianteID: 101, ProduktName: "Bier", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 450, Menge: 1},
		},
	}

	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal zahlung data: %v", err)
	}

	return event.Event{
		ID:       2,
		UserID:   7,
		UserName: "anna",
		Type:     string(kasse.EventTypeZahlungKassiertV1),
		Time:     time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC),
		Subject:  kasse.TischSessionSubject(3, 42),
		Version:  1,
		Data:     raw,
	}
}

// tischablaufSignaturen ist der Signaturauftrags-Stand zu bestellungEvent (ID 1)
// und zahlungEvent (ID 2): je eine eigene quittierte TSE-Transaktion.
func tischablaufSignaturen(t *testing.T) map[int]tse.EventSignatur {
	t.Helper()
	return map[int]tse.EventSignatur{
		1: {ProcessType: "Bestellung-V1", Signatur: testSignatur(t, 4710, 11, "2026-06-16T11:00:00Z", "2026-06-16T11:00:01Z", "BESTELLSIG==")},
		2: {ProcessType: "Kassenbeleg-V1", Signatur: testSignatur(t, 4711, 12, "2026-06-16T12:00:00Z", "2026-06-16T12:00:01Z", "ZAHLSIG==")},
	}
}

// TestMapTischablaufTrennt belegt Revenue-at-payment für den gastronomischen
// Tisch-Ablauf: die Bestellung ist eine geldneutrale AVBestellung (TSE-gesichert,
// informative Positionen, aber UMS_BRUTTO=0.00 und kein Beitrag zu USt, Zahlart
// oder Kassenbestand), die Zahlung der einzige umsatzwirksame Beleg. Beide tragen
// denselben Abrechnungskreis und je eine eigene TSE-Transaktion.
func TestMapTischablaufTrennt(t *testing.T) {
	snapshot := testSnapshot()
	snapshot.Tischnamen = map[int]string{42: "Tisch 42"}

	archive, err := Map(snapshot, []event.Event{bestellungEvent(t), zahlungEvent(t)}, tischablaufSignaturen(t))
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}

	const erstellung = "2026-06-16T14:30:00Z"

	wantRecords := map[string][][]string{
		// Beide Vorgänge sind ihrem Tisch zugeordnet und je TSE-gesichert.
		"allocation_groups.csv": {
			{testSerial, erstellung, "3", bestellungBonID, "Tisch 42"},
			{testSerial, erstellung, "3", zahlungBonID, "Tisch 42"},
		},
		"transactions_tse.csv": {
			{testSerial, erstellung, "3", bestellungBonID, "1", "4710", "2026-06-16T11:00:00.000Z", "2026-06-16T11:00:01.000Z", "Bestellung-V1", "11", "BESTELLSIG==", "", ""},
			{testSerial, erstellung, "3", zahlungBonID, "1", "4711", "2026-06-16T12:00:00.000Z", "2026-06-16T12:00:01.000Z", "Kassenbeleg-V1", "12", "ZAHLSIG==", "", ""},
		},
		// Nur die Zahlung ist geldwirksam: die AVBestellung trägt keine Zahlart bei.
		"datapayment.csv": {
			{testSerial, erstellung, "3", zahlungBonID, "Bar", "Bar", "EUR", "4,50", "4,50"},
		},
		// Umsatz entsteht genau einmal (bei der Zahlung), keine Verdopplung.
		"businesscases.csv": {
			{testSerial, erstellung, "3", "Umsatz", "", "0", "1", "4,50", "3,78", "0,72"},
		},
		"payment.csv": {
			{testSerial, erstellung, "3", "Bar", "Bar", "4,50"},
		},
	}

	for file, want := range wantRecords {
		table := tableByFile(t, archive, file)
		if !reflect.DeepEqual(table.Records, want) {
			t.Errorf("%s records =\n%#v\nwant\n%#v", file, table.Records, want)
		}
	}

	// Die geldneutrale AVBestellung trägt nichts zu transactions_vat/lines_vat bei.
	for _, file := range []string{"transactions_vat.csv", "lines_vat.csv"} {
		table := tableByFile(t, archive, file)
		for _, rec := range table.Records {
			if rec[3] == bestellungBonID {
				t.Errorf("%s enthält Zeile für die AVBestellung: %v", file, rec)
			}
		}
	}

	// lines.csv hält die Bestellpositionen informativ (mit Preis); die geldneutrale
	// AVBestellung trägt keinen GV_TYP — sie ist kein Umsatz-Geschäftsvorfall.
	lines := tableByFile(t, archive, "lines.csv")
	if got := field(t, lines, 0, "GV_TYP"); got != "" {
		t.Errorf("bestellung GV_TYP = %q, want leer (geldneutrale AVBestellung)", got)
	}
	if got := field(t, lines, 0, "STK_BR"); got != "4,50" {
		t.Errorf("bestellung STK_BR = %q, want 4.50 (informativer Preis)", got)
	}
	if got := field(t, lines, 1, "GV_TYP"); got != "Umsatz" {
		t.Errorf("zahlung GV_TYP = %q, want Umsatz", got)
	}

	// BON_TYP trennt offene Bestellung und Zahlungsbeleg; die AVBestellung trägt
	// UMS_BRUTTO=0.00, der Umsatz erscheint erst bei der Zahlung.
	transactions := tableByFile(t, archive, "transactions.csv")
	if got := field(t, transactions, 0, "BON_TYP"); got != "AVBestellung" {
		t.Errorf("bestellung BON_TYP = %q, want AVBestellung", got)
	}
	if got := field(t, transactions, 0, "UMS_BRUTTO"); got != "0,00" {
		t.Errorf("bestellung UMS_BRUTTO = %q, want 0.00", got)
	}
	if got := field(t, transactions, 1, "BON_TYP"); got != "Beleg" {
		t.Errorf("zahlung BON_TYP = %q, want Beleg", got)
	}
	if got := field(t, transactions, 1, "UMS_BRUTTO"); got != "4,50" {
		t.Errorf("zahlung UMS_BRUTTO = %q, want 4.50", got)
	}

	// Nur die Barzahlung fließt in den Kassenbestand, nicht die AVBestellung.
	closing := tableByFile(t, archive, "cashpointclosing.csv")
	if got := field(t, closing, 0, "Z_SE_BARZAHLUNGEN"); got != "4,50" {
		t.Errorf("Z_SE_BARZAHLUNGEN = %q, want 4.50 (nur Zahlung, nicht AVBestellung)", got)
	}
}

// TestAbrechnungskreisFallback synthetisiert "Tisch N", wenn der Tisch nicht
// mehr in den Stammdaten steht (z. B. gelöscht).
func TestAbrechnungskreisFallback(t *testing.T) {
	snapshot := testSnapshot()
	snapshot.Tischnamen = nil // kein Tischname bekannt

	archive, err := Map(snapshot, []event.Event{barverkaufEvent(t)}, barverkaufSignaturen(t))
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}

	groups := tableByFile(t, archive, "allocation_groups.csv")
	if got := field(t, groups, 0, "ABRECHNUNGSKREIS"); got != "Tisch 42" {
		t.Errorf("ABRECHNUNGSKREIS = %q, want Tisch 42", got)
	}
}

const (
	stornoBonID              = "33333333-3333-3333-3333-333333333333"
	korrekturBonID           = "33333333-3333-3333-3333-3333333333cc"
	kombiBonID               = "44444444-4444-4444-4444-444444444444"
	direktverkaufBonID       = "55555555-5555-5555-5555-555555555555"
	direktverkaufStornoBonID = "66666666-6666-6666-6666-666666666666"
	nachsigniertSignedBonID  = "77777777-7777-7777-7777-aaaaaaaaaaaa"
	nachsigniertOutageBonID  = "88888888-8888-8888-8888-bbbbbbbbbbbb"
)

// warenruecknahmeEvent nimmt die zuvor bezahlte Bier-Position (zahlungEvent, ZahlungID
// zahlungBonID) kassenwirksam zurück: negativer Umsatz mit Bar-Rückgabe, am selben
// Tisch 42.
func warenruecknahmeEvent(t *testing.T) event.Event {
	t.Helper()

	data := kasse.StornierungErteiltV1Data{
		StornierungID:          stornoBonID,
		ZahlungID:              zahlungBonID,
		GesamtStornierungCents: 450,
		Kommentar:              "Reklamation",
		Positionen: []kasse.PositionEventData{
			{PositionID: "p1", VarianteID: 101, ProduktName: "Bier", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 450, Menge: 1},
		},
	}

	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal warenruecknahme data: %v", err)
	}

	return event.Event{
		ID:       3,
		UserID:   7,
		UserName: "anna",
		Type:     string(kasse.EventTypeStornierungErteiltV1),
		Time:     time.Date(2026, 6, 16, 13, 0, 0, 0, time.UTC),
		Subject:  kasse.TischSessionSubject(3, 42),
		Version:  3,
		Data:     raw,
	}
}

// korrekturEvent storniert die noch unbezahlte Bier-Position der bestellungEvent-
// Bestellung (gleiche PositionID "p1") geldneutral.
func korrekturEvent(t *testing.T) event.Event {
	t.Helper()

	data := kasse.BestellungKorrigiertV1Data{
		KorrekturID: korrekturBonID,
		GesamtCents: 450,
		Kommentar:   "Versehentlich bestellt",
		Positionen: []kasse.PositionEventData{
			{PositionID: "p1", VarianteID: 101, ProduktName: "Bier", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 450, Menge: 1},
		},
	}

	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal korrektur data: %v", err)
	}

	return event.Event{
		ID:       2,
		UserID:   7,
		UserName: "anna",
		Type:     string(kasse.EventTypeBestellungKorrigiertV1),
		Time:     time.Date(2026, 6, 16, 13, 0, 0, 0, time.UTC),
		Subject:  kasse.TischSessionSubject(3, 42),
		Version:  2,
		Data:     raw,
	}
}

// TestMapKorrekturGeldneutralWithReference belegt das Radierverbot für die
// geldneutrale Korrektur unbezahlter Positionen: sie ist eine AVBestellung mit
// negierter MENGE (kein Vorgangs-Storno, BON_STORNO=0) und verweist per REF_BON_ID
// auf die Ursprungsbestellung. Bestellung und Korrektur sind geldneutral, tragen also
// nichts zu USt, Zahlart und Kassenbestand bei.
func TestMapKorrekturGeldneutralWithReference(t *testing.T) {
	snapshot := testSnapshot()
	snapshot.Tischnamen = map[int]string{42: "Tisch 42"}

	signaturen := map[int]tse.EventSignatur{
		1: {ProcessType: "Bestellung-V1", Signatur: testSignatur(t, 4710, 11, "2026-06-16T11:00:00Z", "2026-06-16T11:00:01Z", "BESTELLSIG==")},
		2: {ProcessType: "Bestellung-V1", Signatur: testSignatur(t, 4712, 13, "2026-06-16T13:00:00Z", "2026-06-16T13:00:01Z", "KORREKTURSIG==")},
	}
	archive, err := Map(snapshot, []event.Event{bestellungEvent(t), korrekturEvent(t)}, signaturen)
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}

	const erstellung = "2026-06-16T14:30:00Z"

	wantRecords := map[string][][]string{
		// Verkettung Korrektur → Ursprungsbestellung (gleiche Sitzung: REF_DATUM/REF_Z_NR identisch).
		"references.csv": {
			{testSerial, erstellung, "3", korrekturBonID, "", "Transaktion", "", erstellung, testSerial, "3", bestellungBonID},
		},
		// Eigene TSE-Signatur der Korrektur.
		"transactions_tse.csv": {
			{testSerial, erstellung, "3", bestellungBonID, "1", "4710", "2026-06-16T11:00:00.000Z", "2026-06-16T11:00:01.000Z", "Bestellung-V1", "11", "BESTELLSIG==", "", ""},
			{testSerial, erstellung, "3", korrekturBonID, "1", "4712", "2026-06-16T13:00:00.000Z", "2026-06-16T13:00:01.000Z", "Bestellung-V1", "13", "KORREKTURSIG==", "", ""},
		},
	}

	for file, want := range wantRecords {
		table := tableByFile(t, archive, file)
		if !reflect.DeepEqual(table.Records, want) {
			t.Errorf("%s records =\n%#v\nwant\n%#v", file, table.Records, want)
		}
	}

	// Geldneutralität: eine reine Bestell-/Korrektur-Sitzung erzeugt keine
	// USt-, Zahlart- oder Geschäftsvorfall-Zeilen.
	for _, file := range []string{"transactions_vat.csv", "lines_vat.csv", "datapayment.csv", "businesscases.csv", "payment.csv"} {
		table := tableByFile(t, archive, file)
		if len(table.Records) != 0 {
			t.Errorf("%s muss leer sein (geldneutrale AVBestellung), hat %d Zeilen", file, len(table.Records))
		}
	}

	// Bonkopf: beide Vorgänge sind geldneutral (UMS_BRUTTO=0.00, BON_STORNO=0).
	transactions := tableByFile(t, archive, "transactions.csv")
	if got := field(t, transactions, 0, "BON_STORNO"); got != "0" {
		t.Errorf("ursprung BON_STORNO = %q, want 0", got)
	}
	if got := field(t, transactions, 0, "UMS_BRUTTO"); got != "0,00" {
		t.Errorf("ursprung UMS_BRUTTO = %q, want 0.00", got)
	}
	if got := field(t, transactions, 1, "BON_STORNO"); got != "0" {
		t.Errorf("korrektur BON_STORNO = %q, want 0", got)
	}
	if got := field(t, transactions, 1, "UMS_BRUTTO"); got != "0,00" {
		t.Errorf("korrektur UMS_BRUTTO = %q, want 0.00", got)
	}

	// Positionsebene: negierte MENGE statt P_STORNO-Flag (DSFinV-K Tz. 4.2.3); der
	// informative Preis bleibt erhalten.
	lines := tableByFile(t, archive, "lines.csv")
	if got := field(t, lines, 1, "MENGE"); got != "-1,000" {
		t.Errorf("korrektur MENGE = %q, want -1.000", got)
	}
	if got := field(t, lines, 1, "P_STORNO"); got != "0" {
		t.Errorf("korrektur P_STORNO = %q, want 0 (negierte MENGE statt Flag)", got)
	}
}

// TestMapWarenruecknahmeNegativeWithZahlungReference belegt die kassenwirksame
// Warenrücknahme bezahlter Positionen: ein negativer Bar-Beleg (BON_TYP Beleg, GV_TYP
// Umsatz, Zahlart Bar, BON_STORNO=0) mit REF_BON_ID auf die Zahlung. Der Bargeldbestand
// gleicht sich gegen die vorausgegangene Zahlung aus (keine Doppelbuchung).
func TestMapWarenruecknahmeNegativeWithZahlungReference(t *testing.T) {
	snapshot := testSnapshot()
	snapshot.Tischnamen = map[int]string{42: "Tisch 42"}

	signaturen := tischablaufSignaturen(t)
	signaturen[3] = tse.EventSignatur{ProcessType: "Kassenbeleg-V1", Signatur: testSignatur(t, 4713, 14, "2026-06-16T13:00:00Z", "2026-06-16T13:00:01Z", "STORNOSIG==")}
	archive, err := Map(snapshot, []event.Event{bestellungEvent(t), zahlungEvent(t), warenruecknahmeEvent(t)}, signaturen)
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}

	const erstellung = "2026-06-16T14:30:00Z"

	wantRecords := map[string][][]string{
		// Verkettung Warenrücknahme → Zahlung (nicht die Bestellung).
		"references.csv": {
			{testSerial, erstellung, "3", stornoBonID, "", "Transaktion", "", erstellung, testSerial, "3", zahlungBonID},
		},
		// Negative Bar-Rückgabe, gegenläufig zur Zahlung.
		"datapayment.csv": {
			{testSerial, erstellung, "3", zahlungBonID, "Bar", "Bar", "EUR", "4,50", "4,50"},
			{testSerial, erstellung, "3", stornoBonID, "Bar", "Bar", "EUR", "-4,50", "-4,50"},
		},
	}

	for file, want := range wantRecords {
		table := tableByFile(t, archive, file)
		if !reflect.DeepEqual(table.Records, want) {
			t.Errorf("%s records =\n%#v\nwant\n%#v", file, table.Records, want)
		}
	}

	// Bonkopf: die Warenrücknahme ist ein Beleg mit negativem Umsatz, ohne
	// Vorgangs-Storno-Kennzeichen.
	transactions := tableByFile(t, archive, "transactions.csv")
	if got := field(t, transactions, 2, "BON_TYP"); got != "Beleg" {
		t.Errorf("warenruecknahme BON_TYP = %q, want Beleg", got)
	}
	if got := field(t, transactions, 2, "BON_STORNO"); got != "0" {
		t.Errorf("warenruecknahme BON_STORNO = %q, want 0", got)
	}
	if got := field(t, transactions, 2, "UMS_BRUTTO"); got != "-4,50" {
		t.Errorf("warenruecknahme UMS_BRUTTO = %q, want -4.50", got)
	}

	// Der Bargeldbestand gleicht sich aus: Zahlung +4.50, Warenrücknahme −4.50.
	closing := tableByFile(t, archive, "cashpointclosing.csv")
	if got := field(t, closing, 0, "Z_SE_BARZAHLUNGEN"); got != "0,00" {
		t.Errorf("Z_SE_BARZAHLUNGEN = %q, want 0.00 (Zahlung und Rückgabe heben sich auf)", got)
	}

	// Umsatz je Steuersatz saldiert sich auf 0 (Zahlung +, Warenrücknahme −).
	businesscases := tableByFile(t, archive, "businesscases.csv")
	if got := field(t, businesscases, 0, "Z_UMS_BRUTTO"); got != "0,00" {
		t.Errorf("businesscases Z_UMS_BRUTTO = %q, want 0.00", got)
	}
}

const umbuchungBonID = "99999999-9999-4999-8999-999999999999"

// umbuchungEventPaar baut das verknüpfte, geldneutrale Umbuchungs-Paar: Abgang auf
// Tisch 42 (BON_ID = UmbuchungID) und Zugang auf Tisch 7. Beide tragen dieselbe
// UmbuchungID.
func umbuchungEventPaar(t *testing.T) (event.Event, event.Event) {
	t.Helper()

	bauen := func(id int, tischID int, posID string, kommentar string) event.Event {
		data := kasse.BestellungUmgebuchtV1Data{
			UmbuchungID:  umbuchungBonID,
			QuellTischID: 42,
			ZielTischID:  7,
			GesamtCents:  450,
			Kommentar:    kommentar,
			Positionen: []kasse.PositionEventData{
				{PositionID: posID, VarianteID: 101, ProduktName: "Bier", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 450, Menge: 1},
			},
		}
		raw, err := json.Marshal(data)
		if err != nil {
			t.Fatalf("marshal umbuchung data: %v", err)
		}
		return event.Event{
			ID:       id,
			UserID:   7,
			UserName: "anna",
			Type:     string(kasse.EventTypeBestellungUmgebuchtV1),
			Time:     time.Date(2026, 6, 16, 13, 30, 0, 0, time.UTC),
			Subject:  kasse.TischSessionSubject(3, tischID),
			Version:  2,
			Data:     raw,
		}
	}

	abgang := bauen(2, 42, "p1", "Umbuchung auf Tisch Tisch 7")
	zugang := bauen(3, 7, "p1z", "Umbuchung von Tisch Tisch 42")
	return abgang, zugang
}

// TestMapUmbuchungGeldneutralMitReferenz belegt: eine Umbuchung erzeugt je Seite eine
// geldneutrale AVBestellung (kein Umsatz, keine Zahlart, keine Kassenbestandswirkung);
// der Zugang verweist in references.csv per REF_BON_ID auf den Abgang (gemeinsame
// UmbuchungID).
func TestMapUmbuchungGeldneutralMitReferenz(t *testing.T) {
	snapshot := testSnapshot()
	snapshot.Tischnamen = map[int]string{42: "Tisch 42", 7: "Tisch 7"}

	abgang, zugang := umbuchungEventPaar(t)
	signaturen := map[int]tse.EventSignatur{
		2: {ProcessType: "Bestellung-V1", Signatur: testSignatur(t, 4722, 22, "2026-06-16T13:30:00Z", "2026-06-16T13:30:01Z", "UMBUCHABSIG==")},
		3: {ProcessType: "Bestellung-V1", Signatur: testSignatur(t, 4723, 23, "2026-06-16T13:30:00Z", "2026-06-16T13:30:01Z", "UMBUCHZUSIG==")},
	}
	archive, err := Map(snapshot, []event.Event{abgang, zugang}, signaturen)
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}

	const erstellung = "2026-06-16T14:30:00Z"
	const zugangBonID = "umbuchung-3"

	wantRecords := map[string][][]string{
		// Verkettung Zugang → Abgang (gemeinsame Sitzung).
		"references.csv": {
			{testSerial, erstellung, "3", zugangBonID, "", "Transaktion", "", erstellung, testSerial, "3", umbuchungBonID},
		},
	}
	for file, want := range wantRecords {
		table := tableByFile(t, archive, file)
		if !reflect.DeepEqual(table.Records, want) {
			t.Errorf("%s records =\n%#v\nwant\n%#v", file, table.Records, want)
		}
	}

	// Geldneutralität: keine USt-, Zahlart- oder Geschäftsvorfall-Zeilen.
	for _, file := range []string{"transactions_vat.csv", "lines_vat.csv", "datapayment.csv", "businesscases.csv", "payment.csv"} {
		table := tableByFile(t, archive, file)
		if len(table.Records) != 0 {
			t.Errorf("%s muss leer sein (geldneutrale Umbuchung), hat %d Zeilen", file, len(table.Records))
		}
	}

	// Beide Seiten sind AVBestellungen ohne Storno-Kennzeichen und ohne Umsatz.
	transactions := tableByFile(t, archive, "transactions.csv")
	for row := 0; row < 2; row++ {
		if got := field(t, transactions, row, "BON_TYP"); got != "AVBestellung" {
			t.Errorf("umbuchung[%d] BON_TYP = %q, want AVBestellung", row, got)
		}
		if got := field(t, transactions, row, "BON_STORNO"); got != "0" {
			t.Errorf("umbuchung[%d] BON_STORNO = %q, want 0", row, got)
		}
		if got := field(t, transactions, row, "UMS_BRUTTO"); got != "0,00" {
			t.Errorf("umbuchung[%d] UMS_BRUTTO = %q, want 0.00", row, got)
		}
	}

	// Kein Bargeld bewegt sich durch die Umbuchung.
	closing := tableByFile(t, archive, "cashpointclosing.csv")
	if got := field(t, closing, 0, "Z_SE_BARZAHLUNGEN"); got != "0,00" {
		t.Errorf("Z_SE_BARZAHLUNGEN = %q, want 0.00 (Umbuchung ist geldneutral)", got)
	}
}

// kombiZahlungEvent kassiert ein Kombi-Menü zum Pauschalpreis 5,01 €.
func kombiZahlungEvent(t *testing.T) event.Event {
	t.Helper()

	data := kasse.ZahlungKassiertV1Data{
		ZahlungID:          kombiBonID,
		GesamtZahlungCents: 501,
		Positionen: []kasse.PositionEventData{
			{PositionID: "p1", VarianteID: 303, ProduktName: "Menü", VarianteName: "", Kategorie: "essen", Steuersatz: "kombi", EinzelpreisCents: 501, Menge: 1},
		},
	}

	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal kombi data: %v", err)
	}

	return event.Event{
		ID:       1,
		UserID:   7,
		UserName: "anna",
		Type:     string(kasse.EventTypeZahlungKassiertV1),
		Time:     time.Date(2026, 6, 16, 12, 30, 0, 0, time.UTC),
		Subject:  kasse.TischSessionSubject(3, 42),
		Version:  1,
		Data:     raw,
	}
}

// TestMapKombiSteuerSplit belegt die Entfaltung einer kombi-Position in 70 % zu
// 7 % und 30 % zu 19 %. lines_vat folgt der Aufteilen-Reihenfolge (ermäßigt,
// regel), transactions_vat der Steuermatrix-Reihenfolge (regel, ermäßigt).
func TestMapKombiSteuerSplit(t *testing.T) {
	signaturen := map[int]tse.EventSignatur{
		1: {ProcessType: "Kassenbeleg-V1", Signatur: testSignatur(t, 4800, 20, "2026-06-16T12:30:00Z", "2026-06-16T12:30:01Z", "KOMBISIG==")},
	}
	archive, err := Map(testSnapshot(), []event.Event{kombiZahlungEvent(t)}, signaturen)
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}

	const erstellung = "2026-06-16T14:30:00Z"

	wantRecords := map[string][][]string{
		"lines_vat.csv": {
			{testSerial, erstellung, "3", kombiBonID, "1", "2", "3,51", "3,28", "0,23"},
			{testSerial, erstellung, "3", kombiBonID, "1", "1", "1,50", "1,26", "0,24"},
		},
		"transactions_vat.csv": {
			{testSerial, erstellung, "3", kombiBonID, "1", "1,50", "1,26", "0,24"},
			{testSerial, erstellung, "3", kombiBonID, "2", "3,51", "3,28", "0,23"},
		},
		"vat.csv": {
			{testSerial, erstellung, "3", "1", "19,00", "Allgemeiner Steuersatz"},
			{testSerial, erstellung, "3", "2", "7,00", "Ermäßigter Steuersatz"},
			{testSerial, erstellung, "3", "3", "10,70", "Durchschnittsatz (§ 24 Abs. 1 Nr. 3 UStG)"},
			{testSerial, erstellung, "3", "4", "5,50", "Durchschnittsatz (§ 24 Abs. 1 Nr. 1 UStG)"},
			{testSerial, erstellung, "3", "5", "0,00", "Nicht Steuerbar"},
			{testSerial, erstellung, "3", "6", "0,00", "Umsatzsteuerfrei"},
			{testSerial, erstellung, "3", "7", "0,00", "UmsatzsteuerNichtErmittelbar"},
		},
	}

	for file, want := range wantRecords {
		table := tableByFile(t, archive, file)
		if !reflect.DeepEqual(table.Records, want) {
			t.Errorf("%s records =\n%#v\nwant\n%#v", file, table.Records, want)
		}
	}
}

// direktverkaufEvent verkauft ein Bier direkt an der Theke (eigener Stream, kein Tisch).
func direktverkaufEvent(t *testing.T) event.Event {
	t.Helper()

	data := kasse.DirektverkaufGetaetigtV1Data{
		VerkaufID:         direktverkaufBonID,
		GesamtbetragCents: 450,
		Positionen: []kasse.PositionEventData{
			{PositionID: "d1", VarianteID: 101, ProduktName: "Bier", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 450, Menge: 1},
		},
	}

	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal direktverkauf data: %v", err)
	}

	return event.Event{
		ID:       1,
		UserID:   7,
		UserName: "anna",
		Type:     string(kasse.EventTypeDirektverkaufGetaetigtV1),
		Time:     time.Date(2026, 6, 16, 14, 0, 0, 0, time.UTC),
		Subject:  kasse.DirektverkaufSubject(3, direktverkaufBonID),
		Version:  1,
		Data:     raw,
	}
}

// direktverkaufStornoEvent storniert den Direktverkauf, eigener Stream mit Referenz.
func direktverkaufStornoEvent(t *testing.T) event.Event {
	t.Helper()

	data := kasse.DirektverkaufStorniertV1Data{
		StornierungID:          direktverkaufStornoBonID,
		VerkaufID:              direktverkaufBonID,
		GesamtStornierungCents: 450,
		Kommentar:              "Falsch eingegeben",
		Positionen: []kasse.PositionEventData{
			{PositionID: "d1", VarianteID: 101, ProduktName: "Bier", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 450, Menge: 1},
		},
	}

	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal direktverkauf storno data: %v", err)
	}

	return event.Event{
		ID:       2,
		UserID:   7,
		UserName: "anna",
		Type:     string(kasse.EventTypeDirektverkaufStorniertV1),
		Time:     time.Date(2026, 6, 16, 14, 5, 0, 0, time.UTC),
		Subject:  kasse.DirektverkaufSubject(3, direktverkaufBonID),
		Version:  1,
		Data:     raw,
	}
}

// TestMapDirektverkaufUndStorno belegt: Direktverkauf und sein Storno sind eigene
// Barbelege ohne Abrechnungskreis; der Storno verweist per REF_BON_ID auf den
// Ursprungsverkauf und kehrt die Vorzeichen um.
func TestMapDirektverkaufUndStorno(t *testing.T) {
	signaturen := map[int]tse.EventSignatur{
		1: {ProcessType: "Kassenbeleg-V1", Signatur: testSignatur(t, 5000, 30, "2026-06-16T14:00:00Z", "2026-06-16T14:00:01Z", "DVSIG==")},
		2: {ProcessType: "Kassenbeleg-V1", Signatur: testSignatur(t, 5001, 31, "2026-06-16T14:05:00Z", "2026-06-16T14:05:01Z", "DVSTORNOSIG==")},
	}
	archive, err := Map(testSnapshot(), []event.Event{direktverkaufEvent(t), direktverkaufStornoEvent(t)}, signaturen)
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}

	const erstellung = "2026-06-16T14:30:00Z"

	wantRecords := map[string][][]string{
		"references.csv": {
			{testSerial, erstellung, "3", direktverkaufStornoBonID, "", "Transaktion", "", erstellung, testSerial, "3", direktverkaufBonID},
		},
		"datapayment.csv": {
			{testSerial, erstellung, "3", direktverkaufBonID, "Bar", "Bar", "EUR", "4,50", "4,50"},
			{testSerial, erstellung, "3", direktverkaufStornoBonID, "Bar", "Bar", "EUR", "-4,50", "-4,50"},
		},
	}

	for file, want := range wantRecords {
		table := tableByFile(t, archive, file)
		if !reflect.DeepEqual(table.Records, want) {
			t.Errorf("%s records =\n%#v\nwant\n%#v", file, table.Records, want)
		}
	}

	// Beide Belege tragen BON_TYP "Beleg"; der Storno ist eine negative
	// Belegdarstellung ohne Vorgangs-Storno-Kennzeichen (BON_STORNO=0).
	transactions := tableByFile(t, archive, "transactions.csv")
	if got := field(t, transactions, 0, "BON_TYP"); got != "Beleg" {
		t.Errorf("direktverkauf BON_TYP = %q, want Beleg", got)
	}
	if got := field(t, transactions, 0, "BON_STORNO"); got != "0" {
		t.Errorf("direktverkauf BON_STORNO = %q, want 0", got)
	}
	if got := field(t, transactions, 1, "BON_STORNO"); got != "0" {
		t.Errorf("direktverkauf-storno BON_STORNO = %q, want 0", got)
	}

	// Direktverkäufe tragen keinen Abrechnungskreis (kein Tischbezug).
	groups := tableByFile(t, archive, "allocation_groups.csv")
	if len(groups.Records) != 0 {
		t.Errorf("allocation_groups must be empty for Direktverkauf, got %d rows", len(groups.Records))
	}
}

// eroeffnetEvent eröffnet die Sitzung mit einem Anfangsbestand von 100,00 €.
func eroeffnetEvent(t *testing.T, betragCents int) event.Event {
	t.Helper()

	data := kasse.KassensitzungEroeffnetV1Data{
		Datum:        "2026-06-16",
		Bezeichnung:  "Sommerfest",
		BetragCents:  betragCents,
		EroeffnetVon: 7,
	}
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal eroeffnet data: %v", err)
	}

	return event.Event{
		ID: 10, UserID: 7, UserName: "anna",
		Type:    string(kasse.EventTypeKassensitzungEroeffnetV1),
		Time:    time.Date(2026, 6, 16, 10, 0, 0, 0, time.UTC),
		Subject: kasse.KassensitzungSubject(3), Version: 1, Data: raw,
	}
}

// geldtransitEvent entnimmt 50,00 € aus der Kasse (z. B. zum Tresor).
func geldtransitEvent(t *testing.T) event.Event {
	t.Helper()

	data := kasse.GeldtransitGebuchtV1Data{
		GeldtransitID: "77777777-7777-4777-8777-777777777777",
		Richtung:      "entnahme",
		BetragCents:   5000,
		Kommentar:     "Abschöpfung Tresor",
		GebuchtVon:    7,
	}
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal geldtransit data: %v", err)
	}

	return event.Event{
		ID: 11, UserID: 7, UserName: "anna",
		Type:    string(kasse.EventTypeGeldtransitGebuchtV1),
		Time:    time.Date(2026, 6, 16, 15, 0, 0, 0, time.UTC),
		Subject: kasse.KassensitzungSubject(3), Version: 1, Data: raw,
	}
}

// differenzEvent bucht einen Kassenfehlbetrag von 1,00 € (Soll − Ist = +100).
func differenzEvent(t *testing.T) event.Event {
	t.Helper()

	data := kasse.DifferenzSollIstGebuchtV1Data{
		BetragCents: 100,
		GebuchtVon:  7,
	}
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal differenz data: %v", err)
	}

	return event.Event{
		ID: 13, UserID: 7, UserName: "anna",
		Type:    string(kasse.EventTypeDifferenzSollIstGebuchtV1),
		Time:    time.Date(2026, 6, 16, 16, 0, 0, 0, time.UTC),
		Subject: kasse.KassensitzungSubject(3), Version: 1, Data: raw,
	}
}

// tagesabschlussEvent schließt die Sitzung ab (Z-Bon). Als letztes Event trägt
// er die höchste Event-ID und wird damit der letzte Beleg der Sitzung.
func tagesabschlussEvent(t *testing.T) event.Event {
	t.Helper()

	data := kasse.TagesabschlussErstelltV1Data{
		ZNr:               3,
		ZeitraumVon:       time.Date(2026, 6, 16, 10, 0, 0, 0, time.UTC),
		ZeitraumBis:       time.Date(2026, 6, 16, 18, 0, 0, 0, time.UTC),
		UmsatzGesamtCents: 450,
		GeldtransitCents:  -5000,
		ErstelltVon:       7,
	}
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal tagesabschluss data: %v", err)
	}

	return event.Event{
		ID: 20, UserID: 7, UserName: "anna",
		Type:    string(kasse.EventTypeTagesabschlussErstelltV1),
		Time:    time.Date(2026, 6, 16, 18, 0, 0, 0, time.UTC),
		Subject: kasse.KassensitzungSubject(3), Version: 1, Data: raw,
	}
}

// TestMapKassenabschlussGemischteSitzung belegt das Kassenabschlussmodul über eine
// gemischte Sitzung: Anfangsbestand, eine geldneutrale Bestellung plus ihre
// Zahlung (Umsatz), ein Direktverkauf (Umsatz) sowie Geldtransit und
// Kassendifferenz. Der Umsatz entsteht nur bei den Zahlungen (Revenue-at-payment),
// nicht bei der Bestellung. businesscases.csv und payment.csv lassen sich gegen die
// Einzelbons abgleichen; cash_per_currency.csv weist den EUR-Bestand aus.
func TestMapKassenabschlussGemischteSitzung(t *testing.T) {
	snapshot := testSnapshot()
	snapshot.Tischnamen = map[int]string{42: "Tisch 42"}

	// Der Direktverkauf erhält eine eigene Event-ID (Fixture-Standard kollidiert
	// mit der Bestellung), damit die Signaturen-Map eindeutig bleibt.
	direktverkauf := direktverkaufEvent(t)
	direktverkauf.ID = 4

	events := []event.Event{
		eroeffnetEvent(t, 10000), // Anfangsbestand 100,00 € (ID 10)
		bestellungEvent(t),       // AVBestellung 4,50 € (Bier, 19 %, geldneutral, ID 1)
		zahlungEvent(t),          // Umsatz 4,50 € (Bier, 19 %, ID 2)
		direktverkauf,            // Umsatz 4,50 € (Bier, 19 %, ID 4)
		geldtransitEvent(t),      // Entnahme −50,00 € (ID 11)
		differenzEvent(t),        // Fehlbetrag −1,00 € (ID 13)
	}

	signaturen := tischablaufSignaturen(t)
	signaturen[10] = tse.EventSignatur{ProcessType: "Kassenbeleg-V1", Signatur: testSignatur(t, 6100, 44, "2026-06-16T10:00:00Z", "2026-06-16T10:00:01Z", "EROEFFNUNGSIG==")}
	signaturen[4] = tse.EventSignatur{ProcessType: "Kassenbeleg-V1", Signatur: testSignatur(t, 5000, 30, "2026-06-16T14:00:00Z", "2026-06-16T14:00:01Z", "DVSIG==")}
	signaturen[11] = tse.EventSignatur{ProcessType: "Kassenbeleg-V1", Signatur: testSignatur(t, 6000, 40, "2026-06-16T15:00:00Z", "2026-06-16T15:00:01Z", "GTSIG==")}
	signaturen[13] = tse.EventSignatur{ProcessType: "Kassenbeleg-V1", Signatur: testSignatur(t, 6002, 42, "2026-06-16T16:00:00Z", "2026-06-16T16:00:01Z", "DIFFSIG==")}

	archive, err := Map(snapshot, events, signaturen)
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}

	const erstellung = "2026-06-16T14:30:00Z"

	wantRecords := map[string][][]string{
		// Je Geschäftsvorfalltyp und Steuersatz: Umsatz (2 × Bier 19 %, nur die
		// Zahlungen) vor den nicht-steuerbaren Bargeldbewegungen (Schlüssel 5). Die
		// geldneutrale AVBestellung erzeugt keine eigene Geschäftsvorfall-Zeile.
		"businesscases.csv": {
			{testSerial, erstellung, "3", "Umsatz", "", "0", "1", "9,00", "7,56", "1,44"},
			{testSerial, erstellung, "3", "Anfangsbestand", "", "0", "5", "100,00", "100,00", "0,00"},
			{testSerial, erstellung, "3", "Geldtransit", "", "0", "5", "-50,00", "-50,00", "0,00"},
			{testSerial, erstellung, "3", "DifferenzSollIst", "", "0", "5", "-1,00", "-1,00", "0,00"},
		},
		// Bar = 100 + 4,50 + 4,50 − 50 − 1 = 58,00; die AVBestellung trägt nichts bei.
		"payment.csv": {
			{testSerial, erstellung, "3", "Bar", "Bar", "58,00"},
		},
		"cash_per_currency.csv": {
			{testSerial, erstellung, "3", "EUR", "58,00"},
		},
	}

	for file, want := range wantRecords {
		table := tableByFile(t, archive, file)
		if !reflect.DeepEqual(table.Records, want) {
			t.Errorf("%s records =\n%#v\nwant\n%#v", file, table.Records, want)
		}
	}

	// Abgleich der Tagessumme gegen die Einzelbons: die Summe der Bonkopf-Brutto
	// (Einzelaufzeichnung) muss mit den Aggregaten je GV-Typ und je Zahlart
	// übereinstimmen (DSFinV-K-Konsistenz Bonmodul ↔ Kassenabschlussmodul).
	bonkopfSumme := summe(t, archive, "transactions.csv", "UMS_BRUTTO")
	gvSumme := summe(t, archive, "businesscases.csv", "Z_UMS_BRUTTO")
	zahlartSumme := summe(t, archive, "payment.csv", "Z_ZAHLART_BETRAG")
	if bonkopfSumme != gvSumme {
		t.Errorf("Summe Bonkopf (%d) ≠ Summe businesscases (%d)", bonkopfSumme, gvSumme)
	}
	if bonkopfSumme != zahlartSumme {
		t.Errorf("Summe Bonkopf (%d) ≠ Summe payment (%d)", bonkopfSumme, zahlartSumme)
	}

	// Alle drei signierten Bargeldbewegungen tragen ihre transactions_tse-Zeile —
	// auch der Anfangsbestand, dessen Eröffnungs-Event im Outbox-Modell
	// signaturpflichtig ist (Bareinlage).
	tse := tableByFile(t, archive, "transactions_tse.csv")
	for _, sig := range []string{"EROEFFNUNGSIG==", "GTSIG==", "DIFFSIG=="} {
		if !hatSignatur(tse, sig) {
			t.Errorf("transactions_tse fehlt Signatur %q", sig)
		}
	}

	// Verprobung ohne Waisen: Jede transactions_tse-Zeile gehört zu einem
	// Bonkopf, und jeder Bonkopf hat genau eine TSE-Zeile (alle Vorgänge der
	// Sitzung sind signaturpflichtig).
	transactions := tableByFile(t, archive, "transactions.csv")
	bonIDs := map[string]bool{}
	for row := range transactions.Records {
		bonIDs[field(t, transactions, row, "BON_ID")] = true
	}
	for row := range tse.Records {
		if bonID := field(t, tse, row, "BON_ID"); !bonIDs[bonID] {
			t.Errorf("transactions_tse-Zeile ohne Bonkopf: BON_ID %q", bonID)
		}
	}
	if len(tse.Records) != len(transactions.Records) {
		t.Errorf("transactions_tse-Zeilen (%d) ≠ Bonkopf-Vorgänge (%d)", len(tse.Records), len(transactions.Records))
	}
}

// TestMapTagesabschlussSigniertErscheintAlsAVSonstigeBon belegt Finding 5: der
// TSE-signierte Tagesabschluss erscheint als geldneutraler AVSonstige-Bon mit
// eigener transactions_tse.csv-Zeile. Weil er geldneutral ist, bleiben die
// Aggregate (businesscases, payment, cash_per_currency) und die Bar-Summen des
// Kassenabschlusses unverändert; als letzter Vorgang der Sitzung wird er
// zusätzlich Z_ENDE_ID, ohne die BON_NR der übrigen Belege zu verschieben.
func TestMapTagesabschlussSigniertErscheintAlsAVSonstigeBon(t *testing.T) {
	const tagesabschlussBonID = "tagesabschluss-20"

	// Dieselbe Sitzung mit und ohne Abschluss: die Aggregate müssen identisch sein.
	ohneAbschluss := []event.Event{
		eroeffnetEvent(t, 10000), // Anfangsbestand 100,00 € (ID 10)
		zahlungEvent(t),          // Umsatz 4,50 € (Bier 19 %, ID 2)
		geldtransitEvent(t),      // Entnahme −50,00 € (ID 11)
		differenzEvent(t),        // Fehlbetrag −1,00 € (ID 13)
	}
	mitAbschluss := append(append([]event.Event{}, ohneAbschluss...), tagesabschlussEvent(t))

	signaturen := map[int]tse.EventSignatur{
		10: {ProcessType: "Kassenbeleg-V1", Signatur: testSignatur(t, 6100, 44, "2026-06-16T10:00:00Z", "2026-06-16T10:00:01Z", "EROEFFNUNGSIG==")},
		2:  {ProcessType: "Kassenbeleg-V1", Signatur: testSignatur(t, 4711, 12, "2026-06-16T12:00:00Z", "2026-06-16T12:00:01Z", "ZAHLSIG==")},
		11: {ProcessType: "Kassenbeleg-V1", Signatur: testSignatur(t, 6000, 40, "2026-06-16T15:00:00Z", "2026-06-16T15:00:01Z", "GTSIG==")},
		13: {ProcessType: "Kassenbeleg-V1", Signatur: testSignatur(t, 6002, 42, "2026-06-16T16:00:00Z", "2026-06-16T16:00:01Z", "DIFFSIG==")},
		20: {ProcessType: "SonstigerVorgang", Signatur: testSignatur(t, 7777, 88, "2026-06-16T18:00:00Z", "2026-06-16T18:00:02Z", "TAGSIG==")},
	}

	archivOhne, err := Map(testSnapshot(), ohneAbschluss, signaturen)
	if err != nil {
		t.Fatalf("Map(ohne) error = %v", err)
	}
	archivMit, err := Map(testSnapshot(), mitAbschluss, signaturen)
	if err != nil {
		t.Fatalf("Map(mit) error = %v", err)
	}

	// Der 0-Bon lässt die Aggregationsdateien unverändert.
	for _, file := range []string{"businesscases.csv", "payment.csv", "cash_per_currency.csv"} {
		ohne := tableByFile(t, archivOhne, file)
		mit := tableByFile(t, archivMit, file)
		if !reflect.DeepEqual(mit.Records, ohne.Records) {
			t.Errorf("%s durch Tagesabschluss verändert:\n mit  %#v\n ohne %#v", file, mit.Records, ohne.Records)
		}
	}

	// cashpointclosing: Bar-Summen und Z_START_ID unverändert, Z_ENDE_ID zeigt auf
	// den Abschluss-Bon (letzte BON_ID im Abschluss).
	closingOhne := tableByFile(t, archivOhne, "cashpointclosing.csv")
	closingMit := tableByFile(t, archivMit, "cashpointclosing.csv")
	for _, spalte := range []string{"Z_START_ID", "Z_SE_ZAHLUNGEN", "Z_SE_BARZAHLUNGEN"} {
		if got, want := field(t, closingMit, 0, spalte), field(t, closingOhne, 0, spalte); got != want {
			t.Errorf("cashpointclosing %s = %q, want %q (unverändert)", spalte, got, want)
		}
	}
	if got := field(t, closingMit, 0, "Z_ENDE_ID"); got != tagesabschlussBonID {
		t.Errorf("Z_ENDE_ID = %q, want %q", got, tagesabschlussBonID)
	}

	// Genau ein AVSonstige-Bon als letzter Bonkopf, geldneutral (UMS_BRUTTO 0,00);
	// die übrigen BON_NR bleiben gegenüber der Sitzung ohne Abschluss unverschoben.
	transOhne := tableByFile(t, archivOhne, "transactions.csv")
	transMit := tableByFile(t, archivMit, "transactions.csv")
	if len(transMit.Records) != len(transOhne.Records)+1 {
		t.Fatalf("transactions: %d Zeilen, want %d", len(transMit.Records), len(transOhne.Records)+1)
	}
	for i := range transOhne.Records {
		if got, want := field(t, transMit, i, "BON_NR"), field(t, transOhne, i, "BON_NR"); got != want {
			t.Errorf("BON_NR Zeile %d = %q, want %q (verschoben)", i, got, want)
		}
	}
	letzte := len(transMit.Records) - 1
	checks := map[string]string{
		"BON_ID":     tagesabschlussBonID,
		"BON_NR":     "5",
		"BON_TYP":    "AVSonstige",
		"BON_NAME":   "Tagesabschluss",
		"BON_STORNO": "0",
		"UMS_BRUTTO": "0,00",
		"BON_START":  "2026-06-16T18:00:00Z",
		"BON_ENDE":   "2026-06-16T18:00:00Z",
	}
	for name, want := range checks {
		if got := field(t, transMit, letzte, name); got != want {
			t.Errorf("Abschluss-Bon %s = %q, want %q", name, got, want)
		}
	}
	sonstige := 0
	for i := range transMit.Records {
		if field(t, transMit, i, "BON_TYP") == "AVSonstige" {
			sonstige++
		}
	}
	if sonstige != 1 {
		t.Errorf("AVSonstige-Bons = %d, want 1", sonstige)
	}

	// Der Abschluss-Bon trägt seine eigene TSE-Zeile mit Signatur.
	tseTab := tableByFile(t, archivMit, "transactions_tse.csv")
	row := tseRowByBonID(t, tseTab, tagesabschlussBonID)
	if got := field(t, tseTab, row, "TSE_TA_SIG"); got != "TAGSIG==" {
		t.Errorf("Abschluss TSE_TA_SIG = %q, want TAGSIG==", got)
	}
	if got := field(t, tseTab, row, "TSE_TA_FEHLER"); got != "" {
		t.Errorf("Abschluss TSE_TA_FEHLER = %q, want leer", got)
	}
}

// TestMapTagesabschlussAusfallTraegtFehlerzeile belegt: ein signaturpflichtiger,
// (noch) unsignierter Tagesabschluss (Auftrag offen, fehlgeschlagen oder
// tse_nicht_konfiguriert) fehlt nicht im Export, sondern trägt — wie jeder andere Vorgang —
// eine TSE_TA_FEHLER-Zeile. So gilt die Invariante „jeder Bonkopf hat genau eine
// TSE-Zeile“ auch für den AVSonstige-Bon.
func TestMapTagesabschlussAusfallTraegtFehlerzeile(t *testing.T) {
	const tagesabschlussBonID = "tagesabschluss-20"

	events := []event.Event{
		zahlungEvent(t),        // signierte Zahlung (ID 2)
		tagesabschlussEvent(t), // Abschluss ohne Signatur (Auftrag unerledigt, ID 20)
	}

	signaturen := map[int]tse.EventSignatur{
		2:  {ProcessType: "Kassenbeleg-V1", Signatur: testSignatur(t, 4711, 12, "2026-06-16T12:00:00Z", "2026-06-16T12:00:01Z", "ZAHLSIG==")},
		20: {ProcessType: "SonstigerVorgang"}, // signaturpflichtig, aber unsigniert
	}

	archive, err := Map(testSnapshot(), events, signaturen)
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}

	tse := tableByFile(t, archive, "transactions_tse.csv")
	row := tseRowByBonID(t, tse, tagesabschlussBonID)
	if got := field(t, tse, row, "TSE_TA_FEHLER"); got != tseFehlerAusfall {
		t.Errorf("Abschluss TSE_TA_FEHLER = %q, want %q", got, tseFehlerAusfall)
	}
	if got := field(t, tse, row, "TSE_TA_SIG"); got != "" {
		t.Errorf("Abschluss TSE_TA_SIG = %q, want leer", got)
	}

	// Invariante: jeder Bonkopf-Vorgang hat genau eine TSE-Zeile.
	transactions := tableByFile(t, archive, "transactions.csv")
	if len(tse.Records) != len(transactions.Records) {
		t.Errorf("transactions_tse-Zeilen (%d) ≠ Bonkopf-Vorgänge (%d)", len(tse.Records), len(transactions.Records))
	}
}

// summe addiert eine Cent-Spalte (formatierte Beträge) über alle Zeilen.
func summe(t *testing.T, a Archive, file, spalte string) int {
	t.Helper()
	table := tableByFile(t, a, file)
	total := 0
	for row := range table.Records {
		total += centsAus(t, field(t, table, row, spalte))
	}
	return total
}

// centsAus parst einen DSFinV-K-Dezimalbetrag ("−50,00", Komma laut amtlicher
// index.xml) zurück nach Cent.
func centsAus(t *testing.T, s string) int {
	t.Helper()
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	ganz, frac, ok := strings.Cut(s, ",")
	if !ok {
		t.Fatalf("kein Dezimalbetrag: %q", s)
	}
	euros, err := strconv.Atoi(ganz)
	if err != nil {
		t.Fatalf("ganzer Teil %q: %v", ganz, err)
	}
	cents, err := strconv.Atoi(frac)
	if err != nil {
		t.Fatalf("Nachkommateil %q: %v", frac, err)
	}
	betrag := euros*100 + cents
	if neg {
		return -betrag
	}
	return betrag
}

// hatSignatur prüft, ob eine TSE-Signatur in transactions_tse.csv vorkommt.
func hatSignatur(table Table, sig string) bool {
	for row := range table.Records {
		for i, c := range table.Columns {
			if c.name == "TSE_TA_SIG" && table.Records[row][i] == sig {
				return true
			}
		}
	}
	return false
}

// zahlungAm baut eine Barzahlung mit gegebener Event-ID und BON_ID.
func zahlungAm(t *testing.T, id int, bonID string, ts time.Time) event.Event {
	t.Helper()

	data := kasse.ZahlungKassiertV1Data{
		ZahlungID:          bonID,
		GesamtZahlungCents: 450,
		Positionen: []kasse.PositionEventData{
			{PositionID: "p1", VarianteID: 101, ProduktName: "Bier", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 450, Menge: 1},
		},
	}

	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal zahlung data: %v", err)
	}

	return event.Event{
		ID: id, UserID: 7, UserName: "anna",
		Type:    string(kasse.EventTypeZahlungKassiertV1),
		Time:    ts,
		Subject: kasse.TischSessionSubject(3, 42), Version: 1, Data: raw,
	}
}

// TestMapNachsigniertVorgang belegt den einen Leseweg: Auch ein erst nach einer
// Störung (verspätet) quittierter Auftrag erscheint vollständig in
// transactions_tse.csv — die Signaturspalten des Auftrags sind die einzige
// Quelle, TSE_TA_VORGANGSART kommt aus dem process_type-Snapshot des Auftrags.
func TestMapNachsigniertVorgang(t *testing.T) {
	prompt := zahlungAm(t, 1, nachsigniertSignedBonID, time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC))
	nachsigniert := zahlungAm(t, 2, nachsigniertOutageBonID, time.Date(2026, 6, 16, 13, 0, 0, 0, time.UTC))

	signaturen := map[int]tse.EventSignatur{
		1: {ProcessType: "Kassenbeleg-V1", Signatur: testSignatur(t, 4711, 12, "2026-06-16T12:00:00Z", "2026-06-16T12:00:01Z", "EVENTSIG==")},
		// Verspätete Quittierung: logTime deutlich nach der Event-Zeit.
		2: {ProcessType: "Kassenbeleg-V1", Signatur: testSignatur(t, 9100, 99, "2026-06-16T13:00:05Z", "2026-06-16T13:00:06Z", "BACKFILLSIG==")},
	}

	archive, err := Map(testSnapshot(), []event.Event{prompt, nachsigniert}, signaturen)
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}

	const erstellung = "2026-06-16T14:30:00Z"
	want := [][]string{
		{testSerial, erstellung, "3", nachsigniertSignedBonID, "1", "4711", "2026-06-16T12:00:00.000Z", "2026-06-16T12:00:01.000Z", "Kassenbeleg-V1", "12", "EVENTSIG==", "", ""},
		{testSerial, erstellung, "3", nachsigniertOutageBonID, "1", "9100", "2026-06-16T13:00:05.000Z", "2026-06-16T13:00:06.000Z", "Kassenbeleg-V1", "99", "BACKFILLSIG==", "", ""},
	}
	tse := tableByFile(t, archive, "transactions_tse.csv")
	if !reflect.DeepEqual(tse.Records, want) {
		t.Errorf("transactions_tse.csv records =\n%#v\nwant\n%#v", tse.Records, want)
	}
}

// TestMapAusfallOhneNachsignierungFehlerzeile belegt: ein signaturpflichtiger
// Vorgang, dessen Auftrag (noch) keine Signatur trägt (offen, fehlgeschlagen
// oder tse_nicht_konfiguriert), fehlt nicht in transactions_tse.csv, sondern trägt eine
// Fehlerzeile mit gesetztem TSE_TA_FEHLER und leerer Signatur. So hat jeder
// Bonkopf-Vorgang genau eine TSE-Zeile.
func TestMapAusfallOhneNachsignierungFehlerzeile(t *testing.T) {
	signiert := zahlungAm(t, 1, nachsigniertSignedBonID, time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC))
	unsigniert := zahlungAm(t, 2, nachsigniertOutageBonID, time.Date(2026, 6, 16, 13, 0, 0, 0, time.UTC))

	signaturen := map[int]tse.EventSignatur{
		1: {ProcessType: "Kassenbeleg-V1", Signatur: testSignatur(t, 4711, 12, "2026-06-16T12:00:00Z", "2026-06-16T12:00:01Z", "EVENTSIG==")},
		2: {ProcessType: "Kassenbeleg-V1"}, // Auftrag existiert, Signatur steht aus
	}

	archive, err := Map(testSnapshot(), []event.Event{signiert, unsigniert}, signaturen)
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}

	const erstellung = "2026-06-16T14:30:00Z"
	want := [][]string{
		{testSerial, erstellung, "3", nachsigniertSignedBonID, "1", "4711", "2026-06-16T12:00:00.000Z", "2026-06-16T12:00:01.000Z", "Kassenbeleg-V1", "12", "EVENTSIG==", "", ""},
		{testSerial, erstellung, "3", nachsigniertOutageBonID, "1", "", "", "", "", "", "", tseFehlerAusfall, ""},
	}
	tse := tableByFile(t, archive, "transactions_tse.csv")
	if !reflect.DeepEqual(tse.Records, want) {
		t.Errorf("transactions_tse.csv records =\n%#v\nwant\n%#v", tse.Records, want)
	}

	// Die Fehlerzeile trägt den Vermerk, aber keine Signatur.
	if got := field(t, tse, 1, "TSE_TA_FEHLER"); got != tseFehlerAusfall {
		t.Errorf("TSE_TA_FEHLER = %q, want %q", got, tseFehlerAusfall)
	}
	if got := field(t, tse, 1, "TSE_TA_SIG"); got != "" {
		t.Errorf("TSE_TA_SIG = %q, want leer", got)
	}

	// Jeder Bonkopf-Vorgang hat genau eine TSE-Zeile (signiert oder als Ausfall).
	transactions := tableByFile(t, archive, "transactions.csv")
	if len(tse.Records) != len(transactions.Records) {
		t.Errorf("transactions_tse-Zeilen (%d) ≠ Bonkopf-Vorgänge (%d)", len(tse.Records), len(transactions.Records))
	}
}

// Unsignierte Vorgänge ALLER Vorgangsarten — nicht nur Zahlung und Direktverkauf —
// müssen im Export eine TSE_TA_FEHLER-Zeile tragen: Der Ausfall ergibt sich
// generisch aus einem Signaturauftrag ohne Signatur (Dokumentationspflicht des
// Ausfalls, AEAO 1.14).
func TestUnsignierteVorgaengeAllerArtenTragenAusfallzeile(t *testing.T) {
	// Die Korrektur erhält eine eigene Event-ID (Fixture-Standard kollidiert mit
	// der Zahlung), damit die Signaturen-Map eindeutig bleibt.
	korrektur := korrekturEvent(t)
	korrektur.ID = 4

	events := []event.Event{
		eroeffnetEvent(t, 10000), // ID 10
		bestellungEvent(t),       // ID 1
		zahlungEvent(t),          // ID 2
		warenruecknahmeEvent(t),  // ID 3
		korrektur,                // ID 4
		geldtransitEvent(t),      // ID 11
		differenzEvent(t),        // ID 13
	}

	// Alle Aufträge existieren (signaturpflichtig), keiner ist quittiert.
	signaturen := map[int]tse.EventSignatur{
		10: {ProcessType: "Kassenbeleg-V1"},
		1:  {ProcessType: "Bestellung-V1"},
		2:  {ProcessType: "Kassenbeleg-V1"},
		3:  {ProcessType: "Kassenbeleg-V1"},
		4:  {ProcessType: "Bestellung-V1"},
		11: {ProcessType: "Kassenbeleg-V1"},
		13: {ProcessType: "Kassenbeleg-V1"},
	}

	archive, err := Map(testSnapshot(), events, signaturen)
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}

	transactions := tableByFile(t, archive, "transactions.csv")
	tse := tableByFile(t, archive, "transactions_tse.csv")

	// Invariante: jeder Bonkopf-Vorgang hat genau eine TSE-Zeile.
	if len(tse.Records) != len(transactions.Records) {
		t.Fatalf("transactions_tse-Zeilen (%d) ≠ Bonkopf-Vorgänge (%d)", len(tse.Records), len(transactions.Records))
	}

	for i := range tse.Records {
		if got := field(t, tse, i, "TSE_TA_FEHLER"); got != tseFehlerAusfall {
			t.Errorf("Zeile %d (BON_ID %s): TSE_TA_FEHLER = %q, want %q", i, field(t, tse, i, "BON_ID"), got, tseFehlerAusfall)
		}
		if got := field(t, tse, i, "TSE_TA_SIG"); got != "" {
			t.Errorf("Zeile %d: TSE_TA_SIG = %q, want leer", i, got)
		}
	}
}

// Der Anfangsbestand trägt die TSE-Signatur seines Eröffnungs-Events im Export.
func TestAnfangsbestandTraegtTSESignatur(t *testing.T) {
	evt := eroeffnetEvent(t, 10000)

	signaturen := map[int]tse.EventSignatur{
		10: {ProcessType: "Kassenbeleg-V1", Signatur: testSignatur(t, 7000, 50, "2026-06-16T10:00:00Z", "2026-06-16T10:00:01Z", "EROEFFNUNGSIG==")},
	}

	archive, err := Map(testSnapshot(), []event.Event{evt}, signaturen)
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}

	tse := tableByFile(t, archive, "transactions_tse.csv")
	if len(tse.Records) != 1 {
		t.Fatalf("expected 1 transactions_tse row, got %d", len(tse.Records))
	}
	if got := field(t, tse, 0, "TSE_TA_SIG"); got != "EROEFFNUNGSIG==" {
		t.Errorf("TSE_TA_SIG = %q, want EROEFFNUNGSIG==", got)
	}
	if got := field(t, tse, 0, "TSE_TA_FEHLER"); got != "" {
		t.Errorf("TSE_TA_FEHLER = %q, want leer", got)
	}
}

func TestErstellungszeitpunkt(t *testing.T) {
	fallback := time.Date(2026, 6, 16, 14, 30, 0, 0, time.UTC)
	abschlussZeit := time.Date(2026, 6, 16, 23, 0, 0, 0, time.UTC)

	abschluss := event.Event{
		ID: 9, Type: string(kasse.EventTypeTagesabschlussErstelltV1),
		Time: abschlussZeit, Subject: kasse.KassensitzungSubject(3), Version: 1,
		Data: json.RawMessage(`{}`),
	}

	// Abgeschlossene Sitzung: Zeit stammt aus dem Tagesabschluss-Event.
	if got := Erstellungszeitpunkt([]event.Event{barverkaufEvent(t), abschluss}, fallback); !got.Equal(abschlussZeit) {
		t.Errorf("Erstellungszeitpunkt = %v, want %v (Tagesabschluss)", got, abschlussZeit)
	}

	// Offene Sitzung ohne Abschluss: fallback (Exportzeitpunkt).
	if got := Erstellungszeitpunkt([]event.Event{barverkaufEvent(t)}, fallback); !got.Equal(fallback) {
		t.Errorf("Erstellungszeitpunkt = %v, want %v (fallback)", got, fallback)
	}
}

func TestMapErzeugtAlle20AmtlichenTabellen(t *testing.T) {
	archive, err := Map(testSnapshot(), []event.Event{barverkaufEvent(t)}, barverkaufSignaturen(t))
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}

	// Reihenfolge und Umfang der amtlichen index.xml; nicht befüllte Tabellen
	// (slaves, pa, itemamounts, subitems) sind Header-only enthalten.
	want := []string{
		"cashpointclosing.csv", "location.csv", "cashregister.csv", "slaves.csv", "pa.csv",
		"tse.csv", "vat.csv", "businesscases.csv", "payment.csv", "cash_per_currency.csv",
		"transactions.csv", "datapayment.csv", "lines.csv", "itemamounts.csv", "subitems.csv",
		"transactions_tse.csv", "transactions_vat.csv", "lines_vat.csv", "allocation_groups.csv", "references.csv",
	}

	var got []string
	for _, tbl := range archive.Tables() {
		got = append(got, tbl.File)
		if (tbl.File == "slaves.csv" || tbl.File == "pa.csv" || tbl.File == "itemamounts.csv" || tbl.File == "subitems.csv") && len(tbl.Records) != 0 {
			t.Errorf("%s muss Header-only sein, hat %d Zeilen", tbl.File, len(tbl.Records))
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("tables = %v, want %v", got, want)
	}
}

func TestMapEmptySessionIsError(t *testing.T) {
	// Eine Sitzung nur mit Eröffnungs-Event hat keinen Beleg.
	eroeffnet := event.Event{
		ID: 1, UserID: 7, UserName: "anna",
		Type: string(kasse.EventTypeKassensitzungEroeffnetV1),
		Time: time.Now().UTC(), Subject: kasse.KassensitzungSubject(3), Version: 1,
		Data: json.RawMessage(`{}`),
	}

	_, err := Map(testSnapshot(), []event.Event{eroeffnet}, nil)
	if err != ErrKeineVorgaenge {
		t.Fatalf("Map() error = %v, want ErrKeineVorgaenge", err)
	}
}

// TestBuildCashregisterVersionAusSnapshot belegt, dass KASSE_SW_VERSION aus dem
// Snapshot kommt und nicht aus einer hardcodierten Konstante. Der Test schlaegt
// fehl, wenn die Version hartcodiert wird oder der Snapshot-Wert ignoriert wird.
func TestBuildCashregisterVersionAusSnapshot(t *testing.T) {
	const wantVersion = "1.2.3-test"
	snap := testSnapshot()
	snap.SoftwareVersion = wantVersion

	archive, err := Map(snap, []event.Event{barverkaufEvent(t)}, barverkaufSignaturen(t))
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}
	cr := tableByFile(t, archive, "cashregister.csv")
	if got := field(t, cr, 0, "KASSE_SW_VERSION"); got != wantVersion {
		t.Errorf("KASSE_SW_VERSION = %q, want %q", got, wantVersion)
	}
}

func tableByFile(t *testing.T, a Archive, file string) Table {
	t.Helper()
	for _, tbl := range a.Tables() {
		if tbl.File == file {
			return tbl
		}
	}
	t.Fatalf("table %q not in archive", file)
	return Table{}
}

func field(t *testing.T, table Table, row int, name string) string {
	t.Helper()
	for i, c := range table.Columns {
		if c.name == name {
			return table.Records[row][i]
		}
	}
	t.Fatalf("column %q not found in %s", name, table.File)
	return ""
}

// tseRowByBonID liefert den Zeilenindex der transactions_tse-Zeile eines Bons.
func tseRowByBonID(t *testing.T, table Table, bonID string) int {
	t.Helper()
	for row := range table.Records {
		if field(t, table, row, "BON_ID") == bonID {
			return row
		}
	}
	t.Fatalf("keine transactions_tse-Zeile für BON_ID %q", bonID)
	return -1
}
