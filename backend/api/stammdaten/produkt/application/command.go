package application

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/internal/logging"
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
	VerschiebeProdukt(ctx context.Context, produktID int, hoch bool) error
	VerschiebeVariante(ctx context.Context, varianteID int, hoch bool) error
	SortiereVariantenAlphabetisch(ctx context.Context, produktID int) error
}

type Command struct {
	ProduktRepo produktRepo
}

func (c Command) CreateProdukt(ctx context.Context, name string, kategorie produkt.Kategorie, steuersatz steuer.Steuersatz) (int, error) {
	log := logging.FromContext(ctx)

	produkt, err := produkt.NewProdukt(name, kategorie, steuersatz)
	if err != nil {
		log.Warn("Invalid produkt data", "error", err, "produkt_name", name)
		return 0, ErrInvalidProduktData
	}

	produktID, err := c.ProduktRepo.CreateProdukt(ctx, produkt)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn("Produkt name already exists", "error", err, "name", produkt.Name)
			return 0, ErrProduktAlreadyExists
		}
		log.Error("Failed to create produkt", "error", err, "name", produkt.Name)
		return 0, ErrDatabase
	}

	log.Info("Produkt created", "produkt_id", produktID)
	return produktID, nil
}

func (c Command) UpdateProdukt(ctx context.Context, produktID int, name string, kategorie produkt.Kategorie, steuersatz steuer.Steuersatz) error {
	log := logging.FromContext(ctx)

	produkt, err := c.ProduktRepo.GetProdukt(ctx, produktID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn("Produkt not found for update", "produkt_id", produktID)
			return ErrProduktNotFound
		}
		log.Error("Failed to retrieve produkt for update", "error", err, "produkt_id", produktID)
		return ErrDatabase
	}

	err = produkt.UpdateDetails(name, kategorie, steuersatz)
	if err != nil {
		log.Warn("Invalid produkt data for update", "error", err, "produkt_id", produktID)
		return ErrInvalidProduktData
	}

	err = c.ProduktRepo.UpdateProdukt(ctx, produkt)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn("Produkt name already exists", "error", err, "name", produkt.Name)
			return ErrProduktAlreadyExists
		}
		log.Error("Failed to update produkt", "error", err, "produkt_id", produktID)
		return ErrDatabase
	}

	log.Info("Produkt updated", "produkt_id", produktID)
	return nil
}

// VerschiebeProdukt tauscht das Produkt mit seinem Nachbarn in derselben
// Kategorie. Am Rand der Kategorie bleibt die Reihenfolge unverändert.
func (c Command) VerschiebeProdukt(ctx context.Context, produktID int, richtung produkt.Richtung) error {
	log := logging.FromContext(ctx)

	err := c.ProduktRepo.VerschiebeProdukt(ctx, produktID, richtung == produkt.RichtungHoch)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn("Produkt not found for move", "produkt_id", produktID)
			return ErrProduktNotFound
		}
		log.Error("Failed to move produkt", "error", err, "produkt_id", produktID)
		return ErrDatabase
	}

	log.Info("Produkt moved", "produkt_id", produktID, "richtung", string(richtung))
	return nil
}

func (c Command) CreateVariante(ctx context.Context, produktID int, name string, preisCents int) (int, error) {
	log := logging.FromContext(ctx)

	_, err := c.ProduktRepo.GetProdukt(ctx, produktID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn("Produkt not found for variante creation", "produkt_id", produktID)
			return 0, ErrProduktNotFound
		}
		log.Error("Failed to retrieve produkt for variante creation", "error", err, "produkt_id", produktID)
		return 0, ErrDatabase
	}

	variante, err := produkt.NewVariante(name, preisCents)
	if err != nil {
		log.Warn("Invalid variante data", "error", err, "variante_name", name)
		return 0, ErrInvalidVarianteData
	}

	varianteID, err := c.ProduktRepo.CreateVariante(ctx, produktID, variante)
	if err != nil {
		log.Error("Failed to create variante", "error", err, "produkt_id", produktID, "name", variante.Name)
		return 0, ErrDatabase
	}

	log.Info("Variante created", "variante_id", varianteID, "produkt_id", produktID)
	return varianteID, nil
}

func (c Command) UpdateVariante(ctx context.Context, varianteID int, name string, preisCents int) error {
	log := logging.FromContext(ctx)

	variante, err := c.ProduktRepo.GetVariante(ctx, varianteID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn("Variante not found for update", "variante_id", varianteID)
			return ErrVarianteNotFound
		}
		log.Error("Failed to retrieve variante for update", "error", err, "variante_id", varianteID)
		return ErrDatabase
	}

	err = variante.UpdateDetails(name, preisCents)
	if err != nil {
		log.Warn("Invalid variante data for update", "error", err, "variante_id", varianteID)
		return ErrInvalidVarianteData
	}

	err = c.ProduktRepo.UpdateVariante(ctx, variante)
	if err != nil {
		log.Error("Failed to update variante", "error", err, "variante_id", varianteID)
		return ErrDatabase
	}

	log.Info("Variante updated", "variante_id", varianteID)
	return nil
}

