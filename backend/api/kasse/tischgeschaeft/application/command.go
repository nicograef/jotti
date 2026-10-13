package application

import (
	"context"
	"errors"
	"unicode/utf8"

	bondruckApp "github.com/nicograef/jotti/backend/api/druck/bondruck/application"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/druckstation"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/tisch"
	"github.com/nicograef/jotti/backend/internal/logging"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
)

type tischRepo interface {
	GetTisch(ctx context.Context, id int) (tisch.Tisch, error)
	GetAktiveTische(ctx context.Context, kassensitzungNr int) ([]tisch.AktiverTisch, error)
	GetAktiveTischeMitFavoriten(ctx context.Context, userID int, kassensitzungNr int) ([]tisch.AktiverTischMitFavorit, error)
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
	GetAktiveKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
}

type produktRepo interface {
	GetVariantenByIDs(ctx context.Context, ids []int) (map[int]produkt.VarianteMitProdukt, error)
	GetProdukteByIDs(ctx context.Context, ids []int) (map[int]produkt.Produkt, error)
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

// getOffeneKassensitzungOderFehler rejects the booking before any TSE roundtrip when no Kassensitzung is open or the barrier is active.
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

// writeEventOCC writes at expectedVersion+1 and maps the UNIQUE violation to ErrConflict.
// expectedVersion must be the state the command validated against (projection or replay), never a
// fresh GetMaxVersion, or UNIQUE(subject, version) cannot catch a stream change since reading.
func writeEventOCC(ctx context.Context, e event.Event, subject string, expectedVersion int, write func(event.Event) (int, error)) error {
	e.Version = expectedVersion + 1

	if _, err := write(e); err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			logging.FromContext(ctx).Warn("OCC conflict", "version", e.Version, "subject", subject)
			return ErrConflict
		}
		if errors.Is(err, db.ErrConflict) {
			logging.FromContext(ctx).Warn("Deadlock on event write", "subject", subject)
			return ErrConflict
		}
		if errors.Is(err, kassenjournal_repo.ErrKassensitzungNichtOffen) {
			logging.FromContext(ctx).Warn("Kassensitzung nicht mehr offen", "subject", subject)
			return ErrKasseNichtGeoeffnet
		}
		return err
	}

	return nil
}

// persistTischEvent writes the tisch-session event with OCC; aktion is the success log message.
func (c Command) persistTischEvent(ctx context.Context, evt event.Event, subject string, expectedVersion int, kassensitzungNr int, tischID int, aktion string) error {
	log := logging.FromContext(ctx)

	if err := writeEventOCC(ctx, evt, subject, expectedVersion, func(versioned event.Event) (int, error) {
		return c.EventRepo.WriteEvent(ctx, versioned, kasse.StreamTypeTischSession, kassensitzungNr)
	}); err != nil {
		// Beides sind fachliche 409-Antworten (OCC-Konflikt; Kassensitzung zwischen Lesen und
		// Schreiben geschlossen) — im ErrDatabase-Fallback würden sie zu 500.
		if errors.Is(err, ErrConflict) || errors.Is(err, ErrKasseNichtGeoeffnet) {
			return err
		}
		log.Error("Failed to write event to database", "error", err, "tisch_id", tischID)
		return ErrDatabase
	}
	log.Info(aktion, "tisch_id", tischID)
	return nil
}

// loadTischState returns subject, kassensitzungNr, tisch name and the projected TischSession.
func (c Command) loadTischState(ctx context.Context, tischID int) (string, int, string, kasse.TischSession, error) {
	log := logging.FromContext(ctx)

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return "", 0, "", kasse.TischSession{}, err
	}

	t, err := c.TischRepo.GetTisch(ctx, tischID)
	if err != nil {
		return "", 0, "", kasse.TischSession{}, fromRepositoryError(err, log, tischID)
	}

	if t.Status != tisch.ActiveStatus {
		log.Warn("Tisch is not active", "tisch_id", tischID, "status", string(t.Status))
		return "", 0, "", kasse.TischSession{}, ErrTischNotActive
	}

	subject := kasse.TischSessionSubject(ks.ZNr, tischID)

	state, err := c.EventRepo.ReadTischSession(ctx, subject)
	if err != nil {
		log.Error("Failed to read tisch session", "error", err, "tisch_id", tischID)
		return "", 0, "", kasse.TischSession{}, ErrDatabase
	}

	return subject, ks.ZNr, t.Name, state, nil
}

