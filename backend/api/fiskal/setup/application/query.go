package application

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/internal/logging"
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
	Umgebungen             *UmgebungsCache
}

type TSEStatus struct {
	Umgebung        string
	IstKonfiguriert bool
}

// TSESetupBefund marks per TSS a client whose serial number equals the Kassen-Seriennummer.
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
	log := logging.FromContext(ctx)

	identitaet, err := q.TSERepo.GetKassenidentitaet(ctx)
	if err != nil {
		log.Error("Failed to retrieve kassenidentitaet", "error", err)
		return tse.Kassenidentitaet{}, ErrDatabase
	}
	return identitaet, nil
}

func (q Query) GetTSEKonfiguration(ctx context.Context) (tse.Konfiguration, error) {
	log := logging.FromContext(ctx)

	c, err := q.TSERepo.GetTSEKonfiguration(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return tse.Konfiguration{}, ErrNotFound
		}
		log.Error("Failed to retrieve tse_konfiguration", "error", err)
		return tse.Konfiguration{}, ErrDatabase
	}

	return c, nil
}

func (q Query) TestTSEVerbindung(ctx context.Context) (tse.VerbindungStatus, error) {
	log := logging.FromContext(ctx)

	conf, err := q.TSERepo.GetTSEKonfiguration(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return tse.VerbindungStatus{}, ErrTSENichtKonfiguriert
		}
		log.Error("Failed to retrieve tse_konfiguration for test", "error", err)
		return tse.VerbindungStatus{}, ErrDatabase
	}

	credentials := conf.Credentials()
	if err := credentials.Validate(); err != nil {
		return tse.VerbindungStatus{}, ErrTSENichtKonfiguriert
	}

	tester, err := q.NewTSEConnectionTester(credentials)
	if err != nil {
		log.Error("Failed to create TSE connection tester", "error", err)
		return tse.VerbindungStatus{}, ErrTSEVerbindungFehlgeschlagen
	}

	status, err := tester.TestConnection(ctx)
	if err != nil {
		log.Warn("TSE connection test failed", "error", err)
		return tse.VerbindungStatus{}, ErrTSEVerbindungFehlgeschlagen
	}

	identitaet, err := q.TSERepo.GetKassenidentitaet(ctx)
	if err != nil {
		log.Error("Failed to retrieve kassenidentitaet for connection test", "error", err)
		return tse.VerbindungStatus{}, ErrDatabase
	}
	status.SeriennummerKorrekt = status.ClientSerialNumber == identitaet.Seriennummer.String()

	return status, nil
}

// CheckTSESetup is side-effect free: it saves nothing and sends only read requests to fiskaly.
func (q Query) CheckTSESetup(ctx context.Context, credentials tse.SetupCredentials) (TSESetupBefund, error) {
	log := logging.FromContext(ctx)

	if err := credentials.Validate(); err != nil {
		return TSESetupBefund{}, ErrTSESetupZugangsdaten
	}

	client, umgebung, tssList, err := oeffneSetupClient(ctx, log, q.NewTSESetupClient, credentials, nil, "setup check")
	if err != nil {
		return TSESetupBefund{}, err
	}

	identitaet, err := q.TSERepo.GetKassenidentitaet(ctx)
	if err != nil {
		log.Error("Failed to retrieve kassenidentitaet for setup check", "error", err)
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
			log.Warn("Failed to list clients during setup check", "error", err, "tss_id", t.ID)
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

func passenderClient(clients []tse.ClientInfo, seriennummer string) *ClientBefund {
	for _, c := range clients {
		if c.SerialNumber == seriennummer {
			return &ClientBefund{ID: c.ID, SerialNumber: c.SerialNumber, State: c.State}
		}
	}
	return nil
}

func (q Query) GetTSEStatus(ctx context.Context) (TSEStatus, error) {
	log := logging.FromContext(ctx)

	conf, err := q.TSERepo.GetTSEKonfiguration(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return TSEStatus{}, nil
		}
		log.Error("Failed to retrieve tse_konfiguration for status", "error", err)
		return TSEStatus{}, ErrDatabase
	}

	status := TSEStatus{
		IstKonfiguriert: conf.IstKonfiguriert(),
	}
	if !status.IstKonfiguriert {
		return status, nil
	}
	credentials := conf.Credentials()
	if umgebung, ok := q.Umgebungen.lies(credentials); ok {
		status.Umgebung = string(umgebung)
		return status, nil
	}
	tester, err := q.NewTSEConnectionTester(credentials)
	if err != nil {
		log.Warn("Failed to create TSE connection tester for status", "error", err)
		return status, nil
	}

	// The status display needs only the environment from the auth token, not a full connection test.
	umgebung, err := tester.Umgebung(ctx)
	if err != nil {
		log.Warn("Failed to determine TSE environment for status", "error", err)
		return status, nil
	}

	q.Umgebungen.merke(credentials, umgebung)
	status.Umgebung = string(umgebung)
	return status, nil
}
