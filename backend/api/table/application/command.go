package application

import (
	"context"
	"errors"

	bondruckApp "github.com/nicograef/jotti/backend/api/bondruck/application"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/nicograef/jotti/backend/domain/settings"
	"github.com/nicograef/jotti/backend/domain/table"
	"github.com/nicograef/jotti/backend/repository/druckauftrag_repo"
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
	ReadTischSession(ctx context.Context, subject string) (kasse.TischSession, error)
	GetMaxVersion(ctx context.Context, subject string) (int, error)
	ReadEventsBySubject(ctx context.Context, subject string) ([]event.Event, error)
}

type kassensitzungenRepo interface {
	GetOffeneKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
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
	EnqueueDruckauftraege(ctx context.Context, auftraege []bondruckApp.Druckauftrag) error
}

type settingsRepo interface {
	GetBondruckEinstellungen(ctx context.Context) (settings.BondruckEinstellungen, error)
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
}

type zahlungKassiertV1Data struct {
	ZahlungID          string                `json:"zahlungId"`
	Positionen         []zahlungPositionData `json:"positionen"`
	GesamtZahlungCents int                   `json:"gesamtZahlungCents"`
	Kommentar          string                `json:"kommentar"`
}

type zahlungPositionData struct {
	PositionID   string `json:"positionId"`
	VarianteID   int    `json:"varianteId"`
	ProduktName  string `json:"produktName"`
	VarianteName string `json:"varianteName"`
	Kategorie    string `json:"kategorie"`
	Einzelpreis  int    `json:"einzelpreis"`
	Menge        int    `json:"menge"`
}

// getOffeneKassensitzungOderFehler retrieves the currently open Kassensitzung.
// Returns ErrKasseNichtGeoeffnet (HTTP 409) when no open Kassensitzung exists.
func (c Command) getOffeneKassensitzungOderFehler(ctx context.Context) (*kasse.Kassensitzung, error) {
	ks, err := c.KassensitzungenRepo.GetOffeneKassensitzung(ctx)
	if err != nil {
		return nil, ErrDatabase
	}
	if ks == nil {
		return nil, ErrKasseNichtGeoeffnet
	}
	return ks, nil
}

// writeEventOCC assigns the next version for the subject and writes the event via
// write, mapping a version conflict (UNIQUE violation) to ErrConflict.
func writeEventOCC(ctx context.Context, repo eventRepo, e event.Event, subject string, write func(event.Event) (int, error)) (int, error) {
	maxVersion, err := repo.GetMaxVersion(ctx, subject)
	if err != nil {
		return 0, err
	}

	e.Version = maxVersion + 1

	eventID, err := write(e)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			zerolog.Ctx(ctx).Warn().
				Int("version", e.Version).
				Str("subject", subject).
				Msg("OCC conflict")
			return 0, ErrConflict
		}
		return 0, err
	}

	return eventID, nil
}

// writeEvent writes an event with optimistic concurrency control.
// Returns ErrConflict on a version conflict.
func writeEvent(ctx context.Context, repo eventRepo, e event.Event, subject string, streamType kasse.StreamType, kassensitzungNr int) (int, error) {
	return writeEventOCC(ctx, repo, e, subject, func(versioned event.Event) (int, error) {
		return repo.WriteEvent(ctx, versioned, streamType, kassensitzungNr)
	})
}

// writeEventWithDruckauftraege writes an event and the print jobs derived from it
// (built from the stored event including its generated ID) in a single transaction.
// Returns ErrConflict on a version conflict.
func writeEventWithDruckauftraege(ctx context.Context, repo eventRepo, e event.Event, subject string, streamType kasse.StreamType, kassensitzungNr int, buildAuftraege func(event.Event) []druckauftrag_repo.NeuerDruckauftrag) (int, error) {
	return writeEventOCC(ctx, repo, e, subject, func(versioned event.Event) (int, error) {
		return repo.WriteEventWithDruckauftraege(ctx, versioned, streamType, kassensitzungNr, buildAuftraege)
	})
}

// konfigurierteDruckstationen returns the configured work-ticket printers, or an
// empty map when no DruckstationRepo is wired (e.g. in tests).
func (c Command) konfigurierteDruckstationen(ctx context.Context) (map[string]bondruckApp.Druckstation, error) {
	if c.DruckstationRepo == nil {
		return nil, nil
	}
	return c.DruckstationRepo.GetKonfigurierteDruckstationen(ctx)
}

// toNeuerDruckauftraege maps bondruck print jobs to the repository's insert type.
func toNeuerDruckauftraege(auftraege []bondruckApp.Druckauftrag) []druckauftrag_repo.NeuerDruckauftrag {
	result := make([]druckauftrag_repo.NeuerDruckauftrag, 0, len(auftraege))
	for _, a := range auftraege {
		result = append(result, druckauftrag_repo.NeuerDruckauftrag{
			ZielIP:   a.ZielIP,
			Payload:  a.Payload,
			BonArt:   a.BonArt,
			Referenz: a.Referenz,
		})
	}
	return result
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

// computeNichtStorniertePositionen replays all events for a subject to compute
// the list of positions that have been ordered but not yet cancelled.
// This is used for stornierung validation (on-demand, not stored in projection).
func (c Command) computeNichtStorniertePositionen(ctx context.Context, subject string) ([]kasse.Position, error) {
	events, err := c.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		return nil, err
	}

	return kasse.ComputeNichtStorniertePositionen(events)
}

