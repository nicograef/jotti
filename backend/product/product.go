package product

import (
	"context"
	"errors"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/rs/zerolog/log"
)

// Product represents a user in the system.
type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	NetPrice    float64   `json:"netPrice"`
	Status      Status    `json:"status"`
	Category    Category  `json:"category"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ProductPublic struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	NetPrice    float64  `json:"netPrice"`
	Category    Category `json:"category"`
}

// IDSchema defines the schema for a user ID.
var IDSchema = z.Int().GTE(1, z.Message("Invalid product ID"))

// NameSchema defines the schema for a product's name.
var NameSchema = z.String().Trim().Min(3, z.Message("Name too short")).Max(30, z.Message("Name too long"))

// DescriptionSchema defines the schema for a product's description.
var DescriptionSchema = z.String().Trim().Min(0).Max(250, z.Message("Description too long"))

// NetPriceSchema defines the schema for a product's net price.
var NetPriceSchema = z.Float64().GTE(0, z.Message("Net price must be non-negative")).LTE(999.99, z.Message("Net price too high"))

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

// ErrProductNotFound is returned when a product is not found.
var ErrProductNotFound = errors.New("product not found")

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

// ErrDatabase is returned when there is a database error.
var ErrDatabase = errors.New("database error")

type persistence interface {
	GetProduct(ctx context.Context, id int) (*Product, error)
	GetAllProducts(ctx context.Context) ([]*Product, error)
	GetActiveProducts(ctx context.Context) ([]*ProductPublic, error)
	CreateProduct(ctx context.Context, name, description string, netPrice float64, category Category) (int, error)
	UpdateProduct(ctx context.Context, id int, name, description string, netPrice float64, category Category) error
	ActivateProduct(ctx context.Context, id int) error
	DeactivateProduct(ctx context.Context, id int) error
}

// Service provides product-related operations.
type Service struct {
	Persistence persistence
}

// CreateProduct creates a new product in the database.
func (s *Service) CreateProduct(ctx context.Context, name, description string, netPrice float64, category Category) (*Product, error) {
	id, err := s.Persistence.CreateProduct(ctx, name, description, netPrice, category)
	if err != nil {
		log.Error().Err(err).Str("name", name).Msg("Failed to create product")
		return nil, ErrDatabase
	}

	product, err := s.Persistence.GetProduct(ctx, id)
	if err != nil {
		log.Error().Err(err).Int("product_id", id).Msg("Failed to retrieve product after creation")
		return nil, ErrDatabase
	}

	return product, nil
}

// UpdateProduct updates an existing product in the database.
func (s *Service) UpdateProduct(ctx context.Context, id int, name, description string, netPrice float64, category Category) (*Product, error) {
	err := s.Persistence.UpdateProduct(ctx, id, name, description, netPrice, category)
	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			return nil, ErrProductNotFound
		}
		log.Error().Err(err).Int("product_id", id).Msg("Failed to update product")
		return nil, ErrDatabase
	}

	updatedProduct, err := s.Persistence.GetProduct(ctx, id)
	if err != nil {
		log.Error().Err(err).Int("product_id", id).Msg("Failed to retrieve updated product")
		return nil, ErrDatabase
	}

	return updatedProduct, nil
}

// GetProduct retrieves a product by its ID.
func (s *Service) GetProduct(ctx context.Context, id int) (*Product, error) {
	product, err := s.Persistence.GetProduct(ctx, id)
	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			return nil, ErrProductNotFound
		}
		log.Error().Err(err).Int("product_id", id).Msg("Failed to retrieve product")
		return nil, ErrDatabase
	}
	return product, nil
}

// GetAllProducts retrieves all products.
func (s *Service) GetAllProducts(ctx context.Context) ([]*Product, error) {
	products, err := s.Persistence.GetAllProducts(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve all products")
		return nil, ErrDatabase
	}
	return products, nil
}

// GetActiveProducts retrieves active products.
func (s *Service) GetActiveProducts(ctx context.Context) ([]*ProductPublic, error) {
	products, err := s.Persistence.GetActiveProducts(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve active products")
		return nil, ErrDatabase
	}
	return products, nil
}

// ActivateProduct sets the status of a product to active.
func (s *Service) ActivateProduct(ctx context.Context, id int) error {
	err := s.Persistence.ActivateProduct(ctx, id)
	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			return ErrProductNotFound
		}
		log.Error().Err(err).Int("product_id", id).Msg("Failed to activate product")
		return ErrDatabase
	}
	return nil
}

// DeactivateProduct sets the status of a product to inactive.
func (s *Service) DeactivateProduct(ctx context.Context, id int) error {
	err := s.Persistence.DeactivateProduct(ctx, id)
	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			return ErrProductNotFound
		}
		log.Error().Err(err).Int("product_id", id).Msg("Failed to deactivate product")
		return ErrDatabase
	}
	return nil
}
