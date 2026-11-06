package application

import (
	"context"
	"errors"
	"slices"

	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/internal/logging"
)

// PfandZurueckgeben writes the Rückgabe as version 1 of its own stream; tischID 0 is the
// Direktverkauf. rueckgabeID is the client's idempotency key: on an OCC conflict a stored Rückgabe
// with that ID and the same table and quantities means idempotent success, another content
// ErrPfandrueckgabeAbweichend, none a real 409.
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

	if err := c.persistVerkaufEvent(ctx, evt, kasse.StreamTypePfandrueckgabe, subject, 0, ks.ZNr, nil); err != nil {
		if errors.Is(err, ErrConflict) {
			gespeichert, found, err := c.getPfandrueckgabe(ctx, rueckgabeID)
			if err != nil {
				return err
			}
			if !found {
				return ErrConflict
			}
			gleich, err := kasse.GleicheRueckgabe(gespeichert, tischID, enrichment.MengenJeVariante(inputs))
			if err != nil {
				log.Error("Failed to compare stored pfandrueckgabe", "error", err, "rueckgabe_id", rueckgabeID)
				return ErrDatabase
			}
			if !gleich {
				log.Warn("Pfandrückgabe mit anderem Inhalt bereits gebucht", "rueckgabe_id", rueckgabeID)
				return ErrPfandrueckgabeAbweichend
			}
			log.Info("Idempotente Pfandrückgabe: rueckgabeId bereits vorhanden", "rueckgabe_id", rueckgabeID)
			return nil
		}
		if errors.Is(err, ErrKasseNichtGeoeffnet) {
			return err
		}
		log.Error("Failed to write pfand zurueckgegeben event", "error", err)
		return ErrDatabase
	}

	log.Info("Pfand zurückgegeben", "rueckgabe_id", rueckgabeID, "tisch_id", tischID)
	return nil
}

// PfandrueckgabeStornieren appends a cancellation to the Rückgabe's stream; it carries the
// Rückgabe's tischId. stornierungID is the client's idempotency key: a replay that already holds it
// answers success without booking.
func (c Command) PfandrueckgabeStornieren(ctx context.Context, userID int, userName string, stornierungID string, rueckgabeID string, positionen []kasse.PositionRef, kommentar string) error {
	log := logging.FromContext(ctx)

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
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
	if enthaeltStornierung(eintrag, stornierungID) {
		log.Info("Idempotente Pfandrückgabe-Stornierung: stornierungId bereits vorhanden", "stornierung_id", stornierungID)
		return nil
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

	if err := c.persistVerkaufEvent(ctx, evt, kasse.StreamTypePfandrueckgabe, subject, events[len(events)-1].Version, ks.ZNr, nil); err != nil {
		if errors.Is(err, ErrConflict) {
			_, aktuell, readErr := c.readPfandrueckgabe(ctx, subject)
			if readErr != nil {
				return readErr
			}
			if enthaeltStornierung(aktuell, stornierungID) {
				log.Info("Idempotente Pfandrückgabe-Stornierung nach OCC-Konflikt", "stornierung_id", stornierungID)
				return nil
			}
			return ErrConflict
		}
		if errors.Is(err, ErrKasseNichtGeoeffnet) {
			return err
		}
		log.Error("Failed to write pfandrueckgabe storniert event", "error", err, "rueckgabe_id", rueckgabeID)
		return ErrDatabase
	}

	log.Info("Pfandrückgabe storniert", "rueckgabe_id", rueckgabeID, "gesamt_stornierung_cents", gesamtStornierungCents)
	return nil
}

// pruefeVerrechnetenVerkauf answers a conflicting settled Direktverkauf: success only if the
// Rückgabe and the Verkauf written with it both match the request. A Verkauf without this Rückgabe,
// or a Rückgabe without this Verkauf, is a different Vorgang under a reused ID.
func (c Command) pruefeVerrechnetenVerkauf(ctx context.Context, verkaufID string, inputs []enrichment.PositionInput, pfand *enrichment.PfandEingabe) error {
	log := logging.FromContext(ctx)

	rueckgabe, found, err := c.getPfandrueckgabe(ctx, pfand.RueckgabeID)
	if err != nil {
		return err
	}
	if !found {
		verkaufExistiert, err := c.EventRepo.EventExistsByTypeAndVorgangsID(ctx, string(kasse.EventTypeDirektverkaufGetaetigtV1), verkaufID, "verkaufId")
		if err != nil {
			log.Error("Failed to lookup direktverkauf idempotency", "error", err, "verkauf_id", verkaufID)
			return ErrDatabase
		}
		if verkaufExistiert {
			log.Warn("Direktverkauf ohne diese Pfandrückgabe bereits gebucht", "verkauf_id", verkaufID, "rueckgabe_id", pfand.RueckgabeID)
			return ErrPfandrueckgabeAbweichend
		}
		return ErrConflict
	}

	zNr, err := kasse.ParseZNrFromSubject(rueckgabe.Subject)
	if err != nil {
		log.Error("Failed to parse pfandrueckgabe subject", "error", err, "subject", rueckgabe.Subject)
		return ErrDatabase
	}
	verkaufEvents, err := c.EventRepo.ReadEventsBySubject(ctx, kasse.DirektverkaufSubject(zNr, verkaufID))
	if err != nil {
		log.Error("Failed to read direktverkauf events", "error", err, "verkauf_id", verkaufID)
		return ErrDatabase
	}

	gleich, err := gleicheVerrechnung(rueckgabe, enrichment.MengenJeVariante(pfand.Positionen), verkaufEvents, enrichment.MengenJeVariante(inputs))
	if err != nil {
		log.Error("Failed to compare stored direktverkauf with pfand", "error", err, "verkauf_id", verkaufID)
		return ErrDatabase
	}
	if !gleich {
		log.Warn("Verrechneter Direktverkauf weicht vom gebuchten ab", "verkauf_id", verkaufID, "rueckgabe_id", pfand.RueckgabeID)
		return ErrPfandrueckgabeAbweichend
	}
	log.Info("Idempotenter Direktverkauf mit Pfand: beide Vorgänge bereits vorhanden", "verkauf_id", verkaufID, "rueckgabe_id", pfand.RueckgabeID)
	return nil
}

func gleicheVerrechnung(rueckgabe event.Event, pfandMengen map[int]int, verkaufEvents []event.Event, verkaufMengen map[int]int) (bool, error) {
	gleich, err := kasse.GleicheRueckgabe(rueckgabe, 0, pfandMengen)
	if err != nil || !gleich {
		return false, err
	}
	verkauf, found := kasse.VerrechnungsPartner(rueckgabe, verkaufEvents, kasse.EventTypeDirektverkaufGetaetigtV1)
	if !found {
		return false, nil
	}
	return kasse.GleicherVerkauf(verkauf, verkaufMengen)
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

func enthaeltStornierung(eintrag kasse.PfandrueckgabeHistorieEintrag, stornierungID string) bool {
	return slices.ContainsFunc(eintrag.Stornierungen, func(s kasse.VorgangStornierung) bool {
		return s.StornierungID == stornierungID
	})
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
