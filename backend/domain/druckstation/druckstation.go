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
	case KategorieKassenbeleg:
		fallthrough
	default:
		return false
	}
}

// ErlaubtBonmodus ist die eine Quelle dieser Regel. Nur der Abholbon darf pro Stück
// drucken (Gäste lösen Einheiten einzeln ein) und mit kein_bon den Direktverkauf stummschalten.
func (k Kategorie) ErlaubtBonmodus(bonmodus Bonmodus) bool {
	if !k.HatBonmodus() {
		return bonmodus == ""
	}

	switch bonmodus {
	case BonmodusProPosition, BonmodusProBestellung:
		return true
	case BonmodusProStueck, BonmodusKeinBon:
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
	// BonmodusKeinBon: ein Direktverkauf druckt keinen Bon, auch nicht an den Produktstationen.
	BonmodusKeinBon Bonmodus = "kein_bon"
)

// AlleBonmodi ist die Backend-Quelle der Wertemenge; die OneOf-Validierung in
// api/druck/station/http/handler.go liest von hier.
func AlleBonmodi() []string {
	return []string{
		string(BonmodusProPosition),
		string(BonmodusProBestellung),
		string(BonmodusProStueck),
		string(BonmodusKeinBon),
	}
}

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
