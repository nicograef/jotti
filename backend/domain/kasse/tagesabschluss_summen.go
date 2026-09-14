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

// ComputeAbschlussSummen aggregiert alle Events einer Kassensitzung zu den
// drei Z-Bon-Summen gemäß reporting.sql:11-44:
//
//	Umsatz        = Zahlungen + Direktverkäufe − Direktverkauf-Storni − Warenrücknahmen
//	Stornierungen = Warenrücknahmen + Korrekturen + Direktverkauf-Storni
//	Geldtransit   = Einlagen − Entnahmen
//
// Ein nicht parsebares Event eines summen-wirksamen Typs bricht die Berechnung ab:
// ein stiller falscher Z-Bon wäre schlimmer als ein blockierter Abschluss.
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
