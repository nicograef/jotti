package http

import (
	"context"
	"net/http"
	"time"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/stammdaten/produkt/application"
	dom "github.com/nicograef/jotti/backend/domain/produkt"
)

type query interface {
	GetAllProducts(ctx context.Context) ([]application.ProduktMitVerkauf, error)
	GetActiveProducts(ctx context.Context) ([]dom.Produkt, error)
}

type QueryHandler struct {
	Query query
}

type variante struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	PreisCents int       `json:"preisCents"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type produkt struct {
	ID           int        `json:"id"`
	Name         string     `json:"name"`
	Kategorie    string     `json:"kategorie"`
	Steuersatz   string     `json:"steuersatz"`
	Status       string     `json:"status"`
	Varianten    []variante `json:"varianten"`
	HatVerkaeufe bool       `json:"hatVerkaeufe"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type getAllProductsResponse struct {
	Produkte []produkt `json:"produkte"`
}

func toVariante(v dom.Variante) variante {
	return variante{
		ID:         v.ID,
		Name:       v.Name,
		PreisCents: v.PreisCents,
		Status:     string(v.Status),
		CreatedAt:  v.CreatedAt,
		UpdatedAt:  v.UpdatedAt,
	}
}

func toProdukt(p dom.Produkt) produkt {
	varianten := make([]variante, 0, len(p.Varianten))
	for _, variante := range p.Varianten {
		varianten = append(varianten, toVariante(variante))
	}

	return produkt{
		ID:         p.ID,
		Name:       p.Name,
		Kategorie:  string(p.Kategorie),
		Steuersatz: string(p.Steuersatz),
		Status:     string(p.Status),
		Varianten:  varianten,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
}

func toProdukte(produkte []dom.Produkt) []produkt {
	produkteResponse := make([]produkt, 0, len(produkte))
	for i := range produkte {
		produkteResponse = append(produkteResponse, toProdukt(produkte[i]))
	}

	return produkteResponse
}

func toProdukteMitVerkauf(produkte []application.ProduktMitVerkauf) []produkt {
	produkteResponse := make([]produkt, 0, len(produkte))
	for i := range produkte {
		dto := toProdukt(produkte[i].Produkt)
		dto.HatVerkaeufe = produkte[i].HatVerkaeufe
		produkteResponse = append(produkteResponse, dto)
	}

	return produkteResponse
}

func (h *QueryHandler) GetAllProductsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		products, err := h.Query.GetAllProducts(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getAllProductsResponse{Produkte: toProdukteMitVerkauf(products)})
	}
}

type getActiveProductsResponse struct {
	Produkte []produkt `json:"produkte"`
}

func (h *QueryHandler) GetActiveProductsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		products, err := h.Query.GetActiveProducts(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getActiveProductsResponse{Produkte: toProdukte(products)})
	}
}
