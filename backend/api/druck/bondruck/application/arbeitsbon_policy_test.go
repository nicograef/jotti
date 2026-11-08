package application

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/api/druck/bondruck/application/escpos"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/steuer"
)

// bonZeitpunkt ist der auf jedem Bon gedruckte Zeitpunkt. Die Konstruktoren setzen
// die Event-Zeit auf jetzt; die Tests stempeln sie fest, damit die Byte-Vergleiche
// gegen den Formatter reproduzierbar sind.
var bonZeitpunkt = time.Date(2026, 3, 17, 19, 34, 0, 0, time.UTC)

// position baut eine vollständig gefüllte Position. Die Event-Konstruktoren
// validieren jedes Feld gegen das Positions-Schema, deshalb tragen die Fixtures
// auch die Felder, die kein Arbeitsbon zeigt (VarianteID, Steuersatz, Preis).
func position(produktName, varianteName, kategorie string, menge int) kasse.Position {
	return kasse.Position{
		VarianteID:       1,
		ProduktName:      produktName,
		VarianteName:     varianteName,
		Kategorie:        kategorie,
		Steuersatz:       string(steuer.RegelSteuersatz),
		EinzelpreisCents: 450,
		Menge:            menge,
	}
}

// makeBestellungEvent baut ein echtes bestellung-aufgenommen:v1-Event über den
// Domänen-Konstruktor: Der Test dekodiert damit dieselbe Event-Form, die im
// Kassenjournal steht — samt der dort erzeugten PositionIDs.
func makeBestellungEvent(t *testing.T, id int, subject string, positionen []kasse.Position, kommentar string) event.Event {
	t.Helper()
	evt, err := kasse.NewBestellungAufgenommenEvent(subject, 3, "Maria", uuid.NewString(), positionen, kommentar)
	if err != nil {
		t.Fatalf("bestellung-aufgenommen bauen: %v", err)
	}
	evt.ID = id
	evt.Time = bonZeitpunkt
	return evt
}

// makeDirektverkaufEvent baut ein echtes direktverkauf-getaetigt:v2-Event über den
// Domänen-Konstruktor, aus demselben Grund wie makeBestellungEvent.
func makeDirektverkaufEvent(t *testing.T, id int, positionen []kasse.Position, kommentar string) event.Event {
	t.Helper()
	verkaufID := uuid.NewString()
	subject := fmt.Sprintf("kassensitzung-1/direktverkauf-%s", verkaufID)
	evt, err := kasse.NewDirektverkaufGetaetigtEvent(subject, verkaufID, 3, "Maria", positionen, kommentar)
	if err != nil {
		t.Fatalf("direktverkauf-getaetigt bauen: %v", err)
	}
	evt.ID = id
	evt.Time = bonZeitpunkt
	return evt
}

func TestCreateArbeitsbonAuftraege_ProPosition(t *testing.T) {
	positionen := []kasse.Position{
		position("Pommes", "gross", "essen", 2),
		position("Bratwurst", "mit Brot", "essen", 1),
	}
	evt := makeBestellungEvent(t, 1, "kassensitzung-1/tisch-7", positionen, "ohne Ketchup")
	konfig := map[string]druckstation.Druckstation{
		"essen": {DruckerIP: "192.168.1.51", Bonmodus: "pro_position"},
	}

	auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, konfig, nil, "Tisch 7")

	if len(auftraege) != 2 {
		t.Fatalf("expected 2 auftraege (one per position), got %d", len(auftraege))
	}
	for _, auftrag := range auftraege {
		if auftrag.ZielIP != "192.168.1.51" {
			t.Errorf("expected ZielIP 192.168.1.51, got %s", auftrag.ZielIP)
		}
		if auftrag.BonArt != "arbeitsbon" {
			t.Errorf("expected BonArt arbeitsbon, got %s", auftrag.BonArt)
		}
		if auftrag.Payload == "" {
			t.Error("expected non-empty payload")
		}
	}
}

