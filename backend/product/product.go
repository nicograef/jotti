package product

import (
	"errors"
	"log"
	"time"

	z "github.com/Oudwins/zog"
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
	GetProduct(id int) (*Product, error)
	GetAllProducts() ([]*Product, error)
	GetActiveProducts() ([]*ProductPublic, error)
	CreateProduct(name, description string, netPrice float64, category Category) (int, error)
	UpdateProduct(id int, name, description string, netPrice float64, category Category) error
	ActivateProduct(id int) error
	DeactivateProduct(id int) error
}

// Service provides product-related operations.
type Service struct {
	Persistence persistence
}

// CreateProduct creates a new product in the database.
func (s *Service) CreateProduct(name, description string, netPrice float64, category Category) (*Product, error) {
	id, err := s.Persistence.CreateProduct(name, description, netPrice, category)
	if err != nil {
		log.Printf("ERROR creating product: %v", err)
		return nil, ErrDatabase
	}

	product, err := s.Persistence.GetProduct(id)
	if err != nil {
		log.Printf("ERROR Failed to retrieve product %d after creation: %v", id, err)
		return nil, ErrDatabase
	}

	return product, nil
}

// UpdateProduct updates an existing product in the database.
func (s *Service) UpdateProduct(id int, name, description string, netPrice float64, category Category) (*Product, error) {
	err := s.Persistence.UpdateProduct(id, name, description, netPrice, category)
	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			return nil, ErrProductNotFound
		}
		log.Printf("ERROR updating product: %v", err)
		return nil, ErrDatabase
	}

	updatedProduct, err := s.Persistence.GetProduct(id)
	if err != nil {
		log.Printf("ERROR Failed to retrieve updated product %d: %v", id, err)
		return nil, ErrDatabase
	}

	return updatedProduct, nil
}

// GetProduct retrieves a product by its ID.
func (s *Service) GetProduct(id int) (*Product, error) {
	product, err := s.Persistence.GetProduct(id)
	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			return nil, ErrProductNotFound
		}
		log.Printf("ERROR retrieving product %d: %v", id, err)
		return nil, ErrDatabase
	}
	return product, nil
}

// GetAllProducts retrieves all products.
func (s *Service) GetAllProducts() ([]*Product, error) {
	products, err := s.Persistence.GetAllProducts()
	if err != nil {
		log.Printf("ERROR retrieving all products: %v", err)
		return nil, ErrDatabase
	}
	return products, nil
}

// GetActiveProducts retrieves active products.
func (s *Service) GetActiveProducts() ([]*ProductPublic, error) {
	products, err := s.Persistence.GetActiveProducts()
	if err != nil {
		log.Printf("ERROR retrieving active products: %v", err)
		return nil, ErrDatabase
	}
	return products, nil
}

// ActivateProduct sets the status of a product to active.
func (s *Service) ActivateProduct(id int) error {
	err := s.Persistence.ActivateProduct(id)
	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			return ErrProductNotFound
		}
		log.Printf("ERROR activating product %d: %v", id, err)
		return ErrDatabase
	}
	return nil
}

// DeactivateProduct sets the status of a product to inactive.
func (s *Service) DeactivateProduct(id int) error {
	err := s.Persistence.DeactivateProduct(id)
	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			return ErrProductNotFound
		}
		log.Printf("ERROR deactivating product %d: %v", id, err)
		return ErrDatabase
	}
	return nil
}
