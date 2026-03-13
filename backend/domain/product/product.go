package product

import (
	"fmt"
	"time"

	z "github.com/Oudwins/zog"
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
	ID        int
	Name      string
	Kategorie Kategorie
	Status    Status
	Varianten []Variante
	CreatedAt time.Time
	UpdatedAt time.Time
}

// IDSchema defines the schema for a product ID.
var IDSchema = z.Int().GTE(1, z.Message("Invalid product ID"))

// NameSchema defines the schema for a product's name.
var NameSchema = z.String().Trim().Min(3, z.Message("Name too short")).Max(100, z.Message("Name too long"))

// KategorieSchema defines the schema for a product category.
var KategorieSchema = z.StringLike[Kategorie]().OneOf(
	[]Kategorie{EssenKategorie, GetraenkKategorie, SonstigesKategorie},
	z.Message("Invalid category"),
)

var ProduktSchema = z.Struct(z.Shape{
	"ID":        IDSchema.Required(),
	"Name":      NameSchema.Required(),
	"Kategorie": KategorieSchema.Required(),
	"Status":    StatusSchema.Required(),
	"Varianten": z.Slice(VarianteSchema).Required(),
	"CreatedAt": z.Time().Required(),
	"UpdatedAt": z.Time().Required(),
})

func (p Produkt) Validate() error {
	if errs := ProduktSchema.Validate(&p); errs != nil {
		issues := z.Issues.FlattenAndCollect(errs)
		return fmt.Errorf("invalid product: %v", issues)
	}
	return nil
}

// NewProduct creates a new Product instance after validating the input parameters.
// The new Product does not have an ID assigned; it is expected to be set by the persistence layer.
func NewProdukt(name string, kategorie Kategorie) (Produkt, error) {
	if issue := NameSchema.Validate(&name); issue != nil {
		return Produkt{}, fmt.Errorf("invalid name")
	}

	if issue := KategorieSchema.Validate(&kategorie); issue != nil {
		return Produkt{}, fmt.Errorf("invalid category")
	}

	produkt := Produkt{
		Name:      name,
		Kategorie: kategorie,
		Status:    ActiveStatus,
		Varianten: []Variante{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	return produkt, nil
}

func (p *Produkt) UpdateDetails(name string, kategorie Kategorie) error {
	if issue := NameSchema.Validate(&name); issue != nil {
		return fmt.Errorf("invalid name")
	}

	if issue := KategorieSchema.Validate(&kategorie); issue != nil {
		return fmt.Errorf("invalid category")
	}

	p.Name = name
	p.Kategorie = kategorie
	p.UpdatedAt = time.Now().UTC()

	return nil
}

func (p *Produkt) Activate() {
	p.Status = ActiveStatus
	p.UpdatedAt = time.Now().UTC()
}

func (p *Produkt) Deactivate() {
	p.Status = InactiveStatus
	p.UpdatedAt = time.Now().UTC()
}

func (p *Produkt) Delete() {
	p.Status = DeletedStatus
	p.UpdatedAt = time.Now().UTC()
}
