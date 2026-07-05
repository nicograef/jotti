package application

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/rs/zerolog"
)

type tseQueryRepo interface {
	GetKassenidentitaet(ctx context.Context) (tse.Kassenidentitaet, error)
	GetTSEKonfiguration(ctx context.Context) (tse.Konfiguration, error)
}

type NewTSEConnectionTester func(credentials tse.Credentials) (tse.ConnectionTester, error)

type NewTSESetupClient func(credentials tse.SetupCredentials) (tse.SetupClient, error)

type Query struct {
	TSERepo                tseQueryRepo
	NewTSEConnectionTester NewTSEConnectionTester
	NewTSESetupClient      NewTSESetupClient
}

type TSEStatus struct {
	Umgebung        string
	IstKonfiguriert bool
}

// TSESetupBefund ist das seiteneffektfreie Ergebnis des Prüf-Schritts: die
// erkannte Umgebung und die vorhandenen TSS samt Zustand. Je TSS wird ein
// bereits passender Client (Seriennummer = Kassen-Seriennummer) ausgewiesen.
type TSESetupBefund struct {
	Umgebung      string
	VorhandeneTSS []TSSBefund
}

type TSSBefund struct {
	ID              string
	State           string
	PassenderClient *ClientBefund
}

type ClientBefund struct {
	ID           string
	SerialNumber string
	State        string
}

func (q Query) GetKassenidentitaet(ctx context.Context) (tse.Kassenidentitaet, error) {
	log := zerolog.Ctx(ctx)

	identitaet, err := q.TSERepo.GetKassenidentitaet(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve kassenidentitaet")
		return tse.Kassenidentitaet{}, ErrDatabase
	}
	return identitaet, nil
}

func (q Query) GetTSEKonfiguration(ctx context.Context) (tse.Konfiguration, error) {
	log := zerolog.Ctx(ctx)

	c, err := q.TSERepo.GetTSEKonfiguration(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return tse.Konfiguration{}, ErrNotFound
		}
		log.Error().Err(err).Msg("Failed to retrieve tse_konfiguration")
		return tse.Konfiguration{}, ErrDatabase
	}

	return c, nil
}

func (q Query) TestTSEVerbindung(ctx context.Context) (tse.VerbindungStatus, error) {
	log := zerolog.Ctx(ctx)

	if q.NewTSEConnectionTester == nil {
		log.Error().Msg("Missing TSE connection tester factory")
		return tse.VerbindungStatus{}, ErrDatabase
	}

	conf, err := q.TSERepo.GetTSEKonfiguration(ctx)
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

	identitaet, err := q.TSERepo.GetKassenidentitaet(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve kassenidentitaet for connection test")
		return tse.VerbindungStatus{}, ErrDatabase
	}
	status.SeriennummerKorrekt = status.ClientSerialNumber == identitaet.Seriennummer.String()

	return status, nil
}

// PruefeTSESetup führt den seiteneffektfreien Befund aus: Es authentifiziert
// sich mit den übergebenen Zugangsdaten, listet die vorhandenen TSS und prüft je
// TSS, ob bereits ein Client mit der Kassen-Seriennummer registriert ist. Es
// wird nichts gespeichert; nur Lese-Requests gehen an fiskaly.
func (q Query) PruefeTSESetup(ctx context.Context, credentials tse.SetupCredentials) (TSESetupBefund, error) {
	log := zerolog.Ctx(ctx)

	if q.NewTSESetupClient == nil {
		log.Error().Msg("Missing TSE setup client factory")
		return TSESetupBefund{}, ErrDatabase
	}
	if err := credentials.Validate(); err != nil {
		return TSESetupBefund{}, ErrTSESetupZugangsdaten
	}

	client, err := q.NewTSESetupClient(credentials)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create TSE setup client")
		return TSESetupBefund{}, ErrTSEVerbindungFehlgeschlagen
	}

	umgebung, tssList, err := client.ListTSS(ctx)
	if err != nil {
		if errors.Is(err, tse.ErrSetupAuthFehlgeschlagen) {
			return TSESetupBefund{}, ErrTSESetupZugangsdaten
		}
		log.Warn().Err(err).Msg("Failed to list TSS during setup check")
		return TSESetupBefund{}, ErrTSEVerbindungFehlgeschlagen
	}

	identitaet, err := q.TSERepo.GetKassenidentitaet(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve kassenidentitaet for setup check")
		return TSESetupBefund{}, ErrDatabase
	}
	seriennummer := identitaet.Seriennummer.String()

	befund := TSESetupBefund{
		Umgebung:      string(umgebung),
		VorhandeneTSS: make([]TSSBefund, 0, len(tssList)),
	}
	for _, t := range tssList {
		clients, err := client.ListClients(ctx, t.ID)
		if err != nil {
			log.Warn().Err(err).Str("tss_id", t.ID).Msg("Failed to list clients during setup check")
			return TSESetupBefund{}, ErrTSEVerbindungFehlgeschlagen
		}

		befund.VorhandeneTSS = append(befund.VorhandeneTSS, TSSBefund{
			ID:              t.ID,
			State:           t.State,
			PassenderClient: passenderClient(clients, seriennummer),
		})
	}

	return befund, nil
}

// passenderClient liefert den Client einer TSS, dessen serial_number der
// Kassen-Seriennummer entspricht — oder nil, wenn es keinen gibt.
func passenderClient(clients []tse.ClientInfo, seriennummer string) *ClientBefund {
	for _, c := range clients {
		if c.SerialNumber == seriennummer {
			return &ClientBefund{ID: c.ID, SerialNumber: c.SerialNumber, State: c.State}
		}
	}
	return nil
}

func (q Query) GetTSEStatus(ctx context.Context) (TSEStatus, error) {
	log := zerolog.Ctx(ctx)

	conf, err := q.TSERepo.GetTSEKonfiguration(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return TSEStatus{}, nil
		}
		log.Error().Err(err).Msg("Failed to retrieve tse_konfiguration for status")
		return TSEStatus{}, ErrDatabase
	}

	status := TSEStatus{
		IstKonfiguriert: conf.IstKonfiguriert(),
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

	// Fuer die Statusanzeige genuegt die Umgebung aus dem Auth-Token — kein
	// voller Verbindungstest (TSS-/Client-Abruf) noetig.
	umgebung, err := tester.Umgebung(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to determine TSE environment for status")
		return status, nil
	}

	status.Umgebung = string(umgebung)
	return status, nil
}
