package application

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/rs/zerolog"
)

type settingsQueryRepo interface {
	GetKassenidentitaet(ctx context.Context) (settings.Kassenidentitaet, error)
	GetBetreiber(ctx context.Context) (settings.Betreiber, error)
	GetBondruckEinstellungen(ctx context.Context) (settings.BondruckEinstellungen, error)
	GetTSEKonfiguration(ctx context.Context) (settings.TSEKonfiguration, error)
}

type tseStatusRepo interface {
	CountOffeneTSENachsignierAuftraege(ctx context.Context) (int, error)
}

type NewTSEConnectionTester func(credentials tse.Credentials) (tse.ConnectionTester, error)

type Query struct {
	SettingsRepo           settingsQueryRepo
	TSEStatusRepo          tseStatusRepo
	NewTSEConnectionTester NewTSEConnectionTester
}

type TSEStatus struct {
	Umgebung               string
	OffeneNachsignierungen int
	IstKonfiguriert        bool
}

func (q Query) GetKassenidentitaet(ctx context.Context) (settings.Kassenidentitaet, error) {
	log := zerolog.Ctx(ctx)

	identitaet, err := q.SettingsRepo.GetKassenidentitaet(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve kassenidentitaet")
		return settings.Kassenidentitaet{}, ErrDatabase
	}
	return identitaet, nil
}

func (q Query) GetBetreiber(ctx context.Context) (settings.Betreiber, error) {
	log := zerolog.Ctx(ctx)

	b, err := q.SettingsRepo.GetBetreiber(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return settings.Betreiber{}, ErrNotFound
		}
		log.Error().Err(err).Msg("Failed to retrieve betreiber")
		return settings.Betreiber{}, ErrDatabase
	}
	return b, nil
}

func (q Query) GetBondruckEinstellungen(ctx context.Context) (settings.BondruckEinstellungen, error) {
	log := zerolog.Ctx(ctx)

	b, err := q.SettingsRepo.GetBondruckEinstellungen(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return settings.BondruckEinstellungen{}, ErrNotFound
		}
		log.Error().Err(err).Msg("Failed to retrieve bondruck_einstellungen")
		return settings.BondruckEinstellungen{}, ErrDatabase
	}
	return b, nil
}

func (q Query) GetTSEKonfiguration(ctx context.Context) (settings.TSEKonfiguration, error) {
	log := zerolog.Ctx(ctx)

	c, err := q.SettingsRepo.GetTSEKonfiguration(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return settings.TSEKonfiguration{}, ErrNotFound
		}
		log.Error().Err(err).Msg("Failed to retrieve tse_konfiguration")
		return settings.TSEKonfiguration{}, ErrDatabase
	}

	return c, nil
}

func (q Query) TestTSEVerbindung(ctx context.Context) (tse.VerbindungStatus, error) {
	log := zerolog.Ctx(ctx)

	if q.NewTSEConnectionTester == nil {
		log.Error().Msg("Missing TSE connection tester factory")
		return tse.VerbindungStatus{}, ErrDatabase
	}

	conf, err := q.SettingsRepo.GetTSEKonfiguration(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return tse.VerbindungStatus{}, ErrTSENichtKonfiguriert
		}
		log.Error().Err(err).Msg("Failed to retrieve tse_konfiguration for test")
		return tse.VerbindungStatus{}, ErrDatabase
	}

	credentials := tse.Credentials{
		ApiKey:    conf.ApiKey,
		ApiSecret: conf.ApiSecret,
		TssID:     conf.TssID,
		ClientID:  conf.ClientID,
	}
	if err := credentials.Validate(); err != nil {
		return tse.VerbindungStatus{}, ErrTSENichtKonfiguriert
	}

	tester, err := q.NewTSEConnectionTester(credentials)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create TSE connection tester")
		return tse.VerbindungStatus{}, ErrTSEVerbindungFehlgeschlagen
	}

	status, err := tester.TestConnection(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("TSE connection test failed")
		return tse.VerbindungStatus{}, ErrTSEVerbindungFehlgeschlagen
	}

	return status, nil
}

func (q Query) GetTSEStatus(ctx context.Context) (TSEStatus, error) {
	log := zerolog.Ctx(ctx)

	if q.TSEStatusRepo == nil {
		log.Error().Msg("Missing TSE status repository")
		return TSEStatus{}, ErrDatabase
	}

	offeneNachsignierungen, err := q.TSEStatusRepo.CountOffeneTSENachsignierAuftraege(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count open TSE retry jobs")
		return TSEStatus{}, ErrDatabase
	}

	conf, err := q.SettingsRepo.GetTSEKonfiguration(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return TSEStatus{OffeneNachsignierungen: offeneNachsignierungen}, nil
		}
		log.Error().Err(err).Msg("Failed to retrieve tse_konfiguration for status")
		return TSEStatus{}, ErrDatabase
	}

	status := TSEStatus{
		OffeneNachsignierungen: offeneNachsignierungen,
		IstKonfiguriert:        conf.IstKonfiguriert(),
	}
	if !status.IstKonfiguriert {
		return status, nil
	}
	if q.NewTSEConnectionTester == nil {
		return status, nil
	}

	tester, err := q.NewTSEConnectionTester(tse.Credentials{
		ApiKey:    conf.ApiKey,
		ApiSecret: conf.ApiSecret,
		TssID:     conf.TssID,
		ClientID:  conf.ClientID,
	})
	if err != nil {
		log.Warn().Err(err).Msg("Failed to create TSE connection tester for status")
		return status, nil
	}

	verbindung, err := tester.TestConnection(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to determine TSE environment for status")
		return status, nil
	}

	status.Umgebung = string(verbindung.Umgebung)
	return status, nil
}
