package application

import (
	"errors"

	"github.com/nicograef/jotti/backend/db"
)

var ErrDatabase = db.ErrDatabase

var ErrUngueltigerZusatzdrucker = errors.New("ungueltiger zusatzdrucker")

var ErrZusatzdruckerNichtGefunden = errors.New("zusatzdrucker nicht gefunden")

var ErrZusatzdruckerNichtKonfiguriert = errors.New("zusatzdrucker nicht konfiguriert")
