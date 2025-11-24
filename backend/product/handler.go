package product

import (
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api"
)

type service interface {
	CreateProduct(name, description string, netPrice float64, category Category) (*Product, error)
	UpdateProduct(id int, name, description string, netPrice float64, category Category) (*Product, error)
	GetAllProducts() ([]*Product, error)
	GetActiveProducts() ([]*ProductPublic, error)
	ActivateProduct(id int) error
	DeactivateProduct(id int) error
}

type Handler struct {
	Service service
}

type createProductRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	NetPrice    float64  `json:"netPrice"`
	Category    Category `json:"category"`
}

var createProductRequestSchema = z.Struct(z.Shape{
	"Name":        NameSchema.Required(),
	"Description": DescriptionSchema.Default(""),
	"NetPrice":    NetPriceSchema.Default(0),
	"Category":    CategorySchema.Required(),
})

type createProductResponse struct {
	Product Product `json:"product"`
}

// CreateProductHandler handles requests to create a new product.
func (h *Handler) CreateProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !api.ValidateMethod(w, r, http.MethodPost) {
			return
		}

		body := createProductRequest{}
		if !api.ReadJSONRequest(w, r, &body) {
			return
		}

		if !api.ValidateBody(w, &body, createProductRequestSchema) {
			return
		}

		product, err := h.Service.CreateProduct(body.Name, body.Description, body.NetPrice, body.Category)
		if err != nil {
			api.SendInternalServerError(w)
			return
		}

		api.SendResponse(w, createProductResponse{
			Product: *product,
		})
	}
}

type updateProductRequest struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	NetPrice    float64  `json:"netPrice"`
	Category    Category `json:"category"`
}

var updateProductRequestSchema = z.Struct(z.Shape{
	"ID":          IDSchema.Required(),
	"Name":        NameSchema.Required(),
	"Description": DescriptionSchema.Default(""),
	"NetPrice":    NetPriceSchema.Default(0),
	"Category":    CategorySchema.Required(),
})

type updateProductResponse struct {
	Product Product `json:"product"`
}

// UpdateProductHandler handles requests to update an existing product.
func (h *Handler) UpdateProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !api.ValidateMethod(w, r, http.MethodPost) {
			return
		}

		body := updateProductRequest{}
		if !api.ReadJSONRequest(w, r, &body) {
			return
		}

		if !api.ValidateBody(w, &body, updateProductRequestSchema) {
			return
		}

		product, err := h.Service.UpdateProduct(body.ID, body.Name, body.Description, body.NetPrice, body.Category)
		if err != nil {
			api.SendInternalServerError(w)
			return
		}

		api.SendResponse(w, updateProductResponse{
			Product: *product,
		})
	}
}

type getAllProductsResponse struct {
	Products []*Product `json:"products"`
}

// GetAllProductsHandler handles requests to retrieve all products.
func (h *Handler) GetAllProductsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !api.ValidateMethod(w, r, http.MethodPost) {
			return
		}

		products, err := h.Service.GetAllProducts()
		if err != nil {
			api.SendInternalServerError(w)
			return
		}

		if products == nil {
			products = []*Product{}
		}

		api.SendResponse(w, getAllProductsResponse{
			Products: products,
		})
	}
}

type getActiveProductsResponse struct {
	Products []*ProductPublic `json:"products"`
}

// GetActiveProductsHandler handles requests to retrieve all active products.
func (h *Handler) GetActiveProductsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !api.ValidateMethod(w, r, http.MethodPost) {
			return
		}

		products, err := h.Service.GetActiveProducts()
		if err != nil {
			api.SendInternalServerError(w)
			return
		}

		if products == nil {
			products = []*ProductPublic{}
		}

		api.SendResponse(w, getActiveProductsResponse{
			Products: products,
		})
	}
}

type activateProductRequest struct {
	ID int `json:"id"`
}

var activateProductRequestSchema = z.Struct(z.Shape{
	"ID": IDSchema.Required(),
})

// ActivateProductHandler handles requests to activate a product.
func (h *Handler) ActivateProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !api.ValidateMethod(w, r, http.MethodPost) {
			return
		}

		body := activateProductRequest{}
		if !api.ReadJSONRequest(w, r, &body) {
			return
		}

		if !api.ValidateBody(w, &body, activateProductRequestSchema) {
			return
		}

		err := h.Service.ActivateProduct(body.ID)
		if err != nil {
			if err == ErrProductNotFound {
				api.SendNotFoundError(w, api.ErrorResponse{
					Message: "Product not found",
					Code:    "product_not_found",
				})
				return
			}
			api.SendInternalServerError(w)
			return
		}

		api.SendEmptyResponse(w)
	}
}

type deactivateTableRequest struct {
	ID int `json:"id"`
}

var deactivateTableRequestSchema = z.Struct(z.Shape{
	"ID": IDSchema.Required(),
})

// DeactivateProductHandler handles requests to deactivate a product.
func (h *Handler) DeactivateProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !api.ValidateMethod(w, r, http.MethodPost) {
			return
		}

		body := deactivateTableRequest{}
		if !api.ReadJSONRequest(w, r, &body) {
			return
		}

		if !api.ValidateBody(w, &body, deactivateTableRequestSchema) {
			return
		}

		err := h.Service.DeactivateProduct(body.ID)
		if err != nil {
			if err == ErrProductNotFound {
				api.SendNotFoundError(w, api.ErrorResponse{
					Message: "Product not found",
					Code:    "product_not_found",
				})
				return
			}
			api.SendInternalServerError(w)
			return
		}

		api.SendEmptyResponse(w)
	}
}
