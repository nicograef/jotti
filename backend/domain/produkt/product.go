package produkt

import (
	"fmt"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/domain/steuer"
)

// Kategorie represents the category of a product.
type Kategorie string

const (
	// EssenKategorie indicates the product belongs to the food category.
	EssenKategorie Kategorie = "essen"
	// GetraenkKategorie indicates the product belongs to the beverage category.
	GetraenkKategorie Kategorie = "getraenk"
	// SonstigesKategorie indicates the product belongs to the other category.
	SonstigesKategorie Kategorie = "sonstiges"
)

type Produkt struct {
	ID         int
	Name       string
	Kategorie  Kategorie
	Steuersatz steuer.Steuersatz
	Status     Status
	Varianten  []Variante
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// IDSchema defines the schema for a product ID.
var IDSchema = z.Int().GTE(1, z.Message("Ungültige Produkt-ID"))

// NameSchema defines the schema for a product's name.
var NameSchema = z.String().Trim().Min(3, z.Message("Name zu kurz")).Max(100, z.Message("Name zu lang"))

// KategorieSchema defines the schema for a product category.
var KategorieSchema = z.StringLike[Kategorie]().OneOf(
	[]Kategorie{EssenKategorie, GetraenkKategorie, SonstigesKategorie},
	z.Message("Ungültige Kategorie"),
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

// NewProdukt creates a new Produkt instance after validating the input parameters.
// The new Produkt does not have an ID assigned; it is expected to be set by the persistence layer.
func NewProdukt(name string, kategorie Kategorie, steuersatz steuer.Steuersatz) (Produkt, error) {
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
		Status:     ActiveStatus,
		Varianten:  []Variante{},
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	return produkt, nil
}

func (p *Produkt) UpdateDetails(name string, kategorie Kategorie, steuersatz steuer.Steuersatz) error {
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
	p.UpdatedAt = time.Now().UTC()

	return nil
}

func (p *Produkt) Delete() {
	p.Status = DeletedStatus
	p.UpdatedAt = time.Now().UTC()
}
