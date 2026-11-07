package http

import (
	"context"
	"net/http"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/kasse/tischgeschaeft/application"
	"github.com/nicograef/jotti/backend/api/middleware"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tisch"
)

type query interface {
	GetTischHistorie(ctx context.Context, tischID int) ([]kasse.HistorieEintrag, error)
	GetTischState(ctx context.Context, tischID int, userID int) (application.TischStateView, error)
	GetAktiveTischeMitFavoriten(ctx context.Context, userID int) ([]tisch.AktiverTischMitFavorit, error)
	GetMeineTischeState(ctx context.Context, userID int) ([]application.TischStateView, error)
}

type QueryHandler struct {
	Query query
}

type getTischHistorieRequest struct {
	TischID int `json:"tischId"`
}

var getTischHistorieSchema = z.Struct(z.Shape{
	"TischID": tisch.TischIDSchema.Required(),
})

type getTischHistorieResponse struct {
	Historie []any `json:"historie"`
}

type position struct {
	PositionID       string `json:"positionId"`
	VarianteID       int    `json:"varianteId"`
	ProduktName      string `json:"produktName"`
	VarianteName     string `json:"varianteName"`
	Kategorie        string `json:"kategorie"`
	Steuersatz       string `json:"steuersatz"`
	EinzelpreisCents int    `json:"einzelpreisCents"`
	Menge            int    `json:"menge"`
	BestellerUserID  int    `json:"bestellerUserId"`
	BestellerName    string `json:"bestellerName"`
}

func toPosition(p kasse.Position) position {
	return position{
		PositionID:       p.PositionID,
		VarianteID:       p.VarianteID,
		ProduktName:      p.ProduktName,
		VarianteName:     p.VarianteName,
		Kategorie:        p.Kategorie,
		Steuersatz:       p.Steuersatz,
		EinzelpreisCents: p.EinzelpreisCents,
		Menge:            p.Menge,
		BestellerUserID:  p.BestellerUserID,
		BestellerName:    p.BestellerName,
	}
}

func toPositionen(positionen []kasse.Position) []position {
	positionenResponse := make([]position, 0, len(positionen))
	for _, position := range positionen {
		positionenResponse = append(positionenResponse, toPosition(position))
	}

	return positionenResponse
}

type bestellung struct {
	Art                    string     `json:"art"`
	ID                     string     `json:"id"`
	UserID                 int        `json:"userId"`
	UserName               string     `json:"userName"`
	TischID                int        `json:"tischId"`
	Positionen             []position `json:"positionen"`
	GesamtPreisCents       int        `json:"gesamtPreisCents"`
	Kommentar              string     `json:"kommentar"`
	AufgenommenAm          time.Time  `json:"aufgenommenAm"`
	StornierbarePositionen []position `json:"stornierbarePositionen"`
	UmbuchbarePositionen   []position `json:"umbuchbarePositionen"`
}

func toBestellung(eintrag kasse.HistorieEintrag) bestellung {
	b := *eintrag.Bestellung
	return bestellung{
		Art:                    string(kasse.HistorieEintragBestellung),
		ID:                     b.ID,
		UserID:                 b.UserID,
		UserName:               b.UserName,
		TischID:                b.TischID,
		Positionen:             toPositionen(b.Positionen),
		GesamtPreisCents:       b.GesamtPreisCents,
		Kommentar:              b.Kommentar,
		AufgenommenAm:          b.AufgenommenAm,
		StornierbarePositionen: toPositionen(eintrag.StornierbarePositionen),
		UmbuchbarePositionen:   toPositionen(eintrag.UmbuchbarePositionen),
	}
}

type zahlung struct {
	Art                string     `json:"art"`
	ID                 string     `json:"id"`
	UserID             int        `json:"userId"`
	UserName           string     `json:"userName"`
	TischID            int        `json:"tischId"`
	Positionen         []position `json:"positionen"`
	GesamtZahlungCents int        `json:"gesamtZahlungCents"`
	Kommentar          string     `json:"kommentar"`
	KassiertAm         time.Time  `json:"kassiertAm"`
}

func toZahlung(z kasse.Zahlung) zahlung {
	return zahlung{
		Art:                string(kasse.HistorieEintragZahlung),
		ID:                 z.ID,
		UserID:             z.UserID,
		UserName:           z.UserName,
		TischID:            z.TischID,
		Positionen:         toPositionen(z.Positionen),
		GesamtZahlungCents: z.GesamtZahlungCents,
		Kommentar:          z.Kommentar,
		KassiertAm:         z.KassiertAm,
	}
}

type stornierung struct {
	Art                    string     `json:"art"`
	ID                     string     `json:"id"`
	UserID                 int        `json:"userId"`
	UserName               string     `json:"userName"`
	TischID                int        `json:"tischId"`
	Positionen             []position `json:"positionen"`
	GesamtStornierungCents int        `json:"gesamtStornierungCents"`
	Kommentar              string     `json:"kommentar"`
	StorniertAm            time.Time  `json:"storniertAm"`
	// BarRueckgabe: true = kassenwirksame Warenrücknahme, false = geldneutrale Korrektur.
	BarRueckgabe bool `json:"barRueckgabe"`
}

func toStornierung(s kasse.Stornierung) stornierung {
	return stornierung{
		Art:                    string(kasse.HistorieEintragStornierung),
		ID:                     s.ID,
		UserID:                 s.UserID,
		UserName:               s.UserName,
		TischID:                s.TischID,
		Positionen:             toPositionen(s.Positionen),
		GesamtStornierungCents: s.GesamtStornierungCents,
		Kommentar:              s.Kommentar,
		StorniertAm:            s.StorniertAm,
		BarRueckgabe:           s.BarRueckgabe,
	}
}

