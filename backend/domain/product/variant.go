package product

import (
	"fmt"
	"time"

	z "github.com/Oudwins/zog"
)

// Status represents the status of a product variant.
type Status string

const (
	// ActiveStatus indicates the product variant is active and usable for service.
	ActiveStatus Status = "active"
	// InactiveStatus indicates the product variant is inactive and not currently in use.
	InactiveStatus Status = "inactive"
)

type Variant struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	PriceCents int       `json:"priceCents"`
	Status     Status    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
}

// PriceCentsSchema defines the schema for a product variant's net price in cents.
var PriceCentsSchema = z.Int().GTE(0, z.Message("Net price must be non-negative")).LTE(99999, z.Message("Net price too high"))

// StatusSchema defines the schema for a product variant status.
var StatusSchema = z.StringLike[Status]().OneOf(
	[]Status{ActiveStatus, InactiveStatus},
	z.Message("Invalid status"),
)

var VariantSchema = z.Struct(z.Shape{
	"ID":         IDSchema.Required(),
	"Name":       NameSchema.Required(),
	"PriceCents": PriceCentsSchema.Required(),
	"Status":     StatusSchema.Required(),
	"CreatedAt":  z.Time().Required(),
})

func (v Variant) Validate() error {
	if errsMap := VariantSchema.Validate(&v); errsMap != nil {
		issues := z.Issues.SanitizeMapAndCollect(errsMap)
		return fmt.Errorf("invalid product variant: %v", issues)
	}
	return nil
}

// NewVariant creates a new Variant instance after validating the input parameters.
// The new Variant does not have an ID assigned; it is expected to be set by the persistence layer.
func NewVariant(name string, priceCents int) (Variant, error) {
	if issue := NameSchema.Validate(&name); issue != nil {
		return Variant{}, fmt.Errorf("invalid name")
	}

	if issue := PriceCentsSchema.Validate(&priceCents); issue != nil {
		return Variant{}, fmt.Errorf("invalid net price")
	}

	variant := Variant{
		Name:       name,
		PriceCents: priceCents,
		Status:     InactiveStatus,
		CreatedAt:  time.Now().UTC(),
	}

	return variant, nil
}

func (v *Variant) Activate() {
	v.Status = ActiveStatus
}

func (v *Variant) Deactivate() {
	v.Status = InactiveStatus
}

func (v *Variant) UpdateDetails(name string, priceCents int) error {
	if issue := NameSchema.Validate(&name); issue != nil {
		return fmt.Errorf("invalid name")
	}

	if issue := PriceCentsSchema.Validate(&priceCents); issue != nil {
		return fmt.Errorf("invalid net price")
	}

	v.Name = name
	v.PriceCents = priceCents

	return nil
}
