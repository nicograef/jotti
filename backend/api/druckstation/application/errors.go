package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
)

var ErrDatabase = db.ErrDatabase

// ErrUngueltigeDruckstation signalisiert, dass die übergebene Druckstation die
// Domain-Validierung nicht besteht (z. B. Bonmodus für kassenbeleg/abholbon).
var ErrUngueltigeDruckstation = errors.New("ungueltige druckstation")
