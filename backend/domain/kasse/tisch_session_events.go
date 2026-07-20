package kasse

import (
	"fmt"
	"slices"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

type EventType string

const (
	EventTypeBestellungAufgenommenV1 EventType = "bestellung-aufgenommen:v1"
	EventTypeZahlungKassiertV1       EventType = "zahlung-kassiert:v1"
	EventTypeStornierungErteiltV1    EventType = "stornierung-erteilt:v1"
	EventTypeBestellungKorrigiertV1  EventType = "bestellung-korrigiert:v1"
	EventTypeBestellungUmgebuchtV1   EventType = "bestellung-umgebucht:v1"
)

// --- Event-Data-Structs ---

type BestellungAufgenommenV1Data struct {
	BestellungID     string              `json:"bestellungId"`
	Positionen       []PositionEventData `json:"positionen"`
	GesamtPreisCents int                 `json:"gesamtPreisCents"`
	Kommentar        string              `json:"kommentar"`
}

var bestellungAufgenommenV1DataSchema = z.Struct(z.Shape{
	"BestellungID": z.String().UUID().Required(),
	"Positionen":   z.Slice(positionSchema).Min(1).Required(),
	// Muss positiv: eine Summe wird über Positionen mit Preis >= 1 Cent gebildet;
	// 0 ist keine gültige Summe (0-Cent-Positionen sind nicht zulässig).
	"GesamtPreisCents": z.Int().GTE(1).Required(),
	"Kommentar":        z.String().Max(100),
})

type ZahlungKassiertV1Data struct {
	ZahlungID          string              `json:"zahlungId"`
	Positionen         []PositionEventData `json:"positionen"`
	GesamtZahlungCents int                 `json:"gesamtZahlungCents"`
	Kommentar          string              `json:"kommentar"`
}

var zahlungKassiertV1DataSchema = z.Struct(z.Shape{
	"ZahlungID":  z.String().UUID().Required(),
	"Positionen": z.Slice(positionSchema).Min(1).Required(),
	// Muss positiv: eine Summe wird über Positionen mit Preis >= 1 Cent gebildet;
	// 0 ist keine gültige Summe (0-Cent-Positionen sind nicht zulässig).
	"GesamtZahlungCents": z.Int().GTE(1).Required(),
	"Kommentar":          z.String().Max(100),
})

// StornierungErteiltV1Data ist die kassenwirksame Warenrücknahme bereits bezahlter
// Positionen: ein negativer Umsatz am Ursprungssteuersatz mit Bar-Rückgabe, signiert
// als Kassenbeleg-V1. ZahlungID referenziert genau die begleichende Zahlung, deren
// Mengen zurückgenommen werden (eine Warenrücknahme je Zahlung). Der Kommentar ist
// Pflicht (Dokumentation des Rückgabegrunds für die Betriebsprüfung).
type StornierungErteiltV1Data struct {
	StornierungID          string              `json:"stornierungId"`
	ZahlungID              string              `json:"zahlungId"`
	Positionen             []PositionEventData `json:"positionen"`
	GesamtStornierungCents int                 `json:"gesamtStornierungCents"`
	Kommentar              string              `json:"kommentar"`
}

var stornierungErteiltV1DataSchema = z.Struct(z.Shape{
	"StornierungID": z.String().UUID().Required(),
	"ZahlungID":     z.String().UUID().Required(),
	"Positionen":    z.Slice(positionSchema).Min(1).Required(),
	// Muss positiv: eine Summe wird über Positionen mit Preis >= 1 Cent gebildet;
	// 0 ist keine gültige Summe (0-Cent-Positionen sind nicht zulässig).
	"GesamtStornierungCents": z.Int().GTE(1).Required(),
	"Kommentar":              z.String().Min(3).Max(100).Required(),
})

// BestellungKorrigiertV1Data ist die geldneutrale Stornierung noch unbezahlter
// Positionen: eine reine Auftragskorrektur ohne Geld- und Umsatzwirkung, signiert als
// Bestellung-V1 (ohne Zahlungszeile). Sie reduziert den offenen Betrag und nimmt die
// Positionen aus den aktiven Listen.
type BestellungKorrigiertV1Data struct {
	KorrekturID string              `json:"korrekturId"`
	Positionen  []PositionEventData `json:"positionen"`
	GesamtCents int                 `json:"gesamtCents"`
	Kommentar   string              `json:"kommentar"`
}

var bestellungKorrigiertV1DataSchema = z.Struct(z.Shape{
	"KorrekturID": z.String().UUID().Required(),
	"Positionen":  z.Slice(positionSchema).Min(1).Required(),
	// Muss positiv: eine Summe wird über Positionen mit Preis >= 1 Cent gebildet;
	// 0 ist keine gültige Summe (0-Cent-Positionen sind nicht zulässig).
	"GesamtCents": z.Int().GTE(1).Required(),
	"Kommentar":   z.String().Max(100),
})

// BestellungUmgebuchtV1Data ist die geldneutrale Umbuchung unbezahlter Positionen
// zwischen zwei Tischen. Quell- und Zielstrom erhalten je ein Event mit derselben
// UmbuchungID; das Event auf dem Quelltisch trägt die ursprünglichen PositionIDs
// (Abgang), das auf dem Zieltisch frische (Zugang). Die Projektion unterscheidet die
// Richtung über den Tisch des Event-Subjects (QuellTischID vs. ZielTischID).
type BestellungUmgebuchtV1Data struct {
	UmbuchungID  string              `json:"umbuchungId"`
	QuellTischID int                 `json:"quellTischId"`
	ZielTischID  int                 `json:"zielTischId"`
	Positionen   []PositionEventData `json:"positionen"`
	GesamtCents  int                 `json:"gesamtCents"`
	// Kommentar trägt den Richtungs-Autotext ("Umbuchung auf/von Tisch X") und ist
	// stets gesetzt. BenutzerKommentar ist der optionale, frei eingegebene Text; er
	// fehlt im JSON, wenn leer (omitempty), damit Events ohne Benutzertext
	// byte-identisch zum ursprünglichen Format bleiben.
	Kommentar         string `json:"kommentar"`
	BenutzerKommentar string `json:"benutzerKommentar,omitempty"`
}

var bestellungUmgebuchtV1DataSchema = z.Struct(z.Shape{
	"UmbuchungID":  z.String().UUID().Required(),
	"QuellTischID": z.Int().GTE(1).Required(),
	"ZielTischID":  z.Int().GTE(1).Required(),
	"Positionen":   z.Slice(positionSchema).Min(1).Required(),
	// Muss positiv: eine Summe wird über Positionen mit Preis >= 1 Cent gebildet;
	// 0 ist keine gültige Summe (0-Cent-Positionen sind nicht zulässig).
	"GesamtCents":       z.Int().GTE(1).Required(),
	"Kommentar":         z.String().Max(100),
	"BenutzerKommentar": z.String().Max(100),
})

// --- Event-Erstellungsfunktionen ---

func NewBestellungAufgenommenEvent(subject string, userID int, userName string, bestellungID string, positionen []Position, kommentar string) (e.Event, error) {
	// Generate PositionIDs for each position (on a copy, so the caller's slice stays untouched)
	positionen = slices.Clone(positionen)
	for i := range positionen {
		positionen[i].PositionID = uuid.New().String()
	}

	gesamtPreisCents := 0
	for _, pos := range positionen {
		gesamtPreisCents += pos.EinzelpreisCents * pos.Menge
	}

	data := BestellungAufgenommenV1Data{
		BestellungID:     bestellungID,
		Positionen:       toPositionenEventData(positionen),
		GesamtPreisCents: gesamtPreisCents,
		Kommentar:        kommentar,
	}

	if err := bestellungAufgenommenV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("bestellung aufgenommen data validation failed: %v", issues)
	}

	return e.New(userID, userName, string(EventTypeBestellungAufgenommenV1), subject, data)
}

