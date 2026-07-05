package application

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/kasse"
	t "github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/rs/zerolog"
)

// TischStateView combines a TischSession with the tisch name for display purposes.
type TischStateView struct {
	TischID               int
	TischName             string
	Subject               string
	SaldoCents            int
	UnbezahltePositionen  []kasse.Position
	AusstehendePositionen []kasse.Position
	GesamtZahlungenCents  int
	// FuerMichErledigt ist true, wenn die anfragende Servicekraft an diesem Tisch
	// keine eigenen ausstehenden und keine eigenen unbezahlten Positionen mehr hat.
	FuerMichErledigt bool
}

type Query struct {
	TischRepo           tischRepo
	EventRepo           eventRepo
	FavoritRepo         favoritRepo
	KassensitzungenRepo kassensitzungenRepo
}

func (q Query) GetAktiveTische(ctx context.Context) ([]t.AktiverTisch, error) {
	log := zerolog.Ctx(ctx)

	kassensitzungNr := 0
	ks, err := q.KassensitzungenRepo.GetOffeneKassensitzung(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get offene kassensitzung for active tische")
		return nil, ErrDatabase
	}
	if ks != nil {
		kassensitzungNr = ks.ZNr
	}

	tische, err := q.TischRepo.GetActiveTables(ctx, kassensitzungNr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve active tische")
		return nil, ErrDatabase
	}

	log.Info().Int("count", len(tische)).Msg("Retrieved active tische")
	return tische, nil
}

func (q Query) GetTischState(ctx context.Context, tischID int, userID int) (TischStateView, error) {
	log := zerolog.Ctx(ctx)

	tisch, err := q.TischRepo.GetTable(ctx, tischID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("tisch_id", tischID).Msg("Tisch not found")
			return TischStateView{}, ErrTischNotFound
		}
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to retrieve tisch")
		return TischStateView{}, ErrDatabase
	}

	kassensitzungNr := 0
	ks, err := q.KassensitzungenRepo.GetOffeneKassensitzung(ctx)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to get offene kassensitzung")
		return TischStateView{}, ErrDatabase
	}
	if ks != nil {
		kassensitzungNr = ks.ZNr
	}

	subject := kasse.TischSessionSubject(kassensitzungNr, tischID)

	state, err := q.EventRepo.ReadTischSession(ctx, subject)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to read tisch session")
		return TischStateView{}, ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Int("saldo_cents", state.SaldoCents).Msg("Retrieved tisch state")
	return TischStateView{
		TischID:               tisch.ID,
		TischName:             tisch.Name,
		Subject:               subject,
		SaldoCents:            state.SaldoCents,
		UnbezahltePositionen:  state.UnbezahltePositionen,
		AusstehendePositionen: state.AusstehendePositionen,
		GesamtZahlungenCents:  state.GesamtZahlungenCents,
		FuerMichErledigt:      kasse.ComputeEigeneArbeitAnTisch(state, userID).Erledigt,
	}, nil
}

func (q Query) GetAktiveTischeMitFavoriten(ctx context.Context, userID int) ([]t.AktiverTischMitFavorit, error) {
	log := zerolog.Ctx(ctx)

	kassensitzungNr := 0
	ks, err := q.KassensitzungenRepo.GetOffeneKassensitzung(ctx)
	if err != nil {
		log.Error().Err(err).Int("user_id", userID).Msg("Failed to get offene kassensitzung")
		return nil, ErrDatabase
	}
	if ks != nil {
		kassensitzungNr = ks.ZNr
	}

	tische, err := q.TischRepo.GetActiveTablesWithFavorites(ctx, userID, kassensitzungNr)
	if err != nil {
		log.Error().Err(err).Int("user_id", userID).Msg("Failed to retrieve active tische mit favoriten")
		return nil, ErrDatabase
	}

	log.Info().Int("user_id", userID).Int("count", len(tische)).Msg("Retrieved active tische mit favoriten")
	return tische, nil
}

func (q Query) GetMeineTischeState(ctx context.Context, userID int) ([]TischStateView, error) {
	log := zerolog.Ctx(ctx)

	favoritIDs, err := q.FavoritRepo.GetByUser(ctx, userID)
	if err != nil {
		log.Error().Err(err).Int("user_id", userID).Msg("Failed to retrieve favorit IDs")
		return nil, ErrDatabase
	}

	if len(favoritIDs) == 0 {
		log.Debug().Int("user_id", userID).Msg("No favoriten for user")
		return []TischStateView{}, nil
	}

	kassensitzungNr := 0
	ks, err := q.KassensitzungenRepo.GetOffeneKassensitzung(ctx)
	if err != nil {
		log.Error().Err(err).Int("user_id", userID).Msg("Failed to get offene kassensitzung")
		return nil, ErrDatabase
	}
	if ks != nil {
		kassensitzungNr = ks.ZNr
	}

	views := make([]TischStateView, 0, len(favoritIDs))
	for _, tischID := range favoritIDs {
		tisch, err := q.TischRepo.GetTable(ctx, tischID)
		if err != nil {
			log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to resolve tisch")
			return nil, ErrDatabase
		}

		subject := kasse.TischSessionSubject(kassensitzungNr, tischID)
		state, err := q.EventRepo.ReadTischSession(ctx, subject)
		if err != nil {
			log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to read tisch session")
			return nil, ErrDatabase
		}

		views = append(views, TischStateView{
			TischID:               tischID,
			TischName:             tisch.Name,
			Subject:               subject,
			SaldoCents:            state.SaldoCents,
			UnbezahltePositionen:  state.UnbezahltePositionen,
			AusstehendePositionen: state.AusstehendePositionen,
			GesamtZahlungenCents:  state.GesamtZahlungenCents,
			FuerMichErledigt:      kasse.ComputeEigeneArbeitAnTisch(state, userID).Erledigt,
		})
	}

	log.Info().Int("user_id", userID).Int("count", len(views)).Msg("Retrieved meine tische state")
	return views, nil
}

func (q Query) GetTischHistorie(ctx context.Context, tischID int) ([]kasse.HistorieEintrag, error) {
	log := zerolog.Ctx(ctx)

	kassensitzungNr := 0
	ks, err := q.KassensitzungenRepo.GetOffeneKassensitzung(ctx)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to get offene kassensitzung for historie")
		return nil, ErrDatabase
	}
	if ks != nil {
		kassensitzungNr = ks.ZNr
	}

	subject := kasse.TischSessionSubject(kassensitzungNr, tischID)
	events, err := q.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		log.Error().Int("tisch_id", tischID).Msg("Failed to read events for tisch")
		return nil, ErrDatabase
	}

	historie, err := kasse.GetHistorieFromEvents(events)
	if err != nil {
		log.Error().Int("tisch_id", tischID).Err(err).Msg("Failed to build historie from events")
		return nil, err
	}

	log.Info().Int("tisch_id", tischID).Int("historie_count", len(historie)).Msg("Retrieved historie for tisch")
	return historie, nil
}
