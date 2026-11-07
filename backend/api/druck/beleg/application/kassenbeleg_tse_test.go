package application

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/betreiber"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/repotest"
)

func TestKassenbelegDrucken_MitSignaturAmAuftrag_ContainsTSEBlock(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	zahlungEvent, err := kasse.NewZahlungKassiertEvent(subject, 1, "Test User", []kasse.Position{
		{
			PositionID:   "11111111-1111-1111-1111-111111111111",
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk", Steuersatz: "regel",
			EinzelpreisCents: 350,
			Menge:            1,
		},
	}, 350, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	var eventData struct {
		ZahlungID string `json:"zahlungId"`
	}
	if err := json.Unmarshal(zahlungEvent.Data, &eventData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}

	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	eventMock.AddEvent(zahlungEvent) // Event-ID 1

	// Die Signatur liegt am quittierten Auftrag — die einzige Signaturquelle.
	tseRepo := &mockTSEAuftragRepo{staende: map[int]tse.SignaturauftragStand{
		1: {Status: tse.StatusErledigt, ErstelltAm: time.Date(2026, 6, 10, 18, 0, 0, 0, time.UTC), Signatur: &tse.Signatur{
			TransaktionNummer: 3001,
			SignaturZaehler:   77,
			TSESeriennummer:   "SW-TSE-SN-0042",
			LogTimeStart:      time.Date(2026, 6, 10, 18, 0, 1, 0, time.UTC),
			LogTimeEnd:        time.Date(2026, 6, 10, 18, 0, 3, 0, time.UTC),
			Signatur:          "SIG-XYZ",
			QRCodeData:        "V0;XYZ",
		}},
	}}

	auftragMock := &mockDruckauftragRepo{}
	betreiberMock := &mockBetreiberRepo{
		betreiber: betreiber.Betreiber{
			Vereinsname: "SV Musterstadt",
			Strasse:     "Musterstrasse 1",
			Plz:         "12345",
			Ort:         "Musterstadt",
			UpdatedAt:   time.Now(),
		},
	}

	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		BetreiberRepo:       betreiberMock,
		DruckauftragRepo:    auftragMock,
		TSERepo:             tseRepo,
	}

	status, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, ZahlungID: eventData.ZahlungID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != BelegStatusEingereiht {
		t.Errorf("expected status eingereiht, got %q", status)
	}

	if len(auftragMock.enqueued) != 1 {
		t.Fatalf("expected exactly 1 enqueued auftrag, got %d", len(auftragMock.enqueued))
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}

	got := string(payload)
	checks := []string{
		"TSE-Daten:",
		"TSE-Transaktion: 3001",
		"Signaturzaehler: 77",
		"TSE-Seriennummer: SW-TSE-SN-0042",
		"TSE-Start: 10.06.2026 20:00:01",
		"TSE-Ende: 10.06.2026 20:00:03",
		"Signatur: SIG-XYZ",
	}

	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Errorf("kassenbeleg payload enthaelt %q nicht; got:\n%q", check, got)
		}
	}

	// Prompte Signatur: kein Nachsigniert-Vermerk (Kriterium: verspätet).
	if strings.Contains(got, "Nachsigniert") {
		t.Errorf("expected no Nachsigniert-Vermerk on promptly signed beleg, got:\n%q", got)
	}
}