func TestCreateArbeitsbonAuftraege_ProBestellung(t *testing.T) {
	positionen := []kasse.Position{
		position("Pommes", "gross", "essen", 1),
		position("Schnitzel", "mit Salat", "essen", 1),
	}
	evt := makeBestellungEvent(t, 2, "kassensitzung-1/tisch-5", positionen, "")
	konfig := map[string]druckstation.Druckstation{
		"essen": {DruckerIP: "192.168.1.51", Bonmodus: "pro_bestellung"},
	}

	auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, konfig, nil, "Tisch 5")

	if len(auftraege) != 1 {
		t.Fatalf("expected 1 sammelbon, got %d", len(auftraege))
	}
	if auftraege[0].Payload == "" {
		t.Error("expected non-empty payload")
	}
}

func TestCreateArbeitsbonAuftraege_NoDruckerFuerKategorie(t *testing.T) {
	positionen := []kasse.Position{
		position("Kaffee", "klein", "sonstiges", 1),
	}
	evt := makeBestellungEvent(t, 3, "kassensitzung-1/tisch-2", positionen, "")
	konfig := map[string]druckstation.Druckstation{
		"essen": {DruckerIP: "192.168.1.51", Bonmodus: "pro_position"},
	}

	auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, konfig, nil, "Tisch 2")

	if len(auftraege) != 0 {
		t.Errorf("expected 0 auftraege (no printer for kategorie), got %d", len(auftraege))
	}
}

func TestCreateArbeitsbonAuftraege_ByteIdentischZumFormatter_ProPosition(t *testing.T) {
	positionen := []kasse.Position{
		position("Pommes", "gross", "essen", 2),
		position("Bratwurst", "mit Brot", "essen", 1),
	}
	evt := makeBestellungEvent(t, 10, "kassensitzung-1/tisch-7", positionen, "ohne Ketchup")
	konfig := map[string]druckstation.Druckstation{
		"essen": {DruckerIP: "192.168.1.51", Bonmodus: "pro_position"},
	}

	auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, konfig, nil, "Tisch 7")
	if len(auftraege) != 2 {
		t.Fatalf("expected 2 auftraege, got %d", len(auftraege))
	}

	expectedFirst := base64.StdEncoding.EncodeToString(escpos.FormatPositionBon(
		positionen[0],
		"Tisch 7",
		evt.UserName,
		evt.Time,
		"ohne Ketchup",
	))
	expectedSecond := base64.StdEncoding.EncodeToString(escpos.FormatPositionBon(
		positionen[1],
		"Tisch 7",
		evt.UserName,
		evt.Time,
		"ohne Ketchup",
	))

	if auftraege[0].Payload != expectedFirst {
		t.Error("first payload is not byte-identical to formatter output")
	}
	if auftraege[1].Payload != expectedSecond {
		t.Error("second payload is not byte-identical to formatter output")
	}
}

func decodePayload(t *testing.T, auftrag druckstation.NeuerDruckauftrag) string {
	t.Helper()
	raw, err := base64.StdEncoding.DecodeString(auftrag.Payload)
	if err != nil {
		t.Fatalf("payload dekodieren: %v", err)
	}
	return string(raw)
}

// Regression #145: the Positionskommentar of a dish must not reach the drinks station.
func TestCreateArbeitsbonAuftraege_PositionskommentarNurAnSeinerStation(t *testing.T) {
	essen := position("Zwiebelkuchen", "Stueck", "essen", 1)
	essen.Kommentar = "ohne Speck"
	positionen := []kasse.Position{essen, position("Bier", "0,5l", "getraenk", 1)}
	evt := makeBestellungEvent(t, 11, "kassensitzung-1/tisch-7", positionen, "")
	konfig := map[string]druckstation.Druckstation{
		"essen":    {DruckerIP: "192.168.1.51", Bonmodus: "pro_bestellung"},
		"getraenk": {DruckerIP: "192.168.1.52", Bonmodus: "pro_bestellung"},
	}

	auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, konfig, nil, "Tisch 7")

	if len(auftraege) != 2 {
		t.Fatalf("expected 2 sammelbons, got %d", len(auftraege))
	}
	for _, auftrag := range auftraege {
		payload := decodePayload(t, auftrag)
		traegtKommentar := strings.Contains(payload, "ohne Speck")
		switch auftrag.ZielIP {
		case "192.168.1.51":
			if !traegtKommentar {
				t.Errorf("Essenbon ohne Positionskommentar: %q", payload)
			}
		case "192.168.1.52":
			if traegtKommentar {
				t.Errorf("Getränkebon trägt den Positionskommentar des Essens: %q", payload)
			}
		default:
			t.Errorf("unexpected ZielIP %s", auftrag.ZielIP)
		}
	}
}