// validatePositionRefs checks that every requested PositionRef exists in the available positions
// and that the requested Menge does not exceed the available Menge.
func validatePositionRefs(available []kasse.Position, requested []kasse.PositionRef) bool {
	for _, ref := range requested {
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
			Einzelpreis:  variant.PreisCents,
			Menge:        input.Menge,
		})
	}

	evt, err := kasse.NewBestellungAufgenommenEvent(subject, userID, userName, positionen, kommentar)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to create bestellung aufgenommen event")
		return err
	}

	druckstationen, err := c.konfigurierteDruckstationen(ctx)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to load druckstationen for arbeitsbon")
		return ErrDatabase
	}

	// Build the work tickets from the stored event (with its generated ID) so the
	// event and its print jobs are written in one transaction (transactional outbox).
	buildAuftraege := func(stored event.Event) []druckauftrag_repo.NeuerDruckauftrag {
		return toNeuerDruckauftraege(bondruckApp.CreateArbeitsbonAuftraegeFromEvent(stored, druckstationen, bondruckApp.DirektverkaufBondruckKonfiguration{}))
	}

	_, err = writeEventWithDruckauftraege(ctx, c.EventRepo, evt, subject, kasse.StreamTypeTischSession, kassensitzungNr, buildAuftraege)
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

	if _, err := writeEvent(ctx, c.EventRepo, evt, subject, kasse.StreamTypeTischSession, kassensitzungNr); err != nil {
		if errors.Is(err, ErrConflict) {
			return ErrConflict
		}
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to write zahlung kassiert event to database")
		return ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Msg("Zahlung kassiert")
	return nil
}

func (c Command) StornierungErteilen(ctx context.Context, userID int, userName string, tischID int, positionen []kasse.PositionRef, kommentar string) error {
	log := zerolog.Ctx(ctx)

	// Tisch-Existenz und Status prüfen, Subject und KS-Nr bestimmen
	subject, kassensitzungNr, _, err := c.loadTischState(ctx, tischID)
	if err != nil {
		return err
	}

	// Stornierungsinvariante: Nur bestellte, nicht-stornierte Positionen können storniert werden
	// (unabhängig vom Bezahlstatus). On-demand event replay to compute nicht-stornierte Positionen.
	nichtStorniert, err := c.computeNichtStorniertePositionen(ctx, subject)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to compute nicht-stornierte Positionen")
		return ErrDatabase
	}

	if !validatePositionRefs(nichtStorniert, positionen) {
		log.Warn().Int("tisch_id", tischID).Msg("Stornierungsinvariante verletzt: angeforderte Positionen nicht stornierbar")
		return ErrPositionNichtStornierbar
	}

	resolvedPositionen, gesamtStornierungCents := resolvePositions(nichtStorniert, positionen)

	evt, err := kasse.NewStornierungErteiltEvent(subject, userID, userName, resolvedPositionen, gesamtStornierungCents, kommentar)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to create stornierung erteilt event")
		return err
	}

	if _, err := writeEvent(ctx, c.EventRepo, evt, subject, kasse.StreamTypeTischSession, kassensitzungNr); err != nil {
		if errors.Is(err, ErrConflict) {
			return ErrConflict
		}
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to write stornierung erteilt event to database")
		return ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Msg("Stornierung erteilt")
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

	if _, err := writeEvent(ctx, c.EventRepo, evt, subject, kasse.StreamTypeTischSession, kassensitzungNr); err != nil {
		if errors.Is(err, ErrConflict) {
			return ErrConflict
		}
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to write ausgabe bestaetigt event to database")
		return ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Msg("Ausgabe bestätigt")
	return nil
}

func (c Command) AuszahlungLeisten(ctx context.Context, userID int, userName string, tischID int, betragCents int, kommentar string) error {
	log := zerolog.Ctx(ctx)

	// Tisch-Existenz und Status prüfen
	subject, kassensitzungNr, _, err := c.loadTischState(ctx, tischID)
	if err != nil {
		return err
	}

	evt, err := kasse.NewAuszahlungGeleistetEvent(subject, userID, userName, betragCents, kommentar)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to create auszahlung geleistet event")
		return err
	}

	if _, err := writeEvent(ctx, c.EventRepo, evt, subject, kasse.StreamTypeTischSession, kassensitzungNr); err != nil {
		if errors.Is(err, ErrConflict) {
			return ErrConflict
		}
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to write auszahlung geleistet event to database")
		return ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Msg("Auszahlung geleistet")
	return nil
}
