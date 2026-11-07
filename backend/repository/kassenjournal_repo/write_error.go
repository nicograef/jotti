package kassenjournal_repo

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
)

// MapWriteError maps a WriteEvent or WriteEventsAtomic error onto the calling context's sentinels.
// A UNIQUE(subject, version) violation or a deadlock means the stream changed since reading.
func MapWriteError(err, conflict, kasseNichtOffen error) error {
	switch {
	case errors.Is(err, db.ErrAlreadyExists), errors.Is(err, db.ErrConflict):
		return conflict
	case errors.Is(err, ErrKassensitzungNichtOffen):
		return kasseNichtOffen
	default:
		return db.ErrDatabase
	}
}
