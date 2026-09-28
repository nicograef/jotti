package application

import (
	"testing"
	"time"
)

// 2026-07-01T23:30:00Z is already 2 July, 01:30 in Europe/Berlin (summer time, UTC+2).
// The archive name uses local time, not UTC.
func TestDateiname_ZeitstempelInBerlinerZeit(t *testing.T) {
	got := dateiname("JOTTI-1", 7, time.Date(2026, 7, 1, 23, 30, 0, 0, time.UTC))

	want := "dsfinvk_JOTTI-1_kassensitzung-7_20260702-013000.zip"
	if got != want {
		t.Errorf("dateiname = %q, want %q", got, want)
	}
}
