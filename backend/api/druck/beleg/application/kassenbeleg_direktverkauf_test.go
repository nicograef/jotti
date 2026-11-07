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

func TestKassenbelegDrucken_Direktverkauf_ExactlyOneAuftrag(t *testing.T) {
	ctx := context.Background()
	verkaufID := uuid.New().String()
	subject := kasse.DirektverkaufSubject(testKassensitzungNr, verkaufID)

	verkaufEvent, err := kasse.NewDirektverkaufGetaetigtEvent(subject, verkaufID, 1, "Test User", []kasse.Position{
		{
			VarianteID:   1,
			ProduktName:  "Cola",
			VarianteName: "0,5l",
			Kategorie:    "getraenk", Steuersatz: "regel",
			EinzelpreisCents: 350,
			Menge:            2,
		},
	}, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	eventMock.AddEvent(verkaufEvent)

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

	if _, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{VerkaufID: verkaufID}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(auftragMock.enqueued) != 1 {
		t.Fatalf("expected exactly 1 enqueued auftrag, got %d", len(auftragMock.enqueued))
	}
	if auftragMock.enqueued[0].BonArt != "kassenbeleg" {
		t.Errorf("expected bon_art kassenbeleg, got %s", auftragMock.enqueued[0].BonArt)
	}
	if !strings.HasPrefix(auftragMock.enqueued[0].Referenz, "direktverkauf-getaetigt:") {
		t.Errorf("expected direktverkauf-getaetigt referenz, got %s", auftragMock.enqueued[0].Referenz)
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}
	if strings.Contains(string(payload), "Erste Bestellung:") {
		t.Errorf("Direktverkauf-Beleg darf keinen Durchbedienen-Klarschriftzeitpunkt enthalten, got:\n%q", string(payload))
	}
}

func TestKassenbelegDrucken_Direktverkauf_NichtGefunden(t *testing.T) {
	ctx := context.Background()
	command := Command{
		EventRepo:           repotest.NewKassenjournalRepo(nil, nil),
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
		BetreiberRepo:       &mockBetreiberRepo{},
		DruckstationRepo:    &mockDruckstationRepo{},
		DruckauftragRepo:    &mockDruckauftragRepo{},
		TSERepo:             &mockTSEAuftragRepo{},
	}

	_, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{VerkaufID: uuid.New().String()})
	if !errors.Is(err, ErrVerkaufNichtGefunden) {
		t.Errorf("expected ErrVerkaufNichtGefunden, got %v", err)
	}
}

