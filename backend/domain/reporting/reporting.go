package reporting

import (
	"time"

	"github.com/nicograef/jotti/backend/domain/steuer"
)

type UmsatzServicekraft struct {
	UserID          int
	UserName        string // eingefrorener Username
	Name            string // live aus users aufgeloester Klarname (nur Admin-Anzeige)
	ZahlungenCents  int
	AnzahlZahlungen int
}

type UmsatzSteuersatz struct {
	Satz        steuer.Steuersatz
	BruttoCents int
	NettoCents  int
	SteuerCents int
}

type StornierungPosition struct {
	ProduktName      string
	VarianteName     string
	Menge            int
	EinzelpreisCents int
}

// ServicekraftRef identifiziert eine Servicekraft in Reporting-Ausgaben:
// stabile Benutzer-ID, eingefrorener Username aus dem Event-Umschlag und der
// live aus users aufgeloeste Klarname (nur Admin-Anzeige, leer wenn der
// Benutzer fehlt). Geteilte Referenz fuer Akteur und Betroffene.
type ServicekraftRef struct {
	UserID   int
	UserName string
	Name     string
}

type StornierungDetail struct {
	Zeitpunkt    time.Time
	Quelle       string // "tisch" oder "direktverkauf"
	BarRueckgabe bool   // true bei kassenwirksamer Warenrücknahme, false bei geldneutraler Korrektur
	TischID      int
	TischName    string
	// Akteur ist die Servicekraft, die den Storno ausgelöst hat (Event-Umschlag).
	Akteur ServicekraftRef
	// Betroffene sind die Servicekräfte, deren Vorgang der Storno rückgängig
	// macht (Storno-Zuordnung): der Kassierer der zurückgenommenen Zahlung, der
	// Verkäufer des stornierten Direktverkaufs bzw. die Besteller der
	// korrigierten Positionen. Nie leer — ohne auflösbaren Verweis steht hier
	// der Akteur.
	Betroffene  []ServicekraftRef
	BetragCents int
	Kommentar   string
	Positionen  []StornierungPosition
}

// StornierungServicekraft aggregiert die Stornierungen einer Servicekraft
// (Anzahl und Betrag) über eine Kassensitzung — als Kontroll-Signal im
// Admin-Dashboard. Aus den StornierungDetail-Zeilen zusammengefasst.
type StornierungServicekraft struct {
	UserID              int
	UserName            string // eingefrorener Username
	Name                string // live aus users aufgeloester Klarname (nur Admin-Anzeige)
	AnzahlStornierungen int
	StornierungenCents  int
}

// ProduktStatistikZeile ist eine flache Ausgabezeile der GetProduktStatistik-Query:
// je Variante die ausgegebene Menge (Produktion) und der Umsatz (Einnahmen) einer
// Kassensitzung. Eingabe der Gruppierung (gruppiereProduktStatistik); erscheint
// nie direkt in einer Response.
type ProduktStatistikZeile struct {
	Kategorie        string
	ProduktName      string
	VarianteID       int
	VarianteName     string
	AusgegebeneMenge int
	UmsatzCents      int
}

// VarianteStatistik ist die aufbereitete Verkaufs-Kennzahl einer Variante:
// ausgegebene Menge und Umsatz, bewusst getrennte Grundlagen (nicht ineinander
// umrechenbar).
type VarianteStatistik struct {
	VarianteID       int
	VarianteName     string
	AusgegebeneMenge int
	UmsatzCents      int
}

// ProduktStatistik gruppiert die Varianten eines Produkts mit Zwischensummen
// über ausgegebene Menge und Umsatz. Ein-Varianten-Produkte tragen genau eine
// Variante; die Zusammenfassung zu einer Zeile ist reine Präsentation.
type ProduktStatistik struct {
	Kategorie        string
	ProduktName      string
	AusgegebeneMenge int
	UmsatzCents      int
	Varianten        []VarianteStatistik
}

type Summary struct {
	GesamtUmsatzCents        int
	GesamtBestellungenCents  int
	GesamtStornierungenCents int
	GeldtransitCents         int
	AnzahlBestellungen       int
	AnzahlStornierungen      int
	AnzahlDirektverkaeufe    int
	DirektverkaufUmsatzCents int
}

type Breakdowns struct {
	UmsatzProServicekraft        []UmsatzServicekraft
	StornierungenProServicekraft []StornierungServicekraft
}