// Before the Signaturauftrag is acknowledged no Druckauftrag is created (ausstehend); afterwards
// the same call prints the Beleg with the TSE section from the Auftrag's signature columns.
func TestKassenbelegDrucken_AusstehendDannEingereiht(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	zahlungEvent, err := kasse.NewZahlungKassiertEvent(subject, 1, "Test User", []kasse.Position{{
		PositionID:       "11111111-1111-4111-8111-111111111111",
		VarianteID:       1,
		ProduktName:      "Cola",
		VarianteName:     "0,5l",
		Kategorie:        "getraenk",
		Steuersatz:       "regel",
		EinzelpreisCents: 350,
		Menge:            1,
	}}, 350, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	var eventData struct {
		ZahlungID string `json:"zahlungId"`
	}
	if err := json.Unmarshal(zahlungEvent.Data, &eventData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}

	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	eventMock.AddEvent(zahlungEvent) // Event-ID 1

	// Auftrag existiert, aber der Worker hat noch nicht quittiert.
	tseRepo := &mockTSEAuftragRepo{staende: map[int]tse.SignaturauftragStand{
		1: {Status: tse.StatusOffen, ErstelltAm: time.Now().UTC()},
	}}

	auftragMock := &mockDruckauftragRepo{}
	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		BetreiberRepo:       belegTestBetreiberMock(),
		DruckauftragRepo:    auftragMock,
		TSERepo:             tseRepo,
	}

	status, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, ZahlungID: eventData.ZahlungID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != BelegStatusAusstehend {
		t.Errorf("expected status ausstehend, got %q", status)
	}
	if len(auftragMock.enqueued) != 0 {
		t.Errorf("expected no druckauftrag while signature is pending, got %d", len(auftragMock.enqueued))
	}

	// Der Worker quittiert — der nächste Abruf liefert den Beleg mit Signatur.
	tseRepo.staende[1] = tse.SignaturauftragStand{Status: tse.StatusErledigt, ErstelltAm: time.Date(2026, 6, 10, 18, 10, 0, 0, time.UTC), Signatur: &tse.Signatur{
		TransaktionNummer: 3002,
		SignaturZaehler:   78,
		TSESeriennummer:   "SW-TSE-SN-0043",
		LogTimeStart:      time.Date(2026, 6, 10, 18, 10, 1, 0, time.UTC),
		LogTimeEnd:        time.Date(2026, 6, 10, 18, 10, 3, 0, time.UTC),
		Signatur:          "SIG-NACHGEHOLT",
		QRCodeData:        "V0;NACHGEHOLT",
	}}

	status, err = command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, ZahlungID: eventData.ZahlungID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != BelegStatusEingereiht {
		t.Errorf("expected status eingereiht, got %q", status)
	}
	if len(auftragMock.enqueued) != 1 {
		t.Fatalf("expected exactly 1 enqueued auftrag, got %d", len(auftragMock.enqueued))
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}

	got := string(payload)
	if !strings.Contains(got, "TSE-Daten:") {
		t.Errorf("expected TSE block from auftrag signature, got:\n%q", got)
	}
	if !strings.Contains(got, "SIG-NACHGEHOLT") {
		t.Errorf("expected auftrag signature in payload, got:\n%q", got)
	}
}

// Verspätete Signatur (später als rund eine Minute nach Auftragserstellung):
// Der Beleg trägt das Nachsigniert-Kennzeichen — auch beim Nachdruck, der über
// denselben Weg läuft.
func TestKassenbelegDrucken_VerspaeteteSignatur_TraegtNachsigniertVermerk(t *testing.T) {
	ctx := context.Background()
	eventMock, zahlungID := belegZahlungFixture(t)

	// Auftrag um 18:00:00 erstellt, Signatur erst gegen 18:07 quittiert.
	tseRepo := &mockTSEAuftragRepo{staende: map[int]tse.SignaturauftragStand{
		1: {Status: tse.StatusErledigt, ErstelltAm: time.Date(2026, 6, 10, 18, 0, 0, 0, time.UTC), Signatur: &tse.Signatur{
			TransaktionNummer: 3005,
			SignaturZaehler:   80,
			TSESeriennummer:   "SW-TSE-SN-0043",
			LogTimeStart:      time.Date(2026, 6, 10, 18, 7, 1, 0, time.UTC),
			LogTimeEnd:        time.Date(2026, 6, 10, 18, 7, 3, 0, time.UTC),
			Signatur:          "SIG-VERSPAETET",
			QRCodeData:        "V0;VERSPAETET",
		}},
	}}

	auftragMock := &mockDruckauftragRepo{}
	command := belegTestCommand(eventMock, tseRepo, auftragMock)

	status, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, ZahlungID: zahlungID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != BelegStatusEingereiht {
		t.Errorf("expected status eingereiht, got %q", status)
	}
	if len(auftragMock.enqueued) != 1 {
		t.Fatalf("expected exactly 1 enqueued auftrag, got %d", len(auftragMock.enqueued))
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}

	got := string(payload)
	if !strings.Contains(got, "TSE-Daten:") {
		t.Errorf("expected TSE block on nachsignierter beleg, got:\n%q", got)
	}
	if !strings.Contains(got, "Nachsigniert am 10.06.2026 20:07:03") {
		t.Errorf("expected Nachsigniert-Vermerk, got:\n%q", got)
	}
}