func TestCreateArbeitsbonAuftraege_ProPosition_PositionskommentarNurAufSeinemEinzelbon(t *testing.T) {
	kommentiert := position("Zwiebelkuchen", "Stueck", "essen", 1)
	kommentiert.Kommentar = "ohne Speck"
	positionen := []kasse.Position{kommentiert, position("Bratwurst", "mit Brot", "essen", 1)}
	evt := makeBestellungEvent(t, 12, "kassensitzung-1/tisch-7", positionen, "")
	konfig := map[string]druckstation.Druckstation{
		"essen": {DruckerIP: "192.168.1.51", Bonmodus: "pro_position"},
	}

	auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, konfig, nil, "Tisch 7")

	if len(auftraege) != 2 {
		t.Fatalf("expected 2 einzelbons, got %d", len(auftraege))
	}
	for _, auftrag := range auftraege {
		payload := decodePayload(t, auftrag)
		zwiebelkuchen := strings.Contains(payload, "Zwiebelkuchen")
		if strings.Contains(payload, "ohne Speck") != zwiebelkuchen {
			t.Errorf("Positionskommentar steht nicht genau auf dem Zwiebelkuchen-Bon: %q", payload)
		}
	}
}

// Regression: Wenn die Tisch-ID nicht dem Tisch-Namen entspricht (z.B. Tisch-ID 18 heißt
// "Tisch 15", weil zwischendurch andere Tische angelegt wurden), muss der übergebene
// tischName auf dem Bon stehen — nicht die ID aus dem Subject.
func TestCreateArbeitsbonAuftraege_TischNameStattID(t *testing.T) {
	positionen := []kasse.Position{
		position("Pommes", "gross", "essen", 1),
	}
	// tisch-18 in der DB, aber der Name ist "Tisch 15"
	evt := makeBestellungEvent(t, 42, "kassensitzung-1/tisch-18", positionen, "")
	konfig := map[string]druckstation.Druckstation{
		"essen": {DruckerIP: "192.168.1.51", Bonmodus: "pro_position"},
	}

	auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, konfig, nil, "Tisch 15")

	if len(auftraege) != 1 {
		t.Fatalf("expected 1 auftrag, got %d", len(auftraege))
	}

	expected := base64.StdEncoding.EncodeToString(escpos.FormatPositionBon(
		positionen[0],
		"Tisch 15",
		evt.UserName,
		evt.Time,
		"",
	))
	if auftraege[0].Payload != expected {
		t.Error("payload must use tischName 'Tisch 15', not tisch ID 18 from subject")
	}
}

// Direktverkauf-Ableitungsregel: ohne konfigurierte Abholbon-Station gehen die Bons
// an die Produktstationen je Kategorie.
func TestCreateArbeitsbonAuftraege_DirektverkaufAnProduktstationen(t *testing.T) {
	positionen := []kasse.Position{
		position("Pommes", "gross", "essen", 1),
		position("Bier", "0,5l", "getraenk", 2),
	}
	evt := makeDirektverkaufEvent(t, 20, positionen, "schnell")
	stationen := map[string]druckstation.Druckstation{
		"essen":    {DruckerIP: "192.168.1.51", Bonmodus: "pro_bestellung"},
		"getraenk": {DruckerIP: "192.168.1.52", Bonmodus: "pro_bestellung"},
	}

	auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, stationen, nil, "")

	if len(auftraege) != 2 {
		t.Fatalf("expected 2 auftraege (one per configured category), got %d", len(auftraege))
	}
	for _, auftrag := range auftraege {
		if auftrag.BonArt != "arbeitsbon" {
			t.Errorf("expected BonArt arbeitsbon, got %s", auftrag.BonArt)
		}
		if auftrag.Referenz != "direktverkauf-getaetigt:20" {
			t.Errorf("expected referenz direktverkauf-getaetigt:20, got %s", auftrag.Referenz)
		}
	}
}