func NewZahlungKassiertEvent(subject string, userID int, userName string, positionen []Position, gesamtZahlungCents int, kommentar string) (e.Event, error) {
	data := ZahlungKassiertV1Data{
		ZahlungID:          uuid.New().String(),
		Positionen:         toPositionenEventData(positionen),
		GesamtZahlungCents: gesamtZahlungCents,
		Kommentar:          kommentar,
	}

	if err := zahlungKassiertV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("zahlung kassiert data validation failed: %v", issues)
	}

	return e.New(userID, userName, string(EventTypeZahlungKassiertV1), subject, data)
}

func NewStornierungErteiltEvent(subject string, userID int, userName string, zahlungID string, positionen []Position, gesamtStornierungCents int, kommentar string) (e.Event, error) {
	data := StornierungErteiltV1Data{
		StornierungID:          uuid.New().String(),
		ZahlungID:              zahlungID,
		Positionen:             toPositionenEventData(positionen),
		GesamtStornierungCents: gesamtStornierungCents,
		Kommentar:              kommentar,
	}

	if err := stornierungErteiltV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("stornierung erteilt data validation failed: %v", issues)
	}

	return e.New(userID, userName, string(EventTypeStornierungErteiltV1), subject, data)
}

func NewBestellungKorrigiertEvent(subject string, userID int, userName string, positionen []Position, gesamtCents int, kommentar string) (e.Event, error) {
	data := BestellungKorrigiertV1Data{
		KorrekturID: uuid.New().String(),
		Positionen:  toPositionenEventData(positionen),
		GesamtCents: gesamtCents,
		Kommentar:   kommentar,
	}

	if err := bestellungKorrigiertV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("bestellung korrigiert data validation failed: %v", issues)
	}

	return e.New(userID, userName, string(EventTypeBestellungKorrigiertV1), subject, data)
}

