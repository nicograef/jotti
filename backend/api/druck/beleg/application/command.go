package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/betreiber"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
)

type eventRepo interface {
	ReadTischSession(ctx context.Context, subject string) (kasse.TischSession, error)
	ReadEventsBySubject(ctx context.Context, subject string) ([]event.Event, error)
}

type kassensitzungenRepo interface {
	GetAktiveKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
}

type druckstationRepo interface {
	GetKonfigurierteDruckstationen(ctx context.Context) (map[string]druckstation.Druckstation, error)
}

type druckauftragRepo interface {
	EnqueueDruckauftraege(ctx context.Context, auftraege []druckauftrag_repo.NeuerDruckauftrag) error
}

// tseAuftragRepo liefert den Signatur-Stand eines Events aus der
// Signaturauftrags-Tabelle und den aktiven Stoerungszeitraum aus dem
// Stoerungsprotokoll — die beiden Eingaben der Signaturstatus-Funktion
// (Beleg-Abruf liest genau eine Signaturquelle) — sowie die Kassenidentitaet
// (Seriennummer) fuer den Beleg-Kopf.
type tseAuftragRepo interface {
	GetSignaturauftragZuEvent(ctx context.Context, eventID int) (tse.SignaturauftragStand, error)
	GetAktiveTSEStoerung(ctx context.Context) (*tse.Stoerung, error)
	GetKassenidentitaet(ctx context.Context) (tse.Kassenidentitaet, error)
}

type betreiberRepo interface {
	GetBetreiber(ctx context.Context) (betreiber.Betreiber, error)
}

type Command struct {
	EventRepo           eventRepo
	KassensitzungenRepo kassensitzungenRepo
	DruckstationRepo    druckstationRepo
	DruckauftragRepo    druckauftragRepo
	BetreiberRepo       betreiberRepo
	TSERepo             tseAuftragRepo
}

func (c Command) getOffeneKassensitzungOderFehler(ctx context.Context) (*kasse.Kassensitzung, error) {
	ks, err := c.KassensitzungenRepo.GetAktiveKassensitzung(ctx)
	if err != nil {
		return nil, ErrDatabase
	}
	if ks == nil {
		return nil, ErrKasseNichtGeoeffnet
	}
	if ks.Status == kasse.KassensitzungWirdAbgeschlossen {
		return nil, ErrKasseWirdAbgeschlossen
	}
	return ks, nil
}
