package product

import (
	"net/http"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api"
)

type service interface {
	CreateProduct(name, description string, netPrice float64, category Category) (*Product, error)
	UpdateProduct(id int, name, description string, netPrice float64, category Category) (*Product, error)
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
	"Description": DescriptionSchema.Required(),
	"NetPrice":    NetPriceSchema.Required(),
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
	"Description": DescriptionSchema.Required(),
	"NetPrice":    NetPriceSchema.Required(),
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