// NewBestellungUmgebuchtEvents erzeugt das verknüpfte Event-Paar einer Umbuchung:
// ein Abgang auf dem Quelltisch (mit den übergebenen, ursprünglichen Positionen) und
// ein Zugang auf dem Zieltisch (mit frischen PositionIDs, damit die Positionen dort
// eigenständig weiterverarbeitet werden können). Beide teilen sich eine UmbuchungID
// und werden vom Aufrufer atomar geschrieben.
func NewBestellungUmgebuchtEvents(zNr int, quellTischID int, zielTischID int, userID int, userName string, quellPositionen []Position, gesamtCents int, quellKommentar string, zielKommentar string, benutzerKommentar string) (e.Event, e.Event, error) {
	umbuchungID := uuid.New().String()

	zielPositionen := slices.Clone(quellPositionen)
	for i := range zielPositionen {
		zielPositionen[i].PositionID = uuid.New().String()
	}

	quellData := BestellungUmgebuchtV1Data{
		UmbuchungID:       umbuchungID,
		QuellTischID:      quellTischID,
		ZielTischID:       zielTischID,
		Positionen:        toPositionenEventData(quellPositionen),
		GesamtCents:       gesamtCents,
		Kommentar:         quellKommentar,
		BenutzerKommentar: benutzerKommentar,
	}
	zielData := quellData
	zielData.Positionen = toPositionenEventData(zielPositionen)
	zielData.Kommentar = zielKommentar

	if err := bestellungUmgebuchtV1DataSchema.Validate(&quellData); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, e.Event{}, fmt.Errorf("bestellung umgebucht (quelle) data validation failed: %v", issues)
	}
	if err := bestellungUmgebuchtV1DataSchema.Validate(&zielData); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, e.Event{}, fmt.Errorf("bestellung umgebucht (ziel) data validation failed: %v", issues)
	}

	quellEvent, err := e.New(userID, userName, string(EventTypeBestellungUmgebuchtV1), TischSessionSubject(zNr, quellTischID), quellData)
	if err != nil {
		return e.Event{}, e.Event{}, err
	}
	zielEvent, err := e.New(userID, userName, string(EventTypeBestellungUmgebuchtV1), TischSessionSubject(zNr, zielTischID), zielData)
	if err != nil {
		return e.Event{}, e.Event{}, err
	}

	return quellEvent, zielEvent, nil
}

// --- Build-from-Event-Funktionen ---

func buildBestellungFromEvent(event e.Event) (Bestellung, error) {
	if event.Type != string(EventTypeBestellungAufgenommenV1) {
		return Bestellung{}, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	tischID, err := ParseTischIDFromSubject(event.Subject)
	if err != nil {
		return Bestellung{}, err
	}

	data := BestellungAufgenommenV1Data{}
	err = e.ParseData(event, &data, bestellungAufgenommenV1DataSchema)
	if err != nil {
		return Bestellung{}, err
	}

	bestellung := Bestellung{
		ID:               data.BestellungID,
		UserID:           event.UserID,
		UserName:         event.UserName,
		TischID:          tischID,
		Positionen:       fromPositionenEventData(data.Positionen),
		GesamtPreisCents: data.GesamtPreisCents,
		Kommentar:        data.Kommentar,
		AufgenommenAm:    event.Time,
	}

	if err := bestellungSchema.Validate(&bestellung); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return Bestellung{}, fmt.Errorf("bestellung validation failed: %v", issues)
	}

	return bestellung, nil
}

