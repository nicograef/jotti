package kasse

import (
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
		case EventTypeZahlungKassiertV1, EventTypeZahlungKassiertV2:
			data, err := e.Decode[ZahlungKassiertV1Data](evt)
			if err != nil {
				return AbschlussSummen{}, err
			}
			s.UmsatzCents += data.GesamtZahlungCents
		case EventTypeStornierungErteiltV1:
			data, err := e.Decode[StornierungErteiltV1Data](evt)
			if err != nil {
				return AbschlussSummen{}, err
			}
			s.UmsatzCents -= data.GesamtStornierungCents
			s.StornierungCents += data.GesamtStornierungCents
		case EventTypeBestellungKorrigiertV1:
			data, err := e.Decode[BestellungKorrigiertV1Data](evt)
			if err != nil {
				return AbschlussSummen{}, err
			}
			s.StornierungCents += data.GesamtCents
		case EventTypeDirektverkaufGetaetigtV1, EventTypeDirektverkaufGetaetigtV2:
			data, err := e.Decode[DirektverkaufGetaetigtV1Data](evt)
			if err != nil {
				return AbschlussSummen{}, err
			}
			s.UmsatzCents += data.GesamtbetragCents
		case EventTypeDirektverkaufStorniertV1:
			data, err := e.Decode[DirektverkaufStorniertV1Data](evt)
			if err != nil {
				return AbschlussSummen{}, err
			}
			s.UmsatzCents -= data.GesamtStornierungCents
			s.StornierungCents += data.GesamtStornierungCents
		case EventTypePfandZurueckgegebenV1:
			data, err := e.Decode[PfandZurueckgegebenV1Data](evt)
			if err != nil {
				return AbschlussSummen{}, err
			}
			s.UmsatzCents -= data.GesamtbetragCents
		case EventTypePfandrueckgabeStorniertV1:
			data, err := e.Decode[PfandrueckgabeStorniertV1Data](evt)
			if err != nil {
				return AbschlussSummen{}, err
			}
			s.UmsatzCents += data.GesamtStornierungCents
		case EventTypeGeldtransitGebuchtV1:
			data, err := e.Decode[GeldtransitGebuchtV1Data](evt)
			if err != nil {
				return AbschlussSummen{}, err
			}
			if data.Richtung == GeldtransitRichtungEinlage {
				s.GeldtransitCents += data.BetragCents
			} else {
				s.GeldtransitCents -= data.BetragCents
			}
		case EventTypeKassensitzungEroeffnetV1,
			EventTypeKassensturzDurchgefuehrtV1,
			EventTypeDifferenzSollIstGebuchtV1,
			EventTypeTagesabschlussErstelltV1,
			EventTypeBestellungAufgenommenV1,
			EventTypeBestellungUmgebuchtV1:
		}
	}
	return s, nil
}
