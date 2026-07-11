//go:build unit

package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/reporting"
	"github.com/nicograef/jotti/backend/domain/steuer"
	"github.com/nicograef/jotti/backend/domain/tisch"
)

type mockReportingRepo struct {
	data             reporting.ReportingData
	liveData         reporting.LiveReportingData
	eigeneUebersicht reporting.EigeneUebersicht
	err              error
}

func (m mockReportingRepo) GetReporting(_ context.Context, _ int) (reporting.ReportingData, error) {
	return m.data, m.err
}

func (m mockReportingRepo) GetEigeneUebersicht(_ context.Context, _ int, _ int) (reporting.EigeneUebersicht, error) {
	return m.eigeneUebersicht, m.err
}

func (m mockReportingRepo) GetLiveReporting(_ context.Context, _ int) (reporting.LiveReportingData, error) {
	return m.liveData, m.err
}

type mockTischSessionRepo struct {
	sessions []kasse.TischSession
	err      error
}

func (m mockTischSessionRepo) GetTischSessionsByKassensitzungNr(_ context.Context, _ int) ([]kasse.TischSession, error) {
	return m.sessions, m.err
}

type mockTischRepo struct {
	tische []tisch.Tisch
	err    error
}

func (m mockTischRepo) GetAllTables(_ context.Context) ([]tisch.Tisch, error) {
	return m.tische, m.err
}

type mockKasseRepo struct {
	kassensitzungNr int
	kassensitzung   *kasse.Kassensitzung
	err             error
}

func (m mockKasseRepo) GetOffeneKassensitzungNr(_ context.Context) (int, error) {
	return m.kassensitzungNr, m.err
}

func (m mockKasseRepo) GetAllKassensitzungen(_ context.Context) ([]kasse.Kassensitzung, error) {
	return nil, m.err
}

func (m mockKasseRepo) GetOffeneKassensitzung(_ context.Context) (*kasse.Kassensitzung, error) {
	return m.kassensitzung, m.err
}

const testKassensitzungNr = 1

