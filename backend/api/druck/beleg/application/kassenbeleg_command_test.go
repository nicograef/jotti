package application

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nicograef/jotti/backend/domain/betreiber"
	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/tse"
	"github.com/nicograef/jotti/backend/repository/repotest"
)

func TestKassenbelegDrucken_SuccessAndReprint(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	zahlungEvent, err := kasse.NewZahlungKassiertEvent(subject, uuid.NewString(), 1, "Test User", []kasse.Position{
		{
			PositionID:   "11111111-1111-1111-1111-111111111111",
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk", Steuersatz: "regel",
			EinzelpreisCents: 350,
			Menge:            2,
		},
	}, 700, "")
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
	eventMock.AddEvent(zahlungEvent)

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
		TSERepo:             &mockTSEAuftragRepo{},
	}

	status, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, ZahlungID: eventData.ZahlungID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status != BelegStatusEingereiht {
		t.Errorf("expected status eingereiht, got %q", status)
	}
	status, err = command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, ZahlungID: eventData.ZahlungID})
	if err != nil {
		t.Fatalf("expected no reprint error, got %v", err)
	}
	if status != BelegStatusEingereiht {
		t.Errorf("expected reprint status eingereiht, got %q", status)
	}

	if len(auftragMock.enqueued) != 2 {
		t.Fatalf("expected 2 enqueued auftraege, got %d", len(auftragMock.enqueued))
	}
	if auftragMock.enqueued[0].BonArt != "kassenbeleg" {
		t.Errorf("expected bon_art kassenbeleg, got %s", auftragMock.enqueued[0].BonArt)
	}
	if auftragMock.enqueued[0].ZielIP != "192.168.1.80" {
		t.Errorf("expected ziel_ip 192.168.1.80, got %s", auftragMock.enqueued[0].ZielIP)
	}
	if !strings.HasPrefix(auftragMock.enqueued[0].Referenz, "zahlung-kassiert:") {
		t.Errorf("expected zahlung-kassiert referenz, got %s", auftragMock.enqueued[0].Referenz)
	}
}

func TestKassenbelegDrucken_ContainsSteuerkennzeichenUndSteuermatrix(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	zahlungEvent, err := kasse.NewZahlungKassiertEvent(subject, uuid.NewString(), 1, "Test User", []kasse.Position{
		{
			PositionID:   "11111111-1111-1111-1111-111111111111",
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk", Steuersatz: "regel",
			EinzelpreisCents: 350,
			Menge:            2,
		},
		{
			PositionID:   "22222222-2222-2222-2222-222222222222",
			VarianteID:   2,
			ProduktName:  "Brezel",
			VarianteName: "normal",
			Kategorie:    "essen", Steuersatz: "ermaessigt",
			EinzelpreisCents: 300,
			Menge:            1,
		},
	}, 1000, "")
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
	eventMock.AddEvent(zahlungEvent)

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
		TSERepo:             &mockTSEAuftragRepo{},
	}

	if _, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, ZahlungID: eventData.ZahlungID}); err != nil {
		t.Fatalf("expected no error, got %v", err)
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
		"GESAMT: 10,00 EUR",
		"= 7,00 EUR (A)",
		"= 3,00 EUR (B)",
		"Steueraufteilung:",
		"A (19 %): Netto 5,88 EUR, Steuer 1,12 EUR, Brutto 7,00 EUR",
		"B (7 %): Netto 2,80 EUR, Steuer 0,20 EUR, Brutto 3,00 EUR",
	}

	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Errorf("kassenbeleg payload enthaelt %q nicht; got:\n%q", check, got)
		}
	}
}

func TestKassenbelegDrucken_Tischzahlung_WithErsteBestellungKlartext(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	zahlungEvent, err := kasse.NewZahlungKassiertEvent(subject, uuid.NewString(), 1, "Test User", []kasse.Position{
		{
			PositionID:   "11111111-1111-4111-8111-111111111111",
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

	ersteBestellung := time.Date(2026, 5, 1, 18, 1, 0, 0, time.UTC)

	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	eventMock.AddEvent(zahlungEvent)
	eventMock.SetTischSession(subject, kasse.TischSession{
		Subject:                subject,
		TischID:                testActiveTisch.ID,
		KassensitzungNr:        testKassensitzungNr,
		ErsteBestellungLogTime: &ersteBestellung,
	})

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
		TSERepo:             &mockTSEAuftragRepo{},
	}

	if _, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, ZahlungID: eventData.ZahlungID}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(auftragMock.enqueued) != 1 {
		t.Fatalf("expected exactly 1 enqueued auftrag, got %d", len(auftragMock.enqueued))
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}

	got := string(payload)
	if !strings.Contains(got, "Erste Bestellung: 01.05.2026 20:01:00") {
		t.Errorf("expected first order klartext in table receipt, got:\n%q", got)
	}
}

func TestKassenbelegDrucken_ZahlungNichtGefunden(t *testing.T) {
	ctx := context.Background()
	command := Command{
		EventRepo:           repotest.NewKassenjournalRepo(nil, nil),
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
		BetreiberRepo:       &mockBetreiberRepo{},
		DruckstationRepo:    &mockDruckstationRepo{},
		DruckauftragRepo:    &mockDruckauftragRepo{},
		TSERepo:             &mockTSEAuftragRepo{},
	}

	_, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, ZahlungID: "11111111-1111-1111-1111-111111111111"})
	if !errors.Is(err, ErrZahlungNichtGefunden) {
		t.Errorf("expected ErrZahlungNichtGefunden, got %v", err)
	}
}