// Endstatus des Auftrags (hier: fehlgeschlagen) ist dokumentierter Ausfall:
// Der Beleg entsteht ohne TSE-Daten, weist den Ausfall aber aus.
func TestKassenbelegDrucken_AusfallEndstatus_BelegMitAusfallvermerk(t *testing.T) {
	ctx := context.Background()
	eventMock, zahlungID := belegZahlungFixture(t)

	tseRepo := &mockTSEAuftragRepo{staende: map[int]tse.SignaturauftragStand{
		1: {Status: tse.StatusFehlgeschlagen, ErstelltAm: time.Now().UTC().Add(-2 * time.Minute)},
	}}

	auftragMock := &mockDruckauftragRepo{}
	command := belegTestCommand(eventMock, tseRepo, auftragMock)

	status, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, ZahlungID: zahlungID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != BelegStatusEingereiht {
		t.Errorf("expected status eingereiht, got %q", status)
	}
	if len(auftragMock.enqueued) != 1 {
		t.Fatalf("expected exactly 1 enqueued auftrag, got %d", len(auftragMock.enqueued))
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}

	got := string(payload)
	if strings.Contains(got, "TSE-Daten:") {
		t.Errorf("expected no TSE block on ausfall beleg, got:\n%q", got)
	}
	if !strings.Contains(got, "Signatur fehlgeschlagen.") {
		t.Errorf("expected Fehlgeschlagen-Vermerk on beleg, got:\n%q", got)
	}
	// fehlgeschlagen ist endgültig: der Beleg darf keine Nachsignierung versprechen.
	if strings.Contains(got, "automatisch nachsigniert") {
		t.Errorf("expected no Nachsignier-Versprechen on fehlgeschlagen beleg, got:\n%q", got)
	}
}

// Ein offener Auftrag während eines aktiven Störungszeitraums ist Ausfall
// (nicht ausstehend): Der Beleg entsteht sofort und weist den Ausfall aus —
// auch in der Aufholphase, solange der Rückstands-Zeitraum aktiv ist.
func TestKassenbelegDrucken_OffenBeiAktiverStoerung_BelegMitAusfallvermerk(t *testing.T) {
	ctx := context.Background()
	eventMock, zahlungID := belegZahlungFixture(t)

	tseRepo := &mockTSEAuftragRepo{
		staende: map[int]tse.SignaturauftragStand{
			1: {Status: tse.StatusOffen, ErstelltAm: time.Now().UTC().Add(-3 * time.Minute)},
		},
		stoerung: &tse.Stoerung{
			Beginn:     time.Now().UTC().Add(-time.Minute),
			GrundArt:   tse.StoerungGrundRueckstand,
			Fehlertext: "Signaturaufträge im Rückstand",
		},
	}

	auftragMock := &mockDruckauftragRepo{}
	command := belegTestCommand(eventMock, tseRepo, auftragMock)

	status, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, ZahlungID: zahlungID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != BelegStatusEingereiht {
		t.Errorf("expected status eingereiht, got %q", status)
	}
	if len(auftragMock.enqueued) != 1 {
		t.Fatalf("expected exactly 1 enqueued auftrag, got %d", len(auftragMock.enqueued))
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}
	if !strings.Contains(string(payload), "TSE-Hinweis:") {
		t.Errorf("expected Ausfallvermerk on beleg, got:\n%q", string(payload))
	}
}
