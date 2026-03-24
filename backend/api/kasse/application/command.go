package application

import (
	"context"
	"errors"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/rs/zerolog"
)

type kassenRepo interface {
	InsertKassensitzung(ctx context.Context, datum time.Time, bezeichnung string) (int, error)
	GetOffeneKassensitzung(ctx context.Context) (*kasse.KassensitzungState, error)
	WriteEvent(ctx context.Context, e event.Event, streamType kasse.StreamType, kassensitzungNr int) (int, error)
	GetMaxVersion(ctx context.Context, subject string) (int, error)
}

type Command struct {
	KassenRepo kassenRepo
}

// getOffeneKassensitzungOderFehler returns the open Kassensitzung or ErrKasseNichtGeoeffnet.
func (c Command) getOffeneKassensitzungOderFehler(ctx context.Context) (*kasse.KassensitzungState, error) {
	ks, err := c.KassenRepo.GetOffeneKassensitzung(ctx)
	if err != nil {
		return nil, ErrDatabase
	}
	if ks == nil {
		return nil, ErrKasseNichtGeoeffnet
	}
	return ks, nil
}

// writeKassensitzungEvent writes a Kassensitzung event with OCC.
func (c Command) writeKassensitzungEvent(ctx context.Context, e event.Event, kassensitzungNr int) error {
	log := zerolog.Ctx(ctx)

	subject := kasse.KassensitzungSubject(kassensitzungNr)
	maxVersion, err := c.KassenRepo.GetMaxVersion(ctx, subject)
	if err != nil {
		return ErrDatabase
	}

	e.Version = maxVersion + 1

	_, err = c.KassenRepo.WriteEvent(ctx, e, kasse.StreamTypeKassensitzung, kassensitzungNr)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn().Int("version", e.Version).Str("subject", subject).Msg("OCC Kassensitzung conflict")
			return ErrKonflikt
		}
		return ErrDatabase
	}

	return nil
}

// KassensitzungEroeffnen opens a new Kassensitzung. Returns ErrKasseAlreadyOpen if one is already open.
func (c Command) KassensitzungEroeffnen(ctx context.Context, userID int, userName string, datum time.Time, bezeichnung string) (int, error) {
	log := zerolog.Ctx(ctx)

	existing, err := c.KassenRepo.GetOffeneKassensitzung(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check for existing open Kassensitzung")
		return 0, ErrDatabase
	}
	if existing != nil {
		log.Warn().Int("z_nr", existing.ZNr).Msg("Kassensitzung already open")
		return 0, ErrKasseAlreadyOpen
	}

	zNr, err := c.KassenRepo.InsertKassensitzung(ctx, datum, bezeichnung)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert Kassensitzung")
		return 0, ErrDatabase
	}

	evt, err := kasse.NewKassensitzungEroeffnetEvent(kasse.KassensitzungSubject(zNr), userID, userName, datum.Format("2006-01-02"), bezeichnung)
	if err != nil {
		log.Error().Err(err).Int("z_nr", zNr).Msg("Failed to create kassensitzung-eroeffnet event")
		return 0, err
	}

	if err := c.writeKassensitzungEvent(ctx, evt, zNr); err != nil {
		return 0, err
	}

	log.Info().Int("z_nr", zNr).Msg("Kassensitzung eroeffnet")
	return zNr, nil
}

// AnfangsbestandSetzen sets the initial cash balance for the open Kassensitzung.
func (c Command) AnfangsbestandSetzen(ctx context.Context, userID int, userName string, betragCents int) error {
	log := zerolog.Ctx(ctx)

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return err
	}

	evt, err := kasse.NewAnfangsbestandGesetztEvent(kasse.KassensitzungSubject(ks.ZNr), userID, userName, betragCents)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to create anfangsbestand-gesetzt event")
		return err
	}

	if err := c.writeKassensitzungEvent(ctx, evt, ks.ZNr); err != nil {
		return err
	}

	log.Info().Int("z_nr", ks.ZNr).Int("betrag_cents", betragCents).Msg("Anfangsbestand gesetzt")
	return nil
}

// KassenbewegungBuchen books a cash movement (privateinlage, privatentnahme, geldtransit).
func (c Command) KassenbewegungBuchen(ctx context.Context, userID int, userName string, art string, betragCents int, kommentar string) error {
	log := zerolog.Ctx(ctx)

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return err
	}

	evt, err := kasse.NewKassenbewegungGebuchtEvent(kasse.KassensitzungSubject(ks.ZNr), userID, userName, art, betragCents, kommentar)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to create kassenbewegung-gebucht event")
		return err
	}

	if err := c.writeKassensitzungEvent(ctx, evt, ks.ZNr); err != nil {
		return err
	}

	log.Info().Int("z_nr", ks.ZNr).Str("art", art).Int("betrag_cents", betragCents).Msg("Kassenbewegung gebucht")
	return nil
}
