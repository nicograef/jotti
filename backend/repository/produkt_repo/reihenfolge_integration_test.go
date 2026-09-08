//go:build integration

package produkt_repo

import (
	"context"
	"slices"
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

// produktReihenfolge liest die persistierte Reihenfolge eines Produkts. Die
// Spalte taucht in keiner Response auf; nur direkt gelesen belegt sie, dass ein
// Tausch tatsaechlich stattgefunden hat.
func produktReihenfolge(t *testing.T, repo Repository, produktID int) int {
	t.Helper()
	var wert int
	if err := repo.db.QueryRow("SELECT reihenfolge FROM produkte WHERE id = $1", produktID).Scan(&wert); err != nil {
		t.Fatalf("reihenfolge lesen: %v", err)
	}
	return wert
}

// setzeProduktReihenfolge erzwingt einen Wert direkt in der Datenbank.
// Gleichstaende entstehen in echten Instanzen durch Bestandsdaten und alte
// Seeds, über das Repository sind sie nicht mehr herstellbar.
func setzeProduktReihenfolge(t *testing.T, repo Repository, produktID int, wert int) {
	t.Helper()
	if _, err := repo.db.Exec("UPDATE produkte SET reihenfolge = $1 WHERE id = $2", wert, produktID); err != nil {
		t.Fatalf("reihenfolge setzen: %v", err)
	}
}

// setzeVarianteReihenfolge erzwingt einen Wert direkt in der Datenbank, siehe
// setzeProduktReihenfolge.
func setzeVarianteReihenfolge(t *testing.T, repo Repository, varianteID int, wert int) {
	t.Helper()
	if _, err := repo.db.Exec("UPDATE produkt_varianten SET reihenfolge = $1 WHERE id = $2", wert, varianteID); err != nil {
		t.Fatalf("reihenfolge setzen: %v", err)
	}
}

// produktNamen liest die Produktnamen in der Reihenfolge, in der das Repository
// sie ausliefert.
func produktNamen(t *testing.T, repo Repository) []string {
	t.Helper()
	alle, err := repo.GetAllProdukte(context.Background())
	if err != nil {
		t.Fatalf("failed to load produkte: %v", err)
	}
	namen := make([]string, 0, len(alle))
	for i := range alle {
		namen = append(namen, alle[i].Name)
	}
	return namen
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

	if got := variantenNamen(t, repo, produktID); !slices.Equal(got, []string{"Erste", "Zweite", "Dritte"}) {
		t.Fatalf("Ausgangsreihenfolge falsch: %v", got)
	}

	if err := repo.VerschiebeVariante(ctx, zweiteID, true); err != nil {
		t.Fatalf("verschieben fehlgeschlagen: %v", err)
	}

	if got := variantenNamen(t, repo, produktID); !slices.Equal(got, []string{"Zweite", "Erste", "Dritte"}) {
		t.Errorf("nach hoch erwartet [Zweite Erste Dritte], got %v", got)
	}

	if err := repo.VerschiebeVariante(ctx, zweiteID, false); err != nil {
		t.Fatalf("verschieben fehlgeschlagen: %v", err)
	}

	if got := variantenNamen(t, repo, produktID); !slices.Equal(got, []string{"Erste", "Zweite", "Dritte"}) {
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

	if got := variantenNamen(t, repo, produktID); !slices.Equal(got, []string{"Erste", "Zweite"}) {
		t.Errorf("Reihenfolge sollte unveraendert bleiben, got %v", got)
	}
}

// Produkte tauschen nur innerhalb ihrer Kategorie: Bier steht am Ende der
// Getraenke und hat einen Nachfolger nur in der Kategorie Essen. Die
// Nachbarkategorie traegt dazu bewusst einen hoeheren, eigenen Wert — faellt
// der Kategoriefilter aus der Nachbarsuche, wandern die Reihenfolge-Werte
// sichtbar.
func TestVerschiebeProdukt_BleibtInSeinerKategorie(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	colaID, _ := repo.CreateProdukt(ctx, newProdukt("Cola", produkt.GetraenkKategorie))
	bierID, _ := repo.CreateProdukt(ctx, newProdukt("Bier", produkt.GetraenkKategorie))
	pommesID, _ := repo.CreateProdukt(ctx, newProdukt("Pommes", produkt.EssenKategorie))
	setzeProduktReihenfolge(t, repo, pommesID, 3)

	if err := repo.VerschiebeProdukt(ctx, bierID, false); err != nil {
		t.Fatalf("erwartet kein Fehler, got %v", err)
	}

	erwartet := map[string]int{"Cola": 1, "Bier": 2, "Pommes": 3}
	for name, id := range map[string]int{"Cola": colaID, "Bier": bierID, "Pommes": pommesID} {
		if got := produktReihenfolge(t, repo, id); got != erwartet[name] {
			t.Errorf("%s: reihenfolge = %d, erwartet %d", name, got, erwartet[name])
		}
	}

	if got := produktNamen(t, repo); !slices.Equal(got, []string{"Pommes", "Cola", "Bier"}) {
		t.Errorf("Bier darf Pommes nicht ueberholen (andere Kategorie), got %v", got)
	}
}

// Die alphabetische Sortierung ordnet nach deutschen Regeln: Umlaute und
// Akzente reihen sich bei ihrem Grundbuchstaben ein, nicht dahinter. Beide
// Akzentzeichen stehen am Wortanfang, weil nur dort die Collation über die
// Position entscheidet. Auf einem Cluster mit der Locale "C" ergaebe dieselbe
// Liste [Banane Zitrone Äpfel Éclair]; die COLLATE-Klausel hält die deutsche
// Reihenfolge unabhaengig von der Cluster-Locale.
func TestSortiereVariantenAlphabetisch_DeutscheCollation(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	produktID, _ := repo.CreateProdukt(ctx, newProdukt("Kaffee", produkt.GetraenkKategorie))
	_, _ = repo.CreateVariante(ctx, produktID, newVariante("Zitrone", 100, produkt.ActiveStatus))
	_, _ = repo.CreateVariante(ctx, produktID, newVariante("Éclair", 200, produkt.ActiveStatus))
	_, _ = repo.CreateVariante(ctx, produktID, newVariante("Äpfel", 300, produkt.ActiveStatus))
	_, _ = repo.CreateVariante(ctx, produktID, newVariante("Banane", 400, produkt.ActiveStatus))

	if err := repo.SortiereVariantenAlphabetisch(ctx, produktID); err != nil {
		t.Fatalf("sortieren fehlgeschlagen: %v", err)
	}

	want := []string{"Äpfel", "Banane", "Éclair", "Zitrone"}
	if got := variantenNamen(t, repo, produktID); !slices.Equal(got, want) {
		t.Errorf("erwartet %v, got %v", want, got)
	}
}

// Zwei Zeilen derselben Kategorie können denselben Reihenfolge-Wert tragen.
// Getauscht werden trotzdem die Raenge: das Verschieben ist kein stiller No-Op.
func TestVerschiebeProdukt_TauschtBeiGleichemWert(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	colaID, _ := repo.CreateProdukt(ctx, newProdukt("Cola", produkt.GetraenkKategorie))
	bierID, _ := repo.CreateProdukt(ctx, newProdukt("Bier", produkt.GetraenkKategorie))
	setzeProduktReihenfolge(t, repo, colaID, 1)
	setzeProduktReihenfolge(t, repo, bierID, 1)

	if err := repo.VerschiebeProdukt(ctx, bierID, true); err != nil {
		t.Fatalf("verschieben fehlgeschlagen: %v", err)
	}

	if got := produktNamen(t, repo); !slices.Equal(got, []string{"Bier", "Cola"}) {
		t.Errorf("erwartet [Bier Cola], got %v", got)
	}
	if bier, cola := produktReihenfolge(t, repo, bierID), produktReihenfolge(t, repo, colaID); bier >= cola {
		t.Errorf("Bier muss vor Cola liegen, got %d und %d", bier, cola)
	}
}

// Derselbe Gleichstand bei Varianten: auch dort tauscht das Verschieben Raenge.
func TestVerschiebeVariante_TauschtBeiGleichemWert(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	produktID, _ := repo.CreateProdukt(ctx, newProdukt("Bier", produkt.GetraenkKategorie))
	kleinID, _ := repo.CreateVariante(ctx, produktID, newVariante("Klein", 300, produkt.ActiveStatus))
	grossID, _ := repo.CreateVariante(ctx, produktID, newVariante("Gross", 450, produkt.ActiveStatus))
	setzeVarianteReihenfolge(t, repo, kleinID, 0)
	setzeVarianteReihenfolge(t, repo, grossID, 0)

	if err := repo.VerschiebeVariante(ctx, grossID, true); err != nil {
		t.Fatalf("verschieben fehlgeschlagen: %v", err)
	}

	if got := variantenNamen(t, repo, produktID); !slices.Equal(got, []string{"Gross", "Klein"}) {
		t.Errorf("erwartet [Gross Klein], got %v", got)
	}
}

// Ein Produkt, das die Kategorie wechselt, haengt sich ans Ende der neuen
// Kategorie und laesst sich dort sofort weiterverschieben. Behielte es seinen
// alten Wert, traefe es dort auf einen bestehenden und bliebe stecken.
func TestVerschiebeProdukt_NachKategoriewechsel(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ctx := context.Background()
	pommesID, _ := repo.CreateProdukt(ctx, newProdukt("Pommes", produkt.EssenKategorie))
	_, _ = repo.CreateProdukt(ctx, newProdukt("Cola", produkt.GetraenkKategorie))
	_, _ = repo.CreateProdukt(ctx, newProdukt("Bier", produkt.GetraenkKategorie))

	pommes := newProdukt("Pommes", produkt.GetraenkKategorie)
	pommes.ID = pommesID
	if err := repo.UpdateProdukt(ctx, pommes); err != nil {
		t.Fatalf("kategoriewechsel fehlgeschlagen: %v", err)
	}

	if got := produktReihenfolge(t, repo, pommesID); got != 3 {
		t.Errorf("Pommes muss ans Ende der Getraenke ruecken (3), got %d", got)
	}
	if got := produktNamen(t, repo); !slices.Equal(got, []string{"Cola", "Bier", "Pommes"}) {
		t.Fatalf("erwartet [Cola Bier Pommes], got %v", got)
	}

	if err := repo.VerschiebeProdukt(ctx, pommesID, true); err != nil {
		t.Fatalf("verschieben fehlgeschlagen: %v", err)
	}

	if got := produktNamen(t, repo); !slices.Equal(got, []string{"Cola", "Pommes", "Bier"}) {
		t.Errorf("erwartet [Cola Pommes Bier], got %v", got)
	}
}
