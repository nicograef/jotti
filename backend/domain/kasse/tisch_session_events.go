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
	EventTypeAusgabeBestaetigtV1     EventType = "ausgabe-bestaetigt:v1"
	EventTypeAuszahlungGeleistetV1   EventType = "auszahlung-geleistet:v1"
)

// --- Event-Data-Structs ---

type BestellungAufgenommenV1Data struct {
	BestellungID     string              `json:"bestellungId"`
	Positionen       []PositionEventData `json:"positionen"`
	GesamtPreisCents int                 `json:"gesamtPreisCents"`
	Kommentar        string              `json:"kommentar"`
	TSETxID          string              `json:"tseTxId,omitempty"`
	TSEData          *TSEData            `json:"tseData,omitempty"`
}

var bestellungAufgenommenV1DataSchema = z.Struct(z.Shape{
	"BestellungID":     z.String().UUID().Required(),
	"Positionen":       z.Slice(positionSchema).Min(1).Required(),
	"GesamtPreisCents": z.Int().GTE(0).Required(),
	"Kommentar":        z.String().Max(100),
})

type ZahlungKassiertV1Data struct {
	ZahlungID          string              `json:"zahlungId"`
	Positionen         []PositionEventData `json:"positionen"`
	GesamtZahlungCents int                 `json:"gesamtZahlungCents"`
	Kommentar          string              `json:"kommentar"`
	TSETxID            string              `json:"tseTxId,omitempty"`
	TSEData            *TSEData            `json:"tseData,omitempty"`
	TSEAusfall         bool                `json:"tseAusfall,omitempty"`
}

var zahlungKassiertV1DataSchema = z.Struct(z.Shape{
	"ZahlungID":          z.String().UUID().Required(),
	"Positionen":         z.Slice(positionSchema).Min(1).Required(),
	"GesamtZahlungCents": z.Int().GTE(0).Required(),
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
	TSETxID                string              `json:"tseTxId,omitempty"`
	TSEData                *TSEData            `json:"tseData,omitempty"`
}

var stornierungErteiltV1DataSchema = z.Struct(z.Shape{
	"StornierungID":          z.String().UUID().Required(),
	"ZahlungID":              z.String().UUID().Required(),
	"Positionen":             z.Slice(positionSchema).Min(1).Required(),
	"GesamtStornierungCents": z.Int().GTE(0).Required(),
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
	TSETxID     string              `json:"tseTxId,omitempty"`
	TSEData     *TSEData            `json:"tseData,omitempty"`
}

var bestellungKorrigiertV1DataSchema = z.Struct(z.Shape{
	"KorrekturID": z.String().UUID().Required(),
	"Positionen":  z.Slice(positionSchema).Min(1).Required(),
	"GesamtCents": z.Int().GTE(0).Required(),
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
	Kommentar    string              `json:"kommentar"`
	TSETxID      string              `json:"tseTxId,omitempty"`
	TSEData      *TSEData            `json:"tseData,omitempty"`
}

var bestellungUmgebuchtV1DataSchema = z.Struct(z.Shape{
	"UmbuchungID":  z.String().UUID().Required(),
	"QuellTischID": z.Int().GTE(1).Required(),
	"ZielTischID":  z.Int().GTE(1).Required(),
	"Positionen":   z.Slice(positionSchema).Min(1).Required(),
	"GesamtCents":  z.Int().GTE(0).Required(),
	"Kommentar":    z.String().Max(100),
})

type AusgabeBestaetigtV1Data struct {
	AusgabeID  string              `json:"ausgabeId"`
	Positionen []PositionEventData `json:"positionen"`
	Kommentar  string              `json:"kommentar"`
}

var ausgabeBestaetigtV1DataSchema = z.Struct(z.Shape{
	"AusgabeID":  z.String().UUID().Required(),
	"Positionen": z.Slice(positionSchema).Min(1).Required(),
	"Kommentar":  z.String().Max(100),
})

type AuszahlungGeleistetV1Data struct {
	AuszahlungID string   `json:"auszahlungId"`
	BetragCents  int      `json:"betragCents"`
	Kommentar    string   `json:"kommentar"`
	TSETxID      string   `json:"tseTxId,omitempty"`
	TSEData      *TSEData `json:"tseData,omitempty"`
}

var auszahlungGeleistetV1DataSchema = z.Struct(z.Shape{
	"AuszahlungID": z.String().UUID().Required(),
	"BetragCents":  z.Int().GTE(1).Required(),
	"Kommentar":    z.String().Min(3).Max(100).Required(),
})

// --- Event-Erstellungsfunktionen ---

func NewBestellungAufgenommenEvent(subject string, userID int, userName string, positionen []Position, kommentar string) (e.Event, error) {
	// Generate PositionIDs for each position (on a copy, so the caller's slice stays untouched)
	positionen = slices.Clone(positionen)
	for i := range positionen {
		positionen[i].PositionID = uuid.New().String()
	}

	gesamtPreisCents := 0
	for _, pos := range positionen {
		gesamtPreisCents += pos.Einzelpreis * pos.Menge
	}

	data := BestellungAufgenommenV1Data{
		BestellungID:     uuid.New().String(),
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
func NewBestellungUmgebuchtEvents(zNr int, quellTischID int, zielTischID int, userID int, userName string, quellPositionen []Position, gesamtCents int, quellKommentar string, zielKommentar string) (e.Event, e.Event, error) {
	umbuchungID := uuid.New().String()

	zielPositionen := slices.Clone(quellPositionen)
	for i := range zielPositionen {
		zielPositionen[i].PositionID = uuid.New().String()
	}

	quellData := BestellungUmgebuchtV1Data{
		UmbuchungID:  umbuchungID,
		QuellTischID: quellTischID,
		ZielTischID:  zielTischID,
		Positionen:   toPositionenEventData(quellPositionen),
		GesamtCents:  gesamtCents,
		Kommentar:    quellKommentar,
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

func NewAusgabeBestaetigtEvent(subject string, userID int, userName string, positionen []Position, kommentar string) (e.Event, error) {
	data := AusgabeBestaetigtV1Data{
		AusgabeID:  uuid.New().String(),
		Positionen: toPositionenEventData(positionen),
		Kommentar:  kommentar,
	}

	if err := ausgabeBestaetigtV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("ausgabe bestaetigt data validation failed: %v", issues)
	}

	return e.New(userID, userName, string(EventTypeAusgabeBestaetigtV1), subject, data)
}

func NewAuszahlungGeleistetEvent(subject string, userID int, userName string, betragCents int, kommentar string) (e.Event, error) {
	data := AuszahlungGeleistetV1Data{
		AuszahlungID: uuid.New().String(),
		BetragCents:  betragCents,
		Kommentar:    kommentar,
	}

	if err := auszahlungGeleistetV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("auszahlung geleistet data validation failed: %v", issues)
	}

	return e.New(userID, userName, string(EventTypeAuszahlungGeleistetV1), subject, data)
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
	}

	if err := stornierungSchema.Validate(&stornierung); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return Stornierung{}, fmt.Errorf("stornierung validation failed: %v", issues)
	}

	return stornierung, nil
}

// buildKorrekturFromEvent baut die geldneutrale Korrektur als Stornierung auf: in
// Historie und UI erscheinen beide Storno-Arten (kassenwirksame Warenrücknahme und
// geldneutrale Korrektur) einheitlich als „Stornierung". Die Daten wurden bei der
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
		ID:           data.UmbuchungID,
		UserID:       event.UserID,
		UserName:     event.UserName,
		TischID:      tischID,
		QuellTischID: data.QuellTischID,
		ZielTischID:  data.ZielTischID,
		Positionen:   fromPositionenEventData(data.Positionen),
		GesamtCents:  data.GesamtCents,
		Kommentar:    data.Kommentar,
		UmgebuchtAm:  event.Time,
	}

	if err := umbuchungSchema.Validate(&umbuchung); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return Umbuchung{}, fmt.Errorf("umbuchung validation failed: %v", issues)
	}

	return umbuchung, nil
}

