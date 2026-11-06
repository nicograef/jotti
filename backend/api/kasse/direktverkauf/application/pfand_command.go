package application

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/internal/logging"
)

// PfandZurueckgeben writes the Rückgabe as version 1 of its own stream; tischID 0 is the
// Direktverkauf. rueckgabeID is the client's idempotency key: on an OCC conflict see
// pruefeReineRueckgabe.
func (c Command) PfandZurueckgeben(ctx context.Context, userID int, userName string, rueckgabeID string, tischID int, inputs []enrichment.PositionInput, kommentar string) error {
	log := logging.FromContext(ctx)

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return err
	}

	if tischID > 0 {
		if err := c.pruefeAktivenTisch(ctx, tischID); err != nil {
			return err
		}
	}

	positionen, err := enrichment.EnrichPfandPositionen(ctx, c.ProduktRepo, inputs)
	if err != nil {
		return err
	}

	subject := kasse.PfandrueckgabeSubject(ks.ZNr, rueckgabeID)

	evt, err := kasse.NewPfandZurueckgegebenEvent(subject, rueckgabeID, tischID, userID, userName, positionen, kommentar)
	if err != nil {
		log.Error("Failed to create pfand zurueckgegeben event", "error", err)
		return err
	}

	if err := c.writeEventOCC(ctx, evt, kasse.StreamTypePfandrueckgabe, subject, 0, ks.ZNr); err != nil {
		if errors.Is(err, ErrConflict) {
			return c.pruefeReineRueckgabe(ctx, rueckgabeID, tischID, inputs)
		}
		return err
	}

	log.Info("Pfand zurückgegeben", "rueckgabe_id", rueckgabeID, "tisch_id", tischID)
	return nil
}

// pruefeReineRueckgabe answers a conflicting pure Rückgabe: success only if the stored Rückgabe books
// this table and these quantities and no Zahlung or Verkauf was written with it. A settled Rückgabe
// was offset, so answering success would pay out cash the journal never booked.
func (c Command) pruefeReineRueckgabe(ctx context.Context, rueckgabeID string, tischID int, inputs []enrichment.PositionInput) error {
	log := logging.FromContext(ctx)

	gespeichert, found, err := c.getPfandrueckgabe(ctx, rueckgabeID)
	if err != nil {
		return err
	}
	if !found {
		return ErrConflict
	}
	gleich, err := kasse.GleicheRueckgabe(gespeichert, tischID, enrichment.MengenJeVariante(inputs))
	if err == nil && gleich {
		var verrechnet bool
		verrechnet, err = c.rueckgabeVerrechnet(ctx, gespeichert, tischID)
		gleich = !verrechnet
	}
	if err != nil {
		log.Error("Failed to compare stored pfandrueckgabe", "error", err, "rueckgabe_id", rueckgabeID)
		return ErrDatabase
	}
	if !gleich {
		log.Warn("Pfandrückgabe mit anderem Inhalt oder verrechnet bereits gebucht", "rueckgabe_id", rueckgabeID)
		return ErrPfandrueckgabeAbweichend
	}
	log.Info("Idempotente Pfandrückgabe: rueckgabeId bereits vorhanden", "rueckgabe_id", rueckgabeID)
	return nil
}

// PfandrueckgabeStornieren appends a cancellation to the Rückgabe's stream; it carries the
// Rückgabe's tischId. stornierungID is the client's idempotency key, unique across all Rückgaben:
// see pruefeGebuchteStornierung.
func (c Command) PfandrueckgabeStornieren(ctx context.Context, userID int, userName string, stornierungID string, rueckgabeID string, positionen []kasse.PositionRef, kommentar string) error {
	log := logging.FromContext(ctx)

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return err
	}

	if gebucht, err := c.pruefeGebuchteStornierung(ctx, stornierungID, rueckgabeID, positionen); gebucht || err != nil {
		return err
	}

	subject := kasse.PfandrueckgabeSubject(ks.ZNr, rueckgabeID)

	events, eintrag, err := c.readPfandrueckgabe(ctx, subject)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return ErrRueckgabeNichtGefunden
	}

	if !kasse.ValidatePositionRefs(eintrag.OffenePositionen, positionen) {
		log.Warn("Storno-Invariante verletzt: angeforderte Pfand-Positionen nicht stornierbar", "rueckgabe_id", rueckgabeID)
		return ErrPositionNichtStornierbar
	}

	resolved, gesamtStornierungCents := kasse.ResolvePositionen(eintrag.OffenePositionen, positionen)

	evt, err := kasse.NewPfandrueckgabeStorniertEvent(subject, stornierungID, rueckgabeID, eintrag.TischID, userID, userName, resolved, gesamtStornierungCents, kommentar)
	if err != nil {
		log.Error("Failed to create pfandrueckgabe storniert event", "error", err, "rueckgabe_id", rueckgabeID)
		return err
	}

	if err := c.writeEventOCC(ctx, evt, kasse.StreamTypePfandrueckgabe, subject, events[len(events)-1].Version, ks.ZNr); err != nil {
		if errors.Is(err, ErrConflict) {
			if gebucht, err := c.pruefeGebuchteStornierung(ctx, stornierungID, rueckgabeID, positionen); gebucht || err != nil {
				return err
			}
		}
		return err
	}

	log.Info("Pfandrückgabe storniert", "rueckgabe_id", rueckgabeID, "gesamt_stornierung_cents", gesamtStornierungCents)
	return nil
}

