package produkt

import (
	"fmt"
	"time"

	z "github.com/Oudwins/zog"
)

// Status represents the status of a product or variant.
type Status string

const (
	// ActiveStatus indicates the product or variant is active and usable for service.
	ActiveStatus Status = "active"
	// InactiveStatus indicates the product or variant is inactive and not currently in use.
	InactiveStatus Status = "inactive"
	// DeletedStatus indicates the product or variant has been soft-deleted.
	DeletedStatus Status = "deleted"
)

type Variante struct {
	ID         int
	Name       string
	PreisCents int
	Status     Status
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// PreisCentsSchema defines the schema for a product variant's gross price in cents.
var PreisCentsSchema = z.Int().GTE(0, z.Message("Preis darf nicht negativ sein")).LTE(99999, z.Message("Preis zu hoch"))

// StatusSchema defines the schema for a product or variant status.
var StatusSchema = z.StringLike[Status]().OneOf(
	[]Status{ActiveStatus, InactiveStatus, DeletedStatus},
	z.Message("Ungültiger Status"),
)

var VarianteSchema = z.Struct(z.Shape{
	"ID":         IDSchema.Required(),
	"Name":       NameSchema.Required(),
	"PreisCents": PreisCentsSchema.Required(),
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

// NewVariante creates a new Variante instance after validating the input parameters.
// The new Variante does not have an ID assigned; it is expected to be set by the persistence layer.
func NewVariante(name string, preisCents int) (Variante, error) {
	if issue := NameSchema.Validate(&name); issue != nil {
		return Variante{}, fmt.Errorf("invalid name")
	}

	if issue := PreisCentsSchema.Validate(&preisCents); issue != nil {
		return Variante{}, fmt.Errorf("invalid price")
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

	if issue := PreisCentsSchema.Validate(&preisCents); issue != nil {
		return fmt.Errorf("invalid price")
	}

	v.Name = name
	v.PreisCents = preisCents
	v.UpdatedAt = time.Now().UTC()

	return nil
}