func buildAusgabeFromEvent(event e.Event) (Ausgabe, error) {
	if event.Type != string(EventTypeAusgabeBestaetigtV1) {
		return Ausgabe{}, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	tischID, err := ParseTischIDFromSubject(event.Subject)
	if err != nil {
		return Ausgabe{}, err
	}

	data := AusgabeBestaetigtV1Data{}
	err = e.ParseData(event, &data, ausgabeBestaetigtV1DataSchema)
	if err != nil {
		return Ausgabe{}, err
	}

	ausgabe := Ausgabe{
		ID:           data.AusgabeID,
		UserID:       event.UserID,
		UserName:     event.UserName,
		TischID:      tischID,
		Positionen:   fromPositionenEventData(data.Positionen),
		Kommentar:    data.Kommentar,
		AusgegebenAm: event.Time,
	}

	if err := ausgabeSchema.Validate(&ausgabe); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return Ausgabe{}, fmt.Errorf("ausgabe validation failed: %v", issues)
	}

	return ausgabe, nil
}

func buildAuszahlungFromEvent(event e.Event) (Auszahlung, error) {
	if event.Type != string(EventTypeAuszahlungGeleistetV1) {
		return Auszahlung{}, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	tischID, err := ParseTischIDFromSubject(event.Subject)
	if err != nil {
		return Auszahlung{}, err
	}

	data := AuszahlungGeleistetV1Data{}
	err = e.ParseData(event, &data, auszahlungGeleistetV1DataSchema)
	if err != nil {
		return Auszahlung{}, err
	}

	auszahlung := Auszahlung{
		ID:          data.AuszahlungID,
		UserID:      event.UserID,
		UserName:    event.UserName,
		TischID:     tischID,
		BetragCents: data.BetragCents,
		Kommentar:   data.Kommentar,
		GeleistetAm: event.Time,
	}

	if err := auszahlungSchema.Validate(&auszahlung); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return Auszahlung{}, fmt.Errorf("auszahlung validation failed: %v", issues)
	}

	return auszahlung, nil
}