// VerschiebeVariante tauscht die Variante mit ihrem Nachbarn im selben Produkt.
// Am Rand der Variantenliste bleibt die Reihenfolge unverändert.
func (c Command) VerschiebeVariante(ctx context.Context, varianteID int, richtung produkt.Richtung) error {
	log := logging.FromContext(ctx)

	err := c.ProduktRepo.VerschiebeVariante(ctx, varianteID, richtung == produkt.RichtungHoch)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn("Variante not found for move", "variante_id", varianteID)
			return ErrVarianteNotFound
		}
		log.Error("Failed to move variante", "error", err, "variante_id", varianteID)
		return ErrDatabase
	}

	log.Info("Variante moved", "variante_id", varianteID, "richtung", string(richtung))
	return nil
}

func (c Command) ActivateVariante(ctx context.Context, varianteID int) error {
	return c.applyVarianteStatusChange(ctx, varianteID, "Variante activated", func(v *produkt.Variante) { v.Activate() })
}

func (c Command) DeactivateVariante(ctx context.Context, varianteID int) error {
	return c.applyVarianteStatusChange(ctx, varianteID, "Variante deactivated", func(v *produkt.Variante) { v.Deactivate() })
}

func (c Command) applyVarianteStatusChange(ctx context.Context, varianteID int, successMsg string, action func(*produkt.Variante)) error {
	log := logging.FromContext(ctx)

	variante, err := c.ProduktRepo.GetVariante(ctx, varianteID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn("Variante not found for status change", "variante_id", varianteID)
			return ErrVarianteNotFound
		}
		log.Error("Failed to retrieve variante for status change", "error", err, "variante_id", varianteID)
		return ErrDatabase
	}

	action(&variante)

	if err := c.ProduktRepo.UpdateVariante(ctx, variante); err != nil {
		log.Error("Failed to update variante", "error", err, "variante_id", varianteID)
		return ErrDatabase
	}

	log.Info(successMsg, "variante_id", varianteID)
	return nil
}

func (c Command) DeleteProdukt(ctx context.Context, produktID int) error {
	log := logging.FromContext(ctx)

	produkt, err := c.ProduktRepo.GetProdukt(ctx, produktID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn("Produkt not found for deletion", "produkt_id", produktID)
			return ErrProduktNotFound
		}
		log.Error("Failed to retrieve produkt for deletion", "error", err, "produkt_id", produktID)
		return ErrDatabase
	}

	for i := range produkt.Varianten {
		produkt.Varianten[i].Delete()
	}
	produkt.Delete()

	err = c.ProduktRepo.DeleteProduktMitVarianten(ctx, produkt)
	if err != nil {
		log.Error("Failed to delete produkt", "error", err, "produkt_id", produktID)
		return ErrDatabase
	}

	log.Info("Produkt deleted", "produkt_id", produktID)
	return nil
}

func (c Command) DeleteVariante(ctx context.Context, produktID int, varianteID int) error {
	log := logging.FromContext(ctx)

	produkt, err := c.ProduktRepo.GetProdukt(ctx, produktID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn("Produkt not found for variante deletion", "produkt_id", produktID)
			return ErrProduktNotFound
		}
		log.Error("Failed to retrieve produkt for variante deletion", "error", err, "produkt_id", produktID)
		return ErrDatabase
	}

	variante, err := c.ProduktRepo.GetVariante(ctx, varianteID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn("Variante not found for deletion", "variante_id", varianteID)
			return ErrVarianteNotFound
		}
		log.Error("Failed to retrieve variante for deletion", "error", err, "variante_id", varianteID)
		return ErrDatabase
	}

	// Ohne den Abgleich löscht eine fremde varianteId die Variante eines anderen
	// Produkts.
	gehoertZumProdukt := false
	for i := range produkt.Varianten {
		if produkt.Varianten[i].ID == variante.ID {
			gehoertZumProdukt = true
			break
		}
	}
	if !gehoertZumProdukt {
		log.Warn("Variante does not belong to produkt", "produkt_id", produktID, "variante_id", varianteID)
		return ErrVarianteNotFound
	}

	variante.Delete()

	err = c.ProduktRepo.UpdateVariante(ctx, variante)
	if err != nil {
		log.Error("Failed to delete variante", "error", err, "variante_id", varianteID)
		return ErrDatabase
	}

	log.Info("Variante deleted", "variante_id", varianteID)
	return nil
}

// SortiereVariantenAlphabetisch ordnet die Varianten eines Produkts einmalig
// alphabetisch. Danach ist es eine gewöhnliche Reihenfolge, die sich mit den
// Pfeilen weiter anpassen lässt - bewusst kein dauerhafter Sortiermodus.
func (c Command) SortiereVariantenAlphabetisch(ctx context.Context, produktID int) error {
	log := logging.FromContext(ctx)

	if _, err := c.ProduktRepo.GetProdukt(ctx, produktID); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn("Produkt not found for sorting", "produkt_id", produktID)
			return ErrProduktNotFound
		}
		log.Error("Failed to retrieve produkt for sorting", "error", err, "produkt_id", produktID)
		return ErrDatabase
	}

	if err := c.ProduktRepo.SortiereVariantenAlphabetisch(ctx, produktID); err != nil {
		log.Error("Failed to sort varianten", "error", err, "produkt_id", produktID)
		return ErrDatabase
	}

	log.Info("Varianten sorted alphabetically", "produkt_id", produktID)
	return nil
}