type umbuchung struct {
	Art                    string     `json:"art"`
	ID                     string     `json:"id"`
	UserID                 int        `json:"userId"`
	UserName               string     `json:"userName"`
	TischID                int        `json:"tischId"`
	QuellTischID           int        `json:"quellTischId"`
	ZielTischID            int        `json:"zielTischId"`
	Positionen             []position `json:"positionen"`
	GesamtCents            int        `json:"gesamtCents"`
	Kommentar              string     `json:"kommentar"`
	BenutzerKommentar      string     `json:"benutzerKommentar"`
	UmgebuchtAm            time.Time  `json:"umgebuchtAm"`
	StornierbarePositionen []position `json:"stornierbarePositionen"`
	UmbuchbarePositionen   []position `json:"umbuchbarePositionen"`
}

func toUmbuchung(eintrag kasse.HistorieEintrag) umbuchung {
	u := *eintrag.Umbuchung
	return umbuchung{
		Art:                    string(kasse.HistorieEintragUmbuchung),
		ID:                     u.ID,
		UserID:                 u.UserID,
		UserName:               u.UserName,
		TischID:                u.TischID,
		QuellTischID:           u.QuellTischID,
		ZielTischID:            u.ZielTischID,
		Positionen:             toPositionen(u.Positionen),
		GesamtCents:            u.GesamtCents,
		Kommentar:              u.Kommentar,
		BenutzerKommentar:      u.BenutzerKommentar,
		UmgebuchtAm:            u.UmgebuchtAm,
		StornierbarePositionen: toPositionen(eintrag.StornierbarePositionen),
		UmbuchbarePositionen:   toPositionen(eintrag.UmbuchbarePositionen),
	}
}

func toHistorie(eintraege []kasse.HistorieEintrag) []any {
	historieResponse := make([]any, 0, len(eintraege))
	for _, eintrag := range eintraege {
		switch eintrag.Art {
		case kasse.HistorieEintragBestellung:
			if eintrag.Bestellung != nil {
				historieResponse = append(historieResponse, toBestellung(eintrag))
			}
		case kasse.HistorieEintragZahlung:
			if eintrag.Zahlung != nil {
				historieResponse = append(historieResponse, toZahlung(*eintrag.Zahlung))
			}
		case kasse.HistorieEintragStornierung:
			if eintrag.Stornierung != nil {
				historieResponse = append(historieResponse, toStornierung(*eintrag.Stornierung))
			}
		case kasse.HistorieEintragUmbuchung:
			if eintrag.Umbuchung != nil {
				historieResponse = append(historieResponse, toUmbuchung(eintrag))
			}
		}
	}

	return historieResponse
}

func (h *QueryHandler) GetTischHistorieHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := getTischHistorieRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, getTischHistorieSchema) {
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

var getTischStateSchema = z.Struct(z.Shape{
	"TischID": tisch.TischIDSchema.Required(),
})

type tischState struct {
	TischID              int        `json:"tischId"`
	TischName            string     `json:"tischName"`
	SaldoCents           int        `json:"saldoCents"`
	UnbezahltePositionen []position `json:"unbezahltePositionen"`
	FuerMichErledigt     bool       `json:"fuerMichErledigt"`
}

func toTischState(s application.TischStateView) tischState {
	return tischState{
		TischID:              s.TischID,
		TischName:            s.TischName,
		SaldoCents:           s.SaldoCents,
		UnbezahltePositionen: toPositionen(s.UnbezahltePositionen),
		FuerMichErledigt:     s.FuerMichErledigt,
	}
}

func (h *QueryHandler) GetTischStateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _, ok := middleware.UserFromContext(r.Context())
		if !ok {
			helper.SendServerError(w)
			return
		}

		body := getTischStateRequest{}
		if !helper.ReadAndValidateBody(w, r, &body, getTischStateSchema) {
			return
		}

		state, err := h.Query.GetTischState(r.Context(), body.TischID, userID)
		if err != nil {
			helper.MapError(w, err, []helper.ErrorCode{
				{Err: application.ErrTischNotFound, Code: "tisch_not_found"},
			})
			return
		}

		helper.SendResponse(w, toTischState(state))
	}
}

type aktiverTischMitFavorit struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	SaldoCents int    `json:"saldoCents"`
	IstFavorit bool   `json:"istFavorit"`
}

type getAktiveTischeMitFavoritenResponse struct {
	Tische []aktiverTischMitFavorit `json:"tische"`
}

func (h *QueryHandler) GetAktiveTischeMitFavoritenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _, ok := middleware.UserFromContext(r.Context())
		if !ok {
			helper.SendServerError(w)
			return
		}

		tische, err := h.Query.GetAktiveTischeMitFavoriten(r.Context(), userID)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		dtos := make([]aktiverTischMitFavorit, len(tische))
		for i, aktiv := range tische {
			dtos[i] = aktiverTischMitFavorit{
				ID:         aktiv.ID,
				Name:       aktiv.Name,
				SaldoCents: aktiv.SaldoCents,
				IstFavorit: aktiv.IstFavorit,
			}
		}

		helper.SendResponse(w, getAktiveTischeMitFavoritenResponse{Tische: dtos})
	}
}

type getMeineTischeStateResponse struct {
	Tische []tischState `json:"tische"`
}

func (h *QueryHandler) GetMeineTischeStateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _, ok := middleware.UserFromContext(r.Context())
		if !ok {
			helper.SendServerError(w)
			return
		}

		tische, err := h.Query.GetMeineTischeState(r.Context(), userID)
		if err != nil {
			helper.SendServerError(w)
			return
		}

		dtos := make([]tischState, len(tische))
		for i, state := range tische {
			dtos[i] = toTischState(state)
		}

		helper.SendResponse(w, getMeineTischeStateResponse{Tische: dtos})
	}
}
