package product

import (
	"fmt"
	"time"

	z "github.com/Oudwins/zog"
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
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Category  Category  `json:"category"`
	Variants  []Variant `json:"variants"`
	CreatedAt time.Time `json:"createdAt"`
}

// IDSchema defines the schema for a product ID.
var IDSchema = z.Int().GTE(1, z.Message("Invalid product ID"))

// NameSchema defines the schema for a product's name.
var NameSchema = z.String().Trim().Min(3, z.Message("Name too short")).Max(100, z.Message("Name too long"))

// CategorySchema defines the schema for a product category.
var CategorySchema = z.StringLike[Category]().OneOf(
	[]Category{FoodCategory, BeverageCategory, OtherCategory},
	z.Message("Invalid category"),
)

var ProductSchema = z.Struct(z.Shape{
	"ID":        IDSchema.Required(),
	"Name":      NameSchema.Required(),
	"Category":  CategorySchema.Required(),
	"Variants":  z.Slice(VariantSchema).Required(),
	"CreatedAt": z.Time().Required(),
})

func (p Product) Validate() error {
	if errs := ProductSchema.Validate(&p); errs != nil {
		issues := z.Issues.FlattenAndCollect(errs)
		return fmt.Errorf("invalid product: %v", issues)
	}
	return nil
}

// NewProduct creates a new Product instance after validating the input parameters.
// The new Product does not have an ID assigned; it is expected to be set by the persistence layer.
func NewProduct(name string, category Category) (Product, error) {
	if issue := NameSchema.Validate(&name); issue != nil {
		return Product{}, fmt.Errorf("invalid name")
	}

	if issue := CategorySchema.Validate(&category); issue != nil {
		return Product{}, fmt.Errorf("invalid category")
	}

	product := Product{
		Name:      name,
		Category:  category,
		Variants:  []Variant{},
		CreatedAt: time.Now().UTC(),
	}

	return product, nil
}

func (p *Product) UpdateDetails(name string, category Category) error {
	if issue := NameSchema.Validate(&name); issue != nil {
		return fmt.Errorf("invalid name")
	}

	if issue := CategorySchema.Validate(&category); issue != nil {
		return fmt.Errorf("invalid category")
	}

	p.Name = name
	p.Category = category

	return nil
}