// BestellungAufnehmen is idempotent via the client's bestellungID: on an OCC conflict a stored event
// with that bestellungId means success (same ID = same operation, payload not compared), none means
// a real conflict (409).
func (c Command) BestellungAufnehmen(ctx context.Context, userID int, userName string, bestellungID string, tischID int, inputs []enrichment.PositionInput, kommentar string) error {
	log := logging.FromContext(ctx)

	subject, kassensitzungNr, tischName, _, err := c.loadTischState(ctx, tischID)
	if err != nil {
		return err
	}

	positionen, err := enrichment.EnrichPositionen(ctx, c.ProduktRepo, inputs)
	if err != nil {
		return err
	}

	evt, err := kasse.NewBestellungAufgenommenEvent(subject, userID, userName, bestellungID, positionen, kommentar)
	if err != nil {
		log.Error("Failed to create bestellung aufgenommen event", "error", err, "tisch_id", tischID)
		return err
	}

	druckstationen, err := c.DruckstationRepo.GetKonfigurierteDruckstationen(ctx)
	if err != nil {
		log.Error("Failed to load druckstationen for arbeitsbon", "error", err, "tisch_id", tischID)
		return ErrDatabase
	}

	buildAuftraege := func(stored event.Event) []druckauftrag_repo.NeuerDruckauftrag {
		return bondruckApp.CreateArbeitsbonAuftraegeFromEvent(stored, druckstationen, tischName)
	}

	// Bestellungen validieren keinen Stream-Zustand (reines Anhängen); die Version wird deshalb erst
	// unmittelbar vor dem Schreiben bestimmt. Bumpt sie dabei die Version eines parallel
	// validierenden Commands (Zahlung, Storno), läuft dieser korrekt in den OCC-Konflikt.
	expectedVersion, err := c.EventRepo.GetMaxVersion(ctx, subject)
	if err != nil {
		log.Error("Failed to load max version for bestellung", "error", err, "tisch_id", tischID)
		return ErrDatabase
	}

	// Writes the event and the Druckaufträge built from it in one transaction
	// (transactional outbox); buildAuftraege sees the stored event with its generated ID.
	err = writeEventOCC(ctx, evt, subject, expectedVersion, func(versioned event.Event) (int, error) {
		return c.EventRepo.WriteEventWithDruckauftraege(ctx, versioned, kasse.StreamTypeTischSession, kassensitzungNr, buildAuftraege)
	})
	if err != nil {
		if errors.Is(err, ErrConflict) {
			exists, lookupErr := c.EventRepo.EventExistsByTypeAndVorgangsID(ctx, string(kasse.EventTypeBestellungAufgenommenV1), bestellungID, "bestellungId")
			if lookupErr != nil {
				log.Error("Failed to lookup bestellung idempotency", "error", lookupErr, "bestellung_id", bestellungID)
				return ErrDatabase
			}
			if exists {
				log.Info("Idempotente Bestellung: bestellungId bereits vorhanden", "bestellung_id", bestellungID)
				return nil
			}
			return ErrConflict
		}
		// Kassensitzung zwischen Lesen und Schreiben geschlossen — 409, kein Datenbankfehler.
		if errors.Is(err, ErrKasseNichtGeoeffnet) {
			return err
		}
		log.Error("Failed to write bestellung aufgenommen event to database", "error", err, "tisch_id", tischID)
		return ErrDatabase
	}

	log.Info("Bestellung aufgenommen", "tisch_id", tischID)
	return nil
}

// maxUmbuchungKommentarBytes ist die Grenze des Kommentars im eingefrorenen
// Event-Schema (bestellungUmgebuchtV1DataSchema); zogs Max zählt Bytes.
const maxUmbuchungKommentarBytes = 100

// truncateBytes kürzt auf höchstens maxBytes und rückt bis zum Anfang der angeschnittenen UTF-8-Folge zurück, damit kein Umlaut zerfällt.
func truncateBytes(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}

	cut := maxBytes
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}

	return s[:cut]
}

// buildUmbuchungKommentar: Tischname (bis 100 Bytes) plus Präfix reißt die Schemagrenze — gekürzt
// wird der Name, das Präfix bleibt ganz.
func buildUmbuchungKommentar(prefix string, tischName string) string {
	return prefix + truncateBytes(tischName, maxUmbuchungKommentarBytes-len(prefix))
}

