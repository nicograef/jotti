//go:build integration

package produkt_repo

import (
	"context"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nicograef/jotti/backend/domain/produkt"
)

// variantenNamen liest die Variantennamen eines Produkts in der Reihenfolge, in
// der das Repository sie ausliefert.
func variantenNamen(t *testing.T, repo Repository, produktID int) []string {
	t.Helper()
	p, err := repo.GetProdukt(context.Background(), produktID)
	if err != nil {
		t.Fatalf("failed to load produkt: %v", err)
	}
	namen := make([]string, 0, len(p.Varianten))
	for i := range p.Varianten {
		namen = append(namen, p.Varianten[i].Name)
	}
	return namen
}

func gleich(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Neue Varianten haengen sich hinten an, und ein Tausch mit dem Nachbarn
// vertauscht genau zwei Eintraege - der Rest der Liste bleibt stehen.
func TestVerschiebeVariante_TauschtMitNachbar(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	produktID, _ := repo.CreateProdukt(ctx, newProdukt("Wein", produkt.GetraenkKategorie))
	_, _ = repo.CreateVariante(ctx, produktID, newVariante("Erste", 100, produkt.ActiveStatus))
	zweiteID, _ := repo.CreateVariante(ctx, produktID, newVariante("Zweite", 200, produkt.ActiveStatus))
	_, _ = repo.CreateVariante(ctx, produktID, newVariante("Dritte", 300, produkt.ActiveStatus))

	if got := variantenNamen(t, repo, produktID); !gleich(got, []string{"Erste", "Zweite", "Dritte"}) {
		t.Fatalf("Ausgangsreihenfolge falsch: %v", got)
	}

	if err := repo.VerschiebeVariante(ctx, zweiteID, true); err != nil {
		t.Fatalf("verschieben fehlgeschlagen: %v", err)
	}

	if got := variantenNamen(t, repo, produktID); !gleich(got, []string{"Zweite", "Erste", "Dritte"}) {
		t.Errorf("nach hoch erwartet [Zweite Erste Dritte], got %v", got)
	}

	if err := repo.VerschiebeVariante(ctx, zweiteID, false); err != nil {
		t.Fatalf("verschieben fehlgeschlagen: %v", err)
	}

	if got := variantenNamen(t, repo, produktID); !gleich(got, []string{"Erste", "Zweite", "Dritte"}) {
		t.Errorf("nach runter erwartet Ausgangsreihenfolge, got %v", got)
	}
}

// Am Rand der Liste gibt es keinen Nachbarn: das Verschieben ist wirkungslos,
// aber kein Fehler.
func TestVerschiebeVariante_AmRandWirkungslos(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	produktID, _ := repo.CreateProdukt(ctx, newProdukt("Wein", produkt.GetraenkKategorie))
	ersteID, _ := repo.CreateVariante(ctx, produktID, newVariante("Erste", 100, produkt.ActiveStatus))
	_, _ = repo.CreateVariante(ctx, produktID, newVariante("Zweite", 200, produkt.ActiveStatus))

	if err := repo.VerschiebeVariante(ctx, ersteID, true); err != nil {
		t.Fatalf("erwartet kein Fehler am Listenrand, got %v", err)
	}

	if got := variantenNamen(t, repo, produktID); !gleich(got, []string{"Erste", "Zweite"}) {
		t.Errorf("Reihenfolge sollte unveraendert bleiben, got %v", got)
	}
}

// Produkte tauschen nur innerhalb ihrer Kategorie: ein Produkt am Anfang seiner
// Kategorie bewegt sich nicht in die davorliegende Kategorie hinein.
func TestVerschiebeProdukt_BleibtInSeinerKategorie(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	_, _ = repo.CreateProdukt(ctx, newProdukt("Pommes", produkt.EssenKategorie))
	colaID, _ := repo.CreateProdukt(ctx, newProdukt("Cola", produkt.GetraenkKategorie))
	_, _ = repo.CreateProdukt(ctx, newProdukt("Bier", produkt.GetraenkKategorie))

	if err := repo.VerschiebeProdukt(ctx, colaID, true); err != nil {
		t.Fatalf("erwartet kein Fehler, got %v", err)
	}

	alle, err := repo.GetAllProdukte(ctx)
	if err != nil {
		t.Fatalf("failed to load produkte: %v", err)
	}
	namen := make([]string, 0, len(alle))
	for i := range alle {
		namen = append(namen, alle[i].Name)
	}

	if !gleich(namen, []string{"Pommes", "Cola", "Bier"}) {
		t.Errorf("Cola darf Pommes nicht ueberholen (andere Kategorie), got %v", namen)
	}
}

// Die alphabetische Sortierung ordnet nach deutschen Regeln: Umlaute und
// Akzente reihen sich bei ihrem Grundbuchstaben ein, nicht dahinter.
func TestSortiereVariantenAlphabetisch_DeutscheCollation(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	produktID, _ := repo.CreateProdukt(ctx, newProdukt("Kaffee", produkt.GetraenkKategorie))
	_, _ = repo.CreateVariante(ctx, produktID, newVariante("Zitrone", 100, produkt.ActiveStatus))
	_, _ = repo.CreateVariante(ctx, produktID, newVariante("Cafe Creme", 200, produkt.ActiveStatus))
	_, _ = repo.CreateVariante(ctx, produktID, newVariante("\u00c4pfel", 300, produkt.ActiveStatus))
	_, _ = repo.CreateVariante(ctx, produktID, newVariante("Banane", 400, produkt.ActiveStatus))

	if err := repo.SortiereVariantenAlphabetisch(ctx, produktID); err != nil {
		t.Fatalf("sortieren fehlgeschlagen: %v", err)
	}

	want := []string{"\u00c4pfel", "Banane", "Cafe Creme", "Zitrone"}
	if got := variantenNamen(t, repo, produktID); !gleich(got, want) {
		t.Errorf("erwartet %v, got %v", want, got)
	}
}
