package kasse

import (
	"fmt"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/google/uuid"
	e "github.com/nicograef/jotti/backend/domain/event"
)

const (
	EventTypeKassensitzungEroeffnetV1   EventType = "kassensitzung-eroeffnet:v1"
	EventTypeGeldtransitGebuchtV1       EventType = "geldtransit-gebucht:v1"
	EventTypeKassensturzDurchgefuehrtV1 EventType = "kassensturz-durchgefuehrt:v1"
	EventTypeDifferenzSollIstGebuchtV1  EventType = "differenz-soll-ist-gebucht:v1"
	EventTypeTagesabschlussErstelltV1   EventType = "tagesabschluss-erstellt:v1"
)

// --- Event-Data-Structs ---

type kassensitzungEroeffnetV1Data struct {
	Datum        string `json:"datum"`
	Bezeichnung  string `json:"bezeichnung"`
	BetragCents  int    `json:"betragCents"`
	EroeffnetVon int    `json:"eroeffnetVon"`
}

var kassensitzungEroeffnetV1DataSchema = z.Struct(z.Shape{
	"Datum":        z.String().Min(8).Max(10).Required(),
	"Bezeichnung":  z.String().Min(1).Max(200).Required(),
	"BetragCents":  z.Int().GTE(0).Required(),
	"EroeffnetVon": z.Int().GTE(1).Required(),
})

type geldtransitGebuchtV1Data struct {
	BewegungID  string `json:"bewegungId"`
	Richtung    string `json:"richtung"` // "einlage" | "entnahme"
	BetragCents int    `json:"betragCents"`
	Kommentar   string `json:"kommentar"`
	GebuchtVon  int    `json:"gebuchtVon"`
}

var geldtransitGebuchtV1DataSchema = z.Struct(z.Shape{
	"BewegungID":  z.String().UUID().Required(),
	"Richtung":    z.String().OneOf([]string{"einlage", "entnahme"}, z.Message("Ungültige Richtung")).Required(),
	"BetragCents": z.Int().GTE(1).Required(),
	"Kommentar":   z.String().Min(3).Max(200).Required(),
	"GebuchtVon":  z.Int().GTE(1).Required(),
})

type kassensturzDurchgefuehrtV1Data struct {
	SollBestandCents int `json:"sollBestandCents"`
	IstBestandCents  int `json:"istBestandCents"`
	DifferenzCents   int `json:"differenzCents"`
	DurchgefuehrtVon int `json:"durchgefuehrtVon"`
}

var kassensturzDurchgefuehrtV1DataSchema = z.Struct(z.Shape{
	"SollBestandCents": z.Int(),
	"IstBestandCents":  z.Int().GTE(0).Required(),
	"DifferenzCents":   z.Int(),
	"DurchgefuehrtVon": z.Int().GTE(1).Required(),
})

type differenzSollIstGebuchtV1Data struct {
	BetragCents int `json:"betragCents"`
	GebuchtVon  int `json:"gebuchtVon"`
}

var differenzSollIstGebuchtV1DataSchema = z.Struct(z.Shape{
	"BetragCents": z.Int().Required(),
	"GebuchtVon":  z.Int().GTE(1).Required(),
})

type tagesabschlussErstelltV1Data struct {
	ZNr               int       `json:"zNr"`
	ZeitraumVon       time.Time `json:"zeitraumVon"`
	ZeitraumBis       time.Time `json:"zeitraumBis"`
	UmsatzGesamtCents int       `json:"umsatzGesamtCents"`
	StornierungCents  int       `json:"stornierungCents"`
	AuszahlungenCents int       `json:"auszahlungenCents"`
	GeldtransitCents  int       `json:"geldtransitCents"`
	ErstelltVon       int       `json:"erstelltVon"`
}

