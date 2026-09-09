//go:build unit

package application

import (
	"testing"
	"time"
)

// 2026-07-01T23:30:00Z ist in Europe/Berlin (Sommerzeit, UTC+2) bereits der
// 2. Juli, 01:30. Der Archivname trägt die Ortszeit, nicht die UTC-Uhr.
func TestDateiname_ZeitstempelInBerlinerZeit(t *testing.T) {
	got := dateiname("JOTTI-1", 7, time.Date(2026, 7, 1, 23, 30, 0, 0, time.UTC))

	want := "dsfinvk_JOTTI-1_kassensitzung-7_20260702-013000.zip"
	if got != want {
		t.Errorf("dateiname = %q, want %q", got, want)
	}
}
