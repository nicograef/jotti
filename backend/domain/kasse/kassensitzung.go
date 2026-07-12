package kasse

import "time"

type KassensitzungStatus string

// Kassensitzung represents the CRUD entity for a Kassensitzung.
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
	// KassensitzungWirdAbgeschlossen ist der transiente Zwischenstatus während des
	// Abschlusses: Ab ihm lehnt der Status-Guard alle Buchungs-Events ab, nur die
	// Abschluss-Events selbst dürfen noch geschrieben werden.
	KassensitzungWirdAbgeschlossen KassensitzungStatus = "wird_abgeschlossen"
	KassensitzungAbgeschlossen     KassensitzungStatus = "abgeschlossen"
)

// Kassenbestand ist der Soll-Kassenbestand einer Kassensitzung mitsamt seiner
// Aufschlüsselung. Reine Projektion des Kassenjournals. Es gilt (vor dem
// Kassensturz, also solange keine Differenz gebucht ist):
//
//	AnfangsbestandCents + BareinnahmenCents + EinlagenCents − EntnahmenCents = SollBestandCents.
type Kassenbestand struct {
	SollBestandCents    int
	AnfangsbestandCents int
	BareinnahmenCents   int
	EinlagenCents       int
	EntnahmenCents      int
}

// Geldtransit ist eine einzelne, gebuchte Bargeldbewegung (Einlage/Entnahme)
// einer Kassensitzung — die Anzeigeform der geldtransit-gebucht:v1-Events für die
// Bewegungsliste. GebuchtVon ist der eingefrorene Anzeigename aus dem Kassenjournal.
type Geldtransit struct {
	Zeitpunkt   time.Time
	Richtung    string
	BetragCents int
	Kommentar   string
	GebuchtVon  string
}
