package reporting

import "time"

type Zeitraum struct {
	Von time.Time
	Bis time.Time
}

type UmsatzServicekraft struct {
	UserID            int
	UserName          string
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

type StornierungPosition struct {
	ProduktName  string `json:"produktName"`
	VarianteName string `json:"varianteName"`
	Menge        int    `json:"menge"`
	Einzelpreis  int    `json:"einzelpreis"`
}

type StornierungDetail struct {
	Zeitpunkt   time.Time
	TischID     int
	TischName   string
	UserID      int
	UserName    string
	BetragCents int
	Kommentar   string
	Positionen  []StornierungPosition
}

type Summary struct {
	GesamtUmsatzCents           int
	GesamtAuszahlungenCents     int
	GesamtBestellungenCents     int
	GesamtStornierungenCents    int
	OffeneSaldiCents            int
	AusstehendAuszahlungenCents int
	AnzahlOffeneTische          int
	AnzahlBestellungen          int
	AnzahlStornierungen         int
}

type Breakdowns struct {
	UmsatzProServicekraft []UmsatzServicekraft
	UmsatzProTisch        []UmsatzTisch
}

type ReportingData struct {
	Zeitraum      Zeitraum
	Summary       Summary
	Breakdowns    Breakdowns
	Stornierungen []StornierungDetail
}

type EigeneUebersicht struct {
	AnzahlBestellungen int
	BestellungenCents  int
	AnzahlZahlungen    int
	ZahlungenCents     int
}
