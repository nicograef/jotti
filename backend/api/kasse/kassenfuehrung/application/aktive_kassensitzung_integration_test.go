//go:build integration

package application

import (
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
)

// TestGetAktiveKassensitzung_ImBarrierestatus_LiefertSitzung: Der Lesepfad der
// Kassentag-Seite liefert die Sitzung auch im Barrierestatus
// 'wird_abgeschlossen'. Bliebe sie hier aus, wirkte die Kasse nach einem
// abgebrochenen Abschluss geschlossen und die Seite zeigte das
// Eröffnen-Formular statt des unterbrochenen Abschlusses.
func TestGetAktiveKassensitzung_ImBarrierestatus_LiefertSitzung(t *testing.T) {
	ctx, _, db, _ := setupKassenfuehrungIntegration(t)

	if _, err := db.Exec(
		"UPDATE kassensitzungen SET status = $1",
		string(kasse.KassensitzungWirdAbgeschlossen),
	); err != nil {
		t.Fatalf("Barrierestatus setzen: %v", err)
	}

	query := Query{KassensitzungenRepo: kassensitzungen_repo.NewRepository(db)}

	ks, err := query.GetAktiveKassensitzung(ctx)
	if err != nil {
		t.Fatalf("GetAktiveKassensitzung: %v", err)
	}
	if ks == nil {
		t.Fatal("erwartet die Sitzung im Barrierestatus, bekam nil")
	}
	if ks.Status != kasse.KassensitzungWirdAbgeschlossen {
		t.Errorf("erwartet Status %q, bekam %q", kasse.KassensitzungWirdAbgeschlossen, ks.Status)
	}
}