// Ist die Abholbon-Station konfiguriert, hat sie Vorrang vor den Produktstationen.
func TestCreateArbeitsbonAuftraege_DirektverkaufAbholbon_ProBestellung(t *testing.T) {
	positionen := []kasse.Position{
		position("Pommes", "gross", "essen", 2),
		position("Bier", "0,5l", "getraenk", 1),
	}
	evt := makeDirektverkaufEvent(t, 21, positionen, "ohne Senf")
	stationen := map[string]druckstation.Druckstation{
		"essen":    {DruckerIP: "192.168.1.51", Bonmodus: "pro_position"},
		"abholbon": {DruckerIP: "192.168.1.77", Bonmodus: "pro_bestellung"},
	}

	auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, stationen, nil, "")

	if len(auftraege) != 1 {
		t.Fatalf("expected exactly 1 abholbon auftrag, got %d", len(auftraege))
	}
	if auftraege[0].ZielIP != "192.168.1.77" {
		t.Errorf("expected ZielIP 192.168.1.77, got %s", auftraege[0].ZielIP)
	}
	if auftraege[0].BonArt != "arbeitsbon" {
		t.Errorf("expected BonArt arbeitsbon, got %s", auftraege[0].BonArt)
	}
	if auftraege[0].Referenz != "direktverkauf-getaetigt:21" {
		t.Errorf("expected referenz direktverkauf-getaetigt:21, got %s", auftraege[0].Referenz)
	}

	expected := base64.StdEncoding.EncodeToString(escpos.FormatDirektverkaufAbholbon(positionen, evt.UserName, evt.Time, "ohne Senf"))
	if auftraege[0].Payload != expected {
		t.Error("abholbon payload is not byte-identical to formatter output")
	}
}

// Bonmodus pro_position der Abholbon-Station erzeugt einen Abholbon je Position.
func TestCreateArbeitsbonAuftraege_DirektverkaufAbholbon_ProPosition(t *testing.T) {
	positionen := []kasse.Position{
		position("Pommes", "gross", "essen", 2),
		position("Bier", "0,5l", "getraenk", 1),
	}
	evt := makeDirektverkaufEvent(t, 22, positionen, "")
	stationen := map[string]druckstation.Druckstation{
		"abholbon": {DruckerIP: "192.168.1.77", Bonmodus: "pro_position"},
	}

	auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, stationen, nil, "")

	if len(auftraege) != 2 {
		t.Fatalf("expected 2 abholbon auftraege (one per position), got %d", len(auftraege))
	}
	for i, auftrag := range auftraege {
		if auftrag.ZielIP != "192.168.1.77" {
			t.Errorf("expected ZielIP 192.168.1.77, got %s", auftrag.ZielIP)
		}
		expected := base64.StdEncoding.EncodeToString(escpos.FormatDirektverkaufAbholbon([]kasse.Position{positionen[i]}, evt.UserName, evt.Time, ""))
		if auftrag.Payload != expected {
			t.Errorf("abholbon payload %d is not byte-identical to formatter output", i)
		}
	}
}

// Bonmodus pro_stueck der Abholbon-Station erzeugt einen Abholbon je Einheit; jeder Bon
// trägt eine Positions-Kopie mit Menge 1.
func TestCreateArbeitsbonAuftraege_DirektverkaufAbholbon_ProStueck(t *testing.T) {
	positionen := []kasse.Position{
		position("Bier", "0,5l", "getraenk", 3),
		position("Pommes", "gross", "essen", 1),
	}
	evt := makeDirektverkaufEvent(t, 24, positionen, "")
	stationen := map[string]druckstation.Druckstation{
		"abholbon": {DruckerIP: "192.168.1.77", Bonmodus: "pro_stueck"},
	}

	auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, stationen, nil, "")

	if len(auftraege) != 4 {
		t.Fatalf("expected 4 abholbon auftraege (one per unit), got %d", len(auftraege))
	}

	einzeln := func(pos kasse.Position) string {
		pos.Menge = 1
		return base64.StdEncoding.EncodeToString(
			escpos.FormatDirektverkaufAbholbon([]kasse.Position{pos}, evt.UserName, evt.Time, ""),
		)
	}
	erwartet := []string{einzeln(positionen[0]), einzeln(positionen[0]), einzeln(positionen[0]), einzeln(positionen[1])}

	for i, auftrag := range auftraege {
		if auftrag.ZielIP != "192.168.1.77" {
			t.Errorf("expected ZielIP 192.168.1.77, got %s", auftrag.ZielIP)
		}
		if auftrag.BonArt != "arbeitsbon" {
			t.Errorf("expected BonArt arbeitsbon, got %s", auftrag.BonArt)
		}
		if auftrag.Referenz != "direktverkauf-getaetigt:24" {
			t.Errorf("expected referenz direktverkauf-getaetigt:24, got %s", auftrag.Referenz)
		}
		if auftrag.Payload != erwartet[i] {
			t.Errorf("abholbon payload %d is not byte-identical to the formatter output for a Menge-1 position", i)
		}
	}
}

