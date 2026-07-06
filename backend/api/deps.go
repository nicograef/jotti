package api

import (
	"database/sql"

	"github.com/nicograef/jotti/backend/config"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/betreiber_repo"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
	"github.com/nicograef/jotti/backend/repository/druckstation_repo"
	"github.com/nicograef/jotti/backend/repository/favorit_repo"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/produkt_repo"
	"github.com/nicograef/jotti/backend/repository/reporting_repo"
	"github.com/nicograef/jotti/backend/repository/tisch_repo"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
	"github.com/nicograef/jotti/backend/repository/user_repo"
)

// Deps bundles all repositories and shared builders for the API wiring layer.
// It is constructed exactly once in app.SetupRoutes and passed to each area constructor.
type Deps struct {
	// Version ist die Build-Version der jotti-Software (gesetzt per ldflags,
	// "dev" im Entwicklungsmodus). Sie wird als KASSE_SW_VERSION in den
	// DSFinV-K-Export geschrieben.
	Version             string
	UserRepo            user_repo.Repository
	ProduktRepo         produkt_repo.Repository
	TischRepo           tisch_repo.Repository
	KassenjournalRepo   kassenjournal_repo.Repository
	KassensitzungenRepo kassensitzungen_repo.Repository
	BetreiberRepo       betreiber_repo.Repository
	TSERepo             tse_repo.Repository
	FavoritRepo         favorit_repo.Repository
	ReportingRepo       reporting_repo.Repository
	DruckstationRepo    druckstation_repo.Repository
	DruckauftragRepo    druckauftrag_repo.Repository

	// Fiskaly client factories — each closes over cfg.FiskalyBaseURL
	NewTSEConnectionTester func(tse.Credentials) (tse.ConnectionTester, error)
	NewTSESetupClient      func(tse.SetupCredentials) (tse.SetupClient, error)
}

// NewDeps constructs all repositories and shared builders exactly once.
func NewDeps(cfg config.Config, db *sql.DB) Deps {
	return Deps{
		UserRepo:            user_repo.NewRepository(db),
		ProduktRepo:         produkt_repo.NewRepository(db),
		TischRepo:           tisch_repo.NewRepository(db),
		KassenjournalRepo:   kassenjournal_repo.NewRepository(db),
		KassensitzungenRepo: kassensitzungen_repo.NewRepository(db),
		BetreiberRepo:       betreiber_repo.NewRepository(db),
		TSERepo:             tse_repo.NewRepository(db),
		FavoritRepo:         favorit_repo.NewRepository(db),
		ReportingRepo:       reporting_repo.NewRepository(db),
		DruckstationRepo:    druckstation_repo.NewRepository(db),
		DruckauftragRepo:    druckauftrag_repo.NewRepository(db),
		NewTSEConnectionTester: func(credentials tse.Credentials) (tse.ConnectionTester, error) {
			return tse_repo.NewFiskalyTSEClient(cfg.FiskalyBaseURL, credentials, nil)
		},
		NewTSESetupClient: func(credentials tse.SetupCredentials) (tse.SetupClient, error) {
			return tse_repo.NewFiskalyTSESetupClient(cfg.FiskalyBaseURL, credentials, nil)
		},
	}
}
