package application

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/rs/zerolog"
)

type produktRepo interface {
	GetProdukt(ctx context.Context, produktID int) (produkt.Produkt, error)
	CreateProdukt(ctx context.Context, produkt produkt.Produkt) (int, error)
	UpdateProdukt(ctx context.Context, produkt produkt.Produkt) error
	GetVariante(ctx context.Context, varianteID int) (produkt.Variante, error)
	CreateVariante(ctx context.Context, produktID int, variante produkt.Variante) (int, error)
	UpdateVariante(ctx context.Context, variante produkt.Variante) error
	GetAllProdukte(ctx context.Context) ([]produkt.Produkt, error)
	GetActiveProdukte(ctx context.Context) ([]produkt.Produkt, error)
	DeleteProduktMitVarianten(ctx context.Context, produkt produkt.Produkt) error
}

type Command struct {
	ProduktRepo produktRepo
}

// Produkt commands

func (c Command) CreateProdukt(ctx context.Context, name string, kategorie produkt.Kategorie, steuersatz steuer.Steuersatz) (int, error) {
	log := zerolog.Ctx(ctx)

	produkt, err := produkt.NewProdukt(name, kategorie, steuersatz)
	if err != nil {
		log.Warn().Err(err).Str("produkt_name", name).Msg("Invalid produkt data")
		return 0, ErrInvalidProduktData
	}

	produktID, err := c.ProduktRepo.CreateProdukt(ctx, produkt)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn().Err(err).Str("name", produkt.Name).Msg("Produkt name already exists")
			return 0, ErrProduktAlreadyExists
		}
		log.Error().Str("name", produkt.Name).Msg("Failed to create produkt")
		return 0, ErrDatabase
	}

	log.Info().Int("produkt_id", produktID).Msg("Produkt created")
	return produktID, nil
}

func (c Command) UpdateProdukt(ctx context.Context, produktID int, name string, kategorie produkt.Kategorie, steuersatz steuer.Steuersatz) error {
	log := zerolog.Ctx(ctx)

	produkt, err := c.ProduktRepo.GetProdukt(ctx, produktID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("produkt_id", produktID).Msg("Produkt not found for update")
			return ErrProduktNotFound
		}
		log.Error().Int("produkt_id", produktID).Msg("Failed to retrieve produkt for update")
		return ErrDatabase
	}

	err = produkt.UpdateDetails(name, kategorie, steuersatz)
	if err != nil {
		log.Warn().Err(err).Int("produkt_id", produktID).Msg("Invalid produkt data for update")
		return ErrInvalidProduktData
	}

	err = c.ProduktRepo.UpdateProdukt(ctx, produkt)
	if err != nil {
		log.Error().Err(err).Int("produkt_id", produktID).Msg("Failed to update produkt")
		return ErrDatabase
	}

	log.Info().Int("produkt_id", produktID).Msg("Produkt updated")
	return nil
}

// Variante commands

func (c Command) CreateVariante(ctx context.Context, produktID int, name string, preisCents int) (int, error) {
	log := zerolog.Ctx(ctx)

	// Verify produkt exists
	_, err := c.ProduktRepo.GetProdukt(ctx, produktID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("produkt_id", produktID).Msg("Produkt not found for variante creation")
			return 0, ErrProduktNotFound
		}
		log.Error().Int("produkt_id", produktID).Msg("Failed to retrieve produkt for variante creation")
		return 0, ErrDatabase
	}

	variante, err := produkt.NewVariante(name, preisCents)
	if err != nil {
		log.Warn().Err(err).Str("variante_name", name).Msg("Invalid variante data")
		return 0, ErrInvalidVarianteData
	}

	varianteID, err := c.ProduktRepo.CreateVariante(ctx, produktID, variante)
	if err != nil {
		log.Error().Int("produkt_id", produktID).Str("name", variante.Name).Msg("Failed to create variante")
		return 0, ErrDatabase
	}

	log.Info().Int("variante_id", varianteID).Int("produkt_id", produktID).Msg("Variante created")
	return varianteID, nil
}

