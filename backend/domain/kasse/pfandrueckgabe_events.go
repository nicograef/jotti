package kasse

import (
	"math"

	z "github.com/Oudwins/zog"
	e "github.com/nicograef/jotti/backend/domain/event"
)

const (
	EventTypePfandZurueckgegebenV1     EventType = "pfand-zurueckgegeben:v1"
	EventTypePfandrueckgabeStorniertV1 EventType = "pfandrueckgabe-storniert:v1"
)

// PfandZurueckgegebenV1Data: the positions are positive; the event type carries the sign
// "Auszahlung". TischID 0 means Direktverkauf.
type PfandZurueckgegebenV1Data struct {
	RueckgabeID       string              `json:"rueckgabeId"`
	TischID           int                 `json:"tischId"`
	Positionen        []PositionEventData `json:"positionen"`
	GesamtbetragCents int                 `json:"gesamtbetragCents"`
	Kommentar         string              `json:"kommentar"`
}

var pfandZurueckgegebenV1DataSchema = z.Struct(z.Shape{
	"RueckgabeID": z.String().UUID().Required(),
	// Kein Required(): zog wertet 0 als fehlend, Tisch 0 ist aber der Direktverkauf.
	"TischID":           z.Int().GTE(0).LTE(math.MaxInt32),
	"Positionen":        z.Slice(positionSchema).Min(1).Required(),
	"GesamtbetragCents": z.Int().GTE(1).Required(),
	"Kommentar":         z.String().Max(100),
})

// PfandrueckgabeStorniertV1Data carries the TischID of the Rückgabe, so readers filter both
// types without looking up the Rückgabe.
type PfandrueckgabeStorniertV1Data struct {
	StornierungID          string              `json:"stornierungId"`
	RueckgabeID            string              `json:"rueckgabeId"`
	TischID                int                 `json:"tischId"`
	Positionen             []PositionEventData `json:"positionen"`
	GesamtStornierungCents int                 `json:"gesamtStornierungCents"`
	Kommentar              string              `json:"kommentar"`
}

var pfandrueckgabeStorniertV1DataSchema = z.Struct(z.Shape{
	"StornierungID":          z.String().UUID().Required(),
	"RueckgabeID":            z.String().UUID().Required(),
	"TischID":                z.Int().GTE(0).LTE(math.MaxInt32),
	"Positionen":             z.Slice(positionSchema).Min(1).Required(),
	"GesamtStornierungCents": z.Int().GTE(1).Required(),
	"Kommentar":              z.String().Min(3).Max(100).Required(),
})

func NewPfandZurueckgegebenEvent(subject string, rueckgabeID string, tischID int, userID int, userName string, positionen []Position, kommentar string) (e.Event, error) {
	positionen = mitNeuenPositionIDs(positionen)
	gesamtbetragCents := summeCents(positionen)

	data := PfandZurueckgegebenV1Data{
		RueckgabeID:       rueckgabeID,
		TischID:           tischID,
		Positionen:        toPositionenEventData(positionen),
		GesamtbetragCents: gesamtbetragCents,
		Kommentar:         kommentar,
	}

	if err := validateEventData(pfandZurueckgegebenV1DataSchema, &data, "pfand zurueckgegeben"); err != nil {
		return e.Event{}, err
	}

	return e.New(userID, userName, string(EventTypePfandZurueckgegebenV1), subject, data)
}

// NewPfandrueckgabeStorniertEvent: gesamtStornierungCents is computed by the caller from the
// not-yet-cancelled positions.
func NewPfandrueckgabeStorniertEvent(subject string, stornierungID string, rueckgabeID string, tischID int, userID int, userName string, positionen []Position, gesamtStornierungCents int, kommentar string) (e.Event, error) {
	data := PfandrueckgabeStorniertV1Data{
		StornierungID:          stornierungID,
		RueckgabeID:            rueckgabeID,
		TischID:                tischID,
		Positionen:             toPositionenEventData(positionen),
		GesamtStornierungCents: gesamtStornierungCents,
		Kommentar:              kommentar,
	}

	if err := validateEventData(pfandrueckgabeStorniertV1DataSchema, &data, "pfandrueckgabe storniert"); err != nil {
		return e.Event{}, err
	}

	return e.New(userID, userName, string(EventTypePfandrueckgabeStorniertV1), subject, data)
}
