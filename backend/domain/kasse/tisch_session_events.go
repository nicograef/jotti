package kasse

import (
	"fmt"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

type EventType string

const (
	EventTypeBestellungAufgenommenV1 EventType = "bestellung-aufgenommen:v1"
	EventTypeZahlungKassiertV1       EventType = "zahlung-kassiert:v1"
	EventTypeStornierungErteiltV1    EventType = "stornierung-erteilt:v1"
	EventTypeAusgabeBestaetigtV1     EventType = "ausgabe-bestaetigt:v1"
	EventTypeAuszahlungGeleistetV1   EventType = "auszahlung-geleistet:v1"
)

// --- Event-Data-Structs ---

type bestellungAufgenommenV1Data struct {
	BestellungID     string              `json:"bestellungId"`
	Positionen       []positionEventData `json:"positionen"`
	GesamtPreisCents int                 `json:"gesamtPreisCents"`
	Kommentar        string              `json:"kommentar"`
}

var bestellungAufgenommenV1DataSchema = z.Struct(z.Shape{
	"BestellungID":     z.String().UUID().Required(),
	"Positionen":       z.Slice(positionSchema).Min(1).Required(),
	"GesamtPreisCents": z.Int().GTE(0).Required(),
	"Kommentar":        z.String().Max(100),
})

type zahlungKassiertV1Data struct {
	ZahlungID          string              `json:"zahlungId"`
	Positionen         []positionEventData `json:"positionen"`
	GesamtZahlungCents int                 `json:"gesamtZahlungCents"`
	Kommentar          string              `json:"kommentar"`
	TSEData            *TSEData            `json:"tseData,omitempty"`
}

var zahlungKassiertV1DataSchema = z.Struct(z.Shape{
	"ZahlungID":          z.String().UUID().Required(),
	"Positionen":         z.Slice(positionSchema).Min(1).Required(),
	"GesamtZahlungCents": z.Int().GTE(0).Required(),
	"Kommentar":          z.String().Max(100),
})

type stornierungErteiltV1Data struct {
	StornierungID          string              `json:"stornierungId"`
	Positionen             []positionEventData `json:"positionen"`
	GesamtStornierungCents int                 `json:"gesamtStornierungCents"`
	Kommentar              string              `json:"kommentar"`
}

var stornierungErteiltV1DataSchema = z.Struct(z.Shape{
	"StornierungID":          z.String().UUID().Required(),
	"Positionen":             z.Slice(positionSchema).Min(1).Required(),
	"GesamtStornierungCents": z.Int().GTE(0).Required(),
	"Kommentar":              z.String().Min(3).Max(100).Required(),
})

type ausgabeBestaetigtV1Data struct {
	AusgabeID  string              `json:"ausgabeId"`
	Positionen []positionEventData `json:"positionen"`
	Kommentar  string              `json:"kommentar"`
}

var ausgabeBestaetigtV1DataSchema = z.Struct(z.Shape{
	"AusgabeID":  z.String().UUID().Required(),
	"Positionen": z.Slice(positionSchema).Min(1).Required(),
	"Kommentar":  z.String().Max(100),
})

type auszahlungGeleistetV1Data struct {
	AuszahlungID string `json:"auszahlungId"`
	BetragCents  int    `json:"betragCents"`
	Kommentar    string `json:"kommentar"`
}

var auszahlungGeleistetV1DataSchema = z.Struct(z.Shape{
	"AuszahlungID": z.String().UUID().Required(),
	"BetragCents":  z.Int().GTE(1).Required(),
	"Kommentar":    z.String().Min(3).Max(100).Required(),
})

// --- Event-Erstellungsfunktionen ---

func NewBestellungAufgenommenEvent(subject string, userID int, userName string, positionen []Position, kommentar string) (e.Event, error) {
	// Generate PositionIDs for each position
	for i := range positionen {
		positionen[i].PositionID = uuid.New().String()
	}

	gesamtPreisCents := 0
	for _, pos := range positionen {
		gesamtPreisCents += pos.Einzelpreis * pos.Menge
	}

	data := bestellungAufgenommenV1Data{
		BestellungID:     uuid.New().String(),
		Positionen:       toPositionenEventData(positionen),
		GesamtPreisCents: gesamtPreisCents,
		Kommentar:        kommentar,
	}

	if err := bestellungAufgenommenV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("bestellung aufgenommen data validation failed: %v", issues)
	}

	event, err := e.New(userID, userName, string(EventTypeBestellungAufgenommenV1), subject, data)
	if err != nil {
		return e.Event{}, err
	}

	return event, nil
}

func NewZahlungKassiertEvent(subject string, userID int, userName string, positionen []Position, gesamtZahlungCents int, kommentar string) (e.Event, error) {
	return newZahlungKassiertEvent(subject, userID, userName, positionen, gesamtZahlungCents, kommentar, nil)
}

