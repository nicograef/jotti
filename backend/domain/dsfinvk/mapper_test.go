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

func TestMapDeclaresOnlyPresentTables(t *testing.T) {
	archive, err := Map(testSnapshot(), []event.Event{barverkaufEvent(t)})
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}

	want := []string{
		"cashpointclosing.csv", "location.csv", "cashregister.csv", "vat.csv", "tse.csv",
		"transactions.csv", "transactions_vat.csv", "datapayment.csv",
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
