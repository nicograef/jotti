package reporting

import (
	"time"

	"github.com/nicograef/jotti/backend/domain/steuer"
)

type UmsatzServicekraft struct {
	UserID            int
	UserName          string // eingefrorener Username
	Name              string // live aus users aufgeloester Klarname (nur Admin-Anzeige)
	ZahlungenCents    int
	AuszahlungenCents int
	AnzahlZahlungen   int
}

type UmsatzTisch struct {
	TischID           int
	TischName         string
	ZahlungenCents    int
	AuszahlungenCents int
	AnzahlZahlungen   int
}

type UmsatzSteuersatz struct {
	Satz        steuer.Steuersatz
	BruttoCents int
	NettoCents  int
	SteuerCents int
}

type StornierungPosition struct {
	ProduktName  string
	VarianteName string
	Menge        int
	Einzelpreis  int
}

type StornierungDetail struct {
	Zeitpunkt   time.Time
	Quelle      string // "tisch" oder "direktverkauf"
	TischID     int
	TischName   string
	UserID      int
	UserName    string // eingefrorener Username
	Name        string // live aus users aufgeloester Klarname (nur Admin-Anzeige)
	BetragCents int
	Kommentar   string
	Positionen  []StornierungPosition
}

type Summary struct {
	GesamtUmsatzCents        int
	GesamtAuszahlungenCents  int
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
	KassensitzungNr             int
	Bezeichnung                 string
	Datum                       time.Time
	OffeneTische                []OffenerTisch
	OffeneSaldiCents            int
	AusstehendAuszahlungenCents int
	Summary                     Summary
	Breakdowns                  Breakdowns
	Stornierungen               []StornierungDetail
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
