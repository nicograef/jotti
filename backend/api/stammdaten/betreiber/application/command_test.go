//go:build unit

package application

import (
	"context"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/betreiber"
)

type spyBetreiberRepo struct {
	gemeldetAm time.Time
}

func (r *spyBetreiberRepo) UpsertBetreiber(ctx context.Context, b betreiber.Betreiber) error {
	return nil
}

func (r *spyBetreiberRepo) SetElsterGemeldetAm(ctx context.Context, gemeldetAm time.Time) error {
	r.gemeldetAm = gemeldetAm
	return nil
}

func (r *spyBetreiberRepo) ClearElsterGemeldetAm(ctx context.Context) error {
	return nil
}

// 2026-07-01T23:30:00Z ist in Europe/Berlin (Sommerzeit, UTC+2) bereits der
// 2. Juli. Das Meldedatum folgt dem Wandkalender, nicht der UTC-Uhr.
func TestSetzeElsterMeldung_MeldedatumInBerlinerZeit(t *testing.T) {
	repo := &spyBetreiberRepo{}
	command := Command{
		BetreiberRepo: repo,
		clock:         func() time.Time { return time.Date(2026, 7, 1, 23, 30, 0, 0, time.UTC) },
	}

	if err := command.SetzeElsterMeldung(context.Background()); err != nil {
		t.Fatalf("SetzeElsterMeldung failed: %v", err)
	}

	want := time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC)
	if !repo.gemeldetAm.Equal(want) {
		t.Errorf("gemeldetAm = %s, want %s", repo.gemeldetAm.Format(time.RFC3339), want.Format(time.RFC3339))
	}
}