var tagesabschlussErstelltV1DataSchema = z.Struct(z.Shape{
	"ZNr":               z.Int().GTE(1).Required(),
	"ZeitraumVon":       z.Time().Required(),
	"ZeitraumBis":       z.Time().Required(),
	"UmsatzGesamtCents": z.Int(),
	"StornierungCents":  z.Int(),
	"AuszahlungenCents": z.Int(),
	"GeldtransitCents":  z.Int(),
	"ErstelltVon":       z.Int().GTE(1).Required(),
})

// --- Event-Erstellungsfunktionen ---

func NewKassensitzungEroeffnetEvent(subject string, userID int, userName string, datum string, bezeichnung string, betragCents int) (e.Event, error) {
	data := kassensitzungEroeffnetV1Data{
		Datum:        datum,
		Bezeichnung:  bezeichnung,
		BetragCents:  betragCents,
		EroeffnetVon: userID,
	}

	if err := kassensitzungEroeffnetV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("kassensitzung eroeffnet data validation failed: %v", issues)
	}

	event, err := e.New(userID, userName, string(EventTypeKassensitzungEroeffnetV1), subject, data)
	if err != nil {
		return e.Event{}, err
	}

	return event, nil
}

func NewGeldtransitGebuchtEvent(subject string, userID int, userName string, richtung string, betragCents int, kommentar string) (e.Event, error) {
	data := geldtransitGebuchtV1Data{
		BewegungID:  uuid.New().String(),
		Richtung:    richtung,
		BetragCents: betragCents,
		Kommentar:   kommentar,
		GebuchtVon:  userID,
	}

	if err := geldtransitGebuchtV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("geldtransit gebucht data validation failed: %v", issues)
	}

	event, err := e.New(userID, userName, string(EventTypeGeldtransitGebuchtV1), subject, data)
	if err != nil {
		return e.Event{}, err
	}

	return event, nil
}

func NewKassensturzDurchgefuehrtEvent(subject string, userID int, userName string, sollBestandCents int, istBestandCents int, differenzCents int) (e.Event, error) {
	data := kassensturzDurchgefuehrtV1Data{
		SollBestandCents: sollBestandCents,
		IstBestandCents:  istBestandCents,
		DifferenzCents:   differenzCents,
		DurchgefuehrtVon: userID,
	}

	if err := kassensturzDurchgefuehrtV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("kassensturz durchgefuehrt data validation failed: %v", issues)
	}

	event, err := e.New(userID, userName, string(EventTypeKassensturzDurchgefuehrtV1), subject, data)
	if err != nil {
		return e.Event{}, err
	}

	return event, nil
}

func NewDifferenzSollIstGebuchtEvent(subject string, userID int, userName string, betragCents int) (e.Event, error) {
	data := differenzSollIstGebuchtV1Data{
		BetragCents: betragCents,
		GebuchtVon:  userID,
	}

	if err := differenzSollIstGebuchtV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("differenz soll-ist gebucht data validation failed: %v", issues)
	}

	event, err := e.New(userID, userName, string(EventTypeDifferenzSollIstGebuchtV1), subject, data)
	if err != nil {
		return e.Event{}, err
	}

	return event, nil
}

func NewTagesabschlussErstelltEvent(subject string, userID int, userName string, zNr int, zeitraumVon time.Time, zeitraumBis time.Time, umsatzGesamtCents int, stornierungCents int, auszahlungenCents int, geldtransitCents int) (e.Event, error) {
	data := tagesabschlussErstelltV1Data{
		ZNr:               zNr,
		ZeitraumVon:       zeitraumVon,
		ZeitraumBis:       zeitraumBis,
		UmsatzGesamtCents: umsatzGesamtCents,
		StornierungCents:  stornierungCents,
		AuszahlungenCents: auszahlungenCents,
		GeldtransitCents:  geldtransitCents,
		ErstelltVon:       userID,
	}

	if err := tagesabschlussErstelltV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("tagesabschluss erstellt data validation failed: %v", issues)
	}

	event, err := e.New(userID, userName, string(EventTypeTagesabschlussErstelltV1), subject, data)
	if err != nil {
		return e.Event{}, err
	}

	return event, nil
}
