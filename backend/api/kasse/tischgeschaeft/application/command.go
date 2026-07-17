package application

import (
	"context"
	"errors"

	bondruckApp "github.com/nicograef/jotti/backend/api/druck/bondruck/application"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/rs/zerolog"
)

type tischRepo interface {
	GetTable(ctx context.Context, id int) (tisch.Tisch, error)
	GetActiveTables(ctx context.Context, kassensitzungNr int) ([]tisch.AktiverTisch, error)
	GetActiveTablesWithFavorites(ctx context.Context, userID int, kassensitzungNr int) ([]tisch.AktiverTischMitFavorit, error)
}

type eventRepo interface {
	WriteEvent(ctx context.Context, e event.Event, streamType kasse.StreamType, kassensitzungNr int) (int, error)
	WriteEventWithDruckauftraege(ctx context.Context, e event.Event, streamType kasse.StreamType, kassensitzungNr int, buildAuftraege func(event.Event) []druckauftrag_repo.NeuerDruckauftrag) (int, error)
	WriteUmbuchung(ctx context.Context, quellEvent event.Event, zielEvent event.Event, kassensitzungNr int) error
	WriteTischSessionEventsAtomic(ctx context.Context, events []event.Event, kassensitzungNr int) error
	ReadTischSession(ctx context.Context, subject string) (kasse.TischSession, error)
	ReadFavoritenTischStates(ctx context.Context, tischIDs []int, kassensitzungNr int) (map[int]kassenjournal_repo.TischNameUndSession, error)
	GetMaxVersion(ctx context.Context, subject string) (int, error)
	ReadEventsBySubject(ctx context.Context, subject string) ([]event.Event, error)
	EventExistsByTypeAndVorgangsID(ctx context.Context, eventType, vorgangsID, jsonKey string) (bool, error)
}

type kassensitzungenRepo interface {
	GetOffeneKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
	GetAktiveKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
}

type produktRepo interface {
	GetVariantsByIDs(ctx context.Context, ids []int) (map[int]produkt.Variante, error)
	GetProductsByIDs(ctx context.Context, ids []int) (map[int]produkt.Produkt, error)
}

type favoritRepo interface {
	GetByUser(ctx context.Context, userID int) ([]int, error)
}

type druckstationRepo interface {
	GetKonfigurierteDruckstationen(ctx context.Context) (map[string]druckstation.Druckstation, error)
}

type Command struct {
	TischRepo           tischRepo
	EventRepo           eventRepo
	ProduktRepo         produktRepo
	FavoritRepo         favoritRepo
	KassensitzungenRepo kassensitzungenRepo
	DruckstationRepo    druckstationRepo
}

// getOffeneKassensitzungOderFehler retrieves the currently open Kassensitzung for a booking.
// Returns ErrKasseNichtGeoeffnet (HTTP 409) when none is active and ErrKasseWirdAbgeschlossen while
// the Kassensitzung is being closed (barrier active), rejecting the booking before any TSE roundtrip.
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

// writeEventOCC writes the event with version expectedVersion+1, mapping a version
// conflict (UNIQUE violation) to ErrConflict.
//
// expectedVersion muss die Version des Zustands sein, gegen den der Command validiert
// hat (Projektion bzw. Replay) — nicht ein frisches GetMaxVersion zum Schreibzeitpunkt.
// Nur so erkennt der UNIQUE(subject, version)-Constraint, dass sich der Stream seit dem
// Lesen geändert hat, und verhindert Doppel-Writes auf Basis veralteter Validierung.
func writeEventOCC(ctx context.Context, e event.Event, subject string, expectedVersion int, write func(event.Event) (int, error)) (int, error) {
	e.Version = expectedVersion + 1

	eventID, err := write(e)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			zerolog.Ctx(ctx).Warn().
				Int("version", e.Version).
				Str("subject", subject).
				Msg("OCC conflict")
			return 0, ErrConflict
		}
		if errors.Is(err, db.ErrConflict) {
			zerolog.Ctx(ctx).Warn().Str("subject", subject).Msg("Deadlock on event write")
			return 0, ErrConflict
		}
		if errors.Is(err, kassenjournal_repo.ErrKassensitzungNichtOffen) {
			zerolog.Ctx(ctx).Warn().Str("subject", subject).Msg("Kassensitzung nicht mehr offen")
			return 0, ErrKasseNichtGeoeffnet
		}
		return 0, err
	}

	return eventID, nil
}