// pruefeGebuchteStornierung reports whether stornierungID is already booked. A booked one answers
// success only if it cancels the same quantities of the same Rückgabe, else
// ErrPfandrueckgabeAbweichend: a reused ID would otherwise skip or duplicate a Storno.
func (c Command) pruefeGebuchteStornierung(ctx context.Context, stornierungID string, rueckgabeID string, positionen []kasse.PositionRef) (bool, error) {
	log := logging.FromContext(ctx)

	gespeichert, found, err := c.EventRepo.GetPfandrueckgabeStornierung(ctx, stornierungID)
	if err != nil {
		log.Error("Failed to read pfandrueckgabe stornierung", "error", err, "stornierung_id", stornierungID)
		return false, ErrDatabase
	}
	if !found {
		return false, nil
	}
	gleich, err := kasse.GleicheStornierung(gespeichert, rueckgabeID, positionen)
	if err != nil {
		log.Error("Failed to compare stored pfandrueckgabe stornierung", "error", err, "stornierung_id", stornierungID)
		return true, ErrDatabase
	}
	if !gleich {
		log.Warn("Pfandrückgabe-Stornierung mit anderem Inhalt bereits gebucht", "stornierung_id", stornierungID, "rueckgabe_id", rueckgabeID)
		return true, ErrPfandrueckgabeAbweichend
	}
	log.Info("Idempotente Pfandrückgabe-Stornierung: stornierungId bereits vorhanden", "stornierung_id", stornierungID)
	return true, nil
}

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

// rueckgabeVerrechnet reports whether a Zahlung (tischID > 0) or a Verkauf (tischID 0) was written
// with the stored Rückgabe.
func (c Command) rueckgabeVerrechnet(ctx context.Context, rueckgabe event.Event, tischID int) (bool, error) {
	zNr, err := kasse.ParseZNrFromSubject(rueckgabe.Subject)
	if err != nil {
		return false, err
	}
	var partner []event.Event
	typen := []kasse.EventType{kasse.EventTypeZahlungKassiertV1, kasse.EventTypeZahlungKassiertV2}
	if tischID == 0 {
		typen = []kasse.EventType{kasse.EventTypeDirektverkaufGetaetigtV1, kasse.EventTypeDirektverkaufGetaetigtV2}
		partner, err = c.EventRepo.ReadDirektverkaufEvents(ctx, zNr)
	} else {
		partner, err = c.EventRepo.ReadEventsBySubject(ctx, kasse.TischSessionSubject(zNr, tischID))
	}
	if err != nil {
		return false, err
	}
	_, found := kasse.VerrechnungsPartner(rueckgabe, partner, typen...)
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

func (c Command) readPfandrueckgabe(ctx context.Context, subject string) ([]event.Event, kasse.PfandrueckgabeHistorieEintrag, error) {
	log := logging.FromContext(ctx)

	events, err := c.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		log.Error("Failed to read pfandrueckgabe events", "error", err, "subject", subject)
		return nil, kasse.PfandrueckgabeHistorieEintrag{}, ErrDatabase
	}
	eintrag, err := kasse.BuildPfandrueckgabeHistorieEintrag(events)
	if err != nil {
		log.Error("Failed to replay pfandrueckgabe", "error", err, "subject", subject)
		return nil, kasse.PfandrueckgabeHistorieEintrag{}, ErrDatabase
	}
	return events, eintrag, nil
}

func (c Command) pruefeAktivenTisch(ctx context.Context, tischID int) error {
	log := logging.FromContext(ctx)

	t, err := c.TischRepo.GetTisch(ctx, tischID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn("Tisch not found", "tisch_id", tischID)
			return ErrTischNichtGefunden
		}
		log.Error("Failed to load tisch", "error", err, "tisch_id", tischID)
		return ErrDatabase
	}
	if t.Status != tisch.ActiveStatus {
		log.Warn("Tisch is not active", "tisch_id", tischID, "status", string(t.Status))
		return ErrTischNichtAktiv
	}
	return nil
}
