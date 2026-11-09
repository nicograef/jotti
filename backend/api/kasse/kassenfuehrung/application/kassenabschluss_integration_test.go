//go:build integration

package application

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/repository/kassenjournal_repo"
	"github.com/nicograef/jotti/backend/repository/kassensitzungen_repo"
	"github.com/nicograef/jotti/backend/repository/tse_repo"
)

// The Abschluss reads the Soll-Bestand from real bookings: a paid Tisch order of 350, an Einlage of
// 1000 and an Entnahme of 500 over an Anfangsbestand of 0 give Soll 850; Ist 800 leaves a Fehlbetrag of 50.
func TestKasseAbschliessen_ZaehltEchteBuchungen(t *testing.T) {
	ctx, _, db, userID := setupKassenfuehrungIntegration(t)
	repo := kassenjournal_repo.NewRepository(db)
	cmd := Command{
		KassenjournalRepo:   repo,
		KassensitzungenRepo: kassensitzungen_repo.NewRepository(db),
		TSERepo:             tse_repo.NewRepository(db),
		DruckauftragRepo:    &fakeDruckauftraege{},
	}

	bezahlteBestellungAmTisch(ctx, t, db, userID, 350)
	if err := cmd.GeldtransitBuchen(ctx, userID, "test", uuid.NewString(), "einlage", 1000, "Wechselgeld"); err != nil {
		t.Fatalf("Einlage: %v", err)
	}
	if err := cmd.GeldtransitBuchen(ctx, userID, "test", uuid.NewString(), "entnahme", 500, "Abschöpfung"); err != nil {
		t.Fatalf("Entnahme: %v", err)
	}
	signaturauftraegeErledigen(t, db)

	if _, err := cmd.KasseAbschliessen(ctx, userID, "test", 800); err != nil {
		t.Fatalf("KasseAbschliessen: %v", err)
	}

	var zNr int
	if err := db.QueryRow("SELECT z_nr FROM kassensitzungen").Scan(&zNr); err != nil {
		t.Fatalf("z_nr lesen: %v", err)
	}
	events, err := repo.ReadEventsBySubject(ctx, kasse.KassensitzungSubject(zNr))
	if err != nil {
		t.Fatalf("Events lesen: %v", err)
	}
	daten := map[string]json.RawMessage{}
	for _, evt := range events {
		daten[evt.Type] = evt.Data
	}

	var kassensturz kasse.KassensturzDurchgefuehrtV1Data
	unmarshalEventData(t, daten, kasse.EventTypeKassensturzDurchgefuehrtV1, &kassensturz)
	if kassensturz.SollBestandCents != 850 || kassensturz.IstBestandCents != 800 || kassensturz.DifferenzCents != 50 {
		t.Errorf("kassensturz soll/ist/differenz = %d/%d/%d, erwartet 850/800/50", kassensturz.SollBestandCents, kassensturz.IstBestandCents, kassensturz.DifferenzCents)
	}

	var differenz kasse.DifferenzSollIstGebuchtV1Data
	unmarshalEventData(t, daten, kasse.EventTypeDifferenzSollIstGebuchtV1, &differenz)
	if differenz.BetragCents != 50 {
		t.Errorf("differenz betragCents = %d, erwartet 50", differenz.BetragCents)
	}

	var tagesabschluss kasse.TagesabschlussErstelltV1Data
	unmarshalEventData(t, daten, kasse.EventTypeTagesabschlussErstelltV1, &tagesabschluss)
	if tagesabschluss.UmsatzGesamtCents != 350 || tagesabschluss.GeldtransitCents != 500 {
		t.Errorf("tagesabschluss umsatz/geldtransit = %d/%d, erwartet 350/500", tagesabschluss.UmsatzGesamtCents, tagesabschluss.GeldtransitCents)
	}

	var status string
	if err := db.QueryRow("SELECT status FROM kassensitzungen").Scan(&status); err != nil {
		t.Fatalf("kassensitzung status lesen: %v", err)
	}
	if status != string(kasse.KassensitzungAbgeschlossen) {
		t.Errorf("erwartet Status %q, gespeichert: %q", kasse.KassensitzungAbgeschlossen, status)
	}
}

func unmarshalEventData(t *testing.T, daten map[string]json.RawMessage, typ kasse.EventType, ziel any) {
	t.Helper()
	data, ok := daten[string(typ)]
	if !ok {
		t.Fatalf("kein %s-Event geschrieben", typ)
	}
	if err := json.Unmarshal(data, ziel); err != nil {
		t.Fatalf("%s lesen: %v", typ, err)
	}
}
