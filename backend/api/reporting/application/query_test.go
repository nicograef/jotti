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
	produktStatistik []reporting.ProduktStatistikZeile
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

func (m mockReportingRepo) GetProduktStatistik(_ context.Context, _ int) ([]reporting.ProduktStatistikZeile, error) {
	return m.produktStatistik, m.err
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

func (m mockKasseRepo) GetAbgeschlosseneKassensitzungen(_ context.Context) ([]reporting.AbgeschlosseneSitzung, error) {
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

func TestGetReporting_AggregiertStornierungenProServicekraft(t *testing.T) {
	data := reporting.ReportingData{
		KassensitzungNr: testKassensitzungNr,
		Summary: reporting.Summary{
			AnzahlStornierungen:      3,
			GesamtStornierungenCents: 1050,
		},
		Stornierungen: []reporting.StornierungDetail{
			{Akteur: reporting.ServicekraftRef{UserID: 3, UserName: "felix", Name: "Felix W."}, BetragCents: 500},
			{Akteur: reporting.ServicekraftRef{UserID: 7, UserName: "sophie", Name: "Sophie B."}, BetragCents: 250},
			{Akteur: reporting.ServicekraftRef{UserID: 3, UserName: "felix", Name: "Felix W."}, BetragCents: 300},
		},
	}

	q := Query{ReportingRepo: mockReportingRepo{data: data}, KassensitzungenRepo: mockKasseRepo{kassensitzungNr: testKassensitzungNr}}

	result, err := q.GetReporting(context.Background(), testKassensitzungNr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertStornoAggregat(t, result.Breakdowns.StornierungenProServicekraft, result.Summary)

	// Reihenfolge folgt dem ersten Auftreten (felix vor sophie).
	agg := result.Breakdowns.StornierungenProServicekraft
	if len(agg) != 2 {
		t.Fatalf("expected 2 servicekraefte with stornos, got %d: %+v", len(agg), agg)
	}
	if agg[0].UserID != 3 || agg[0].UserName != "felix" || agg[0].Name != "Felix W." || agg[0].AnzahlStornierungen != 2 || agg[0].StornierungenCents != 800 {
		t.Errorf("unexpected felix aggregate: %+v", agg[0])
	}
	if agg[1].UserID != 7 || agg[1].AnzahlStornierungen != 1 || agg[1].StornierungenCents != 250 {
		t.Errorf("unexpected sophie aggregate: %+v", agg[1])
	}
}

// assertStornoAggregat prüft die Kern-Invariante: die Summe über alle
// Servicekräfte entspricht anzahlStornierungen/gesamtStornierungenCents der Summary.
func assertStornoAggregat(t *testing.T, agg []reporting.StornierungServicekraft, summary reporting.Summary) {
	t.Helper()
	var summeAnzahl, summeCents int
	for _, a := range agg {
		summeAnzahl += a.AnzahlStornierungen
		summeCents += a.StornierungenCents
	}
	if summeAnzahl != summary.AnzahlStornierungen {
		t.Errorf("Summe AnzahlStornierungen %d != Summary %d", summeAnzahl, summary.AnzahlStornierungen)
	}
	if summeCents != summary.GesamtStornierungenCents {
		t.Errorf("Summe StornierungenCents %d != Summary %d", summeCents, summary.GesamtStornierungenCents)
	}
}

func TestGruppiereProduktStatistik_GruppiertSortiertUndSummiert(t *testing.T) {
	// Bewusst unsortierte, kategorieübergreifende Eingabe: Sonstiges vor Essen,
	// Varianten und Produkte in wechselnder Mengen-Reihenfolge.
	zeilen := []reporting.ProduktStatistikZeile{
		{Kategorie: "sonstiges", ProduktName: "Los", VarianteID: 90, VarianteName: "Los", AusgegebeneMenge: 4, UmsatzCents: 400},
		{Kategorie: "getraenk", ProduktName: "Cola", VarianteID: 20, VarianteName: "0,5 l", AusgegebeneMenge: 3, UmsatzCents: 900},
		{Kategorie: "getraenk", ProduktName: "Cola", VarianteID: 21, VarianteName: "0,3 l", AusgegebeneMenge: 10, UmsatzCents: 2000},
		{Kategorie: "essen", ProduktName: "Pommes", VarianteID: 10, VarianteName: "groß", AusgegebeneMenge: 5, UmsatzCents: 1500},
		{Kategorie: "essen", ProduktName: "Pommes", VarianteID: 11, VarianteName: "klein", AusgegebeneMenge: 5, UmsatzCents: 1000},
		{Kategorie: "essen", ProduktName: "Wurst", VarianteID: 12, VarianteName: "Wurst", AusgegebeneMenge: 8, UmsatzCents: 2400},
	}

	produkte := gruppiereProduktStatistik(zeilen)

	// Sechs Varianten-Zeilen fallen zu vier Produkt-Gruppen zusammen
	// (Pommes und Cola je zweivariantig, Wurst und Los einvariantig).
	if len(produkte) != 4 {
		t.Fatalf("expected 4 produkte, got %d: %+v", len(produkte), produkte)
	}

	// Essen zuerst; Pommes (Menge 10) vor Wurst (Menge 8).
	if produkte[0].Kategorie != "essen" || produkte[0].ProduktName != "Pommes" {
		t.Fatalf("expected Pommes first in Essen, got %+v", produkte[0])
	}
	if produkte[0].AusgegebeneMenge != 10 || produkte[0].UmsatzCents != 2500 {
		t.Errorf("expected Pommes subtotal menge 10 / umsatz 2500, got %d / %d", produkte[0].AusgegebeneMenge, produkte[0].UmsatzCents)
	}
	// Varianten je Produkt nach Menge absteigend, Name als Tiebreaker bei
	// Gleichstand (groß vor klein: beide 5).
	if len(produkte[0].Varianten) != 2 || produkte[0].Varianten[0].VarianteName != "groß" || produkte[0].Varianten[1].VarianteName != "klein" {
		t.Errorf("expected Pommes-Varianten groß, klein (Name-Tiebreaker), got %+v", produkte[0].Varianten)
	}

	if produkte[1].Kategorie != "essen" || produkte[1].ProduktName != "Wurst" {
		t.Errorf("expected Wurst second in Essen, got %+v", produkte[1])
	}

	// Getränke nach Essen; Cola-Varianten 0,3 l (10) vor 0,5 l (3).
	if produkte[2].Kategorie != "getraenk" || produkte[2].ProduktName != "Cola" {
		t.Fatalf("expected Cola in Getränke, got %+v", produkte[2])
	}
	if produkte[2].AusgegebeneMenge != 13 || produkte[2].UmsatzCents != 2900 {
		t.Errorf("expected Cola subtotal menge 13 / umsatz 2900, got %d / %d", produkte[2].AusgegebeneMenge, produkte[2].UmsatzCents)
	}
	if produkte[2].Varianten[0].VarianteName != "0,3 l" || produkte[2].Varianten[1].VarianteName != "0,5 l" {
		t.Errorf("expected Cola-Varianten 0,3 l vor 0,5 l (Menge absteigend), got %+v", produkte[2].Varianten)
	}

	// Sonstiges zuletzt; Ein-Varianten-Produkt behält genau eine Variante.
	if produkte[3].Kategorie != "sonstiges" || produkte[3].ProduktName != "Los" || len(produkte[3].Varianten) != 1 {
		t.Errorf("expected Los as single-variant Sonstiges product, got %+v", produkte[3])
	}
}

func TestGruppiereProduktStatistik_LeereEingabe(t *testing.T) {
	produkte := gruppiereProduktStatistik(nil)
	if produkte == nil {
		t.Fatal("expected non-nil empty slice for empty input")
	}
	if len(produkte) != 0 {
		t.Errorf("expected empty result, got %+v", produkte)
	}
}

func TestGetReporting_ReichtProduktStatistikDurch(t *testing.T) {
	zeilen := []reporting.ProduktStatistikZeile{
		{Kategorie: "essen", ProduktName: "Pommes", VarianteID: 10, VarianteName: "groß", AusgegebeneMenge: 5, UmsatzCents: 1500},
	}
	q := Query{
		ReportingRepo:       mockReportingRepo{produktStatistik: zeilen},
		KassensitzungenRepo: mockKasseRepo{kassensitzungNr: testKassensitzungNr},
	}

	result, err := q.GetReporting(context.Background(), testKassensitzungNr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.ProduktStatistik) != 1 || result.ProduktStatistik[0].ProduktName != "Pommes" {
		t.Fatalf("expected ProduktStatistik with Pommes, got %+v", result.ProduktStatistik)
	}
	if result.ProduktStatistik[0].AusgegebeneMenge != 5 || result.ProduktStatistik[0].UmsatzCents != 1500 {
		t.Errorf("expected Pommes subtotal 5 / 1500, got %+v", result.ProduktStatistik[0])
	}
}

func TestGetLiveReporting_ReichtProduktStatistikDurch(t *testing.T) {
	ks := &kasse.Kassensitzung{ZNr: testKassensitzungNr, Status: kasse.KassensitzungOffen}
	zeilen := []reporting.ProduktStatistikZeile{
		{Kategorie: "getraenk", ProduktName: "Cola", VarianteID: 20, VarianteName: "0,5 l", AusgegebeneMenge: 3, UmsatzCents: 900},
	}
	q := Query{
		ReportingRepo:       mockReportingRepo{produktStatistik: zeilen},
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
	if len(result.ProduktStatistik) != 1 || result.ProduktStatistik[0].ProduktName != "Cola" {
		t.Fatalf("expected ProduktStatistik with Cola, got %+v", result.ProduktStatistik)
	}
}

func TestGetReporting_DatabaseError(t *testing.T) {
	q := Query{ReportingRepo: mockReportingRepo{err: errors.New("db connection failed")}, KassensitzungenRepo: mockKasseRepo{kassensitzungNr: testKassensitzungNr}}

	_, err := q.GetReporting(context.Background(), testKassensitzungNr)
	if !errors.Is(err, ErrDatabase) {
		t.Fatalf("expected ErrDatabase, got %v", err)
	}
}

func TestGetEigeneUebersicht_ReichtStatistikDurch(t *testing.T) {
	base := reporting.EigeneUebersicht{
		AnzahlBestellungen: 4,
		BestellungenCents:  3000,
		AnzahlZahlungen:    2,
		ZahlungenCents:     1500,
	}
	q := Query{
		ReportingRepo:       mockReportingRepo{eigeneUebersicht: base},
		KassensitzungenRepo: mockKasseRepo{kassensitzungNr: testKassensitzungNr},
	}

	result, err := q.GetEigeneUebersicht(context.Background(), 7)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.AnzahlBestellungen != 4 || result.BestellungenCents != 3000 || result.AnzahlZahlungen != 2 || result.ZahlungenCents != 1500 {
		t.Errorf("expected base reporting numbers to be preserved, got %+v", result)
	}
}

func TestGetEigeneUebersicht_DatabaseError(t *testing.T) {
	q := Query{
		ReportingRepo:       mockReportingRepo{err: errors.New("db error")},
		KassensitzungenRepo: mockKasseRepo{kassensitzungNr: testKassensitzungNr},
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
				{PositionID: "p1", Menge: 2, EinzelpreisCents: 375, BestellerUserID: 7, BestellerName: "Anna"},
			},
		},
		{
			// Bert (8) hat offene Arbeit, aber keinen kassierten Umsatz.
			TischID: 1,
			UnbezahltePositionen: []kasse.Position{
				{PositionID: "p2", Menge: 1, EinzelpreisCents: 300, BestellerUserID: 8, BestellerName: "Bert"},
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
	// OffenCents wird aus der Domäne durchgereicht: 2 × 375 = 750 Cent.
	if anna.OffeneTische[0].OffenCents != 750 {
		t.Errorf("expected Anna OffenCents 750, got %d", anna.OffeneTische[0].OffenCents)
	}
	// Der offene Betrag wird auf Servicekraft-Ebene aggregiert (Summe über Tische).
	if anna.OffenCents != 750 {
		t.Errorf("expected Anna Servicekraft-OffenCents 750, got %d", anna.OffenCents)
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
	if len(bert.OffeneTische) != 1 || bert.OffeneTische[0].TischID != 1 || bert.OffeneTische[0].AnzahlOffen != 1 {
		t.Errorf("expected Bert offen an Tisch 1, got %+v", bert.OffeneTische)
	}
	if bert.OffeneTische[0].OffenCents != 300 {
		t.Errorf("expected Bert OffenCents 300, got %d", bert.OffeneTische[0].OffenCents)
	}
	if bert.OffenCents != 300 {
		t.Errorf("expected Bert Servicekraft-OffenCents 300, got %d", bert.OffenCents)
	}
}

func TestGetLiveReporting_AggregiertStornierungenProServicekraft(t *testing.T) {
	ks := &kasse.Kassensitzung{ZNr: testKassensitzungNr, Status: kasse.KassensitzungOffen}
	liveData := reporting.LiveReportingData{
		KassensitzungNr: testKassensitzungNr,
		Summary: reporting.Summary{
			AnzahlStornierungen:      2,
			GesamtStornierungenCents: 750,
		},
		Stornierungen: []reporting.StornierungDetail{
			{Akteur: reporting.ServicekraftRef{UserID: 3, UserName: "felix", Name: "Felix W."}, BetragCents: 500},
			{Akteur: reporting.ServicekraftRef{UserID: 7, UserName: "sophie", Name: "Sophie B."}, BetragCents: 250},
		},
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

	assertStornoAggregat(t, result.Breakdowns.StornierungenProServicekraft, result.Summary)
	if len(result.Breakdowns.StornierungenProServicekraft) != 2 {
		t.Fatalf("expected 2 servicekraefte with stornos, got %+v", result.Breakdowns.StornierungenProServicekraft)
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
