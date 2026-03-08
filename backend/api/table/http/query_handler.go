package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/table/application"
	t "github.com/nicograef/jotti/backend/domain/table"
)

type query interface {
	GetTisch(ctx context.Context, id int) (t.Tisch, error)
	GetAllTische(ctx context.Context) ([]t.Tisch, error)
	GetAktiveTische(ctx context.Context) ([]t.Tisch, error)
	GetTischHistorie(ctx context.Context, tischID int) ([]any, error)
	GetTischSaldo(ctx context.Context, tischID int) (int, error)
	GetTischUnbezahlt(ctx context.Context, tischID int) ([]t.Position, error)
	GetTischUngeliefert(ctx context.Context, tischID int) ([]t.Position, error)
}

type QueryHandler struct {
	Query query
}

type getTisch struct {
	ID int `json:"id"`
}

type getTischResponse struct {
	Tisch t.Tisch `json:"tisch"`
}

func (h QueryHandler) GetTischHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getTisch{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		tisch, err := h.Query.GetTisch(r.Context(), body.ID)
		if err != nil {
			if errors.Is(err, application.ErrTischNotFound) {
				helper.SendClientError(w, "tisch_not_found", nil)
				return
			} else {
				helper.SendServerError(w)
				return
			}
		}

		helper.SendResponse(w, getTischResponse{Tisch: tisch})
	}
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

type getTischSaldo struct {
	TischID int `json:"tischId"`
}

type getTischSaldoResponse struct {
	SaldoCents int `json:"saldoCents"`
}

func (h QueryHandler) GetTischSaldoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getTischSaldo{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		saldoCents, err := h.Query.GetTischSaldo(r.Context(), body.TischID)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getTischSaldoResponse{SaldoCents: saldoCents})
	}
}

type getTischUnbezahlt struct {
	TischID int `json:"tischId"`
}

type getTischUnbezahltResponse struct {
	Positionen []t.Position `json:"positionen"`
}

func (h QueryHandler) GetTischUnbezahltHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getTischUnbezahlt{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		positionen, err := h.Query.GetTischUnbezahlt(r.Context(), body.TischID)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getTischUnbezahltResponse{Positionen: positionen})
	}
}

type getTischUngeliefert struct {
	TischID int `json:"tischId"`
}

type getTischUngeliefertResponse struct {
	Positionen []t.Position `json:"positionen"`
}

func (h QueryHandler) GetTischUngeliefertHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getTischUngeliefert{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		positionen, err := h.Query.GetTischUngeliefert(r.Context(), body.TischID)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getTischUngeliefertResponse{Positionen: positionen})
	}
}