func TestKassenbelegDrucken_KassenbelegDruckerNichtKonfiguriert(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	zahlungEvent, err := kasse.NewZahlungKassiertEvent(subject, uuid.NewString(), 1, "Test User", []kasse.Position{
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
	eventMock.AddEvent(zahlungEvent)

	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
		BetreiberRepo:       &mockBetreiberRepo{},
		DruckstationRepo:    &mockDruckstationRepo{},
		DruckauftragRepo:    &mockDruckauftragRepo{},
		TSERepo:             &mockTSEAuftragRepo{},
	}

	if _, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, ZahlungID: eventData.ZahlungID}); !errors.Is(err, ErrKassenbelegDruckerNichtKonfiguriert) {
		t.Errorf("expected ErrKassenbelegDruckerNichtKonfiguriert, got %v", err)
	}
}

func TestKassenbelegDrucken_TischStorno_DruckbarAlsStornobeleg(t *testing.T) {
	ctx := context.Background()
	subject := kasse.TischSessionSubject(testKassensitzungNr, testActiveTisch.ID)

	orderEvent, _ := kasse.NewBestellungAufgenommenEvent(subject, 1, "Test User", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", []kasse.Position{{
		VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 2,
	}}, "")
	var orderData kasse.BestellungAufgenommenV1Data
	if err := json.Unmarshal(orderEvent.Data, &orderData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}
	posID := orderData.Positionen[0].PositionID

	paymentEvent, _ := kasse.NewZahlungKassiertEvent(subject, uuid.NewString(), 1, "Test User", []kasse.Position{{
		PositionID: posID, VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 2,
	}}, 700, "")
	var paymentData kasse.ZahlungKassiertV1Data
	if err := json.Unmarshal(paymentEvent.Data, &paymentData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}

	stornoEvent, err := kasse.NewStornierungErteiltEvent(subject, uuid.NewString(), 2, "Leitung", paymentData.ZahlungID, []kasse.Position{{
		PositionID: posID, VarianteID: 1, ProduktName: "Cola", VarianteName: "0,5l", Kategorie: "getraenk", Steuersatz: "regel", EinzelpreisCents: 350, Menge: 2,
	}}, 700, "Rückgabe")
	if err != nil {
		t.Fatalf("expected no storno event error, got %v", err)
	}
	var stornoData kasse.StornierungErteiltV1Data
	if err := json.Unmarshal(stornoEvent.Data, &stornoData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}

	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	eventMock.AddEvent(orderEvent)   // ID 1
	eventMock.AddEvent(paymentEvent) // ID 2
	eventMock.AddEvent(stornoEvent)  // ID 3

	// Die Signatur des Storno-Vorgangs liegt am quittierten Auftrag.
	tseRepo := &mockTSEAuftragRepo{staende: map[int]tse.SignaturauftragStand{
		3: {Status: tse.StatusErledigt, ErstelltAm: time.Date(2026, 6, 10, 19, 20, 0, 0, time.UTC), Signatur: &tse.Signatur{
			TransaktionNummer: 4003,
			SignaturZaehler:   101,
			TSESeriennummer:   "SW-TSE-SN-0044",
			LogTimeStart:      time.Date(2026, 6, 10, 19, 20, 1, 0, time.UTC),
			LogTimeEnd:        time.Date(2026, 6, 10, 19, 20, 3, 0, time.UTC),
			Signatur:          "SIG-STORNO",
			QRCodeData:        "V0;STORNO",
		}},
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

	if _, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{TischID: testActiveTisch.ID, StornierungID: stornoData.StornierungID}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(auftragMock.enqueued) != 1 {
		t.Fatalf("expected exactly 1 enqueued auftrag, got %d", len(auftragMock.enqueued))
	}
	if !strings.HasPrefix(auftragMock.enqueued[0].Referenz, "stornierung-erteilt:") {
		t.Errorf("expected stornierung-erteilt referenz, got %s", auftragMock.enqueued[0].Referenz)
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}

	got := string(payload)
	// Referenz auf den Ursprungs-Zahlungsbeleg (Event-ID 2) und negativer Betrag.
	for _, check := range []string{"STORNOBELEG", "Storno zu Bon-Nr: 2", "GESAMT: -7,00 EUR", "SIG-STORNO"} {
		if !strings.Contains(got, check) {
			t.Errorf("expected %q in stornobeleg, got:\n%q", check, got)
		}
	}
}
