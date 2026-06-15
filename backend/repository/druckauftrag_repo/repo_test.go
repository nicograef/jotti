//go:build integration

package druckauftrag_repo

import (
	"context"
	"database/sql"
	"strconv"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	dbpkg "github.com/nicograef/jotti/backend/db"
)

func setup(t *testing.T) (Repository, func(t *testing.T)) {
	t.Helper()
	database := dbpkg.OpenTestDatabase()

	_, err := database.Exec("DELETE FROM druckauftraege")
	if err != nil {
		t.Fatalf("Failed to reset druckauftraege: %v", err)
	}

	return NewRepository(database), func(t *testing.T) {
		_, err := database.Exec("DELETE FROM druckauftraege")
		if err != nil {
			t.Fatalf("Failed to reset druckauftraege: %v", err)
		}
		database.Close()
	}
}

func TestEnqueueAndGetOffeneDruckauftraege(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	err := repo.EnqueueDruckauftraege(context.Background(), []NeuerDruckauftrag{
		{
			ZielIP:   "192.168.1.51",
			Payload:  "AAA=",
			BonArt:   "arbeitsbon",
			Referenz: "bestellung-aufgenommen:1",
		},
		{
			ZielIP:   "192.168.1.52",
			Payload:  "BBB=",
			BonArt:   "arbeitsbon",
			Referenz: "bestellung-aufgenommen:2",
		},
	})
	if err != nil {
		t.Fatalf("Expected no enqueue error, got %v", err)
	}

	offene, err := repo.GetOffeneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 2 {
		t.Fatalf("Expected 2 offene auftraege, got %d", len(offene))
	}
	if offene[0].ID == 0 || offene[1].ID == 0 {
		t.Fatalf("Expected generated IDs, got %+v", offene)
	}
	if offene[0].ZielIP != "192.168.1.51" || offene[1].ZielIP != "192.168.1.52" {
		t.Fatalf("Unexpected order or ziel_ip: %+v", offene)
	}
}

func TestMeldeDruckergebnis_QuittiertErfolge(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	id := enqueueOne(t, repo, "192.168.1.51")

	err := repo.MeldeDruckergebnis(context.Background(), []int{id}, nil)
	if err != nil {
		t.Fatalf("Expected no MeldeDruckergebnis error, got %v", err)
	}

	offene, err := repo.GetOffeneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 0 {
		t.Fatalf("Expected no offene auftraege after quittieren, got %d", len(offene))
	}
}

func TestMeldeDruckergebnis_QuittierenIstIdempotent(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	id := enqueueOne(t, repo, "192.168.1.51")

	for i := 0; i < 2; i++ {
		if err := repo.MeldeDruckergebnis(context.Background(), []int{id}, nil); err != nil {
			t.Fatalf("Expected no MeldeDruckergebnis error on call %d, got %v", i+1, err)
		}
	}

	status, versuche, _ := readAuftrag(t, repo, id)
	if status != "gedruckt" {
		t.Fatalf("Expected status gedruckt after doppelter Meldung, got %q", status)
	}
	if versuche != 0 {
		t.Fatalf("Expected versuche to stay 0 for a successful auftrag, got %d", versuche)
	}
}