func TestKassenbelegDrucken_Direktverkauf_MitSignaturAmAuftrag(t *testing.T) {
	ctx := context.Background()
	verkaufID := uuid.New().String()
	subject := kasse.DirektverkaufSubject(testKassensitzungNr, verkaufID)

	verkaufEvent, err := kasse.NewDirektverkaufGetaetigtEvent(subject, verkaufID, 1, "Test User", []kasse.Position{{
		VarianteID:   1,
		ProduktName:  "Cola",
		VarianteName: "0,5l",
		Kategorie:    "getraenk", Steuersatz: "regel",
		EinzelpreisCents: 350,
		Menge:            2,
	}}, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	eventMock.AddEvent(verkaufEvent) // Event-ID 1

	tseRepo := &mockTSEAuftragRepo{staende: map[int]tse.SignaturauftragStand{
		1: {Status: tse.StatusErledigt, ErstelltAm: time.Date(2026, 6, 10, 19, 0, 0, 0, time.UTC), Signatur: &tse.Signatur{
			TransaktionNummer: 4001,
			SignaturZaehler:   99,
			TSESeriennummer:   "SW-TSE-SN-0044",
			LogTimeStart:      time.Date(2026, 6, 10, 19, 0, 1, 0, time.UTC),
			LogTimeEnd:        time.Date(2026, 6, 10, 19, 0, 3, 0, time.UTC),
			Signatur:          "SIG-DIREKTVERKAUF",
			QRCodeData:        "V0;DIREKTVERKAUF",
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

	if _, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{VerkaufID: verkaufID}); err != nil {
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
	for _, check := range []string{"TSE-Daten:", "TSE-Transaktion: 4001", "Signaturzaehler: 99", "TSE-Seriennummer: SW-TSE-SN-0044", "SIG-DIREKTVERKAUF", "V0;DIREKTVERKAUF"} {
		if !strings.Contains(got, check) {
			t.Errorf("expected %q in direktverkauf receipt, got:\n%q", check, got)
		}
	}
}

func TestKassenbelegDrucken_DirektverkaufStorno_DruckbarAlsStornobeleg(t *testing.T) {
	ctx := context.Background()
	verkaufID := uuid.New().String()
	subject := kasse.DirektverkaufSubject(testKassensitzungNr, verkaufID)

	verkaufEvent, err := kasse.NewDirektverkaufGetaetigtEvent(subject, verkaufID, 1, "Test User", []kasse.Position{{
		VarianteID:   1,
		ProduktName:  "Cola",
		VarianteName: "0,5l",
		Kategorie:    "getraenk", Steuersatz: "regel",
		EinzelpreisCents: 350,
		Menge:            2,
	}}, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	stornoEvent, err := kasse.NewDirektverkaufStorniertEvent(subject, verkaufID, 2, "Leitung", []kasse.Position{{
		PositionID:   "11111111-1111-4111-8111-111111111111",
		VarianteID:   1,
		ProduktName:  "Cola",
		VarianteName: "0,5l",
		Kategorie:    "getraenk", Steuersatz: "regel",
		EinzelpreisCents: 350,
		Menge:            2,
	}}, 700, "Rückgabe")
	if err != nil {
		t.Fatalf("expected no storno event error, got %v", err)
	}

	var stornoData kasse.DirektverkaufStorniertV1Data
	if err := json.Unmarshal(stornoEvent.Data, &stornoData); err != nil {
		t.Fatalf("expected no unmarshal error, got %v", err)
	}

	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	eventMock.AddEvent(verkaufEvent) // Event-ID 1
	eventMock.AddEvent(stornoEvent)  // Event-ID 2

	// Die Signatur des Storno-Vorgangs liegt am quittierten Auftrag.
	tseRepo := &mockTSEAuftragRepo{staende: map[int]tse.SignaturauftragStand{
		2: {Status: tse.StatusErledigt, ErstelltAm: time.Date(2026, 6, 10, 19, 20, 0, 0, time.UTC), Signatur: &tse.Signatur{
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

	if _, err := command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{VerkaufID: verkaufID, StornierungID: stornoData.StornierungID}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(auftragMock.enqueued) != 1 {
		t.Fatalf("expected exactly 1 enqueued auftrag, got %d", len(auftragMock.enqueued))
	}
	if !strings.HasPrefix(auftragMock.enqueued[0].Referenz, "direktverkauf-storniert:") {
		t.Errorf("expected direktverkauf-storniert referenz, got %s", auftragMock.enqueued[0].Referenz)
	}

	payload, err := base64.StdEncoding.DecodeString(auftragMock.enqueued[0].Payload)
	if err != nil {
		t.Fatalf("expected base64 payload, got decode error: %v", err)
	}

	got := string(payload)
	for _, check := range []string{"STORNOBELEG", "Storno zu Bon-Nr: 1", "GESAMT: -7,00 EUR", "-3,50 x 2 = -7,00 EUR", "SIG-STORNO"} {
		if !strings.Contains(got, check) {
			t.Errorf("expected %q in stornobeleg, got:\n%q", check, got)
		}
	}
}

func TestKassenbelegDrucken_DirektverkaufStorno_NichtGefunden(t *testing.T) {
	ctx := context.Background()
	verkaufID := uuid.New().String()
	subject := kasse.DirektverkaufSubject(testKassensitzungNr, verkaufID)

	verkaufEvent, err := kasse.NewDirektverkaufGetaetigtEvent(subject, verkaufID, 1, "Test User", []kasse.Position{{
		VarianteID:   1,
		ProduktName:  "Cola",
		VarianteName: "0,5l",
		Kategorie:    "getraenk", Steuersatz: "regel",
		EinzelpreisCents: 350,
		Menge:            1,
	}}, "")
	if err != nil {
		t.Fatalf("expected no event error, got %v", err)
	}

	eventMock := repotest.NewKassenjournalRepo(nil, nil)
	eventMock.AddEvent(verkaufEvent)

	command := Command{
		EventRepo:           eventMock,
		KassensitzungenRepo: repotest.NewKassensitzungenRepo(testOpenKS, nil),
		BetreiberRepo:       belegTestBetreiberMock(),
		DruckstationRepo:    &mockDruckstationRepo{konfig: kassenbelegStationen},
		DruckauftragRepo:    &mockDruckauftragRepo{},
		TSERepo:             &mockTSEAuftragRepo{},
	}

	_, err = command.KassenbelegDrucken(ctx, KassenbelegDruckenCommand{VerkaufID: verkaufID, StornierungID: uuid.New().String()})
	if !errors.Is(err, ErrStornierungNichtGefunden) {
		t.Errorf("expected ErrStornierungNichtGefunden, got %v", err)
	}
}
