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
	GetAktiveTische(ctx context.Context) ([]t.AktiverTisch, error)
	GetTischHistorie(ctx context.Context, tischID int) ([]t.HistorieEintrag, error)
	GetTischState(ctx context.Context, tischID int) (t.TischState, error)
}

type QueryHandler struct {
	Query query
}

type tisch struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type getAllTischeResponse struct {
	Tische []tisch `json:"tische"`
}

func toTisch(src t.Tisch) tisch {
	return tisch{
		ID:        src.ID,
		Name:      src.Name,
		Status:    string(src.Status),
		CreatedAt: src.CreatedAt,
		UpdatedAt: src.UpdatedAt,
	}
}

func toTische(tische []t.Tisch) []tisch {
	tischeResponse := make([]tisch, 0, len(tische))
	for _, tisch := range tische {
		tischeResponse = append(tischeResponse, toTisch(tisch))
	}

	return tischeResponse
}

func (h QueryHandler) GetAllTischeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tische, err := h.Query.GetAllTische(r.Context())
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getAllTischeResponse{Tische: toTische(tische)})
	}
}

type aktiverTisch struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	SaldoCents int    `json:"saldoCents"`
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
				ID:         tisch.ID,
				Name:       tisch.Name,
				SaldoCents: tisch.SaldoCents,
			}
		}

		helper.SendResponse(w, getAktiveTischeResponse{Tische: aktiveTische})
	}
}

type getTischHistorieRequest struct {
	TischID int `json:"tischId"`
}

type getTischHistorieResponse struct {
	Historie []any `json:"historie"`
}

type position struct {
	PositionID   string `json:"positionId"`
	VarianteID   int    `json:"varianteId"`
	ProduktName  string `json:"produktName"`
	VarianteName string `json:"varianteName"`
	Kategorie    string `json:"kategorie"`
	Einzelpreis  int    `json:"einzelpreis"`
	Menge        int    `json:"menge"`
}

func toPosition(p t.Position) position {
	return position{
		PositionID:   p.PositionID,
		VarianteID:   p.VarianteID,
		ProduktName:  p.ProduktName,
		VarianteName: p.VarianteName,
		Kategorie:    p.Kategorie,
		Einzelpreis:  p.Einzelpreis,
		Menge:        p.Menge,
	}
}

func toPositionen(positionen []t.Position) []position {
	positionenResponse := make([]position, 0, len(positionen))
	for _, position := range positionen {
		positionenResponse = append(positionenResponse, toPosition(position))
	}

	return positionenResponse
}

type bestellung struct {
	ID               string     `json:"id"`
	UserID           int        `json:"userId"`
	TischID          int        `json:"tischId"`
	Positionen       []position `json:"positionen"`
	GesamtPreisCents int        `json:"gesamtPreisCents"`
	Kommentar        string     `json:"kommentar"`
	AufgenommenAm    time.Time  `json:"aufgenommenAm"`
}

func toBestellung(b t.Bestellung) bestellung {
	return bestellung{
		ID:               b.ID,
		UserID:           b.UserID,
		TischID:          b.TischID,
		Positionen:       toPositionen(b.Positionen),
		GesamtPreisCents: b.GesamtPreisCents,
		Kommentar:        b.Kommentar,
		AufgenommenAm:    b.AufgenommenAm,
	}
}

type ausgabe struct {
	ID           string     `json:"id"`
	UserID       int        `json:"userId"`
	TischID      int        `json:"tischId"`
	Positionen   []position `json:"positionen"`
	Kommentar    string     `json:"kommentar"`
	AusgegebenAm time.Time  `json:"ausgegebenAm"`
}

func toAusgabe(a t.Ausgabe) ausgabe {
	return ausgabe{
		ID:           a.ID,
		UserID:       a.UserID,
		TischID:      a.TischID,
		Positionen:   toPositionen(a.Positionen),
		Kommentar:    a.Kommentar,
		AusgegebenAm: a.AusgegebenAm,
	}
}

type zahlung struct {
	ID                 string     `json:"id"`
	UserID             int        `json:"userId"`
	TischID            int        `json:"tischId"`
	Positionen         []position `json:"positionen"`
	GesamtZahlungCents int        `json:"gesamtZahlungCents"`
	Kommentar          string     `json:"kommentar"`
	KassiertAm         time.Time  `json:"kassiertAm"`
}

