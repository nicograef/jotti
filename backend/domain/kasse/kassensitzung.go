package kasse

import (
	"math"
	"time"

	z "github.com/Oudwins/zog"
)

type KassensitzungStatus string

// KassensitzungNrSchema bounds a Z-Nummer like an ID: the int4 column holds at most MaxInt32. Each
// call builds a fresh schema, because zog's Required() mutates its receiver and one field is optional.
func KassensitzungNrSchema() *z.NumberSchema[int] {
	return z.Int().
		GTE(1, z.Message("Ungültige Kassensitzungsnummer")).
		LTE(math.MaxInt32, z.Message("Ungültige Kassensitzungsnummer"))
}

type Kassensitzung struct {
	ZNr         int
	Datum       time.Time
	Bezeichnung string
	Status      KassensitzungStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

const (
	KassensitzungOffen KassensitzungStatus = "offen"
	// KassensitzungWirdAbgeschlossen ist der transiente Zwischenstatus: ab ihm lehnt der
	// Status-Guard alle Buchungs-Events ab, nur Abschluss-Events dürfen noch geschrieben werden.
	KassensitzungWirdAbgeschlossen KassensitzungStatus = "wird_abgeschlossen"
	KassensitzungAbgeschlossen     KassensitzungStatus = "abgeschlossen"
)

// Kassenbestand is the Soll-Kassenbestand of a Kassensitzung, projected from the Kassenjournal.
// Until a Differenz is booked, Anfangsbestand + Bareinnahmen + Einlagen − Entnahmen = SollBestand.
type Kassenbestand struct {
	SollBestandCents    int
	AnfangsbestandCents int
	BareinnahmenCents   int
	EinlagenCents       int
	EntnahmenCents      int
}

// SollBestandOhneDifferenzCents summiert die vier Komponenten. SollBestandCents zieht
// zusätzlich eine gebuchte Differenz ab (sie gleicht Soll an den gezählten Ist-Bestand an).
func (k Kassenbestand) SollBestandOhneDifferenzCents() int {
	return k.AnfangsbestandCents + k.BareinnahmenCents + k.EinlagenCents - k.EntnahmenCents
}

// Geldtransit ist eine gebuchte Bargeldbewegung (Einlage/Entnahme) aus den
// geldtransit-gebucht:v1-Events. GebuchtVon ist der eingefrorene Anzeigename.
type Geldtransit struct {
	Zeitpunkt   time.Time
	Richtung    string
	BetragCents int
	Kommentar   string
	GebuchtVon  string
}
