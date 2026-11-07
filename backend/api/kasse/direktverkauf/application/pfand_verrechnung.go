package application

import (
	"context"

	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/internal/logging"
)

// pruefeVerrechnetenVerkauf answers a conflicting settled Direktverkauf: success only if the
// Rückgabe and the Verkauf written with it both match the request. A Verkauf without this Rückgabe,
// or a Rückgabe without this Verkauf, is a different Vorgang under a reused ID.
func (c Command) pruefeVerrechnetenVerkauf(ctx context.Context, verkaufID string, inputs []enrichment.PositionInput, pfand *enrichment.PfandEingabe) error {
	log := logging.FromContext(ctx)

	rueckgabe, rueckgabeGefunden, err := c.getPfandrueckgabe(ctx, pfand.RueckgabeID)
	if err != nil {
		return err
	}
	verkauf, verkaufGefunden, err := c.getDirektverkauf(ctx, verkaufID)
	if err != nil {
		return err
	}
	if !rueckgabeGefunden && !verkaufGefunden {
		return ErrConflict
	}

	gleich := rueckgabeGefunden && verkaufGefunden
	if gleich {
		gleich, err = gleicheVerrechnung(rueckgabe, enrichment.MengenJeVariante(pfand.Positionen), verkauf, enrichment.MengenJeVariante(inputs))
		if err != nil {
			log.Error("Failed to compare stored direktverkauf with pfand", "error", err, "verkauf_id", verkaufID)
			return ErrDatabase
		}
	}
	if !gleich {
		log.Warn("Verrechneter Direktverkauf weicht vom gebuchten ab", "verkauf_id", verkaufID, "rueckgabe_id", pfand.RueckgabeID)
		return ErrPfandrueckgabeAbweichend
	}
	log.Info("Idempotenter Direktverkauf mit Pfand: beide Vorgänge bereits vorhanden", "verkauf_id", verkaufID, "rueckgabe_id", pfand.RueckgabeID)
	return nil
}

func gleicheVerrechnung(rueckgabe event.Event, pfandMengen map[int]int, verkauf event.Event, verkaufMengen map[int]int) (bool, error) {
	gleich, err := kasse.GleicheRueckgabe(rueckgabe, 0, pfandMengen)
	if err != nil || !gleich {
		return false, err
	}
	if _, found := kasse.VerrechnungsPartner(rueckgabe, []event.Event{verkauf}, kasse.EventTypeDirektverkaufGetaetigtV1, kasse.EventTypeDirektverkaufGetaetigtV2); !found {
		return false, nil
	}
	return kasse.GleicherVerkauf(verkauf, verkaufMengen)
}

// pruefeUnverrechnetenVerkauf answers a conflicting Direktverkauf without Pfand: success only if the
// stored Verkauf was booked without a Rückgabe too, else it collected only the net price under this ID.
func (c Command) pruefeUnverrechnetenVerkauf(ctx context.Context, verkaufID string) error {
	log := logging.FromContext(ctx)

	verkauf, found, err := c.getDirektverkauf(ctx, verkaufID)
	if err != nil {
		return err
	}
	if !found {
		return ErrConflict
	}
	verrechnet, err := c.verkaufVerrechnet(ctx, verkauf)
	if err != nil {
		log.Error("Failed to look up the pfandrueckgabe of a direktverkauf", "error", err, "verkauf_id", verkaufID)
		return ErrDatabase
	}
	if verrechnet {
		log.Warn("Direktverkauf mit Pfandrückgabe bereits gebucht", "verkauf_id", verkaufID)
		return ErrPfandrueckgabeAbweichend
	}
	log.Info("Idempotenter Direktverkauf: verkaufId bereits vorhanden", "verkauf_id", verkaufID)
	return nil
}

// verkaufVerrechnet reports whether a Rückgabe was written with the stored Verkauf.
func (c Command) verkaufVerrechnet(ctx context.Context, verkauf event.Event) (bool, error) {
	zNr, err := kasse.ParseZNrFromSubject(verkauf.Subject)
	if err != nil {
		return false, err
	}
	rueckgaben, err := c.EventRepo.ReadPfandrueckgabeEvents(ctx, zNr, 0)
	if err != nil {
		return false, err
	}
	_, found := kasse.VerrechnungsPartner(verkauf, rueckgaben, kasse.EventTypePfandZurueckgegebenV1)
	return found, nil
}

func (c Command) getDirektverkauf(ctx context.Context, verkaufID string) (event.Event, bool, error) {
	verkauf, found, err := c.EventRepo.GetDirektverkauf(ctx, verkaufID)
	if err != nil {
		logging.FromContext(ctx).Error("Failed to read direktverkauf", "error", err, "verkauf_id", verkaufID)
		return event.Event{}, false, ErrDatabase
	}
	return verkauf, found, nil
}

func (c Command) getPfandrueckgabe(ctx context.Context, rueckgabeID string) (event.Event, bool, error) {
	rueckgabe, found, err := c.EventRepo.GetPfandrueckgabe(ctx, rueckgabeID)
	if err != nil {
		logging.FromContext(ctx).Error("Failed to read pfandrueckgabe", "error", err, "rueckgabe_id", rueckgabeID)
		return event.Event{}, false, ErrDatabase
	}
	return rueckgabe, found, nil
}
