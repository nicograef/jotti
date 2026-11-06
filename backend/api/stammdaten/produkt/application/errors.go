package application

import (
	"errors"
	"log/slog"

	"github.com/nicograef/jotti/backend/db"
)

var ErrProduktNotFound = errors.New("produkt not found")

var ErrProduktAlreadyExists = errors.New("produkt already exists")

var ErrVarianteNotFound = errors.New("variante not found")

var ErrDatabase = db.ErrDatabase

var ErrInvalidProduktData = errors.New("invalid produkt data")

var ErrInvalidVarianteData = errors.New("invalid variante data")

// fromProduktError maps a repository error on a produkt; vorgang and logArgs identify the call site.
func fromProduktError(err error, log *slog.Logger, vorgang string, logArgs ...any) error {
	logArgs = append([]any{"vorgang", vorgang}, logArgs...)
	switch {
	case errors.Is(err, db.ErrNotFound):
		log.Warn("Produkt not found", logArgs...)
		return ErrProduktNotFound
	case errors.Is(err, db.ErrAlreadyExists):
		log.Warn("Produkt name already exists", append(logArgs, "error", err)...)
		return ErrProduktAlreadyExists
	default:
		log.Error("Produkt database error", append(logArgs, "error", err)...)
		return ErrDatabase
	}
}

// fromVarianteError maps a repository error on a variante; vorgang and logArgs identify the call site.
func fromVarianteError(err error, log *slog.Logger, vorgang string, logArgs ...any) error {
	logArgs = append([]any{"vorgang", vorgang}, logArgs...)
	if errors.Is(err, db.ErrNotFound) {
		log.Warn("Variante not found", logArgs...)
		return ErrVarianteNotFound
	}
	log.Error("Variante database error", append(logArgs, "error", err)...)
	return ErrDatabase
}
