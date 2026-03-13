package http

import (
	"context"
	"net/http"

	"github.com/nicograef/jotti/backend/api/helper"
	t "github.com/nicograef/jotti/backend/domain/table"
)

type query interface {
	GetAllTische(ctx context.Context) ([]t.Tisch, error)
	GetAktiveTische(ctx context.Context) ([]t.Tisch, error)
	GetTischHistorie(ctx context.Context, tischID int) ([]any, error)
	GetTischState(ctx context.Context, tischID int) (t.TischState, error)
}

type QueryHandler struct {
	Query query
}

type getAllTischeResponse struct {
	Tische []t.Tisch `json:"tische"`
}

func (h QueryHandler) GetAllTischeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tische, err := h.Query.GetAllTische(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getAllTischeResponse{Tische: tische})
	}
}

type aktiverTisch struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type getAktiveTischeResponse struct {
	Tische []aktiverTisch `json:"tische"`
}

func (h QueryHandler) GetAktiveTischeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tische, err := h.Query.GetAktiveTische(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		aktiveTische := make([]aktiverTisch, len(tische))
		for i, tisch := range tische {
			aktiveTische[i] = aktiverTisch{
				ID:   tisch.ID,
				Name: tisch.Name,
			}
		}

		helper.SendResponse(w, getAktiveTischeResponse{Tische: aktiveTische})
	}
}

type getTischHistorie struct {
	TischID int `json:"tischId"`
}

type getTischHistorieResponse struct {
	Historie []any `json:"historie"`
}

func (h QueryHandler) GetTischHistorieHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getTischHistorie{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		historie, err := h.Query.GetTischHistorie(r.Context(), body.TischID)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getTischHistorieResponse{Historie: historie})
	}
}

type getTischState struct {
	TischID int `json:"tischId"`
}

type getTischStateResponse struct {
	TischID                int          `json:"tischId"`
	TischName              string       `json:"tischName"`
	SaldoCents             int          `json:"saldoCents"`
	UnbezahltePositionen   []t.Position `json:"unbezahltePositionen"`
	UngeliefertePositionen []t.Position `json:"ungeliefertePositionen"`
	GesamtZahlungenCents   int          `json:"gesamtZahlungenCents"`
}

func (h QueryHandler) GetTischStateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getTischState{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		state, err := h.Query.GetTischState(r.Context(), body.TischID)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		unbezahlt := state.UnbezahltePositionen
		if unbezahlt == nil {
			unbezahlt = []t.Position{}
		}
		ungeliefert := state.UngeliefertePositionen
		if ungeliefert == nil {
			ungeliefert = []t.Position{}
		}

		helper.SendResponse(w, getTischStateResponse{
			TischID:                state.TischID,
			TischName:              state.TischName,
			SaldoCents:             state.SaldoCents,
			UnbezahltePositionen:   unbezahlt,
			UngeliefertePositionen: ungeliefert,
			GesamtZahlungenCents:   state.GesamtZahlungenCents,
		})
	}
}
