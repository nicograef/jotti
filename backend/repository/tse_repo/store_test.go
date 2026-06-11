//go:build integration

package tse_repo

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
)

func setupStore(t *testing.T) (Store, *sql.DB, func(t *testing.T)) {
	t.Helper()
	database := dbpkg.OpenTestDatabase()

	reset := func(t *testing.T) {
		t.Helper()
		if _, err := database.Exec("DELETE FROM tse_signaturen"); err != nil {
			t.Fatalf("Failed to reset tse_signaturen: %v", err)
		}
		if _, err := database.Exec("DELETE FROM tse_nachsignier_auftraege"); err != nil {
			t.Fatalf("Failed to reset tse_nachsignier_auftraege: %v", err)
		}
	}
	reset(t)

	return NewStore(database), database, func(t *testing.T) {
		reset(t)
		database.Close()
	}
}

func insertAuftrag(t *testing.T, database *sql.DB, txID string) int {
	t.Helper()
	var id int
	err := database.QueryRow(`
		INSERT INTO tse_nachsignier_auftraege (tx_id, process_type, process_data, status, naechster_versuch_am, erstellt_am)
		VALUES ($1, 'Kassenbeleg-V1', 'Beleg^0.00_2.55_0.00_0.00_0.00^2.55:Bar', 'offen', NOW(), NOW())
		RETURNING id
	`, txID).Scan(&id)
	if err != nil {
		t.Fatalf("Failed to insert auftrag: %v", err)
	}
	return id
}

// Ein Fehlversuch verschiebt den naechsten Versuch in die Zukunft (Backoff):
// Der fehlschlagende Auftrag verschwindet aus dem Worker-Batch, ein neuerer
// Auftrag bleibt abholbar — kein Head-of-Line-Blocking mehr.
func TestTSENachsignierFehlversuch_BackoffBlockiertNeuereNicht(t *testing.T) {
	store, database, teardown := setupStore(t)
	defer teardown(t)
	ctx := context.Background()

	fehlschlagendID := insertAuftrag(t, database, "tx-fehlschlagend")
	neuererID := insertAuftrag(t, database, "tx-neuer")

	if err := store.TSENachsignierAuftragFehlversuch(ctx, fehlschlagendID, "fiskaly timeout"); err != nil {
		t.Fatalf("Expected no fehlversuch error, got %v", err)
	}

	offene, err := store.GetOffeneTSENachsignierAuftraege(ctx, 20)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 1 || offene[0].ID != neuererID {
		t.Fatalf("Expected only the newer auftrag to be due, got %+v", offene)
	}

	auftraege, err := store.GetTSENachsignierAuftraege(ctx)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	// Neueste zuerst: tx-neuer vor tx-fehlschlagend.
	if len(auftraege) != 2 || auftraege[1].ID != fehlschlagendID {
		t.Fatalf("Expected both auftraege newest first, got %+v", auftraege)
	}
	if auftraege[1].Status != "offen" || auftraege[1].Versuche != 1 || auftraege[1].LetzterFehler != "fiskaly timeout" {
		t.Fatalf("Expected recorded fehlversuch, got %+v", auftraege[1])
	}
}

// Nach MaxNachsignierVersuche Fehlversuchen wechselt der Auftrag auf
// fehlgeschlagen; Zuruecksetzen reiht ihn wieder ein, Verwerfen beendet ihn.
func TestTSENachsignierAuftrag_FehlgeschlagenZuruecksetzenVerwerfen(t *testing.T) {
	store, database, teardown := setupStore(t)
	defer teardown(t)
	ctx := context.Background()

	id := insertAuftrag(t, database, "tx-dauerhaft")

	for i := 0; i < MaxNachsignierVersuche; i++ {
		if err := store.TSENachsignierAuftragFehlversuch(ctx, id, "fiskaly down"); err != nil {
			t.Fatalf("Expected no fehlversuch error, got %v", err)
		}
	}

	auftraege, err := store.GetTSENachsignierAuftraege(ctx)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(auftraege) != 1 || auftraege[0].Status != "fehlgeschlagen" || auftraege[0].Versuche != MaxNachsignierVersuche {
		t.Fatalf("Expected fehlgeschlagenen auftrag after max versuche, got %+v", auftraege)
	}

	// Ein weiterer Fehlversuch aendert nichts mehr (Status-Guard offen).
	if err := store.TSENachsignierAuftragFehlversuch(ctx, id, "noch ein fehler"); err != nil {
		t.Fatalf("Expected no fehlversuch error, got %v", err)
	}
	auftraege, _ = store.GetTSENachsignierAuftraege(ctx)
	if auftraege[0].Versuche != MaxNachsignierVersuche {
		t.Fatalf("Expected versuche to stay at %d, got %d", MaxNachsignierVersuche, auftraege[0].Versuche)
	}

	// Zuruecksetzen: fehlgeschlagen -> offen, sofort wieder faellig.
	if err := store.TSENachsignierAuftragZuruecksetzen(ctx, id); err != nil {
		t.Fatalf("Expected no zuruecksetzen error, got %v", err)
	}
	offene, err := store.GetOffeneTSENachsignierAuftraege(ctx, 20)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 1 || offene[0].ID != id {
		t.Fatalf("Expected reset auftrag to be due again, got %+v", offene)
	}

	// Erneut fehlschlagen lassen und verwerfen.
	for i := 0; i < MaxNachsignierVersuche; i++ {
		if err := store.TSENachsignierAuftragFehlversuch(ctx, id, "fiskaly down"); err != nil {
			t.Fatalf("Expected no fehlversuch error, got %v", err)
		}
	}
	if err := store.TSENachsignierAuftragVerwerfen(ctx, id); err != nil {
		t.Fatalf("Expected no verwerfen error, got %v", err)
	}
	auftraege, _ = store.GetTSENachsignierAuftraege(ctx)
	if len(auftraege) != 1 || auftraege[0].Status != "verworfen" {
		t.Fatalf("Expected verworfenen auftrag, got %+v", auftraege)
	}
}

// Zuruecksetzen und Verwerfen wirken nur auf fehlgeschlagene Auftraege.
func TestTSENachsignierAuftrag_StatusGuards(t *testing.T) {
	store, database, teardown := setupStore(t)
	defer teardown(t)
	ctx := context.Background()

	id := insertAuftrag(t, database, "tx-offen")

	if err := store.TSENachsignierAuftragVerwerfen(ctx, id); err != nil {
		t.Fatalf("Expected no verwerfen error, got %v", err)
	}
	if err := store.TSENachsignierAuftragZuruecksetzen(ctx, id); err != nil {
		t.Fatalf("Expected no zuruecksetzen error, got %v", err)
	}

	auftraege, err := store.GetTSENachsignierAuftraege(ctx)
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(auftraege) != 1 || auftraege[0].Status != "offen" {
		t.Fatalf("Expected offenen auftrag to stay offen, got %+v", auftraege)
	}
}
