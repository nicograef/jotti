package enrichment_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/repository/repotest"
)

// A failed catalogue read is a server error, not an unknown product the Servicekraft could fix by reloading.
func TestEnrichPositionen_RepoFehlerIstDatenbankfehler(t *testing.T) {
	repo := repotest.NewProduktRepo(nil, errors.New("connection refused"))

	_, err := enrichment.EnrichPositionen(context.Background(), repo, []enrichment.PositionInput{{ProduktID: 1, VarianteID: 1, Menge: 1}})
	if !errors.Is(err, db.ErrDatabase) {
		t.Errorf("expected db.ErrDatabase, got %v", err)
	}
}
