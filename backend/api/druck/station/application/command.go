package application

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/nicograef/jotti/backend/internal/logging"

	"github.com/nicograef/jotti/backend/api/druck/bondruck/application/escpos"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
)

type druckstationCommandRepo interface {
	UpsertDruckstation(ctx context.Context, station druckstation.Druckstation) error
	GetKonfigurierteDruckstationen(ctx context.Context) (map[string]druckstation.Druckstation, error)
}

type druckauftragCommandRepo interface {
	EnqueueDruckauftraege(ctx context.Context, auftraege []druckauftrag_repo.NeuerDruckauftrag) error
}

type Command struct {
	DruckstationRepo druckstationCommandRepo
	DruckauftragRepo druckauftragCommandRepo
}

func (c Command) UpsertDruckstation(ctx context.Context, kategorie, druckerIP, bonmodus string) error {
	log := logging.FromContext(ctx)

	station, err := druckstation.NewDruckstation(
		druckstation.Kategorie(kategorie),
		druckerIP,
		druckstation.Bonmodus(bonmodus),
	)
	if err != nil {
		log.Warn("Invalid druckstation", "error", err, "kategorie", kategorie)
		return ErrUngueltigeDruckstation
	}

	if err := c.DruckstationRepo.UpsertDruckstation(ctx, station); err != nil {
		log.Error("Failed to upsert druckstation", "error", err, "kategorie", kategorie)
		return ErrDatabase
	}

	log.Info("Druckstation updated", "kategorie", kategorie, "druckerIP", druckerIP)
	return nil
}

// TestbonDrucken reiht einen Testbon (Stationsname + Zeitstempel) in die Outbox
// ein. Es gibt keinen Status-Rückkanal: Schlägt der Druck fehl, erscheint der
// Auftrag wie jeder andere in den fehlgeschlagenen Druckaufträgen.
func (c Command) TestbonDrucken(ctx context.Context, kategorie string) error {
	log := logging.FromContext(ctx)

	kat := druckstation.Kategorie(kategorie)

	stationen, err := c.DruckstationRepo.GetKonfigurierteDruckstationen(ctx)
	if err != nil {
		log.Error("Failed to load druckstationen for testbon", "error", err, "kategorie", kategorie)
		return ErrDatabase
	}

	station, ok := stationen[kategorie]
	if !ok || station.DruckerIP == "" {
		log.Warn("Testbon requested for druckstation without configured IP", "kategorie", kategorie)
		return ErrDruckstationNichtKonfiguriert
	}

	payload := escpos.FormatTestbon(kat.Anzeigename(), time.Now())

	auftrag := druckauftrag_repo.NeuerDruckauftrag{
		ZielIP:   station.DruckerIP,
		Payload:  base64.StdEncoding.EncodeToString(payload),
		BonArt:   "testbon",
		Referenz: "testdruck:" + kategorie,
	}

	if err := c.DruckauftragRepo.EnqueueDruckauftraege(ctx, []druckauftrag_repo.NeuerDruckauftrag{auftrag}); err != nil {
		log.Error("Failed to enqueue testbon", "error", err, "kategorie", kategorie)
		return ErrDatabase
	}

	log.Info("Testbon queued", "kategorie", kategorie, "druckerIP", station.DruckerIP)
	return nil
}
