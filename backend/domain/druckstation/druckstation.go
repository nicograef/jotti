package druckstation

import (
	"fmt"
	"net/netip"
)

// Kategorie unterscheidet die fünf Druckstationen: die drei Produktkategorien
// (essen, getraenk, sonstiges) sowie die Sonderstationen Kassenbeleg und Abholbon.
type Kategorie string

const (
	KategorieEssen       Kategorie = "essen"
	KategorieGetraenk    Kategorie = "getraenk"
	KategorieSonstiges   Kategorie = "sonstiges"
	KategorieKassenbeleg Kategorie = "kassenbeleg"
	KategorieAbholbon    Kategorie = "abholbon"
)

// HatBonmodus meldet, ob die Station einen Bonmodus trägt: die drei
// Produktkategorien (essen, getraenk, sonstiges) sowie der Abholbon werden
// wahlweise pro Position oder pro Bestellung gedruckt. Nur der Kassenbeleg
// (ein einzelner Zahlungsbeleg) hat keinen Bonmodus.
func (k Kategorie) HatBonmodus() bool {
	switch k {
	case KategorieEssen, KategorieGetraenk, KategorieSonstiges, KategorieAbholbon:
		return true
	default:
		return false
	}
}

// Anzeigename liefert die deutschsprachige Bezeichnung der Station (etwa für
// den Testbon-Kopf). Für unbekannte Kategorien fällt sie auf den Rohwert zurück.
func (k Kategorie) Anzeigename() string {
	switch k {
	case KategorieEssen:
		return "Essen"
	case KategorieGetraenk:
		return "Getränk"
	case KategorieSonstiges:
		return "Sonstiges"
	case KategorieKassenbeleg:
		return "Kassenbeleg"
	case KategorieAbholbon:
		return "Abholbon"
	default:
		return string(k)
	}
}

func (k Kategorie) isValid() bool {
	switch k {
	case KategorieEssen, KategorieGetraenk, KategorieSonstiges, KategorieKassenbeleg, KategorieAbholbon:
		return true
	default:
		return false
	}
}

// Bonmodus bestimmt, wie Arbeitsbons einer Produktstation gedruckt werden.
// Für die Sonderstationen Kassenbeleg und Abholbon ist der Bonmodus leer.
type Bonmodus string

const (
	BonmodusProPosition   Bonmodus = "pro_position"
	BonmodusProBestellung Bonmodus = "pro_bestellung"
)

type Druckstation struct {
	Kategorie Kategorie
	DruckerIP string   // IPv4-Adresse, leer = kein Drucker konfiguriert
	Bonmodus  Bonmodus // nur für Produktkategorien gesetzt, sonst leer
}

func (d Druckstation) Validate() error {
	if !d.Kategorie.isValid() {
		return fmt.Errorf("invalid kategorie")
	}

	if d.Kategorie.HatBonmodus() {
		if d.Bonmodus != BonmodusProPosition && d.Bonmodus != BonmodusProBestellung {
			return fmt.Errorf("invalid bonmodus for %s", d.Kategorie)
		}
	} else if d.Bonmodus != "" {
		return fmt.Errorf("bonmodus not allowed for %s", d.Kategorie)
	}

	if d.DruckerIP != "" {
		addr, err := netip.ParseAddr(d.DruckerIP)
		if err != nil || !addr.Is4() {
			return fmt.Errorf("invalid drucker_ip")
		}
	}

	return nil
}

// NewDruckstation erzeugt eine validierte Druckstation. Der Bonmodus ist nur für
// Produktkategorien zulässig und dort verpflichtend; kassenbeleg/abholbon tragen
// keinen Bonmodus.
func NewDruckstation(kategorie Kategorie, druckerIP string, bonmodus Bonmodus) (Druckstation, error) {
	d := Druckstation{
		Kategorie: kategorie,
		DruckerIP: druckerIP,
		Bonmodus:  bonmodus,
	}

	if err := d.Validate(); err != nil {
		return Druckstation{}, err
	}

	return d, nil
}
