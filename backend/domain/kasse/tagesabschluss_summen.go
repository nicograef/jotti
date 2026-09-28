package kasse

import (
	"encoding/json"
	"fmt"

	e "github.com/nicograef/jotti/backend/domain/event"
)

type AbschlussSummen struct {
	UmsatzCents      int
	StornierungCents int
	GeldtransitCents int
}

// ComputeAbschlussSummen yields the three Z-Bon sums with the formulas of GetReportingStats in
// sqlc/queries/reporting.sql. An unparsable sum-relevant event aborts: a silently wrong Z-Bon is
// worse than a blocked Abschluss.
func ComputeAbschlussSummen(events []e.Event) (AbschlussSummen, error) {
	var s AbschlussSummen
	for _, evt := range events {
		switch EventType(evt.Type) {
		case EventTypeZahlungKassiertV1:
			var data ZahlungKassiertV1Data
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return AbschlussSummen{}, fmt.Errorf("event %d (%s): %w", evt.ID, evt.Type, err)
			}
			s.UmsatzCents += data.GesamtZahlungCents
		case EventTypeStornierungErteiltV1:
			var data StornierungErteiltV1Data
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return AbschlussSummen{}, fmt.Errorf("event %d (%s): %w", evt.ID, evt.Type, err)
			}
			s.UmsatzCents -= data.GesamtStornierungCents
			s.StornierungCents += data.GesamtStornierungCents
		case EventTypeBestellungKorrigiertV1:
			var data BestellungKorrigiertV1Data
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return AbschlussSummen{}, fmt.Errorf("event %d (%s): %w", evt.ID, evt.Type, err)
			}
			s.StornierungCents += data.GesamtCents
		case EventTypeDirektverkaufGetaetigtV1:
			var data DirektverkaufGetaetigtV1Data
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return AbschlussSummen{}, fmt.Errorf("event %d (%s): %w", evt.ID, evt.Type, err)
			}
			s.UmsatzCents += data.GesamtbetragCents
		case EventTypeDirektverkaufStorniertV1:
			var data DirektverkaufStorniertV1Data
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return AbschlussSummen{}, fmt.Errorf("event %d (%s): %w", evt.ID, evt.Type, err)
			}
			s.UmsatzCents -= data.GesamtStornierungCents
			s.StornierungCents += data.GesamtStornierungCents
		case EventTypeGeldtransitGebuchtV1:
			var data GeldtransitGebuchtV1Data
			if err := json.Unmarshal(evt.Data, &data); err != nil {
				return AbschlussSummen{}, fmt.Errorf("event %d (%s): %w", evt.ID, evt.Type, err)
			}
			if data.Richtung == GeldtransitRichtungEinlage {
				s.GeldtransitCents += data.BetragCents
			} else {
				s.GeldtransitCents -= data.BetragCents
			}
		}
	}
	return s, nil
}
