package application

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/rs/zerolog"

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
	log := zerolog.Ctx(ctx)

	station, err := druckstation.NewDruckstation(
		druckstation.Kategorie(kategorie),
		druckerIP,
		druckstation.Bonmodus(bonmodus),
	)
	if err != nil {
		log.Warn().Err(err).Str("kategorie", kategorie).Msg("Invalid druckstation")
		return ErrUngueltigeDruckstation
	}

	if err := c.DruckstationRepo.UpsertDruckstation(ctx, station); err != nil {
		log.Error().Err(err).Str("kategorie", kategorie).Msg("Failed to upsert druckstation")
		return ErrDatabase
	}

	log.Info().Str("kategorie", kategorie).Str("druckerIP", druckerIP).Msg("Druckstation updated")
	return nil
}

// TestbonDrucken reiht einen Testbon (Stationsname + Zeitstempel) für die
// angegebene Kategorie in die Outbox ein. Ist für die Kategorie kein Drucker
// konfiguriert (keine IP), wird ErrDruckstationNichtKonfiguriert zurückgegeben.
// Es gibt keinen eigenen Status-Rückkanal: schlägt der Druck fehl, erscheint der
// Auftrag wie jeder andere in den fehlgeschlagenen Druckaufträgen.
func (c Command) TestbonDrucken(ctx context.Context, kategorie string) error {
	log := zerolog.Ctx(ctx)

	kat := druckstation.Kategorie(kategorie)

	stationen, err := c.DruckstationRepo.GetKonfigurierteDruckstationen(ctx)
	if err != nil {
		log.Error().Err(err).Str("kategorie", kategorie).Msg("Failed to load druckstationen for testbon")
		return ErrDatabase
	}

	station, ok := stationen[kategorie]
	if !ok || station.DruckerIP == "" {
		log.Warn().Str("kategorie", kategorie).Msg("Testbon requested for druckstation without configured IP")
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
		log.Error().Err(err).Str("kategorie", kategorie).Msg("Failed to enqueue testbon")
		return ErrDatabase
	}

	log.Info().Str("kategorie", kategorie).Str("druckerIP", station.DruckerIP).Msg("Testbon queued")
	return nil
}
