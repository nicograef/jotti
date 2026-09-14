package kasse

import "time"

type KassensitzungStatus string

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

// Kassenbestand ist der Soll-Kassenbestand einer Kassensitzung, eine reine Projektion
// des Kassenjournals. Solange keine Differenz gebucht ist, gilt:
//
//	AnfangsbestandCents + BareinnahmenCents + EinlagenCents − EntnahmenCents = SollBestandCents.
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