// writeEvent writes an event with optimistic concurrency control against
// expectedVersion. Returns ErrConflict on a version conflict.
func writeEvent(ctx context.Context, repo eventRepo, e event.Event, subject string, expectedVersion int, streamType kasse.StreamType, kassensitzungNr int) (int, error) {
	return writeEventOCC(ctx, e, subject, expectedVersion, func(versioned event.Event) (int, error) {
		return repo.WriteEvent(ctx, versioned, streamType, kassensitzungNr)
	})
}

// writeEventWithDruckauftraege writes an event and the Druckaufträge derived from it
// (built from the stored event including its generated ID) in a single transaction
// (transactional outbox). Returns ErrConflict on a version conflict.
func writeEventWithDruckauftraege(ctx context.Context, repo eventRepo, e event.Event, subject string, expectedVersion int, streamType kasse.StreamType, kassensitzungNr int, buildAuftraege func(event.Event) []druckauftrag_repo.NeuerDruckauftrag) (int, error) {
	return writeEventOCC(ctx, e, subject, expectedVersion, func(versioned event.Event) (int, error) {
		return repo.WriteEventWithDruckauftraege(ctx, versioned, streamType, kassensitzungNr, buildAuftraege)
	})
}

// persistTischEvent writes a tisch-session event with OCC against expectedVersion
// (die Version des gelesenen Zustands, gegen den validiert wurde). An OCC conflict
// maps to ErrConflict, any other write error to ErrDatabase. aktion is the success
// log message.
func (c Command) persistTischEvent(ctx context.Context, evt event.Event, subject string, expectedVersion int, kassensitzungNr int, tischID int, aktion string) error {
	log := zerolog.Ctx(ctx)

	if _, err := writeEvent(ctx, c.EventRepo, evt, subject, expectedVersion, kasse.StreamTypeTischSession, kassensitzungNr); err != nil {
		if errors.Is(err, ErrConflict) {
			return ErrConflict
		}
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to write event to database")
		return ErrDatabase
	}
	log.Info().Int("tisch_id", tischID).Msg(aktion)
	return nil
}

// loadTischState loads and validates the Tisch, then reads its projected TischSession.
// Returns the subject, kassensitzungNr, and TischSession state. Returns
// ErrKasseNichtGeoeffnet if no open Kassensitzung exists, ErrTischNotFound if the
// Tisch doesn't exist and ErrTischNotActive if it is not active.
func (c Command) loadTischState(ctx context.Context, tischID int) (string, int, kasse.TischSession, error) {
	log := zerolog.Ctx(ctx)

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return "", 0, kasse.TischSession{}, err
	}

	t, err := c.TischRepo.GetTable(ctx, tischID)
	if err != nil {
		return "", 0, kasse.TischSession{}, fromRepositoryError(err, log, tischID)
	}

	if t.Status != tisch.ActiveStatus {
		log.Warn().Int("tisch_id", tischID).Str("status", string(t.Status)).Msg("Tisch is not active")
		return "", 0, kasse.TischSession{}, ErrTischNotActive
	}

	subject := kasse.TischSessionSubject(ks.ZNr, tischID)

	state, err := c.EventRepo.ReadTischSession(ctx, subject)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to read tisch session")
		return "", 0, kasse.TischSession{}, ErrDatabase
	}

	return subject, ks.ZNr, state, nil
}

