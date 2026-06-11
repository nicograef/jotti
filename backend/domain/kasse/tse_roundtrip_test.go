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
	dataOhne := assertTypedRoundtrip[bestellungAufgenommenV1Data](t, evtOhne.Data)
	if dataOhne.TSEData != nil {
		t.Fatal("expected nil TSEData for event without TSE")
	}

	evtMit, err := NewBestellungAufgenommenEventMitTSE(testSubject, 1, "User", positionen, "", testRoundtripTSEData("Bestellung-V1"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataMit := assertTypedRoundtrip[bestellungAufgenommenV1Data](t, evtMit.Data)
	if dataMit.TSEData == nil {
		t.Fatal("expected TSEData for event with TSE")
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
	dataOhne := assertTypedRoundtrip[zahlungKassiertV1Data](t, evtOhne.Data)
	if dataOhne.TSEData != nil {
		t.Fatal("expected nil TSEData for event without TSE")
	}

	evtMit, err := NewZahlungKassiertEventMitTSE(testSubject, 1, "User", positionen, 350, "", testRoundtripTSEData("Kassenbeleg-V1"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataMit := assertTypedRoundtrip[zahlungKassiertV1Data](t, evtMit.Data)
	if dataMit.TSEData == nil {
		t.Fatal("expected TSEData for event with TSE")
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

	evtOhne, err := NewStornierungErteiltEvent(testSubject, 1, "User", positionen, 350, "Reklamation")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataOhne := assertTypedRoundtrip[stornierungErteiltV1Data](t, evtOhne.Data)
	if dataOhne.TSEData != nil {
		t.Fatal("expected nil TSEData for event without TSE")
	}

	evtMit, err := NewStornierungErteiltEventMitTSE(testSubject, 1, "User", positionen, 350, "Reklamation", testRoundtripTSEData("Kassenbeleg-V1"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataMit := assertTypedRoundtrip[stornierungErteiltV1Data](t, evtMit.Data)
	if dataMit.TSEData == nil {
		t.Fatal("expected TSEData for event with TSE")
	}
}

func TestEventDataRoundtrip_Auszahlung_WithAndWithoutTSE(t *testing.T) {
	evtOhne, err := NewAuszahlungGeleistetEvent(testSubject, 1, "User", 500, "Rueckzahlung")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataOhne := assertTypedRoundtrip[auszahlungGeleistetV1Data](t, evtOhne.Data)
	if dataOhne.TSEData != nil {
		t.Fatal("expected nil TSEData for event without TSE")
	}

	evtMit, err := NewAuszahlungGeleistetEventMitTSE(testSubject, 1, "User", 500, "Rueckzahlung", testRoundtripTSEData("Kassenbeleg-V1"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataMit := assertTypedRoundtrip[auszahlungGeleistetV1Data](t, evtMit.Data)
	if dataMit.TSEData == nil {
		t.Fatal("expected TSEData for event with TSE")
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
	dataOhne := assertTypedRoundtrip[direktverkaufGetaetigtV1Data](t, evtOhne.Data)
	if dataOhne.TSEData != nil {
		t.Fatal("expected nil TSEData for event without TSE")
	}

	evtMit, err := NewDirektverkaufGetaetigtEventMitTSE(subject, verkaufID, 1, "User", positionen, "", testRoundtripTSEData("Kassenbeleg-V1"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataMit := assertTypedRoundtrip[direktverkaufGetaetigtV1Data](t, evtMit.Data)
	if dataMit.TSEData == nil {
		t.Fatal("expected TSEData for event with TSE")
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
	dataOhne := assertTypedRoundtrip[direktverkaufStorniertV1Data](t, evtOhne.Data)
	if dataOhne.TSEData != nil {
		t.Fatal("expected nil TSEData for event without TSE")
	}

	evtMit, err := NewDirektverkaufStorniertEventMitTSE(subject, verkaufID, 1, "User", positionen, 350, "Rueckgabe", testRoundtripTSEData("Kassenbeleg-V1"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataMit := assertTypedRoundtrip[direktverkaufStorniertV1Data](t, evtMit.Data)
	if dataMit.TSEData == nil {
		t.Fatal("expected TSEData for event with TSE")
	}
}

func TestEventDataRoundtrip_Geldtransit_WithAndWithoutTSE(t *testing.T) {
	subject := KassensitzungSubject(1)

	evtOhne, err := NewGeldtransitGebuchtEvent(subject, 1, "User", "einlage", 1000, "Wechselgeld")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataOhne := assertTypedRoundtrip[geldtransitGebuchtV1Data](t, evtOhne.Data)
	if dataOhne.TSEData != nil {
		t.Fatal("expected nil TSEData for event without TSE")
	}

	evtMit, err := NewGeldtransitGebuchtEventMitTSE(subject, 1, "User", "einlage", 1000, "Wechselgeld", testRoundtripTSEData("Kassenbeleg-V1"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataMit := assertTypedRoundtrip[geldtransitGebuchtV1Data](t, evtMit.Data)
	if dataMit.TSEData == nil {
		t.Fatal("expected TSEData for event with TSE")
	}
}

func TestEventDataRoundtrip_Differenz_WithAndWithoutTSE(t *testing.T) {
	subject := KassensitzungSubject(1)

	evtOhne, err := NewDifferenzSollIstGebuchtEvent(subject, 1, "User", -250)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataOhne := assertTypedRoundtrip[differenzSollIstGebuchtV1Data](t, evtOhne.Data)
	if dataOhne.TSEData != nil {
		t.Fatal("expected nil TSEData for event without TSE")
	}

	evtMit, err := NewDifferenzSollIstGebuchtEventMitTSE(subject, 1, "User", -250, testRoundtripTSEData("Kassenbeleg-V1"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataMit := assertTypedRoundtrip[differenzSollIstGebuchtV1Data](t, evtMit.Data)
	if dataMit.TSEData == nil {
		t.Fatal("expected TSEData for event with TSE")
	}
}

func TestEventDataRoundtrip_Tagesabschluss_WithAndWithoutTSE(t *testing.T) {
	subject := KassensitzungSubject(1)
	von := time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC)
	bis := time.Date(2026, 6, 10, 22, 0, 0, 0, time.UTC)

	evtOhne, err := NewTagesabschlussErstelltEvent(subject, 1, "User", 1, von, bis, 10000, 500, 300, 200)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataOhne := assertTypedRoundtrip[tagesabschlussErstelltV1Data](t, evtOhne.Data)
	if dataOhne.TSEData != nil {
		t.Fatal("expected nil TSEData for event without TSE")
	}

	evtMit, err := NewTagesabschlussErstelltEventMitTSE(subject, 1, "User", 1, von, bis, 10000, 500, 300, 200, testRoundtripTSEData("SonstigerVorgang"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	dataMit := assertTypedRoundtrip[tagesabschlussErstelltV1Data](t, evtMit.Data)
	if dataMit.TSEData == nil {
		t.Fatal("expected TSEData for event with TSE")
	}
}
