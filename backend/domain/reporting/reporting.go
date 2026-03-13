package reporting

import "time"

type Zeitraum struct {
	Von time.Time `json:"von"`
	Bis time.Time `json:"bis"`
}

type UmsatzServicekraft struct {
	UserID          int    `json:"userId"`
	UserName        string `json:"userName"`
	ZahlungenCents  int    `json:"zahlungenCents"`
	AnzahlZahlungen int    `json:"anzahlZahlungen"`
}

type StornierungPosition struct {
	ProduktName  string `json:"produktName"`
	VarianteName string `json:"varianteName"`
	Menge        int    `json:"menge"`
	Einzelpreis  int    `json:"einzelpreis"`
}

type StornierungDetail struct {
	Zeitpunkt   time.Time             `json:"zeitpunkt"`
	TischID     int                   `json:"tischId"`
	TischName   string                `json:"tischName"`
	UserID      int                   `json:"userId"`
	UserName    string                `json:"userName"`
	BetragCents int                   `json:"betragCents"`
	Kommentar   string                `json:"kommentar"`
	Positionen  []StornierungPosition `json:"positionen"`
}

type DashboardData struct {
	GesamtUmsatzCents        int `json:"gesamtUmsatzCents"`
	AnzahlOffeneTische       int `json:"anzahlOffeneTische"`
	AnzahlBestellungen       int `json:"anzahlBestellungen"`
	AnzahlStornierungen      int `json:"anzahlStornierungen"`
	GesamtBestellungenCents  int `json:"gesamtBestellungenCents"`
	GesamtStornierungenCents int `json:"gesamtStornierungenCents"`
}

type TagesabrechnungData struct {
	Zeitraum                 Zeitraum             `json:"zeitraum"`
	GesamtUmsatzCents        int                  `json:"gesamtUmsatzCents"`
	GesamtBestellungenCents  int                  `json:"gesamtBestellungenCents"`
	GesamtStornierungenCents int                  `json:"gesamtStornierungenCents"`
	OffeneSaldiCents         int                  `json:"offeneSaldiCents"`
	AnzahlBestellungen       int                  `json:"anzahlBestellungen"`
	AnzahlStornierungen      int                  `json:"anzahlStornierungen"`
	UmsatzProServicekraft    []UmsatzServicekraft `json:"umsatzProServicekraft"`
	Stornierungen            []StornierungDetail  `json:"stornierungen"`
}
