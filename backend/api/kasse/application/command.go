package application

import (
	"context"
	"errors"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/rs/zerolog"
)

type kassenjournalRepo interface {
	WriteEvent(ctx context.Context, e event.Event, streamType kasse.StreamType, kassensitzungNr int) (int, error)
	GetMaxVersion(ctx context.Context, subject string) (int, error)
	ReadEventsBySubject(ctx context.Context, subject string) ([]event.Event, error)
	GetKassenbestand(ctx context.Context, kassensitzungNr int) (int, error)
	GetTischSessionsByKassensitzungNr(ctx context.Context, kassensitzungNr int) ([]kasse.TischSession, error)
}

type kassensitzungenRepo interface {
	InsertKassensitzung(ctx context.Context, datum time.Time, bezeichnung string) (int, error)
	GetOffeneKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
}

type settingsRepo interface {
	GetBetreiber(ctx context.Context) (settings.Betreiber, error)
}

type Command struct {
	KassenjournalRepo   kassenjournalRepo
	KassensitzungenRepo kassensitzungenRepo
	SettingsRepo        settingsRepo
}

// getOffeneKassensitzungOderFehler returns the open Kassensitzung or ErrKasseNichtGeoeffnet.
func (c Command) getOffeneKassensitzungOderFehler(ctx context.Context) (*kasse.Kassensitzung, error) {
	ks, err := c.KassensitzungenRepo.GetOffeneKassensitzung(ctx)
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
	maxVersion, err := c.KassenjournalRepo.GetMaxVersion(ctx, subject)
	if err != nil {
		return ErrDatabase
	}

	e.Version = maxVersion + 1

	_, err = c.KassenjournalRepo.WriteEvent(ctx, e, kasse.StreamTypeKassensitzung, kassensitzungNr)
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
func (c Command) KassensitzungEroeffnen(ctx context.Context, userID int, userName string, bezeichnung string, betragCents int) (int, error) {
	log := zerolog.Ctx(ctx)

	betreiber, err := c.SettingsRepo.GetBetreiber(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check betreiber configuration")
		return 0, ErrDatabase
	}
	if err = betreiber.Validate(); err != nil {
		log.Warn().Err(err).Msg("Kassensitzung blocked: betreiber not configured")
		return 0, ErrBetreiberNichtKonfiguriert
	}

	existing, err := c.KassensitzungenRepo.GetOffeneKassensitzung(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check for existing open Kassensitzung")
		return 0, ErrDatabase
	}
	if existing != nil {
		log.Warn().Int("z_nr", existing.ZNr).Msg("Kassensitzung already open")
		return 0, ErrKasseAlreadyOpen
	}

	berliner, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		log.Error().Err(err).Msg("Failed to load Europe/Berlin timezone")
		return 0, ErrDatabase
	}
	datum := time.Now().In(berliner).Truncate(24 * time.Hour)

	zNr, err := c.KassensitzungenRepo.InsertKassensitzung(ctx, datum, bezeichnung)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert Kassensitzung")
		return 0, ErrDatabase
	}

	evt, err := kasse.NewKassensitzungEroeffnetEvent(kasse.KassensitzungSubject(zNr), userID, userName, datum.Format("2006-01-02"), bezeichnung, betragCents)
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

// GeldtransitBuchen books a Geldtransit (einlage or entnahme).
func (c Command) GeldtransitBuchen(ctx context.Context, userID int, userName string, richtung string, betragCents int, kommentar string) error {
	log := zerolog.Ctx(ctx)

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return err
	}

	evt, err := kasse.NewGeldtransitGebuchtEvent(kasse.KassensitzungSubject(ks.ZNr), userID, userName, richtung, betragCents, kommentar)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to create geldtransit-gebucht event")
		return err
	}

	if err := c.writeKassensitzungEvent(ctx, evt, ks.ZNr); err != nil {
		return err
	}

	log.Info().Int("z_nr", ks.ZNr).Str("richtung", richtung).Int("betrag_cents", betragCents).Msg("Geldtransit gebucht")
	return nil
}

