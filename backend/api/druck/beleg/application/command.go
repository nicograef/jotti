package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/betreiber"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
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
	EnqueueDruckauftraege(ctx context.Context, auftraege []druckstation.NeuerDruckauftrag) error
}

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
