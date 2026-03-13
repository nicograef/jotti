package http

import (
	"context"
	"net/http"
	"time"

	"github.com/nicograef/jotti/backend/api/helper"
	t "github.com/nicograef/jotti/backend/domain/table"
)

type query interface {
	GetAllTische(ctx context.Context) ([]t.Tisch, error)
	GetAktiveTische(ctx context.Context) ([]t.Tisch, error)
	GetTischHistorie(ctx context.Context, tischID int) ([]t.HistorieEintrag, error)
	GetTischState(ctx context.Context, tischID int) (t.TischState, error)
}

type QueryHandler struct {
	Query query
}

type tischDTO struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type getAllTischeResponse struct {
	Tische []tischDTO `json:"tische"`
}

func toTischDTO(tisch t.Tisch) tischDTO {
	return tischDTO{
		ID:        tisch.ID,
		Name:      tisch.Name,
		Status:    string(tisch.Status),
		CreatedAt: tisch.CreatedAt,
		UpdatedAt: tisch.UpdatedAt,
	}
}

func toTischDTOs(tische []t.Tisch) []tischDTO {
	tischDTOs := make([]tischDTO, 0, len(tische))
	for _, tisch := range tische {
		tischDTOs = append(tischDTOs, toTischDTO(tisch))
	}

	return tischDTOs
}

func (h QueryHandler) GetAllTischeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tische, err := h.Query.GetAllTische(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getAllTischeResponse{Tische: toTischDTOs(tische)})
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

type positionDTO struct {
	PositionID   string `json:"positionId"`
	VarianteID   int    `json:"varianteId"`
	ProduktName  string `json:"produktName"`
	VarianteName string `json:"varianteName"`
	Kategorie    string `json:"kategorie"`
	Einzelpreis  int    `json:"einzelpreis"`
	Menge        int    `json:"menge"`
}

func toPositionDTO(p t.Position) positionDTO {
	return positionDTO{
		PositionID:   p.PositionID,
		VarianteID:   p.VarianteID,
		ProduktName:  p.ProduktName,
		VarianteName: p.VarianteName,
		Kategorie:    p.Kategorie,
		Einzelpreis:  p.Einzelpreis,
		Menge:        p.Menge,
	}
}

func toPositionDTOs(positionen []t.Position) []positionDTO {
	positionDTOs := make([]positionDTO, 0, len(positionen))
	for _, position := range positionen {
		positionDTOs = append(positionDTOs, toPositionDTO(position))
	}

	return positionDTOs
}

type bestellungDTO struct {
	ID               string        `json:"id"`
	UserID           int           `json:"userId"`
	TischID          int           `json:"tischId"`
	Positionen       []positionDTO `json:"positionen"`
	GesamtPreisCents int           `json:"gesamtPreisCents"`
	Kommentar        string        `json:"kommentar"`
	AufgegebenAm     time.Time     `json:"aufgegebenAm"`
}

func toBestellungDTO(b t.Bestellung) bestellungDTO {
	return bestellungDTO{
		ID:               b.ID,
		UserID:           b.UserID,
		TischID:          b.TischID,
		Positionen:       toPositionDTOs(b.Positionen),
		GesamtPreisCents: b.GesamtPreisCents,
		Kommentar:        b.Kommentar,
		AufgegebenAm:     b.AufgegebenAm,
	}
}

type lieferungDTO struct {
	ID          string        `json:"id"`
	UserID      int           `json:"userId"`
	TischID     int           `json:"tischId"`
	Positionen  []positionDTO `json:"positionen"`
	Kommentar   string        `json:"kommentar"`
	GeliefertAm time.Time     `json:"geliefertAm"`
}

func toLieferungDTO(l t.Lieferung) lieferungDTO {
	return lieferungDTO{
		ID:          l.ID,
		UserID:      l.UserID,
		TischID:     l.TischID,
		Positionen:  toPositionDTOs(l.Positionen),
		Kommentar:   l.Kommentar,
		GeliefertAm: l.GeliefertAm,
	}
}

type zahlungDTO struct {
	ID                 string        `json:"id"`
	UserID             int           `json:"userId"`
	TischID            int           `json:"tischId"`
	Positionen         []positionDTO `json:"positionen"`
	GesamtZahlungCents int           `json:"gesamtZahlungCents"`
	Kommentar          string        `json:"kommentar"`
	RegistriertAm      time.Time     `json:"registriertAm"`
}

func toZahlungDTO(z t.Zahlung) zahlungDTO {
	return zahlungDTO{
		ID:                 z.ID,
		UserID:             z.UserID,
		TischID:            z.TischID,
		Positionen:         toPositionDTOs(z.Positionen),
		GesamtZahlungCents: z.GesamtZahlungCents,
		Kommentar:          z.Kommentar,
		RegistriertAm:      z.RegistriertAm,
	}
}

type stornierungDTO struct {
	ID                     string        `json:"id"`
	UserID                 int           `json:"userId"`
	TischID                int           `json:"tischId"`
	Positionen             []positionDTO `json:"positionen"`
	GesamtStornierungCents int           `json:"gesamtStornierungCents"`
	Kommentar              string        `json:"kommentar"`
	StorniertAm            time.Time     `json:"storniertAm"`
}

func toStornierungDTO(s t.Stornierung) stornierungDTO {
	return stornierungDTO{
		ID:                     s.ID,
		UserID:                 s.UserID,
		TischID:                s.TischID,
		Positionen:             toPositionDTOs(s.Positionen),
		GesamtStornierungCents: s.GesamtStornierungCents,
		Kommentar:              s.Kommentar,
		StorniertAm:            s.StorniertAm,
	}
}

func toHistorieDTO(eintraege []t.HistorieEintrag) []any {
	historieDTO := make([]any, 0, len(eintraege))
	for _, eintrag := range eintraege {
		switch eintrag.Kind {
		case t.HistorieEintragBestellung:
			if eintrag.Bestellung != nil {
				historieDTO = append(historieDTO, toBestellungDTO(*eintrag.Bestellung))
			}
		case t.HistorieEintragLieferung:
			if eintrag.Lieferung != nil {
				historieDTO = append(historieDTO, toLieferungDTO(*eintrag.Lieferung))
			}
		case t.HistorieEintragZahlung:
			if eintrag.Zahlung != nil {
				historieDTO = append(historieDTO, toZahlungDTO(*eintrag.Zahlung))
			}
		case t.HistorieEintragStornierung:
			if eintrag.Stornierung != nil {
				historieDTO = append(historieDTO, toStornierungDTO(*eintrag.Stornierung))
			}
		}
	}

	return historieDTO
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

		helper.SendResponse(w, getTischHistorieResponse{Historie: toHistorieDTO(historie)})
	}
}

type getTischState struct {
	TischID int `json:"tischId"`
}

type getTischStateResponse struct {
	TischID                int           `json:"tischId"`
	TischName              string        `json:"tischName"`
	SaldoCents             int           `json:"saldoCents"`
	UnbezahltePositionen   []positionDTO `json:"unbezahltePositionen"`
	UngeliefertePositionen []positionDTO `json:"ungeliefertePositionen"`
	GesamtZahlungenCents   int           `json:"gesamtZahlungenCents"`
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

		unbezahlt := toPositionDTOs(state.UnbezahltePositionen)
		if unbezahlt == nil {
			unbezahlt = []positionDTO{}
		}
		ungeliefert := toPositionDTOs(state.UngeliefertePositionen)
		if ungeliefert == nil {
			ungeliefert = []positionDTO{}
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