// KassensturzDurchfuehren performs a Soll/Ist comparison of the cash balance.
// Two-Event Pattern: writes kassensturz-durchgefuehrt:v1 always,
// and differenz-soll-ist-gebucht:v1 additionally when differenzCents != 0.
func (c Command) KassensturzDurchfuehren(ctx context.Context, userID int, userName string, istBestandCents int) error {
	log := zerolog.Ctx(ctx)

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return err
	}

	sollBestandCents, err := c.KassenjournalRepo.GetKassenbestand(ctx, ks.ZNr)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to get Kassenbestand for Kassensturz")
		return ErrDatabase
	}

	differenzCents := sollBestandCents - istBestandCents

	subject := kasse.KassensitzungSubject(ks.ZNr)

	evt, err := kasse.NewKassensturzDurchgefuehrtEvent(subject, userID, userName, sollBestandCents, istBestandCents, differenzCents)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to create kassensturz-durchgefuehrt event")
		return err
	}

	if err := c.writeKassensitzungEvent(ctx, evt, ks.ZNr); err != nil {
		return err
	}

	// Two-Event Pattern: write differenz-soll-ist-gebucht if differenz != 0
	if differenzCents != 0 {
		diffEvt, err := kasse.NewDifferenzSollIstGebuchtEvent(subject, userID, userName, differenzCents)
		if err != nil {
			log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to create differenz-soll-ist-gebucht event")
			return err
		}

		if err := c.writeKassensitzungEvent(ctx, diffEvt, ks.ZNr); err != nil {
			return err
		}
	}

	log.Info().Int("z_nr", ks.ZNr).
		Int("soll_cents", sollBestandCents).
		Int("ist_cents", istBestandCents).
		Int("differenz_cents", differenzCents).
		Msg("Kassensturz durchgefuehrt")
	return nil
}

// TagesabschlussErstellen creates the Z-Receipt and closes the Kassensitzung.
// Invariants:
//   - Kassensturz must be completed (kassensturz-durchgefuehrt:v1 in event stream)
//   - Tisch-Saldo-Sperre: all tisch sessions must have saldo_cents = 0
func (c Command) TagesabschlussErstellen(ctx context.Context, userID int, userName string) error {
	log := zerolog.Ctx(ctx)

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return err
	}

	subject := kasse.KassensitzungSubject(ks.ZNr)

	// Invariant: Kassensturz must be completed
	events, err := c.KassenjournalRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to read KS events for Tagesabschluss")
		return ErrDatabase
	}

	kassensturzDone := false
	for _, evt := range events {
		if evt.Type == string(kasse.EventTypeKassensturzDurchgefuehrtV1) {
			kassensturzDone = true
			break
		}
	}
	if !kassensturzDone {
		log.Warn().Int("z_nr", ks.ZNr).Msg("Tagesabschluss rejected: Kassensturz not completed")
		return ErrKassensturzErforderlich
	}

	// Invariant: Tisch-Saldo-Sperre — all tisch sessions must have saldo_cents = 0
	sessions, err := c.KassenjournalRepo.GetTischSessionsByKassensitzungNr(ctx, ks.ZNr)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to get tisch sessions for Tagesabschluss")
		return ErrDatabase
	}

	for _, s := range sessions {
		if s.SaldoCents != 0 {
			log.Warn().Int("z_nr", ks.ZNr).Int("tisch_id", s.TischID).Int("saldo_cents", s.SaldoCents).
				Msg("Tagesabschluss rejected: Tisch has non-zero saldo")
			return ErrTischeSaldoOffen
		}
	}

	now := time.Now().UTC()
	// Aggregate values are 0 here — actual reporting uses SQL aggregation queries.
	// The event records the structural fact of the Tagesabschluss; detailed numbers come from GetReporting.
	tagesabschlussEvt, err := kasse.NewTagesabschlussErstelltEvent(
		subject, userID, userName,
		ks.ZNr,
		ks.CreatedAt, now,
		0, 0, 0, 0,
	)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to create tagesabschluss-erstellt event")
		return err
	}

	if err := c.writeKassensitzungEvent(ctx, tagesabschlussEvt, ks.ZNr); err != nil {
		return err
	}

	log.Info().Int("z_nr", ks.ZNr).Msg("Tagesabschluss erstellt")
	return nil
}
