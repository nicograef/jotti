package produkt

import (
	"fmt"
	"time"

	z "github.com/Oudwins/zog"
)

type Status string

const (
	ActiveStatus   Status = "active"
	InactiveStatus Status = "inactive"
	DeletedStatus  Status = "deleted"
)

type Variante struct {
	ID         int
	Name       string
	PreisCents int
	Status     Status
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// VarianteMitProdukt verbindet eine Variante mit der ID ihres Produkts: die
// Anreicherung prüft damit die vom Client gesendete Paarung gegen die Datenbank.
// Die Zuordnung ist reine Persistenz und darum kein Feld von Variante.
type VarianteMitProdukt struct {
	Variante  Variante
	ProduktID int
}

// PreisCentsSchema validates a variant's gross price in cents. It is required by
// definition, so call sites use it directly and must not call .Required() again —
// zog's .Required() mutates the receiver in place, and re-mutating a shared
// exported schema is a footgun.
var PreisCentsSchema = z.Int().
	GTE(1, z.Message("Preis muss mindestens 1 Cent betragen")).
	LTE(99999, z.Message("Preis zu hoch")).
	Required(z.Message("Preis muss mindestens 1 Cent betragen"))

var StatusSchema = z.StringLike[Status]().OneOf(
	[]Status{ActiveStatus, InactiveStatus, DeletedStatus},
	z.Message("Ungültiger Status"),
)

var VarianteSchema = z.Struct(z.Shape{
	"ID":         IDSchema.Required(),
	"Name":       NameSchema.Required(),
	"PreisCents": PreisCentsSchema,
	"Status":     StatusSchema.Required(),
	"CreatedAt":  z.Time().Required(),
	"UpdatedAt":  z.Time().Required(),
})

func (v Variante) Validate() error {
	if errs := VarianteSchema.Validate(&v); errs != nil {
		issues := z.Issues.FlattenAndCollect(errs)
		return fmt.Errorf("invalid product variant: %v", issues)
	}
	return nil
}

// NewVariante assigns no ID; the persistence layer sets it.
func NewVariante(name string, preisCents int) (Variante, error) {
	if issue := NameSchema.Validate(&name); issue != nil {
		return Variante{}, fmt.Errorf("invalid name")
	}

	if issues := PreisCentsSchema.Validate(&preisCents); issues != nil {
		return Variante{}, fmt.Errorf("invalid price: %s", issues[0].Message)
	}

	variante := Variante{
		Name:       name,
		PreisCents: preisCents,
		Status:     InactiveStatus,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	return variante, nil
}

func (v *Variante) Activate() {
	v.Status = ActiveStatus
	v.UpdatedAt = time.Now().UTC()
}

func (v *Variante) Deactivate() {
	v.Status = InactiveStatus
	v.UpdatedAt = time.Now().UTC()
}

func (v *Variante) Delete() {
	v.Status = DeletedStatus
	v.UpdatedAt = time.Now().UTC()
}

func (v *Variante) UpdateDetails(name string, preisCents int) error {
	if issue := NameSchema.Validate(&name); issue != nil {
		return fmt.Errorf("invalid name")
	}

	if issues := PreisCentsSchema.Validate(&preisCents); issues != nil {
		return fmt.Errorf("invalid price: %s", issues[0].Message)
	}

	v.Name = name
	v.PreisCents = preisCents
	v.UpdatedAt = time.Now().UTC()

	return nil
}