// Metadaten sind die Sitzungs-Kopfdaten für den formalen Tagesbericht, rein aus
// den Journal-Events projiziert: Eröffnungs- und Abschlusszeitpunkt, der
// abschließende Benutzer und die Kassensturz-Differenz. Alle Felder sind
// optional, solange die zugehörigen Events fehlen (z. B. offene Sitzung).
type Metadaten struct {
	EroeffnetAm               *time.Time
	AbgeschlossenAm           *time.Time
	AbgeschlossenVon          string // eingefrorener user_name des Tagesabschluss-Events
	KassensturzDifferenzCents *int
}

type ReportingData struct {
	KassensitzungNr     int
	Metadaten           Metadaten
	Summary             Summary
	Breakdowns          Breakdowns
	UmsatzProSteuersatz []UmsatzSteuersatz
	Stornierungen       []StornierungDetail
	// ProduktStatistik ist die fertig gruppierte und sortierte Verkaufsstatistik
	// je Produkt/Variante (Kategorie-Abschnitte). Von der Anwendungsschicht aus
	// den flachen ProduktStatistikZeile-Werten gebaut.
	ProduktStatistik []ProduktStatistik
}

// AbgeschlosseneSitzung ist ein Eintrag der Kassenberichte-Sitzungsliste: die
// abgeschlossene Kassensitzung mit ihrem Gesamtumsatz und Abschlusszeitpunkt aus
// dem tagesabschluss-erstellt:v1-Event. AbgeschlossenAm ist optional, falls das
// Event ausnahmsweise fehlt.
type AbgeschlosseneSitzung struct {
	ZNr               int
	Datum             time.Time
	Bezeichnung       string
	UmsatzGesamtCents int
	AbgeschlossenAm   *time.Time
}

type OffenerTisch struct {
	TischID    int
	TischName  string
	SaldoCents int
}

type LiveReportingData struct {
	KassensitzungNr  int
	Bezeichnung      string
	Datum            time.Time
	OffeneTische     []OffenerTisch
	OffeneSaldiCents int
	Summary          Summary
	Breakdowns       Breakdowns
	// Servicekraefte ist die Live-Sicht pro Servicekraft: kassierter Umsatz
	// (aus Breakdowns.UmsatzProServicekraft) zusammengeführt mit der offenen
	// eigenen Arbeit über die Tisch-Sessions, per user_id gemerged.
	Servicekraefte []ServicekraftLive
	Stornierungen  []StornierungDetail
	// ProduktStatistik ist die fertig gruppierte und sortierte Verkaufsstatistik
	// der offenen Sitzung — identische Form und Regeln wie in der Abrechnung.
	ProduktStatistik []ProduktStatistik
}

// ServicekraftLive ist die Live-Sicht auf eine Servicekraft im Admin-Dashboard:
// ihr kassierter Umsatz, zusammengeführt mit ihrer offenen eigenen Arbeit (per
// user_id). Personen mit offener Arbeit, aber ohne kassierten Umsatz erscheinen
// ebenfalls (dann mit Null-Umsatz).
type ServicekraftLive struct {
	UserID          int
	UserName        string // eingefrorener Username
	Name            string // live aus users aufgeloester Klarname (leer bei reiner offener Arbeit)
	ZahlungenCents  int
	AnzahlZahlungen int
	// OffenCents ist der noch offene (unbezahlte) Betrag der eigenen Arbeit über
	// alle Tische — die Servicekraft-Ebene-Summe über OffeneTische.OffenCents.
	OffenCents int
	// OffeneTische listet die Tische mit offener eigener Arbeit (aufsteigend nach
	// Tisch-ID); leer wenn die Servicekraft fertig ist.
	OffeneTische []OffeneArbeitTisch
	// Erledigt ist true, wenn keine offene eigene Arbeit mehr besteht.
	Erledigt bool
}

// OffeneArbeitTisch ist die offene eigene Arbeit einer Servicekraft an einem
// Tisch, angereichert um den Tisch-Namen für die Anzeige.
type OffeneArbeitTisch struct {
	TischID     int
	TischName   string
	AnzahlOffen int
	// OffenCents ist der noch offene (unbezahlte) Betrag der eigenen Positionen
	// an diesem Tisch.
	OffenCents int
}

type EigeneUebersicht struct {
	AnzahlBestellungen int
	BestellungenCents  int
	AnzahlZahlungen    int
	ZahlungenCents     int
}
