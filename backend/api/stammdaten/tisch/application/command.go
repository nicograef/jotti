package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/rs/zerolog"
)

type tischRepo interface {
	GetTable(ctx context.Context, id int) (tisch.Tisch, error)
	CreateTable(ctx context.Context, t tisch.Tisch) (int, error)
	UpdateTable(ctx context.Context, t tisch.Tisch) error
	GetAllTables(ctx context.Context) ([]tisch.Tisch, error)
	TischHatOffenenSaldo(ctx context.Context, tischID int) (bool, error)
}

type favoritRepo interface {
	Add(ctx context.Context, userID, tischID int) error
	Remove(ctx context.Context, userID, tischID int) error
	RemoveByTisch(ctx context.Context, tischID int) error
}

type Command struct {
	TischRepo   tischRepo
	FavoritRepo favoritRepo
}

func (c Command) FavoritHinzufuegen(ctx context.Context, userID, tischID int) error {
	log := zerolog.Ctx(ctx)

	t, err := c.TischRepo.GetTable(ctx, tischID)
	if err != nil {
		return fromRepositoryError(err, log, tischID)
	}

	if t.Status != tisch.ActiveStatus {
		log.Warn().Int("tisch_id", tischID).Str("status", string(t.Status)).Msg("Tisch is not active")
		return ErrTischNotActive
	}

	if err := c.FavoritRepo.Add(ctx, userID, tischID); err != nil {
		log.Error().Err(err).Int("user_id", userID).Int("tisch_id", tischID).Msg("Failed to add favorit")
		return ErrDatabase
	}

	log.Info().Int("user_id", userID).Int("tisch_id", tischID).Msg("Favorit added")
	return nil
}

func (c Command) FavoritEntfernen(ctx context.Context, userID, tischID int) error {
	log := zerolog.Ctx(ctx)

	if err := c.FavoritRepo.Remove(ctx, userID, tischID); err != nil {
		log.Error().Err(err).Int("user_id", userID).Int("tisch_id", tischID).Msg("Failed to remove favorit")
		return ErrDatabase
	}

	log.Info().Int("user_id", userID).Int("tisch_id", tischID).Msg("Favorit removed")
	return nil
}

func (c Command) TischErstellen(ctx context.Context, name string) (int, error) {
	log := zerolog.Ctx(ctx)

	tisch, err := tisch.NewTisch(name)
	if err != nil {
		log.Warn().Err(err).Str("tisch_name", name).Msg("Invalid tisch data")
		return 0, ErrInvalidTischData
	}

	id, err := c.TischRepo.CreateTable(ctx, tisch)
	if err != nil {
		return 0, fromRepositoryError(err, log, 0)
	}

	log.Info().Int("tisch_id", id).Msg("Tisch created")
	return id, nil
}

func (c Command) TischAktualisieren(ctx context.Context, id int, name string) error {
	log := zerolog.Ctx(ctx)

	tisch, err := c.TischRepo.GetTable(ctx, id)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	err = tisch.Rename(name)
	if err != nil {
		log.Warn().Err(err).Int("tisch_id", id).Msg("Invalid tisch data for update")
		return ErrInvalidTischData
	}

	err = c.TischRepo.UpdateTable(ctx, tisch)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	log.Info().Int("tisch_id", id).Msg("Tisch updated")
	return nil
}

func (c Command) TischAktivieren(ctx context.Context, id int) error {
	return c.applyTischStatusChange(ctx, id, false, "Tisch activated", func(t *tisch.Tisch) { t.Activate() })
}

func (c Command) TischDeaktivieren(ctx context.Context, id int) error {
	return c.applyTischStatusChange(ctx, id, true, "Tisch deactivated", func(t *tisch.Tisch) { t.Deactivate() })
}

func (c Command) TischLoeschen(ctx context.Context, id int) error {
	return c.applyTischStatusChange(ctx, id, true, "Tisch deleted", func(t *tisch.Tisch) { t.Delete() })
}

// applyTischStatusChange lädt den Tisch, wendet action an und persistiert das
// Ergebnis. Ist guardSaldo gesetzt (Deaktivieren, Löschen), wird ein Tisch mit
// offenem Saldo in der offenen Kassensitzung abgelehnt — das Backend erzwingt
// den Schutz als Single Source of Truth, unabhängig davon, was das Frontend
// anbietet. Führt action den Tisch in den Status 'deleted', werden zusätzlich
// seine Favoriten-Markierungen entfernt.
func (c Command) applyTischStatusChange(ctx context.Context, id int, guardSaldo bool, successMsg string, action func(*tisch.Tisch)) error {
	log := zerolog.Ctx(ctx)

	t, err := c.TischRepo.GetTable(ctx, id)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	if guardSaldo {
		hatSaldo, err := c.TischRepo.TischHatOffenenSaldo(ctx, id)
		if err != nil {
			log.Error().Err(err).Int("tisch_id", id).Msg("Failed to check open saldo for status change")
			return ErrDatabase
		}
		if hatSaldo {
			log.Warn().Int("tisch_id", id).Msg("Cannot change status of tisch with open saldo")
			return ErrTischSaldoOffen
		}
	}

	action(&t)

	// Ein gelöschter Tisch erscheint nicht mehr in der Tischauswahl; eine
	// zurückbleibende Markierung wäre für die betroffene Servicekraft weder
	// sichtbar noch abwählbar und hinge dauerhaft in ihrer Tischübersicht.
	// Der Cleanup läuft vor dem Statuswechsel: Scheitert er, bleibt der Tisch
	// unangetastet und der Löschvorgang ist unverändert wiederholbar.
	// Ein deaktivierter Tisch bleibt bewusst markiert — er kommt wieder.
	if t.Status == tisch.DeletedStatus {
		if err := c.FavoritRepo.RemoveByTisch(ctx, id); err != nil {
			log.Error().Err(err).Int("tisch_id", id).Msg("Failed to remove favoriten of deleted tisch")
			return ErrDatabase
		}
	}

	if err := c.TischRepo.UpdateTable(ctx, t); err != nil {
		return fromRepositoryError(err, log, id)
	}
	log.Info().Int("tisch_id", id).Msg(successMsg)
	return nil
}
