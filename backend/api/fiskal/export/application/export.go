// Package application orchestrates the DSFinV-K export: it loads a Kassensitzung's
// events and master data and hands them to the pure dsfinvk mapper.
package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nicograef/jotti/backend/api/fiskal/dsfinvk"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/betreiber"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/internal/zeit"
	"github.com/rs/zerolog"
)

var (
	ErrDatabase                   = db.ErrDatabase
	ErrKassensitzungNichtGefunden = errors.New("kassensitzung nicht gefunden")
	ErrLeereKassensitzung         = dsfinvk.ErrKeineVorgaenge
)

type kassenjournalRepo interface {
	// ReadEventsByKassensitzung returns the events with each one's signature state
	// (LEFT JOIN on the Signaturaufträge: no entry = no signature required).
	ReadEventsByKassensitzung(ctx context.Context, kassensitzungNr int) ([]event.Event, map[int]tse.EventSignatur, error)
}

type kassensitzungenRepo interface {
	GetOffeneKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
	GetAllKassensitzungen(ctx context.Context) ([]kasse.Kassensitzung, error)
}

type betreiberRepo interface {
	GetBetreiber(ctx context.Context) (betreiber.Betreiber, error)
}

type tseRepo interface {
	GetKassenidentitaet(ctx context.Context) (tse.Kassenidentitaet, error)
	GetTSEStammdaten(ctx context.Context) (tse.Stammdaten, error)
}

type tischRepo interface {
	// GetAlleTischNamen must include deleted tables: the export names the
	// Abrechnungskreise of past sessions, and a table may be deleted after the Tagesabschluss.
	GetAlleTischNamen(ctx context.Context) (map[int]string, error)
}

type Export struct {
	KassenjournalRepo   kassenjournalRepo
	KassensitzungenRepo kassensitzungenRepo
	BetreiberRepo       betreiberRepo
	TSERepo             tseRepo
	TischRepo           tischRepo
	// Version is the jotti build version set via ldflags, exported as KASSE_SW_VERSION
	// in cashregister.csv.
	Version string
}

type Archiv struct {
	Dateiname string
	Inhalt    []byte
}

// Erstellen builds the DSFinV-K archive for the chosen Kassensitzung. nr == 0 selects
// the default session: the open one, else the latest closed.
func (e Export) Erstellen(ctx context.Context, nr int) (Archiv, error) {
	log := zerolog.Ctx(ctx)

	ks, err := e.resolveKassensitzung(ctx, nr)
	if err != nil {
		return Archiv{}, err
	}

	events, signaturen, err := e.KassenjournalRepo.ReadEventsByKassensitzung(ctx, ks.ZNr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read events for dsfinvk export")
		return Archiv{}, ErrDatabase
	}

	erstellung := dsfinvk.Erstellungszeitpunkt(events, time.Now().UTC())

	snapshot, err := e.snapshot(ctx, ks, erstellung)
	if err != nil {
		return Archiv{}, err
	}

	if dsfinvk.ZertifikatZuLang(snapshot.TSEStammdaten.Zertifikat) {
		log.Warn().
			Int("kassensitzung", ks.ZNr).
			Int("laenge", len(snapshot.TSEStammdaten.Zertifikat)).
			Msg("TSE certificate exceeds two DSFinV-K fields; TSE_ZERTIFIKAT left empty in export")
	}

	inhalt, err := dsfinvk.BuildArchive(snapshot, events, signaturen)
	if err != nil {
		if errors.Is(err, dsfinvk.ErrKeineVorgaenge) {
			return Archiv{}, ErrLeereKassensitzung
		}
		log.Error().Err(err).Msg("Failed to build dsfinvk archive")
		return Archiv{}, err
	}

	log.Info().Int("kassensitzung", ks.ZNr).Msg("Created DSFinV-K export")
	return Archiv{
		Dateiname: dateiname(snapshot.KasseSeriennummer, ks.ZNr, snapshot.Erstellung),
		Inhalt:    inhalt,
	}, nil
}

func (e Export) resolveKassensitzung(ctx context.Context, nr int) (kasse.Kassensitzung, error) {
	log := zerolog.Ctx(ctx)

	if nr > 0 {
		alle, err := e.KassensitzungenRepo.GetAllKassensitzungen(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Failed to list kassensitzungen")
			return kasse.Kassensitzung{}, ErrDatabase
		}
		for _, ks := range alle {
			if ks.ZNr == nr {
				return ks, nil
			}
		}
		return kasse.Kassensitzung{}, ErrKassensitzungNichtGefunden
	}

	// The export wants exactly the open session; a session in barrier status reaches
	// the same export via the latest-session branch below.
	offen, err := e.KassensitzungenRepo.GetOffeneKassensitzung(ctx) //nolint:forbidigo
	if err != nil {
		log.Error().Err(err).Msg("Failed to get offene kassensitzung")
		return kasse.Kassensitzung{}, ErrDatabase
	}
	if offen != nil {
		return *offen, nil
	}

	alle, err := e.KassensitzungenRepo.GetAllKassensitzungen(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list kassensitzungen")
		return kasse.Kassensitzung{}, ErrDatabase
	}
	if len(alle) == 0 {
		return kasse.Kassensitzung{}, ErrKassensitzungNichtGefunden
	}
	// GetAllKassensitzungen sorts by datum DESC, so alle[0] is the latest.
	return alle[0], nil
}

func (e Export) snapshot(ctx context.Context, ks kasse.Kassensitzung, erstellung time.Time) (dsfinvk.Snapshot, error) {
	log := zerolog.Ctx(ctx)

	ident, err := e.TSERepo.GetKassenidentitaet(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get kassenidentitaet")
		return dsfinvk.Snapshot{}, ErrDatabase
	}
	betreiber, err := e.BetreiberRepo.GetBetreiber(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get betreiber")
		return dsfinvk.Snapshot{}, ErrDatabase
	}
	stammdaten, err := e.TSERepo.GetTSEStammdaten(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get tse stammdaten")
		return dsfinvk.Snapshot{}, ErrDatabase
	}
	tischnamen, err := e.TischRepo.GetAlleTischNamen(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get tischnamen")
		return dsfinvk.Snapshot{}, ErrDatabase
	}

	return dsfinvk.Snapshot{
		KasseSeriennummer: ident.Seriennummer.String(),
		Erstellung:        erstellung,
		KassensitzungNr:   ks.ZNr,
		Betreiber:         betreiber,
		TSEStammdaten:     stammdaten,
		SoftwareVersion:   e.Version,
		Tischnamen:        tischnamen,
	}, nil
}

// dateiname builds the archive name from serial number, session and timestamp. The
// UTC timestamp appears in German local time; UTC would name the previous day just after midnight.
func dateiname(seriennummer string, nr int, zeitpunkt time.Time) string {
	return fmt.Sprintf("dsfinvk_%s_kassensitzung-%d_%s.zip", seriennummer, nr, zeitpunkt.In(zeit.Berlin).Format("20060102-150405"))
}
