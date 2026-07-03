package application

import (
	"context"
	"errors"

	bondruckApp "github.com/nicograef/jotti/backend/api/bondruck/application"
	tseApp "github.com/nicograef/jotti/backend/api/tse/application"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/nicograef/jotti/backend/domain/table"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/rs/zerolog"
)

type tableRepo interface {
	GetTable(ctx context.Context, id int) (table.Tisch, error)
	CreateTable(ctx context.Context, t table.Tisch) (int, error)
	UpdateTable(ctx context.Context, t table.Tisch) error
	GetAllTables(ctx context.Context) ([]table.Tisch, error)
	GetActiveTables(ctx context.Context, kassensitzungNr int) ([]table.AktiverTisch, error)
	GetActiveTablesWithFavorites(ctx context.Context, userID int, kassensitzungNr int) ([]table.AktiverTischMitFavorit, error)
}

type eventRepo interface {
	WriteEvent(ctx context.Context, e event.Event, streamType kasse.StreamType, kassensitzungNr int) (int, error)
	WriteEventWithDruckauftraege(ctx context.Context, e event.Event, streamType kasse.StreamType, kassensitzungNr int, buildAuftraege func(event.Event) []druckauftrag_repo.NeuerDruckauftrag) (int, error)
	WriteEventWithDruckauftraegeUndNachsignierAuftrag(ctx context.Context, e event.Event, streamType kasse.StreamType, kassensitzungNr int, buildAuftraege func(event.Event) []druckauftrag_repo.NeuerDruckauftrag, txID string, processType string, processData string) (int, error)
	WriteEventWithNachsignierAuftrag(ctx context.Context, e event.Event, streamType kasse.StreamType, kassensitzungNr int, txID string, processType string, processData string) (int, error)
	WriteUmbuchung(ctx context.Context, quellEvent event.Event, zielEvent event.Event, nachsignierungen []kassenjournal_repo.TSENachsignierung, kassensitzungNr int) error
	WriteTischSessionEventsAtomic(ctx context.Context, events []event.Event, nachsignierungen []kassenjournal_repo.TSENachsignierung, kassensitzungNr int) error
	ReadTischSession(ctx context.Context, subject string) (kasse.TischSession, error)
	GetMaxVersion(ctx context.Context, subject string) (int, error)
	ReadEventsBySubject(ctx context.Context, subject string) ([]event.Event, error)
	GetTSESignaturByTxID(ctx context.Context, txID string) (kasse.TSEData, error)
}

type kassensitzungenRepo interface {
	GetOffeneKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
	GetAktiveKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
}

type productRepo interface {
	GetProduct(ctx context.Context, productID int) (product.Produkt, error)
	GetVariant(ctx context.Context, variantID int) (product.Variante, error)
	GetVariantsByIDs(ctx context.Context, ids []int) (map[int]product.Variante, error)
	GetProductsByIDs(ctx context.Context, ids []int) (map[int]product.Produkt, error)
}

type favoritRepo interface {
	Add(ctx context.Context, userID, tischID int) error
	Remove(ctx context.Context, userID, tischID int) error
	GetByUser(ctx context.Context, userID int) ([]int, error)
}

type druckstationRepo interface {
	GetKonfigurierteDruckstationen(ctx context.Context) (map[string]bondruckApp.Druckstation, error)
}

type druckauftragRepo interface {
	EnqueueDruckauftraege(ctx context.Context, auftraege []druckauftrag_repo.NeuerDruckauftrag) error
}

type settingsRepo interface {
	GetBetreiber(ctx context.Context) (settings.Betreiber, error)
	GetKassenidentitaet(ctx context.Context) (settings.Kassenidentitaet, error)
}

// BestellPositionInput represents the input for a single position in an order.
// The application layer enriches this with product/variant details (fat events).
type BestellPositionInput struct {
	ProduktID  int
	VarianteID int
	Menge      int
}