// Pfand-Positionen erscheinen in keinem Bonmodus auf einem Abholbon; ein reiner Pfand-Verkauf
// erzeugt keinen Auftrag.
func TestCreateArbeitsbonAuftraege_DirektverkaufAbholbon_OhnePfand(t *testing.T) {
	bier := position("Bier", "0,5l", "getraenk", 2)
	pfand := position("Glaspfand", "0,5l", "getraenk", 2)
	pfand.EinzelpreisCents = 200
	pfand.Pfand = true

	abholbon := func(evt event.Event, positionen ...kasse.Position) string {
		return base64.StdEncoding.EncodeToString(escpos.FormatDirektverkaufAbholbon(positionen, evt.UserName, evt.Time, ""))
	}
	bierEinheit := bier
	bierEinheit.Menge = 1

	tests := []struct {
		bonmodus druckstation.Bonmodus
		erwartet func(evt event.Event) []string
	}{
		{druckstation.BonmodusProBestellung, func(evt event.Event) []string { return []string{abholbon(evt, bier)} }},
		{druckstation.BonmodusProPosition, func(evt event.Event) []string { return []string{abholbon(evt, bier)} }},
		{druckstation.BonmodusProStueck, func(evt event.Event) []string {
			return []string{abholbon(evt, bierEinheit), abholbon(evt, bierEinheit)}
		}},
	}
	for _, tt := range tests {
		t.Run(string(tt.bonmodus), func(t *testing.T) {
			stationen := map[string]druckstation.Druckstation{
				"getraenk": {DruckerIP: "192.168.1.52", Bonmodus: "pro_position"},
				"abholbon": {DruckerIP: "192.168.1.77", Bonmodus: tt.bonmodus},
			}

			evt := makeDirektverkaufEvent(t, 26, []kasse.Position{bier, pfand}, "")
			auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, stationen, nil, "")

			erwartet := tt.erwartet(evt)
			if len(auftraege) != len(erwartet) {
				t.Fatalf("expected %d abholbons with Bier only, got %d", len(erwartet), len(auftraege))
			}
			for i, auftrag := range auftraege {
				if auftrag.ZielIP != "192.168.1.77" {
					t.Errorf("expected ZielIP 192.168.1.77, got %s", auftrag.ZielIP)
				}
				if auftrag.Payload != erwartet[i] {
					t.Errorf("abholbon %d is not the Bier-only formatter output: %q", i, decodePayload(t, auftrag))
				}
			}

			nurPfand := makeDirektverkaufEvent(t, 27, []kasse.Position{pfand}, "")
			if auftraege := CreateArbeitsbonAuftraegeFromEvent(nurPfand, stationen, nil, ""); len(auftraege) != 0 {
				t.Errorf("expected no auftrag for a Pfand-only sale, got %d", len(auftraege))
			}
		})
	}
}