func (c Command) BestellungUmbuchen(ctx context.Context, userID int, userName string, quellTischID int, zielTischID int, positionen []kasse.PositionRef, benutzerKommentar string) error {
	log := logging.FromContext(ctx)

	if quellTischID == zielTischID {
		log.Warn("Umbuchung mit gleichem Quell- und Ziel-Tisch", "tisch_id", quellTischID)
		return ErrUmbuchungGleicherTisch
	}

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return err
	}

	quellTisch, err := c.TischRepo.GetTisch(ctx, quellTischID)
	if err != nil {
		return fromRepositoryError(err, log, quellTischID)
	}
	if quellTisch.Status != tisch.ActiveStatus {
		log.Warn("Quell-Tisch ist nicht aktiv", "tisch_id", quellTischID, "status", string(quellTisch.Status))
		return ErrTischNotActive
	}

	zielTisch, err := c.TischRepo.GetTisch(ctx, zielTischID)
	if err != nil {
		return fromRepositoryError(err, log, zielTischID)
	}
	if zielTisch.Status != tisch.ActiveStatus {
		log.Warn("Ziel-Tisch ist nicht aktiv", "tisch_id", zielTischID, "status", string(zielTisch.Status))
		return ErrTischNotActive
	}

	quellSubject := kasse.TischSessionSubject(ks.ZNr, quellTischID)
	zielSubject := kasse.TischSessionSubject(ks.ZNr, zielTischID)

	quellState, err := c.EventRepo.ReadTischSession(ctx, quellSubject)
	if err != nil {
		log.Error("Failed to read source tisch session", "error", err, "tisch_id", quellTischID)
		return ErrDatabase
	}

	if !kasse.ValidatePositionRefs(quellState.UnbezahltePositionen, positionen) {
		log.Warn("Umbuchungsinvariante verletzt: Positionen nicht umbuchbar", "quell_tisch_id", quellTischID)
		return ErrPositionNichtUmbuchbar
	}

	resolvedPositionen, gesamtCents := kasse.ResolvePositionen(quellState.UnbezahltePositionen, positionen)

	quellKommentar := buildUmbuchungKommentar("Umbuchung auf ", zielTisch.Name)
	zielKommentar := buildUmbuchungKommentar("Umbuchung von ", quellTisch.Name)

	quellEvent, zielEvent, err := kasse.NewBestellungUmgebuchtEvents(ks.ZNr, quellTischID, zielTischID, userID, userName, resolvedPositionen, gesamtCents, quellKommentar, zielKommentar, benutzerKommentar)
	if err != nil {
		log.Error("Failed to create umbuchung events", "error", err, "quell_tisch_id", quellTischID, "ziel_tisch_id", zielTischID)
		return err
	}

	// The source uses OCC against its validated state; the target validates nothing (pure append), so
	// its version is read right before writing. Both sides get their Signaturauftrag in the same commit.
	zielMaxVersion, err := c.EventRepo.GetMaxVersion(ctx, zielSubject)
	if err != nil {
		log.Error("Failed to load max version for target subject", "error", err, "ziel_tisch_id", zielTischID)
		return ErrDatabase
	}

	quellEvent.Version = quellState.LastEventVersion + 1
	zielEvent.Version = zielMaxVersion + 1

	err = c.EventRepo.WriteUmbuchung(ctx, quellEvent, zielEvent, ks.ZNr)
	if err != nil {
		if errors.Is(err, kassenjournal_repo.ErrKassensitzungNichtOffen) {
			log.Warn("Kassensitzung nicht mehr offen", "quell_tisch_id", quellTischID)
			return ErrKasseNichtGeoeffnet
		}
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn("OCC conflict bei Bestellung umbuchen", "quell_version", quellEvent.Version, "ziel_version", zielEvent.Version, "quell_subject", quellSubject, "ziel_subject", zielSubject)
			return ErrConflict
		}

		log.Error("Failed to write umbuchung", "error", err, "quell_tisch_id", quellTischID, "ziel_tisch_id", zielTischID)
		return ErrDatabase
	}

	log.Info("Bestellung umgebucht", "quell_tisch_id", quellTischID, "ziel_tisch_id", zielTischID)
	return nil
}