// BestellungAufnehmen nimmt eine Bestellung für einen Tisch auf.
// bestellungID ist eine client-seitig erzeugte UUID (Idempotenz-Schlüssel). Bei
// UniqueViolation (Duplikat-Einreichung) wird per bestellungId nachgeschlagen:
// Treffer = idempotente Erfolgsantwort; kein Treffer = echter OCC-Konflikt (409).
// Gleiche ID bedeutet denselben Vorgang — der Payload wird nicht verglichen.
func (c Command) BestellungAufnehmen(ctx context.Context, userID int, userName string, bestellungID string, tischID int, inputs []enrichment.PositionInput, kommentar string) error {
	log := zerolog.Ctx(ctx)

	// Tisch-Existenz, Status prüfen und KS + Subject bestimmen
	subject, kassensitzungNr, _, err := c.loadTischState(ctx, tischID)
	if err != nil {
		return err
	}

	positionen, err := enrichment.EnrichPositionen(ctx, c.ProduktRepo, inputs)
	if err != nil {
		return err
	}

	evt, err := kasse.NewBestellungAufgenommenEvent(subject, userID, userName, bestellungID, positionen, kommentar)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to create bestellung aufgenommen event")
		return err
	}

	druckstationen, err := c.DruckstationRepo.GetKonfigurierteDruckstationen(ctx)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to load druckstationen for arbeitsbon")
		return ErrDatabase
	}

	// Build the Druckaufträge from the stored event (with its generated ID) so the
	// event and its print jobs are written in one transaction (transactional outbox).
	buildAuftraege := func(stored event.Event) []druckauftrag_repo.NeuerDruckauftrag {
		return bondruckApp.CreateArbeitsbonAuftraegeFromEvent(stored, druckstationen)
	}

	// Bestellungen validieren keinen Stream-Zustand (reines Anhängen); die Version wird
	// deshalb erst unmittelbar vor dem Schreiben bestimmt. Bumpt eine Bestellung die
	// Version zwischen Lesen und Schreiben eines validierenden Commands (Zahlung,
	// Storno, …), läuft dieser korrekt in den OCC-Konflikt.
	expectedVersion, err := c.EventRepo.GetMaxVersion(ctx, subject)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to load max version for bestellung")
		return ErrDatabase
	}

	_, err = writeEventWithDruckauftraege(ctx, c.EventRepo, evt, subject, expectedVersion, kasse.StreamTypeTischSession, kassensitzungNr, buildAuftraege)
	if err != nil {
		if errors.Is(err, ErrConflict) {
			// Idempotenz-Check: Ist der Konflikt eine Duplikat-Einreichung (gleiche bestellungId)
			// oder ein echter OCC-Konflikt?
			exists, lookupErr := c.EventRepo.EventExistsByTypeAndVorgangsID(ctx, string(kasse.EventTypeBestellungAufgenommenV1), bestellungID, "bestellungId")
			if lookupErr != nil {
				log.Error().Err(lookupErr).Str("bestellung_id", bestellungID).Msg("Failed to lookup bestellung idempotency")
				return ErrDatabase
			}
			if exists {
				log.Info().Str("bestellung_id", bestellungID).Msg("Idempotente Bestellung: bestellungId bereits vorhanden")
				return nil
			}
			return ErrConflict
		}
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to write bestellung aufgenommen event to database")
		return ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Msg("Bestellung aufgenommen")
	return nil
}

const maxUmbuchungKommentarRunes = 100

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}

	runes := []rune(s)
	if len(runes) <= max {
		return s
	}

	return string(runes[:max])
}

func buildUmbuchungKommentar(prefix string, tischName string) string {
	prefixRunes := len([]rune(prefix))
	if prefixRunes >= maxUmbuchungKommentarRunes {
		return truncateRunes(prefix, maxUmbuchungKommentarRunes)
	}

	maxTischNameRunes := maxUmbuchungKommentarRunes - prefixRunes
	return prefix + truncateRunes(tischName, maxTischNameRunes)
}

