package application

import (
	"context"
	"encoding/base64"
	"errors"
	"time"

	"github.com/nicograef/jotti/backend/api/druck/bondruck/application/escpos"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/internal/logging"
)

type zusatzdruckerCommandRepo interface {
	GetZusatzdrucker(ctx context.Context, id int) (druckstation.Zusatzdrucker, error)
	CreateZusatzdrucker(ctx context.Context, zd druckstation.Zusatzdrucker) (int, error)
	UpdateZusatzdrucker(ctx context.Context, zd druckstation.Zusatzdrucker) error
	DeleteZusatzdrucker(ctx context.Context, id int) error
	AssignProdukt(ctx context.Context, produktID, zusatzdruckerID int) error
	UnassignProdukt(ctx context.Context, produktID int) error
}

type druckauftragCommandRepo interface {
	EnqueueDruckauftraege(ctx context.Context, auftraege []druckstation.NeuerDruckauftrag) error
}

type Command struct {
	ZusatzdruckerRepo zusatzdruckerCommandRepo
	DruckauftragRepo  druckauftragCommandRepo
}

// Anlegen starts a Zusatzdrucker without IP and with Bonmodus pro_position.
func (c Command) Anlegen(ctx context.Context, name string) (int, error) {
	log := logging.FromContext(ctx)

	zd, err := druckstation.NewZusatzdrucker(name, "", druckstation.BonmodusProPosition)
	if err != nil {
		log.Warn("Invalid zusatzdrucker", "error", err)
		return 0, ErrUngueltigerZusatzdrucker
	}

	id, err := c.ZusatzdruckerRepo.CreateZusatzdrucker(ctx, zd)
	if err != nil {
		log.Error("Failed to create zusatzdrucker", "error", err)
		return 0, ErrDatabase
	}

	log.Info("Zusatzdrucker created", "id", id)
	return id, nil
}

func (c Command) Aendern(ctx context.Context, id int, name, druckerIP, bonmodus string) error {
	log := logging.FromContext(ctx)

	zd, err := druckstation.NewZusatzdrucker(name, druckerIP, druckstation.Bonmodus(bonmodus))
	if err != nil {
		log.Warn("Invalid zusatzdrucker", "error", err, "id", id)
		return ErrUngueltigerZusatzdrucker
	}
	zd.ID = id

	if err := c.ZusatzdruckerRepo.UpdateZusatzdrucker(ctx, zd); err != nil {
		return mapRepoError(ctx, err, "Failed to update zusatzdrucker", id)
	}

	log.Info("Zusatzdrucker updated", "id", id, "druckerIP", druckerIP)
	return nil
}

// Loeschen removes the Zusatzdrucker with its Produkt-Zuordnungen; those Produkte print at their Kategorie again.
func (c Command) Loeschen(ctx context.Context, id int) error {
	if err := c.ZusatzdruckerRepo.DeleteZusatzdrucker(ctx, id); err != nil {
		return mapRepoError(ctx, err, "Failed to delete zusatzdrucker", id)
	}

	logging.FromContext(ctx).Info("Zusatzdrucker deleted", "id", id)
	return nil
}

// TestbonDrucken queues a Testbon carrying the Zusatzdrucker's name. Like at a Druckstation there is no
// status channel: a failed print shows up among the failed Druckaufträge.
func (c Command) TestbonDrucken(ctx context.Context, id int) error {
	log := logging.FromContext(ctx)

	zd, err := c.ZusatzdruckerRepo.GetZusatzdrucker(ctx, id)
	if err != nil {
		return mapRepoError(ctx, err, "Failed to load zusatzdrucker for testbon", id)
	}

	if zd.DruckerIP == "" {
		log.Warn("Testbon requested for zusatzdrucker without IP", "id", id)
		return ErrZusatzdruckerNichtKonfiguriert
	}

	payload := escpos.FormatTestbon(zd.Name, time.Now())

	auftrag := druckstation.NeuerDruckauftrag{
		ZielIP:   zd.DruckerIP,
		Payload:  base64.StdEncoding.EncodeToString(payload),
		BonArt:   druckstation.BonArtTestbon,
		Referenz: druckstation.ZusatzdruckerTestdruckReferenz(zd.Name),
	}

	if err := c.DruckauftragRepo.EnqueueDruckauftraege(ctx, []druckstation.NeuerDruckauftrag{auftrag}); err != nil {
		log.Error("Failed to enqueue testbon", "error", err, "id", id)
		return ErrDatabase
	}

	log.Info("Testbon queued", "id", id, "druckerIP", zd.DruckerIP)
	return nil
}

// ProduktZuordnen moves an existing Zuordnung of the Produkt to this Zusatzdrucker.
func (c Command) ProduktZuordnen(ctx context.Context, produktID, zusatzdruckerID int) error {
	log := logging.FromContext(ctx)

	if err := c.ZusatzdruckerRepo.AssignProdukt(ctx, produktID, zusatzdruckerID); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn("Produkt or zusatzdrucker not found", "produktID", produktID, "zusatzdruckerID", zusatzdruckerID)
			return ErrZuordnungNichtMoeglich
		}
		log.Error("Failed to assign produkt", "error", err, "produktID", produktID, "zusatzdruckerID", zusatzdruckerID)
		return ErrDatabase
	}

	log.Info("Produkt assigned to zusatzdrucker", "produktID", produktID, "zusatzdruckerID", zusatzdruckerID)
	return nil
}

// ZuordnungEntfernen succeeds without an existing Zuordnung; the Produkt prints at its Kategorie again.
func (c Command) ZuordnungEntfernen(ctx context.Context, produktID int) error {
	log := logging.FromContext(ctx)

	if err := c.ZusatzdruckerRepo.UnassignProdukt(ctx, produktID); err != nil {
		log.Error("Failed to unassign produkt", "error", err, "produktID", produktID)
		return ErrDatabase
	}

	log.Info("Produkt unassigned from zusatzdrucker", "produktID", produktID)
	return nil
}

func mapRepoError(ctx context.Context, err error, msg string, id int) error {
	if errors.Is(err, db.ErrNotFound) {
		logging.FromContext(ctx).Warn("Zusatzdrucker not found", "id", id)
		return ErrZusatzdruckerNichtGefunden
	}
	logging.FromContext(ctx).Error(msg, "error", err, "id", id)
	return ErrDatabase
}
