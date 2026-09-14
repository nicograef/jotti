package druckstation

import (
	"fmt"
	"net/netip"
	"slices"
	"time"
)

type Kategorie string

const (
	KategorieEssen       Kategorie = "essen"
	KategorieGetraenk    Kategorie = "getraenk"
	KategorieSonstiges   Kategorie = "sonstiges"
	KategorieKassenbeleg Kategorie = "kassenbeleg"
	KategorieAbholbon    Kategorie = "abholbon"
)

// AlleKategorien ist die einzige Quelle der Wertemenge; die OneOf-Validierung in
// api/druck/station/http/handler.go liest von hier.
func AlleKategorien() []string {
	return []string{
		string(KategorieEssen),
		string(KategorieGetraenk),
		string(KategorieSonstiges),
		string(KategorieKassenbeleg),
		string(KategorieAbholbon),
	}
}

func (k Kategorie) HatBonmodus() bool {
	switch k {
	case KategorieEssen, KategorieGetraenk, KategorieSonstiges, KategorieAbholbon:
		return true
	default:
		return false
	}
}

// ErlaubtBonmodus ist die eine Quelle dieser Regel: der Abholbon darf zusätzlich
// pro Stück drucken (je Einheit ein Bon), weil Gäste im Direktverkauf mehrere
// Einheiten auf einmal kaufen und einzeln einlösen.
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
	return slices.Contains(AlleKategorien(), string(k))
}

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
