package kasse

import (
	"strings"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/steuer"
)

type Position struct {
	PositionID       string
	VarianteID       int
	ProduktName      string
	VarianteName     string
	Kategorie        string
	Steuersatz       string
	EinzelpreisCents int
	Menge            int
	// BestellerUserID und BestellerName sind reine Projektionsfelder: beim Anwenden des
	// bestellung-aufgenommen-Events aus dem Event-Umschlag getagt, nie mitserialisiert.
	BestellerUserID int
	BestellerName   string
}

// Bezeichnung is the canonical position name (product + variant, single space, trimmed) —
// the single place in the backend that composes these two fields. No brackets, no dedup.
func (p Position) Bezeichnung() string {
	return strings.TrimSpace(p.ProduktName + " " + p.VarianteName)
}

// PositionEventData is the event-store form of Position. The json keys are frozen —
// immutable events.
type PositionEventData struct {
	PositionID       string `json:"positionId"`
	VarianteID       int    `json:"varianteId"`
	ProduktName      string `json:"produktName"`
	VarianteName     string `json:"varianteName"`
	Kategorie        string `json:"kategorie"`
	Steuersatz       string `json:"steuersatz"`
	EinzelpreisCents int    `json:"einzelpreisCents"`
	Menge            int    `json:"menge"`
}

// toPositionenEventData drops the Besteller fields deliberately — the besteller is already
// recorded in the event envelope's UserID/UserName.
func toPositionenEventData(positionen []Position) []PositionEventData {
	out := make([]PositionEventData, len(positionen))
	for i, p := range positionen {
		out[i] = PositionEventData{
			PositionID:       p.PositionID,
			VarianteID:       p.VarianteID,
			ProduktName:      p.ProduktName,
			VarianteName:     p.VarianteName,
			Kategorie:        p.Kategorie,
			Steuersatz:       p.Steuersatz,
			EinzelpreisCents: p.EinzelpreisCents,
			Menge:            p.Menge,
		}
	}
	return out
}

// fromPositionenEventData leaves the Besteller fields zero; they are tagged from the event
// envelope when the bestellung-aufgenommen event is applied.
func fromPositionenEventData(positionen []PositionEventData) []Position {
	out := make([]Position, len(positionen))
	for i, p := range positionen {
		out[i] = PositionFromEventData(p)
	}
	return out
}

// PositionFromEventData is the single source of truth for this mapping; the Besteller
// fields stay zero (the event form carries no besteller).
func PositionFromEventData(p PositionEventData) Position {
	return Position{
		PositionID:       p.PositionID,
		VarianteID:       p.VarianteID,
		ProduktName:      p.ProduktName,
		VarianteName:     p.VarianteName,
		Kategorie:        p.Kategorie,
		Steuersatz:       p.Steuersatz,
		EinzelpreisCents: p.EinzelpreisCents,
		Menge:            p.Menge,
	}
}

// PositionEingabeSchema begrenzt die Menge nur auf dem Eingabeweg: Die Obergrenze schützt
// `EinzelpreisCents * Menge` vor dem int-Überlauf, der auf einen plausiblen Kleinbetrag
// zurückwickelt. In `positionSchema`, das auch jedes gelesene Event validiert, machte eine
// Grenze bestehende Events unlesbar. Das Schema ist per Definition required — nie erneut
// `.Required()` aufrufen (zog mutiert den Empfänger in place).
var PositionEingabeSchema = z.Int().
	GTE(1, z.Message("Menge muss mindestens 1 betragen")).
	LTE(999, z.Message("Menge zu hoch")).
	Required(z.Message("Menge muss mindestens 1 betragen"))

var positionSchema = z.Struct(z.Shape{
	"PositionID":       z.String().UUID().Required(),
	"VarianteID":       produkt.IDSchema.Required(),
	"ProduktName":      produkt.NameSchema.Required(),
	"VarianteName":     produkt.NameSchema.Required(),
	"Kategorie":        z.String().OneOf([]string{string(produkt.EssenKategorie), string(produkt.GetraenkKategorie), string(produkt.SonstigesKategorie)}, z.Message("Ungültige Kategorie")).Required(),
	"Steuersatz":       z.String().OneOf([]string{string(steuer.RegelSteuersatz), string(steuer.ErmaessigtSteuersatz), string(steuer.BefreitSteuersatz), string(steuer.KombiSteuersatz)}, z.Message("Ungültiger Steuersatz")).Required(),
	"EinzelpreisCents": produkt.PreisCentsSchema,
	"Menge":            z.Int().GTE(1, z.Message("Menge muss mindestens 1 betragen")).Required(),
})

// PositionRef is the lightweight request-side reference, enriched to a fat Position in the
// command layer before it is stored in an event.
type PositionRef struct {
	PositionID string
	Menge      int
}

type Bestellung struct {
	ID     string
	UserID int
	// UserName ist der eingefrorene Username der bestellenden Servicekraft aus dem
	// Event-Umschlag; spätere Umbenennungen ändern alte Einträge nicht.
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
	"GesamtPreisCents": z.Int().GTE(1).Required(),
	"Kommentar":        z.String().Max(100),
	"AufgenommenAm":    z.Time().Required(),
})
