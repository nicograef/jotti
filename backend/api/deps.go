package api

import (
	"database/sql"

	direktverkaufApp "github.com/nicograef/jotti/backend/api/kasse/direktverkauf/application"
	pfandrueckgabeApp "github.com/nicograef/jotti/backend/api/kasse/pfandrueckgabe/application"
	tischgeschaeftApp "github.com/nicograef/jotti/backend/api/kasse/tischgeschaeft/application"
	reportingApp "github.com/nicograef/jotti/backend/api/reporting/application"
	produktApp "github.com/nicograef/jotti/backend/api/stammdaten/produkt/application"
	tischApp "github.com/nicograef/jotti/backend/api/stammdaten/tisch/application"

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

type Deps struct {
	// Version geht als KASSE_SW_VERSION in den DSFinV-K-Export; per ldflags gesetzt ("dev" in der Entwicklung).
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

	NewTSEConnectionTester func(tse.Credentials) (tse.ConnectionTester, error)
	NewTSESetupClient      func(tse.SetupCredentials) (tse.SetupClient, error)
}

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

// The services below are mounted in more than one area; each is wired once here, so no area can
// miss a dependency another area has.

func (d Deps) produktQuery() produktApp.Query {
	return produktApp.Query{ProduktRepo: d.ProduktRepo}
}

func (d Deps) tischCommand() tischApp.Command {
	return tischApp.Command{TischRepo: d.TischRepo, FavoritRepo: d.FavoritRepo}
}

func (d Deps) reportingQuery() reportingApp.Query {
	return reportingApp.Query{
		ReportingRepo:       d.ReportingRepo,
		KassensitzungenRepo: d.KassensitzungenRepo,
		TischSessionRepo:    d.KassenjournalRepo,
		TischRepo:           d.TischRepo,
	}
}

func (d Deps) tischgeschaeftCommand() tischgeschaeftApp.Command {
	return tischgeschaeftApp.Command{
		TischRepo:           d.TischRepo,
		EventRepo:           d.KassenjournalRepo,
		ProduktRepo:         d.ProduktRepo,
		KassensitzungenRepo: d.KassensitzungenRepo,
		DruckstationRepo:    d.DruckstationRepo,
	}
}

func (d Deps) direktverkaufCommand() direktverkaufApp.Command {
	return direktverkaufApp.Command{
		EventRepo:           d.KassenjournalRepo,
		ProduktRepo:         d.ProduktRepo,
		KassensitzungenRepo: d.KassensitzungenRepo,
		DruckstationRepo:    d.DruckstationRepo,
	}
}

func (d Deps) pfandrueckgabeCommand() pfandrueckgabeApp.Command {
	return pfandrueckgabeApp.Command{
		EventRepo:           d.KassenjournalRepo,
		ProduktRepo:         d.ProduktRepo,
		KassensitzungenRepo: d.KassensitzungenRepo,
		TischRepo:           d.TischRepo,
	}
}