func (c Command) UpdateVariante(ctx context.Context, varianteID int, name string, preisCents int) error {
	log := zerolog.Ctx(ctx)

	variante, err := c.ProduktRepo.GetVariante(ctx, varianteID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("variante_id", varianteID).Msg("Variante not found for update")
			return ErrVarianteNotFound
		}
		log.Error().Int("variante_id", varianteID).Msg("Failed to retrieve variante for update")
		return ErrDatabase
	}

	err = variante.UpdateDetails(name, preisCents)
	if err != nil {
		log.Warn().Err(err).Int("variante_id", varianteID).Msg("Invalid variante data for update")
		return ErrInvalidVarianteData
	}

	err = c.ProduktRepo.UpdateVariante(ctx, variante)
	if err != nil {
		log.Error().Err(err).Int("variante_id", varianteID).Msg("Failed to update variante")
		return ErrDatabase
	}

	log.Info().Int("variante_id", varianteID).Msg("Variante updated")
	return nil
}

func (c Command) ActivateVariante(ctx context.Context, varianteID int) error {
	return c.applyVarianteStatusChange(ctx, varianteID, "Variante activated", func(v *produkt.Variante) { v.Activate() })
}

func (c Command) DeactivateVariante(ctx context.Context, varianteID int) error {
	return c.applyVarianteStatusChange(ctx, varianteID, "Variante deactivated", func(v *produkt.Variante) { v.Deactivate() })
}

func (c Command) applyVarianteStatusChange(ctx context.Context, varianteID int, successMsg string, action func(*produkt.Variante)) error {
	log := zerolog.Ctx(ctx)

	variante, err := c.ProduktRepo.GetVariante(ctx, varianteID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("variante_id", varianteID).Msg("Variante not found for status change")
			return ErrVarianteNotFound
		}
		log.Error().Int("variante_id", varianteID).Msg("Failed to retrieve variante for status change")
		return ErrDatabase
	}

	action(&variante)

	if err := c.ProduktRepo.UpdateVariante(ctx, variante); err != nil {
		log.Error().Err(err).Int("variante_id", varianteID).Msg("Failed to update variante")
		return ErrDatabase
	}

	log.Info().Int("variante_id", varianteID).Msg(successMsg)
	return nil
}

func (c Command) DeleteProdukt(ctx context.Context, produktID int) error {
	log := zerolog.Ctx(ctx)

	produkt, err := c.ProduktRepo.GetProdukt(ctx, produktID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("produkt_id", produktID).Msg("Produkt not found for deletion")
			return ErrProduktNotFound
		}
		log.Error().Int("produkt_id", produktID).Msg("Failed to retrieve produkt for deletion")
		return ErrDatabase
	}

	for i := range produkt.Varianten {
		produkt.Varianten[i].Delete()
	}
	produkt.Delete()

	err = c.ProduktRepo.DeleteProduktMitVarianten(ctx, produkt)
	if err != nil {
		log.Error().Err(err).Int("produkt_id", produktID).Msg("Failed to delete produkt")
		return ErrDatabase
	}

	log.Info().Int("produkt_id", produktID).Msg("Produkt deleted")
	return nil
}

func (c Command) DeleteVariante(ctx context.Context, produktID int, varianteID int) error {
	log := zerolog.Ctx(ctx)

	_, err := c.ProduktRepo.GetProdukt(ctx, produktID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("produkt_id", produktID).Msg("Produkt not found for variante deletion")
			return ErrProduktNotFound
		}
		log.Error().Int("produkt_id", produktID).Msg("Failed to retrieve produkt for variante deletion")
		return ErrDatabase
	}

	variante, err := c.ProduktRepo.GetVariante(ctx, varianteID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Int("variante_id", varianteID).Msg("Variante not found for deletion")
			return ErrVarianteNotFound
		}
		log.Error().Int("variante_id", varianteID).Msg("Failed to retrieve variante for deletion")
		return ErrDatabase
	}

	variante.Delete()

	err = c.ProduktRepo.UpdateVariante(ctx, variante)
	if err != nil {
		log.Error().Err(err).Int("variante_id", varianteID).Msg("Failed to delete variante")
		return ErrDatabase
	}

	log.Info().Int("variante_id", varianteID).Msg("Variante deleted")
	return nil
}
