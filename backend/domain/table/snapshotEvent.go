package table

import (
	"fmt"
	"strconv"
	"time"

	z "github.com/Oudwins/zog"
	e "github.com/nicograef/jotti/backend/domain/event"
)

type snapshotV1Data struct {
	SaldoCents             int        `json:"saldoCents"`
	UnbezahltePositionen   []Position `json:"unbezahltePositionen"`
	UngeliefertePositionen []Position `json:"ungeliefertePositionen"`
	GesamtZahlungenCents   int        `json:"gesamtZahlungenCents"`
}

var snapshotV1DataSchema = z.Struct(z.Shape{
	"SaldoCents":             z.Int(),                 // Can be 0 (no outstanding balance)
	"UnbezahltePositionen":   z.Slice(positionSchema), // Can be empty slice
	"UngeliefertePositionen": z.Slice(positionSchema), // Can be empty slice
	"GesamtZahlungenCents":   z.Int(),                 // Can be 0 (no payments yet)
})

// Snapshot represents the materialized state at a point in time
type Snapshot struct {
	TischID                int
	SaldoCents             int
	UnbezahltePositionen   []Position
	UngeliefertePositionen []Position
	GesamtZahlungenCents   int
	CreatedAt              time.Time
}

func NewSnapshotEvent(userID int, userName string, tischID int, saldo int, unbezahlt, ungeliefert []Position, gesamtZahlungen int) (e.Event, error) {
	data := snapshotV1Data{
		SaldoCents:             saldo,
		UnbezahltePositionen:   unbezahlt,
		UngeliefertePositionen: ungeliefert,
		GesamtZahlungenCents:   gesamtZahlungen,
	}

	if err := snapshotV1DataSchema.Validate(&data); err != nil {
		issues := z.Issues.FlattenAndCollect(err)
		return e.Event{}, fmt.Errorf("snapshot data validation failed: %v", issues)
	}

	return e.New(userID, userName, string(EventTypeSnapshotV1), "tisch:"+strconv.Itoa(tischID), data)
}

func buildSnapshotFromEvent(event e.Event) (Snapshot, error) {
	if event.Type != string(EventTypeSnapshotV1) {
		return Snapshot{}, fmt.Errorf("unsupported event type: %s", event.Type)
	}

	tischID, err := parseTischIDFromSubject(event.Subject)
	if err != nil {
		return Snapshot{}, err
	}

	var data snapshotV1Data
	if err := e.ParseData(event, &data, snapshotV1DataSchema); err != nil {
		return Snapshot{}, err
	}

	return Snapshot{
		TischID:                tischID,
		SaldoCents:             data.SaldoCents,
		UnbezahltePositionen:   data.UnbezahltePositionen,
		UngeliefertePositionen: data.UngeliefertePositionen,
		GesamtZahlungenCents:   data.GesamtZahlungenCents,
		CreatedAt:              event.Time,
	}, nil
}
