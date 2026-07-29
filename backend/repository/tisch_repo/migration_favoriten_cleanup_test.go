//go:build integration

package tisch_repo

import (
	"context"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/tisch"
)

// Pfad der geprüften Migration, relativ zum Paketverzeichnis. Der Test liest die
// Datei zur Laufzeit, statt ihre Anweisung zu kopieren — nur so schlägt er an,
// wenn sich die Bedingung in der Migration ändert.
const favoritenCleanupMigration = "../../../database/migrations/06_favoriten_cleanup.up.sql"

func erstelleTisch(t *testing.T, repo Repository, name string, status tisch.Status) int {
	t.Helper()

	now := time.Now().UTC()
	id, err := repo.CreateTable(context.Background(), tisch.Tisch{
		Name: name, Status: status, CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("Failed to create tisch %q: %v", name, err)
	}
	return id
}

// favoritenCleanupAusfuehren spielt die Migrationsdatei unverändert ein.
func favoritenCleanupAusfuehren(t *testing.T, repo Repository) {
	t.Helper()

	anweisungen, err := os.ReadFile(favoritenCleanupMigration)
	if err != nil {
		t.Fatalf("Failed to read migration %s: %v", favoritenCleanupMigration, err)
	}
	if _, err := repo.db.Exec(string(anweisungen)); err != nil {
		t.Fatalf("Failed to apply migration %s: %v", favoritenCleanupMigration, err)
	}
}

// 06_favoriten_cleanup.up.sql räumt Markierungen ab, deren Tisch gelöscht ist.
// Der Bestand deckt alle drei Tisch-Status ab: nur die Markierung des gelöschten
// Tisches darf verschwinden. Ein deaktivierter Tisch behält seine Markierung —
// er kommt zurück, sobald er wieder aktiv geschaltet wird.
//
// Beide Aussagen hängen an demselben Bestand, damit eine invertierte Bedingung
// (EXISTS statt NOT EXISTS) den Test zwingend bricht: sie beträfe genau die
// Markierungen, die bleiben müssen.
func TestFavoritenCleanupMigrationDB(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	aktivID := erstelleTisch(t, repo, "Cleanup Aktiv", tisch.ActiveStatus)
	inaktivID := erstelleTisch(t, repo, "Cleanup Inaktiv", tisch.InactiveStatus)
	geloeschtID := erstelleTisch(t, repo, "Cleanup Geloescht", tisch.DeletedStatus)

	userID := setupFavoritenUser(t, repo)
	markiereFavorit(t, repo, userID, aktivID)
	markiereFavorit(t, repo, userID, inaktivID)
	markiereFavorit(t, repo, userID, geloeschtID)

	vorher := favoritenVonUser(t, repo, userID)
	if len(vorher) != 3 {
		t.Fatalf("expected 3 marked tische before the migration, got %v", vorher)
	}

	favoritenCleanupAusfuehren(t, repo)

	nachher := favoritenVonUser(t, repo, userID)
	erwartet := []int{aktivID, inaktivID}
	slices.Sort(erwartet)
	if !slices.Equal(nachher, erwartet) {
		t.Errorf("expected only the deleted tisch %d to lose its marking, got %v (expected %v)", geloeschtID, nachher, erwartet)
	}
}