func toZahlung(z t.Zahlung) zahlung {
	return zahlung{
		ID:                 z.ID,
		UserID:             z.UserID,
		TischID:            z.TischID,
		Positionen:         toPositionen(z.Positionen),
		GesamtZahlungCents: z.GesamtZahlungCents,
		Kommentar:          z.Kommentar,
		KassiertAm:         z.KassiertAm,
	}
}

type stornierung struct {
	ID                     string     `json:"id"`
	UserID                 int        `json:"userId"`
	TischID                int        `json:"tischId"`
	Positionen             []position `json:"positionen"`
	GesamtStornierungCents int        `json:"gesamtStornierungCents"`
	Kommentar              string     `json:"kommentar"`
	StorniertAm            time.Time  `json:"storniertAm"`
}

func toStornierung(s t.Stornierung) stornierung {
	return stornierung{
		ID:                     s.ID,
		UserID:                 s.UserID,
		TischID:                s.TischID,
		Positionen:             toPositionen(s.Positionen),
		GesamtStornierungCents: s.GesamtStornierungCents,
		Kommentar:              s.Kommentar,
		StorniertAm:            s.StorniertAm,
	}
}

type auszahlung struct {
	ID          string    `json:"id"`
	UserID      int       `json:"userId"`
	TischID     int       `json:"tischId"`
	BetragCents int       `json:"betragCents"`
	Kommentar   string    `json:"kommentar"`
	GeleistetAm time.Time `json:"geleistetAm"`
}

func toAuszahlung(a t.Auszahlung) auszahlung {
	return auszahlung{
		ID:          a.ID,
		UserID:      a.UserID,
		TischID:     a.TischID,
		BetragCents: a.BetragCents,
		Kommentar:   a.Kommentar,
		GeleistetAm: a.GeleistetAm,
	}
}

func toHistorie(eintraege []t.HistorieEintrag) []any {
	historieResponse := make([]any, 0, len(eintraege))
	for _, eintrag := range eintraege {
		switch eintrag.Art {
		case t.HistorieEintragBestellung:
			if eintrag.Bestellung != nil {
				historieResponse = append(historieResponse, toBestellung(*eintrag.Bestellung))
			}
		case t.HistorieEintragAusgabe:
			if eintrag.Ausgabe != nil {
				historieResponse = append(historieResponse, toAusgabe(*eintrag.Ausgabe))
			}
		case t.HistorieEintragZahlung:
			if eintrag.Zahlung != nil {
				historieResponse = append(historieResponse, toZahlung(*eintrag.Zahlung))
			}
		case t.HistorieEintragStornierung:
			if eintrag.Stornierung != nil {
				historieResponse = append(historieResponse, toStornierung(*eintrag.Stornierung))
			}
		case t.HistorieEintragAuszahlung:
			if eintrag.Auszahlung != nil {
				historieResponse = append(historieResponse, toAuszahlung(*eintrag.Auszahlung))
			}
		}
	}

	return historieResponse
}

func (h QueryHandler) GetTischHistorieHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getTischHistorieRequest{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		historie, err := h.Query.GetTischHistorie(r.Context(), body.TischID)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		helper.SendResponse(w, getTischHistorieResponse{Historie: toHistorie(historie)})
	}
}

type getTischStateRequest struct {
	TischID int `json:"tischId"`
}

type getTischStateResponse struct {
	TischID               int        `json:"tischId"`
	TischName             string     `json:"tischName"`
	SaldoCents            int        `json:"saldoCents"`
	UnbezahltePositionen  []position `json:"unbezahltePositionen"`
	AusstehendePositionen []position `json:"ausstehendePositionen"`
	GesamtZahlungenCents  int        `json:"gesamtZahlungenCents"`
}

func (h QueryHandler) GetTischStateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getTischStateRequest{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		state, err := h.Query.GetTischState(r.Context(), body.TischID)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		unbezahlt := toPositionen(state.UnbezahltePositionen)
		if unbezahlt == nil {
			unbezahlt = []position{}
		}
		ausstehend := toPositionen(state.AusstehendePositionen)
		if ausstehend == nil {
			ausstehend = []position{}
		}

		helper.SendResponse(w, getTischStateResponse{
			TischID:               state.TischID,
			TischName:             state.TischName,
			SaldoCents:            state.SaldoCents,
			UnbezahltePositionen:  unbezahlt,
			AusstehendePositionen: ausstehend,
			GesamtZahlungenCents:  state.GesamtZahlungenCents,
		})
	}
}
