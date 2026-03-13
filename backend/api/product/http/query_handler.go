package http

import (
	"context"
	"net/http"
	"time"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/product/application"
	"github.com/nicograef/jotti/backend/domain/product"
)

type productReader interface {
	GetAllProducts(ctx context.Context) ([]product.Produkt, error)
	GetActiveProducts(ctx context.Context) ([]product.Produkt, error)
}

type QueryHandler struct {
	ProductRepo productReader
}

type varianteDTO struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	PreisCents int       `json:"preisCents"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type produktDTO struct {
	ID        int           `json:"id"`
	Name      string        `json:"name"`
	Kategorie string        `json:"kategorie"`
	Status    string        `json:"status"`
	Varianten []varianteDTO `json:"varianten"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
}

type getAllProductsResponse struct {
	Produkte []produktDTO `json:"produkte"`
}

func toVarianteDTO(v product.Variante) varianteDTO {
	return varianteDTO{
		ID:         v.ID,
		Name:       v.Name,
		PreisCents: v.PreisCents,
		Status:     string(v.Status),
		CreatedAt:  v.CreatedAt,
		UpdatedAt:  v.UpdatedAt,
	}
}

func toProduktDTO(p product.Produkt) produktDTO {
	varianten := make([]varianteDTO, 0, len(p.Varianten))
	for _, variante := range p.Varianten {
		varianten = append(varianten, toVarianteDTO(variante))
	}

	return produktDTO{
		ID:        p.ID,
		Name:      p.Name,
		Kategorie: string(p.Kategorie),
		Status:    string(p.Status),
		Varianten: varianten,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func toProduktDTOs(produkte []product.Produkt) []produktDTO {
	produktDTOs := make([]produktDTO, 0, len(produkte))
	for _, produkt := range produkte {
		produktDTOs = append(produktDTOs, toProduktDTO(produkt))
	}

	return produktDTOs
}

func (h *QueryHandler) GetAllProductsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		products, err := application.GetAllProducts(ctx, h.ProductRepo)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getAllProductsResponse{Produkte: toProduktDTOs(products)})
	}
}

type getActiveProductsResponse struct {
	Produkte []produktDTO `json:"produkte"`
}

func (h *QueryHandler) GetActiveProductsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		products, err := application.GetActiveProducts(ctx, h.ProductRepo)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getActiveProductsResponse{Produkte: toProduktDTOs(products)})
	}
}
