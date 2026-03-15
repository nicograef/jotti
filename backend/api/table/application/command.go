package application

import (
	"context"
	"errors"
	"strconv"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/event"
	"github.com/nicograef/jotti/backend/domain/product"
	"github.com/nicograef/jotti/backend/domain/table"
	"github.com/rs/zerolog"
)

type tableRepo interface {
	GetTable(ctx context.Context, id int) (table.Tisch, error)
	CreateTable(ctx context.Context, t table.Tisch) (int, error)
	UpdateTable(ctx context.Context, t table.Tisch) error
	GetAllTables(ctx context.Context) ([]table.Tisch, error)
	GetActiveTables(ctx context.Context) ([]table.AktiverTisch, error)
}

type eventRepo interface {
	WriteEvent(ctx context.Context, event event.Event) (int, error)
	ReadTableState(ctx context.Context, tischID int) (table.TischState, error)
	GetMaxVersion(ctx context.Context, subject string) (int, error)
	ReadEventsBySubject(ctx context.Context, subject string) ([]event.Event, error)
}

type productRepo interface {
	GetProduct(ctx context.Context, productID int) (product.Produkt, error)
	GetVariant(ctx context.Context, variantID int) (product.Variante, error)
}

// BestellPositionInput represents the input for a single position in an order.
// The application layer enriches this with product/variant details (fat events).
type BestellPositionInput struct {
	ProduktID  int `json:"produktId"`
	VarianteID int `json:"varianteId"`
	Menge      int `json:"menge"`
}

type Command struct {
	TableRepo   tableRepo
	EventRepo   eventRepo
	ProductRepo productRepo
}

// writeEvent writes an event with optimistic concurrency control.
// It reads the current max version for the subject, sets event.Version = maxVersion + 1,
// and writes the event. Returns ErrConflict on UNIQUE constraint violation (version conflict).
func writeEvent(ctx context.Context, repo eventRepo, e event.Event, subject string) error {
	maxVersion, err := repo.GetMaxVersion(ctx, subject)
	if err != nil {
		return err
	}

	e.Version = maxVersion + 1

	_, err = repo.WriteEvent(ctx, e)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			zerolog.Ctx(ctx).Warn().
				Int("version", e.Version).
				Str("subject", subject).
				Msg("OCC conflict")
			return ErrConflict
		}
		return err
	}

	return nil
}

// loadTischState loads and validates the tisch, then reads its projected state.
// Returns ErrTischNotFound if the tisch doesn't exist, ErrTischNotActive if not active.
func (c Command) loadTischState(ctx context.Context, tischID int) (table.TischState, error) {
	log := zerolog.Ctx(ctx)

	tisch, err := c.TableRepo.GetTable(ctx, tischID)
	if err != nil {
		return table.TischState{}, fromRepositoryError(err, log, tischID)
	}

	if tisch.Status != table.ActiveStatus {
		log.Warn().Int("tisch_id", tischID).Str("status", string(tisch.Status)).Msg("Tisch is not active")
		return table.TischState{}, ErrTischNotActive
	}

	state, err := c.EventRepo.ReadTableState(ctx, tischID)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to read table state")
		return table.TischState{}, ErrDatabase
	}

	return state, nil
}

// computeNichtStorniertePositionen replays all events for a subject to compute
// the list of positions that have been ordered but not yet cancelled.
// This is used for stornierung validation (on-demand, not stored in projection).
func (c Command) computeNichtStorniertePositionen(ctx context.Context, subject string) ([]table.Position, error) {
	events, err := c.EventRepo.ReadEventsBySubject(ctx, subject)
	if err != nil {
		return nil, err
	}

	return table.ComputeNichtStorniertePositionen(events)
}

// validatePositionRefs checks that every requested PositionRef exists in the available positions
// and that the requested Menge does not exceed the available Menge.
func validatePositionRefs(available []table.Position, requested []table.PositionRef) bool {
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
	return c.applyTischStatusChange(
		ctx,
		id,
		"Tisch activated",
		func(t *table.Tisch) { t.Activate() },
	)
}

func (c Command) TischDeaktivieren(ctx context.Context, id int) error {
	return c.applyTischStatusChange(
		ctx,
		id,
		"Tisch deactivated",
		func(t *table.Tisch) { t.Deactivate() },
	)
}

func (c Command) TischLoeschen(ctx context.Context, id int) error {
	return c.applyTischStatusChange(
		ctx,
		id,
		"Tisch deleted",
		func(t *table.Tisch) { t.Delete() },
	)
}

func (c Command) applyTischStatusChange(
	ctx context.Context,
	id int,
	successMsg string,
	action func(*table.Tisch),
) error {
	log := zerolog.Ctx(ctx)

	tisch, err := c.TableRepo.GetTable(ctx, id)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	action(&tisch)

	err = c.TableRepo.UpdateTable(ctx, tisch)
	if err != nil {
		return fromRepositoryError(err, log, id)
	}

	log.Info().Int("tisch_id", id).Msg(successMsg)
	return nil
}

