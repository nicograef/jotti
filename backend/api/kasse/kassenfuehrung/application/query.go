package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/internal/logging"
)

type Query struct {
	KassenjournalRepo   kassenjournalRepo
	KassensitzungenRepo kassensitzungenRepo
}

// GetAktiveKassensitzung returns the Kassensitzung in status 'offen' or 'wird_abgeschlossen', nil if
// none — the barrier included, so an interrupted Kassenabschluss does not read as a closed Kasse.
func (q Query) GetAktiveKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error) {
	log := logging.FromContext(ctx)

	ks, err := q.KassensitzungenRepo.GetAktiveKassensitzung(ctx)
	if err != nil {
		log.Error("Failed to get aktive Kassensitzung", "error", err)
		return nil, ErrDatabase
	}

	log.Debug("Retrieved aktive Kassensitzung")
	return ks, nil
}

func (q Query) GetKassenbestand(ctx context.Context, kassensitzungNr int) (kasse.Kassenbestand, error) {
	log := logging.FromContext(ctx)

	bestand, err := q.KassenjournalRepo.GetKassenbestand(ctx, kassensitzungNr)
	if err != nil {
		log.Error("Failed to get Kassenbestand", "error", err, "z_nr", kassensitzungNr)
		return kasse.Kassenbestand{}, ErrDatabase
	}

	log.Debug("Retrieved Kassenbestand", "z_nr", kassensitzungNr, "bestand_cents", bestand.SollBestandCents)
	return bestand, nil
}

// GetGeldtransitListe returns the Einlagen/Entnahmen of the Kassensitzung, newest first.
func (q Query) GetGeldtransitListe(ctx context.Context, kassensitzungNr int) ([]kasse.Geldtransit, error) {
	log := logging.FromContext(ctx)

	buchungen, err := q.KassenjournalRepo.GetGeldtransitListe(ctx, kassensitzungNr)
	if err != nil {
		log.Error("Failed to get Geldtransit-Liste", "error", err, "z_nr", kassensitzungNr)
		return nil, ErrDatabase
	}

	log.Debug("Retrieved Geldtransit-Liste", "z_nr", kassensitzungNr, "anzahl", len(buchungen))
	return buchungen, nil
}
