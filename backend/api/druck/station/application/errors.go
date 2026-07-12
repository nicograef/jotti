package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
)

var ErrDatabase = db.ErrDatabase

// ErrUngueltigeDruckstation signalisiert, dass die übergebene Druckstation die
// Domain-Validierung nicht besteht (z. B. Bonmodus für kassenbeleg/abholbon).
var ErrUngueltigeDruckstation = errors.New("ungueltige druckstation")

// ErrDruckstationNichtKonfiguriert signalisiert, dass für die angeforderte
// Kategorie kein Drucker (keine IP) konfiguriert ist — ein Testbon lässt sich
// dann nicht senden.
var ErrDruckstationNichtKonfiguriert = errors.New("druckstation nicht konfiguriert")
