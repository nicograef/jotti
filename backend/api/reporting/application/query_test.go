//go:build unit

package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nicograef/jotti/backend/domain/reporting"
)

type mockReportingRepo struct {
	data reporting.ReportingData
	err  error
}

func (m mockReportingRepo) GetReporting(_ context.Context, _ reporting.Zeitraum) (reporting.ReportingData, error) {
	return m.data, m.err
}

var testZeitraum = reporting.Zeitraum{
	Von: time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC),
	Bis: time.Date(2026, 3, 14, 23, 59, 0, 0, time.UTC),
}

func TestGetReporting_HappyPath(t *testing.T) {
	expected := reporting.ReportingData{
		Zeitraum: testZeitraum,
		Summary:  reporting.Summary{GesamtUmsatzCents: 5000},
		Breakdowns: reporting.Breakdowns{
			UmsatzProServicekraft: []reporting.UmsatzServicekraft{},
			UmsatzProTisch:        []reporting.UmsatzTisch{},
		},
		Stornierungen: []reporting.StornierungDetail{},
	}

	q := Query{ReportingRepo: mockReportingRepo{data: expected}}

	result, err := q.GetReporting(context.Background(), testZeitraum)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Summary.GesamtUmsatzCents != 5000 {
		t.Errorf("expected 5000 cents, got %d", result.Summary.GesamtUmsatzCents)
	}
	if result.Zeitraum.Von != testZeitraum.Von {
		t.Errorf("expected Von %v, got %v", testZeitraum.Von, result.Zeitraum.Von)
	}
}

func TestGetReporting_DatabaseError(t *testing.T) {
	q := Query{ReportingRepo: mockReportingRepo{err: errors.New("db connection failed")}}

	_, err := q.GetReporting(context.Background(), testZeitraum)
	if !errors.Is(err, ErrDatabase) {
		t.Fatalf("expected ErrDatabase, got %v", err)
	}
}
