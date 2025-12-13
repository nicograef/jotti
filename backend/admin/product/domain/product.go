package domain

import (
	"fmt"
	"time"

	z "github.com/Oudwins/zog"
)

// Status represents the status of a product.
type Status string

const (
	// ActiveStatus indicates the product is active and usable for service.
	ActiveStatus Status = "active"
	// InactiveStatus indicates the product is inactive and not currently in use.
	InactiveStatus Status = "inactive"
	// DeletedStatus indicates the product has been deleted and is no longer in use.
	DeletedStatus Status = "deleted"
)

// Category represents the category of a product.
type Category string

const (
	// FoodCategory indicates the product belongs to the food category.
	FoodCategory Category = "food"
	// BeverageCategory indicates the product belongs to the beverage category.
	BeverageCategory Category = "beverage"
	// OtherCategory indicates the product belongs to the other category.
	OtherCategory Category = "other"
)

type Product struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	NetPriceCents int       `json:"netPriceCents"`
	Status        Status    `json:"status"`
	Category      Category  `json:"category"`
	CreatedAt     time.Time `json:"createdAt"`
}

// IDSchema defines the schema for a user ID.
var IDSchema = z.Int().GTE(1, z.Message("Invalid product ID"))

// NameSchema defines the schema for a product's name.
var NameSchema = z.String().Trim().Min(3, z.Message("Name too short")).Max(30, z.Message("Name too long"))

// DescriptionSchema defines the schema for a product's description.
var DescriptionSchema = z.String().Trim().Min(0).Max(250, z.Message("Description too long"))

// NetPriceCentsSchema defines the schema for a product's net price in cents.
var NetPriceCentsSchema = z.Int().GTE(0, z.Message("Net price must be non-negative")).LTE(99999, z.Message("Net price too high"))

// StatusSchema defines the schema for a product status.
var StatusSchema = z.StringLike[Status]().OneOf(
	[]Status{ActiveStatus, InactiveStatus, DeletedStatus},
	z.Message("Invalid status"),
)

// CategorySchema defines the schema for a product category.
var CategorySchema = z.StringLike[Category]().OneOf(
	[]Category{FoodCategory, BeverageCategory, OtherCategory},
	z.Message("Invalid category"),
)

var ProductSchema = z.Struct(z.Shape{
	"ID":            IDSchema.Required(),
	"Name":          NameSchema.Required(),
	"Description":   DescriptionSchema.Required(),
	"NetPriceCents": NetPriceCentsSchema.Required(),
	"Status":        StatusSchema.Required(),
	"Category":      CategorySchema.Required(),
	"CreatedAt":     z.Time().Required(),
})

func (p Product) Validate() error {
	if errsMap := ProductSchema.Validate(&p); errsMap != nil {
		issues := z.Issues.SanitizeMapAndCollect(errsMap)
		return fmt.Errorf("invalid product: %v", issues)
	}
	return nil
}

// NewProduct creates a new Product instance after validating the input parameters.
// The new Product does not have an ID assigned; it is expected to be set by the persistence layer.
func NewProduct(name, description string, netPriceCents int, category Category) (Product, error) {
	if issue := NameSchema.Validate(&name); issue != nil {
		return Product{}, fmt.Errorf("invalid name")
	}

	if issue := DescriptionSchema.Validate(&description); issue != nil {
		return Product{}, fmt.Errorf("invalid description")
	}

	if issue := NetPriceCentsSchema.Validate(&netPriceCents); issue != nil {
		return Product{}, fmt.Errorf("invalid net price")
	}

	if issue := CategorySchema.Validate(&category); issue != nil {
		return Product{}, fmt.Errorf("invalid category")
	}

	return Product{
		Name:          name,
		Description:   description,
		NetPriceCents: netPriceCents,
		Status:        InactiveStatus,
		Category:      category,
		CreatedAt:     time.Now(),
	}, nil
}

func (p *Product) Activate() {
	p.Status = ActiveStatus
}

func (p *Product) Deactivate() {
	p.Status = InactiveStatus
}

func (p *Product) Update(name, description string, netPriceCents int, category Category) error {
	if issue := NameSchema.Validate(&name); issue != nil {
		return fmt.Errorf("invalid name")
	}

	if issue := DescriptionSchema.Validate(&description); issue != nil {
		return fmt.Errorf("invalid description")
	}

	if issue := NetPriceCentsSchema.Validate(&netPriceCents); issue != nil {
		return fmt.Errorf("invalid net price")
	}

	if issue := CategorySchema.Validate(&category); issue != nil {
		return fmt.Errorf("invalid category")
	}

	p.Name = name
	p.Description = description
	p.NetPriceCents = netPriceCents
	p.Category = category

	return nil
}