func TestMeldeDruckergebnis_FehlversuchZaehlung(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	id := enqueueOne(t, repo, "192.168.1.51")

	// Erste beiden Fehlversuche: Auftrag bleibt offen, versuche/letzter_fehler werden aktualisiert.
	for versuch := 1; versuch <= 2; versuch++ {
		fehler := "drucker nicht erreichbar #" + strconv.Itoa(versuch)
		if err := repo.MeldeDruckergebnis(context.Background(), nil, []Fehlversuch{{ID: id, Fehler: fehler}}); err != nil {
			t.Fatalf("Expected no error on fehlversuch %d, got %v", versuch, err)
		}

		status, versuche, letzterFehler := readAuftrag(t, repo, id)
		if status != "offen" {
			t.Fatalf("After fehlversuch %d expected status offen, got %q", versuch, status)
		}
		if versuche != versuch {
			t.Fatalf("After fehlversuch %d expected versuche %d, got %d", versuch, versuch, versuche)
		}
		if letzterFehler != fehler {
			t.Fatalf("After fehlversuch %d expected letzter_fehler %q, got %q", versuch, fehler, letzterFehler)
		}

		offene, err := repo.GetOffeneDruckauftraege(context.Background())
		if err != nil {
			t.Fatalf("Expected no read error, got %v", err)
		}
		if len(offene) != 1 {
			t.Fatalf("After fehlversuch %d expected auftrag still offen in poll, got %d", versuch, len(offene))
		}
	}

	// Dritter Fehlversuch: Auftrag wird fehlgeschlagen und verschwindet aus dem Poll.
	if err := repo.MeldeDruckergebnis(context.Background(), nil, []Fehlversuch{{ID: id, Fehler: "endgueltig"}}); err != nil {
		t.Fatalf("Expected no error on dritter fehlversuch, got %v", err)
	}

	status, versuche, letzterFehler := readAuftrag(t, repo, id)
	if status != "fehlgeschlagen" {
		t.Fatalf("After dritter fehlversuch expected status fehlgeschlagen, got %q", status)
	}
	if versuche != MaxDruckversuche {
		t.Fatalf("After dritter fehlversuch expected versuche %d, got %d", MaxDruckversuche, versuche)
	}
	if letzterFehler != "endgueltig" {
		t.Fatalf("After dritter fehlversuch expected letzter_fehler %q, got %q", "endgueltig", letzterFehler)
	}

	offene, err := repo.GetOffeneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 0 {
		t.Fatalf("Expected fehlgeschlagener auftrag to disappear from poll, got %d offene", len(offene))
	}
}

func TestGetFehlgeschlageneDruckauftraege_NurFehlgeschlagene(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	err := repo.EnqueueDruckauftraege(context.Background(), []NeuerDruckauftrag{
		{ZielIP: "192.168.1.51", Payload: "AAA=", BonArt: "arbeitsbon", Referenz: "offen-ref"},
		{ZielIP: "192.168.1.52", Payload: "BBB=", BonArt: "arbeitsbon", Referenz: "fehl-ref"},
	})
	if err != nil {
		t.Fatalf("Expected no enqueue error, got %v", err)
	}

	offene, err := repo.GetOffeneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 2 {
		t.Fatalf("Expected 2 offene auftraege, got %d", len(offene))
	}
	offenID := offene[0].ID
	fehlID := offene[1].ID
	makeFehlgeschlagen(t, repo, fehlID, "endgueltig")

	fehlgeschlagene, err := repo.GetFehlgeschlageneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(fehlgeschlagene) != 1 {
		t.Fatalf("Expected exactly 1 fehlgeschlagener auftrag, got %d", len(fehlgeschlagene))
	}

	a := fehlgeschlagene[0]
	if a.ID != fehlID {
		t.Fatalf("Expected fehlgeschlagener auftrag id %d, got %d", fehlID, a.ID)
	}
	if a.ID == offenID {
		t.Fatalf("Offener auftrag should not appear among fehlgeschlagene")
	}
	if a.BonArt != "arbeitsbon" || a.ZielIP != "192.168.1.52" || a.Referenz != "fehl-ref" {
		t.Fatalf("Unexpected auftrag fields: %+v", a)
	}
	if a.Versuche != MaxDruckversuche {
		t.Fatalf("Expected versuche %d, got %d", MaxDruckversuche, a.Versuche)
	}
	if a.LetzterFehler != "endgueltig" {
		t.Fatalf("Expected letzter_fehler %q, got %q", "endgueltig", a.LetzterFehler)
	}
	if a.ErstelltAm.IsZero() {
		t.Fatalf("Expected erstellt_am to be set")
	}
}

