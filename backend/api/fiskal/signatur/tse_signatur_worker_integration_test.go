//go:build integration

package signatur

import (
	"context"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nicograef/jotti/backend/db/dbtest"
)

// While another session holds the advisory lock, the worker is refused without failing fast.
// After release, the next tick's retry acquires it.
func TestTSESignaturWorker_AdvisoryLock_ZweiteSessionHaeltLock(t *testing.T) {
	db := dbtest.Open()
	t.Cleanup(func() { _ = db.Close() })
	ctx := context.Background()

	// First "instance": its own pinned session holds the lock.
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

	// Second instance waits: ensureLock returns false, the app keeps running.
	if worker.ensureLock(ctx) {
		t.Fatal("Worker hat den Lock erhalten, obwohl eine zweite Session ihn haelt")
	}

	// The holder releases; the next tick's retry acquires the lock.
	if _, err := halter.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", tseSignaturWorkerLockKey); err != nil {
		t.Fatalf("Halter-Lock freigeben: %v", err)
	}

	if !worker.ensureLock(ctx) {
		t.Fatal("Worker hat den Lock nach der Freigabe nicht erworben")
	}

	// The lock is session-scoped: a further session is refused.
	var frei bool
	if err := db.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", tseSignaturWorkerLockKey).Scan(&frei); err != nil {
		t.Fatalf("Kontroll-Lock pruefen: %v", err)
	}
	if frei {
		t.Error("Lock war trotz haltendem Worker frei erwerbbar")
	}
}