func TestGetReporting_HappyPath(t *testing.T) {
	expected := reporting.ReportingData{
		KassensitzungNr: testKassensitzungNr,
		Summary: reporting.Summary{
			GesamtUmsatzCents:        5000,
			AnzahlDirektverkaeufe:    2,
			DirektverkaufUmsatzCents: 2200,
		},
		Breakdowns: reporting.Breakdowns{
			UmsatzProServicekraft: []reporting.UmsatzServicekraft{},
		},
		UmsatzProSteuersatz: []reporting.UmsatzSteuersatz{},
		Stornierungen:       []reporting.StornierungDetail{},
	}

	q := Query{ReportingRepo: mockReportingRepo{data: expected}, KassensitzungenRepo: mockKasseRepo{kassensitzungNr: testKassensitzungNr}}

	result, err := q.GetReporting(context.Background(), testKassensitzungNr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Summary.GesamtUmsatzCents != 5000 {
		t.Errorf("expected 5000 cents, got %d", result.Summary.GesamtUmsatzCents)
	}
	if result.Summary.AnzahlDirektverkaeufe != 2 {
		t.Errorf("expected anzahl direktverkaeufe 2, got %d", result.Summary.AnzahlDirektverkaeufe)
	}
	if result.Summary.DirektverkaufUmsatzCents != 2200 {
		t.Errorf("expected direktverkauf umsatz 2200, got %d", result.Summary.DirektverkaufUmsatzCents)
	}
	if result.KassensitzungNr != testKassensitzungNr {
		t.Errorf("expected KassensitzungNr %d, got %d", testKassensitzungNr, result.KassensitzungNr)
	}
}

func TestGetReporting_BerechnetUmsatzProSteuersatz(t *testing.T) {
	data := reporting.ReportingData{
		KassensitzungNr: testKassensitzungNr,
		UmsatzProSteuersatz: []reporting.UmsatzSteuersatz{
			{Satz: steuer.RegelSteuersatz, BruttoCents: 1190},
			{Satz: steuer.ErmaessigtSteuersatz, BruttoCents: 107},
			{Satz: steuer.BefreitSteuersatz, BruttoCents: 500},
			{Satz: steuer.KombiSteuersatz, BruttoCents: 1000},
		},
	}

	q := Query{ReportingRepo: mockReportingRepo{data: data}, KassensitzungenRepo: mockKasseRepo{kassensitzungNr: testKassensitzungNr}}

	result, err := q.GetReporting(context.Background(), testKassensitzungNr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	bySatz := map[steuer.Steuersatz]reporting.UmsatzSteuersatz{}
	for _, eintrag := range result.UmsatzProSteuersatz {
		bySatz[eintrag.Satz] = eintrag
	}

	regel := bySatz[steuer.RegelSteuersatz]
	if regel.BruttoCents != 1490 || regel.NettoCents != 1252 || regel.SteuerCents != 238 {
		t.Fatalf("unexpected regel values: %+v", regel)
	}

	ermaessigt := bySatz[steuer.ErmaessigtSteuersatz]
	if ermaessigt.BruttoCents != 807 || ermaessigt.NettoCents != 754 || ermaessigt.SteuerCents != 53 {
		t.Fatalf("unexpected ermaessigt values: %+v", ermaessigt)
	}

	befreit := bySatz[steuer.BefreitSteuersatz]
	if befreit.BruttoCents != 500 || befreit.NettoCents != 500 || befreit.SteuerCents != 0 {
		t.Fatalf("unexpected befreit values: %+v", befreit)
	}

	if _, hasKombi := bySatz[steuer.KombiSteuersatz]; hasKombi {
		t.Fatalf("did not expect kombi row in result: %+v", result.UmsatzProSteuersatz)
	}
}

// Zeilenbasis statt Aggregatbasis (B9): Zwei Kombi-Zeilen à 10,05 € runden je
// Zeile (2 × 7,04 € ermäßigt = 14,08 €), nicht auf dem Aggregat (20,10 € →
// 14,07 €). Warenrücknahmen kommen als negative Zeilen und mindern die
// Aufschlüsselung, statt bei negativem Aggregat zu verschwinden.
func TestGetReporting_UmsatzProSteuersatzRechnetJeZeile(t *testing.T) {
	data := reporting.ReportingData{
		KassensitzungNr: testKassensitzungNr,
		UmsatzProSteuersatz: []reporting.UmsatzSteuersatz{
			{Satz: steuer.KombiSteuersatz, BruttoCents: 1005},
			{Satz: steuer.KombiSteuersatz, BruttoCents: 1005},
			{Satz: steuer.RegelSteuersatz, BruttoCents: -450},
		},
	}

	q := Query{ReportingRepo: mockReportingRepo{data: data}, KassensitzungenRepo: mockKasseRepo{kassensitzungNr: testKassensitzungNr}}

	result, err := q.GetReporting(context.Background(), testKassensitzungNr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	bySatz := map[steuer.Steuersatz]reporting.UmsatzSteuersatz{}
	for _, eintrag := range result.UmsatzProSteuersatz {
		bySatz[eintrag.Satz] = eintrag
	}

	ermaessigt := bySatz[steuer.ErmaessigtSteuersatz]
	if ermaessigt.BruttoCents != 1408 || ermaessigt.NettoCents != 1316 || ermaessigt.SteuerCents != 92 {
		t.Fatalf("unexpected ermaessigt values (Zeilenbasis): %+v", ermaessigt)
	}

	regel := bySatz[steuer.RegelSteuersatz]
	if regel.BruttoCents != 152 || regel.NettoCents != 128 || regel.SteuerCents != 24 {
		t.Fatalf("unexpected regel values (Warenrücknahme abgezogen): %+v", regel)
	}
}

func TestGetReporting_DatabaseError(t *testing.T) {
	q := Query{ReportingRepo: mockReportingRepo{err: errors.New("db connection failed")}, KassensitzungenRepo: mockKasseRepo{kassensitzungNr: testKassensitzungNr}}

	_, err := q.GetReporting(context.Background(), testKassensitzungNr)
	if !errors.Is(err, ErrDatabase) {
		t.Fatalf("expected ErrDatabase, got %v", err)
	}
}

func TestGetEigeneUebersicht_MitOffenerArbeit(t *testing.T) {
	base := reporting.EigeneUebersicht{
		AnzahlBestellungen: 4,
		BestellungenCents:  3000,
		AnzahlZahlungen:    2,
		ZahlungenCents:     1500,
	}
	sessions := []kasse.TischSession{
		{
			TischID: 3,
			UnbezahltePositionen: []kasse.Position{
				{PositionID: "p1", Menge: 1, BestellerUserID: 7},
			},
		},
		{
			// Nur fremde Arbeit -> für User 7 erledigt, darf nicht erscheinen.
			TischID: 1,
			UnbezahltePositionen: []kasse.Position{
				{PositionID: "p2", Menge: 1, BestellerUserID: 8},
			},
		},
	}
	q := Query{
		ReportingRepo:       mockReportingRepo{eigeneUebersicht: base},
		KassensitzungenRepo: mockKasseRepo{kassensitzungNr: testKassensitzungNr},
		TischSessionRepo:    mockTischSessionRepo{sessions: sessions},
		TischRepo:           mockTischRepo{tische: []tisch.Tisch{{ID: 3, Name: "Tisch 3"}, {ID: 1, Name: "Tisch 1"}}},
	}

	result, err := q.GetEigeneUebersicht(context.Background(), 7)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.AnzahlBestellungen != 4 || result.BestellungenCents != 3000 || result.AnzahlZahlungen != 2 || result.ZahlungenCents != 1500 {
		t.Errorf("expected base reporting numbers to be preserved, got %+v", result)
	}
	if result.AlleErledigt {
		t.Errorf("expected open work, AlleErledigt should be false")
	}
	if len(result.OffeneTische) != 1 {
		t.Fatalf("expected 1 offener Tisch, got %d: %+v", len(result.OffeneTische), result.OffeneTische)
	}
	offen := result.OffeneTische[0]
	if offen.TischID != 3 || offen.TischName != "Tisch 3" || offen.AnzahlOffen != 1 {
		t.Errorf("unexpected offener Tisch: %+v", offen)
	}
}

func TestGetEigeneUebersicht_AllesErledigt(t *testing.T) {
	sessions := []kasse.TischSession{
		{
			TischID: 1,
			UnbezahltePositionen: []kasse.Position{
				{PositionID: "p1", Menge: 1, BestellerUserID: 8},
			},
		},
	}
	q := Query{
		ReportingRepo:       mockReportingRepo{},
		KassensitzungenRepo: mockKasseRepo{kassensitzungNr: testKassensitzungNr},
		TischSessionRepo:    mockTischSessionRepo{sessions: sessions},
		TischRepo:           mockTischRepo{tische: []tisch.Tisch{{ID: 1, Name: "Tisch 1"}}},
	}

	result, err := q.GetEigeneUebersicht(context.Background(), 7)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.AlleErledigt {
		t.Errorf("expected AlleErledigt true for servicekraft without own work")
	}
	if len(result.OffeneTische) != 0 {
		t.Errorf("expected no offene Tische, got %+v", result.OffeneTische)
	}
}

func TestGetEigeneUebersicht_DatabaseError(t *testing.T) {
	q := Query{
		ReportingRepo:       mockReportingRepo{err: errors.New("db error")},
		KassensitzungenRepo: mockKasseRepo{kassensitzungNr: testKassensitzungNr},
		TischSessionRepo:    mockTischSessionRepo{},
		TischRepo:           mockTischRepo{},
	}

	_, err := q.GetEigeneUebersicht(context.Background(), 7)
	if !errors.Is(err, ErrDatabase) {
		t.Fatalf("expected ErrDatabase, got %v", err)
	}
}

func TestGetLiveReporting_KeineOffeneSitzung(t *testing.T) {
	q := Query{
		ReportingRepo:       mockReportingRepo{},
		KassensitzungenRepo: mockKasseRepo{kassensitzung: nil},
	}

	result, err := q.GetLiveReporting(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result when no kassensitzung is open, got %+v", result)
	}
}

func TestGetLiveReporting_HappyPath(t *testing.T) {
	datum := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	ks := &kasse.Kassensitzung{
		ZNr:         testKassensitzungNr,
		Bezeichnung: "Sommerfest Tag 1",
		Datum:       datum,
		Status:      kasse.KassensitzungOffen,
	}
	liveData := reporting.LiveReportingData{
		KassensitzungNr:  testKassensitzungNr,
		OffeneSaldiCents: 1200,
		OffeneTische: []reporting.OffenerTisch{
			{TischID: 3, TischName: "Tisch 3", SaldoCents: 1200},
		},
		Summary: reporting.Summary{GesamtUmsatzCents: 45000},
	}

	q := Query{
		ReportingRepo:       mockReportingRepo{liveData: liveData},
		KassensitzungenRepo: mockKasseRepo{kassensitzung: ks},
		TischSessionRepo:    mockTischSessionRepo{},
		TischRepo:           mockTischRepo{},
	}

	result, err := q.GetLiveReporting(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.KassensitzungNr != testKassensitzungNr {
		t.Errorf("expected KassensitzungNr %d, got %d", testKassensitzungNr, result.KassensitzungNr)
	}
	if result.Bezeichnung != "Sommerfest Tag 1" {
		t.Errorf("expected Bezeichnung 'Sommerfest Tag 1', got %q", result.Bezeichnung)
	}
	if !result.Datum.Equal(datum) {
		t.Errorf("expected Datum %v, got %v", datum, result.Datum)
	}
	if result.OffeneSaldiCents != 1200 {
		t.Errorf("expected OffeneSaldiCents 1200, got %d", result.OffeneSaldiCents)
	}
	if len(result.OffeneTische) != 1 {
		t.Errorf("expected 1 offener Tisch, got %d", len(result.OffeneTische))
	}
}

func TestGetLiveReporting_MergesServicekraefteByUserID(t *testing.T) {
	ks := &kasse.Kassensitzung{ZNr: testKassensitzungNr, Status: kasse.KassensitzungOffen}
	liveData := reporting.LiveReportingData{
		KassensitzungNr: testKassensitzungNr,
		Breakdowns: reporting.Breakdowns{
			UmsatzProServicekraft: []reporting.UmsatzServicekraft{
				{UserID: 7, UserName: "Anna", Name: "Anna A.", ZahlungenCents: 1500, AnzahlZahlungen: 2},
				{UserID: 9, UserName: "Cleo", Name: "Cleo C.", ZahlungenCents: 900, AnzahlZahlungen: 1},
			},
		},
	}
	sessions := []kasse.TischSession{
		{
			// Anna (7, hat Umsatz) hat hier noch offene Arbeit.
			TischID: 3,
			UnbezahltePositionen: []kasse.Position{
				{PositionID: "p1", Menge: 1, BestellerUserID: 7, BestellerName: "Anna"},
			},
		},
		{
			// Bert (8) hat offene Arbeit, aber keinen kassierten Umsatz.
			TischID: 1,
			UnbezahltePositionen: []kasse.Position{
				{PositionID: "p2", Menge: 1, BestellerUserID: 8, BestellerName: "Bert"},
			},
		},
	}
	q := Query{
		ReportingRepo:       mockReportingRepo{liveData: liveData},
		KassensitzungenRepo: mockKasseRepo{kassensitzung: ks},
		TischSessionRepo:    mockTischSessionRepo{sessions: sessions},
		TischRepo:           mockTischRepo{tische: []tisch.Tisch{{ID: 3, Name: "Tisch 3"}, {ID: 1, Name: "Tisch 1"}}},
	}

	result, err := q.GetLiveReporting(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Servicekraefte) != 3 {
		t.Fatalf("expected 3 servicekraefte (2 mit Umsatz + Bert ohne), got %d: %+v", len(result.Servicekraefte), result.Servicekraefte)
	}

	// Umsatz-Servicekräfte zuerst, in Umsatz-Reihenfolge.
	anna := result.Servicekraefte[0]
	if anna.UserID != 7 || anna.ZahlungenCents != 1500 || anna.Erledigt {
		t.Errorf("expected Anna mit Umsatz und offener Arbeit, got %+v", anna)
	}
	if len(anna.OffeneTische) != 1 || anna.OffeneTische[0].TischID != 3 || anna.OffeneTische[0].TischName != "Tisch 3" || anna.OffeneTische[0].AnzahlOffen != 1 {
		t.Errorf("expected Anna offen an Tisch 3, got %+v", anna.OffeneTische)
	}

	cleo := result.Servicekraefte[1]
	if cleo.UserID != 9 || !cleo.Erledigt || len(cleo.OffeneTische) != 0 {
		t.Errorf("expected Cleo mit Umsatz aber fertig, got %+v", cleo)
	}

	// Person mit offener Arbeit, aber ohne Umsatz, wird angehängt.
	bert := result.Servicekraefte[2]
	if bert.UserID != 8 || bert.UserName != "Bert" || bert.Name != "" || bert.ZahlungenCents != 0 || bert.Erledigt {
		t.Errorf("expected Bert ohne Umsatz mit offener Arbeit, got %+v", bert)
	}
	if len(bert.OffeneTische) != 1 || bert.OffeneTische[0].TischID != 1 || bert.OffeneTische[0].AnzahlUnbezahlt != 1 {
		t.Errorf("expected Bert offen an Tisch 1, got %+v", bert.OffeneTische)
	}
}

func TestGetLiveReporting_DatabaseError_KassensitzungRepo(t *testing.T) {
	q := Query{
		ReportingRepo:       mockReportingRepo{},
		KassensitzungenRepo: mockKasseRepo{err: errors.New("db error")},
	}

	_, err := q.GetLiveReporting(context.Background())
	if !errors.Is(err, ErrDatabase) {
		t.Fatalf("expected ErrDatabase, got %v", err)
	}
}

func TestGetLiveReporting_DatabaseError_ReportingRepo(t *testing.T) {
	ks := &kasse.Kassensitzung{ZNr: testKassensitzungNr, Status: kasse.KassensitzungOffen}
	q := Query{
		ReportingRepo:       mockReportingRepo{err: errors.New("db error")},
		KassensitzungenRepo: mockKasseRepo{kassensitzung: ks},
	}

	_, err := q.GetLiveReporting(context.Background())
	if !errors.Is(err, ErrDatabase) {
		t.Fatalf("expected ErrDatabase, got %v", err)
	}
}
