package application

import (
	"context"
	"errors"
	"time"

	tseApp "github.com/nicograef/jotti/backend/api/tse/application"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/reporting"
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/rs/zerolog"
)

type kassenjournalRepo interface {
	WriteEvent(ctx context.Context, e event.Event, streamType kasse.StreamType, kassensitzungNr int) (int, error)
	WriteEventWithNachsignierAuftrag(ctx context.Context, e event.Event, streamType kasse.StreamType, kassensitzungNr int, txID string, processType string, processData string) (int, error)
	GetMaxVersion(ctx context.Context, subject string) (int, error)
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

type reportingRepo interface {
	GetReporting(ctx context.Context, kassensitzungNr int) (reporting.ReportingData, error)
}

type Command struct {
	KassenjournalRepo   kassenjournalRepo
	KassensitzungenRepo kassensitzungenRepo
	SettingsRepo        settingsRepo
	ReportingRepo       reportingRepo
	TSESignierer        tseApp.Signierer
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

// writeKassensitzungEvent writes a Kassensitzung event with OCC against expectedVersion.
// expectedVersion ist die Version des Zustands, gegen den der Command validiert hat
// (frischer Stream: 0). Ein UNIQUE(subject, version)-Konflikt wird zu ErrKonflikt.
func (c Command) writeKassensitzungEvent(ctx context.Context, e event.Event, kassensitzungNr int, expectedVersion int) error {
	log := zerolog.Ctx(ctx)

	subject := kasse.KassensitzungSubject(kassensitzungNr)
	e.Version = expectedVersion + 1

	_, err := c.KassenjournalRepo.WriteEvent(ctx, e, kasse.StreamTypeKassensitzung, kassensitzungNr)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn().Int("version", e.Version).Str("subject", subject).Msg("OCC Kassensitzung conflict")
			return ErrKonflikt
		}
		return ErrDatabase
	}

	return nil
}

func (c Command) writeKassensitzungEventWithNachsignierAuftrag(ctx context.Context, e event.Event, kassensitzungNr int, expectedVersion int, txID string, processType string, processData string) error {
	log := zerolog.Ctx(ctx)

	subject := kasse.KassensitzungSubject(kassensitzungNr)
	e.Version = expectedVersion + 1

	_, err := c.KassenjournalRepo.WriteEventWithNachsignierAuftrag(ctx, e, kasse.StreamTypeKassensitzung, kassensitzungNr, txID, processType, processData)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn().Int("version", e.Version).Str("subject", subject).Msg("OCC Kassensitzung conflict")
			return ErrKonflikt
		}
		return ErrDatabase
	}

	return nil
}

// writeSignedKassensitzungEvent writes a signed Kassensitzung event, attaching the TSE retry job
// when the signing produced a Nachsignier-Auftrag and writing the event on its own otherwise.
func (c Command) writeSignedKassensitzungEvent(ctx context.Context, signierung tseApp.Signierung, kassensitzungNr int, expectedVersion int) error {
	if signierung.NachsignierAuftrag != nil {
		na := signierung.NachsignierAuftrag
		return c.writeKassensitzungEventWithNachsignierAuftrag(ctx, signierung.Event, kassensitzungNr, expectedVersion, na.TxID, na.ProcessType, na.ProcessData)
	}
	return c.writeKassensitzungEvent(ctx, signierung.Event, kassensitzungNr, expectedVersion)
}

// KassensitzungEroeffnen opens a new Kassensitzung. Returns ErrKasseAlreadyOpen if one is already open.
func (c Command) KassensitzungEroeffnen(ctx context.Context, userID int, userName string, bezeichnung string, betragCents int) (int, error) {
	log := zerolog.Ctx(ctx)

	betreiber, err := c.SettingsRepo.GetBetreiber(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn().Msg("Kassensitzung blocked: betreiber not configured")
			return 0, ErrBetreiberNichtKonfiguriert
		}
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

	// Frischer Stream (neue z_nr): erwartete Version 0, das Eröffnungs-Event ist version = 1.
	if err := c.writeKassensitzungEvent(ctx, evt, zNr, 0); err != nil {
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

	signierung, err := c.signGeldtransitGebuchtEvent(ctx, evt, richtung, betragCents)
	if err != nil {
		return err
	}

	// Geldtransit validiert keinen Stream-Zustand (reines Anhängen); die Version wird
	// erst unmittelbar vor dem Schreiben bestimmt.
	maxVersion, err := c.KassenjournalRepo.GetMaxVersion(ctx, kasse.KassensitzungSubject(ks.ZNr))
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to load max version for geldtransit")
		return ErrDatabase
	}

	if err := c.writeSignedKassensitzungEvent(ctx, signierung, ks.ZNr, maxVersion); err != nil {
		return err
	}

	msg := "Geldtransit gebucht"
	if signierung.NachsignierAuftrag != nil {
		msg += " (unsigniert, Nachsignierung vorgemerkt)"
	}
	log.Info().Int("z_nr", ks.ZNr).Str("richtung", richtung).Int("betrag_cents", betragCents).Msg(msg)
	return nil
}

