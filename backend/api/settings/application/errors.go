package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
)

var ErrDatabase = db.ErrDatabase
var ErrNotFound = db.ErrNotFound
var ErrTSENichtKonfiguriert = errors.New("tse_not_configured")
var ErrTSEVerbindungFehlgeschlagen = errors.New("tse_connection_failed")