func (c Command) BestellungUmbuchen(ctx context.Context, userID int, userName string, quellTischID int, zielTischID int, positionen []kasse.PositionRef, benutzerKommentar string) error {
	log := zerolog.Ctx(ctx)

	if quellTischID == zielTischID {
		log.Warn().Int("tisch_id", quellTischID).Msg("Umbuchung mit gleichem Quell- und Ziel-Tisch")
		return ErrUmbuchungGleicherTisch
	}

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return err
	}

	quellTisch, err := c.TischRepo.GetTable(ctx, quellTischID)
	if err != nil {
		return fromRepositoryError(err, log, quellTischID)
	}
	if quellTisch.Status != tisch.ActiveStatus {
		log.Warn().Int("tisch_id", quellTischID).Str("status", string(quellTisch.Status)).Msg("Quell-Tisch ist nicht aktiv")
		return ErrTischNotActive
	}

	zielTisch, err := c.TischRepo.GetTable(ctx, zielTischID)
	if err != nil {
		return fromRepositoryError(err, log, zielTischID)
	}
	if zielTisch.Status != tisch.ActiveStatus {
		log.Warn().Int("tisch_id", zielTischID).Str("status", string(zielTisch.Status)).Msg("Ziel-Tisch ist nicht aktiv")
		return ErrTischNotActive
	}

	quellSubject := kasse.TischSessionSubject(ks.ZNr, quellTischID)
	zielSubject := kasse.TischSessionSubject(ks.ZNr, zielTischID)

	quellState, err := c.EventRepo.ReadTischSession(ctx, quellSubject)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", quellTischID).Msg("Failed to read source tisch session")
		return ErrDatabase
	}

	if !kasse.ValidatePositionRefs(quellState.UnbezahltePositionen, positionen) {
		log.Warn().Int("quell_tisch_id", quellTischID).Msg("Umbuchungsinvariante verletzt: Positionen nicht umbuchbar")
		return ErrPositionNichtUmbuchbar
	}

	resolvedPositionen, gesamtCents := kasse.ResolvePositionen(quellState.UnbezahltePositionen, positionen)

	quellKommentar := buildUmbuchungKommentar("Umbuchung auf ", zielTisch.Name)
	zielKommentar := buildUmbuchungKommentar("Umbuchung von ", quellTisch.Name)

	quellEvent, zielEvent, err := kasse.NewBestellungUmgebuchtEvents(ks.ZNr, quellTischID, zielTischID, userID, userName, resolvedPositionen, gesamtCents, quellKommentar, zielKommentar, benutzerKommentar)
	if err != nil {
		log.Error().Err(err).Int("quell_tisch_id", quellTischID).Int("ziel_tisch_id", zielTischID).Msg("Failed to create umbuchung events")
		return err
	}

	// Quelle: OCC gegen den validierten Zustand (die Umbuchbarkeit wurde gegen die
	// Quell-Projektion geprüft). Ziel: dort wird kein Zustand validiert (reines
	// Anhängen), die Version kommt erst unmittelbar vor dem Schreiben. Beide Seiten
	// erhalten ihren Signaturauftrag im selben Commit (fiskalische Projektion).
	zielMaxVersion, err := c.EventRepo.GetMaxVersion(ctx, zielSubject)
	if err != nil {
		log.Error().Err(err).Int("ziel_tisch_id", zielTischID).Msg("Failed to load max version for target subject")
		return ErrDatabase
	}

	quellEvent.Version = quellState.LastEventVersion + 1
	zielEvent.Version = zielMaxVersion + 1

	err = c.EventRepo.WriteUmbuchung(ctx, quellEvent, zielEvent, ks.ZNr)
	if err != nil {
		if errors.Is(err, kassenjournal_repo.ErrKassensitzungNichtOffen) {
			log.Warn().Int("quell_tisch_id", quellTischID).Msg("Kassensitzung nicht mehr offen")
			return ErrKasseNichtGeoeffnet
		}
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn().
				Int("quell_version", quellEvent.Version).
				Int("ziel_version", zielEvent.Version).
				Str("quell_subject", quellSubject).
				Str("ziel_subject", zielSubject).
				Msg("OCC conflict bei Bestellung umbuchen")
			return ErrConflict
		}

		log.Error().Err(err).Int("quell_tisch_id", quellTischID).Int("ziel_tisch_id", zielTischID).Msg("Failed to write umbuchung")
		return ErrDatabase
	}

	log.Info().Int("quell_tisch_id", quellTischID).Int("ziel_tisch_id", zielTischID).Msg("Bestellung umgebucht")
	return nil
}

func (c Command) ZahlungKassieren(ctx context.Context, userID int, userName string, tischID int, positionen []kasse.PositionRef, kommentar string) error {
	log := zerolog.Ctx(ctx)

	// Tisch-Existenz, Status und State laden
	subject, kassensitzungNr, state, err := c.loadTischState(ctx, tischID)
	if err != nil {
		return err
	}

	// Bezahl-Invariante: nur unbezahlte Positionen können bezahlt werden
	if !kasse.ValidatePositionRefs(state.UnbezahltePositionen, positionen) {
		log.Warn().Int("tisch_id", tischID).Msg("Bezahl-Invariante verletzt: angeforderte Positionen nicht verfügbar")
		return ErrPositionNichtBezahlbar
	}

	resolvedPositionen, gesamtZahlungCents := kasse.ResolvePositionen(state.UnbezahltePositionen, positionen)

	evt, err := kasse.NewZahlungKassiertEvent(subject, userID, userName, resolvedPositionen, gesamtZahlungCents, kommentar)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to create zahlung kassiert event")
		return err
	}

	// OCC gegen den validierten Zustand: Hat sich der Stream seit dem Lesen geändert
	// (z. B. eine parallele Zahlung), schlägt der Write mit 409 fehl.
	return c.persistTischEvent(ctx, evt, subject, state.LastEventVersion, kassensitzungNr, tischID, "Zahlung kassiert")
}