type Command struct {
	TableRepo           tableRepo
	EventRepo           eventRepo
	ProductRepo         productRepo
	FavoritRepo         favoritRepo
	KassensitzungenRepo kassensitzungenRepo
	DruckstationRepo    druckstationRepo
	DruckauftragRepo    druckauftragRepo
	SettingsRepo        settingsRepo
	TSESignierer        tseApp.Signierer
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
// Lesen geändert hat (z. B. während der TSE-Signierung), und verhindert Doppel-Writes
// auf Basis veralteter Validierung.
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

// writeEvent writes an event with optimistic concurrency control against expectedVersion.
// Returns ErrConflict on a version conflict.
func writeEvent(ctx context.Context, repo eventRepo, e event.Event, subject string, expectedVersion int, streamType kasse.StreamType, kassensitzungNr int) (int, error) {
	return writeEventOCC(ctx, e, subject, expectedVersion, func(versioned event.Event) (int, error) {
		return repo.WriteEvent(ctx, versioned, streamType, kassensitzungNr)
	})
}

// writeEventWithDruckauftraege writes an event and the print jobs derived from it
// (built from the stored event including its generated ID) in a single transaction.
// Returns ErrConflict on a version conflict.
func writeEventWithDruckauftraege(ctx context.Context, repo eventRepo, e event.Event, subject string, expectedVersion int, streamType kasse.StreamType, kassensitzungNr int, buildAuftraege func(event.Event) []druckauftrag_repo.NeuerDruckauftrag) (int, error) {
	return writeEventOCC(ctx, e, subject, expectedVersion, func(versioned event.Event) (int, error) {
		return repo.WriteEventWithDruckauftraege(ctx, versioned, streamType, kassensitzungNr, buildAuftraege)
	})
}

func writeEventWithDruckauftraegeUndNachsignierAuftrag(ctx context.Context, repo eventRepo, e event.Event, subject string, expectedVersion int, streamType kasse.StreamType, kassensitzungNr int, buildAuftraege func(event.Event) []druckauftrag_repo.NeuerDruckauftrag, txID string, processType string, processData string) (int, error) {
	return writeEventOCC(ctx, e, subject, expectedVersion, func(versioned event.Event) (int, error) {
		return repo.WriteEventWithDruckauftraegeUndNachsignierAuftrag(ctx, versioned, streamType, kassensitzungNr, buildAuftraege, txID, processType, processData)
	})
}

// writeEventWithNachsignierAuftrag writes an event plus a TSE retry job in one
// transaction. Returns ErrConflict on a version conflict.
func writeEventWithNachsignierAuftrag(ctx context.Context, repo eventRepo, e event.Event, subject string, expectedVersion int, streamType kasse.StreamType, kassensitzungNr int, txID string, processType string, processData string) (int, error) {
	return writeEventOCC(ctx, e, subject, expectedVersion, func(versioned event.Event) (int, error) {
		return repo.WriteEventWithNachsignierAuftrag(ctx, versioned, streamType, kassensitzungNr, txID, processType, processData)
	})
}

// persistSignedTischEvent writes a signed tisch-session event: when the signing produced a
// Nachsignier-Auftrag the event is written together with that TSE retry job, otherwise on its own.
// expectedVersion ist die Version des gelesenen Zustands, gegen den validiert wurde.
// An OCC conflict maps to ErrConflict, any other write error to ErrDatabase. aktion is the success
// log message; on the deferred-signing path it is suffixed accordingly.
func (c Command) persistSignedTischEvent(ctx context.Context, signierung tseApp.Signierung, subject string, expectedVersion int, kassensitzungNr int, tischID int, aktion string) error {
	log := zerolog.Ctx(ctx)

	if signierung.NachsignierAuftrag != nil {
		na := signierung.NachsignierAuftrag
		if _, err := writeEventWithNachsignierAuftrag(ctx, c.EventRepo, signierung.Event, subject, expectedVersion, kasse.StreamTypeTischSession, kassensitzungNr, na.TxID, na.ProcessType, na.ProcessData); err != nil {
			if errors.Is(err, ErrConflict) {
				return ErrConflict
			}
			log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to write event with TSE-nachsignierung")
			return ErrDatabase
		}
		log.Info().Int("tisch_id", tischID).Msg(aktion + " (unsigniert, Nachsignierung vorgemerkt)")
		return nil
	}

	if _, err := writeEvent(ctx, c.EventRepo, signierung.Event, subject, expectedVersion, kasse.StreamTypeTischSession, kassensitzungNr); err != nil {
		if errors.Is(err, ErrConflict) {
			return ErrConflict
		}
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to write event to database")
		return ErrDatabase
	}
	log.Info().Int("tisch_id", tischID).Msg(aktion)
	return nil
}

// konfigurierteDruckstationen returns the configured work-ticket printers, or an
// empty map when no DruckstationRepo is wired (e.g. in tests).
func (c Command) konfigurierteDruckstationen(ctx context.Context) (map[string]bondruckApp.Druckstation, error) {
	if c.DruckstationRepo == nil {
		return nil, nil
	}
	return c.DruckstationRepo.GetKonfigurierteDruckstationen(ctx)
}

// loadTischState loads and validates the tisch, then reads its projected tisch session.
// Returns the subject, kassensitzungNr, and TischSession state.
// Returns ErrKasseNichtGeoeffnet if no open Kassensitzung exists.
// Returns ErrTischNotFound if the tisch doesn't exist, ErrTischNotActive if not active.
func (c Command) loadTischState(ctx context.Context, tischID int) (string, int, kasse.TischSession, error) {
	log := zerolog.Ctx(ctx)

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return "", 0, kasse.TischSession{}, err
	}

	tisch, err := c.TableRepo.GetTable(ctx, tischID)
	if err != nil {
		return "", 0, kasse.TischSession{}, fromRepositoryError(err, log, tischID)
	}

	if tisch.Status != table.ActiveStatus {
		log.Warn().Int("tisch_id", tischID).Str("status", string(tisch.Status)).Msg("Tisch is not active")
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

// validatePositionRefs checks that every requested PositionRef exists in the available positions,
// that no PositionID is referenced more than once (duplicates would add up unnoticed), and that
// the requested Menge does not exceed the available Menge.
func validatePositionRefs(available []kasse.Position, requested []kasse.PositionRef) bool {
	seen := make(map[string]bool, len(requested))
	for _, ref := range requested {
		if seen[ref.PositionID] {
			return false
		}
		seen[ref.PositionID] = true
		found := false
		for _, pos := range available {
			if pos.PositionID == ref.PositionID {
				if ref.Menge > pos.Menge {
					return false
				}
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (c Command) FavoritHinzufuegen(ctx context.Context, userID, tischID int) error {
	log := zerolog.Ctx(ctx)

	tisch, err := c.TableRepo.GetTable(ctx, tischID)
	if err != nil {
		return fromRepositoryError(err, log, tischID)
	}

	if tisch.Status != table.ActiveStatus {
		log.Warn().Int("tisch_id", tischID).Str("status", string(tisch.Status)).Msg("Tisch is not active")
		return ErrTischNotActive
	}

	if err := c.FavoritRepo.Add(ctx, userID, tischID); err != nil {
		log.Error().Err(err).Int("user_id", userID).Int("tisch_id", tischID).Msg("Failed to add favorit")
		return ErrDatabase
	}

	log.Info().Int("user_id", userID).Int("tisch_id", tischID).Msg("Favorit added")
	return nil
}

func (c Command) FavoritEntfernen(ctx context.Context, userID, tischID int) error {
	log := zerolog.Ctx(ctx)

	if err := c.FavoritRepo.Remove(ctx, userID, tischID); err != nil {
		log.Error().Err(err).Int("user_id", userID).Int("tisch_id", tischID).Msg("Failed to remove favorit")
		return ErrDatabase
	}

	log.Info().Int("user_id", userID).Int("tisch_id", tischID).Msg("Favorit removed")
	return nil
}

func (c Command) TischErstellen(ctx context.Context, name string) (int, error) {
	log := zerolog.Ctx(ctx)

	tisch, err := table.NewTisch(name)
	if err != nil {
		log.Warn().Err(err).Str("tisch_name", name).Msg("Invalid tisch data")
		return 0, ErrInvalidTischData
	}

	id, err := c.TableRepo.CreateTable(ctx, tisch)
	if err != nil {
		return 0, fromRepositoryError(err, log, 0)
	}

	log.Info().Int("tisch_id", id).Msg("Tisch created")
	return id, nil
}

func (c Command) TischAktualisieren(ctx context.Context, id int, name string) error {
	log := zerolog.Ctx(ctx)

	tisch, err := c.TableRepo.GetTable(ctx, id)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	err = tisch.Rename(name)
	if err != nil {
		log.Warn().Err(err).Int("tisch_id", id).Msg("Invalid tisch data for update")
		return ErrInvalidTischData
	}

	err = c.TableRepo.UpdateTable(ctx, tisch)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	log.Info().Int("tisch_id", id).Msg("Tisch updated")
	return nil
}

func (c Command) TischAktivieren(ctx context.Context, id int) error {
	return c.applyTischStatusChange(ctx, id, "Tisch activated", func(t *table.Tisch) { t.Activate() })
}

func (c Command) TischDeaktivieren(ctx context.Context, id int) error {
	return c.applyTischStatusChange(ctx, id, "Tisch deactivated", func(t *table.Tisch) { t.Deactivate() })
}

func (c Command) TischLoeschen(ctx context.Context, id int) error {
	return c.applyTischStatusChange(ctx, id, "Tisch deleted", func(t *table.Tisch) { t.Delete() })
}

func (c Command) applyTischStatusChange(ctx context.Context, id int, successMsg string, action func(*table.Tisch)) error {
	log := zerolog.Ctx(ctx)

	tisch, err := c.TableRepo.GetTable(ctx, id)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}
	action(&tisch)
	if err := c.TableRepo.UpdateTable(ctx, tisch); err != nil {
		return fromRepositoryError(err, log, id)
	}
	log.Info().Int("tisch_id", id).Msg(successMsg)
	return nil
}

func (c Command) BestellungAufnehmen(ctx context.Context, userID int, userName string, tischID int, inputs []BestellPositionInput, kommentar string) error {
	log := zerolog.Ctx(ctx)

	// Tisch-Existenz, Status prüfen und KS + Subject bestimmen
	subject, kassensitzungNr, _, err := c.loadTischState(ctx, tischID)
	if err != nil {
		return err
	}

	// Collect unique variant and product IDs for batch enrichment
	varianteIDs := make([]int, 0, len(inputs))
	produktIDs := make([]int, 0, len(inputs))
	seenVarianten := make(map[int]bool, len(inputs))
	seenProdukte := make(map[int]bool, len(inputs))
	for _, input := range inputs {
		if !seenVarianten[input.VarianteID] {
			varianteIDs = append(varianteIDs, input.VarianteID)
			seenVarianten[input.VarianteID] = true
		}
		if !seenProdukte[input.ProduktID] {
			produktIDs = append(produktIDs, input.ProduktID)
			seenProdukte[input.ProduktID] = true
		}
	}

	// Batch-fetch all required variants and products in one query each
	variantenByID, err := c.ProductRepo.GetVariantsByIDs(ctx, varianteIDs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to batch-fetch variants for position enrichment")
		return ErrProduktNotFound
	}
	produkteByID, err := c.ProductRepo.GetProductsByIDs(ctx, produktIDs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to batch-fetch products for position enrichment")
		return ErrProduktNotFound
	}

	// Enrich positions with product/variant data (fat events)
	positionen := make([]kasse.Position, 0, len(inputs))
	for _, input := range inputs {
		variant, ok := variantenByID[input.VarianteID]
		if !ok {
			log.Error().Int("variante_id", input.VarianteID).Msg("Variant not found in batch result")
			return ErrProduktNotFound
		}
		prod, ok := produkteByID[input.ProduktID]
		if !ok {
			log.Error().Int("produkt_id", input.ProduktID).Msg("Product not found in batch result")
			return ErrProduktNotFound
		}

		positionen = append(positionen, kasse.Position{
			VarianteID:   input.VarianteID,
			ProduktName:  prod.Name,
			VarianteName: variant.Name,
			Kategorie:    string(prod.Kategorie),
			Steuersatz:   string(prod.Steuersatz),
			Einzelpreis:  variant.PreisCents,
			Menge:        input.Menge,
		})
	}

	evt, err := kasse.NewBestellungAufgenommenEvent(subject, userID, userName, positionen, kommentar)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to create bestellung aufgenommen event")
		return err
	}

	signierung, err := c.signBestellungAufgenommenEvent(ctx, evt, positionen)
	if err != nil {
		return err
	}
	evt = signierung.Event

	druckstationen, err := c.konfigurierteDruckstationen(ctx)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to load druckstationen for arbeitsbon")
		return ErrDatabase
	}

	// Build the work tickets from the stored event (with its generated ID) so the
	// event and its print jobs are written in one transaction (transactional outbox).
	buildAuftraege := func(stored event.Event) []druckauftrag_repo.NeuerDruckauftrag {
		return bondruckApp.CreateArbeitsbonAuftraegeFromEvent(stored, druckstationen)
	}

	// Bestellungen validieren keinen Stream-Zustand (reines Anhängen); die Version wird
	// deshalb erst unmittelbar vor dem Schreiben bestimmt, damit parallele Bestellungen
	// am selben Tisch nicht über die gesamte Signier-Dauer hinweg kollidieren. Bumpt eine
	// Bestellung die Version zwischen Lesen und Schreiben eines validierenden Commands
	// (Zahlung, Storno, …), läuft dieser korrekt in den OCC-Konflikt.
	expectedVersion, err := c.EventRepo.GetMaxVersion(ctx, subject)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to load max version for bestellung")
		return ErrDatabase
	}

	if signierung.NachsignierAuftrag != nil {
		if _, err = writeEventWithDruckauftraegeUndNachsignierAuftrag(
			ctx,
			c.EventRepo,
			evt,
			subject,
			expectedVersion,
			kasse.StreamTypeTischSession,
			kassensitzungNr,
			buildAuftraege,
			signierung.NachsignierAuftrag.TxID,
			signierung.NachsignierAuftrag.ProcessType,
			signierung.NachsignierAuftrag.ProcessData,
		); err != nil {
			if errors.Is(err, ErrConflict) {
				return ErrConflict
			}
			log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to write bestellung event with TSE-nachsignierung")
			return ErrDatabase
		}

		log.Info().Int("tisch_id", tischID).Msg("Bestellung aufgenommen (unsigniert, Nachsignierung vorgemerkt)")
		return nil
	}

	_, err = writeEventWithDruckauftraege(ctx, c.EventRepo, evt, subject, expectedVersion, kasse.StreamTypeTischSession, kassensitzungNr, buildAuftraege)
	if err != nil {
		if errors.Is(err, ErrConflict) {
			return ErrConflict
		}
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to write bestellung aufgenommen event to database")
		return ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Msg("Bestellung aufgenommen")
	return nil
}

// resolvePositions resolves PositionRefs to full Positions using available positions.
// Returns resolved positions and total amount in cents.
func resolvePositions(available []kasse.Position, refs []kasse.PositionRef) ([]kasse.Position, int) {
	resolved := make([]kasse.Position, 0, len(refs))
	totalCents := 0
	for _, ref := range refs {
		for _, pos := range available {
			if pos.PositionID == ref.PositionID {
				resolved = append(resolved, kasse.Position{
					PositionID:   pos.PositionID,
					VarianteID:   pos.VarianteID,
					ProduktName:  pos.ProduktName,
					VarianteName: pos.VarianteName,
					Kategorie:    pos.Kategorie,
					Steuersatz:   pos.Steuersatz,
					Einzelpreis:  pos.Einzelpreis,
					Menge:        ref.Menge,
				})
				totalCents += pos.Einzelpreis * ref.Menge
				break
			}
		}
	}
	return resolved, totalCents
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

func (c Command) BestellungUmbuchen(ctx context.Context, userID int, userName string, quellTischID int, zielTischID int, positionen []kasse.PositionRef) error {
	log := zerolog.Ctx(ctx)

	if quellTischID == zielTischID {
		log.Warn().Int("tisch_id", quellTischID).Msg("Umbuchung mit gleichem Quell- und Ziel-Tisch")
		return ErrUmbuchungGleicherTisch
	}

	ks, err := c.getOffeneKassensitzungOderFehler(ctx)
	if err != nil {
		return err
	}

	quellTisch, err := c.TableRepo.GetTable(ctx, quellTischID)
	if err != nil {
		return fromRepositoryError(err, log, quellTischID)
	}
	if quellTisch.Status != table.ActiveStatus {
		log.Warn().Int("tisch_id", quellTischID).Str("status", string(quellTisch.Status)).Msg("Quell-Tisch ist nicht aktiv")
		return ErrTischNotActive
	}

	zielTisch, err := c.TableRepo.GetTable(ctx, zielTischID)
	if err != nil {
		return fromRepositoryError(err, log, zielTischID)
	}
	if zielTisch.Status != table.ActiveStatus {
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

	if !validatePositionRefs(quellState.UnbezahltePositionen, positionen) {
		log.Warn().Int("quell_tisch_id", quellTischID).Msg("Umbuchungsinvariante verletzt: Positionen nicht umbuchbar")
		return ErrPositionNichtUmbuchbar
	}

	resolvedPositionen, gesamtCents := resolvePositions(quellState.UnbezahltePositionen, positionen)

	quellKommentar := buildUmbuchungKommentar("Umbuchung auf Tisch ", zielTisch.Name)
	zielKommentar := buildUmbuchungKommentar("Umbuchung von Tisch ", quellTisch.Name)

	quellEvent, zielEvent, err := kasse.NewBestellungUmgebuchtEvents(ks.ZNr, quellTischID, zielTischID, userID, userName, resolvedPositionen, gesamtCents, quellKommentar, zielKommentar)
	if err != nil {
		log.Error().Err(err).Int("quell_tisch_id", quellTischID).Int("ziel_tisch_id", zielTischID).Msg("Failed to create umbuchung events")
		return err
	}

	// Beide Seiten sind geldneutrale Bestellungen und werden je mit eigener
	// TSE-Transaktion als Bestellung-V1 signiert: der Abgang mit negativen,
	// der Zugang mit positiven Mengen (Anhang I).
	quellSignierung, err := c.signBestellungUmgebuchtEvent(ctx, quellEvent, resolvedPositionen, -1)
	if err != nil {
		return err
	}
	zielSignierung, err := c.signBestellungUmgebuchtEvent(ctx, zielEvent, resolvedPositionen, +1)
	if err != nil {
		return err
	}

	// Quelle: OCC gegen den validierten Zustand (die Umbuchbarkeit wurde gegen die
	// Quell-Projektion geprüft). Ziel: dort wird kein Zustand validiert (reines
	// Anhängen), die Version kommt erst unmittelbar vor dem Schreiben.
	zielMaxVersion, err := c.EventRepo.GetMaxVersion(ctx, zielSubject)
	if err != nil {
		log.Error().Err(err).Int("ziel_tisch_id", zielTischID).Msg("Failed to load max version for target subject")
		return ErrDatabase
	}

	quellSigniert := quellSignierung.Event
	zielSigniert := zielSignierung.Event
	quellSigniert.Version = quellState.LastEventVersion + 1
	zielSigniert.Version = zielMaxVersion + 1

	nachsignierungen := nachsignierungenAusSignierungen(quellSignierung, zielSignierung)

	err = c.EventRepo.WriteUmbuchung(ctx, quellSigniert, zielSigniert, nachsignierungen, ks.ZNr)
	if err != nil {
		if errors.Is(err, kassenjournal_repo.ErrKassensitzungNichtOffen) {
			log.Warn().Int("quell_tisch_id", quellTischID).Msg("Kassensitzung nicht mehr offen")
			return ErrKasseNichtGeoeffnet
		}
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn().
				Int("quell_version", quellSigniert.Version).
				Int("ziel_version", zielSigniert.Version).
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

// nachsignierungenAusSignierungen sammelt die Nachsignier-Aufträge der Seiten, deren
// Signierung bei der Erfassung fehlschlug (TSE-Ausfall), damit der Worker sie atomar
// mit den Events nachholt.
func nachsignierungenAusSignierungen(signierungen ...tseApp.Signierung) []kassenjournal_repo.TSENachsignierung {
	var nachsignierungen []kassenjournal_repo.TSENachsignierung
	for _, s := range signierungen {
		if s.NachsignierAuftrag == nil {
			continue
		}
		nachsignierungen = append(nachsignierungen, kassenjournal_repo.TSENachsignierung{
			TxID:        s.NachsignierAuftrag.TxID,
			ProcessType: s.NachsignierAuftrag.ProcessType,
			ProcessData: s.NachsignierAuftrag.ProcessData,
		})
	}
	return nachsignierungen
}

func (c Command) ZahlungKassieren(ctx context.Context, userID int, userName string, tischID int, positionen []kasse.PositionRef, kommentar string) error {
	log := zerolog.Ctx(ctx)

	// Tisch-Existenz, Status und State laden
	subject, kassensitzungNr, state, err := c.loadTischState(ctx, tischID)
	if err != nil {
		return err
	}

	// Bezahl-Invariante: nur unbezahlte Positionen können bezahlt werden
	if !validatePositionRefs(state.UnbezahltePositionen, positionen) {
		log.Warn().Int("tisch_id", tischID).Msg("Bezahl-Invariante verletzt: angeforderte Positionen nicht verfügbar")
		return ErrPositionNichtBezahlbar
	}

	resolvedPositionen, gesamtZahlungCents := resolvePositions(state.UnbezahltePositionen, positionen)

	evt, err := kasse.NewZahlungKassiertEvent(subject, userID, userName, resolvedPositionen, gesamtZahlungCents, kommentar)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to create zahlung kassiert event")
		return err
	}

	signierung, err := c.signZahlungKassiertEvent(ctx, evt, resolvedPositionen, gesamtZahlungCents)
	if err != nil {
		return err
	}

	// OCC gegen den validierten Zustand: Hat sich der Stream seit dem Lesen geändert
	// (z. B. parallele Zahlung während der TSE-Signierung), schlägt der Write mit 409 fehl.
	return c.persistSignedTischEvent(ctx, signierung, subject, state.LastEventVersion, kassensitzungNr, tischID, "Zahlung kassiert")
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

	signierungen, err := c.signStornoAufteilung(ctx, subject, userID, userName, aufteilung, kommentar)
	if err != nil {
		return err
	}

	// OCC gegen den validierten Zustand: Basis ist die höchste Version des Replays,
	// gegen den die Storno-Aufteilung berechnet wurde.
	expectedVersion := 0
	if len(events) > 0 {
		expectedVersion = events[len(events)-1].Version
	}

	return c.persistStornoEvents(ctx, signierungen, subject, expectedVersion, kassensitzungNr, tischID)
}

// signStornoAufteilung erzeugt und signiert die Events einer aufgeteilten Storno-
// Aktion: zuerst die geldneutrale Korrektur (falls unbezahlte Mengen vorliegen), dann
// je betroffener Zahlung eine kassenwirksame Warenrücknahme. Die Signierungen werden in
// Schreibreihenfolge zurückgegeben (jede mit eigener TSE-Transaktion).
func (c Command) signStornoAufteilung(ctx context.Context, subject string, userID int, userName string, aufteilung kasse.StornoAufteilung, kommentar string) ([]tseApp.Signierung, error) {
	log := zerolog.Ctx(ctx)

	var signierungen []tseApp.Signierung

	if len(aufteilung.Korrektur) > 0 {
		evt, err := kasse.NewBestellungKorrigiertEvent(subject, userID, userName, aufteilung.Korrektur, aufteilung.KorrekturCents, kommentar)
		if err != nil {
			log.Error().Err(err).Str("subject", subject).Msg("Failed to create bestellung korrigiert event")
			return nil, err
		}
		signierung, err := c.signBestellungKorrigiertEvent(ctx, evt, aufteilung.Korrektur)
		if err != nil {
			return nil, err
		}
		signierungen = append(signierungen, signierung)
	}

	for _, wr := range aufteilung.Warenruecknahmen {
		evt, err := kasse.NewStornierungErteiltEvent(subject, userID, userName, wr.ZahlungID, wr.Positionen, wr.GesamtCents, kommentar)
		if err != nil {
			log.Error().Err(err).Str("subject", subject).Msg("Failed to create stornierung erteilt event")
			return nil, err
		}
		signierung, err := c.signStornierungErteiltEvent(ctx, evt, wr.Positionen, wr.GesamtCents)
		if err != nil {
			return nil, err
		}
		signierungen = append(signierungen, signierung)
	}

	return signierungen, nil
}

// persistStornoEvents weist den signierten Storno-Events fortlaufende Versionen ab der
// erwarteten Version (Stand des validierten Replays) zu und schreibt sie atomar (mit
// etwaigen Nachsignier-Aufträgen). Ein OCC-Konflikt wird zu ErrConflict.
func (c Command) persistStornoEvents(ctx context.Context, signierungen []tseApp.Signierung, subject string, expectedVersion int, kassensitzungNr int, tischID int) error {
	log := zerolog.Ctx(ctx)

	events := make([]event.Event, 0, len(signierungen))
	for i, signierung := range signierungen {
		evt := signierung.Event
		evt.Version = expectedVersion + 1 + i
		events = append(events, evt)
	}

	nachsignierungen := nachsignierungenAusSignierungen(signierungen...)

	if err := c.EventRepo.WriteTischSessionEventsAtomic(ctx, events, nachsignierungen, kassensitzungNr); err != nil {
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

func (c Command) AusgabeBestaetigen(ctx context.Context, userID int, userName string, tischID int, positionen []kasse.PositionRef, kommentar string) error {
	log := zerolog.Ctx(ctx)

	// Tisch-Existenz, Status und State laden
	subject, kassensitzungNr, state, err := c.loadTischState(ctx, tischID)
	if err != nil {
		return err
	}

	// Ausgabe-Invariante: nur ausstehende Positionen können ausgegeben werden
	if !validatePositionRefs(state.AusstehendePositionen, positionen) {
		log.Warn().Int("tisch_id", tischID).Msg("Ausgabe-Invariante verletzt: angeforderte Positionen nicht ausgebbar")
		return ErrPositionNichtAusgebbar
	}

	resolvedPositionen, _ := resolvePositions(state.AusstehendePositionen, positionen)

	evt, err := kasse.NewAusgabeBestaetigtEvent(subject, userID, userName, resolvedPositionen, kommentar)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to create ausgabe bestaetigt event")
		return err
	}

	// OCC gegen den validierten Zustand (siehe ZahlungKassieren).
	if _, err := writeEvent(ctx, c.EventRepo, evt, subject, state.LastEventVersion, kasse.StreamTypeTischSession, kassensitzungNr); err != nil {
		if errors.Is(err, ErrConflict) {
			return ErrConflict
		}
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to write ausgabe bestaetigt event to database")
		return ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Msg("Ausgabe bestätigt")
	return nil
}