func TestCreateArbeitsbonAuftraege_DirektverkaufAbholbon_ProStueck_PositionskommentarAufJedemStueckbon(t *testing.T) {
	wurst := position("Bratwurst", "mit Brot", "essen", 2)
	wurst.Kommentar = "1 davon ohne Senf"
	evt := makeDirektverkaufEvent(t, 25, []kasse.Position{wurst, position("Bier", "0,5l", "getraenk", 1)}, "")
	stationen := map[string]druckstation.Druckstation{
		"abholbon": {DruckerIP: "192.168.1.77", Bonmodus: "pro_stueck"},
	}

	auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, stationen, nil, "")

	if len(auftraege) != 3 {
		t.Fatalf("expected 3 stückbons, got %d", len(auftraege))
	}
	wurstBons := 0
	for _, auftrag := range auftraege {
		payload := decodePayload(t, auftrag)
		bratwurst := strings.Contains(payload, "Bratwurst")
		if bratwurst {
			wurstBons++
		}
		if strings.Contains(payload, "1 davon ohne Senf") != bratwurst {
			t.Errorf("Positionskommentar on the wrong stückbon (Bratwurst %v): %q", bratwurst, payload)
		}
	}
	if wurstBons != 2 {
		t.Errorf("expected 2 Bratwurst stückbons, got %d", wurstBons)
	}
}

// Ohne konfigurierte Druckstationen entstehen für einen Direktverkauf keine Aufträge.
func TestCreateArbeitsbonAuftraege_DirektverkaufOhneStationen(t *testing.T) {
	evt := makeDirektverkaufEvent(t, 23, []kasse.Position{position("Pommes", "gross", "essen", 1)}, "")

	auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, map[string]druckstation.Druckstation{}, nil, "")

	if len(auftraege) != 0 {
		t.Errorf("expected 0 auftraege without configured stations, got %d", len(auftraege))
	}
}

// Bonmodus kein_bon am Abholbon schaltet den Direktverkauf stumm, mit und ohne
// Abholbon-IP und trotz konfigurierter Produktstationen. Tischbestellungen drucken weiter.
func TestCreateArbeitsbonAuftraege_DirektverkaufAus(t *testing.T) {
	positionen := []kasse.Position{
		position("Pommes", "gross", "essen", 1),
		position("Bier", "0,5l", "getraenk", 2),
	}
	produktstationen := func(abholbonIP string) map[string]druckstation.Druckstation {
		return map[string]druckstation.Druckstation{
			"essen":    {DruckerIP: "192.168.1.51", Bonmodus: "pro_position"},
			"getraenk": {DruckerIP: "192.168.1.52", Bonmodus: "pro_bestellung"},
			"abholbon": {DruckerIP: abholbonIP, Bonmodus: "kein_bon"},
		}
	}

	cases := []struct {
		name      string
		evt       event.Event
		stationen map[string]druckstation.Druckstation
		want      int
	}{
		{"direktverkauf, abholbon ohne IP", makeDirektverkaufEvent(t, 40, positionen, ""), produktstationen(""), 0},
		{"direktverkauf, abholbon mit IP", makeDirektverkaufEvent(t, 41, positionen, ""), produktstationen("192.168.1.77"), 0},
		{"tischbestellung druckt weiter", makeBestellungEvent(t, 42, "kassensitzung-1/tisch-1", positionen, ""), produktstationen(""), 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			auftraege := CreateArbeitsbonAuftraegeFromEvent(tc.evt, tc.stationen, nil, "Tisch 1")
			if len(auftraege) != tc.want {
				t.Errorf("expected %d auftraege, got %d", tc.want, len(auftraege))
			}
		})
	}
}

// variante gives a fixture position its own VarianteID, the key of the Zusatzdrucker-Zuordnung.
func variante(pos kasse.Position, varianteID int) kasse.Position {
	pos.VarianteID = varianteID
	return pos
}

func sammelbon(evt event.Event, kontext string, positionen ...kasse.Position) string {
	return base64.StdEncoding.EncodeToString(escpos.FormatSammelBon(positionen, kontext, evt.UserName, evt.Time, ""))
}

type erwarteterAuftrag struct {
	zielIP  string
	payload string
}