// StornierungErteilen führt eine „Stornieren"-Aktion aus und teilt sie serverseitig
// nach Bezahlstatus auf: unbezahlte Mengen werden geldneutral korrigiert
// (ein bestellung-korrigiert), bezahlte Mengen werden ihren begleichenden Zahlungen
// FIFO zugeordnet und je Zahlung als kassenwirksame Warenrücknahme zurückgenommen
// (ein stornierung-erteilt mit genau einer ZahlungID). Jedes entstehende Event trägt
// eine eigene TSE-Transaktion; alle werden atomar geschrieben (alles-oder-nichts).
func (c Command) StornierungErteilen(ctx context.Context, userID int, userName string, tischID int, positionen []kasse.PositionRef, kommentar string) error {
	log := zerolog.Ctx(ctx)

	// Tisch-Existenz und Status prüfen, Subject und KS-Nr bestimmen
	subject, kassensitzungNr, _, err := c.loadTischState(ctx, tischID)
	if err != nil {
		return err
	}

	events, err := c.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to read events for stornierung")
		return ErrDatabase
	}

	// Routing nach Bezahlstatus (FIFO je Zahlung). false = angeforderte Menge übersteigt
	// die noch stornierbare Menge.
	aufteilung, ok := kasse.ComputeStornoAufteilung(events, positionen)
	if !ok {
		log.Warn().Int("tisch_id", tischID).Msg("Stornierungsinvariante verletzt: angeforderte Positionen nicht stornierbar")
		return ErrPositionNichtStornierbar
	}

	stornoEvents, err := buildStornoEvents(ctx, subject, userID, userName, aufteilung, kommentar)
	if err != nil {
		return err
	}

	// OCC gegen den validierten Zustand: Basis ist die höchste Version des Replays,
	// gegen den die Storno-Aufteilung berechnet wurde.
	expectedVersion := 0
	if len(events) > 0 {
		expectedVersion = events[len(events)-1].Version
	}

	return c.persistStornoEvents(ctx, stornoEvents, subject, expectedVersion, kassensitzungNr, tischID)
}

// buildStornoEvents erzeugt die Events einer aufgeteilten Storno-Aktion in
// Schreibreihenfolge: zuerst die geldneutrale Korrektur (falls unbezahlte Mengen
// vorliegen), dann je betroffener Zahlung eine kassenwirksame Warenrücknahme.
// Jedes Event erhält beim Schreiben seinen eigenen Signaturauftrag.
func buildStornoEvents(ctx context.Context, subject string, userID int, userName string, aufteilung kasse.StornoAufteilung, kommentar string) ([]event.Event, error) {
	log := zerolog.Ctx(ctx)

	var events []event.Event

	if len(aufteilung.Korrektur) > 0 {
		evt, err := kasse.NewBestellungKorrigiertEvent(subject, userID, userName, aufteilung.Korrektur, aufteilung.KorrekturCents, kommentar)
		if err != nil {
			log.Error().Err(err).Str("subject", subject).Msg("Failed to create bestellung korrigiert event")
			return nil, err
		}
		events = append(events, evt)
	}

	for _, wr := range aufteilung.Warenruecknahmen {
		evt, err := kasse.NewStornierungErteiltEvent(subject, userID, userName, wr.ZahlungID, wr.Positionen, wr.GesamtCents, kommentar)
		if err != nil {
			log.Error().Err(err).Str("subject", subject).Msg("Failed to create stornierung erteilt event")
			return nil, err
		}
		events = append(events, evt)
	}

	return events, nil
}

// persistStornoEvents weist den Storno-Events fortlaufende Versionen ab der
// erwarteten Version (Stand des validierten Replays) zu und schreibt sie atomar
// (je Event mit seinem Signaturauftrag). Ein OCC-Konflikt wird zu ErrConflict.
func (c Command) persistStornoEvents(ctx context.Context, stornoEvents []event.Event, subject string, expectedVersion int, kassensitzungNr int, tischID int) error {
	log := zerolog.Ctx(ctx)

	events := make([]event.Event, 0, len(stornoEvents))
	for i, evt := range stornoEvents {
		evt.Version = expectedVersion + 1 + i
		events = append(events, evt)
	}

	if err := c.EventRepo.WriteTischSessionEventsAtomic(ctx, events, kassensitzungNr); err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn().Int("tisch_id", tischID).Str("subject", subject).Msg("OCC conflict bei Stornierung")
			return ErrConflict
		}
		if errors.Is(err, kassenjournal_repo.ErrKassensitzungNichtOffen) {
			log.Warn().Int("tisch_id", tischID).Msg("Kassensitzung nicht mehr offen")
			return ErrKasseNichtGeoeffnet
		}
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to write stornierung events")
		return ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Int("anzahl_events", len(events)).Msg("Stornierung erteilt")
	return nil
}
