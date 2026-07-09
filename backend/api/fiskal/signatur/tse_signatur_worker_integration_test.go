//go:build integration

package signatur

import (
	"context"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
)

// Der Advisory Lock sichert die Single-Prozess-Annahme: Haelt eine zweite
// Session den Lock, bekommt der Worker ihn nicht (kein Fail-Fast, Retry am
// naechsten Tick); nach der Freigabe erwirbt der Retry ihn.
func TestTSESignaturWorker_AdvisoryLock_ZweiteSessionHaeltLock(t *testing.T) {
	db := dbpkg.OpenTestDatabase()
	t.Cleanup(func() { _ = db.Close() })
	ctx := context.Background()

	// Erste „Instanz": eigene, gepinnte Session haelt den Lock.
	halter, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Halter-Connection oeffnen: %v", err)
	}
	t.Cleanup(func() { _ = halter.Close() })

	var gehalten bool
	if err := halter.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", tseSignaturWorkerLockKey).Scan(&gehalten); err != nil {
		t.Fatalf("Halter-Lock erwerben: %v", err)
	}
	if !gehalten {
		t.Fatal("Halter-Session hat den Lock nicht bekommen — haelt ihn ein anderer Prozess?")
	}

	worker := &tseSignaturWorker{lockDB: db}
	defer worker.releaseLock()

	// Zweite Instanz wartet: ensureLock liefert false, die App laeuft weiter.
	if worker.ensureLock(ctx) {
		t.Fatal("Worker hat den Lock erhalten, obwohl eine zweite Session ihn haelt")
	}

	// Freigabe durch die haltende Session — der Retry am naechsten Tick erwirbt den Lock.
	if _, err := halter.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", tseSignaturWorkerLockKey); err != nil {
		t.Fatalf("Halter-Lock freigeben: %v", err)
	}

	if !worker.ensureLock(ctx) {
		t.Fatal("Worker hat den Lock nach der Freigabe nicht erworben")
	}

	// Der Lock ist session-gebunden gehalten: Eine weitere Session bekommt ihn nicht.
	var frei bool
	if err := db.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", tseSignaturWorkerLockKey).Scan(&frei); err != nil {
		t.Fatalf("Kontroll-Lock pruefen: %v", err)
	}
	if frei {
		t.Fatal("Lock war trotz haltendem Worker frei erwerbbar")
	}
}
