package product

import (
	"errors"

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
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	NetPriceCents int      `json:"netPriceCents"`
	Category      Category `json:"category"`
}

// IDSchema defines the schema for a user ID.
var IDSchema = z.Int().GTE(1, z.Message("Invalid product ID"))

// NameSchema defines the schema for a product's name.
var NameSchema = z.String().Trim().Min(3, z.Message("Name too short")).Max(30, z.Message("Name too long"))

// DescriptionSchema defines the schema for a product's description.
var DescriptionSchema = z.String().Trim().Min(0).Max(250, z.Message("Description too long"))

// NetPriceCentsSchema defines the schema for a product's net price in cents.
var NetPriceCentsSchema = z.Int().GTE(0, z.Message("Net price must be non-negative")).LTE(99999, z.Message("Net price too high"))

// CategorySchema defines the schema for a product category.
var CategorySchema = z.StringLike[Category]().OneOf(
	[]Category{FoodCategory, BeverageCategory, OtherCategory},
	z.Message("Invalid category"),
)

// ErrProductNotFound is returned when a product is not found.
var ErrProductNotFound = errors.New("product not found")

// ErrDatabase is returned when there is a database error.
var ErrDatabase = errors.New("database error")
