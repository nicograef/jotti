package druckstation

import (
	"fmt"
	"net/netip"
	"time"
)

// Kategorie unterscheidet die fünf Druckstationen: die drei Produktkategorien
// (essen, getränk, sonstiges) sowie die Sonderstationen Kassenbeleg und Abholbon.
type Kategorie string

const (
	KategorieEssen       Kategorie = "essen"
	KategorieGetraenk    Kategorie = "getraenk"
	KategorieSonstiges   Kategorie = "sonstiges"
	KategorieKassenbeleg Kategorie = "kassenbeleg"
	KategorieAbholbon    Kategorie = "abholbon"
)

// HatBonmodus meldet, ob die Station überhaupt einen Bonmodus trägt: die drei
// Produktkategorien (essen, getränk, sonstiges) und der Abholbon tragen einen,
// nur der Kassenbeleg (ein einzelner Zahlungsbeleg) nicht. Welche Modi die
// Station im Einzelnen zulässt, sagt ErlaubtBonmodus.
func (k Kategorie) HatBonmodus() bool {
	switch k {
	case KategorieEssen, KategorieGetraenk, KategorieSonstiges, KategorieAbholbon:
		return true
	default:
		return false
	}
}

// ErlaubtBonmodus meldet, ob der Bonmodus zu dieser Station passt. Die drei
// Produktkategorien drucken pro Position oder pro Bestellung; der Abholbon
// zusätzlich pro Stück (je Einheit einer Position ein eigener Bon), weil Gäste
// im Direktverkauf mehrere Einheiten auf einmal kaufen und einzeln einlösen.
// Der Kassenbeleg lässt nur den leeren Bonmodus zu. Die Methode ist die eine
// Quelle dieser Regel für Validate und die vorgelagerten Schemas.
func (k Kategorie) ErlaubtBonmodus(bonmodus Bonmodus) bool {
	if !k.HatBonmodus() {
		return bonmodus == ""
	}

	switch bonmodus {
	case BonmodusProPosition, BonmodusProBestellung:
		return true
	case BonmodusProStueck:
		return k == KategorieAbholbon
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

// Bonmodus bestimmt, wie die Bons einer Station gedruckt werden: ein Bon je
// Position, ein Sammelbon je Bestellung oder — nur am Abholbon — ein Bon je
// Einheit. Für die Station Kassenbeleg ist der Bonmodus leer.
type Bonmodus string

const (
	BonmodusProPosition   Bonmodus = "pro_position"
	BonmodusProBestellung Bonmodus = "pro_bestellung"
	BonmodusProStueck     Bonmodus = "pro_stueck"
)

type Druckstation struct {
	Kategorie Kategorie
	DruckerIP string   // IPv4-Adresse, leer = kein Drucker konfiguriert
	Bonmodus  Bonmodus // gesetzt außer beim Kassenbeleg
}

func (d Druckstation) Validate() error {
	if !d.Kategorie.isValid() {
		return fmt.Errorf("invalid kategorie")
	}

	if !d.Kategorie.ErlaubtBonmodus(d.Bonmodus) {
		return fmt.Errorf("invalid bonmodus %q for %s", d.Bonmodus, d.Kategorie)
	}

	if d.DruckerIP != "" {
		addr, err := netip.ParseAddr(d.DruckerIP)
		if err != nil || !addr.Is4() {
			return fmt.Errorf("invalid drucker_ip")
		}
	}

	return nil
}

// NewDruckstation erzeugt eine validierte Druckstation. Der Bonmodus ist für
// alle Stationen außer dem Kassenbeleg verpflichtend; welche Werte je Station
// zulässig sind, entscheidet Kategorie.ErlaubtBonmodus.
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

// FehlgeschlagenerDruckauftrag ist ein nach der maximalen Zahl von Zustellversuchen
// aufgegebener Druckauftrag, wie ihn die Druckstationen-Seite zur Verwaltung anzeigt.
type FehlgeschlagenerDruckauftrag struct {
	ID            int
	BonArt        string
	ZielIP        string
	Referenz      string
	Versuche      int
	LetzterFehler string
	ErstelltAm    time.Time
}