func (c Command) BestellungAufnehmen(ctx context.Context, userID int, userName string, tischID int, inputs []BestellPositionInput, kommentar string) error {
	log := zerolog.Ctx(ctx)

	// Tisch-Existenz und Status prüfen (Bestellungen brauchen keinen Invarianten-Check)
	if _, err := c.loadTischState(ctx, tischID); err != nil {
		return err
	}

	// Enrich positions with product/variant data (fat events)
	positionen := make([]table.Position, 0, len(inputs))
	for _, input := range inputs {
		variant, err := c.ProductRepo.GetVariant(ctx, input.VarianteID)
		if err != nil {
			log.Error().Err(err).Int("variante_id", input.VarianteID).Msg("Failed to get variant for position enrichment")
			return ErrProduktNotFound
		}

		prod, err := c.ProductRepo.GetProduct(ctx, input.ProduktID)
		if err != nil {
			log.Error().Err(err).Int("produkt_id", input.ProduktID).Msg("Failed to get product for position enrichment")
			return ErrProduktNotFound
		}

		positionen = append(positionen, table.Position{
			VarianteID:   input.VarianteID,
			ProduktName:  prod.Name,
			VarianteName: variant.Name,
			Kategorie:    string(prod.Kategorie),
			Einzelpreis:  variant.PreisCents,
			Menge:        input.Menge,
		})
	}

	event, err := table.NewBestellungAufgenommenEvent(userID, userName, tischID, positionen, kommentar)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to create bestellung aufgenommen event")
		return err
	}

	subject := "tisch:" + strconv.Itoa(tischID)
	if err := writeEvent(ctx, c.EventRepo, event, subject); err != nil {
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
func resolvePositions(available []table.Position, refs []table.PositionRef) ([]table.Position, int) {
	resolved := make([]table.Position, 0, len(refs))
	totalCents := 0
	for _, ref := range refs {
		for _, pos := range available {
			if pos.PositionID == ref.PositionID {
				resolved = append(resolved, table.Position{
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

func (c Command) ZahlungKassieren(ctx context.Context, userID int, userName string, tischID int, positionen []table.PositionRef, kommentar string) error {
	log := zerolog.Ctx(ctx)

	// Tisch-Existenz, Status und State laden
	state, err := c.loadTischState(ctx, tischID)
	if err != nil {
		return err
	}

	// Bezahl-Invariante: nur unbezahlte Positionen können bezahlt werden
	if !validatePositionRefs(state.UnbezahltePositionen, positionen) {
		log.Warn().Int("tisch_id", tischID).Msg("Bezahl-Invariante verletzt: angeforderte Positionen nicht verfügbar")
		return ErrPositionNichtBezahlbar
	}

	resolvedPositionen, gesamtZahlungCents := resolvePositions(state.UnbezahltePositionen, positionen)

	event, err := table.NewZahlungKassiertEvent(userID, userName, tischID, resolvedPositionen, gesamtZahlungCents, kommentar)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to create zahlung kassiert event")
		return err
	}

	subject := "tisch:" + strconv.Itoa(tischID)
	if err := writeEvent(ctx, c.EventRepo, event, subject); err != nil {
		if errors.Is(err, ErrConflict) {
			return ErrConflict
		}
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to write zahlung kassiert event to database")
		return ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Msg("Zahlung kassiert")
	return nil
}

func (c Command) StornierungErteilen(ctx context.Context, userID int, userName string, tischID int, positionen []table.PositionRef, kommentar string) error {
	log := zerolog.Ctx(ctx)

	// Tisch-Existenz und Status prüfen
	if _, err := c.loadTischState(ctx, tischID); err != nil {
		return err
	}

	// Stornierungsinvariante: Nur bestellte, nicht-stornierte Positionen können storniert werden
	// (unabhängig vom Bezahlstatus). On-demand event replay to compute nicht-stornierte Positionen.
	subject := "tisch:" + strconv.Itoa(tischID)
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

	event, err := table.NewStornierungErteiltEvent(userID, userName, tischID, resolvedPositionen, gesamtStornierungCents, kommentar)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to create stornierung erteilt event")
		return err
	}

	if err := writeEvent(ctx, c.EventRepo, event, subject); err != nil {
		if errors.Is(err, ErrConflict) {
			return ErrConflict
		}
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to write stornierung erteilt event to database")
		return ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Msg("Stornierung erteilt")
	return nil
}

func (c Command) AusgabeBestaetigen(ctx context.Context, userID int, userName string, tischID int, positionen []table.PositionRef, kommentar string) error {
	log := zerolog.Ctx(ctx)

	// Tisch-Existenz, Status und State laden
	state, err := c.loadTischState(ctx, tischID)
	if err != nil {
		return err
	}

	// Ausgabe-Invariante: nur ausstehende Positionen können ausgegeben werden
	if !validatePositionRefs(state.AusstehendePositionen, positionen) {
		log.Warn().Int("tisch_id", tischID).Msg("Ausgabe-Invariante verletzt: angeforderte Positionen nicht ausgebbar")
		return ErrPositionNichtAusgebbar
	}

	resolvedPositionen, _ := resolvePositions(state.AusstehendePositionen, positionen)

	event, err := table.NewAusgabeBestaetigtEvent(userID, userName, tischID, resolvedPositionen, kommentar)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to create ausgabe bestaetigt event")
		return err
	}

	subject := "tisch:" + strconv.Itoa(tischID)
	if err := writeEvent(ctx, c.EventRepo, event, subject); err != nil {
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

	// Tisch-Existenz und Status prüfen (kein Saldo-Precondition-Check)
	if _, err := c.loadTischState(ctx, tischID); err != nil {
		return err
	}

	event, err := table.NewAuszahlungGeleistetEvent(userID, userName, tischID, betragCents, kommentar)
	if err != nil {
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to create auszahlung geleistet event")
		return err
	}

	subject := "tisch:" + strconv.Itoa(tischID)
	if err := writeEvent(ctx, c.EventRepo, event, subject); err != nil {
		if errors.Is(err, ErrConflict) {
			return ErrConflict
		}
		log.Error().Err(err).Int("tisch_id", tischID).Msg("Failed to write auszahlung geleistet event to database")
		return ErrDatabase
	}

	log.Info().Int("tisch_id", tischID).Msg("Auszahlung geleistet")
	return nil
}