// KasseAbschliessen schließt die Kasse in einem Schritt ab: Kassensturz,
// Differenzbuchung (bei Differenz ungleich Null) und Tagesabschluss.
//
// Feste Schreibreihenfolge:
//  1. kassensturz-durchgefuehrt:v1 (immer)
//  2. differenz-soll-ist-gebucht:v1 (nur bei Differenz ungleich Null, signiert)
//  3. tagesabschluss-erstellt:v1 (signiert, schließt die Kassensitzung)
//
// Invariante: Tisch-Saldo-Sperre — alle Tisch-Sessions müssen saldo_cents = 0
// haben. Die Tagessummen des Z-Bons kommen aus GetReporting.
//
// Teilfehler: Schlägt ein Schreibvorgang nach dem ersten Event fehl, bleibt die
// Kassensitzung offen und der Abschluss kann wiederholt werden. Es gibt bewusst
// keine umschließende Transaktion über alle Events.
func (c Command) KasseAbschliessen(ctx context.Context, userID int, userName string, istBestandCents int) error {
	log := zerolog.Ctx(ctx)

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return err
	}

	subject := kasse.KassensitzungSubject(ks.ZNr)

	// OCC-Anker VOR der Bestandsberechnung: Bucht ein paralleler Geldtransit in den
	// KS-Stream, nachdem der Soll-Bestand gelesen wurde, wäre die Differenz veraltet —
	// der Kassensturz-Write läuft dann in den Versionskonflikt und der Abschluss wird
	// mit frischem Bestand wiederholt.
	expectedVersion, err := c.KassenjournalRepo.GetMaxVersion(ctx, subject)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to load max version for Kassenabschluss")
		return ErrDatabase
	}

	sollBestandCents, err := c.KassenjournalRepo.GetKassenbestand(ctx, ks.ZNr)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to get Kassenbestand for Kassenabschluss")
		return ErrDatabase
	}
	differenzCents := sollBestandCents - istBestandCents

	// Invariant: Tisch-Saldo-Sperre — all tisch sessions must have saldo_cents = 0
	sessions, err := c.KassenjournalRepo.GetTischSessionsByKassensitzungNr(ctx, ks.ZNr)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to get tisch sessions for Kassenabschluss")
		return ErrDatabase
	}
	for _, s := range sessions {
		if s.SaldoCents != 0 {
			log.Warn().Int("z_nr", ks.ZNr).Int("tisch_id", s.TischID).Int("saldo_cents", s.SaldoCents).
				Msg("Kassenabschluss rejected: Tisch has non-zero saldo")
			return ErrTischeSaldoOffen
		}
	}

	// 1. Kassensturz
	kassensturzEvt, err := kasse.NewKassensturzDurchgefuehrtEvent(subject, userID, userName, sollBestandCents, istBestandCents, differenzCents)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to create kassensturz-durchgefuehrt event")
		return err
	}
	if err := c.writeKassensitzungEvent(ctx, kassensturzEvt, ks.ZNr, expectedVersion); err != nil {
		return err
	}
	expectedVersion++

	// 2. Differenzbuchung nur bei Differenz ungleich Null
	if differenzCents != 0 {
		diffEvt, err := kasse.NewDifferenzSollIstGebuchtEvent(subject, userID, userName, differenzCents)
		if err != nil {
			log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to create differenz-soll-ist-gebucht event")
			return err
		}
		signierung, err := c.signDifferenzSollIstGebuchtEvent(ctx, diffEvt, differenzCents)
		if err != nil {
			return err
		}
		if err := c.writeSignedKassensitzungEvent(ctx, signierung, ks.ZNr, expectedVersion); err != nil {
			return err
		}
		expectedVersion++
	}

	// 3. Tagesabschluss mit echten Tagessummen aus dem Reporting
	reportingData, err := c.ReportingRepo.GetReporting(ctx, ks.ZNr)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to get reporting for Tagesabschluss")
		return ErrDatabase
	}
	summary := reportingData.Summary

	now := time.Now().UTC()
	tagesabschlussEvt, err := kasse.NewTagesabschlussErstelltEvent(
		subject, userID, userName,
		ks.ZNr,
		ks.CreatedAt, now,
		summary.GesamtUmsatzCents, summary.GesamtStornierungenCents,
		summary.GeldtransitCents,
	)
	if err != nil {
		log.Error().Err(err).Int("z_nr", ks.ZNr).Msg("Failed to create tagesabschluss-erstellt event")
		return err
	}
	signierung, err := c.signTagesabschlussErstelltEvent(ctx, tagesabschlussEvt, ks.ZNr, ks.CreatedAt, now)
	if err != nil {
		return err
	}
	if err := c.writeSignedKassensitzungEvent(ctx, signierung, ks.ZNr, expectedVersion); err != nil {
		return err
	}

	msg := "Kasse abgeschlossen"
	if signierung.NachsignierAuftrag != nil {
		msg += " (Tagesabschluss unsigniert, Nachsignierung vorgemerkt)"
	}
	log.Info().Int("z_nr", ks.ZNr).
		Int("soll_cents", sollBestandCents).
		Int("ist_cents", istBestandCents).
		Int("differenz_cents", differenzCents).
		Msg(msg)
	return nil
}