func assertAuftraege(t *testing.T, auftraege []druckstation.NeuerDruckauftrag, referenz string, erwartet []erwarteterAuftrag) {
	t.Helper()
	if len(auftraege) != len(erwartet) {
		t.Fatalf("expected %d auftraege, got %d", len(erwartet), len(auftraege))
	}
	for i, auftrag := range auftraege {
		if auftrag.ZielIP != erwartet[i].zielIP {
			t.Errorf("auftrag %d: expected ZielIP %s, got %s", i, erwartet[i].zielIP, auftrag.ZielIP)
		}
		if auftrag.Payload != erwartet[i].payload {
			t.Errorf("auftrag %d: unexpected payload %q", i, decodePayload(t, auftrag))
		}
		if auftrag.BonArt != druckstation.BonArtArbeitsbon {
			t.Errorf("auftrag %d: expected BonArt arbeitsbon, got %s", i, auftrag.BonArt)
		}
		if auftrag.Referenz != referenz {
			t.Errorf("auftrag %d: expected referenz %s, got %s", i, referenz, auftrag.Referenz)
		}
	}
}

// Ein zugeordnetes Produkt druckt am Zusatzdrucker, nicht an der Station seiner Kategorie; auch bei
// gleicher IP bleiben es zwei Aufträge.
func TestCreateArbeitsbonAuftraege_ZusatzdruckerVorKategorie(t *testing.T) {
	pommes := variante(position("Pommes", "gross", "essen", 1), 1)
	kuchen := variante(position("Kuchen", "Stueck", "essen", 2), 2)
	evt := makeBestellungEvent(t, 50, "kassensitzung-1/tisch-7", []kasse.Position{pommes, kuchen}, "")
	stationen := map[string]druckstation.Druckstation{
		"essen": {DruckerIP: "192.168.1.51", Bonmodus: "pro_bestellung"},
	}
	zusatzdrucker := map[int]druckstation.Zusatzdrucker{
		2: {ID: 1, Name: "Kaffee & Kuchen", DruckerIP: "192.168.1.51", Bonmodus: "pro_bestellung"},
	}

	auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, stationen, zusatzdrucker, "Tisch 7")

	assertAuftraege(t, auftraege, "bestellung-aufgenommen:50", []erwarteterAuftrag{
		{"192.168.1.51", sammelbon(evt, "Tisch 7", pommes)},
		{"192.168.1.51", sammelbon(evt, "Tisch 7", kuchen)},
	})
}

// Ein Zusatzdrucker ohne IP druckt nichts, und seine Produkte fallen nicht auf die Kategorie zurück.
func TestCreateArbeitsbonAuftraege_ZusatzdruckerOhneIPOhneRueckfall(t *testing.T) {
	pommes := variante(position("Pommes", "gross", "essen", 1), 1)
	kuchen := variante(position("Kuchen", "Stueck", "essen", 1), 2)
	evt := makeBestellungEvent(t, 51, "kassensitzung-1/tisch-7", []kasse.Position{kuchen, pommes}, "")
	stationen := map[string]druckstation.Druckstation{
		"essen": {DruckerIP: "192.168.1.51", Bonmodus: "pro_bestellung"},
	}
	zusatzdrucker := map[int]druckstation.Zusatzdrucker{
		2: {ID: 1, Name: "Kaffee & Kuchen", DruckerIP: "", Bonmodus: "pro_bestellung"},
	}

	auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, stationen, zusatzdrucker, "Tisch 7")

	assertAuftraege(t, auftraege, "bestellung-aufgenommen:51", []erwarteterAuftrag{
		{"192.168.1.51", sammelbon(evt, "Tisch 7", pommes)},
	})
}

// pro_bestellung ergibt je Zusatzdrucker einen Sammelbon mit genau seinen Positionen, in der
// Reihenfolge des ersten Auftretens.
func TestCreateArbeitsbonAuftraege_ZusatzdruckerSammelbonJeDrucker(t *testing.T) {
	kuchen := variante(position("Kuchen", "Stueck", "essen", 1), 2)
	kaffee := variante(position("Kaffee", "Tasse", "getraenk", 2), 3)
	torte := variante(position("Torte", "Stueck", "essen", 1), 4)
	evt := makeBestellungEvent(t, 52, "kassensitzung-1/tisch-7", []kasse.Position{kuchen, kaffee, torte}, "")
	kuchentheke := druckstation.Zusatzdrucker{ID: 1, Name: "Kuchentheke", DruckerIP: "192.168.1.61", Bonmodus: "pro_bestellung"}
	kaffeebar := druckstation.Zusatzdrucker{ID: 2, Name: "Kaffeebar", DruckerIP: "192.168.1.62", Bonmodus: "pro_bestellung"}
	zusatzdrucker := map[int]druckstation.Zusatzdrucker{2: kuchentheke, 3: kaffeebar, 4: kuchentheke}

	auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, map[string]druckstation.Druckstation{}, zusatzdrucker, "Tisch 7")

	assertAuftraege(t, auftraege, "bestellung-aufgenommen:52", []erwarteterAuftrag{
		{"192.168.1.61", sammelbon(evt, "Tisch 7", kuchen, torte)},
		{"192.168.1.62", sammelbon(evt, "Tisch 7", kaffee)},
	})
}

