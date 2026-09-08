//go:build integration

package enrichment_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
	dbpkg "github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/produkt"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/repository/produkt_repo"
)

// katalog legt zwei Produkte mit je einer Variante an: „Pommes" (essen,
// ermaessigt) und „Cola" (getraenk, regel). Die Steuersätze unterscheiden sich,
// damit eine falsch angenommene Paarung sichtbar den Steuersatz verschöbe.
func katalog(t *testing.T) (repo produkt_repo.Repository, pommesID, pommesVarianteID, colaID, colaVarianteID int) {
	t.Helper()

	db := dbpkg.OpenTestDatabase()
	clean(t, db)
	t.Cleanup(func() {
		clean(t, db)
		_ = db.Close()
	})

	repo = produkt_repo.NewRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()

	neuesProdukt := func(name string, kategorie produkt.Kategorie, satz steuer.Steuersatz) int {
		t.Helper()
		id, err := repo.CreateProdukt(ctx, produkt.Produkt{
			Name:       name,
			Kategorie:  kategorie,
			Steuersatz: satz,
			Status:     produkt.ActiveStatus,
			Varianten:  []produkt.Variante{},
			CreatedAt:  now,
			UpdatedAt:  now,
		})
		if err != nil {
			t.Fatalf("create produkt %q: %v", name, err)
		}
		return id
	}

	neueVariante := func(produktID int, name string, preisCents int) int {
		t.Helper()
		id, err := repo.CreateVariante(ctx, produktID, produkt.Variante{
			Name:       name,
			PreisCents: preisCents,
			Status:     produkt.ActiveStatus,
			CreatedAt:  now,
			UpdatedAt:  now,
		})
		if err != nil {
			t.Fatalf("create variante %q: %v", name, err)
		}
		return id
	}

	pommesID = neuesProdukt("Pommes", produkt.EssenKategorie, steuer.ErmaessigtSteuersatz)
	pommesVarianteID = neueVariante(pommesID, "gross", 450)
	colaID = neuesProdukt("Cola", produkt.GetraenkKategorie, steuer.RegelSteuersatz)
	colaVarianteID = neueVariante(colaID, "0,5l", 300)

	return repo, pommesID, pommesVarianteID, colaID, colaVarianteID
}

func clean(t *testing.T, db *sql.DB) {
	t.Helper()
	for _, stmt := range []string{"DELETE FROM produkt_varianten", "DELETE FROM produkte"} {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("clean %q: %v", stmt, err)
		}
	}
}

// Eine fremde Variante darf keine Position erzeugen: Sonst erbte die Position
// Kategorie und Steuersatz des mitgesendeten Produkts, den Preis aber von der
// Variante eines anderen Produkts.
func TestEnrichPositionen_FremdeVarianteWirdAbgelehnt(t *testing.T) {
	repo, _, pommesVarianteID, colaID, _ := katalog(t)

	_, err := enrichment.EnrichPositionen(context.Background(), repo, []enrichment.PositionInput{
		{ProduktID: colaID, VarianteID: pommesVarianteID, Menge: 1},
	})

	if !errors.Is(err, enrichment.ErrProduktNotFound) {
		t.Fatalf("expected ErrProduktNotFound for a foreign variant, got %v", err)
	}
}

func TestEnrichPositionen_EigeneVarianteWirdAngereichert(t *testing.T) {
	repo, pommesID, pommesVarianteID, _, _ := katalog(t)

	positionen, err := enrichment.EnrichPositionen(context.Background(), repo, []enrichment.PositionInput{
		{ProduktID: pommesID, VarianteID: pommesVarianteID, Menge: 2},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(positionen) != 1 {
		t.Fatalf("expected 1 position, got %d", len(positionen))
	}
	pos := positionen[0]
	if pos.ProduktName != "Pommes" || pos.VarianteName != "gross" {
		t.Errorf("expected Pommes/gross, got %q/%q", pos.ProduktName, pos.VarianteName)
	}
	if pos.Kategorie != string(produkt.EssenKategorie) || pos.Steuersatz != string(steuer.ErmaessigtSteuersatz) {
		t.Errorf("expected essen/ermaessigt, got %q/%q", pos.Kategorie, pos.Steuersatz)
	}
	if pos.EinzelpreisCents != 450 {
		t.Errorf("expected 450 cents, got %d", pos.EinzelpreisCents)
	}
}
