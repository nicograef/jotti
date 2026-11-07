// Package kassenfehler maps the open-Kassensitzung check and the journal-write failures onto
// the calling context's sentinels, so the mapping and its logging exist once.
package kassenfehler

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/internal/logging"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
)

type kassensitzungenRepo interface {
	GetAktiveKassensitzung(ctx context.Context) (*kasse.Kassensitzung, error)
}

// Sentinels names the calling context's errors. A nil field marks a failure the context never
// expects: it becomes db.ErrDatabase.
type Sentinels struct {
	KasseNichtGeoeffnet    error
	KasseWirdAbgeschlossen error
	Conflict               error
	TischNichtAktiv        error
}

// OffeneKassensitzung rejects the operation before any TSE roundtrip when no Kassensitzung is
// open or the Kassenabschluss barrier is active.
func (s Sentinels) OffeneKassensitzung(ctx context.Context, repo kassensitzungenRepo) (*kasse.Kassensitzung, error) {
	ks, err := repo.GetAktiveKassensitzung(ctx)
	if err != nil {
		logging.FromContext(ctx).Error("Failed to load aktive Kassensitzung", "error", err)
		return nil, db.ErrDatabase
	}
	if ks == nil {
		return nil, oderDatenbankfehler(s.KasseNichtGeoeffnet)
	}
	if ks.Status == kasse.KassensitzungWirdAbgeschlossen {
		return nil, oderDatenbankfehler(s.KasseWirdAbgeschlossen)
	}
	return ks, nil
}

// Schreibfehler maps a WriteEvent or WriteEventsAtomic error and logs it: a rejection as a
// warning, anything unmapped as an error. A UNIQUE(subject, version) violation or a deadlock
// means the stream changed since reading.
func (s Sentinels) Schreibfehler(ctx context.Context, err error, logArgs ...any) error {
	var mapped error
	switch {
	case errors.Is(err, db.ErrAlreadyExists), errors.Is(err, db.ErrConflict):
		mapped = s.Conflict
	case errors.Is(err, kassenjournal_repo.ErrKassensitzungNichtOffen):
		mapped = s.KasseNichtGeoeffnet
	case errors.Is(err, kassenjournal_repo.ErrTischNichtAktiv):
		mapped = s.TischNichtAktiv
	}
	args := append([]any{"error", err}, logArgs...)
	if mapped == nil {
		logging.FromContext(ctx).Error("Failed to write events", args...)
		return db.ErrDatabase
	}
	logging.FromContext(ctx).Warn("Event write rejected", args...)
	return mapped
}

func oderDatenbankfehler(sentinel error) error {
	if sentinel == nil {
		return db.ErrDatabase
	}
	return sentinel
}
