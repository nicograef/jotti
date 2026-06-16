//go:build unit

package dsfinvk

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/settings"
)

const (
	testSerial = "11111111-2222-3333-4444-555555555555"
	testBonID  = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
)

func testSnapshot() Snapshot {
	steuernummer := "12345/67890"
	return Snapshot{
		KasseSeriennummer: testSerial,
		Erstellung:        time.Date(2026, 6, 16, 14, 30, 0, 0, time.UTC),
		KassensitzungNr:   3,
		Betreiber: settings.Betreiber{
			Vereinsname:  "TSV Beispiel",
			Strasse:      "Hauptstr. 1",
			Plz:          "12345",
			Ort:          "Musterdorf",
			Steuernummer: &steuernummer,
		},
		TSEStammdaten: settings.TSEStammdaten{
			SignaturAlgorithmus: "ecdsa-plain-SHA256",
			PublicKey:           "PUBKEY==",
			Zertifikat:          "CERTBASE64",
			LogTimeFormat:       "unixTime",
		},
	}
}

// barverkaufEvent baut die einfachste fiskalische Zahlung: ein Bier (19 %) und
// eine Brezel (7 %), bar kassiert und TSE-signiert.
func barverkaufEvent(t *testing.T) event.Event {
	t.Helper()

	data := kasse.ZahlungKassiertV1Data{
		ZahlungID:          testBonID,
		GesamtZahlungCents: 1050,
		Positionen: []kasse.PositionEventData{
			{PositionID: "p1", VarianteID: 101, ProduktName: "Bier", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 450, Menge: 2},
			{PositionID: "p2", VarianteID: 202, ProduktName: "Brezel", VarianteName: "", Kategorie: "essen", Steuersatz: "ermaessigt", Einzelpreis: 150, Menge: 1},
		},
		TSEData: &kasse.TSEData{
			TransactionNumber: 4711,
			SignatureCounter:  12,
			SerialNumberTSE:   "abc123serial",
			LogTimeStart:      "2026-06-16T12:00:00Z",
			LogTimeEnd:        "2026-06-16T12:00:01Z",
			Signature:         "SIGBASE64==",
			ProcessType:       "Kassenbeleg-V1",
			QRCodeData:        "V0;keydata",
		},
	}

	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal zahlung data: %v", err)
	}

	return event.Event{
		ID:       1,
		UserID:   7,
		UserName: "Anna",
		Type:     string(kasse.EventTypeZahlungKassiertV1),
		Time:     time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC),
		Subject:  kasse.TischSessionSubject(3, 42),
		Version:  1,
		Data:     raw,
	}
}

func TestMapBarverkaufGoldenRows(t *testing.T) {
	archive, err := Map(testSnapshot(), []event.Event{barverkaufEvent(t)})
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}

	const erstellung = "2026-06-16T14:30:00Z"

	wantRecords := map[string][][]string{
		"transactions_vat.csv": {
			{testSerial, erstellung, "3", testBonID, "1", "9.00", "7.56", "1.44"},
			{testSerial, erstellung, "3", testBonID, "2", "1.50", "1.40", "0.10"},
		},
		"lines.csv": {
			{testSerial, erstellung, "3", testBonID, "1", "", "Bier 0,5l", "", "Umsatz", "", "", "0", "0", "101", "", "getraenk", "getraenk", "2.000", "", "", "4.50"},
			{testSerial, erstellung, "3", testBonID, "2", "", "Brezel", "", "Umsatz", "", "", "0", "0", "202", "", "essen", "essen", "1.000", "", "", "1.50"},
		},
		"lines_vat.csv": {
			{testSerial, erstellung, "3", testBonID, "1", "1", "9.00", "7.56", "1.44"},
			{testSerial, erstellung, "3", testBonID, "2", "2", "1.50", "1.40", "0.10"},
		},
		"transactions_tse.csv": {
			{testSerial, erstellung, "3", testBonID, "1", "4711", "2026-06-16T12:00:00Z", "2026-06-16T12:00:01Z", "Kassenbeleg-V1", "12", "SIGBASE64==", "", "V0;keydata"},
		},
		"vat.csv": {
			{testSerial, erstellung, "3", "1", "19.00", "Allgemeiner Steuersatz"},
			{testSerial, erstellung, "3", "2", "7.00", "Ermäßigter Steuersatz"},
		},
		"tse.csv": {
			{testSerial, erstellung, "3", "1", "abc123serial", "ecdsa-plain-SHA256", "unixTime", "UTF-8", "PUBKEY==", "CERTBASE64", ""},
		},
		"location.csv": {
			{testSerial, erstellung, "3", "TSV Beispiel", "Hauptstr. 1", "12345", "Musterdorf", "DEU", ""},
		},
		"cashregister.csv": {
			{testSerial, erstellung, "3", "jotti", "jotti mPOS", testSerial, "jotti", "1.0", "EUR", ""},
		},
		"cashpointclosing.csv": {
			{testSerial, erstellung, "3", "", "2.5", testBonID, testBonID, "TSV Beispiel", "Hauptstr. 1", "12345", "Musterdorf", "DEU", "12345/67890", "", "10.50", "10.50"},
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
		"BEDIENER_NAME": "Anna",
		"UMS_BRUTTO":    "10.50",
	}
	for name, want := range checks {
		if got := field(t, transactions, 0, name); got != want {
			t.Errorf("transactions %s = %q, want %q", name, got, want)
		}
	}
}

