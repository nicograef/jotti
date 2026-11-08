package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
)

var ErrDatabase = db.ErrDatabase

var ErrUngueltigerZusatzdrucker = errors.New("ungueltiger zusatzdrucker")

var ErrZusatzdruckerNichtGefunden = errors.New("zusatzdrucker nicht gefunden")

var ErrZusatzdruckerNichtKonfiguriert = errors.New("zusatzdrucker nicht konfiguriert")

// ErrZuordnungNichtMoeglich covers a deleted Produkt and a missing Zusatzdrucker alike: one INSERT … SELECT
// tells them apart from nothing.
var ErrZuordnungNichtMoeglich = errors.New("zuordnung nicht moeglich")