func (c Command) ZahlungKassieren(ctx context.Context, userID int, userName string, tischID int, positionen []kasse.PositionRef, kommentar string) error {
	log := logging.FromContext(ctx)

	subject, kassensitzungNr, _, state, err := c.loadTischState(ctx, tischID)
	if err != nil {
		return err
	}

	if !kasse.ValidatePositionRefs(state.UnbezahltePositionen, positionen) {
		log.Warn("Bezahl-Invariante verletzt: angeforderte Positionen nicht verfügbar", "tisch_id", tischID)
		return ErrPositionNichtBezahlbar
	}

	resolvedPositionen, gesamtZahlungCents := kasse.ResolvePositionen(state.UnbezahltePositionen, positionen)

	evt, err := kasse.NewZahlungKassiertEvent(subject, userID, userName, resolvedPositionen, gesamtZahlungCents, kommentar)
	if err != nil {
		log.Error("Failed to create zahlung kassiert event", "error", err, "tisch_id", tischID)
		return err
	}

	return c.persistTischEvent(ctx, evt, subject, state.LastEventVersion, kassensitzungNr, tischID, "Zahlung kassiert")
}

// StornierungErteilen splits by payment status: unpaid quantities become one bestellung-korrigiert,
// paid ones one stornierung-erteilt per Zahlung (FIFO). Each event gets its own TSE transaction, all
// are written atomically (docs/handbuch.md §3.7).
func (c Command) StornierungErteilen(ctx context.Context, userID int, userName string, tischID int, positionen []kasse.PositionRef, kommentar string) error {
	log := logging.FromContext(ctx)

	subject, kassensitzungNr, _, _, err := c.loadTischState(ctx, tischID)
	if err != nil {
		return err
	}

	events, err := c.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		log.Error("Failed to read events for stornierung", "error", err, "tisch_id", tischID)
		return ErrDatabase
	}

	// false = angeforderte Positionen sind nicht stornierbar: die Menge übersteigt die noch
	// stornierbare Menge, oder eine PositionID ist doppelt referenziert.
	aufteilung, ok := kasse.ComputeStornoAufteilung(events, positionen)
	if !ok {
		log.Warn("Stornierungsinvariante verletzt: angeforderte Positionen nicht stornierbar", "tisch_id", tischID)
		return ErrPositionNichtStornierbar
	}

	stornoEvents, err := buildStornoEvents(ctx, subject, userID, userName, aufteilung, kommentar)
	if err != nil {
		return err
	}

	expectedVersion := 0
	if len(events) > 0 {
		expectedVersion = events[len(events)-1].Version
	}

	return c.persistStornoEvents(ctx, stornoEvents, subject, expectedVersion, kassensitzungNr, tischID)
}

func buildStornoEvents(ctx context.Context, subject string, userID int, userName string, aufteilung kasse.StornoAufteilung, kommentar string) ([]event.Event, error) {
	log := logging.FromContext(ctx)

	var events []event.Event

	if len(aufteilung.Korrektur) > 0 {
		evt, err := kasse.NewBestellungKorrigiertEvent(subject, userID, userName, aufteilung.Korrektur, aufteilung.KorrekturCents, kommentar)
		if err != nil {
			log.Error("Failed to create bestellung korrigiert event", "error", err, "subject", subject)
			return nil, err
		}
		events = append(events, evt)
	}

	for _, wr := range aufteilung.Warenruecknahmen {
		evt, err := kasse.NewStornierungErteiltEvent(subject, userID, userName, wr.ZahlungID, wr.Positionen, wr.GesamtCents, kommentar)
		if err != nil {
			log.Error("Failed to create stornierung erteilt event", "error", err, "subject", subject)
			return nil, err
		}
		events = append(events, evt)
	}

	return events, nil
}

func (c Command) persistStornoEvents(ctx context.Context, stornoEvents []event.Event, subject string, expectedVersion int, kassensitzungNr int, tischID int) error {
	log := logging.FromContext(ctx)

	events := make([]event.Event, 0, len(stornoEvents))
	for i, evt := range stornoEvents {
		evt.Version = expectedVersion + 1 + i
		events = append(events, evt)
	}

	if err := c.EventRepo.WriteTischSessionEventsAtomic(ctx, events, kassensitzungNr); err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn("OCC conflict bei Stornierung", "tisch_id", tischID, "subject", subject)
			return ErrConflict
		}
		if errors.Is(err, kassenjournal_repo.ErrKassensitzungNichtOffen) {
			log.Warn("Kassensitzung nicht mehr offen", "tisch_id", tischID)
			return ErrKasseNichtGeoeffnet
		}
		log.Error("Failed to write stornierung events", "error", err, "tisch_id", tischID)
		return ErrDatabase
	}

	log.Info("Stornierung erteilt", "tisch_id", tischID, "anzahl_events", len(events))
	return nil
}