const (
	bestellungBonID = "11111111-1111-1111-1111-111111111111"
	zahlungBonID    = "22222222-2222-2222-2222-222222222222"
)

// bestellungEvent baut eine offene Bestellung (Forderungsentstehung): ein Bier,
// TSE-signiert als Bestellung-V1, am Tisch 42 aufgenommen.
func bestellungEvent(t *testing.T) event.Event {
	t.Helper()

	data := kasse.BestellungAufgenommenV1Data{
		BestellungID:     bestellungBonID,
		GesamtPreisCents: 450,
		Positionen: []kasse.PositionEventData{
			{PositionID: "p1", VarianteID: 101, ProduktName: "Bier", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 450, Menge: 1},
		},
		TSEData: &kasse.TSEData{
			TransactionNumber: 4710,
			SignatureCounter:  11,
			SerialNumberTSE:   "abc123serial",
			LogTimeStart:      "2026-06-16T11:00:00Z",
			LogTimeEnd:        "2026-06-16T11:00:01Z",
			Signature:         "BESTELLSIG==",
			ProcessType:       "Bestellung-V1",
			QRCodeData:        "V0;bestell",
		},
	}

	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal bestellung data: %v", err)
	}

	return event.Event{
		ID:       1,
		UserID:   7,
		UserName: "Anna",
		Type:     string(kasse.EventTypeBestellungAufgenommenV1),
		Time:     time.Date(2026, 6, 16, 11, 0, 0, 0, time.UTC),
		Subject:  kasse.TischSessionSubject(3, 42),
		Version:  1,
		Data:     raw,
	}
}

// zahlungEvent baut die spätere Barzahlung desselben Tisches (Umsatzrealisierung
// und Forderungsauflösung), TSE-signiert als Kassenbeleg-V1.
func zahlungEvent(t *testing.T) event.Event {
	t.Helper()

	data := kasse.ZahlungKassiertV1Data{
		ZahlungID:          zahlungBonID,
		GesamtZahlungCents: 450,
		Positionen: []kasse.PositionEventData{
			{PositionID: "p1", VarianteID: 101, ProduktName: "Bier", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", Einzelpreis: 450, Menge: 1},
		},
		TSEData: &kasse.TSEData{
			TransactionNumber: 4711,
			SignatureCounter:  12,
			SerialNumberTSE:   "abc123serial",
			LogTimeStart:      "2026-06-16T12:00:00Z",
			LogTimeEnd:        "2026-06-16T12:00:01Z",
			Signature:         "ZAHLSIG==",
			ProcessType:       "Kassenbeleg-V1",
			QRCodeData:        "V0;zahl",
		},
	}

	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal zahlung data: %v", err)
	}

	return event.Event{
		ID:       2,
		UserID:   7,
		UserName: "Anna",
		Type:     string(kasse.EventTypeZahlungKassiertV1),
		Time:     time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC),
		Subject:  kasse.TischSessionSubject(3, 42),
		Version:  1,
		Data:     raw,
	}
}