func NewZahlungKassiertEventMitTSE(subject string, userID int, userName string, positionen []Position, gesamtZahlungCents int, kommentar string, tseData *TSEData) (e.Event, error) {
	return newZahlungKassiertEvent(subject, userID, userName, positionen, gesamtZahlungCents, kommentar, tseData)
}

func newZahlungKassiertEvent(subject string, userID int, userName string, positionen []Position, gesamtZahlungCents int, kommentar string, tseData *TSEData) (e.Event, error) {
	if tseData != nil {
		if err := tseData.Validate(); err != nil {
			return e.Event{}, err
		}
	}

	data := zahlungKassiertV1Data{
		ZahlungID:          uuid.New().String(),
		Positionen:         toPositionenEventData(positionen),
		GesamtZahlungCents: gesamtZahlungCents,
		Kommentar:          kommentar,
		TSEData:            tseData,
	}

	if err := zahlungKassiertV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("zahlung kassiert data validation failed: %v", issues)
	}

	event, err := e.New(userID, userName, string(EventTypeZahlungKassiertV1), subject, data)
	if err != nil {
		return e.Event{}, err
	}

	return event, nil
}

func NewStornierungErteiltEvent(subject string, userID int, userName string, positionen []Position, gesamtStornierungCents int, kommentar string) (e.Event, error) {
	data := stornierungErteiltV1Data{
		StornierungID:          uuid.New().String(),
		Positionen:             toPositionenEventData(positionen),
		GesamtStornierungCents: gesamtStornierungCents,
		Kommentar:              kommentar,
	}

	if err := stornierungErteiltV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("stornierung erteilt data validation failed: %v", issues)
	}

	event, err := e.New(userID, userName, string(EventTypeStornierungErteiltV1), subject, data)
	if err != nil {
		return e.Event{}, err
	}

	return event, nil
}

func NewAusgabeBestaetigtEvent(subject string, userID int, userName string, positionen []Position, kommentar string) (e.Event, error) {
	data := ausgabeBestaetigtV1Data{
		AusgabeID:  uuid.New().String(),
		Positionen: toPositionenEventData(positionen),
		Kommentar:  kommentar,
	}

	if err := ausgabeBestaetigtV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("ausgabe bestaetigt data validation failed: %v", issues)
	}

	event, err := e.New(userID, userName, string(EventTypeAusgabeBestaetigtV1), subject, data)
	if err != nil {
		return e.Event{}, err
	}

	return event, nil
}

func NewAuszahlungGeleistetEvent(subject string, userID int, userName string, betragCents int, kommentar string) (e.Event, error) {
	data := auszahlungGeleistetV1Data{
		AuszahlungID: uuid.New().String(),
		BetragCents:  betragCents,
		Kommentar:    kommentar,
	}

	if err := auszahlungGeleistetV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("auszahlung geleistet data validation failed: %v", issues)
	}

	event, err := e.New(userID, userName, string(EventTypeAuszahlungGeleistetV1), subject, data)
	if err != nil {
		return e.Event{}, err
	}

	return event, nil
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

	data := bestellungAufgenommenV1Data{}
	err = e.ParseData(event, &data, bestellungAufgenommenV1DataSchema)
	if err != nil {
		return Bestellung{}, err
	}

	bestellung := Bestellung{
		ID:               data.BestellungID,
		UserID:           event.UserID,
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

	data := zahlungKassiertV1Data{}
	err = e.ParseData(event, &data, zahlungKassiertV1DataSchema)
	if err != nil {
		return Zahlung{}, err
	}

	zahlung := Zahlung{
		ID:                 data.ZahlungID,
		UserID:             event.UserID,
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

	data := stornierungErteiltV1Data{}
	err = e.ParseData(event, &data, stornierungErteiltV1DataSchema)
	if err != nil {
		return Stornierung{}, err
	}

	stornierung := Stornierung{
		ID:                     data.StornierungID,
		UserID:                 event.UserID,
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

func buildAusgabeFromEvent(event e.Event) (Ausgabe, error) {
	if event.Type != string(EventTypeAusgabeBestaetigtV1) {
		return Ausgabe{}, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	tischID, err := ParseTischIDFromSubject(event.Subject)
	if err != nil {
		return Ausgabe{}, err
	}

	data := ausgabeBestaetigtV1Data{}
	err = e.ParseData(event, &data, ausgabeBestaetigtV1DataSchema)
	if err != nil {
		return Ausgabe{}, err
	}

	ausgabe := Ausgabe{
		ID:           data.AusgabeID,
		UserID:       event.UserID,
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

	data := auszahlungGeleistetV1Data{}
	err = e.ParseData(event, &data, auszahlungGeleistetV1DataSchema)
	if err != nil {
		return Auszahlung{}, err
	}

	auszahlung := Auszahlung{
		ID:          data.AuszahlungID,
		UserID:      event.UserID,
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
