//go:build integration

package druckauftrag_repo

import (
	"context"
	"database/sql"
	"strconv"
	"testing"
	"time"

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
		_ = database.Close()
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

func TestReportDruckergebnis_QuittiertErfolge(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	id := enqueueEinen(t, repo, "192.168.1.51")

	err := repo.ReportDruckergebnis(context.Background(), []int{id}, nil)
	if err != nil {
		t.Fatalf("Expected no ReportDruckergebnis error, got %v", err)
	}

	offene, err := repo.GetOffeneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 0 {
		t.Fatalf("Expected no offene auftraege after quittieren, got %d", len(offene))
	}
}

func TestReportDruckergebnis_QuittierenIstIdempotent(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	id := enqueueEinen(t, repo, "192.168.1.51")

	for i := 0; i < 2; i++ {
		if err := repo.ReportDruckergebnis(context.Background(), []int{id}, nil); err != nil {
			t.Fatalf("Expected no ReportDruckergebnis error on call %d, got %v", i+1, err)
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

func TestReportDruckergebnis_FehlversuchZaehlung(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	id := enqueueEinen(t, repo, "192.168.1.51")

	// Fehlversuche 1..5: Auftrag bleibt offen, versuche/letzter_fehler werden
	// aktualisiert und die Backoff-Faelligkeit wird in die Zukunft gesetzt, sodass
	// der Auftrag bis dahin nicht mehr im Poll erscheint.
	for versuch := 1; versuch <= MaxDruckversuche-1; versuch++ {
		fehler := "drucker nicht erreichbar #" + strconv.Itoa(versuch)
		if err := repo.ReportDruckergebnis(context.Background(), nil, []Fehlversuch{{ID: id, Fehler: fehler}}); err != nil {
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

		naechsterVersuch := readNaechsterVersuch(t, repo, id)
		if !naechsterVersuch.Valid {
			t.Fatalf("After fehlversuch %d expected naechster_versuch_ab to be set", versuch)
		}
		if !naechsterVersuch.Time.After(time.Now()) {
			t.Fatalf("After fehlversuch %d expected naechster_versuch_ab in the future, got %v", versuch, naechsterVersuch.Time)
		}

		offene, err := repo.GetOffeneDruckauftraege(context.Background())
		if err != nil {
			t.Fatalf("Expected no read error, got %v", err)
		}
		if len(offene) != 0 {
			t.Fatalf("After fehlversuch %d expected auftrag not faellig in poll, got %d", versuch, len(offene))
		}
	}

	// Sechster Fehlversuch: Auftrag wird fehlgeschlagen und verschwindet aus dem Poll.
	if err := repo.ReportDruckergebnis(context.Background(), nil, []Fehlversuch{{ID: id, Fehler: "endgueltig"}}); err != nil {
		t.Fatalf("Expected no error on letzten fehlversuch, got %v", err)
	}

	status, versuche, letzterFehler := readAuftrag(t, repo, id)
	if status != "fehlgeschlagen" {
		t.Fatalf("After letztem fehlversuch expected status fehlgeschlagen, got %q", status)
	}
	if versuche != MaxDruckversuche {
		t.Fatalf("After letztem fehlversuch expected versuche %d, got %d", MaxDruckversuche, versuche)
	}
	if letzterFehler != "endgueltig" {
		t.Fatalf("After letztem fehlversuch expected letzter_fehler %q, got %q", "endgueltig", letzterFehler)
	}

	offene, err := repo.GetOffeneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 0 {
		t.Fatalf("Expected fehlgeschlagener auftrag to disappear from poll, got %d offene", len(offene))
	}
}

func TestReportDruckergebnis_StaleFehlversuchIstNoOp(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	// Ein Auftrag wird erst gedruckt; danach trifft (verspaetet oder doppelt) noch
	// ein Fehlversuch fuer dieselbe ID ein — der Auftrag ist nicht mehr offen.
	// Ein Nachfolger an derselben Ziel-IP zeigt, dass der stale Fehlversuch auch
	// die Warteschlange dieses Druckers nicht bremst.
	warteschlange := enqueueMehrere(t, repo, "192.168.1.51", 2)
	gedrucktID, nachfolgerID := warteschlange[0], warteschlange[1]
	if err := repo.ReportDruckergebnis(context.Background(), []int{gedrucktID}, nil); err != nil {
		t.Fatalf("Expected no error quittieren, got %v", err)
	}

	// Ein zweiter, frisch offener Auftrag im selben Ergebnis-Batch beweist, dass der
	// stale Fehlversuch den Zyklus nicht per Rollback abbricht (der Status-Guard
	// liefert ErrNoRows, das als No-Op behandelt wird).
	offenID := enqueueMehrere(t, repo, "192.168.1.52", 1)[0]

	err := repo.ReportDruckergebnis(
		context.Background(),
		[]int{offenID},
		[]Fehlversuch{{ID: gedrucktID, Fehler: "verspaetet gemeldet"}},
	)
	if err != nil {
		t.Fatalf("Expected stale fehlversuch to be a no-op, got %v", err)
	}

	// Der bereits gedruckte Auftrag bleibt unveraendert (kein Fehlversuch angerechnet).
	status, versuche, letzterFehler := readAuftrag(t, repo, gedrucktID)
	if status != "gedruckt" || versuche != 0 || letzterFehler != "" {
		t.Fatalf("Expected gedruckter auftrag unchanged, got status=%q versuche=%d letzterFehler=%q", status, versuche, letzterFehler)
	}

	// Der frische Auftrag desselben Batches wurde korrekt quittiert — der Zyklus
	// wurde also nicht abgebrochen.
	statusOffen, _, _ := readAuftrag(t, repo, offenID)
	if statusOffen != "gedruckt" {
		t.Fatalf("Expected zweiten auftrag im selben batch gedruckt, got %q", statusOffen)
	}

	// Die Warteschlange der Ziel-IP des stale Fehlversuchs bleibt ungebremst.
	if naechsterVersuch := readNaechsterVersuch(t, repo, nachfolgerID); naechsterVersuch.Valid {
		t.Fatalf("Expected no backoff on the queue after a stale fehlversuch, got %v", naechsterVersuch.Time)
	}
	offene, err := repo.GetOffeneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 1 || offene[0].ID != nachfolgerID {
		t.Fatalf("Expected nachfolger %d to stay faellig after a stale fehlversuch, got %+v", nachfolgerID, offene)
	}
}

func TestReportDruckergebnis_FehlversuchBremstNurWarteschlangeDerZielIP(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	gebremst := enqueueMehrere(t, repo, "192.168.1.51", 5)
	andere := enqueueMehrere(t, repo, "192.168.1.52", 2)

	fehlversuch := []Fehlversuch{{ID: gebremst[0], Fehler: "drucker beschaeftigt"}}
	if err := repo.ReportDruckergebnis(context.Background(), nil, fehlversuch); err != nil {
		t.Fatalf("Expected no error on fehlversuch, got %v", err)
	}

	// Die ganze Warteschlange des betroffenen Druckers wartet — nicht nur der
	// gescheiterte Auftrag. Der andere Drucker bleibt davon unberuehrt.
	assertOffeneAuftraege(t, repo, andere)

	// Nach Ablauf der Wartezeit sind beide Warteschlangen wieder faellig.
	simuliereWartezeitAbgelaufen(t, repo, "192.168.1.51")

	assertOffeneAuftraege(t, repo, append(gebremst, andere...))
}

func TestReportDruckergebnis_WarteschlangeBleibtInIDReihenfolge(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ids := enqueueMehrere(t, repo, "192.168.1.51", 5)

	fehlversuch := []Fehlversuch{{ID: ids[0], Fehler: "drucker beschaeftigt"}}
	if err := repo.ReportDruckergebnis(context.Background(), nil, fehlversuch); err != nil {
		t.Fatalf("Expected no error on fehlversuch, got %v", err)
	}

	// Unmittelbar nach dem Fehlversuch wartet die ganze Warteschlange. Laese
	// GetOffeneDruckauftraege nur die Faelligkeit der eigenen Zeile, waeren die
	// Nachfolger jetzt sofort faellig und wuerden den gebremsten Auftrag ueberholen.
	assertOffeneAuftraege(t, repo, nil)

	// Nach Ablauf der Wartezeit laeuft sie in ID-Reihenfolge weiter.
	simuliereWartezeitAbgelaufen(t, repo, "192.168.1.51")

	assertOffeneAuftraege(t, repo, ids)
}

func TestReportDruckergebnis_FehlgeschlagenerAuftragBremstWarteschlangeNicht(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	auftrag := enqueueEinen(t, repo, "192.168.1.51")

	// Fehlversuche 1..5 bremsen die Warteschlange jeweils; das Relay versucht erst
	// nach Ablauf der Wartezeit erneut, was hier per SQL simuliert wird.
	for versuch := 1; versuch < MaxDruckversuche; versuch++ {
		fehlversuch := []Fehlversuch{{ID: auftrag, Fehler: "drucker beschaeftigt"}}
		if err := repo.ReportDruckergebnis(context.Background(), nil, fehlversuch); err != nil {
			t.Fatalf("Expected no error on fehlversuch %d, got %v", versuch, err)
		}
		simuliereWartezeitAbgelaufen(t, repo, "192.168.1.51")
	}

	// Erst jetzt reiht sich ein weiterer Bon desselben Druckers ein: seine
	// Faelligkeit ist noch unberuehrt (NULL) und macht damit sichtbar, ob der
	// letzte Fehlversuch die Warteschlange anfasst.
	nachfolger := enqueueEinen(t, repo, "192.168.1.51")

	// Der MaxDruckversuche-te Fehlversuch nimmt den Auftrag aus dem Rennen — die
	// Warteschlange bekommt deshalb keinen Backoff mehr.
	fehlversuch := []Fehlversuch{{ID: auftrag, Fehler: "endgueltig"}}
	if err := repo.ReportDruckergebnis(context.Background(), nil, fehlversuch); err != nil {
		t.Fatalf("Expected no error on letztem fehlversuch, got %v", err)
	}

	status, _, _ := readAuftrag(t, repo, auftrag)
	if status != "fehlgeschlagen" {
		t.Fatalf("Expected status fehlgeschlagen after letztem fehlversuch, got %q", status)
	}
	if naechsterVersuch := readNaechsterVersuch(t, repo, nachfolger); naechsterVersuch.Valid {
		t.Fatalf("Expected the warteschlange to stay untouched by a fehlgeschlagener auftrag, got naechster_versuch_ab %v", naechsterVersuch.Time)
	}

	offene, err := repo.GetOffeneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 1 || offene[0].ID != nachfolger {
		t.Fatalf("Expected only nachfolger %d faellig, got %+v", nachfolger, offene)
	}
}

func TestReportDruckergebnis_WarteschlangeWartetAufDieLaengsteWartezeit(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	ids := enqueueMehrere(t, repo, "192.168.1.51", 2)

	// ids[0] auf zwei Fehlversuche bringen: sein naechster Fehlversuch ergibt 30 s
	// Wartezeit, waehrend der erste Fehlversuch von ids[1] nur 5 s ergibt.
	for versuch := 1; versuch <= 2; versuch++ {
		fehlversuch := []Fehlversuch{{ID: ids[0], Fehler: "drucker beschaeftigt"}}
		if err := repo.ReportDruckergebnis(context.Background(), nil, fehlversuch); err != nil {
			t.Fatalf("Expected no error on fehlversuch %d, got %v", versuch, err)
		}
		simuliereWartezeitAbgelaufen(t, repo, "192.168.1.51")
	}

	// Beide Auftraege derselben Ziel-IP melden im selben Batch einen Fehlversuch:
	// ids[0] seinen dritten (30 s Wartezeit), ids[1] seinen ersten (5 s).
	fehlversuche := []Fehlversuch{
		{ID: ids[0], Fehler: "drucker beschaeftigt"},
		{ID: ids[1], Fehler: "drucker beschaeftigt"},
	}
	if err := repo.ReportDruckergebnis(context.Background(), nil, fehlversuche); err != nil {
		t.Fatalf("Expected no error on gemischtem batch, got %v", err)
	}

	// Nach Ablauf der kuerzeren Wartezeit bleibt die Warteschlange gesperrt: ids[1]
	// waere fuer sich genommen faellig, wuerde damit aber den noch wartenden
	// ids[0] ueberholen.
	simuliereZeitvergangen(t, repo, "192.168.1.51", 10*time.Second)
	assertOffeneAuftraege(t, repo, nil)

	// Erst nach der laengeren Wartezeit laeuft die Warteschlange wieder an.
	simuliereZeitvergangen(t, repo, "192.168.1.51", 30*time.Second)
	assertOffeneAuftraege(t, repo, ids)
}

func TestGetOffeneDruckauftraege_NeuerAuftragUeberholtGebremsteWarteschlangeNicht(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	gebremst := enqueueMehrere(t, repo, "192.168.1.51", 2)
	andere := enqueueMehrere(t, repo, "192.168.1.52", 1)

	fehlversuch := []Fehlversuch{{ID: gebremst[0], Fehler: "drucker beschaeftigt"}}
	if err := repo.ReportDruckergebnis(context.Background(), nil, fehlversuch); err != nil {
		t.Fatalf("Expected no error on fehlversuch, got %v", err)
	}

	// Waehrend das Backoff-Fenster laeuft, kommt ein neuer Bon fuer denselben
	// Drucker dazu — seine eigene Faelligkeit ist NULL, er ist also fuer sich
	// genommen sofort faellig.
	neu := enqueueEinen(t, repo, "192.168.1.51")
	if naechsterVersuch := readNaechsterVersuch(t, repo, neu); naechsterVersuch.Valid {
		t.Fatalf("Expected a freshly enqueued auftrag to carry no faelligkeit, got %v", naechsterVersuch.Time)
	}

	// Die gebremste Ziel-IP wird komplett uebersprungen, der neue Auftrag ueberholt
	// die wartende Warteschlange also nicht. Der andere Drucker bleibt faellig.
	offene, err := repo.GetOffeneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 1 || offene[0].ID != andere[0] {
		t.Fatalf("Expected only auftrag %d of the ungebremste ziel_ip to be faellig, got %+v", andere[0], offene)
	}

	// Nach Ablauf der Wartezeit laeuft die Warteschlange in ID-Reihenfolge weiter —
	// der neue Auftrag zuletzt.
	simuliereWartezeitAbgelaufen(t, repo, "192.168.1.51")

	offene, err = repo.GetOffeneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	erwartet := []int{gebremst[0], gebremst[1], andere[0], neu}
	if len(offene) != len(erwartet) {
		t.Fatalf("Expected %d faellige auftraege after the wartezeit, got %+v", len(erwartet), offene)
	}
	for i, id := range erwartet {
		if offene[i].ID != id {
			t.Fatalf("Expected auftrag %d at position %d, got %+v", id, i, offene)
		}
	}
}

func TestGetOffeneDruckauftraege_RespektiertFaelligkeit(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	id := enqueueEinen(t, repo, "192.168.1.51")

	offene, err := repo.GetOffeneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 1 {
		t.Fatalf("Expected 1 offener auftrag direkt nach enqueue, got %d", len(offene))
	}

	if _, err := repo.db.Exec("UPDATE druckauftraege SET naechster_versuch_ab = NOW() + INTERVAL '1 hour' WHERE id = $1", id); err != nil {
		t.Fatalf("Failed to set naechster_versuch_ab into the future: %v", err)
	}

	offene, err = repo.GetOffeneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 0 {
		t.Fatalf("Expected auftrag not faellig while naechster_versuch_ab is in the future, got %d", len(offene))
	}

	if _, err := repo.db.Exec("UPDATE druckauftraege SET naechster_versuch_ab = NOW() - INTERVAL '1 second' WHERE id = $1", id); err != nil {
		t.Fatalf("Failed to set naechster_versuch_ab into the past: %v", err)
	}

	offene, err = repo.GetOffeneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 1 || offene[0].ID != id {
		t.Fatalf("Expected auftrag faellig again once naechster_versuch_ab is in the past, got %+v", offene)
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

func TestRetryDruckauftrag_SetztOffenUndVersucheNull(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	id := enqueueEinen(t, repo, "192.168.1.51")
	makeFehlgeschlagen(t, repo, id, "endgueltig")

	if err := repo.RetryDruckauftrag(context.Background(), id); err != nil {
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

	naechsterVersuch := readNaechsterVersuch(t, repo, id)
	if naechsterVersuch.Valid {
		t.Fatalf("Expected naechster_versuch_ab cleared after erneut versuchen, got %v", naechsterVersuch.Time)
	}
}

func TestDiscardDruckauftrag_SetztVerworfenUndBleibtErhalten(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	id := enqueueEinen(t, repo, "192.168.1.51")
	makeFehlgeschlagen(t, repo, id, "endgueltig")

	if err := repo.DiscardDruckauftrag(context.Background(), id); err != nil {
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
	id := enqueueEinen(t, repo, "192.168.1.51")

	if err := repo.RetryDruckauftrag(context.Background(), id); err != nil {
		t.Fatalf("Expected no error on erneut versuchen, got %v", err)
	}
	if err := repo.DiscardDruckauftrag(context.Background(), id); err != nil {
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

func TestDiscardAlleFehlgeschlagenen_NurFehlgeschlagene(t *testing.T) {
	repo, teardown := setup(t)
	defer teardown(t)

	err := repo.EnqueueDruckauftraege(context.Background(), []NeuerDruckauftrag{
		{ZielIP: "192.168.1.51", Payload: "AAA=", BonArt: "arbeitsbon", Referenz: "fehl-ref-1"},
		{ZielIP: "192.168.1.52", Payload: "BBB=", BonArt: "arbeitsbon", Referenz: "fehl-ref-2"},
		{ZielIP: "192.168.1.53", Payload: "CCC=", BonArt: "arbeitsbon", Referenz: "offen-ref"},
		{ZielIP: "192.168.1.54", Payload: "DDD=", BonArt: "arbeitsbon", Referenz: "gedruckt-ref"},
	})
	if err != nil {
		t.Fatalf("Expected no enqueue error, got %v", err)
	}
	offene, err := repo.GetOffeneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 4 {
		t.Fatalf("Expected 4 offene auftraege, got %d", len(offene))
	}
	fehlID1 := offene[0].ID
	fehlID2 := offene[1].ID
	offenID := offene[2].ID
	gedrucktID := offene[3].ID

	makeFehlgeschlagen(t, repo, fehlID1, "endgueltig 1")
	makeFehlgeschlagen(t, repo, fehlID2, "endgueltig 2")
	if err := repo.ReportDruckergebnis(context.Background(), []int{gedrucktID}, nil); err != nil {
		t.Fatalf("Expected no error quittieren, got %v", err)
	}

	verworfen, err := repo.DiscardAlleFehlgeschlagenen(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if verworfen != 2 {
		t.Fatalf("Expected 2 verworfene auftraege, got %d", verworfen)
	}

	fehlgeschlagene, err := repo.GetFehlgeschlageneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(fehlgeschlagene) != 0 {
		t.Fatalf("Expected no fehlgeschlagene auftraege after discard-alle, got %d", len(fehlgeschlagene))
	}

	status1, _, _ := readAuftrag(t, repo, fehlID1)
	if status1 != "verworfen" {
		t.Fatalf("Expected auftrag %d verworfen, got %q", fehlID1, status1)
	}
	status2, _, _ := readAuftrag(t, repo, fehlID2)
	if status2 != "verworfen" {
		t.Fatalf("Expected auftrag %d verworfen, got %q", fehlID2, status2)
	}

	offene, err = repo.GetOffeneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != 1 || offene[0].ID != offenID {
		t.Fatalf("Expected offener auftrag %d to stay offen and unberuehrt, got %+v", offenID, offene)
	}

	statusGedruckt, _, _ := readAuftrag(t, repo, gedrucktID)
	if statusGedruckt != "gedruckt" {
		t.Fatalf("Expected gedruckter auftrag %d to stay gedruckt, got %q", gedrucktID, statusGedruckt)
	}
}

// makeFehlgeschlagen treibt einen Auftrag über MaxDruckversuche Fehlversuche in
// den Status fehlgeschlagen.
func makeFehlgeschlagen(t *testing.T, repo Repository, id int, letzterFehler string) {
	t.Helper()
	for i := 0; i < MaxDruckversuche; i++ {
		if err := repo.ReportDruckergebnis(context.Background(), nil, []Fehlversuch{{ID: id, Fehler: letzterFehler}}); err != nil {
			t.Fatalf("Failed to drive auftrag %d into fehlgeschlagen: %v", id, err)
		}
	}
	status, _, _ := readAuftrag(t, repo, id)
	if status != "fehlgeschlagen" {
		t.Fatalf("Expected auftrag %d to be fehlgeschlagen, got %q", id, status)
	}
}

// enqueueEinen reiht einen einzelnen Auftrag fuer eine Ziel-IP ein und liefert
// dessen ID (die hoechste dieser Ziel-IP). Seine Faelligkeit ist wie bei jedem
// neuen Auftrag NULL.
func enqueueEinen(t *testing.T, repo Repository, zielIP string) int {
	t.Helper()
	err := repo.EnqueueDruckauftraege(context.Background(), []NeuerDruckauftrag{
		{ZielIP: zielIP, Payload: "AAA=", BonArt: "arbeitsbon", Referenz: "bestellung-aufgenommen:1"},
	})
	if err != nil {
		t.Fatalf("Expected no enqueue error, got %v", err)
	}

	var id int
	err = repo.db.QueryRow(
		"SELECT id FROM druckauftraege WHERE ziel_ip = $1 ORDER BY id DESC LIMIT 1", zielIP,
	).Scan(&id)
	if err != nil {
		t.Fatalf("Failed to read newest auftrag for ziel_ip %s: %v", zielIP, err)
	}
	return id
}

// enqueueMehrere legt anzahl Auftraege fuer dieselbe Ziel-IP an und liefert deren
// IDs in aufsteigender Reihenfolge — also in Zustellreihenfolge.
func enqueueMehrere(t *testing.T, repo Repository, zielIP string, anzahl int) []int {
	t.Helper()
	auftraege := make([]NeuerDruckauftrag, 0, anzahl)
	for i := 1; i <= anzahl; i++ {
		auftraege = append(auftraege, NeuerDruckauftrag{
			ZielIP:   zielIP,
			Payload:  "AAA=",
			BonArt:   "arbeitsbon",
			Referenz: "bestellung-aufgenommen:" + strconv.Itoa(i),
		})
	}
	if err := repo.EnqueueDruckauftraege(context.Background(), auftraege); err != nil {
		t.Fatalf("Expected no enqueue error, got %v", err)
	}

	rows, err := repo.db.Query("SELECT id FROM druckauftraege WHERE ziel_ip = $1 ORDER BY id ASC", zielIP)
	if err != nil {
		t.Fatalf("Failed to read auftraege for ziel_ip %s: %v", zielIP, err)
	}
	defer func() { _ = rows.Close() }()

	ids := make([]int, 0, anzahl)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("Failed to scan auftrag id: %v", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("Failed to iterate auftraege for ziel_ip %s: %v", zielIP, err)
	}
	if len(ids) != anzahl {
		t.Fatalf("Expected %d auftraege for ziel_ip %s, got %d", anzahl, zielIP, len(ids))
	}
	return ids
}

// simuliereWartezeitAbgelaufen setzt die Faelligkeit aller Auftraege einer
// Ziel-IP in die Vergangenheit, statt die Backoff-Wartezeit real abzuwarten.
func simuliereWartezeitAbgelaufen(t *testing.T, repo Repository, zielIP string) {
	t.Helper()
	_, err := repo.db.Exec(
		"UPDATE druckauftraege SET naechster_versuch_ab = NOW() - INTERVAL '1 second' WHERE ziel_ip = $1 AND status = 'offen'",
		zielIP,
	)
	if err != nil {
		t.Fatalf("Failed to expire naechster_versuch_ab for ziel_ip %s: %v", zielIP, err)
	}
}

// simuliereZeitvergangen laesst dauer verstreichen, indem es die Faelligkeiten
// aller offenen Auftraege einer Ziel-IP um diese Spanne nach vorne schiebt. So
// laufen unterschiedlich lange Wartezeiten derselben Warteschlange nacheinander
// ab, ohne real zu warten.
func simuliereZeitvergangen(t *testing.T, repo Repository, zielIP string, dauer time.Duration) {
	t.Helper()
	_, err := repo.db.Exec(
		"UPDATE druckauftraege SET naechster_versuch_ab = naechster_versuch_ab - make_interval(secs => $1) WHERE ziel_ip = $2 AND status = 'offen'",
		dauer.Seconds(), zielIP,
	)
	if err != nil {
		t.Fatalf("Failed to advance naechster_versuch_ab for ziel_ip %s: %v", zielIP, err)
	}
}

// assertOffeneAuftraege prueft den beobachtbaren Vertrag von
// GetOffeneDruckauftraege: genau diese Auftraege sind faellig, in genau dieser
// Reihenfolge.
func assertOffeneAuftraege(t *testing.T, repo Repository, erwartet []int) {
	t.Helper()
	offene, err := repo.GetOffeneDruckauftraege(context.Background())
	if err != nil {
		t.Fatalf("Expected no read error, got %v", err)
	}
	if len(offene) != len(erwartet) {
		t.Fatalf("Expected faellige auftraege %v, got %+v", erwartet, offene)
	}
	for i, id := range erwartet {
		if offene[i].ID != id {
			t.Fatalf("Expected auftrag %d at position %d of %v, got %+v", id, i, erwartet, offene)
		}
	}
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

func readNaechsterVersuch(t *testing.T, repo Repository, id int) sql.NullTime {
	t.Helper()
	var naechsterVersuch sql.NullTime
	err := repo.db.QueryRow(
		"SELECT naechster_versuch_ab FROM druckauftraege WHERE id = $1", id,
	).Scan(&naechsterVersuch)
	if err != nil {
		t.Fatalf("Failed to read naechster_versuch_ab for auftrag %d: %v", id, err)
	}
	return naechsterVersuch
}
