package reporting

import "time"

type Zeitraum struct {
	Von time.Time
	Bis time.Time
}

type UmsatzServicekraft struct {
	UserID          int
	UserName        string
	ZahlungenCents  int
	AnzahlZahlungen int
}

type StornierungPosition struct {
	ProduktName  string
	VarianteName string
	Menge        int
	Einzelpreis  int
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

type DashboardData struct {
	GesamtUmsatzCents        int
	AnzahlOffeneTische       int
	AnzahlBestellungen       int
	AnzahlStornierungen      int
	GesamtBestellungenCents  int
	GesamtStornierungenCents int
}

type TagesabrechnungData struct {
	Zeitraum                 Zeitraum
	GesamtUmsatzCents        int
	GesamtBestellungenCents  int
	GesamtStornierungenCents int
	OffeneSaldiCents         int
	AnzahlBestellungen       int
	AnzahlStornierungen      int
	UmsatzProServicekraft    []UmsatzServicekraft
	Stornierungen            []StornierungDetail
}