func TestDruckauftragErneutVersuchen_SetztOffenUndVersucheNull(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	id := enqueueOne(t, repo, "192.168.1.51")
	makeFehlgeschlagen(t, repo, id, "endgueltig")

	if err := repo.DruckauftragErneutVersuchen(context.Background(), id); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	status, versuche, letzterFehler := readAuftrag(t, repo, id)
	if status != "offen" {
		t.Fatalf("Expected status offen after erneut versuchen, got %q", status)
	}
	if versuche != 0 {
		t.Fatalf("Expected versuche 0 after erneut versuchen, got %d", versuche)
	}
	if letzterFehler != "" {
		t.Fatalf("Expected letzter_fehler cleared after erneut versuchen, got %q", letzterFehler)
	}

	offene, err := repo.GetOffeneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 1 || offene[0].ID != id {
		t.Fatalf("Expected auftrag %d back in poll, got %+v", id, offene)
	}
}

func TestDruckauftragVerwerfen_SetztVerworfenUndBleibtErhalten(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	id := enqueueOne(t, repo, "192.168.1.51")
	makeFehlgeschlagen(t, repo, id, "endgueltig")

	if err := repo.DruckauftragVerwerfen(context.Background(), id); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	status, _, _ := readAuftrag(t, repo, id)
	if status != "verworfen" {
		t.Fatalf("Expected status verworfen, got %q", status)
	}

	offene, err := repo.GetOffeneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 0 {
		t.Fatalf("Expected verworfener auftrag to stay out of poll, got %d offene", len(offene))
	}

	fehlgeschlagene, err := repo.GetFehlgeschlageneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(fehlgeschlagene) != 0 {
		t.Fatalf("Expected verworfener auftrag to leave the fehlgeschlagene list, got %d", len(fehlgeschlagene))
	}
}

func TestDruckauftragTransitionen_NurAusFehlgeschlagen(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	// Auftrag bleibt offen (nicht fehlgeschlagen): beide Übergänge sind No-Ops.
	id := enqueueOne(t, repo, "192.168.1.51")

	if err := repo.DruckauftragErneutVersuchen(context.Background(), id); err != nil {
		t.Fatalf("Expected no error on erneut versuchen, got %v", err)
	}
	if err := repo.DruckauftragVerwerfen(context.Background(), id); err != nil {
		t.Fatalf("Expected no error on verwerfen, got %v", err)
	}

	status, versuche, _ := readAuftrag(t, repo, id)
	if status != "offen" {
		t.Fatalf("Expected offener auftrag to stay offen, got %q", status)
	}
	if versuche != 0 {
		t.Fatalf("Expected versuche to stay 0, got %d", versuche)
	}
}

// makeFehlgeschlagen treibt einen Auftrag über MaxDruckversuche Fehlversuche in
// den Status fehlgeschlagen.
func makeFehlgeschlagen(t *testing.T, repo Repository, id int, letzterFehler string) {
	t.Helper()
	for i := 0; i < MaxDruckversuche; i++ {
		if err := repo.MeldeDruckergebnis(context.Background(), nil, []Fehlversuch{{ID: id, Fehler: letzterFehler}}); err != nil {
			t.Fatalf("Failed to drive auftrag %d into fehlgeschlagen: %v", id, err)
		}
	}
	status, _, _ := readAuftrag(t, repo, id)
	if status != "fehlgeschlagen" {
		t.Fatalf("Expected auftrag %d to be fehlgeschlagen, got %q", id, status)
	}
}

func enqueueOne(t *testing.T, repo Repository, zielIP string) int {
	t.Helper()
	err := repo.EnqueueDruckauftraege(context.Background(), []NeuerDruckauftrag{
		{ZielIP: zielIP, Payload: "AAA=", BonArt: "arbeitsbon", Referenz: "bestellung-aufgenommen:1"},
	})
	if err != nil {
		t.Fatalf("Expected no enqueue error, got %v", err)
	}

	offene, err := repo.GetOffeneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 1 {
		t.Fatalf("Expected 1 offener auftrag, got %d", len(offene))
	}
	return offene[0].ID
}

func readAuftrag(t *testing.T, repo Repository, id int) (status string, versuche int, letzterFehler string) {
	t.Helper()
	var fehler sql.NullString
	err := repo.db.QueryRow(
		"SELECT status, versuche, letzter_fehler FROM druckauftraege WHERE id = $1", id,
	).Scan(&status, &versuche, &fehler)
	if err != nil {
		t.Fatalf("Failed to read auftrag %d: %v", id, err)
	}
	return status, versuche, fehler.String
}
