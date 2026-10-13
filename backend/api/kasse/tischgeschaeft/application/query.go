package application

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/kasse"
	t "github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/internal/logging"
)

type TischStateView struct {
	TischID              int
	TischName            string
	Subject              string
	SaldoCents           int
	UnbezahltePositionen []kasse.Position
	// FuerMichErledigt: die anfragende Servicekraft hat an diesem Tisch keine eigenen unbezahlten Positionen mehr.
	FuerMichErledigt bool
}

type Query struct {
	TischRepo           tischRepo
	EventRepo           eventRepo
	FavoritRepo         favoritRepo
	KassensitzungenRepo kassensitzungenRepo
}

func (q Query) aktiveKassensitzungNr(ctx context.Context) (int, error) {
	ks, err := q.KassensitzungenRepo.GetAktiveKassensitzung(ctx)
	if err != nil {
		return 0, err
	}
	if ks == nil {
		return 0, nil
	}
	return ks.ZNr, nil
}

func (q Query) GetAktiveTische(ctx context.Context) ([]t.AktiverTisch, error) {
	log := logging.FromContext(ctx)

	kassensitzungNr, err := q.aktiveKassensitzungNr(ctx)
	if err != nil {
		log.Error("Failed to get aktive kassensitzung for active tische", "error", err)
		return nil, ErrDatabase
	}

	tische, err := q.TischRepo.GetAktiveTische(ctx, kassensitzungNr)
	if err != nil {
		log.Error("Failed to retrieve active tische", "error", err)
		return nil, ErrDatabase
	}

	log.Info("Retrieved active tische", "count", len(tische))
	return tische, nil
}

func (q Query) GetTischState(ctx context.Context, tischID int, userID int) (TischStateView, error) {
	log := logging.FromContext(ctx)

	tisch, err := q.TischRepo.GetTisch(ctx, tischID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn("Tisch not found", "tisch_id", tischID)
			return TischStateView{}, ErrTischNotFound
		}
		log.Error("Failed to retrieve tisch", "error", err, "tisch_id", tischID)
		return TischStateView{}, ErrDatabase
	}

	kassensitzungNr, err := q.aktiveKassensitzungNr(ctx)
	if err != nil {
		log.Error("Failed to get aktive kassensitzung", "error", err, "tisch_id", tischID)
		return TischStateView{}, ErrDatabase
	}

	subject := kasse.TischSessionSubject(kassensitzungNr, tischID)

	state, err := q.EventRepo.ReadTischSession(ctx, subject)
	if err != nil {
		log.Error("Failed to read tisch session", "error", err, "tisch_id", tischID)
		return TischStateView{}, ErrDatabase
	}

	log.Info("Retrieved tisch state", "tisch_id", tischID, "saldo_cents", state.SaldoCents)
	return TischStateView{
		TischID:              tisch.ID,
		TischName:            tisch.Name,
		Subject:              subject,
		SaldoCents:           state.SaldoCents,
		UnbezahltePositionen: state.UnbezahltePositionen,
		FuerMichErledigt:     kasse.ComputeEigeneArbeitAnTisch(state, userID).Erledigt,
	}, nil
}

func (q Query) GetAktiveTischeMitFavoriten(ctx context.Context, userID int) ([]t.AktiverTischMitFavorit, error) {
	log := logging.FromContext(ctx)

	kassensitzungNr, err := q.aktiveKassensitzungNr(ctx)
	if err != nil {
		log.Error("Failed to get aktive kassensitzung", "error", err, "user_id", userID)
		return nil, ErrDatabase
	}

	tische, err := q.TischRepo.GetAktiveTischeMitFavoriten(ctx, userID, kassensitzungNr)
	if err != nil {
		log.Error("Failed to retrieve active tische mit favoriten", "error", err, "user_id", userID)
		return nil, ErrDatabase
	}

	log.Info("Retrieved active tische mit favoriten", "user_id", userID, "count", len(tische))
	return tische, nil
}

func (q Query) GetMeineTischeState(ctx context.Context, userID int) ([]TischStateView, error) {
	log := logging.FromContext(ctx)

	favoritIDs, err := q.FavoritRepo.GetByUser(ctx, userID)
	if err != nil {
		log.Error("Failed to retrieve favorit IDs", "error", err, "user_id", userID)
		return nil, ErrDatabase
	}

	if len(favoritIDs) == 0 {
		log.Debug("No favoriten for user", "user_id", userID)
		return []TischStateView{}, nil
	}

	kassensitzungNr, err := q.aktiveKassensitzungNr(ctx)
	if err != nil {
		log.Error("Failed to get aktive kassensitzung", "error", err, "user_id", userID)
		return nil, ErrDatabase
	}

	// Ein Batch statt N+1: Name + projizierte Session aller Favoriten in einer Query.
	states, err := q.EventRepo.ReadFavoritenTischStates(ctx, favoritIDs, kassensitzungNr)
	if err != nil {
		log.Error("Failed to batch-read favoriten tisch states", "error", err, "user_id", userID)
		return nil, ErrDatabase
	}

	views := make([]TischStateView, 0, len(favoritIDs))
	for _, tischID := range favoritIDs {
		entry, ok := states[tischID]
		if !ok {
			// Kein Tisch (gelöscht/unbekannt): verwaisten Favoriten überspringen. Ein Abbruch nähme
			// der Servicekraft wegen eines einzigen gelöschten Tisches die gesamte Tischübersicht.
			log.Warn("Skipped favorit with unresolvable tisch", "tisch_id", tischID, "user_id", userID)
			continue
		}

		state := entry.Session
		views = append(views, TischStateView{
			TischID:              tischID,
			TischName:            entry.Name,
			Subject:              kasse.TischSessionSubject(kassensitzungNr, tischID),
			SaldoCents:           state.SaldoCents,
			UnbezahltePositionen: state.UnbezahltePositionen,
			FuerMichErledigt:     kasse.ComputeEigeneArbeitAnTisch(state, userID).Erledigt,
		})
	}

	log.Info("Retrieved meine tische state", "user_id", userID, "count", len(views))
	return views, nil
}

func (q Query) GetTischHistorie(ctx context.Context, tischID int) ([]kasse.HistorieEintrag, error) {
	log := logging.FromContext(ctx)

	kassensitzungNr, err := q.aktiveKassensitzungNr(ctx)
	if err != nil {
		log.Error("Failed to get aktive kassensitzung for historie", "error", err, "tisch_id", tischID)
		return nil, ErrDatabase
	}

	subject := kasse.TischSessionSubject(kassensitzungNr, tischID)
	events, err := q.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		log.Error("Failed to read events for tisch", "error", err, "tisch_id", tischID)
		return nil, ErrDatabase
	}

	historie, err := kasse.GetHistorieFromEvents(events)
	if err != nil {
		log.Error("Failed to build historie from events", "tisch_id", tischID, "error", err)
		return nil, err
	}

	log.Info("Retrieved historie for tisch", "tisch_id", tischID, "historie_count", len(historie))
	return historie, nil
}
