package kassenfehler

import (
	"context"
	"errors"
	"testing"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
)

var (
	errConflict        = errors.New("conflict")
	errNichtGeoeffnet  = errors.New("kasse nicht geoeffnet")
	errImAbschluss     = errors.New("kasse wird abgeschlossen")
	errTischNichtAktiv = errors.New("tisch nicht aktiv")
)

func TestSchreibfehler(t *testing.T) {
	mitTisch := Sentinels{Conflict: errConflict, KasseNichtGeoeffnet: errNichtGeoeffnet, KasseWirdAbgeschlossen: errImAbschluss, TischNichtAktiv: errTischNichtAktiv}
	ohneTisch := Sentinels{Conflict: errConflict, KasseNichtGeoeffnet: errNichtGeoeffnet, KasseWirdAbgeschlossen: errImAbschluss}
	ohneBarriere := Sentinels{Conflict: errConflict, KasseNichtGeoeffnet: errNichtGeoeffnet}

	tests := []struct {
		name      string
		sentinels Sentinels
		err       error
		want      error
	}{
		{name: "UNIQUE(subject, version)", sentinels: mitTisch, err: db.ErrAlreadyExists, want: errConflict},
		{name: "Deadlock", sentinels: mitTisch, err: db.ErrConflict, want: errConflict},
		{name: "Kassensitzung seit dem Lesen geschlossen", sentinels: mitTisch, err: kassenjournal_repo.ErrKassensitzungNichtOffen, want: errNichtGeoeffnet},
		{name: "Abschluss-Barriere seit dem Lesen gesetzt", sentinels: mitTisch, err: kassenjournal_repo.ErrKassensitzungWirdAbgeschlossen, want: errImAbschluss},
		{name: "Barriere-Sentinel im Kontext nicht erwartet", sentinels: ohneBarriere, err: kassenjournal_repo.ErrKassensitzungWirdAbgeschlossen, want: db.ErrDatabase},
		{name: "Tisch seit dem Lesen deaktiviert oder gelöscht", sentinels: mitTisch, err: kassenjournal_repo.ErrTischNichtAktiv, want: errTischNichtAktiv},
		{name: "Tisch-Sentinel im Kontext nicht erwartet", sentinels: ohneTisch, err: kassenjournal_repo.ErrTischNichtAktiv, want: db.ErrDatabase},
		{name: "sonstiger Fehler", sentinels: mitTisch, err: errors.New("connection reset"), want: db.ErrDatabase},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sentinels.Schreibfehler(context.Background(), tt.err); !errors.Is(got, tt.want) {
				t.Errorf("Schreibfehler(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
