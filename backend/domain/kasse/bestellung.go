package kasse

import (
	"strings"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/steuer"
)

type Position struct {
	PositionID   string
	VarianteID   int
	ProduktName  string
	VarianteName string
	Kategorie    string
	Steuersatz   string
	Einzelpreis  int
	Menge        int
	// BestellerUserID und BestellerName sind reine Projektions-/Anzeigefelder:
	// die Servicekraft, die die Bestellung aufgenommen hat. Sie werden beim
	// Anwenden des bestellung-aufgenommen-Events aus dem Event-Umschlag getagt
	// (eingefrorener Username) und nicht in der Event-Form serialisiert.
	BestellerUserID int
	BestellerName   string
}

// Bezeichnung is the canonical position name: product name and variant name
// joined by a single space, trimmed at the edges. No brackets, no dedup —
// the single place in the backend that composes these two fields.
func (p Position) Bezeichnung() string {
	return strings.TrimSpace(p.ProduktName + " " + p.VarianteName)
}

// PositionEventData is the serialization-friendly representation of Position for the event store.
// The json-keys are stable and must not be changed (immutable events).
type PositionEventData struct {
	PositionID   string `json:"positionId"`
	VarianteID   int    `json:"varianteId"`
	ProduktName  string `json:"produktName"`
	VarianteName string `json:"varianteName"`
	Kategorie    string `json:"kategorie"`
	Steuersatz   string `json:"steuersatz"`
	Einzelpreis  int    `json:"einzelpreis"`
	Menge        int    `json:"menge"`
}

// toPositionenEventData maps projection positions to their event form. The
// Besteller fields live only in the projection and are deliberately dropped here
// (the besteller is already recorded in the event envelope's UserID/UserName).
func toPositionenEventData(positionen []Position) []PositionEventData {
	out := make([]PositionEventData, len(positionen))
	for i, p := range positionen {
		out[i] = PositionEventData{
			PositionID:   p.PositionID,
			VarianteID:   p.VarianteID,
			ProduktName:  p.ProduktName,
			VarianteName: p.VarianteName,
			Kategorie:    p.Kategorie,
			Steuersatz:   p.Steuersatz,
			Einzelpreis:  p.Einzelpreis,
			Menge:        p.Menge,
		}
	}
	return out
}

// fromPositionenEventData maps event-form positions back to projection positions.
// The Besteller fields are left zero here; they are tagged from the event
// envelope when the bestellung-aufgenommen event is applied.
func fromPositionenEventData(positionen []PositionEventData) []Position {
	out := make([]Position, len(positionen))
	for i, p := range positionen {
		out[i] = PositionFromEventData(p)
	}
	return out
}

// PositionFromEventData maps a single event-form position to a projection
// position. The Besteller fields are left zero (the event form carries no
// besteller); it is the single source of truth for this mapping, used both
// inside the projection and by callers that read raw event positions.
func PositionFromEventData(p PositionEventData) Position {
	return Position{
		PositionID:   p.PositionID,
		VarianteID:   p.VarianteID,
		ProduktName:  p.ProduktName,
		VarianteName: p.VarianteName,
		Kategorie:    p.Kategorie,
		Steuersatz:   p.Steuersatz,
		Einzelpreis:  p.Einzelpreis,
		Menge:        p.Menge,
	}
}

var positionSchema = z.Struct(z.Shape{
	"PositionID":   z.String().UUID().Required(),
	"VarianteID":   produkt.IDSchema.Required(),
	"ProduktName":  produkt.NameSchema.Required(),
	"VarianteName": produkt.NameSchema.Required(),
	"Kategorie":    z.String().OneOf([]string{string(produkt.EssenKategorie), string(produkt.GetraenkKategorie), string(produkt.SonstigesKategorie)}, z.Message("Ungültige Kategorie")).Required(),
	"Steuersatz":   z.String().OneOf([]string{string(steuer.RegelSteuersatz), string(steuer.ErmaessigtSteuersatz), string(steuer.BefreitSteuersatz), string(steuer.KombiSteuersatz)}, z.Message("Ungültiger Steuersatz")).Required(),
	"Einzelpreis":  produkt.PreisCentsSchema.Required(),
	"Menge":        z.Int().GTE(1, z.Message("Menge muss mindestens 1 betragen")).Required(),
})

// PositionRef is a lightweight reference used in API request commands for payment/delivery/cancellation.
// Enriched to Position (fat) in the command layer before being stored in events.
type PositionRef struct {
	PositionID string
	Menge      int
}

type Bestellung struct {
	ID     string
	UserID int
	// UserName ist der eingefrorene Username der bestellenden Servicekraft aus
	// dem Event-Umschlag. Er beschriftet die Bestellung in der Storno-/Umbuch-
	// Historie, damit die Serviceleitung fremde Bestellungen ohne zusätzlichen
	// Klick findet. Spätere Umbenennungen ändern alte Einträge nicht.
	UserName         string
	TischID          int
	Positionen       []Position
	GesamtPreisCents int
	Kommentar        string
	AufgenommenAm    time.Time
}

var bestellungSchema = z.Struct(z.Shape{
	"ID":               z.String().UUID().Required(),
	"UserID":           z.Int().GTE(1).Required(),
	"UserName":         z.String().Min(1).Required(),
	"TischID":          z.Int().GTE(1).Required(),
	"Positionen":       z.Slice(positionSchema).Min(1).Required(),
	"GesamtPreisCents": z.Int().GTE(0).Required(),
	"Kommentar":        z.String().Max(100),
	"AufgenommenAm":    z.Time().Required(),
})
