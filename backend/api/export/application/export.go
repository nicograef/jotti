// Package application orchestriert den DSFinV-K-Export: es lädt Events und
// Stammdaten einer Kassensitzung und reicht sie an den reinen dsfinvk-Mapper
// weiter. Die fiskalische Transformation selbst liegt im Domain-Paket dsfinvk.
package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/dsfinvk"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/nicograef/jotti/backend/domain/table"
	"github.com/rs/zerolog"
)

var (
	ErrDatabase = db.ErrDatabase
	// ErrKassensitzungNichtGefunden meldet eine unbekannte Kassensitzung (404).
	ErrKassensitzungNichtGefunden = errors.New("kassensitzung nicht gefunden")
	// ErrLeereKassensitzung meldet eine Sitzung ohne abrechenbare Vorgänge (400).
	ErrLeereKassensitzung = dsfinvk.ErrKeineVorgaenge
)

type kassenjournalRepo interface {
	ReadEventsByKassensitzung(ctx context.Context, kassensitzungNr int) ([]event.Event, error)
}

type kassensitzungenRepo interface {
	GetOffeneKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
	GetAllKassensitzungen(ctx context.Context) ([]kasse.Kassensitzung, error)
}

type settingsRepo interface {
	GetKassenidentitaet(ctx context.Context) (settings.Kassenidentitaet, error)
	GetBetreiber(ctx context.Context) (settings.Betreiber, error)
	GetTSEStammdaten(ctx context.Context) (settings.TSEStammdaten, error)
}

type tableRepo interface {
	GetAllTables(ctx context.Context) ([]table.Tisch, error)
}

// Export ist der App-Service, der das DSFinV-K-Archiv einer Kassensitzung
// erzeugt.
type Export struct {
	KassenjournalRepo   kassenjournalRepo
	KassensitzungenRepo kassensitzungenRepo
	SettingsRepo        settingsRepo
	TableRepo           tableRepo
}

// Archiv ist das fertige DSFinV-K-ZIP samt sprechendem Dateinamen.
type Archiv struct {
	Dateiname string
	Inhalt    []byte
}

// Erstellen erzeugt das DSFinV-K-Archiv für die gewählte Kassensitzung. nr == 0
// wählt die Standard-Sitzung: die offene, sonst die jüngste abgeschlossene.
func (e Export) Erstellen(ctx context.Context, nr int) (Archiv, error) {
	log := zerolog.Ctx(ctx)

	ks, err := e.resolveKassensitzung(ctx, nr)
	if err != nil {
		return Archiv{}, err
	}

	events, err := e.KassenjournalRepo.ReadEventsByKassensitzung(ctx, ks.ZNr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read events for dsfinvk export")
		return Archiv{}, ErrDatabase
	}

	erstellung := dsfinvk.Erstellungszeitpunkt(events, time.Now().UTC())

	snapshot, err := e.snapshot(ctx, ks, erstellung)
	if err != nil {
		return Archiv{}, err
	}

	inhalt, err := dsfinvk.BuildArchive(snapshot, events)
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

// resolveKassensitzung wählt die zu exportierende Sitzung. Eine explizit
// angeforderte Nummer muss existieren, sonst ErrKassensitzungNichtGefunden.
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

	offen, err := e.KassensitzungenRepo.GetOffeneKassensitzung(ctx)
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
	// GetAllKassensitzungen ist nach datum DESC sortiert — alle[0] ist die jüngste.
	return alle[0], nil
}

// snapshot lädt die Stammdaten für den Export. erstellung ist Z_ERSTELLUNG: bei
// einer abgeschlossenen Sitzung die Zeit des Tagesabschluss-Events, bei einer
// offenen der Exportzeitpunkt.
func (e Export) snapshot(ctx context.Context, ks kasse.Kassensitzung, erstellung time.Time) (dsfinvk.Snapshot, error) {
	log := zerolog.Ctx(ctx)

	ident, err := e.SettingsRepo.GetKassenidentitaet(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get kassenidentitaet")
		return dsfinvk.Snapshot{}, ErrDatabase
	}
	betreiber, err := e.SettingsRepo.GetBetreiber(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get betreiber")
		return dsfinvk.Snapshot{}, ErrDatabase
	}
	stammdaten, err := e.SettingsRepo.GetTSEStammdaten(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get tse stammdaten")
		return dsfinvk.Snapshot{}, ErrDatabase
	}
	tische, err := e.TableRepo.GetAllTables(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get tische")
		return dsfinvk.Snapshot{}, ErrDatabase
	}

	tischnamen := make(map[int]string, len(tische))
	for _, t := range tische {
		tischnamen[t.ID] = t.Name
	}

	return dsfinvk.Snapshot{
		KasseSeriennummer: ident.Seriennummer.String(),
		Erstellung:        erstellung,
		KassensitzungNr:   ks.ZNr,
		Betreiber:         betreiber,
		TSEStammdaten:     stammdaten,
		Tischnamen:        tischnamen,
	}, nil
}

// dateiname baut den sprechenden Archivnamen aus Seriennummer, Kassensitzung
// und Zeitstempel.
func dateiname(seriennummer string, nr int, zeitpunkt time.Time) string {
	return fmt.Sprintf("dsfinvk_%s_kassensitzung-%d_%s.zip", seriennummer, nr, zeitpunkt.Format("20060102-150405"))
}
