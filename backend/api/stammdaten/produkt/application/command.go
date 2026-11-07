package application

import (
	"context"
	"slices"

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
	DeleteProduktMitVarianten(ctx context.Context, produkt produkt.Produkt) error
	VerschiebeProdukt(ctx context.Context, produktID int, hoch bool) error
	VerschiebeVariante(ctx context.Context, varianteID int, hoch bool) error
	SortiereVariantenAlphabetisch(ctx context.Context, produktID int) error
}

type Command struct {
	ProduktRepo produktRepo
}

func (c Command) CreateProdukt(ctx context.Context, name string, kategorie produkt.Kategorie, steuersatz steuer.Steuersatz, pfand bool) (int, error) {
	log := logging.FromContext(ctx)

	produkt, err := produkt.NewProdukt(name, kategorie, steuersatz, pfand)
	if err != nil {
		log.Warn("Invalid produkt data", "error", err, "produkt_name", name)
		return 0, ErrInvalidProduktData
	}

	produktID, err := c.ProduktRepo.CreateProdukt(ctx, produkt)
	if err != nil {
		return 0, fromProduktError(err, log, "creation", "name", produkt.Name)
	}

	log.Info("Produkt created", "produkt_id", produktID)
	return produktID, nil
}

func (c Command) UpdateProdukt(ctx context.Context, produktID int, name string, kategorie produkt.Kategorie, steuersatz steuer.Steuersatz, pfand bool) error {
	log := logging.FromContext(ctx)

	produkt, err := c.ProduktRepo.GetProdukt(ctx, produktID)
	if err != nil {
		return fromProduktError(err, log, "update", "produkt_id", produktID)
	}

	err = produkt.UpdateDetails(name, kategorie, steuersatz, pfand)
	if err != nil {
		log.Warn("Invalid produkt data for update", "error", err, "produkt_id", produktID)
		return ErrInvalidProduktData
	}

	err = c.ProduktRepo.UpdateProdukt(ctx, produkt)
	if err != nil {
		return fromProduktError(err, log, "update", "produkt_id", produktID, "name", produkt.Name)
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
		return fromProduktError(err, log, "move", "produkt_id", produktID)
	}

	log.Info("Produkt moved", "produkt_id", produktID, "richtung", string(richtung))
	return nil
}

func (c Command) CreateVariante(ctx context.Context, produktID int, name string, preisCents int) (int, error) {
	log := logging.FromContext(ctx)

	_, err := c.ProduktRepo.GetProdukt(ctx, produktID)
	if err != nil {
		return 0, fromProduktError(err, log, "variante creation", "produkt_id", produktID)
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
		return fromVarianteError(err, log, "update", "variante_id", varianteID)
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
		return fromVarianteError(err, log, "move", "variante_id", varianteID)
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
		return fromVarianteError(err, log, "status change", "variante_id", varianteID)
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
		return fromProduktError(err, log, "deletion", "produkt_id", produktID)
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

	eigenesProdukt, err := c.ProduktRepo.GetProdukt(ctx, produktID)
	if err != nil {
		return fromProduktError(err, log, "variante deletion", "produkt_id", produktID)
	}

	// Only a Variante of this Produkt is deleted; a foreign varianteId would delete another Produkt's.
	idx := slices.IndexFunc(eigenesProdukt.Varianten, func(v produkt.Variante) bool { return v.ID == varianteID })
	if idx < 0 {
		log.Warn("Variante does not belong to produkt", "produkt_id", produktID, "variante_id", varianteID)
		return ErrVarianteNotFound
	}
	variante := eigenesProdukt.Varianten[idx]
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
		return fromProduktError(err, log, "sorting", "produkt_id", produktID)
	}

	if err := c.ProduktRepo.SortiereVariantenAlphabetisch(ctx, produktID); err != nil {
		log.Error("Failed to sort varianten", "error", err, "produkt_id", produktID)
		return ErrDatabase
	}

	log.Info("Varianten sorted alphabetically", "produkt_id", produktID)
	return nil
}
