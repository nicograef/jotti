package kasse

import (
	"slices"
	"strings"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/internal/zeichen"
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
	// Pfand marks a line of a Pfand-Produkt; only zahlung-kassiert:v2 and direktverkauf-getaetigt:v2
	// store it, every other event form reads it as false.
	Pfand bool
	// BestellerUserID und BestellerName sind reine Projektionsfelder: beim Anwenden des
	// bestellung-aufgenommen-Events aus dem Event-Umschlag getagt, nie mitserialisiert.
	BestellerUserID int
	BestellerName   string
	// Kommentar is the Positionskommentar; only bestellung-aufgenommen:v1, direktverkauf-getaetigt:v2
	// and the request input set it, every other event form reads it as empty.
	Kommentar string
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

// PositionV2EventData is the position of zahlung-kassiert:v2 and direktverkauf-getaetigt:v2: the
// v1 position plus the Pfand flag of its product at the time of sale. The json keys are frozen.
type PositionV2EventData struct {
	PositionEventData
	Pfand bool `json:"pfand"`
}

// KommentiertePositionEventData is the position of bestellung-aufgenommen:v1: the base form plus
// the optional Positionskommentar, omitted when empty so a position without it keeps the frozen bytes (D22).
type KommentiertePositionEventData struct {
	PositionEventData
	Kommentar string `json:"kommentar,omitempty"`
}

// KommentiertePositionV2EventData is the position of direktverkauf-getaetigt:v2: the v2 form plus
// the optional Positionskommentar, omitted when empty (D22).
type KommentiertePositionV2EventData struct {
	PositionV2EventData
	Kommentar string `json:"kommentar,omitempty"`
}

// PositionsKommentarSchema caps the Positionskommentar; event and request schemas share it.
var PositionsKommentarSchema = z.String().TestFunc(zeichen.Max(100))

func toKommentiertePositionenEventData(positionen []Position) []KommentiertePositionEventData {
	basis := toPositionenEventData(positionen)
	out := make([]KommentiertePositionEventData, len(positionen))
	for i, p := range positionen {
		out[i] = KommentiertePositionEventData{PositionEventData: basis[i], Kommentar: p.Kommentar}
	}
	return out
}

// PositionenMitKommentar maps the positions of bestellung-aufgenommen:v1 including their Kommentar.
func PositionenMitKommentar(positionen []KommentiertePositionEventData) []Position {
	out := make([]Position, len(positionen))
	for i, p := range positionen {
		out[i] = PositionFromEventData(p.PositionEventData)
		out[i].Kommentar = p.Kommentar
	}
	return out
}

// PositionenOhneKommentar is the base form for readers the Positionskommentar must not reach:
// projection, Storno split, TSE data and DSFinV-K export.
func PositionenOhneKommentar(positionen []KommentiertePositionEventData) []PositionEventData {
	out := make([]PositionEventData, len(positionen))
	for i, p := range positionen {
		out[i] = p.PositionEventData
	}
	return out
}

func toPositionenV2EventData(positionen []Position) []PositionV2EventData {
	basis := toPositionenEventData(positionen)
	out := make([]PositionV2EventData, len(positionen))
	for i, p := range positionen {
		out[i] = PositionV2EventData{PositionEventData: basis[i], Pfand: p.Pfand}
	}
	return out
}

func toKommentiertePositionenV2EventData(positionen []Position) []KommentiertePositionV2EventData {
	basis := toPositionenV2EventData(positionen)
	out := make([]KommentiertePositionV2EventData, len(positionen))
	for i, p := range positionen {
		out[i] = KommentiertePositionV2EventData{PositionV2EventData: basis[i], Kommentar: p.Kommentar}
	}
	return out
}

// PositionenV2MitKommentar maps the positions of direktverkauf-getaetigt:v2 including their Kommentar.
func PositionenV2MitKommentar(positionen []KommentiertePositionV2EventData) []Position {
	out := make([]Position, len(positionen))
	for i, p := range positionen {
		out[i] = PositionFromEventData(p.PositionEventData)
		out[i].Pfand = p.Pfand
		out[i].Kommentar = p.Kommentar
	}
	return out
}

// PositionenV2OhneKommentar is the v2 form with the Pfand flag for readers the Positionskommentar
// must not reach: the DSFinV-K export.
func PositionenV2OhneKommentar(positionen []KommentiertePositionV2EventData) []PositionV2EventData {
	out := make([]PositionV2EventData, len(positionen))
	for i, p := range positionen {
		out[i] = p.PositionV2EventData
	}
	return out
}

func positionenV1(positionen []PositionV2EventData) []PositionEventData {
	out := make([]PositionEventData, len(positionen))
	for i, p := range positionen {
		out[i] = p.PositionEventData
	}
	return out
}

// VarianteIDs lists the VarianteID of each position, in order and with duplicates.
func VarianteIDs(positionen []Position) []int {
	ids := make([]int, len(positionen))
	for i, p := range positionen {
		ids[i] = p.VarianteID
	}
	return ids
}

// MitPfandKennzeichen sets Pfand on each position whose variant belongs to a Pfand-Produkt.
func MitPfandKennzeichen(positionen []Position, pfandVarianten map[int]bool) []Position {
	out := slices.Clone(positionen)
	for i := range out {
		out[i].Pfand = pfandVarianten[out[i].VarianteID]
	}
	return out
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

// PositionenFromEventData leaves the Besteller fields zero; they are tagged from the event
// envelope when the bestellung-aufgenommen event is applied.
func PositionenFromEventData(positionen []PositionEventData) []Position {
	out := make([]Position, len(positionen))
	for i, p := range positionen {
		out[i] = PositionFromEventData(p)
	}
	return out
}

// SteuermatrixPositionen feeds steuer.Steuermatrix: the gross amount and rate of each position.
func SteuermatrixPositionen(positionen []Position) []steuer.SteuermatrixPosition {
	out := make([]steuer.SteuermatrixPosition, len(positionen))
	for i, p := range positionen {
		out[i] = steuer.SteuermatrixPosition{
			Brutto:     p.EinzelpreisCents * p.Menge,
			Steuersatz: steuer.Steuersatz(p.Steuersatz),
		}
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

// PositionEingabeSchema caps Menge on input only: the cap keeps `EinzelpreisCents * Menge` from
// overflowing int, and in positionSchema, which also validates stored events, it would make them
// unreadable. It is already required; never call `.Required()` again, since zog mutates in place.
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

var kommentiertePositionSchema = positionSchema.Extend(z.Shape{"Kommentar": PositionsKommentarSchema})

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
	"UserName":         z.String().TestFunc(zeichen.Min(1)).Required(),
	"TischID":          z.Int().GTE(1).Required(),
	"Positionen":       z.Slice(positionSchema).Min(1).Required(),
	"GesamtPreisCents": z.Int().GTE(1).Required(),
	"Kommentar":        z.String().TestFunc(zeichen.Max(100)),
	"AufgenommenAm":    z.Time().Required(),
})
