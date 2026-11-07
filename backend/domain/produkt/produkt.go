package produkt

import (
	"fmt"
	"math"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/internal/zeichen"
)

type Kategorie string

const (
	EssenKategorie     Kategorie = "essen"
	GetraenkKategorie  Kategorie = "getraenk"
	SonstigesKategorie Kategorie = "sonstiges"
)

// Richtung ist die Verschieberichtung in der Anzeigereihenfolge. Die Reihenfolge
// selbst bleibt reine Persistenz: kein Feld am Aggregat, keine Response liefert
// sie — das Backend gibt die fertig sortierte Liste aus.
type Richtung string

const (
	RichtungHoch   Richtung = "hoch"
	RichtungRunter Richtung = "runter"
)

// Pfand marks the Varianten the Pfandrückgabe offers; Zahlung and Direktverkauf freeze it per
// position, so a later change never touches a booked Vorgang (D20).
type Produkt struct {
	ID         int
	Name       string
	Kategorie  Kategorie
	Steuersatz steuer.Steuersatz
	Pfand      bool
	Status     Status
	Varianten  []Variante
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// IDSchema is bounded at both ends. The upper bound is the largest value the
// int4 column holds: a request carrying more can only be wrong, and the schema
// answers 400 instead of letting it fail inside the driver.
var IDSchema = z.Int().
	GTE(1, z.Message("Ungültige Produkt-ID")).
	LTE(math.MaxInt32, z.Message("Ungültige Produkt-ID"))

var NameSchema = z.String().Trim().TestFunc(zeichen.Min(3), z.Message("Name zu kurz")).TestFunc(zeichen.Max(100), z.Message("Name zu lang"))

var KategorieSchema = z.StringLike[Kategorie]().OneOf(
	[]Kategorie{EssenKategorie, GetraenkKategorie, SonstigesKategorie},
	z.Message("Ungültige Kategorie"),
)

var RichtungSchema = z.StringLike[Richtung]().OneOf(
	[]Richtung{RichtungHoch, RichtungRunter},
	z.Message("Ungültige Richtung"),
)

var SteuersatzSchema = steuer.SteuersatzSchema

var ProduktSchema = z.Struct(z.Shape{
	"ID":         IDSchema.Required(),
	"Name":       NameSchema.Required(),
	"Kategorie":  KategorieSchema.Required(),
	"Steuersatz": SteuersatzSchema.Required(),
	"Status":     StatusSchema.Required(),
	"Varianten":  z.Slice(VarianteSchema).Required(),
	"CreatedAt":  z.Time().Required(),
	"UpdatedAt":  z.Time().Required(),
})

func (p Produkt) Validate() error {
	if errs := ProduktSchema.Validate(&p); errs != nil {
		issues := z.Issues.FlattenAndCollect(errs)
		return fmt.Errorf("invalid product: %v", issues)
	}
	return nil
}

// NewProdukt assigns no ID; the persistence layer sets it.
func NewProdukt(name string, kategorie Kategorie, steuersatz steuer.Steuersatz, pfand bool) (Produkt, error) {
	if issue := NameSchema.Validate(&name); issue != nil {
		return Produkt{}, fmt.Errorf("invalid name")
	}

	if issue := KategorieSchema.Validate(&kategorie); issue != nil {
		return Produkt{}, fmt.Errorf("invalid category")
	}

	if issue := SteuersatzSchema.Validate(&steuersatz); issue != nil {
		return Produkt{}, fmt.Errorf("invalid tax rate")
	}

	produkt := Produkt{
		Name:       name,
		Kategorie:  kategorie,
		Steuersatz: steuersatz,
		Pfand:      pfand,
		Status:     ActiveStatus,
		Varianten:  []Variante{},
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	return produkt, nil
}

func (p *Produkt) UpdateDetails(name string, kategorie Kategorie, steuersatz steuer.Steuersatz, pfand bool) error {
	if issue := NameSchema.Validate(&name); issue != nil {
		return fmt.Errorf("invalid name")
	}

	if issue := KategorieSchema.Validate(&kategorie); issue != nil {
		return fmt.Errorf("invalid category")
	}

	if issue := SteuersatzSchema.Validate(&steuersatz); issue != nil {
		return fmt.Errorf("invalid tax rate")
	}

	p.Name = name
	p.Kategorie = kategorie
	p.Steuersatz = steuersatz
	p.Pfand = pfand
	p.UpdatedAt = time.Now().UTC()

	return nil
}

func (p *Produkt) Delete() {
	p.Status = DeletedStatus
	p.UpdatedAt = time.Now().UTC()
}
