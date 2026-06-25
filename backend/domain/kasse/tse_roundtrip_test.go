//go:build unit

package kasse

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

func testRoundtripTSEData(processType string) *TSEData {
	return &TSEData{
		TransactionNumber: 101,
		SignatureCounter:  202,
		SerialNumberTSE:   "TSE-SN-1",
		LogTimeStart:      "2026-06-10T20:00:01Z",
		LogTimeEnd:        "2026-06-10T20:00:02Z",
		Signature:         "SIG-XYZ",
		ProcessType:       processType,
		QRCodeData:        "qr-data",
	}
}

func assertTypedRoundtrip[T any](t *testing.T, raw []byte) T {
	t.Helper()

	var first T
	if err := json.Unmarshal(raw, &first); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}

	marshaled, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("expected no marshal error, got %v", err)
	}

	var second T
	if err := json.Unmarshal(marshaled, &second); err != nil {
		t.Fatalf("expected no unmarshal error on roundtrip, got %v", err)
	}

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("expected lossless roundtrip\nfirst:  %+v\nsecond: %+v", first, second)
	}

	return first
}

func TestEventDataRoundtrip_Bestellung_WithAndWithoutTSE(t *testing.T) {
	positionen := []Position{testPosition(1, "Cola", "0,5l", "getraenk", 350, 2)}

	evtOhne, err := NewBestellungAufgenommenEvent(testSubject, 1, "User", positionen, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataOhne := assertTypedRoundtrip[BestellungAufgenommenV1Data](t, evtOhne.Data)
	if dataOhne.TSEData != nil || dataOhne.TSETxID != "" {
		t.Fatal("expected no TSE fields for event without TSE")
	}

	signed, err := EmbedTSEInBestellungAufgenommen(evtOhne, "tx-bestellung", testRoundtripTSEData("Bestellung-V1"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataMit := assertTypedRoundtrip[BestellungAufgenommenV1Data](t, signed.Data)
	if dataMit.TSEData == nil {
		t.Fatal("expected TSEData for signed event")
	}
	if dataMit.TSETxID != "tx-bestellung" {
		t.Fatalf("expected tseTxId to survive roundtrip, got %q", dataMit.TSETxID)
	}
}

func TestEventDataRoundtrip_Zahlung_WithAndWithoutTSE(t *testing.T) {
	positionen := []Position{{
		PositionID:   uuid.NewString(),
		VarianteID:   1,
		ProduktName:  "Cola",
		VarianteName: "0,5l",
		Kategorie:    "getraenk",
		Steuersatz:   "regel",
		Einzelpreis:  350,
		Menge:        1,
	}}

	evtOhne, err := NewZahlungKassiertEvent(testSubject, 1, "User", positionen, 350, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataOhne := assertTypedRoundtrip[ZahlungKassiertV1Data](t, evtOhne.Data)
	if dataOhne.TSEData != nil || dataOhne.TSETxID != "" || dataOhne.TSEAusfall {
		t.Fatal("expected no TSE fields for event without TSE")
	}

	signed, err := EmbedTSEInZahlungKassiert(evtOhne, "tx-zahlung", testRoundtripTSEData("Kassenbeleg-V1"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataMit := assertTypedRoundtrip[ZahlungKassiertV1Data](t, signed.Data)
	if dataMit.TSEData == nil {
		t.Fatal("expected TSEData for signed event")
	}
	if dataMit.TSETxID != "tx-zahlung" || dataMit.TSEAusfall {
		t.Fatalf("expected tseTxId set and tseAusfall false on success, got txId=%q ausfall=%v", dataMit.TSETxID, dataMit.TSEAusfall)
	}

	ausfall, err := EmbedTSEInZahlungKassiert(evtOhne, "tx-ausfall", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataAusfall := assertTypedRoundtrip[ZahlungKassiertV1Data](t, ausfall.Data)
	if dataAusfall.TSEData != nil {
		t.Fatal("expected nil TSEData on Ausfall")
	}
	if dataAusfall.TSETxID != "tx-ausfall" || !dataAusfall.TSEAusfall {
		t.Fatalf("expected tseTxId set and tseAusfall true to survive roundtrip, got txId=%q ausfall=%v", dataAusfall.TSETxID, dataAusfall.TSEAusfall)
	}
}

func TestEventDataRoundtrip_Stornierung_WithAndWithoutTSE(t *testing.T) {
	positionen := []Position{{
		PositionID:   uuid.NewString(),
		VarianteID:   1,
		ProduktName:  "Cola",
		VarianteName: "0,5l",
		Kategorie:    "getraenk",
		Steuersatz:   "regel",
		Einzelpreis:  350,
		Menge:        1,
	}}

	evtOhne, err := NewStornierungErteiltEvent(testSubject, 1, "User", "11111111-1111-1111-1111-111111111111", positionen, 350, "Reklamation")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataOhne := assertTypedRoundtrip[StornierungErteiltV1Data](t, evtOhne.Data)
	if dataOhne.TSEData != nil || dataOhne.TSETxID != "" {
		t.Fatal("expected no TSE fields for event without TSE")
	}

	signed, err := EmbedTSEInStornierungErteilt(evtOhne, "tx-storno", testRoundtripTSEData("Kassenbeleg-V1"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataMit := assertTypedRoundtrip[StornierungErteiltV1Data](t, signed.Data)
	if dataMit.TSEData == nil {
		t.Fatal("expected TSEData for signed event")
	}
	if dataMit.TSETxID != "tx-storno" {
		t.Fatalf("expected tseTxId to survive roundtrip, got %q", dataMit.TSETxID)
	}
}

func TestEventDataRoundtrip_DirektverkaufGetaetigt_WithAndWithoutTSE(t *testing.T) {
	verkaufID := uuid.NewString()
	subject := DirektverkaufSubject(1, verkaufID)
	positionen := []Position{testPosition(1, "Cola", "0,5l", "getraenk", 350, 2)}

	evtOhne, err := NewDirektverkaufGetaetigtEvent(subject, verkaufID, 1, "User", positionen, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataOhne := assertTypedRoundtrip[DirektverkaufGetaetigtV1Data](t, evtOhne.Data)
	if dataOhne.TSEData != nil || dataOhne.TSETxID != "" || dataOhne.TSEAusfall {
		t.Fatal("expected no TSE fields for event without TSE")
	}

	signed, err := EmbedTSEInDirektverkaufGetaetigt(evtOhne, "tx-verkauf", testRoundtripTSEData("Kassenbeleg-V1"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataMit := assertTypedRoundtrip[DirektverkaufGetaetigtV1Data](t, signed.Data)
	if dataMit.TSEData == nil {
		t.Fatal("expected TSEData for signed event")
	}
	if dataMit.TSETxID != "tx-verkauf" || dataMit.TSEAusfall {
		t.Fatalf("expected tseTxId set and tseAusfall false on success, got txId=%q ausfall=%v", dataMit.TSETxID, dataMit.TSEAusfall)
	}

	ausfall, err := EmbedTSEInDirektverkaufGetaetigt(evtOhne, "tx-verkauf-ausfall", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataAusfall := assertTypedRoundtrip[DirektverkaufGetaetigtV1Data](t, ausfall.Data)
	if dataAusfall.TSEData != nil {
		t.Fatal("expected nil TSEData on Ausfall")
	}
	if dataAusfall.TSETxID != "tx-verkauf-ausfall" || !dataAusfall.TSEAusfall {
		t.Fatalf("expected tseTxId set and tseAusfall true to survive roundtrip, got txId=%q ausfall=%v", dataAusfall.TSETxID, dataAusfall.TSEAusfall)
	}
}

func TestEventDataRoundtrip_DirektverkaufStorniert_WithAndWithoutTSE(t *testing.T) {
	verkaufID := uuid.NewString()
	subject := DirektverkaufSubject(1, verkaufID)
	positionen := []Position{{
		PositionID:   uuid.NewString(),
		VarianteID:   1,
		ProduktName:  "Cola",
		VarianteName: "0,5l",
		Kategorie:    "getraenk",
		Steuersatz:   "regel",
		Einzelpreis:  350,
		Menge:        1,
	}}

	evtOhne, err := NewDirektverkaufStorniertEvent(subject, verkaufID, 1, "User", positionen, 350, "Rueckgabe")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataOhne := assertTypedRoundtrip[DirektverkaufStorniertV1Data](t, evtOhne.Data)
	if dataOhne.TSEData != nil || dataOhne.TSETxID != "" {
		t.Fatal("expected no TSE fields for event without TSE")
	}

	signed, err := EmbedTSEInDirektverkaufStorniert(evtOhne, "tx-verkauf-storno", testRoundtripTSEData("Kassenbeleg-V1"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataMit := assertTypedRoundtrip[DirektverkaufStorniertV1Data](t, signed.Data)
	if dataMit.TSEData == nil {
		t.Fatal("expected TSEData for signed event")
	}
	if dataMit.TSETxID != "tx-verkauf-storno" {
		t.Fatalf("expected tseTxId to survive roundtrip, got %q", dataMit.TSETxID)
	}
}

func TestEventDataRoundtrip_Geldtransit_WithAndWithoutTSE(t *testing.T) {
	subject := KassensitzungSubject(1)

	evtOhne, err := NewGeldtransitGebuchtEvent(subject, 1, "User", "einlage", 1000, "Wechselgeld")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataOhne := assertTypedRoundtrip[GeldtransitGebuchtV1Data](t, evtOhne.Data)
	if dataOhne.TSEData != nil || dataOhne.TSETxID != "" {
		t.Fatal("expected no TSE fields for event without TSE")
	}

	signed, err := EmbedTSEInGeldtransitGebucht(evtOhne, "tx-geldtransit", testRoundtripTSEData("Kassenbeleg-V1"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataMit := assertTypedRoundtrip[GeldtransitGebuchtV1Data](t, signed.Data)
	if dataMit.TSEData == nil {
		t.Fatal("expected TSEData for signed event")
	}
	if dataMit.TSETxID != "tx-geldtransit" {
		t.Fatalf("expected tseTxId to survive roundtrip, got %q", dataMit.TSETxID)
	}
}

func TestEventDataRoundtrip_Differenz_WithAndWithoutTSE(t *testing.T) {
	subject := KassensitzungSubject(1)

	evtOhne, err := NewDifferenzSollIstGebuchtEvent(subject, 1, "User", -250)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataOhne := assertTypedRoundtrip[DifferenzSollIstGebuchtV1Data](t, evtOhne.Data)
	if dataOhne.TSEData != nil || dataOhne.TSETxID != "" {
		t.Fatal("expected no TSE fields for event without TSE")
	}

	signed, err := EmbedTSEInDifferenzSollIstGebucht(evtOhne, "tx-differenz", testRoundtripTSEData("Kassenbeleg-V1"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataMit := assertTypedRoundtrip[DifferenzSollIstGebuchtV1Data](t, signed.Data)
	if dataMit.TSEData == nil {
		t.Fatal("expected TSEData for signed event")
	}
	if dataMit.TSETxID != "tx-differenz" {
		t.Fatalf("expected tseTxId to survive roundtrip, got %q", dataMit.TSETxID)
	}
}

func TestEventDataRoundtrip_Tagesabschluss_WithAndWithoutTSE(t *testing.T) {
	subject := KassensitzungSubject(1)
	von := time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC)
	bis := time.Date(2026, 6, 10, 22, 0, 0, 0, time.UTC)

	evtOhne, err := NewTagesabschlussErstelltEvent(subject, 1, "User", 1, von, bis, 10000, 500, 200)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataOhne := assertTypedRoundtrip[TagesabschlussErstelltV1Data](t, evtOhne.Data)
	if dataOhne.TSEData != nil || dataOhne.TSETxID != "" {
		t.Fatal("expected no TSE fields for event without TSE")
	}

	signed, err := EmbedTSEInTagesabschlussErstellt(evtOhne, "tx-tagesabschluss", testRoundtripTSEData("SonstigerVorgang"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataMit := assertTypedRoundtrip[TagesabschlussErstelltV1Data](t, signed.Data)
	if dataMit.TSEData == nil {
		t.Fatal("expected TSEData for signed event")
	}
	if dataMit.TSETxID != "tx-tagesabschluss" {
		t.Fatalf("expected tseTxId to survive roundtrip, got %q", dataMit.TSETxID)
	}
}
