package reporting

import (
	"time"

	"github.com/nicograef/jotti/backend/domain/steuer"
)

// AbrechnungServicekraft ist die Bargeld-Abrechnung des Tischservice für eine
// Servicekraft: AbzugebenCents = KassiertCents − RuecknahmenCents.
// AnzahlStornierungen zählt beide Tisch-Storno-Arten (Rücknahmen und
// geldneutrale Korrekturen). Direktverkäufe bleiben außen vor — der
// Direktverkauf hat eine eigene Kasse.
type AbrechnungServicekraft struct {
	UserID              int
	UserName            string // eingefrorener Username
	Name                string // live aus users aufgelöster Klarname (nur Admin-Anzeige)
	KassiertCents       int
	AnzahlZahlungen     int
	RuecknahmenCents    int
	AnzahlStornierungen int
	AbzugebenCents      int
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

// ServicekraftRef identifiziert eine Servicekraft: stabile Benutzer-ID,
// eingefrorener Username aus dem Event-Umschlag, live aus users aufgelöster
// Klarname (nur Admin-Anzeige, leer wenn der Benutzer fehlt).
type ServicekraftRef struct {
	UserID   int
	UserName string
	Name     string
}

// QuelleTisch und QuelleDirektverkauf sind die einzigen erlaubten Werte für
// StornierungDetail.Quelle und müssen mit dem CASE-Ausdruck in GetStornierungen
// (backend/sqlc/queries/reporting.sql) übereinstimmen.
const (
	QuelleTisch         = "tisch"
	QuelleDirektverkauf = "direktverkauf"
)

type StornierungDetail struct {
	Zeitpunkt    time.Time
	Quelle       string // QuelleTisch oder QuelleDirektverkauf
	BarRueckgabe bool   // true bei kassenwirksamer Warenrücknahme, false bei geldneutraler Korrektur
	TischID      int
	TischName    string
	// Akteur ist die Servicekraft, die den Storno ausgelöst hat (Event-Umschlag).
	Akteur ServicekraftRef
	// Betroffene sind die Servicekräfte, deren Vorgang der Storno rückgängig macht:
	// Kassierer der zurückgenommenen Zahlung, Verkäufer des stornierten
	// Direktverkaufs, Besteller der korrigierten Positionen. Nie leer — ohne
	// auflösbaren Verweis steht hier der Akteur.
	Betroffene  []ServicekraftRef
	BetragCents int
	Kommentar   string
	Positionen  []StornierungPosition
}

// ProduktStatistikZeile ist eine flache Ausgabezeile der GetProduktStatistik-Query
// (je Variante ausgegebene Menge und Umsatz einer Kassensitzung). Eingabe von
// gruppiereProduktStatistik; erscheint nie direkt in einer Response.
type ProduktStatistikZeile struct {
	Kategorie        string
	ProduktName      string
	VarianteID       int
	VarianteName     string
	AusgegebeneMenge int
	UmsatzCents      int
}

// VarianteStatistik trägt Menge und Umsatz als bewusst getrennte Grundlagen
// (nicht ineinander umrechenbar).
type VarianteStatistik struct {
	VarianteID       int
	VarianteName     string
	AusgegebeneMenge int
	UmsatzCents      int
}

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
	AbrechnungProServicekraft []AbrechnungServicekraft
}

// Metadaten sind die Sitzungs-Kopfdaten des Tagesberichts, rein aus den
// Journal-Events projiziert. Alle Felder sind optional, solange die zugehörigen
// Events fehlen (z. B. offene Sitzung).
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
	// ProduktStatistik ist fertig gruppiert und sortiert, von der Anwendungsschicht
	// aus den ProduktStatistikZeile-Werten gebaut.
	ProduktStatistik []ProduktStatistik
}

// AbgeschlosseneSitzung ist ein Eintrag der Kassenberichte-Sitzungsliste, aus dem
// tagesabschluss-erstellt:v1-Event. AbgeschlossenAm ist optional, falls das Event
// ausnahmsweise fehlt.
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
	// Servicekraefte führt die Abrechnung (Breakdowns.AbrechnungProServicekraft)
	// mit der offenen eigenen Arbeit aus den Tisch-Sessions zusammen (per user_id).
	Servicekraefte   []ServicekraftLive
	Stornierungen    []StornierungDetail
	ProduktStatistik []ProduktStatistik
}

// ServicekraftLive ist die Live-Sicht auf eine Servicekraft im Admin-Dashboard:
// Abrechnung plus offene eigene Arbeit (per user_id). Personen mit offener Arbeit
// ohne eigene Abrechnungszeile erscheinen mit Null-Beträgen.
type ServicekraftLive struct {
	UserID              int
	UserName            string // eingefrorener Username
	Name                string // live aus users aufgelöster Klarname (leer bei reiner offener Arbeit)
	KassiertCents       int
	RuecknahmenCents    int
	AnzahlStornierungen int
	AbzugebenCents      int
	// OffenCents ist die Summe über OffeneTische.OffenCents.
	OffenCents int
	// OffeneTische sind die Tische mit offener eigener Arbeit, aufsteigend nach Tisch-ID.
	OffeneTische []OffeneArbeitTisch
	Erledigt     bool
}

type OffeneArbeitTisch struct {
	TischID     int
	TischName   string
	AnzahlOffen int
	// OffenCents ist der offene (unbezahlte) Betrag der eigenen Positionen hier.
	OffenCents int
}

// EigeneUebersicht sind die KPIs einer Servicekraft auf ihrem Service-Dashboard.
// Rücknahmen folgen der Storno-Zuordnung: gezählt wird, was von einer Zahlung
// dieser Servicekraft zurückgenommen wurde, egal wer storniert hat. Geldneutrale
// Korrekturen bleiben außen vor.
type EigeneUebersicht struct {
	AnzahlBestellungen int
	BestellungenCents  int
	AnzahlZahlungen    int
	ZahlungenCents     int
	AnzahlRuecknahmen  int
	RuecknahmenCents   int
	// AbzugebenCents ist ZahlungenCents - RuecknahmenCents und nie negativ.
	AbzugebenCents int
}