func buildZahlungFromEvent(event e.Event) (Zahlung, error) {
	if event.Type != string(EventTypeZahlungKassiertV1) {
		return Zahlung{}, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	tischID, err := ParseTischIDFromSubject(event.Subject)
	if err != nil {
		return Zahlung{}, err
	}

	data := ZahlungKassiertV1Data{}
	err = e.ParseData(event, &data, zahlungKassiertV1DataSchema)
	if err != nil {
		return Zahlung{}, err
	}

	zahlung := Zahlung{
		ID:                 data.ZahlungID,
		UserID:             event.UserID,
		UserName:           event.UserName,
		TischID:            tischID,
		Positionen:         fromPositionenEventData(data.Positionen),
		GesamtZahlungCents: data.GesamtZahlungCents,
		Kommentar:          data.Kommentar,
		KassiertAm:         event.Time,
	}

	if err := zahlungSchema.Validate(&zahlung); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return Zahlung{}, fmt.Errorf("zahlung validation failed: %v", issues)
	}

	return zahlung, nil
}

func buildStornierungFromEvent(event e.Event) (Stornierung, error) {
	if event.Type != string(EventTypeStornierungErteiltV1) {
		return Stornierung{}, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	tischID, err := ParseTischIDFromSubject(event.Subject)
	if err != nil {
		return Stornierung{}, err
	}

	data := StornierungErteiltV1Data{}
	err = e.ParseData(event, &data, stornierungErteiltV1DataSchema)
	if err != nil {
		return Stornierung{}, err
	}

	stornierung := Stornierung{
		ID:                     data.StornierungID,
		UserID:                 event.UserID,
		UserName:               event.UserName,
		TischID:                tischID,
		Positionen:             fromPositionenEventData(data.Positionen),
		GesamtStornierungCents: data.GesamtStornierungCents,
		Kommentar:              data.Kommentar,
		StorniertAm:            event.Time,
		BarRueckgabe:           true,
	}

	if err := stornierungSchema.Validate(&stornierung); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return Stornierung{}, fmt.Errorf("stornierung validation failed: %v", issues)
	}

	return stornierung, nil
}

// buildKorrekturFromEvent baut die geldneutrale Korrektur als Stornierung auf. In
// Historie und UI erscheinen beide Storno-Arten (kassenwirksame Warenrücknahme und
// geldneutrale Korrektur) als „Stornierung", werden aber über das abgeleitete Feld
// BarRueckgabe (hier false) sichtbar unterschieden. Die Daten wurden bei der
// Event-Erstellung validiert, daher keine erneute Schema-Prüfung (der Kommentar ist
// hier optional, anders als bei der Warenrücknahme).
func buildKorrekturFromEvent(event e.Event) (Stornierung, error) {
	if event.Type != string(EventTypeBestellungKorrigiertV1) {
		return Stornierung{}, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	tischID, err := ParseTischIDFromSubject(event.Subject)
	if err != nil {
		return Stornierung{}, err
	}

	data := BestellungKorrigiertV1Data{}
	err = e.ParseData(event, &data, bestellungKorrigiertV1DataSchema)
	if err != nil {
		return Stornierung{}, err
	}

	return Stornierung{
		ID:                     data.KorrekturID,
		UserID:                 event.UserID,
		UserName:               event.UserName,
		TischID:                tischID,
		Positionen:             fromPositionenEventData(data.Positionen),
		GesamtStornierungCents: data.GesamtCents,
		Kommentar:              data.Kommentar,
		StorniertAm:            event.Time,
		BarRueckgabe:           false,
	}, nil
}

func buildUmbuchungFromEvent(event e.Event) (Umbuchung, error) {
	if event.Type != string(EventTypeBestellungUmgebuchtV1) {
		return Umbuchung{}, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	tischID, err := ParseTischIDFromSubject(event.Subject)
	if err != nil {
		return Umbuchung{}, err
	}

	data := BestellungUmgebuchtV1Data{}
	err = e.ParseData(event, &data, bestellungUmgebuchtV1DataSchema)
	if err != nil {
		return Umbuchung{}, err
	}

	umbuchung := Umbuchung{
		ID:                data.UmbuchungID,
		UserID:            event.UserID,
		UserName:          event.UserName,
		TischID:           tischID,
		QuellTischID:      data.QuellTischID,
		ZielTischID:       data.ZielTischID,
		Positionen:        fromPositionenEventData(data.Positionen),
		GesamtCents:       data.GesamtCents,
		Kommentar:         data.Kommentar,
		BenutzerKommentar: data.BenutzerKommentar,
		UmgebuchtAm:       event.Time,
	}

	if err := umbuchungSchema.Validate(&umbuchung); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return Umbuchung{}, fmt.Errorf("umbuchung validation failed: %v", issues)
	}

	return umbuchung, nil
}
