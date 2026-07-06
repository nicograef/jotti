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

type UmsatzTisch struct {
	TischID         int
	TischName       string
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

type StornierungDetail struct {
	Zeitpunkt    time.Time
	Quelle       string // "tisch" oder "direktverkauf"
	BarRueckgabe bool   // true bei kassenwirksamer Warenrücknahme, false bei geldneutraler Korrektur
	TischID      int
	TischName    string
	UserID       int
	UserName     string // eingefrorener Username
	Name         string // live aus users aufgeloester Klarname (nur Admin-Anzeige)
	BetragCents  int
	Kommentar    string
	Positionen   []StornierungPosition
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
	UmsatzProServicekraft []UmsatzServicekraft
	UmsatzProTisch        []UmsatzTisch
}

type ReportingData struct {
	KassensitzungNr     int
	Summary             Summary
	Breakdowns          Breakdowns
	UmsatzProSteuersatz []UmsatzSteuersatz
	Stornierungen       []StornierungDetail
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
	// OffeneTische listet die Tische mit offener eigener Arbeit (aufsteigend nach
	// Tisch-ID); leer wenn die Servicekraft fertig ist.
	OffeneTische []OffeneArbeitTisch
	// Erledigt ist true, wenn keine offene eigene Arbeit mehr besteht.
	Erledigt bool
}

// OffeneArbeitTisch ist die offene eigene Arbeit einer Servicekraft an einem
// Tisch, angereichert um den Tisch-Namen für die Anzeige.
type OffeneArbeitTisch struct {
	TischID          int
	TischName        string
	AnzahlAusstehend int
	AnzahlUnbezahlt  int
	AnzahlOffen      int
}

type EigeneUebersicht struct {
	AnzahlBestellungen int
	BestellungenCents  int
	AnzahlZahlungen    int
	ZahlungenCents     int
	// OffeneTische listet die Tische der offenen Kassensitzung, an denen die
	// Servicekraft noch offene eigene Arbeit hat (aufsteigend nach Tisch-ID).
	OffeneTische []OffeneArbeitTisch
	// AlleErledigt ist true, wenn die Servicekraft an keinem Tisch noch offene
	// eigene Arbeit hat.
	AlleErledigt bool
}
