package druckstation

import (
	"fmt"
	"math"
	"net/netip"
	"slices"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/internal/zeichen"
)

// Zusatzdrucker is a named printer next to the Druckstationen. A Produkt assigned to it prints its
// Arbeitsbon there instead of at the Druckstation of its Kategorie.
type Zusatzdrucker struct {
	ID        int
	Name      string
	DruckerIP string // IPv4-Adresse, leer = kein Drucker konfiguriert
	Bonmodus  Bonmodus
}

// ZusatzdruckerIDSchema is bounded by the int4 column.
var ZusatzdruckerIDSchema = z.Int().
	GTE(1, z.Message("Ungültige Zusatzdrucker-ID")).
	LTE(math.MaxInt32, z.Message("Ungültige Zusatzdrucker-ID"))

var ZusatzdruckerNameSchema = z.String().Trim().TestFunc(zeichen.Min(3), z.Message("Name zu kurz")).TestFunc(zeichen.Max(100), z.Message("Name zu lang"))

// ZusatzdruckerBonmodi is the value set of a Zusatzdrucker's Bonmodus; pro_stueck and kein_bon belong to the Abholbon.
func ZusatzdruckerBonmodi() []string {
	return []string{string(BonmodusProPosition), string(BonmodusProBestellung)}
}

func (zd Zusatzdrucker) Validate() error {
	// zog skips the tests of an empty value unless the schema is Required.
	name := zd.Name
	if issue := ZusatzdruckerNameSchema.Validate(&name); issue != nil || name == "" || name != zd.Name {
		return fmt.Errorf("invalid name")
	}

	if !slices.Contains(ZusatzdruckerBonmodi(), string(zd.Bonmodus)) {
		return fmt.Errorf("invalid bonmodus %q", zd.Bonmodus)
	}

	if zd.DruckerIP != "" {
		addr, err := netip.ParseAddr(zd.DruckerIP)
		if err != nil || !addr.Is4() {
			return fmt.Errorf("invalid drucker_ip")
		}
	}

	return nil
}

// NewZusatzdrucker trims the name and assigns no ID; the persistence layer sets it.
func NewZusatzdrucker(name, druckerIP string, bonmodus Bonmodus) (Zusatzdrucker, error) {
	if issue := ZusatzdruckerNameSchema.Validate(&name); issue != nil {
		return Zusatzdrucker{}, fmt.Errorf("invalid name")
	}

	zd := Zusatzdrucker{
		Name:      name,
		DruckerIP: druckerIP,
		Bonmodus:  bonmodus,
	}

	if err := zd.Validate(); err != nil {
		return Zusatzdrucker{}, err
	}

	return zd, nil
}

// ProduktZuordnung is a non-deleted Produkt with its Zusatzdrucker; ZusatzdruckerID 0 means none.
type ProduktZuordnung struct {
	ProduktID         int
	ProduktName       string
	Kategorie         produkt.Kategorie
	ZusatzdruckerID   int
	ZusatzdruckerName string
}
