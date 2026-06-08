//go:build unit

package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/kasse"
	"github.com/nicograef/jotti/backend/domain/reporting"
)

type mockReportingRepo struct {
	data     reporting.ReportingData
	liveData reporting.LiveReportingData
	err      error
}

func (m mockReportingRepo) GetReporting(_ context.Context, _ int) (reporting.ReportingData, error) {
	return m.data, m.err
}

func (m mockReportingRepo) GetEigeneUebersicht(_ context.Context, _ int, _ int) (reporting.EigeneUebersicht, error) {
	return reporting.EigeneUebersicht{}, m.err
}

func (m mockReportingRepo) GetLiveReporting(_ context.Context, _ int) (reporting.LiveReportingData, error) {
	return m.liveData, m.err
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
			UmsatzProTisch:        []reporting.UmsatzTisch{},
		},
		Stornierungen: []reporting.StornierungDetail{},
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

func TestGetReporting_DatabaseError(t *testing.T) {
	q := Query{ReportingRepo: mockReportingRepo{err: errors.New("db connection failed")}, KassensitzungenRepo: mockKasseRepo{kassensitzungNr: testKassensitzungNr}}

	_, err := q.GetReporting(context.Background(), testKassensitzungNr)
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