// Im Direktverkauf gelten „Kein Bon" und der Abholbon vor der Zuordnung.
func TestCreateArbeitsbonAuftraege_DirektverkaufAbholbonIgnoriertZusatzdrucker(t *testing.T) {
	pommes := variante(position("Pommes", "gross", "essen", 1), 1)
	kuchen := variante(position("Kuchen", "Stueck", "essen", 1), 2)
	zusatzdrucker := map[int]druckstation.Zusatzdrucker{
		2: {ID: 1, Name: "Kaffee & Kuchen", DruckerIP: "192.168.1.61", Bonmodus: "pro_position"},
	}
	stationen := func(bonmodus druckstation.Bonmodus) map[string]druckstation.Druckstation {
		return map[string]druckstation.Druckstation{
			"essen":    {DruckerIP: "192.168.1.51", Bonmodus: "pro_position"},
			"abholbon": {DruckerIP: "192.168.1.77", Bonmodus: bonmodus},
		}
	}

	t.Run("abholbon", func(t *testing.T) {
		evt := makeDirektverkaufEvent(t, 53, []kasse.Position{pommes, kuchen}, "")
		auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, stationen(druckstation.BonmodusProBestellung), zusatzdrucker, "")
		abholbon := base64.StdEncoding.EncodeToString(escpos.FormatDirektverkaufAbholbon([]kasse.Position{pommes, kuchen}, evt.UserName, evt.Time, ""))
		assertAuftraege(t, auftraege, "direktverkauf-getaetigt:53", []erwarteterAuftrag{{"192.168.1.77", abholbon}})
	})

	t.Run("kein bon", func(t *testing.T) {
		evt := makeDirektverkaufEvent(t, 54, []kasse.Position{pommes, kuchen}, "")
		if auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, stationen(druckstation.BonmodusKeinBon), zusatzdrucker, ""); len(auftraege) != 0 {
			t.Errorf("expected no auftrag with kein_bon, got %d", len(auftraege))
		}
	})
}

// Ohne Abholbon druckt ein zugeordnetes Produkt im Direktverkauf am Zusatzdrucker.
func TestCreateArbeitsbonAuftraege_DirektverkaufAmZusatzdrucker(t *testing.T) {
	pommes := variante(position("Pommes", "gross", "essen", 1), 1)
	kuchen := variante(position("Kuchen", "Stueck", "essen", 1), 2)
	evt := makeDirektverkaufEvent(t, 55, []kasse.Position{kuchen, pommes}, "")
	stationen := map[string]druckstation.Druckstation{
		"essen": {DruckerIP: "192.168.1.51", Bonmodus: "pro_bestellung"},
	}
	zusatzdrucker := map[int]druckstation.Zusatzdrucker{
		2: {ID: 1, Name: "Kaffee & Kuchen", DruckerIP: "192.168.1.61", Bonmodus: "pro_position"},
	}

	auftraege := CreateArbeitsbonAuftraegeFromEvent(evt, stationen, zusatzdrucker, "")

	assertAuftraege(t, auftraege, "direktverkauf-getaetigt:55", []erwarteterAuftrag{
		{"192.168.1.61", base64.StdEncoding.EncodeToString(escpos.FormatPositionBon(kuchen, "Direktverkauf", evt.UserName, evt.Time, ""))},
		{"192.168.1.51", sammelbon(evt, "Direktverkauf", pommes)},
	})
}
