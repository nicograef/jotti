package http

import (
	"context"
	"net/http"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/stammdaten/produkt/application"
	dom "github.com/nicograef/jotti/backend/domain/produkt"
)

type query interface {
	GetAllProdukte(ctx context.Context) ([]dom.Produkt, error)
	GetActiveProdukte(ctx context.Context) (application.AktiveProdukte, error)
}

type QueryHandler struct {
	Query query
}

type variante struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	PreisCents int    `json:"preisCents"`
	Status     string `json:"status"`
}

type produkt struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	Kategorie  string     `json:"kategorie"`
	Steuersatz string     `json:"steuersatz"`
	Pfand      bool       `json:"pfand"`
	Status     string     `json:"status"`
	Varianten  []variante `json:"varianten"`
}

type getAllProdukteResponse struct {
	Produkte []produkt `json:"produkte"`
}

type getAktiveProdukteResponse struct {
	Produkte      []produkt `json:"produkte"`
	PfandProdukte []produkt `json:"pfandProdukte"`
}

func toVariante(v dom.Variante) variante {
	return variante{
		ID:         v.ID,
		Name:       v.Name,
		PreisCents: v.PreisCents,
		Status:     string(v.Status),
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
		Pfand:      p.Pfand,
		Status:     string(p.Status),
		Varianten:  varianten,
	}
}

func toProdukte(produkte []dom.Produkt) []produkt {
	produkteResponse := make([]produkt, 0, len(produkte))
	for i := range produkte {
		produkteResponse = append(produkteResponse, toProdukt(produkte[i]))
	}

	return produkteResponse
}

func (h *QueryHandler) GetAllProdukteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		produkte, err := h.Query.GetAllProdukte(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getAllProdukteResponse{Produkte: toProdukte(produkte)})
	}
}

func (h *QueryHandler) GetActiveProdukteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		aktive, err := h.Query.GetActiveProdukte(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getAktiveProdukteResponse{
			Produkte:      toProdukte(aktive.Produkte),
			PfandProdukte: toProdukte(aktive.PfandProdukte),
		})
	}
}
