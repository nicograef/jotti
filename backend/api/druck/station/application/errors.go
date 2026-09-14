package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
)

var ErrDatabase = db.ErrDatabase

var ErrUngueltigeDruckstation = errors.New("ungueltige druckstation")

var ErrDruckstationNichtKonfiguriert = errors.New("druckstation nicht konfiguriert")