// TestMapTischablaufTrennt belegt die getrennte Abbildung des gastronomischen
// Tisch-Ablaufs: die Bestellung erscheint als Forderungsentstehung (kein
// Umsatz), die Zahlung als Umsatzrealisierung. Beide tragen denselben
// Abrechnungskreis und je eine eigene TSE-Transaktion.
func TestMapTischablaufTrennt(t *testing.T) {
	snapshot := testSnapshot()
	snapshot.Tischnamen = map[int]string{42: "Tisch 42"}

	archive, err := Map(snapshot, []event.Event{bestellungEvent(t), zahlungEvent(t)})
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}

	const erstellung = "2026-06-16T14:30:00Z"

	wantRecords := map[string][][]string{
		// Getrennte Geschäftsvorfälle: Bestellung = AVBestellung, Zahlung = Beleg.
		"allocation_groups.csv": {
			{testSerial, erstellung, "3", bestellungBonID, "Tisch 42"},
			{testSerial, erstellung, "3", zahlungBonID, "Tisch 42"},
		},
		"datapayment.csv": {
			{testSerial, erstellung, "3", bestellungBonID, "Forderungsentstehung", "Forderungsentstehung", "EUR", "4.50", "4.50"},
			{testSerial, erstellung, "3", zahlungBonID, "Bar", "Bar", "EUR", "4.50", "4.50"},
		},
		"transactions_tse.csv": {
			{testSerial, erstellung, "3", bestellungBonID, "1", "4710", "2026-06-16T11:00:00Z", "2026-06-16T11:00:01Z", "Bestellung-V1", "11", "BESTELLSIG==", "", "V0;bestell"},
			{testSerial, erstellung, "3", zahlungBonID, "1", "4711", "2026-06-16T12:00:00Z", "2026-06-16T12:00:01Z", "Kassenbeleg-V1", "12", "ZAHLSIG==", "", "V0;zahl"},
		},
	}

	for file, want := range wantRecords {
		table := tableByFile(t, archive, file)
		if !reflect.DeepEqual(table.Records, want) {
			t.Errorf("%s records =\n%#v\nwant\n%#v", file, table.Records, want)
		}
	}

	// Forderungsentstehung trägt keinen Umsatz: Bestellung GV_TYP, Zahlung GV_TYP.
	lines := tableByFile(t, archive, "lines.csv")
	if got := field(t, lines, 0, "GV_TYP"); got != "Forderungsentstehung" {
		t.Errorf("bestellung GV_TYP = %q, want Forderungsentstehung", got)
	}
	if got := field(t, lines, 1, "GV_TYP"); got != "Umsatz" {
		t.Errorf("zahlung GV_TYP = %q, want Umsatz", got)
	}

	// BON_TYP trennt offene Bestellung und Zahlungsbeleg.
	transactions := tableByFile(t, archive, "transactions.csv")
	if got := field(t, transactions, 0, "BON_TYP"); got != "AVBestellung" {
		t.Errorf("bestellung BON_TYP = %q, want AVBestellung", got)
	}
	if got := field(t, transactions, 1, "BON_TYP"); got != "Beleg" {
		t.Errorf("zahlung BON_TYP = %q, want Beleg", got)
	}

	// Nur die Barzahlung fließt in den Kassenbestand, nicht die Forderung.
	closing := tableByFile(t, archive, "cashpointclosing.csv")
	if got := field(t, closing, 0, "Z_SE_BARZAHLUNGEN"); got != "4.50" {
		t.Errorf("Z_SE_BARZAHLUNGEN = %q, want 4.50 (nur Zahlung, nicht Forderung)", got)
	}
}

// TestAbrechnungskreisFallback synthetisiert "Tisch N", wenn der Tisch nicht
// mehr in den Stammdaten steht (z. B. gelöscht).
func TestAbrechnungskreisFallback(t *testing.T) {
	snapshot := testSnapshot()
	snapshot.Tischnamen = nil // kein Tischname bekannt

	archive, err := Map(snapshot, []event.Event{barverkaufEvent(t)})
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}

	groups := tableByFile(t, archive, "allocation_groups.csv")
	if got := field(t, groups, 0, "ABRECHNUNGSKREIS"); got != "Tisch 42" {
		t.Errorf("ABRECHNUNGSKREIS = %q, want Tisch 42", got)
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

func TestMapDeclaresOnlyPresentTables(t *testing.T) {
	archive, err := Map(testSnapshot(), []event.Event{barverkaufEvent(t)})
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}

	want := []string{
		"cashpointclosing.csv", "location.csv", "cashregister.csv", "vat.csv", "tse.csv",
		"transactions.csv", "allocation_groups.csv", "transactions_vat.csv", "datapayment.csv",
		"lines.csv", "lines_vat.csv", "transactions_tse.csv",
	}

	var got []string
	for _, tbl := range archive.Tables() {
		got = append(got, tbl.File)
		if tbl.File == "slaves.csv" || tbl.File == "pa.csv" {
			t.Errorf("archive must not contain %s", tbl.File)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("tables = %v, want %v", got, want)
	}
}

func TestMapEmptySessionIsError(t *testing.T) {
	// Eine Sitzung nur mit Eröffnungs-Event hat keinen Beleg.
	eroeffnet := event.Event{
		ID: 1, UserID: 7, UserName: "Anna",
		Type: string(kasse.EventTypeKassensitzungEroeffnetV1),
		Time: time.Now().UTC(), Subject: kasse.KassensitzungSubject(3), Version: 1,
		Data: json.RawMessage(`{}`),
	}

	_, err := Map(testSnapshot(), []event.Event{eroeffnet})
	if err != ErrKeineVorgaenge {
		t.Fatalf("Map() error = %v, want ErrKeineVorgaenge", err)
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
